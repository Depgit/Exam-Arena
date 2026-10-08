package questions

import (
	"context"
	"fmt"
	"sort"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlagStore struct {
	db *pgxpool.Pool
}

func NewFlagStore(db *pgxpool.Pool) *FlagStore {
	return &FlagStore{db: db}
}

// CreateQuestionFlag records a flag. A second open flag by the same player
// on the same question fails with a unique violation.
func (r *FlagStore) CreateQuestionFlag(ctx context.Context, reporterID, questionID, reason string, description *string) (*models.QuestionFlag, error) {
	f := &models.QuestionFlag{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO reports (reporter_id, target_type, target_id, reason, description)
		VALUES ($1, 'question', $2, $3, $4)
		RETURNING id, target_id, reporter_id, reason, description, status, created_at, resolved_at
	`, reporterID, questionID, reason, description).Scan(
		&f.ID, &f.QuestionID, &f.ReporterID, &f.Reason, &f.Description, &f.Status, &f.CreatedAt, &f.ResolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create question flag: %w", err)
	}
	return f, nil
}

// ListFlaggedQuestions returns flagged questions with their flags, most
// flagged first. status is a report status, or "all".
func (r *FlagStore) ListFlaggedQuestions(ctx context.Context, status string) ([]models.FlaggedQuestion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.target_id, r.reporter_id, u.username, r.reason, r.description,
		       r.status, r.created_at, r.resolved_at,
		       q.body, q.status, q.explanation, c.name, t.name
		FROM reports r
		JOIN questions q       ON q.id = r.target_id
		JOIN exam_categories c ON c.id = q.exam_category_id
		JOIN topics t          ON t.id = q.topic_id
		JOIN users u           ON u.id = r.reporter_id
		WHERE r.target_type = 'question'
		  AND ($1 = 'all' OR r.status::text = $1)
		ORDER BY r.created_at DESC
	`, status)
	if err != nil {
		return nil, fmt.Errorf("list flagged questions: %w", err)
	}
	defer rows.Close()

	byQuestion := map[string]*models.FlaggedQuestion{}
	var order []string
	for rows.Next() {
		var f models.QuestionFlag
		var body, qStatus, category, topic string
		var explanation *string
		if err := rows.Scan(
			&f.ID, &f.QuestionID, &f.ReporterID, &f.Reporter, &f.Reason, &f.Description,
			&f.Status, &f.CreatedAt, &f.ResolvedAt,
			&body, &qStatus, &explanation, &category, &topic,
		); err != nil {
			return nil, err
		}
		fq, ok := byQuestion[f.QuestionID]
		if !ok {
			fq = &models.FlaggedQuestion{
				QuestionID:     f.QuestionID,
				Body:           body,
				QuestionStatus: qStatus,
				Category:       category,
				Topic:          topic,
				Explanation:    explanation,
				Options:        []models.QuestionOption{},
				Reasons:        map[string]int{},
				LatestFlagAt:   f.CreatedAt, // rows are newest first
			}
			byQuestion[f.QuestionID] = fq
			order = append(order, f.QuestionID)
		}
		fq.FlagCount++
		fq.Reasons[f.Reason]++
		fq.Flags = append(fq.Flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]models.FlaggedQuestion, 0, len(order))
	for _, id := range order {
		out = append(out, *byQuestion[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].FlagCount > out[j].FlagCount })
	return out, nil
}

// ReviewQuestionFlags closes every open flag on a question with the given
// status and returns how many were closed.
func (r *FlagStore) ReviewQuestionFlags(ctx context.Context, questionID, adminID, status string) (int, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE reports
		SET status = $3, reviewed_by = $2, resolved_at = now()
		WHERE target_type = 'question' AND target_id = $1 AND status = 'open'
	`, questionID, adminID, status)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// CountOpenFlaggedQuestions counts questions with at least one open flag.
func (r *FlagStore) CountOpenFlaggedQuestions(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT target_id) FROM reports
		WHERE target_type = 'question' AND status = 'open'
	`).Scan(&n)
	return n, err
}
