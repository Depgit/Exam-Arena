-- 007: procedurally generated questions (internal/questiongen).
--
-- Generated questions are ordinary rows in `questions`, so matches, practice,
-- the daily challenge, saved answers and flags all keep working unchanged.
-- These columns record where a row came from so the pool keeper can rotate
-- generated rows without ever touching hand-written ones, and so any
-- generated question can be rebuilt exactly from (generator_key, seed).

ALTER TABLE questions
  ADD COLUMN IF NOT EXISTS source         VARCHAR(16) NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS generator_key  VARCHAR(64),
  ADD COLUMN IF NOT EXISTS generator_seed BIGINT;

-- Pool keeper lookups: "published generated questions in this category".
CREATE INDEX IF NOT EXISTS idx_questions_generated
  ON questions(exam_category_id, status, created_at)
  WHERE source = 'generated';

-- English is paused for now. Deactivated, not deleted: its past matches,
-- ratings and stats stay intact, and it comes back with
--   UPDATE exam_categories SET is_active = true WHERE code = 'ENGLISH';
UPDATE exam_categories SET is_active = false WHERE code = 'ENGLISH';
