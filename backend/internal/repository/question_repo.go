package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/exam-arena/internal/models"
	"github.com/google/uuid"
)

type QuestionRepo struct {
	db *sql.DB
}

func NewQuestionRepo(db *sql.DB) *QuestionRepo {
	return &QuestionRepo{db: db}
}

func (r *QuestionRepo) GetRandomQuestions(ctx context.Context, categoryID string, count int) ([]models.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM questions
		WHERE exam_category_id = $1 AND status = 'published'
		ORDER BY RANDOM()
		LIMIT $2
	`, categoryID, count)
	if err != nil {
		return nil, fmt.Errorf("get random questions: %w", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(
			&q.ID, &q.ExamCategoryID, &q.TopicID, &q.QuestionType,
			&q.Difficulty, &q.Language, &q.Body, &q.Explanation,
			&q.EstimatedTimeSeconds, &q.Status, &q.CreatedAt,
		); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	for i := range questions {
		options, err := r.GetOptions(ctx, questions[i].ID)
		if err != nil {
			return nil, err
		}
		questions[i].Options = options
	}

	return questions, nil
}

func (r *QuestionRepo) GetByID(ctx context.Context, id string) (*models.Question, error) {
	q := &models.Question{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM questions WHERE id = $1
	`, id).Scan(
		&q.ID, &q.ExamCategoryID, &q.TopicID, &q.QuestionType,
		&q.Difficulty, &q.Language, &q.Body, &q.Explanation,
		&q.EstimatedTimeSeconds, &q.Status, &q.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	options, err := r.GetOptions(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	q.Options = options

	return q, nil
}

func (r *QuestionRepo) GetOptions(ctx context.Context, questionID string) ([]models.QuestionOption, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question_id, option_text, is_correct, order_index
		FROM question_options WHERE question_id = $1 ORDER BY order_index
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []models.QuestionOption
	for rows.Next() {
		var o models.QuestionOption
		var isCorrectInt int
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.OptionText, &isCorrectInt, &o.OrderIndex); err != nil {
			return nil, err
		}
		o.IsCorrect = isCorrectInt == 1
		options = append(options, o)
	}
	return options, nil
}

func (r *QuestionRepo) Create(ctx context.Context, q *models.Question, options []models.QuestionOption) (*models.Question, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if q.ID == "" {
		q.ID = uuid.New().String()
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, q.ID, q.ExamCategoryID, q.TopicID, q.QuestionType, q.Difficulty, q.Body, q.Explanation, q.EstimatedTimeSeconds, q.Status)
	if err != nil {
		return nil, err
	}

	for i, opt := range options {
		optID := uuid.New().String()
		isCorrectInt := 0
		if opt.IsCorrect {
			isCorrectInt = 1
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO question_options (id, question_id, option_text, is_correct, order_index)
			VALUES ($1, $2, $3, $4, $5)
		`, optID, q.ID, opt.OptionText, isCorrectInt, i+1)
		if err != nil {
			return nil, err
		}
		options[i].ID = optID
		options[i].QuestionID = q.ID
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	q.Options = options
	return q, nil
}

func (r *QuestionRepo) GetAllPublished(ctx context.Context, categoryID string) ([]models.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM   questions
		WHERE  exam_category_id = $1
		  AND  status = 'published'
		ORDER  BY created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get all published: %w", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(
			&q.ID, &q.ExamCategoryID, &q.TopicID, &q.QuestionType,
			&q.Difficulty, &q.Language, &q.Body, &q.Explanation,
			&q.EstimatedTimeSeconds, &q.Status, &q.CreatedAt,
		); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	if len(questions) == 0 {
		return questions, nil
	}

	// Prepare IN clause for options
	placeholders := make([]string, len(questions))
	args := make([]interface{}, len(questions))
	for i, q := range questions {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = q.ID
	}

	query := fmt.Sprintf(`
		SELECT id, question_id, option_text, is_correct, order_index
		FROM   question_options
		WHERE  question_id IN (%s)
		ORDER  BY question_id, order_index
	`, strings.Join(placeholders, ","))

	optRows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get options bulk: %w", err)
	}
	defer optRows.Close()

	qMap := make(map[string]*models.Question, len(questions))
	for i := range questions {
		qMap[questions[i].ID] = &questions[i]
	}

	for optRows.Next() {
		var o models.QuestionOption
		var isCorrectInt int
		if err := optRows.Scan(&o.ID, &o.QuestionID, &o.OptionText, &isCorrectInt, &o.OrderIndex); err != nil {
			return nil, err
		}
		o.IsCorrect = isCorrectInt == 1
		if q, ok := qMap[o.QuestionID]; ok {
			q.Options = append(q.Options, o)
		}
	}

	return questions, nil
}

func (r *QuestionRepo) GetActiveCategoryIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM exam_categories WHERE is_active = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *QuestionRepo) Publish(ctx context.Context, questionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE questions SET status = 'published', published_at = CURRENT_TIMESTAMP WHERE id = $1
	`, questionID)
	return err
}

func (r *QuestionRepo) GetQuestionCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM questions WHERE status = 'published'`).Scan(&count)
	return count, err
}

func (r *QuestionRepo) IsOptionCorrect(ctx context.Context, optionID string) (bool, error) {
	var isCorrectInt int
	err := r.db.QueryRowContext(ctx, `SELECT is_correct FROM question_options WHERE id = $1`, optionID).Scan(&isCorrectInt)
	if err != nil {
		return false, err
	}
	return isCorrectInt == 1, nil
}
