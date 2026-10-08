package leaderboard

import (
	"context"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (r *Store) GetLeaderboard(ctx context.Context, categoryID string, limit, offset int) ([]models.LeaderboardEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ur.user_id, u.username, u.display_name, u.avatar_url,
		       ur.rating, ur.matches_played,
		       ROW_NUMBER() OVER (ORDER BY ur.rating DESC) AS rank
		FROM user_ratings ur
		JOIN users u ON u.id = ur.user_id AND u.deleted_at IS NULL AND NOT u.is_guest AND NOT u.is_bot
		WHERE ur.exam_category_id = $1
		ORDER BY ur.rating DESC
		LIMIT $2 OFFSET $3
	`, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.LeaderboardEntry{}
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(
			&e.UserID, &e.Username, &e.DisplayName, &e.AvatarURL,
			&e.Rating, &e.MatchesPlayed, &e.Rank,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *Store) GetCategoryByCode(ctx context.Context, code string) (*models.ExamCategory, error) {
	cat := &models.ExamCategory{}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name, description, is_active FROM exam_categories WHERE code = $1
	`, code).Scan(&cat.ID, &cat.Code, &cat.Name, &cat.Description, &cat.IsActive)
	if err != nil {
		return nil, err
	}
	return cat, nil
}
