-- 003: admin-controlled category order, and the subject categories.
--
-- Categories are listed by sort_order, then name. Mathematics, Logical
-- Reasoning and English lead; every existing category keeps the default 100
-- and so follows them alphabetically.
ALTER TABLE exam_categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 100;

INSERT INTO exam_categories (code, name, description, sort_order) VALUES
  ('MATH',      'Mathematics',       'Arithmetic, algebra, geometry and quantitative aptitude.', 1),
  ('REASONING', 'Logical Reasoning', 'Verbal and non-verbal logical reasoning.',                 2),
  ('ENGLISH',   'English',           'Grammar, vocabulary and reading comprehension.',          3)
ON CONFLICT (code) DO UPDATE SET sort_order = EXCLUDED.sort_order;
