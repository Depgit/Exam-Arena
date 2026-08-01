package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/google/uuid"
)

type MatchRepo struct {
	db *sql.DB
}

func NewMatchRepo(db *sql.DB) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) CreateMatch(ctx context.Context, matchType, categoryID string, timerSeconds, maxPlayers int, roomCode *string) (*models.Match, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO matches (id, match_type, exam_category_id, timer_seconds, max_players, room_code)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, matchType, categoryID, timerSeconds, maxPlayers, roomCode)
	if err != nil {
		return nil, fmt.Errorf("create match: %w", err)
	}

	return r.GetMatch(ctx, id)
}

func (r *MatchRepo) GetMatch(ctx context.Context, matchID string) (*models.Match, error) {
	m := &models.Match{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, match_type, status, exam_category_id, room_code, tournament_id,
		       timer_seconds, max_players, created_at, started_at, ended_at
		FROM matches WHERE id = $1
	`, matchID).Scan(
		&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID, &m.RoomCode,
		&m.TournamentID, &m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt,
		&m.StartedAt, &m.EndedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *MatchRepo) GetMatchByRoomCode(ctx context.Context, roomCode string) (*models.Match, error) {
	m := &models.Match{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, match_type, status, exam_category_id, room_code, tournament_id,
		       timer_seconds, max_players, created_at, started_at, ended_at
		FROM matches WHERE room_code = $1 AND status IN ('waiting', 'starting')
	`, roomCode).Scan(
		&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID, &m.RoomCode,
		&m.TournamentID, &m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt,
		&m.StartedAt, &m.EndedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *MatchRepo) UpdateMatchStatus(ctx context.Context, matchID, status string) error {
	query := `UPDATE matches SET status = $2`
	switch status {
	case "in_progress":
		query += `, started_at = CURRENT_TIMESTAMP`
	case "completed", "abandoned":
		query += `, ended_at = CURRENT_TIMESTAMP`
	}
	query += ` WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, matchID, status)
	return err
}

func (r *MatchRepo) AddPlayer(ctx context.Context, matchID, userID string, ratingBefore *int) (*models.MatchPlayer, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO match_players (id, match_id, user_id, rating_before)
		VALUES ($1, $2, $3, $4)
	`, id, matchID, userID, ratingBefore)
	if err != nil {
		return nil, fmt.Errorf("add player: %w", err)
	}

	mp := &models.MatchPlayer{}
	err = r.db.QueryRowContext(ctx, `
		SELECT id, match_id, user_id, score, connection_status, joined_at
		FROM match_players WHERE id = $1
	`, id).Scan(&mp.ID, &mp.MatchID, &mp.UserID, &mp.Score, &mp.ConnectionStatus, &mp.JoinedAt)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

func (r *MatchRepo) GetMatchPlayers(ctx context.Context, matchID string) ([]models.MatchPlayer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT mp.id, mp.match_id, mp.user_id, u.username, mp.score,
		       mp.final_rank, mp.rating_before, mp.rating_after, mp.rating_delta,
		       mp.connection_status, mp.joined_at
		FROM match_players mp
		JOIN users u ON u.id = mp.user_id
		WHERE mp.match_id = $1
		ORDER BY mp.score DESC
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.MatchPlayer
	for rows.Next() {
		var p models.MatchPlayer
		if err := rows.Scan(
			&p.ID, &p.MatchID, &p.UserID, &p.Username, &p.Score,
			&p.FinalRank, &p.RatingBefore, &p.RatingAfter, &p.RatingDelta,
			&p.ConnectionStatus, &p.JoinedAt,
		); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

func (r *MatchRepo) GetPlayerCount(ctx context.Context, matchID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM match_players WHERE match_id = $1`, matchID).Scan(&count)
	return count, err
}

func (r *MatchRepo) AddMatchQuestions(ctx context.Context, matchID string, questionIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO match_questions (match_id, question_id, order_index) VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, qid := range questionIDs {
		if _, err := stmt.ExecContext(ctx, matchID, qid, i+1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MatchRepo) GetMatchQuestions(ctx context.Context, matchID string) ([]models.MatchQuestion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT match_id, question_id, order_index
		FROM match_questions WHERE match_id = $1 ORDER BY order_index
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.MatchQuestion
	for rows.Next() {
		var q models.MatchQuestion
		if err := rows.Scan(&q.MatchID, &q.QuestionID, &q.OrderIndex); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *MatchRepo) SaveAnswer(ctx context.Context, ans *models.MatchAnswer) error {
	id := uuid.New().String()
	isCorrectInt := 0
	if ans.IsCorrect {
		isCorrectInt = 1
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO match_answers (id, match_id, user_id, question_id, selected_option_id, is_correct, time_taken_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, ans.MatchID, ans.UserID, ans.QuestionID, ans.SelectedOptionID, isCorrectInt, ans.TimeTakenMs)
	return err
}

func (r *MatchRepo) UpdatePlayerScore(ctx context.Context, matchID, userID string, score int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE match_players SET score = $3 WHERE match_id = $1 AND user_id = $2
	`, matchID, userID, score)
	return err
}

func (r *MatchRepo) UpdatePlayerRating(ctx context.Context, matchID, userID string, ratingAfter, ratingDelta int, finalRank int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE match_players SET rating_after = $3, rating_delta = $4, final_rank = $5
		WHERE match_id = $1 AND user_id = $2
	`, matchID, userID, ratingAfter, ratingDelta, finalRank)
	return err
}

func (r *MatchRepo) GetMatchAnswers(ctx context.Context, matchID, userID string) ([]models.MatchAnswer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, match_id, user_id, question_id, selected_option_id, is_correct, time_taken_ms, answered_at
		FROM match_answers WHERE match_id = $1 AND user_id = $2
	`, matchID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []models.MatchAnswer
	for rows.Next() {
		var a models.MatchAnswer
		var isCorrectInt int
		if err := rows.Scan(&a.ID, &a.MatchID, &a.UserID, &a.QuestionID, &a.SelectedOptionID,
			&isCorrectInt, &a.TimeTakenMs, &a.AnsweredAt); err != nil {
			return nil, err
		}
		a.IsCorrect = isCorrectInt == 1
		answers = append(answers, a)
	}
	return answers, nil
}

func (r *MatchRepo) GetUserMatchHistory(ctx context.Context, userID string, limit, offset int) ([]models.Match, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.match_type, m.status, m.exam_category_id, m.room_code,
		       m.timer_seconds, m.max_players, m.created_at, m.started_at, m.ended_at
		FROM matches m
		JOIN match_players mp ON mp.match_id = m.id
		WHERE mp.user_id = $1 AND m.status = 'completed'
		ORDER BY m.ended_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		if err := rows.Scan(&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID, &m.RoomCode,
			&m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt, &m.StartedAt, &m.EndedAt); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

// --- Matchmaking Queue ---

func (r *MatchRepo) AddToQueue(ctx context.Context, entry *models.QueueEntry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO matchmaking_queue (user_id, exam_category_id, match_type, rating)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			exam_category_id = $2, match_type = $3, rating = $4, queued_at = CURRENT_TIMESTAMP
	`, entry.UserID, entry.ExamCategoryID, entry.MatchType, entry.Rating)
	return err
}

func (r *MatchRepo) RemoveFromQueue(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM matchmaking_queue WHERE user_id = $1`, userID)
	return err
}

func (r *MatchRepo) FindMatch(ctx context.Context, categoryID, matchType string, rating, ratingRange int) ([]models.QueueEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, exam_category_id, match_type, rating, queued_at
		FROM matchmaking_queue
		WHERE exam_category_id = $1 AND match_type = $2
		  AND rating BETWEEN $3 AND $4
		ORDER BY queued_at ASC
		LIMIT 2
	`, categoryID, matchType, rating-ratingRange, rating+ratingRange)
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

func (r *MatchRepo) GetQueueEntries(ctx context.Context) ([]models.QueueEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
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

func (r *MatchRepo) GetActiveMatchCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE status IN ('waiting','starting','in_progress')`).Scan(&count)
	return count, err
}

func (r *MatchRepo) GetTotalMatchCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches`).Scan(&count)
	return count, err
}
