package repository

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepo struct {
	db *pgxpool.Pool
}

func NewMatchRepo(db *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) CreateMatch(ctx context.Context, matchType, categoryID string, timerSeconds, maxPlayers int, roomCode *string) (*models.Match, error) {
	m := &models.Match{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO matches (match_type, exam_category_id, timer_seconds, max_players, room_code)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, match_type, status, exam_category_id, room_code, timer_seconds, max_players, created_at
	`, matchType, categoryID, timerSeconds, maxPlayers, roomCode).Scan(
		&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID,
		&m.RoomCode, &m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create match: %w", err)
	}
	return m, nil
}

func (r *MatchRepo) GetMatch(ctx context.Context, matchID string) (*models.Match, error) {
	m := &models.Match{}
	err := r.db.QueryRow(ctx, `
		SELECT id, match_type, status, exam_category_id, room_code, tournament_id,
		       timer_seconds, max_players, created_at, started_at, ended_at
		FROM matches WHERE id = $1
	`, matchID).Scan(
		&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID, &m.RoomCode,
		&m.TournamentID, &m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt,
		&m.StartedAt, &m.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *MatchRepo) GetMatchByRoomCode(ctx context.Context, roomCode string) (*models.Match, error) {
	m := &models.Match{}
	err := r.db.QueryRow(ctx, `
		SELECT id, match_type, status, exam_category_id, room_code, tournament_id,
		       timer_seconds, max_players, created_at, started_at, ended_at
		FROM matches WHERE room_code = $1 AND status IN ('waiting', 'starting')
	`, roomCode).Scan(
		&m.ID, &m.MatchType, &m.Status, &m.ExamCategoryID, &m.RoomCode,
		&m.TournamentID, &m.TimerSeconds, &m.MaxPlayers, &m.CreatedAt,
		&m.StartedAt, &m.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
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
		query += `, started_at = now()`
	case "completed", "abandoned":
		query += `, ended_at = now()`
	}
	query += ` WHERE id = $1`
	_, err := r.db.Exec(ctx, query, matchID, status)
	return err
}

// CancelWaitingMatch cancels a match that has not started yet. It reports
// false when the match is no longer waiting (it started or already ended).
func (r *MatchRepo) CancelWaitingMatch(ctx context.Context, matchID string) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE matches SET status = 'cancelled', ended_at = now()
		WHERE id = $1 AND status = 'waiting'
	`, matchID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *MatchRepo) AddPlayer(ctx context.Context, matchID, userID string, ratingBefore *int) (*models.MatchPlayer, error) {
	mp := &models.MatchPlayer{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO match_players (match_id, user_id, rating_before)
		VALUES ($1, $2, $3)
		RETURNING id, match_id, user_id, score, connection_status, joined_at
	`, matchID, userID, ratingBefore).Scan(
		&mp.ID, &mp.MatchID, &mp.UserID, &mp.Score, &mp.ConnectionStatus, &mp.JoinedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add player: %w", err)
	}
	return mp, nil
}

func (r *MatchRepo) GetMatchPlayers(ctx context.Context, matchID string) ([]models.MatchPlayer, error) {
	rows, err := r.db.Query(ctx, `
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
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM match_players WHERE match_id = $1`, matchID).Scan(&count)
	return count, err
}

func (r *MatchRepo) AddMatchQuestions(ctx context.Context, matchID string, questionIDs []string) error {
	batch := &pgx.Batch{}
	for i, qid := range questionIDs {
		batch.Queue(`INSERT INTO match_questions (match_id, question_id, order_index) VALUES ($1, $2, $3)`,
			matchID, qid, i+1)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range questionIDs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *MatchRepo) GetMatchQuestions(ctx context.Context, matchID string) ([]models.MatchQuestion, error) {
	rows, err := r.db.Query(ctx, `
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
	_, err := r.db.Exec(ctx, `
		INSERT INTO match_answers (match_id, user_id, question_id, selected_option_id, is_correct, time_taken_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (match_id, user_id, question_id) DO NOTHING
	`, ans.MatchID, ans.UserID, ans.QuestionID, ans.SelectedOptionID, ans.IsCorrect, ans.TimeTakenMs)
	return err
}

func (r *MatchRepo) UpdatePlayerScore(ctx context.Context, matchID, userID string, score int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE match_players SET score = $3 WHERE match_id = $1 AND user_id = $2
	`, matchID, userID, score)
	return err
}

func (r *MatchRepo) UpdatePlayerRating(ctx context.Context, matchID, userID string, ratingAfter, ratingDelta int, finalRank int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE match_players SET rating_after = $3, rating_delta = $4, final_rank = $5
		WHERE match_id = $1 AND user_id = $2
	`, matchID, userID, ratingAfter, ratingDelta, finalRank)
	return err
}

func (r *MatchRepo) GetMatchAnswers(ctx context.Context, matchID, userID string) ([]models.MatchAnswer, error) {
	rows, err := r.db.Query(ctx, `
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
		if err := rows.Scan(&a.ID, &a.MatchID, &a.UserID, &a.QuestionID, &a.SelectedOptionID,
			&a.IsCorrect, &a.TimeTakenMs, &a.AnsweredAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, nil
}

func (r *MatchRepo) GetUserMatchHistory(ctx context.Context, userID string, limit, offset int) ([]models.Match, error) {
	rows, err := r.db.Query(ctx, `
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
	_, err := r.db.Exec(ctx, `
		INSERT INTO matchmaking_queue (user_id, exam_category_id, match_type, rating)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			exam_category_id = $2, match_type = $3, rating = $4, queued_at = now()
	`, entry.UserID, entry.ExamCategoryID, entry.MatchType, entry.Rating)
	return err
}

func (r *MatchRepo) RemoveFromQueue(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM matchmaking_queue WHERE user_id = $1`, userID)
	return err
}

func (r *MatchRepo) FindMatch(ctx context.Context, categoryID, matchType string, rating, ratingRange int) ([]models.QueueEntry, error) {
	rows, err := r.db.Query(ctx, `
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

func (r *MatchRepo) GetActiveMatchCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM matches WHERE status IN ('waiting','starting','in_progress')`).Scan(&count)
	return count, err
}

func (r *MatchRepo) GetTotalMatchCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM matches`).Scan(&count)
	return count, err
}
