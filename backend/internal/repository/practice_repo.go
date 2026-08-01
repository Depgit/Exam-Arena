package repository

import (
	"context"
	"database/sql"

	"github.com/exam-arena/internal/models"
	"github.com/google/uuid"
)

type PracticeRepo struct {
	db *sql.DB
}

func NewPracticeRepo(db *sql.DB) *PracticeRepo {
	return &PracticeRepo{db: db}
}

type PracticeSession struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	ExamCategoryID string  `json:"exam_category_id"`
	TopicID        *string `json:"topic_id"`
	Difficulty     *string `json:"difficulty"`
	QuestionCount  int     `json:"question_count"`
	Status         string  `json:"status"`
}

func (r *PracticeRepo) CreateSession(ctx context.Context, userID, categoryID string, topicID *string, difficulty *string, questionCount int) (*PracticeSession, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO practice_sessions (id, user_id, exam_category_id, topic_id, difficulty, question_count)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, userID, categoryID, topicID, difficulty, questionCount)
	if err != nil {
		return nil, err
	}
	return r.GetSession(ctx, id)
}

func (r *PracticeRepo) AddSessionQuestion(ctx context.Context, sessionID, questionID string, orderIndex int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO practice_session_questions (session_id, question_id, order_index)
		VALUES ($1, $2, $3)
	`, sessionID, questionID, orderIndex)
	return err
}

func (r *PracticeRepo) AnswerQuestion(ctx context.Context, sessionID, questionID string, optionID *string, isCorrect bool, timeTakenMs int) error {
	isCorrectInt := 0
	if isCorrect {
		isCorrectInt = 1
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE practice_session_questions
		SET selected_option_id = $3, is_correct = $4, time_taken_ms = $5, answered_at = CURRENT_TIMESTAMP
		WHERE session_id = $1 AND question_id = $2
	`, sessionID, questionID, optionID, isCorrectInt, timeTakenMs)
	return err
}

func (r *PracticeRepo) EndSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE practice_sessions SET status = 'completed', ended_at = CURRENT_TIMESTAMP WHERE id = $1
	`, sessionID)
	return err
}

func (r *PracticeRepo) GetSession(ctx context.Context, sessionID string) (*PracticeSession, error) {
	s := &PracticeSession{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, exam_category_id, topic_id, difficulty, question_count, status
		FROM practice_sessions WHERE id = $1
	`, sessionID).Scan(
		&s.ID, &s.UserID, &s.ExamCategoryID, &s.TopicID, &s.Difficulty, &s.QuestionCount, &s.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

type PracticeQuestionResult struct {
	QuestionID string                   `json:"question_id"`
	OrderIndex int                      `json:"order_index"`
	IsCorrect  *bool                    `json:"is_correct"`
	TimeTaken  *int                     `json:"time_taken_ms"`
	Question   models.QuestionForPlayer `json:"question"`
}
