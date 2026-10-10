package chat

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/exam-arena/internal/platform/realtime"
)

// Message is one chat line as players see it.
type Message struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name,omitempty"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	MaxLength   = 300 // characters per message
	historySize = 50  // messages shown when the chat opens
	keepInDB    = 500 // older rows are trimmed

	// Flood control per player: at most burstMax messages per burstWindow,
	// and the same text can't be repeated within repeatWindow.
	burstMax     = 5
	burstWindow  = 10 * time.Second
	repeatWindow = 30 * time.Second
)

var (
	ErrEmpty    = errors.New("message is empty")
	ErrTooLong  = errors.New("message is too long (max 300 characters)")
	ErrTooFast  = errors.New("you're sending messages too fast — wait a few seconds")
	ErrRepeated = errors.New("you just sent that")
)

// Service runs the global chat: checks and stores new messages, keeps the
// latest ones in memory, and pushes each one live to every connected player
// as a "chat_message" WebSocket event.
type Service struct {
	store *Store
	hub   *realtime.Hub

	mu      sync.Mutex
	recent  []Message             // newest last, at most historySize
	senders map[string]*senderLog // flood control per user
}

type senderLog struct {
	times    []time.Time
	lastBody string
	lastAt   time.Time
}

func NewService(store *Store, hub *realtime.Hub) *Service {
	return &Service{store: store, hub: hub, senders: map[string]*senderLog{}}
}

// Load fills the in-memory history from the database (call once at start).
func (s *Service) Load(ctx context.Context) error {
	msgs, err := s.store.Recent(ctx, historySize)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.recent = msgs
	s.mu.Unlock()
	return nil
}

// Recent returns the latest messages, oldest first. No database call.
func (s *Service) Recent() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message(nil), s.recent...)
}

// Clean trims spaces, turns line breaks and other control characters into
// single spaces, and enforces the length limit.
func Clean(body string) (string, error) {
	body = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, body)
	body = strings.Join(strings.Fields(body), " ")
	if body == "" {
		return "", ErrEmpty
	}
	if len([]rune(body)) > MaxLength {
		return "", ErrTooLong
	}
	return body, nil
}

// allow applies flood control for one user at time now.
func (s *Service) allow(userID, body string, now time.Time) error {
	log := s.senders[userID]
	if log == nil {
		log = &senderLog{}
		s.senders[userID] = log
	}
	kept := log.times[:0]
	for _, t := range log.times {
		if now.Sub(t) < burstWindow {
			kept = append(kept, t)
		}
	}
	log.times = kept
	if len(log.times) >= burstMax {
		return ErrTooFast
	}
	if strings.EqualFold(body, log.lastBody) && now.Sub(log.lastAt) < repeatWindow {
		return ErrRepeated
	}
	log.times = append(log.times, now)
	log.lastBody, log.lastAt = body, now
	return nil
}

// Send checks, stores and broadcasts a message from a player.
func (s *Service) Send(ctx context.Context, userID, username, body string) (*Message, error) {
	clean, err := Clean(body)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	err = s.allow(userID, clean, time.Now())
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}

	m := &Message{UserID: userID, Username: username, Body: clean}
	if err := s.store.Save(ctx, m); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.recent = append(s.recent, *m)
	if len(s.recent) > historySize {
		s.recent = s.recent[len(s.recent)-historySize:]
	}
	trim := m.ID%50 == 0 // every 50th message, trim old rows
	s.mu.Unlock()

	s.hub.Broadcast(realtime.Message{Type: "chat_message", Payload: m.payload()})
	if trim {
		go func() {
			if err := s.store.KeepNewest(context.Background(), keepInDB); err != nil {
				slog.Warn("chat trim failed", "error", err)
			}
		}()
	}
	return m, nil
}

// payload is the message in the shape WebSocket events carry (same JSON
// fields as Message).
func (m *Message) payload() map[string]interface{} {
	p := map[string]interface{}{
		"id":         m.ID,
		"user_id":    m.UserID,
		"username":   m.Username,
		"body":       m.Body,
		"created_at": m.CreatedAt,
	}
	if m.DisplayName != nil {
		p["display_name"] = *m.DisplayName
	}
	return p
}
