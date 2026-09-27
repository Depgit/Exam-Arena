package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
)

// MatchmakingService is a thin adapter between HTTP handlers and the
// in-memory Queue. The Engine drives the pairing loop; this service
// only exposes Join / Leave / Stats to the outside world.
type MatchmakingService struct {
	matchRepo *repository.MatchRepo
	userRepo  *repository.UserRepo
	queue     *matchmaking.Queue
}

func NewMatchmakingService(
	matchRepo *repository.MatchRepo,
	userRepo *repository.UserRepo,
	queue *matchmaking.Queue,
) *MatchmakingService {
	return &MatchmakingService{
		matchRepo: matchRepo,
		userRepo:  userRepo,
		queue:     queue,
	}
}

// JoinQueueRequest is decoded from the HTTP request body.
type JoinQueueRequest struct {
	ExamCategoryID string `json:"exam_category_id"`
	MatchType      string `json:"match_type"`
}

// JoinQueue adds the player to the in-memory queue and mirrors the entry
// to Postgres so the queue can be reconstructed after a server restart.
func (s *MatchmakingService) JoinQueue(ctx context.Context, userID, username string, req JoinQueueRequest) error {
	if req.ExamCategoryID == "" {
		return fmt.Errorf("exam_category_id is required")
	}
	switch req.MatchType {
	case "":
		req.MatchType = "ranked"
	case "ranked", matchmaking.MatchTypeArena:
	case "daily_challenge":
		return fmt.Errorf("the daily challenge is played from the home page, not the queue")
	default:
		return fmt.Errorf("match_type must be \"ranked\" or \"arena\"")
	}

	// Ensure a rating row exists so we have a meaningful Elo score.
	rating, err := s.userRepo.EnsureRating(ctx, userID, req.ExamCategoryID)
	if err != nil {
		return fmt.Errorf("ensure rating: %w", err)
	}

	// Build the in-memory queue entry.
	entry := matchmaking.Entry{
		UserID:         userID,
		Username:       username,
		ExamCategoryID: req.ExamCategoryID,
		MatchType:      req.MatchType,
		Rating:         rating.Rating,
		QueuedAt:       time.Now(),
	}

	// Push to in-memory queue (replaces any stale entry for this user).
	s.queue.Push(entry)

	// Mirror to Postgres (best-effort — the in-memory queue is authoritative).
	dbEntry := toDBQueueEntry(entry)
	if err := s.matchRepo.AddToQueue(ctx, &dbEntry); err != nil {
		// Log but do not fail — the player is already in the live queue.
		slog.Warn("failed to mirror queue entry to DB", "user", userID, "error", err)
	}

	slog.Info("player queued",
		"user", username,
		"user_id", userID,
		"rating", rating.Rating,
		"category", req.ExamCategoryID,
		"match_type", req.MatchType,
	)
	return nil
}

// LeaveQueue removes the player from both the in-memory queue and Postgres.
func (s *MatchmakingService) LeaveQueue(ctx context.Context, userID string) error {
	s.queue.Remove(userID)

	if err := s.matchRepo.RemoveFromQueue(ctx, userID); err != nil {
		slog.Warn("failed to remove queue entry from DB", "user", userID, "error", err)
	}
	return nil
}

// IsQueued reports whether the player is currently in the in-memory queue.
func (s *MatchmakingService) IsQueued(userID string) bool {
	return s.queue.IsQueued(userID)
}

// QueueStats returns the depth of every active matchmaking pool.
// Useful for the admin dashboard and the /health endpoint.
func (s *MatchmakingService) QueueStats() map[string]int {
	return s.queue.Stats()
}

// RestoreFromDB re-populates the in-memory queue from the Postgres
// matchmaking_queue table. Call this once at server startup so that
// players who were queued before a restart are not silently dropped.
func (s *MatchmakingService) RestoreFromDB(ctx context.Context) error {
	entries, err := s.matchRepo.GetQueueEntries(ctx)
	if err != nil {
		return fmt.Errorf("restore queue from DB: %w", err)
	}

	for _, e := range entries {
		// We do not have the username stored in matchmaking_queue, so fetch it.
		user, err := s.userRepo.GetByID(ctx, e.UserID)
		if err != nil || user == nil {
			slog.Warn("skipping queue restore for unknown user", "user_id", e.UserID)
			continue
		}

		s.queue.Push(matchmaking.Entry{
			UserID:         e.UserID,
			Username:       user.Username,
			ExamCategoryID: e.ExamCategoryID,
			MatchType:      e.MatchType,
			Rating:         e.Rating,
			QueuedAt:       e.QueuedAt,
		})
	}

	slog.Info("queue restored from DB", "entries", len(entries))
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────

// toDBQueueEntry converts the in-memory matchmaking.Entry into the
// models.QueueEntry that the repository layer expects.
// Keeping this conversion in one place means changing models.QueueEntry
// only requires an update here, not scattered across the codebase.
func toDBQueueEntry(e matchmaking.Entry) models.QueueEntry {
	return models.QueueEntry{
		UserID:         e.UserID,
		ExamCategoryID: e.ExamCategoryID,
		MatchType:      e.MatchType,
		Rating:         e.Rating,
		QueuedAt:       e.QueuedAt,
	}
}
