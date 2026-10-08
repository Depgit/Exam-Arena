package practice

import (
	"context"

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

type Session struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	ExamCategoryID string  `json:"exam_category_id"`
	TopicID        *string `json:"topic_id"`
	Difficulty     *string `json:"difficulty"`
	QuestionCount  int     `json:"question_count"`
	Status         string  `json:"status"`
}

func (r *Store) CreateSession(ctx context.Context, userID, categoryID string, topicID *string, difficulty *string, questionCount int) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO practice_sessions (user_id, exam_category_id, topic_id, difficulty, question_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, exam_category_id, topic_id, difficulty, question_count, status
	`, userID, categoryID, topicID, difficulty, questionCount).Scan(
		&s.ID, &s.UserID, &s.ExamCategoryID, &s.TopicID, &s.Difficulty, &s.QuestionCount, &s.Status,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Store) AddSessionQuestion(ctx context.Context, sessionID, questionID string, orderIndex int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO practice_session_questions (session_id, question_id, order_index)
		VALUES ($1, $2, $3)
	`, sessionID, questionID, orderIndex)
	return err
}

func (r *Store) AnswerQuestion(ctx context.Context, sessionID, questionID string, optionID *string, isCorrect bool, timeTakenMs int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE practice_session_questions
		SET selected_option_id = $3, is_correct = $4, time_taken_ms = $5, answered_at = now()
		WHERE session_id = $1 AND question_id = $2
	`, sessionID, questionID, optionID, isCorrect, timeTakenMs)
	return err
}

func (r *Store) EndSession(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE practice_sessions SET status = 'completed', ended_at = now() WHERE id = $1
	`, sessionID)
	return err
}

func (r *Store) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	s := &Session{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, exam_category_id, topic_id, difficulty, question_count, status
		FROM practice_sessions WHERE id = $1
	`, sessionID).Scan(
		&s.ID, &s.UserID, &s.ExamCategoryID, &s.TopicID, &s.Difficulty, &s.QuestionCount, &s.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

type QuestionResult struct {
	QuestionID string                   `json:"question_id"`
	OrderIndex int                      `json:"order_index"`
	IsCorrect  *bool                    `json:"is_correct"`
	TimeTaken  *int                     `json:"time_taken_ms"`
	Question   models.QuestionForPlayer `json:"question"`
}
