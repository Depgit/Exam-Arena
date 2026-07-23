package repository

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, display_name)
		VALUES ($1, $2, $3, $1)
		RETURNING id, username, email, display_name, role, status, created_at, updated_at
	`, username, email, passwordHash).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string, ip string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET last_login_at = now(), last_login_ip = $2 WHERE id = $1
	`, userID, ip)
	return err
}

func (r *UserRepo) GetStatistics(ctx context.Context, userID string) ([]models.UserStatistics, error) {
	rows, err := r.db.Query(ctx, `
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
	rows, err := r.db.Query(ctx, `
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
	err := r.db.QueryRow(ctx, `
		SELECT user_id, exam_category_id, rating, matches_played, updated_at
		FROM user_ratings WHERE user_id = $1 AND exam_category_id = $2
	`, userID, categoryID).Scan(
		&rating.UserID, &rating.ExamCategoryID, &rating.Rating, &rating.MatchesPlayed, &rating.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rating, nil
}

func (r *UserRepo) EnsureRating(ctx context.Context, userID, categoryID string) (*models.UserRating, error) {
	rating := &models.UserRating{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_ratings (user_id, exam_category_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, exam_category_id) DO NOTHING
		RETURNING user_id, exam_category_id, rating, matches_played, updated_at
	`, userID, categoryID).Scan(
		&rating.UserID, &rating.ExamCategoryID, &rating.Rating, &rating.MatchesPlayed, &rating.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return r.GetRating(ctx, userID, categoryID)
		}
		return nil, err
	}
	return rating, nil
}

func (r *UserRepo) UpdateRating(ctx context.Context, userID, categoryID string, newRating, matchesPlayed int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_ratings SET rating = $3, matches_played = $4, updated_at = now()
		WHERE user_id = $1 AND exam_category_id = $2
	`, userID, categoryID, newRating, matchesPlayed)
	return err
}

func (r *UserRepo) InsertRatingHistory(ctx context.Context, userID, categoryID, matchID string, oldRating, newRating, delta int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO rating_history (user_id, exam_category_id, match_id, old_rating, new_rating, delta)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, categoryID, matchID, oldRating, newRating, delta)
	return err
}

func (r *UserRepo) UpdateStatistics(ctx context.Context, userID, categoryID string, won bool, questionsCorrect, questionsTotal int) error {
	winInc, lossInc := 0, 0
	if won {
		winInc = 1
	} else {
		lossInc = 1
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO user_statistics (user_id, exam_category_id, total_matches, wins, losses,
		                              current_win_streak, longest_win_streak, total_questions_solved, overall_accuracy)
		VALUES ($1, $2, 1, $3, $4, $5, $5, $6, $7)
		ON CONFLICT (user_id, exam_category_id) DO UPDATE SET
			total_matches = user_statistics.total_matches + 1,
			wins = user_statistics.wins + $3,
			losses = user_statistics.losses + $4,
			current_win_streak = CASE WHEN $3 = 1 THEN user_statistics.current_win_streak + 1 ELSE 0 END,
			longest_win_streak = GREATEST(user_statistics.longest_win_streak,
				CASE WHEN $3 = 1 THEN user_statistics.current_win_streak + 1 ELSE user_statistics.longest_win_streak END),
			longest_losing_streak = GREATEST(user_statistics.longest_losing_streak,
				CASE WHEN $4 = 1 THEN
					CASE WHEN user_statistics.current_win_streak = 0 THEN user_statistics.losses - user_statistics.wins + 1 ELSE 1 END
				ELSE user_statistics.longest_losing_streak END),
			total_questions_solved = user_statistics.total_questions_solved + $6,
			overall_accuracy = CASE WHEN (user_statistics.total_questions_solved + $8) > 0
				THEN ROUND(((user_statistics.total_questions_solved * user_statistics.overall_accuracy / 100.0 + $6)::NUMERIC /
					(user_statistics.total_questions_solved + $8)::NUMERIC) * 100, 2)
				ELSE 0 END,
			updated_at = now()
	`, userID, categoryID, winInc, lossInc,
		winInc, // current_win_streak init
		questionsCorrect,
		func() float64 {
			if questionsTotal == 0 {
				return 0
			}
			return float64(questionsCorrect) / float64(questionsTotal) * 100
		}(),
		questionsTotal,
	)
	return err
}

func (r *UserRepo) GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}
