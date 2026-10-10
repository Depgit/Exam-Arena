package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (r *Store) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
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
func (r *Store) CreateGuest(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error) {
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

// EnsureBot returns the id of the bot account with this username, creating
// it if needed. Bots can't log in: their password hash is of a random
// secret that is immediately discarded.
func (r *Store) EnsureBot(ctx context.Context, username, displayName, passwordHash string) (string, error) {
	if _, err := r.db.Exec(ctx, `
		INSERT INTO users (username, email, password_hash, display_name, is_bot)
		VALUES ($1, $1 || '@bots.invalid', $2, $3, true)
		ON CONFLICT DO NOTHING
	`, username, passwordHash, displayName); err != nil {
		return "", fmt.Errorf("create bot %s: %w", username, err)
	}
	// Only ever use a row that is actually a bot: never take over a real
	// account that happens to have this username.
	var id string
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE username = $1 AND is_bot`, username).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("bot %s unavailable: %w", username, err)
	}
	return id, nil
}

// IsGuest reports whether the account is a demo (guest) account. A missing
// account counts as a guest, so a deleted user can't keep playing on an
// old token.
func (r *Store) IsGuest(ctx context.Context, userID string) (bool, error) {
	var guest bool
	err := r.db.QueryRow(ctx,
		`SELECT is_guest FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(&guest)
	if err == pgx.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("is guest: %w", err)
	}
	return guest, nil
}

// DeleteStaleGuests removes guest accounts older than maxAge that left
// nothing other players depend on. Their own data (ratings, stats, practice,
// friendships, daily attempts…) cascades away. Guests who played a match or
// filed a report are kept so opponents' history and moderation stay intact.
func (r *Store) DeleteStaleGuests(ctx context.Context, maxAge time.Duration) (int64, error) {
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

func (r *Store) GetByID(ctx context.Context, id string) (*models.User, error) {
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

// PlayerMatch is one result of a player search.
type PlayerMatch struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
}

// likeEscaper makes user input safe inside a LIKE pattern: %, _ and \ are
// matched literally instead of acting as wildcards.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// SearchPlayers finds real players whose username or display name contains
// query (case-insensitive), for "add a friend" suggestions. Exact and
// starts-with matches come first. Guests, bots, admins, suspended accounts
// and excludeID (the person searching) are left out.
func (r *Store) SearchPlayers(ctx context.Context, query, excludeID string, limit int) ([]PlayerMatch, error) {
	q := likeEscaper.Replace(strings.TrimSpace(query))
	rows, err := r.db.Query(ctx, `
		SELECT id, username, display_name
		FROM users
		WHERE deleted_at IS NULL AND status = 'active'
		  AND NOT is_guest AND NOT is_bot AND role <> 'admin'
		  AND id::text <> $2
		  AND (username ILIKE '%' || $1 || '%' ESCAPE '\'
		       OR display_name ILIKE '%' || $1 || '%' ESCAPE '\')
		ORDER BY lower(username::text) = lower($3) DESC,
		         username ILIKE $1 || '%' ESCAPE '\' DESC,
		         length(username::text), username
		LIMIT $4
	`, q, excludeID, strings.TrimSpace(query), limit)
	if err != nil {
		return nil, fmt.Errorf("search players: %w", err)
	}
	defer rows.Close()
	out := []PlayerMatch{}
	for rows.Next() {
		var m PlayerMatch
		if err := rows.Scan(&m.UserID, &m.Username, &m.DisplayName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Store) GetByUsername(ctx context.Context, username string) (*models.User, error) {
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

func (r *Store) GetByEmail(ctx context.Context, email string) (*models.User, error) {
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

func (r *Store) UpdateLastLogin(ctx context.Context, userID string, ip string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET last_login_at = now(), last_active_at = now(), last_login_ip = $2 WHERE id = $1
	`, userID, ip)
	return err
}

func (r *Store) GetStatistics(ctx context.Context, userID string) ([]models.UserStatistics, error) {
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

func (r *Store) GetRatings(ctx context.Context, userID string) ([]models.UserRating, error) {
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

func (r *Store) GetRating(ctx context.Context, userID, categoryID string) (*models.UserRating, error) {
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

func (r *Store) EnsureRating(ctx context.Context, userID, categoryID string) (*models.UserRating, error) {
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

func (r *Store) UpdateRating(ctx context.Context, userID, categoryID string, newRating, matchesPlayed int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_ratings SET rating = $3, matches_played = $4, updated_at = now()
		WHERE user_id = $1 AND exam_category_id = $2
	`, userID, categoryID, newRating, matchesPlayed)
	return err
}

func (r *Store) InsertRatingHistory(ctx context.Context, userID, categoryID, matchID string, oldRating, newRating, delta int) error {
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

func (r *Store) UpdateStatistics(ctx context.Context, userID, categoryID string, outcome MatchOutcome, questionsCorrect, questionsTotal int) error {
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
func (r *Store) TouchLastActive(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_active_at = now() WHERE id = $1`, userID)
	return err
}

// CountActiveSince counts users active since the given time. Users in
// onlineIDs are always counted: a long-open session may not have touched
// last_active_at recently.
func (r *Store) CountActiveSince(ctx context.Context, since time.Time, onlineIDs []string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE deleted_at IS NULL
		  AND (last_active_at >= $1 OR id::text = ANY($2))
	`, since, onlineIDs).Scan(&count)
	return count, err
}

// GetUserCount counts real players and, separately, demo (guest) accounts.
// Bots and admins are in neither.
func (r *Store) GetUserCount(ctx context.Context) (players, demo int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE NOT is_guest),
			COUNT(*) FILTER (WHERE is_guest)
		FROM users
		WHERE deleted_at IS NULL AND NOT is_bot AND role <> 'admin'
	`).Scan(&players, &demo)
	return players, demo, err
}
