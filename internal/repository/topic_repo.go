package repository

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TopicRepo struct {
	db *pgxpool.Pool
}

func NewTopicRepo(db *pgxpool.Pool) *TopicRepo {
	return &TopicRepo{db: db}
}

const topicColumns = `id, exam_category_id, parent_topic_id, name, created_at`

func scanTopic(row pgx.Row) (*models.Topic, error) {
	t := &models.Topic{}
	if err := row.Scan(&t.ID, &t.ExamCategoryID, &t.ParentTopicID, &t.Name, &t.CreatedAt); err != nil {
		return nil, err
	}
	return t, nil
}

// Create inserts a topic and returns it with the generated ID and timestamp.
// A missing exam category or parent topic surfaces as a foreign-key error;
// use IsForeignKeyViolation to map it to a 400.
func (r *TopicRepo) Create(ctx context.Context, t *models.Topic) (*models.Topic, error) {
	created, err := scanTopic(r.db.QueryRow(ctx, `
		INSERT INTO topics (exam_category_id, parent_topic_id, name)
		VALUES ($1, $2, $3)
		RETURNING `+topicColumns,
		t.ExamCategoryID, t.ParentTopicID, t.Name,
	))
	if err != nil {
		return nil, fmt.Errorf("create topic: %w", err)
	}
	return created, nil
}

// GetByID returns nil, nil when no topic has that ID.
func (r *TopicRepo) GetByID(ctx context.Context, id string) (*models.Topic, error) {
	t, err := scanTopic(r.db.QueryRow(ctx, `SELECT `+topicColumns+` FROM topics WHERE id = $1`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get topic: %w", err)
	}
	return t, nil
}

// ExistsByName reports whether the category already has a topic with this
// name, compared case-insensitively.
func (r *TopicRepo) ExistsByName(ctx context.Context, categoryID, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM topics WHERE exam_category_id = $1 AND lower(name) = lower($2)
		)
	`, categoryID, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("topic exists: %w", err)
	}
	return exists, nil
}

// List returns topics ordered by name. When categoryID is non-empty only
// that category's topics are returned. The result is never nil so an empty
// list serialises as [] rather than null.
func (r *TopicRepo) List(ctx context.Context, categoryID string) ([]models.Topic, error) {
	query := `SELECT ` + topicColumns + ` FROM topics`
	args := []interface{}{}
	if categoryID != "" {
		query += ` WHERE exam_category_id = $1`
		args = append(args, categoryID)
	}
	query += ` ORDER BY exam_category_id, name`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	defer rows.Close()

	topics := []models.Topic{}
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.ExamCategoryID, &t.ParentTopicID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}
