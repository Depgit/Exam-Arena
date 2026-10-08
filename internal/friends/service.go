package friends

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/exam-arena/internal/match"
	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/cache"
	"github.com/exam-arena/internal/platform/database"
	"github.com/exam-arena/internal/platform/realtime"
	"github.com/exam-arena/internal/questions"
	"github.com/exam-arena/internal/users"
)

// ChallengeTTL is how long a friend challenge can be declined or cancelled
// through the friends API. The underlying room code keeps working like any
// other friend match room.
const ChallengeTTL = 10 * time.Minute

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrCannotFriendSelf   = errors.New("you cannot add yourself as a friend")
	ErrAlreadyFriends     = errors.New("you are already friends")
	ErrRequestPending     = errors.New("friend request already sent")
	ErrRequestUnavailable = errors.New("cannot send a friend request to this user")
	ErrRequestNotFound    = errors.New("friend request not found")
	ErrNotFriends         = errors.New("you are not friends with this user")
	ErrFriendOffline      = errors.New("your friend is offline")
	ErrCategoryNotFound   = errors.New("exam category not found")
	ErrChallengeNotFound  = errors.New("challenge not found or no longer pending")
)

// friendChallenge is kept in the in-memory cache while a challenge is open,
// so either side can decline/cancel it and the other side can be told.
type friendChallenge struct {
	MatchID        string
	RoomCode       string
	ChallengerID   string
	ChallengerName string
	FriendID       string
}

func challengeKey(matchID string) string { return "challenge:" + matchID }

type Service struct {
	friendRepo   *Store
	userRepo     *users.Store
	matchRepo    *match.Store
	questionRepo *questions.Store
	matchService *match.Service
	hub          *realtime.Hub
	cache        cache.Cache
}

func NewService(
	friendRepo *Store,
	userRepo *users.Store,
	matchRepo *match.Store,
	questionRepo *questions.Store,
	matchService *match.Service,
	hub *realtime.Hub,
	cache cache.Cache,
) *Service {
	return &Service{
		friendRepo:   friendRepo,
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		questionRepo: questionRepo,
		matchService: matchService,
		hub:          hub,
		cache:        cache,
	}
}

// List returns the user's accepted friends and pending requests in both
// directions, with each friend's live online status.
func (s *Service) List(ctx context.Context, userID string) (*models.FriendsList, error) {
	entries, err := s.friendRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	list := &models.FriendsList{
		Friends:  []models.FriendEntry{},
		Incoming: []models.FriendEntry{},
		Outgoing: []models.FriendEntry{},
	}
	for _, e := range entries {
		e.Online = s.hub.IsOnline(e.UserID)
		switch {
		case e.Status == "accepted":
			list.Friends = append(list.Friends, e)
		case e.Direction == "incoming":
			list.Incoming = append(list.Incoming, e)
		default:
			list.Outgoing = append(list.Outgoing, e)
		}
	}
	return list, nil
}

// SendRequest sends a friend request to the user with the given username.
// If that user already sent us a pending request, it is accepted instead and
// accepted is true.
func (s *Service) SendRequest(ctx context.Context, userID, username, targetUsername string) (f *models.Friendship, accepted bool, err error) {
	target, err := s.userRepo.GetByUsername(ctx, strings.TrimSpace(targetUsername))
	if err != nil {
		return nil, false, err
	}
	if target == nil || target.Status != "active" {
		return nil, false, ErrUserNotFound
	}
	if target.ID == userID {
		return nil, false, ErrCannotFriendSelf
	}

	existing, err := s.friendRepo.GetBetween(ctx, userID, target.ID)
	if err != nil {
		return nil, false, err
	}

	switch {
	case existing == nil:
		f, err = s.friendRepo.Create(ctx, userID, target.ID)
		if database.IsUniqueViolation(err) {
			// The other user sent a request at the same moment.
			return nil, false, ErrRequestPending
		}
	case existing.Status == "accepted":
		return nil, false, ErrAlreadyFriends
	case existing.Status == "blocked":
		return nil, false, ErrRequestUnavailable
	case existing.Status == "pending" && existing.RequesterID == userID:
		return nil, false, ErrRequestPending
	case existing.Status == "pending":
		// Mutual request: accept theirs.
		f, err = s.friendRepo.Respond(ctx, existing.ID, "accepted")
		if err != nil {
			return nil, false, err
		}
		if f == nil {
			return nil, false, ErrRequestNotFound
		}
		s.hub.SendToUser(target.ID, realtime.Message{
			Type: "friend_request_accepted",
			Payload: map[string]interface{}{
				"friendship_id": f.ID,
				"user_id":       userID,
				"username":      username,
			},
		})
		return f, true, nil
	default: // declined — allow asking again
		f, err = s.friendRepo.Reopen(ctx, existing.ID, userID, target.ID)
		if err == nil && f == nil {
			return nil, false, ErrRequestPending
		}
	}
	if err != nil {
		return nil, false, err
	}

	s.hub.SendToUser(target.ID, realtime.Message{
		Type: "friend_request",
		Payload: map[string]interface{}{
			"friendship_id": f.ID,
			"from_user_id":  userID,
			"from_username": username,
		},
	})
	return f, false, nil
}

