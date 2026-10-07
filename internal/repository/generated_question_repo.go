package repository

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Storage for procedurally generated questions (internal/questiongen).
// Generated questions are ordinary `questions` rows with source='generated',
// so everything that reads questions keeps working; these methods are only
// for the pool keeper that creates, rotates and prunes them.

// GeneratedQuestion is one question ready to insert.
type GeneratedQuestion struct {
	CategoryID       string
	TopicID          string
	Difficulty       string
	Body             string
	Explanation      string
	EstimatedSeconds int
	GeneratorKey     string
	GeneratorSeed    int64
	Options          []string
	Correct          int
}

// NewUUID returns a random (v4) UUID. IDs are made client-side so a whole
// batch of questions and options goes in with two statements.
func NewUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ActiveCategoryIDByCode returns the id of an active category, or "" if
// there is none with that code.
func (r *QuestionRepo) ActiveCategoryIDByCode(ctx context.Context, code string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT id FROM exam_categories WHERE code = $1 AND is_active = true`, code).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return id, err
}

// EnsureTopic returns the id of the top-level topic with this name in the
// category, creating it if needed.
func (r *QuestionRepo) EnsureTopic(ctx context.Context, categoryID, name string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		SELECT id FROM topics
		WHERE exam_category_id = $1 AND name = $2 AND parent_topic_id IS NULL
		ORDER BY created_at LIMIT 1
	`, categoryID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return "", fmt.Errorf("find topic: %w", err)
	}
	err = r.db.QueryRow(ctx,
		`INSERT INTO topics (exam_category_id, name) VALUES ($1, $2) RETURNING id`,
		categoryID, name).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create topic: %w", err)
	}
	return id, nil
}

// GeneratedPoolState reports, per difficulty, how many generated questions
// are published, plus every published generated body (to avoid duplicates).
func (r *QuestionRepo) GeneratedPoolState(ctx context.Context, categoryID string) (map[string]int, map[string]bool, error) {
	rows, err := r.db.Query(ctx, `
		SELECT difficulty::text, body FROM questions
		WHERE exam_category_id = $1 AND source = 'generated' AND status = 'published'
	`, categoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("generated pool state: %w", err)
	}
	defer rows.Close()
	counts := map[string]int{}
	bodies := map[string]bool{}
	for rows.Next() {
		var d, body string
		if err := rows.Scan(&d, &body); err != nil {
			return nil, nil, err
		}
		counts[d]++
		bodies[body] = true
	}
	return counts, bodies, rows.Err()
}

