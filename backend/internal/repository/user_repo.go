package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, display_name)
		VALUES ($1, $2, $3, $4, $5)
	`, id, username, email, passwordHash, username)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       country_code, preferred_language, role, status,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.CountryCode,
		&user.PreferredLanguage, &user.Role, &user.Status,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       country_code, preferred_language, role, status,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE username = $1 AND deleted_at IS NULL
	`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.CountryCode,
		&user.PreferredLanguage, &user.Role, &user.Status,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       country_code, preferred_language, role, status,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.CountryCode,
		&user.PreferredLanguage, &user.Role, &user.Status,
		&user.EmailVerifiedAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string, ip string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET last_login_at = CURRENT_TIMESTAMP, last_login_ip = $2 WHERE id = $1
	`, userID, ip)
	return err
}

func (r *UserRepo) GetStatistics(ctx context.Context, userID string) ([]models.UserStatistics, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, exam_category_id, total_matches, wins, losses, draws,
		       current_win_streak, longest_win_streak, longest_losing_streak,
		       total_questions_solved, total_practice_sessions, overall_accuracy,
		       avg_solving_time_ms
		FROM user_statistics WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserStatistics
	for rows.Next() {
		var s models.UserStatistics
		if err := rows.Scan(
			&s.UserID, &s.ExamCategoryID, &s.TotalMatches, &s.Wins, &s.Losses, &s.Draws,
			&s.CurrentWinStreak, &s.LongestWinStreak, &s.LongestLosingStreak,
			&s.TotalQuestionsSolved, &s.TotalPracticeSessions, &s.OverallAccuracy,
			&s.AvgSolvingTimeMs,
		); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *UserRepo) GetRatings(ctx context.Context, userID string) ([]models.UserRating, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, exam_category_id, rating, matches_played, updated_at
		FROM user_ratings WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []models.UserRating
	for rows.Next() {
		var r models.UserRating
		if err := rows.Scan(&r.UserID, &r.ExamCategoryID, &r.Rating, &r.MatchesPlayed, &r.UpdatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, r)
	}
	return ratings, nil
}

func (r *UserRepo) GetRating(ctx context.Context, userID, categoryID string) (*models.UserRating, error) {
	rating := &models.UserRating{}
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, exam_category_id, rating, matches_played, updated_at
		FROM user_ratings WHERE user_id = $1 AND exam_category_id = $2
	`, userID, categoryID).Scan(
		&rating.UserID, &rating.ExamCategoryID, &rating.Rating, &rating.MatchesPlayed, &rating.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rating, nil
}

func (r *UserRepo) EnsureRating(ctx context.Context, userID, categoryID string) (*models.UserRating, error) {
	_, _ = r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO user_ratings (user_id, exam_category_id, rating, matches_played)
		VALUES ($1, $2, 1200, 0)
	`, userID, categoryID)
	return r.GetRating(ctx, userID, categoryID)
}

func (r *UserRepo) UpdateRating(ctx context.Context, userID, categoryID string, newRating, matchesPlayed int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_ratings SET rating = $3, matches_played = $4, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND exam_category_id = $2
	`, userID, categoryID, newRating, matchesPlayed)
	return err
}

func (r *UserRepo) InsertRatingHistory(ctx context.Context, userID, categoryID, matchID string, oldRating, newRating, delta int) error {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO rating_history (id, user_id, exam_category_id, match_id, old_rating, new_rating, delta)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, userID, categoryID, matchID, oldRating, newRating, delta)
	return err
}

func (r *UserRepo) UpdateStatistics(ctx context.Context, userID, categoryID string, won bool, questionsCorrect, questionsTotal int) error {
	winInc, lossInc := 0, 0
	if won {
		winInc = 1
	} else {
		lossInc = 1
	}

	accuracy := float64(0)
	if questionsTotal > 0 {
		accuracy = float64(questionsCorrect) / float64(questionsTotal) * 100
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_statistics (user_id, exam_category_id, total_matches, wins, losses,
		                              current_win_streak, longest_win_streak, total_questions_solved, overall_accuracy)
		VALUES ($1, $2, 1, $3, $4, $5, $5, $6, $7)
		ON CONFLICT (user_id, exam_category_id) DO UPDATE SET
			total_matches = user_statistics.total_matches + 1,
			wins = user_statistics.wins + $3,
			losses = user_statistics.losses + $4,
			current_win_streak = CASE WHEN $3 = 1 THEN user_statistics.current_win_streak + 1 ELSE 0 END,
			longest_win_streak = MAX(user_statistics.longest_win_streak,
				CASE WHEN $3 = 1 THEN user_statistics.current_win_streak + 1 ELSE user_statistics.longest_win_streak END),
			total_questions_solved = user_statistics.total_questions_solved + $6,
			overall_accuracy = ROUND(((user_statistics.total_questions_solved * user_statistics.overall_accuracy / 100.0 + $6) /
				(user_statistics.total_questions_solved + $8)) * 100, 2),
			updated_at = CURRENT_TIMESTAMP
	`, userID, categoryID, winInc, lossInc,
		winInc,
		questionsCorrect,
		accuracy,
		questionsTotal,
	)
	return err
}

func (r *UserRepo) GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}
