

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

-- All the CREATE TYPE, CREATE TABLE, CREATE INDEX statements from the schema...
-- (Paste the entire schema SQL verbatim here)

DO $$ BEGIN
  CREATE TYPE user_role AS ENUM ('user', 'admin', 'moderator');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE account_status AS ENUM ('active', 'suspended', 'banned', 'deleted');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE question_type AS ENUM ('mcq_single', 'mcq_multiple', 'integer');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE question_status AS ENUM ('draft', 'in_review', 'published', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE match_type AS ENUM ('ranked', 'friend', 'arena', 'tournament', 'daily_challenge');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE match_status AS ENUM ('waiting', 'starting', 'in_progress', 'completed', 'abandoned', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE player_conn_status AS ENUM ('connected', 'disconnected', 'reconnecting', 'left');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE tournament_status AS ENUM ('scheduled', 'registration_open', 'in_progress', 'completed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE friend_req_status AS ENUM ('pending', 'accepted', 'declined', 'blocked');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE report_status AS ENUM ('open', 'reviewed', 'resolved', 'dismissed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE report_target AS ENUM ('question', 'player', 'content');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE notification_type AS ENUM (
    'friend_invite', 'tournament_start', 'match_found',
    'achievement_unlocked', 'daily_challenge', 'system_announcement'
  );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username          CITEXT UNIQUE NOT NULL,
  email             CITEXT UNIQUE NOT NULL,
  password_hash     TEXT NOT NULL,
  display_name      VARCHAR(60),
  avatar_url        TEXT,
  country_code      CHAR(2),
  preferred_language VARCHAR(10) DEFAULT 'en',
  role              user_role DEFAULT 'user',
  status            account_status DEFAULT 'active',
  email_verified_at TIMESTAMPTZ,
  last_login_at     TIMESTAMPTZ,
  last_login_ip     INET,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_settings (
  user_id   UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  settings  JSONB NOT NULL DEFAULT '{}',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_auth_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL,
  purpose     VARCHAR(30) NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user ON user_auth_tokens(user_id, purpose);

CREATE TABLE IF NOT EXISTS exam_categories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code        VARCHAR(20) UNIQUE NOT NULL,
  name        VARCHAR(100) NOT NULL,
  description TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS topics (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id) ON DELETE CASCADE,
  parent_topic_id   UUID REFERENCES topics(id) ON DELETE CASCADE,
  name              VARCHAR(100) NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_topics_category ON topics(exam_category_id);
CREATE INDEX IF NOT EXISTS idx_topics_parent ON topics(parent_topic_id);

CREATE TABLE IF NOT EXISTS tags (
  id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name  CITEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS questions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_category_id      UUID NOT NULL REFERENCES exam_categories(id),
  topic_id              UUID NOT NULL REFERENCES topics(id),
  question_type         question_type NOT NULL DEFAULT 'mcq_single',
  difficulty            difficulty_level NOT NULL,
  language              VARCHAR(10) NOT NULL DEFAULT 'en',
  body                  TEXT NOT NULL,
  explanation           TEXT,
  estimated_time_seconds SMALLINT NOT NULL DEFAULT 60,
  status                question_status NOT NULL DEFAULT 'draft',
  created_by            UUID REFERENCES users(id),
  reviewed_by           UUID REFERENCES users(id),
  published_at          TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_questions_selection ON questions(exam_category_id, topic_id, difficulty, status)
  WHERE status = 'published';

CREATE TABLE IF NOT EXISTS question_options (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  option_text  TEXT NOT NULL,
  is_correct   BOOLEAN NOT NULL DEFAULT false,
  order_index  SMALLINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_options_question ON question_options(question_id);

CREATE TABLE IF NOT EXISTS question_images (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  image_url    TEXT NOT NULL,
  alt_text     VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS question_tags (
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  tag_id       UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (question_id, tag_id)
);

CREATE TABLE IF NOT EXISTS practice_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id UUID NOT NULL REFERENCES exam_categories(id),
  topic_id         UUID REFERENCES topics(id),
  difficulty       difficulty_level,
  question_count   SMALLINT NOT NULL,
  started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at         TIMESTAMPTZ,
  status           VARCHAR(20) NOT NULL DEFAULT 'in_progress'
);
CREATE INDEX IF NOT EXISTS idx_practice_user ON practice_sessions(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS practice_session_questions (
  session_id     UUID NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  question_id    UUID NOT NULL REFERENCES questions(id),
  order_index    SMALLINT NOT NULL,
  selected_option_id UUID REFERENCES question_options(id),
  is_correct     BOOLEAN,
  time_taken_ms  INTEGER,
  answered_at    TIMESTAMPTZ,
  PRIMARY KEY (session_id, question_id)
);

CREATE TABLE IF NOT EXISTS matches (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  match_type        match_type NOT NULL,
  status            match_status NOT NULL DEFAULT 'waiting',
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  room_code         VARCHAR(10),
  tournament_id     UUID,
  timer_seconds     SMALLINT NOT NULL,
  max_players       SMALLINT NOT NULL DEFAULT 2,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at        TIMESTAMPTZ,
  ended_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_matches_room_code ON matches(room_code) WHERE room_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status) WHERE status IN ('waiting','starting','in_progress');

CREATE TABLE IF NOT EXISTS match_players (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  match_id        UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  user_id         UUID NOT NULL REFERENCES users(id),
  score            INTEGER NOT NULL DEFAULT 0,
  final_rank      SMALLINT,
  rating_before   INTEGER,
  rating_after    INTEGER,
  rating_delta    INTEGER,
  connection_status player_conn_status NOT NULL DEFAULT 'connected',
  joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  left_at         TIMESTAMPTZ,
  UNIQUE (match_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_match_players_user ON match_players(user_id, joined_at DESC);

CREATE TABLE IF NOT EXISTS match_questions (
  match_id     UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  question_id  UUID NOT NULL REFERENCES questions(id),
  order_index  SMALLINT NOT NULL,
  PRIMARY KEY (match_id, order_index)
);

CREATE TABLE IF NOT EXISTS match_answers (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  match_id           UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  user_id            UUID NOT NULL REFERENCES users(id),
  question_id        UUID NOT NULL REFERENCES questions(id),
  selected_option_id UUID REFERENCES question_options(id),
  is_correct         BOOLEAN NOT NULL,
  time_taken_ms      INTEGER NOT NULL,
  answered_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (match_id, user_id, question_id)
);
CREATE INDEX IF NOT EXISTS idx_match_answers_match ON match_answers(match_id);

CREATE TABLE IF NOT EXISTS user_ratings (
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  rating            INTEGER NOT NULL DEFAULT 1200,
  matches_played    INTEGER NOT NULL DEFAULT 0,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, exam_category_id)
);
CREATE INDEX IF NOT EXISTS idx_user_ratings_lookup ON user_ratings(exam_category_id, rating);

CREATE TABLE IF NOT EXISTS rating_history (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id),
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  match_id          UUID REFERENCES matches(id),
  old_rating        INTEGER NOT NULL,
  new_rating        INTEGER NOT NULL,
  delta             INTEGER NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_rating_history_user ON rating_history(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS seasons (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(60) NOT NULL,
  start_date  TIMESTAMPTZ NOT NULL,
  end_date    TIMESTAMPTZ NOT NULL,
  status      VARCHAR(20) NOT NULL DEFAULT 'upcoming'
);

CREATE TABLE IF NOT EXISTS season_ratings (
  season_id         UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  rating            INTEGER NOT NULL DEFAULT 1200,
  matches_played    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (season_id, user_id, exam_category_id)
);
CREATE INDEX IF NOT EXISTS idx_season_ratings_rank ON season_ratings(season_id, exam_category_id, rating DESC);

CREATE TABLE IF NOT EXISTS tournaments (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_category_id    UUID NOT NULL REFERENCES exam_categories(id),
  name                VARCHAR(150) NOT NULL,
  description         TEXT,
  status              tournament_status NOT NULL DEFAULT 'scheduled',
  rules               JSONB NOT NULL DEFAULT '{}',
  max_participants    INTEGER,
  registration_start  TIMESTAMPTZ NOT NULL,
  registration_end    TIMESTAMPTZ NOT NULL,
  start_time          TIMESTAMPTZ NOT NULL,
  end_time            TIMESTAMPTZ,
  created_by          UUID REFERENCES users(id),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tournament_participants (
  tournament_id  UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  registered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  final_rank     INTEGER,
  final_score    INTEGER,
  PRIMARY KEY (tournament_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_tourney_participants_rank ON tournament_participants(tournament_id, final_rank);

CREATE TABLE IF NOT EXISTS user_statistics (
  user_id                   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id          UUID NOT NULL REFERENCES exam_categories(id),
  total_matches             INTEGER NOT NULL DEFAULT 0,
  wins                      INTEGER NOT NULL DEFAULT 0,
  losses                    INTEGER NOT NULL DEFAULT 0,
  draws                     INTEGER NOT NULL DEFAULT 0,
  current_win_streak        INTEGER NOT NULL DEFAULT 0,
  longest_win_streak        INTEGER NOT NULL DEFAULT 0,
  longest_losing_streak     INTEGER NOT NULL DEFAULT 0,
  total_questions_solved    INTEGER NOT NULL DEFAULT 0,
  total_practice_sessions   INTEGER NOT NULL DEFAULT 0,
  overall_accuracy          NUMERIC(5,2) NOT NULL DEFAULT 0,
  avg_solving_time_ms       INTEGER,
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, exam_category_id)
);

CREATE TABLE IF NOT EXISTS user_topic_statistics (
  user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  topic_id            UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
  questions_attempted INTEGER NOT NULL DEFAULT 0,
  questions_correct   INTEGER NOT NULL DEFAULT 0,
  avg_time_ms         INTEGER,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, topic_id)
);

CREATE TABLE IF NOT EXISTS achievements (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         VARCHAR(50) UNIQUE NOT NULL,
  name         VARCHAR(100) NOT NULL,
  description  TEXT,
  icon_url     TEXT,
  criteria     JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS user_achievements (
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
  unlocked_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, achievement_id)
);

CREATE TABLE IF NOT EXISTS friendships (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  addressee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status        friend_req_status NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  responded_at  TIMESTAMPTZ,
  CHECK (requester_id <> addressee_id),
  UNIQUE (requester_id, addressee_id)
);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships(addressee_id, status);

CREATE TABLE IF NOT EXISTS reports (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id  UUID NOT NULL REFERENCES users(id),
  target_type  report_target NOT NULL,
  target_id    UUID NOT NULL,
  reason       VARCHAR(100) NOT NULL,
  description  TEXT,
  status       report_status NOT NULL DEFAULT 'open',
  reviewed_by  UUID REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  resolved_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);

CREATE TABLE IF NOT EXISTS notifications (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type        notification_type NOT NULL,
  payload     JSONB NOT NULL DEFAULT '{}',
  is_read     BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);

CREATE TABLE IF NOT EXISTS matchmaking_queue (
  user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  match_type        match_type NOT NULL,
  rating            INTEGER NOT NULL,
  queued_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_matchmaking_lookup ON matchmaking_queue(exam_category_id, match_type, rating);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  admin_id    UUID NOT NULL REFERENCES users(id),
  action      VARCHAR(100) NOT NULL,
  target_type VARCHAR(50) NOT NULL,
  target_id   UUID,
  metadata    JSONB NOT NULL DEFAULT '{}',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_admin ON admin_audit_logs(admin_id, created_at DESC);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
  CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TRIGGER trg_questions_updated_at BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TRIGGER trg_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Seed exam categories
INSERT INTO exam_categories (code, name, description) VALUES
  ('SSC', 'SSC Exams', 'Staff Selection Commission exams'),
  ('BANKING', 'Banking Exams', 'IBPS, SBI and other banking exams'),
  ('RAILWAYS', 'Railway Exams', 'RRB and railway recruitment exams'),
  ('UPSC', 'UPSC Civil Services', 'Union Public Service Commission exams'),
  ('CAT', 'CAT/MBA', 'Common Admission Test and MBA entrance exams'),
  ('STATE_PSC', 'State PSC', 'State Public Service Commission exams')
ON CONFLICT (code) DO NOTHING;