// Respond accepts or declines a pending request addressed to userID.
func (s *Service) Respond(ctx context.Context, userID, username, friendshipID string, accept bool) (*models.Friendship, error) {
	existing, err := s.friendRepo.GetByID(ctx, friendshipID)
	if err != nil {
		if database.IsInvalidInput(err) {
			return nil, ErrRequestNotFound
		}
		return nil, err
	}
	if existing == nil || existing.AddresseeID != userID || existing.Status != "pending" {
		return nil, ErrRequestNotFound
	}

	status := "declined"
	if accept {
		status = "accepted"
	}
	f, err := s.friendRepo.Respond(ctx, friendshipID, status)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, ErrRequestNotFound
	}

	if accept {
		s.hub.SendToUser(f.RequesterID, realtime.Message{
			Type: "friend_request_accepted",
			Payload: map[string]interface{}{
				"friendship_id": f.ID,
				"user_id":       userID,
				"username":      username,
			},
		})
	}
	return f, nil
}

// Remove deletes the friendship between userID and otherUserID. It covers
// unfriending, withdrawing an outgoing request and dismissing an incoming one.
func (s *Service) Remove(ctx context.Context, userID, otherUserID string) error {
	existing, err := s.friendRepo.GetBetween(ctx, userID, otherUserID)
	if err != nil {
		if database.IsInvalidInput(err) {
			return ErrNotFriends
		}
		return err
	}
	if existing == nil || (existing.Status != "accepted" && existing.Status != "pending") {
		return ErrNotFriends
	}
	return s.friendRepo.Delete(ctx, existing.ID)
}

// ChallengeResult is returned to the challenger.
type ChallengeResult struct {
	MatchID          string `json:"match_id"`
	RoomCode         string `json:"room_code"`
	FriendID         string `json:"friend_id"`
	ExamCategoryID   string `json:"exam_category_id"`
	ExamCategoryName string `json:"exam_category_name"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// Challenge opens a friend match room and invites friendID to it over the
// WebSocket. The friend accepts by joining the room code
// (POST /matches/friend/join), or declines via DeclineOrCancelChallenge.
func (s *Service) Challenge(ctx context.Context, userID, username, friendID, categoryID string) (*ChallengeResult, error) {
	if friendID == userID {
		return nil, ErrNotFriends
	}
	f, err := s.friendRepo.GetBetween(ctx, userID, friendID)
	if err != nil {
		if database.IsInvalidInput(err) {
			return nil, ErrNotFriends
		}
		return nil, err
	}
	if f == nil || f.Status != "accepted" {
		return nil, ErrNotFriends
	}
	if !s.hub.IsOnline(friendID) {
		return nil, ErrFriendOffline
	}

	category, err := s.questionRepo.GetActiveCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}
	if !s.matchService.HasEnoughQuestions(category.ID) {
		return nil, questions.ErrNotEnoughQuestions
	}

	match, roomCode, err := s.matchService.CreateFriendMatch(ctx, userID, username, category.ID)
	if err != nil {
		return nil, fmt.Errorf("create challenge match: %w", err)
	}

	s.cache.Set(challengeKey(match.ID), &friendChallenge{
		MatchID:        match.ID,
		RoomCode:       roomCode,
		ChallengerID:   userID,
		ChallengerName: username,
		FriendID:       friendID,
	}, ChallengeTTL)

	s.hub.SendToUser(friendID, realtime.Message{
		Type: "friend_challenge",
		Payload: map[string]interface{}{
			"match_id":           match.ID,
			"room_code":          roomCode,
			"from_user_id":       userID,
			"from_username":      username,
			"exam_category_id":   category.ID,
			"exam_category_name": category.Name,
			"expires_in_seconds": int(ChallengeTTL.Seconds()),
		},
	})

	slog.Info("friend challenge sent", "match_id", match.ID, "from", userID, "to", friendID)

	return &ChallengeResult{
		MatchID:          match.ID,
		RoomCode:         roomCode,
		FriendID:         friendID,
		ExamCategoryID:   category.ID,
		ExamCategoryName: category.Name,
		ExpiresInSeconds: int(ChallengeTTL.Seconds()),
	}, nil
}

// DeclineOrCancelChallenge closes an open challenge. The invited friend
// declines it; the challenger cancels it. The waiting match is cancelled and
// the other side is notified.
func (s *Service) DeclineOrCancelChallenge(ctx context.Context, userID, username, matchID string) error {
	raw, ok := s.cache.Get(challengeKey(matchID))
	if !ok {
		return ErrChallengeNotFound
	}
	ch := raw.(*friendChallenge)
	if userID != ch.FriendID && userID != ch.ChallengerID {
		return ErrChallengeNotFound
	}

	cancelled, err := s.matchRepo.CancelWaitingMatch(ctx, matchID)
	if err != nil {
		return err
	}
	s.cache.Delete(challengeKey(matchID))
	if !cancelled {
		// The friend already joined and the match started.
		return ErrChallengeNotFound
	}

	msgType, notify := "friend_challenge_declined", ch.ChallengerID
	if userID == ch.ChallengerID {
		msgType, notify = "friend_challenge_cancelled", ch.FriendID
	}
	s.hub.SendToUser(notify, realtime.Message{
		Type: msgType,
		Payload: map[string]interface{}{
			"match_id":    matchID,
			"by_user_id":  userID,
			"by_username": username,
		},
	})
	return nil
}