// InsertGenerated tops one category+difficulty up to target, using qs as
// the supply, and returns how many it inserted.
//
// Several server instances can share one database (e.g. local Docker and
// Render), so the work runs under a transaction-scoped advisory lock and
// recounts inside it: a second instance either waits its turn and finds
// nothing to do, or skips. Transaction-scoped locks also work through
// Supabase's transaction-mode pooler, unlike session locks.
func (r *QuestionRepo) InsertGenerated(ctx context.Context, categoryID, difficulty string, target int, qs []GeneratedQuestion) (int, error) {
	if len(qs) == 0 {
		return 0, nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var locked bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtext('questiongen:' || $1))`,
		categoryID+":"+difficulty).Scan(&locked); err != nil {
		return 0, fmt.Errorf("pool lock: %w", err)
	}
	if !locked {
		return 0, nil // another instance is filling this pool right now
	}
	var have int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM questions
		WHERE exam_category_id = $1 AND difficulty = $2::difficulty_level
		  AND source = 'generated' AND status = 'published'
	`, categoryID, difficulty).Scan(&have); err != nil {
		return 0, fmt.Errorf("pool count: %w", err)
	}
	if need := target - have; need < len(qs) {
		if need <= 0 {
			return 0, nil
		}
		qs = qs[:need]
	}

	n := len(qs)
	ids := make([]string, n)
	cats := make([]string, n)
	topics := make([]string, n)
	diffs := make([]string, n)
	bodies := make([]string, n)
	expls := make([]string, n)
	secs := make([]int32, n)
	keys := make([]string, n)
	seeds := make([]int64, n)

	var optIDs, optQIDs, optTexts []string
	var optCorrect []bool
	var optOrder []int32

	for i, q := range qs {
		ids[i] = NewUUID()
		cats[i], topics[i], diffs[i] = q.CategoryID, q.TopicID, q.Difficulty
		bodies[i], expls[i], secs[i] = q.Body, q.Explanation, int32(q.EstimatedSeconds)
		keys[i], seeds[i] = q.GeneratorKey, q.GeneratorSeed
		for j, text := range q.Options {
			optIDs = append(optIDs, NewUUID())
			optQIDs = append(optQIDs, ids[i])
			optTexts = append(optTexts, text)
			optCorrect = append(optCorrect, j == q.Correct)
			optOrder = append(optOrder, int32(j+1))
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, language,
		                       body, explanation, estimated_time_seconds, status, published_at,
		                       source, generator_key, generator_seed)
		SELECT id, cat, topic, 'mcq_single', diff::difficulty_level, 'en',
		       body, expl, secs, 'published', now(),
		       'generated', key, seed
		FROM unnest($1::uuid[], $2::uuid[], $3::uuid[], $4::text[], $5::text[], $6::text[],
		            $7::int[], $8::text[], $9::bigint[])
		     AS t(id, cat, topic, diff, body, expl, secs, key, seed)
	`, ids, cats, topics, diffs, bodies, expls, secs, keys, seeds); err != nil {
		return 0, fmt.Errorf("insert generated questions: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO question_options (id, question_id, option_text, is_correct, order_index)
		SELECT * FROM unnest($1::uuid[], $2::uuid[], $3::text[], $4::bool[], $5::int[])
	`, optIDs, optQIDs, optTexts, optCorrect, optOrder); err != nil {
		return 0, fmt.Errorf("insert generated options: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return n, nil
}

// CountPlayers counts real players for the question cap: not guests, not
// admins, not deleted.
func (r *QuestionRepo) CountPlayers(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM users WHERE NOT is_guest AND role <> 'admin' AND deleted_at IS NULL`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count players: %w", err)
	}
	return n, nil
}

// ArchiveOldestBeyond keeps only the newest keep published questions in a
// category — hand-written or generated alike — and archives the rest.
// Archived questions leave matches, practice and new daily challenges but
// stay stored, so past matches and answers that used them remain intact
// (and an admin can republish one).
func (r *QuestionRepo) ArchiveOldestBeyond(ctx context.Context, categoryID string, keep int) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE questions SET status = 'archived', updated_at = now()
		WHERE id IN (
			SELECT id FROM questions
			WHERE exam_category_id = $1 AND status = 'published'
			ORDER BY created_at DESC, id
			OFFSET $2
		)
	`, categoryID, keep)
	if err != nil {
		return 0, fmt.Errorf("archive beyond cap: %w", err)
	}
	return tag.RowsAffected(), nil
}

// RetireOldestGenerated archives the n oldest published generated questions
// at one difficulty. Archived rows stay in place so past matches, answers,
// practice sessions and daily challenges that used them remain intact.
func (r *QuestionRepo) RetireOldestGenerated(ctx context.Context, categoryID, difficulty string, n int) (int64, error) {
	if n <= 0 {
		return 0, nil
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE questions SET status = 'archived', updated_at = now()
		WHERE id IN (
			SELECT id FROM questions
			WHERE exam_category_id = $1 AND difficulty = $2::difficulty_level
			  AND source = 'generated' AND status = 'published'
			ORDER BY created_at ASC
			LIMIT $3
		)
	`, categoryID, difficulty, n)
	if err != nil {
		return 0, fmt.Errorf("retire generated: %w", err)
	}
	return tag.RowsAffected(), nil
}

// PruneArchivedGenerated deletes archived generated questions that nothing
// refers to. daily_challenges and reports point at questions without a
// foreign key, so they are checked explicitly; the others would block the
// delete anyway, and checking them keeps it from failing half-way.
func (r *QuestionRepo) PruneArchivedGenerated(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM questions q
		WHERE q.source = 'generated' AND q.status = 'archived'
		  AND NOT EXISTS (SELECT 1 FROM match_questions mq WHERE mq.question_id = q.id)
		  AND NOT EXISTS (SELECT 1 FROM match_answers ma WHERE ma.question_id = q.id)
		  AND NOT EXISTS (SELECT 1 FROM practice_session_questions ps WHERE ps.question_id = q.id)
		  AND NOT EXISTS (SELECT 1 FROM daily_challenges dc WHERE q.id = ANY(dc.question_ids))
		  AND NOT EXISTS (SELECT 1 FROM reports rp WHERE rp.target_type = 'question' AND rp.target_id = q.id)
	`)
	if err != nil {
		return 0, fmt.Errorf("prune generated: %w", err)
	}
	return tag.RowsAffected(), nil
}
