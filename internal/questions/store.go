package questions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool

	// Active categories are read on almost every page but change only when
	// an admin reorders them (or a migration toggles one), so keep them for
	// a short while instead of a database round trip per request.
	catMu      sync.Mutex
	catCached  []models.ExamCategory
	catFetched time.Time
}

const categoryCacheTTL = time.Minute

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// GetRandomQuestions picks count random published questions for practice.
// difficulty ("easy" | "medium" | "hard") narrows the pick; nil or "" means
// mixed.
func (r *Store) GetRandomQuestions(ctx context.Context, categoryID string, difficulty *string, count int) ([]models.Question, error) {
	diff := ""
	if difficulty != nil {
		diff = *difficulty
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM questions
		WHERE exam_category_id = $1 AND status = 'published'
		  AND ($2 = '' OR difficulty::text = $2)
		ORDER BY RANDOM()
		LIMIT $3
	`, categoryID, diff, count)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return questions, r.attachOptions(ctx, questions)
}

// attachOptions loads the options for every question in one query (no N+1).
func (r *Store) attachOptions(ctx context.Context, questions []models.Question) error {
	if len(questions) == 0 {
		return nil
	}
	qIDs := make([]string, len(questions))
	qMap := make(map[string]*models.Question, len(questions))
	for i := range questions {
		qIDs[i] = questions[i].ID
		qMap[questions[i].ID] = &questions[i]
	}

	optRows, err := r.db.Query(ctx, `
		SELECT id, question_id, option_text, is_correct, order_index
		FROM   question_options
		WHERE  question_id = ANY($1)
		ORDER  BY question_id, order_index
	`, qIDs)
	if err != nil {
		return fmt.Errorf("get options bulk: %w", err)
	}
	defer optRows.Close()

	for optRows.Next() {
		var o models.QuestionOption
		if err := optRows.Scan(&o.ID, &o.QuestionID, &o.OptionText, &o.IsCorrect, &o.OrderIndex); err != nil {
			return err
		}
		if q, ok := qMap[o.QuestionID]; ok {
			q.Options = append(q.Options, o)
		}
	}
	return optRows.Err()
}

func (r *Store) GetByID(ctx context.Context, id string) (*models.Question, error) {
	q := &models.Question{}
	err := r.db.QueryRow(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM questions WHERE id = $1
	`, id).Scan(
		&q.ID, &q.ExamCategoryID, &q.TopicID, &q.QuestionType,
		&q.Difficulty, &q.Language, &q.Body, &q.Explanation,
		&q.EstimatedTimeSeconds, &q.Status, &q.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
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

func (r *Store) GetOptions(ctx context.Context, questionID string) ([]models.QuestionOption, error) {
	rows, err := r.db.Query(ctx, `
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
		if err := rows.Scan(&o.ID, &o.QuestionID, &o.OptionText, &o.IsCorrect, &o.OrderIndex); err != nil {
			return nil, err
		}
		options = append(options, o)
	}
	return options, nil
}

func (r *Store) Create(ctx context.Context, q *models.Question, options []models.QuestionOption) (*models.Question, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO questions (exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`, q.ExamCategoryID, q.TopicID, q.QuestionType, q.Difficulty, q.Body, q.Explanation, q.EstimatedTimeSeconds, q.Status,
	).Scan(&q.ID, &q.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i, opt := range options {
		err = tx.QueryRow(ctx, `
			INSERT INTO question_options (question_id, option_text, is_correct, order_index)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, q.ID, opt.OptionText, opt.IsCorrect, i+1).Scan(&options[i].ID)
		if err != nil {
			return nil, err
		}
		options[i].QuestionID = q.ID
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	q.Options = options
	return q, nil
}

// Add these two methods to the existing Store struct.
// The rest of the file is unchanged.

// GetAllPublished returns every published question (with options) for one category.
// Called by QuestionBank.Warm() at startup and on each refresh tick.
func (r *Store) GetAllPublished(ctx context.Context, categoryID string) ([]models.Question, error) {
	rows, err := r.db.Query(ctx, `
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

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return questions, r.attachOptions(ctx, questions)
}

// GetActiveCategoryIDs returns the UUIDs of all active exam categories.
// Used by QuestionBank to know which categories to warm.
func (r *Store) GetActiveCategoryIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id FROM exam_categories WHERE is_active = true
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

func (r *Store) GetActiveCategories(ctx context.Context) ([]models.ExamCategory, error) {
	r.catMu.Lock()
	if r.catCached != nil && time.Since(r.catFetched) < categoryCacheTTL {
		out := append([]models.ExamCategory(nil), r.catCached...)
		r.catMu.Unlock()
		return out, nil
	}
	r.catMu.Unlock()

	categories, err := r.loadActiveCategories(ctx)
	if err != nil {
		return nil, err
	}
	r.catMu.Lock()
	r.catCached, r.catFetched = categories, time.Now()
	r.catMu.Unlock()
	return append([]models.ExamCategory(nil), categories...), nil
}

func (r *Store) invalidateCategories() {
	r.catMu.Lock()
	r.catCached = nil
	r.catMu.Unlock()
}

func (r *Store) loadActiveCategories(ctx context.Context) ([]models.ExamCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, name, COALESCE(description, ''), is_active, sort_order
		FROM exam_categories WHERE is_active = true
		ORDER BY sort_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ExamCategory
	for rows.Next() {
		var cat models.ExamCategory
		if err := rows.Scan(&cat.ID, &cat.Code, &cat.Name, &cat.Description, &cat.IsActive, &cat.SortOrder); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// GetActiveCategory returns one active category, or nil if the id is unknown
// or the category is inactive.
func (r *Store) GetActiveCategory(ctx context.Context, id string) (*models.ExamCategory, error) {
	cat := &models.ExamCategory{IsActive: true}
	err := r.db.QueryRow(ctx, `
		SELECT id, code, name FROM exam_categories WHERE id = $1 AND is_active = true
	`, id).Scan(&cat.ID, &cat.Code, &cat.Name)
	if err != nil {
		if err == pgx.ErrNoRows || database.IsInvalidInput(err) {
			return nil, nil
		}
		return nil, err
	}
	return cat, nil
}

// SetCategorySortOrder changes where a category appears in lists. It
// reports false when the category does not exist.
func (r *Store) SetCategorySortOrder(ctx context.Context, id string, sortOrder int) (bool, error) {
	tag, err := r.db.Exec(ctx, `UPDATE exam_categories SET sort_order = $2 WHERE id = $1`, id, sortOrder)
	r.invalidateCategories()
	if err != nil {
		if database.IsInvalidInput(err) {
			return false, nil
		}
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// GetByIDsWithOptions loads questions with their options (including
// is_correct), returned in the order of ids. Unknown ids are skipped.
func (r *Store) GetByIDsWithOptions(ctx context.Context, ids []string) ([]models.Question, error) {
	if len(ids) == 0 {
		return []models.Question{}, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, exam_category_id, topic_id, question_type, difficulty,
		       language, body, explanation, estimated_time_seconds, status, created_at
		FROM questions WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("get questions by ids: %w", err)
	}
	byID := make(map[string]*models.Question, len(ids))
	for rows.Next() {
		q := &models.Question{}
		if err := rows.Scan(
			&q.ID, &q.ExamCategoryID, &q.TopicID, &q.QuestionType,
			&q.Difficulty, &q.Language, &q.Body, &q.Explanation,
			&q.EstimatedTimeSeconds, &q.Status, &q.CreatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		byID[q.ID] = q
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	optRows, err := r.db.Query(ctx, `
		SELECT id, question_id, option_text, is_correct, order_index
		FROM question_options WHERE question_id = ANY($1)
		ORDER BY question_id, order_index
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("get options by question ids: %w", err)
	}
	defer optRows.Close()
	for optRows.Next() {
		var o models.QuestionOption
		if err := optRows.Scan(&o.ID, &o.QuestionID, &o.OptionText, &o.IsCorrect, &o.OrderIndex); err != nil {
			return nil, err
		}
		if q, ok := byID[o.QuestionID]; ok {
			q.Options = append(q.Options, o)
		}
	}

	out := make([]models.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := byID[id]; ok {
			out = append(out, *q)
		}
	}
	return out, nil
}

// PickDailyQuestionIDs chooses n published questions from active categories
// in a pseudo-random order that is fixed for a given seed (the date).
func (r *Store) PickDailyQuestionIDs(ctx context.Context, seed string, n int) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT q.id
		FROM questions q
		JOIN exam_categories c ON c.id = q.exam_category_id
		WHERE q.status = 'published' AND c.is_active = true
		ORDER BY md5(q.id::text || $1)
		LIMIT $2
	`, seed, n)
	if err != nil {
		return nil, fmt.Errorf("pick daily questions: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Archive takes a question out of play.
func (r *Store) Archive(ctx context.Context, questionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE questions SET status = 'archived', updated_at = now() WHERE id = $1
	`, questionID)
	return err
}

func (r *Store) Publish(ctx context.Context, questionID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE questions SET status = 'published', published_at = now() WHERE id = $1
	`, questionID)
	return err
}

func (r *Store) GetQuestionCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM questions WHERE status = 'published'`).Scan(&count)
	return count, err
}

func (r *Store) IsOptionCorrect(ctx context.Context, optionID string) (bool, error) {
	var isCorrect bool
	err := r.db.QueryRow(ctx, `SELECT is_correct FROM question_options WHERE id = $1`, optionID).Scan(&isCorrect)
	if err != nil {
		return false, err
	}
	return isCorrect, nil
}
