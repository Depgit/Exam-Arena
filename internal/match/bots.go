package match

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	randv2 "math/rand/v2"
	"sync"
	"time"

	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/passwords"
	"github.com/exam-arena/internal/platform/realtime"
	"github.com/exam-arena/internal/questions"
)

// Bot opponents.
//
// With few players online, most searches find nobody. Instead of a dead
// "Searching…" screen, a player can battle a bot. The bot plays through
// SubmitAnswer exactly like a human, so scoring, live score updates, the
// early finish and match_end all work unchanged. Bot matches are unrated:
// no rating or win/loss change, so bots can't be farmed for rating.

type botProfile struct {
	Username    string // reserved prefix "bot." — registration refuses it
	DisplayName string
	Rating      int     // shown to the player; never changes
	Accuracy    float64 // chance of answering each question correctly
	MinDelay    time.Duration
	MaxDelay    time.Duration
}

// Ordered weakest → strongest; botFor picks by the player's rating.
var botProfiles = []botProfile{
	{"bot.rookie", "🤖 Rookie", 1050, 0.50, 5 * time.Second, 8 * time.Second},
	{"bot.ace", "🤖 Ace", 1250, 0.70, 4 * time.Second, 7 * time.Second},
	{"bot.master", "🤖 Master", 1450, 0.88, 2 * time.Second, 5 * time.Second},
}

// botFor picks the bot closest to the player's level.
func botFor(rating int) botProfile {
	switch {
	case rating < 1150:
		return botProfiles[0]
	case rating < 1350:
		return botProfiles[1]
	default:
		return botProfiles[2]
	}
}

var (
	botIDsMu sync.Mutex
	botIDs   = map[string]string{} // username → user id, created on first use
)

func (s *Service) botUserID(ctx context.Context, b botProfile) (string, error) {
	botIDsMu.Lock()
	defer botIDsMu.Unlock()
	if id, ok := botIDs[b.Username]; ok {
		return id, nil
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	hash, err := passwords.Hash(hex.EncodeToString(secret))
	if err != nil {
		return "", err
	}
	id, err := s.userRepo.EnsureBot(ctx, b.Username, b.DisplayName, hash)
	if err != nil {
		return "", err
	}
	botIDs[b.Username] = id
	return id, nil
}

// StartBotMatch starts an unrated match between the player and a bot.
// The player receives match_start over WebSocket, like any other match.
func (s *Service) StartBotMatch(ctx context.Context, userID, username, categoryID string) (string, error) {
	if !s.HasEnoughQuestions(categoryID) {
		return "", questions.ErrNotEnoughQuestions
	}

	// Leaving the queue is part of choosing a bot; do it server-side too so
	// a slow client can't end up both in a bot match and paired by the engine.
	s.queue.Remove(userID)
	if err := s.queueStore.RemoveFromQueue(ctx, userID); err != nil {
		slog.Warn("failed to clear queue mirror for bot match", "error", err, "user", userID)
	}

	rating, err := s.userRepo.EnsureRating(ctx, userID, categoryID)
	if err != nil {
		return "", fmt.Errorf("ensure rating: %w", err)
	}
	bot := botFor(rating.Rating)
	botID, err := s.botUserID(ctx, bot)
	if err != nil {
		return "", err
	}

	picked, err := s.questionBank.GetRandom(categoryID, QuestionsPerMatch)
	if err != nil {
		return "", questions.ErrNotEnoughQuestions
	}

	match, err := s.matchRepo.CreateMatch(ctx, "arena", categoryID, MatchTimerSeconds, 2, nil)
	if err != nil {
		return "", fmt.Errorf("create bot match: %w", err)
	}
	qIDs := make([]string, len(picked))
	for i, q := range picked {
		qIDs[i] = q.ID
	}
	if err := s.matchRepo.AddMatchQuestions(ctx, match.ID, qIDs); err != nil {
		slog.Error("failed to save bot match questions", "error", err)
	}
	ratingBefore := rating.Rating
	if _, err := s.matchRepo.AddPlayer(ctx, match.ID, userID, &ratingBefore); err != nil {
		return "", fmt.Errorf("add player: %w", err)
	}
	if _, err := s.matchRepo.AddPlayer(ctx, match.ID, botID, nil); err != nil {
		return "", fmt.Errorf("add bot: %w", err)
	}
	if err := s.matchRepo.UpdateMatchStatus(ctx, match.ID, "in_progress"); err != nil {
		slog.Error("failed to start bot match", "error", err)
	}

	timerCtx, cancelTimer := context.WithCancel(context.Background())
	lm := &LiveMatch{
		MatchID:      match.ID,
		CategoryID:   categoryID,
		MatchType:    "arena",
		Questions:    picked,
		PlayerStates: map[string]*LivePlayer{},
		StartedAt:    time.Now(),
		TimerSeconds: MatchTimerSeconds,
		BotUserID:    botID,
		cancelTimer:  cancelTimer,
	}
	for _, p := range []struct{ id, name string }{{userID, username}, {botID, bot.DisplayName}} {
		lm.PlayerStates[p.id] = &LivePlayer{UserID: p.id, Username: p.name, AnsweredIDs: map[string]bool{}}
	}
	s.cache.Set(liveMatchKey(match.ID), lm, 35*time.Minute)

	s.hub.SendToUser(userID, realtime.Message{
		Type: "match_start",
		Payload: map[string]interface{}{
			"match_id":      match.ID,
			"match_type":    "bot",
			"timer_seconds": MatchTimerSeconds,
			"questions":     questions.ForPlayers(picked, matchShuffleKey(match.ID, userID)),
			"players": []map[string]interface{}{
				{"user_id": userID, "username": username, "rating": rating.Rating},
				{"user_id": botID, "username": bot.DisplayName, "rating": bot.Rating, "is_bot": true},
			},
		},
	})

	entries := []matchmaking.Entry{{UserID: userID, Username: username}, {UserID: botID, Username: bot.DisplayName}}
	go s.runMatchTimer(timerCtx, match.ID, entries, time.Duration(MatchTimerSeconds)*time.Second)
	go s.runBot(timerCtx, lm, botID, bot)

	slog.Info("bot match started", "match_id", match.ID, "player", username, "bot", bot.Username)
	return match.ID, nil
}

// runBot answers every question with human-like timing and accuracy, until
// it's done or the match ends. It submits through the normal answer path.
func (s *Service) runBot(ctx context.Context, lm *LiveMatch, botID string, b botProfile) {
	r := randv2.New(randv2.NewPCG(uint64(time.Now().UnixNano()), randv2.Uint64()))
	for _, q := range lm.Questions {
		delay := b.MinDelay + time.Duration(r.Int64N(int64(b.MaxDelay-b.MinDelay)+1))
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		err := s.SubmitAnswer(context.Background(), AnswerRequest{
			MatchID:     lm.MatchID,
			UserID:      botID,
			QuestionID:  q.ID,
			OptionID:    botPick(r, q, b.Accuracy),
			TimeTakenMs: int(delay / time.Millisecond),
		})
		if err != nil {
			return // match over
		}
	}
}

// botPick returns the correct option with probability accuracy, otherwise
// a random wrong one.
func botPick(r *randv2.Rand, q models.Question, accuracy float64) string {
	var correct string
	var wrong []string
	for _, o := range q.Options {
		if o.IsCorrect {
			correct = o.ID
		} else {
			wrong = append(wrong, o.ID)
		}
	}
	if correct != "" && (len(wrong) == 0 || r.Float64() < accuracy) {
		return correct
	}
	if len(wrong) == 0 {
		return ""
	}
	return wrong[r.IntN(len(wrong))]
}
