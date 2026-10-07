package repository

import (
	"context"
	"fmt"
	"time"

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
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, display_name, role, status, created_at, updated_at
	`, username, email, passwordHash, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// CreateGuest inserts a throwaway demo account. The password hash is of a
// random secret nobody knows, so the account can only be used through the
// token handed out when it was created.
func (r *UserRepo) CreateGuest(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, display_name, is_guest)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id, username, email, display_name, role, status, is_guest, created_at, updated_at
	`, username, email, passwordHash, displayName).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.Role, &user.Status, &user.IsGuest, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create guest: %w", err)
	}
	return user, nil
}

// DeleteStaleGuests removes guest accounts older than maxAge that left
// nothing other players depend on. Their own data (ratings, stats, practice,
// friendships, daily attempts…) cascades away. Guests who played a match or
// filed a report are kept so opponents' history and moderation stay intact.
func (r *UserRepo) DeleteStaleGuests(ctx context.Context, maxAge time.Duration) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM users u
		WHERE u.is_guest AND u.created_at < now() - make_interval(secs => $1)
		  AND NOT EXISTS (SELECT 1 FROM match_players mp WHERE mp.user_id = u.id)
		  AND NOT EXISTS (SELECT 1 FROM match_answers ma WHERE ma.user_id = u.id)
		  AND NOT EXISTS (SELECT 1 FROM rating_history rh WHERE rh.user_id = u.id)
		  AND NOT EXISTS (SELECT 1 FROM reports rp WHERE rp.reporter_id = u.id)
	`, maxAge.Seconds())
	if err != nil {
		return 0, fmt.Errorf("delete stale guests: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       country_code, preferred_language, role, status, is_guest,
		       email_verified_at, last_login_at, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.CountryCode,
		&user.PreferredLanguage, &user.Role, &user.Status, &user.IsGuest,
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
		UPDATE users SET last_login_at = now(), last_active_at = now(), last_login_ip = $2 WHERE id = $1
	`, userID, ip)
	return err
}

func (r *UserRepo) GetStatistics(ctx context.Context, userID string) ([]models.UserStatistics, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.user_id, s.exam_category_id, c.code, c.name,
		       s.total_matches, s.wins, s.losses, s.draws,
		       s.current_win_streak, s.longest_win_streak, s.longest_losing_streak,
		       s.total_questions_solved, s.total_practice_sessions, s.overall_accuracy,
		       s.avg_solving_time_ms
		FROM user_statistics s
		JOIN exam_categories c ON c.id = s.exam_category_id
		WHERE s.user_id = $1
		ORDER BY c.sort_order, c.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.UserStatistics{}
	for rows.Next() {
		var s models.UserStatistics
		if err := rows.Scan(
			&s.UserID, &s.ExamCategoryID, &s.ExamCategoryCode, &s.ExamCategoryName, &s.TotalMatches, &s.Wins, &s.Losses, &s.Draws,
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
		SELECT r.user_id, r.exam_category_id, c.code, c.name, r.rating, r.matches_played, r.updated_at
		FROM user_ratings r
		JOIN exam_categories c ON c.id = r.exam_category_id
		WHERE r.user_id = $1
		ORDER BY c.sort_order, c.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := []models.UserRating{}
	for rows.Next() {
		var r models.UserRating
		if err := rows.Scan(&r.UserID, &r.ExamCategoryID, &r.ExamCategoryCode, &r.ExamCategoryName, &r.Rating, &r.MatchesPlayed, &r.UpdatedAt); err != nil {
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

// MatchOutcome is one player's result in a finished match.
type MatchOutcome int

const (
	OutcomeLoss MatchOutcome = iota
	OutcomeDraw
	OutcomeWin
)

func (r *UserRepo) UpdateStatistics(ctx context.Context, userID, categoryID string, outcome MatchOutcome, questionsCorrect, questionsTotal int) error {
	winInc, lossInc, drawInc := 0, 0, 0
	switch outcome {
	case OutcomeWin:
		winInc = 1
	case OutcomeLoss:
		lossInc = 1
	case OutcomeDraw:
		drawInc = 1
	}

	// A draw ends the current win streak but is not counted as a loss.
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_statistics (user_id, exam_category_id, total_matches, wins, losses, draws,
		                              current_win_streak, longest_win_streak, total_questions_solved, overall_accuracy)
		VALUES ($1, $2, 1, $3, $4, $9, $5, $5, $6, $7)
		ON CONFLICT (user_id, exam_category_id) DO UPDATE SET
			total_matches = user_statistics.total_matches + 1,
			wins = user_statistics.wins + $3,
			losses = user_statistics.losses + $4,
			draws = user_statistics.draws + $9,
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
		drawInc,
	)
	return err
}

// TouchLastActive records that the user is using the app right now.
func (r *UserRepo) TouchLastActive(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_active_at = now() WHERE id = $1`, userID)
	return err
}

// CountActiveSince counts users active since the given time. Users in
// onlineIDs are always counted: a long-open session may not have touched
// last_active_at recently.
func (r *UserRepo) CountActiveSince(ctx context.Context, since time.Time, onlineIDs []string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE deleted_at IS NULL
		  AND (last_active_at >= $1 OR id::text = ANY($2))
	`, since, onlineIDs).Scan(&count)
	return count, err
}

func (r *UserRepo) GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}
