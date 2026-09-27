-- 005: players flag faulty questions through the existing reports table
-- (target_type = 'question').

-- A player can have only one open flag per question.
CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_open_question_flag
  ON reports(reporter_id, target_id)
  WHERE target_type = 'question' AND status = 'open';

-- Admin review: flags grouped by question and status.
CREATE INDEX IF NOT EXISTS idx_reports_target ON reports(target_type, status, target_id);
