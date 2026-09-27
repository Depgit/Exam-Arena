package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DailyRepo struct {
	db *pgxpool.Pool
}

func NewDailyRepo(db *pgxpool.Pool) *DailyRepo {
	return &DailyRepo{db: db}
}

func (r *DailyRepo) GetChallenge(ctx context.Context, date time.Time) (*models.DailyChallenge, error) {
	c := &models.DailyChallenge{}
	err := r.db.QueryRow(ctx, `
		SELECT challenge_date, question_ids::text[], time_limit_seconds
		FROM daily_challenges WHERE challenge_date = $1
	`, date).Scan(&c.Date, &c.QuestionIDs, &c.TimeLimitSeconds)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get daily challenge: %w", err)
	}
	return c, nil
}

// CreateChallenge stores the day's question set unless another request
// already did (first writer wins).
func (r *DailyRepo) CreateChallenge(ctx context.Context, date time.Time, questionIDs []string, timeLimitSeconds int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO daily_challenges (challenge_date, question_ids, time_limit_seconds)
		VALUES ($1, $2::uuid[], $3)
		ON CONFLICT (challenge_date) DO NOTHING
	`, date, questionIDs, timeLimitSeconds)
	return err
}

const attemptColumns = `id, challenge_date, user_id, started_at, completed_at, correct, total, time_taken_ms, answers`

func scanAttempt(row pgx.Row) (*models.DailyAttempt, error) {
	a := &models.DailyAttempt{}
	var answers []byte
	err := row.Scan(&a.ID, &a.Date, &a.UserID, &a.StartedAt, &a.CompletedAt, &a.Correct, &a.Total, &a.TimeTakenMs, &answers)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = json.Unmarshal(answers, &a.Answers)
	return a, nil
}

func (r *DailyRepo) GetAttempt(ctx context.Context, date time.Time, userID string) (*models.DailyAttempt, error) {
	a, err := scanAttempt(r.db.QueryRow(ctx, `
		SELECT `+attemptColumns+` FROM daily_challenge_attempts
		WHERE challenge_date = $1 AND user_id = $2
	`, date, userID))
	if err != nil {
		return nil, fmt.Errorf("get daily attempt: %w", err)
	}
	return a, nil
}

// StartAttempt creates the user's attempt for the day, or returns the
// existing one.
func (r *DailyRepo) StartAttempt(ctx context.Context, date time.Time, userID string, total int) (*models.DailyAttempt, error) {
	a, err := scanAttempt(r.db.QueryRow(ctx, `
		INSERT INTO daily_challenge_attempts (challenge_date, user_id, total)
		VALUES ($1, $2, $3)
		ON CONFLICT (challenge_date, user_id) DO NOTHING
		RETURNING `+attemptColumns,
		date, userID, total))
	if err != nil {
		return nil, fmt.Errorf("start daily attempt: %w", err)
	}
	if a == nil {
		return r.GetAttempt(ctx, date, userID)
	}
	return a, nil
}

// CompleteAttempt records the graded result. It reports false if the
// attempt was already completed.
func (r *DailyRepo) CompleteAttempt(ctx context.Context, attemptID string, correct, total, timeTakenMs int, answers []models.DailyAnswerRecord) (bool, error) {
	raw, err := json.Marshal(answers)
	if err != nil {
		return false, err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE daily_challenge_attempts
		SET completed_at = now(), correct = $2, total = $3, time_taken_ms = $4, answers = $5
		WHERE id = $1 AND completed_at IS NULL
	`, attemptID, correct, total, timeTakenMs, raw)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// Leaderboard ranks completed attempts: most correct, then fastest.
func (r *DailyRepo) Leaderboard(ctx context.Context, date time.Time, limit int) ([]models.DailyEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT RANK() OVER (ORDER BY a.correct DESC, a.time_taken_ms ASC),
		       u.id, u.username, u.display_name, a.correct, a.total, a.time_taken_ms
		FROM daily_challenge_attempts a
		JOIN users u ON u.id = a.user_id
		WHERE a.challenge_date = $1 AND a.completed_at IS NOT NULL
		ORDER BY a.correct DESC, a.time_taken_ms ASC, a.completed_at ASC
		LIMIT $2
	`, date, limit)
	if err != nil {
		return nil, fmt.Errorf("daily leaderboard: %w", err)
	}
	defer rows.Close()
	entries := []models.DailyEntry{}
	for rows.Next() {
		var e models.DailyEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.DisplayName, &e.Correct, &e.Total, &e.TimeTakenMs); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Standing returns a completed attempt's rank (ties share a rank) and the
// number of players who have completed the day's challenge.
func (r *DailyRepo) Standing(ctx context.Context, date time.Time, correct, timeTakenMs int) (rank, participants int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
		  1 + COUNT(*) FILTER (WHERE correct > $2 OR (correct = $2 AND time_taken_ms < $3)),
		  COUNT(*)
		FROM daily_challenge_attempts
		WHERE challenge_date = $1 AND completed_at IS NOT NULL
	`, date, correct, timeTakenMs).Scan(&rank, &participants)
	return rank, participants, err
}

func (r *DailyRepo) CountParticipants(ctx context.Context, date time.Time) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM daily_challenge_attempts
		WHERE challenge_date = $1 AND completed_at IS NOT NULL
	`, date).Scan(&n)
	return n, err
}
