package matchmaking

import (
	"context"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store keeps a copy of the in-memory queue in Postgres (matchmaking_queue).
//
// The in-memory Queue is the real one; this copy exists only so players
// who were waiting are put back in the queue after a server restart
// (Service.RestoreFromDB).
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// AddToQueue records (or updates) a waiting player.
func (r *Store) AddToQueue(ctx context.Context, entry *models.QueueEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO matchmaking_queue (user_id, exam_category_id, match_type, rating)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			exam_category_id = $2, match_type = $3, rating = $4, queued_at = now()
	`, entry.UserID, entry.ExamCategoryID, entry.MatchType, entry.Rating)
	return err
}

// RemoveFromQueue forgets a player who left the queue or got matched.
func (r *Store) RemoveFromQueue(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM matchmaking_queue WHERE user_id = $1`, userID)
	return err
}

// GetQueueEntries returns every waiting player, oldest first.
func (r *Store) GetQueueEntries(ctx context.Context) ([]models.QueueEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, exam_category_id, match_type, rating, queued_at
		FROM matchmaking_queue ORDER BY queued_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.QueueEntry
	for rows.Next() {
		var e models.QueueEntry
		if err := rows.Scan(&e.UserID, &e.ExamCategoryID, &e.MatchType, &e.Rating, &e.QueuedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
