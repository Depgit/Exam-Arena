-- 006: the solo daily challenge.
--
-- daily_challenges fixes each day's question set the first time anyone asks
-- for it, so everyone gets the same questions even if new ones are
-- published later that day.
CREATE TABLE IF NOT EXISTS daily_challenges (
  challenge_date      DATE PRIMARY KEY,
  question_ids        UUID[] NOT NULL,
  time_limit_seconds  INTEGER NOT NULL,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One attempt per user per day.
CREATE TABLE IF NOT EXISTS daily_challenge_attempts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  challenge_date  DATE NOT NULL REFERENCES daily_challenges(challenge_date) ON DELETE CASCADE,
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at    TIMESTAMPTZ,
  correct         INTEGER NOT NULL DEFAULT 0,
  total           INTEGER NOT NULL DEFAULT 0,
  time_taken_ms   INTEGER,
  answers         JSONB NOT NULL DEFAULT '[]',
  UNIQUE (challenge_date, user_id)
);

CREATE INDEX IF NOT EXISTS idx_daily_attempts_board
  ON daily_challenge_attempts(challenge_date, correct DESC, time_taken_ms)
  WHERE completed_at IS NOT NULL;
