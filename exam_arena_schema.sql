-- =====================================================================
-- EXAM ARENA — PRODUCTION DATABASE SCHEMA
-- Engine target : PostgreSQL 15+
-- Design goal   : Postgres = single source of truth for everything.
--                 Redis is NOT required to run this schema today.
--                 Sections marked [REDIS-READY] are the pieces you will
--                 migrate to Redis later WITHOUT changing any other
--                 table. See schema-design.md for the migration plan.
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "citext";       -- case-insensitive email/username

-- =====================================================================
-- 0. ENUM TYPES
-- =====================================================================

CREATE TYPE user_role        AS ENUM ('user', 'admin', 'moderator');
CREATE TYPE account_status   AS ENUM ('active', 'suspended', 'banned', 'deleted');
CREATE TYPE question_type    AS ENUM ('mcq_single', 'mcq_multiple', 'integer');
CREATE TYPE question_status  AS ENUM ('draft', 'in_review', 'published', 'archived');
CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');

CREATE TYPE match_type       AS ENUM ('ranked', 'friend', 'arena', 'tournament', 'daily_challenge');
CREATE TYPE match_status     AS ENUM ('waiting', 'starting', 'in_progress', 'completed', 'abandoned', 'cancelled');
CREATE TYPE player_conn_status AS ENUM ('connected', 'disconnected', 'reconnecting', 'left');

CREATE TYPE tournament_status AS ENUM ('scheduled', 'registration_open', 'in_progress', 'completed', 'cancelled');

CREATE TYPE friend_req_status AS ENUM ('pending', 'accepted', 'declined', 'blocked');
CREATE TYPE report_status     AS ENUM ('open', 'reviewed', 'resolved', 'dismissed');
CREATE TYPE report_target     AS ENUM ('question', 'player', 'content');

CREATE TYPE notification_type AS ENUM (
  'friend_invite', 'tournament_start', 'match_found',
  'achievement_unlocked', 'daily_challenge', 'system_announcement'
);

-- =====================================================================
-- 1. USERS & AUTH
-- =====================================================================

CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username          CITEXT UNIQUE NOT NULL,
  email             CITEXT UNIQUE NOT NULL,
  password_hash     TEXT NOT NULL,
  display_name      VARCHAR(60),
  avatar_url        TEXT,
  country_code      CHAR(2),                       -- ISO 3166-1 alpha-2
  preferred_language VARCHAR(10) DEFAULT 'en',
  role              user_role DEFAULT 'user',
  status            account_status DEFAULT 'active',
  email_verified_at TIMESTAMPTZ,
  last_login_at     TIMESTAMPTZ,
  last_login_ip     INET,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ                     -- soft delete
);
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;

-- Free-form, evolving preferences. JSONB means new settings never need a migration.
CREATE TABLE user_settings (
  user_id   UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  settings  JSONB NOT NULL DEFAULT '{}',            -- e.g. {"sound": true, "theme": "dark"}
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Password reset / email verification tokens. Short-lived by nature —
-- [REDIS-READY] first candidate to move to Redis with a TTL (SETEX).
CREATE TABLE user_auth_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL,
  purpose     VARCHAR(30) NOT NULL,                 -- 'password_reset' | 'email_verify'
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_auth_tokens_user ON user_auth_tokens(user_id, purpose);

-- =====================================================================
-- 2. EXAM TAXONOMY & QUESTION BANK
-- =====================================================================

CREATE TABLE exam_categories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code        VARCHAR(20) UNIQUE NOT NULL,           -- 'SSC','BANKING','RAILWAYS','UPSC','CAT','STATE_PSC'
  name        VARCHAR(100) NOT NULL,
  description TEXT,
  is_active   BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Self-referencing so a "topic" can have "subtopics" without a second table.
CREATE TABLE topics (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id) ON DELETE CASCADE,
  parent_topic_id   UUID REFERENCES topics(id) ON DELETE CASCADE,   -- NULL = top-level topic
  name              VARCHAR(100) NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_topics_category ON topics(exam_category_id);
CREATE INDEX idx_topics_parent ON topics(parent_topic_id);

CREATE TABLE tags (
  id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name  CITEXT UNIQUE NOT NULL
);

CREATE TABLE questions (
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
-- Matchmaking's hottest read: "give me N published questions for category+topic+difficulty"
CREATE INDEX idx_questions_selection ON questions(exam_category_id, topic_id, difficulty, status)
  WHERE status = 'published';

CREATE TABLE question_options (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  option_text  TEXT NOT NULL,
  is_correct   BOOLEAN NOT NULL DEFAULT false,
  order_index  SMALLINT NOT NULL
);
CREATE INDEX idx_options_question ON question_options(question_id);

CREATE TABLE question_images (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  image_url    TEXT NOT NULL,
  alt_text     VARCHAR(255)
);

CREATE TABLE question_tags (
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  tag_id       UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (question_id, tag_id)
);

-- =====================================================================
-- 3. PRACTICE MODE (no rating impact)
-- =====================================================================

CREATE TABLE practice_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id UUID NOT NULL REFERENCES exam_categories(id),
  topic_id         UUID REFERENCES topics(id),
  difficulty       difficulty_level,
  question_count   SMALLINT NOT NULL,
  started_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at         TIMESTAMPTZ,
  status           VARCHAR(20) NOT NULL DEFAULT 'in_progress'  -- in_progress|completed|abandoned
);
CREATE INDEX idx_practice_user ON practice_sessions(user_id, started_at DESC);

CREATE TABLE practice_session_questions (
  session_id     UUID NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  question_id    UUID NOT NULL REFERENCES questions(id),
  order_index    SMALLINT NOT NULL,
  selected_option_id UUID REFERENCES question_options(id),
  is_correct     BOOLEAN,
  time_taken_ms  INTEGER,
  answered_at    TIMESTAMPTZ,
  PRIMARY KEY (session_id, question_id)
);

-- =====================================================================
-- 4. MATCHES — unified table for ranked / friend / arena / tournament
-- =====================================================================

CREATE TABLE matches (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  match_type        match_type NOT NULL,
  status            match_status NOT NULL DEFAULT 'waiting',
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  room_code         VARCHAR(10),                    -- only for friend battles
  tournament_id     UUID,                           -- FK added after tournaments table
  timer_seconds     SMALLINT NOT NULL,
  max_players       SMALLINT NOT NULL DEFAULT 2,     -- 2 for 1v1, >2 for arena
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at        TIMESTAMPTZ,
  ended_at          TIMESTAMPTZ
);
CREATE INDEX idx_matches_room_code ON matches(room_code) WHERE room_code IS NOT NULL;
CREATE INDEX idx_matches_status ON matches(status) WHERE status IN ('waiting','starting','in_progress');

CREATE TABLE match_players (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  match_id        UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  user_id         UUID NOT NULL REFERENCES users(id),
  score            INTEGER NOT NULL DEFAULT 0,
  final_rank      SMALLINT,                          -- 1st, 2nd... (arena/tournament)
  rating_before   INTEGER,
  rating_after    INTEGER,
  rating_delta    INTEGER,
  connection_status player_conn_status NOT NULL DEFAULT 'connected',
  joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  left_at         TIMESTAMPTZ,
  UNIQUE (match_id, user_id)
);
CREATE INDEX idx_match_players_user ON match_players(user_id, joined_at DESC);

-- Guarantees "same questions, same order" for every player in a match.
CREATE TABLE match_questions (
  match_id     UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  question_id  UUID NOT NULL REFERENCES questions(id),
  order_index  SMALLINT NOT NULL,
  PRIMARY KEY (match_id, order_index)
);

CREATE TABLE match_answers (
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
CREATE INDEX idx_match_answers_match ON match_answers(match_id);

-- =====================================================================
-- 5. RATING SYSTEM (Elo-style)
-- =====================================================================

-- Current rating per user per exam category — fast read for matchmaking & profile.
CREATE TABLE user_ratings (
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  rating            INTEGER NOT NULL DEFAULT 1200,
  matches_played    INTEGER NOT NULL DEFAULT 0,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, exam_category_id)
);
-- Powers matchmaking's "find opponents near my rating" query.
CREATE INDEX idx_user_ratings_lookup ON user_ratings(exam_category_id, rating);

-- Append-only audit trail — never updated, only inserted. This is what
-- lets you rebuild user_ratings or a Redis leaderboard from scratch at any time.
CREATE TABLE rating_history (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id),
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  match_id          UUID REFERENCES matches(id),
  old_rating        INTEGER NOT NULL,
  new_rating        INTEGER NOT NULL,
  delta             INTEGER NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rating_history_user ON rating_history(user_id, created_at DESC);

-- =====================================================================
-- 6. SEASONS & LEADERBOARDS
-- =====================================================================
-- NOTE: These tables are the SOURCE OF TRUTH. The actual "live" leaderboard
-- the user scrolls through should be served from a cache (materialized view
-- today, Redis ZSET tomorrow) rebuilt from this data — see schema-design.md.

CREATE TABLE seasons (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(60) NOT NULL,
  start_date  TIMESTAMPTZ NOT NULL,
  end_date    TIMESTAMPTZ NOT NULL,
  status      VARCHAR(20) NOT NULL DEFAULT 'upcoming'  -- upcoming|active|completed
);

CREATE TABLE season_ratings (
  season_id         UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  rating            INTEGER NOT NULL DEFAULT 1200,
  matches_played    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (season_id, user_id, exam_category_id)
);
CREATE INDEX idx_season_ratings_rank ON season_ratings(season_id, exam_category_id, rating DESC);

-- =====================================================================
-- 7. TOURNAMENTS
-- =====================================================================

CREATE TABLE tournaments (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_category_id    UUID NOT NULL REFERENCES exam_categories(id),
  name                VARCHAR(150) NOT NULL,
  description         TEXT,
  status              tournament_status NOT NULL DEFAULT 'scheduled',
  rules               JSONB NOT NULL DEFAULT '{}',     -- flexible: rounds, scoring, elimination type
  max_participants    INTEGER,
  registration_start  TIMESTAMPTZ NOT NULL,
  registration_end    TIMESTAMPTZ NOT NULL,
  start_time          TIMESTAMPTZ NOT NULL,
  end_time            TIMESTAMPTZ,
  created_by          UUID REFERENCES users(id),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE matches
  ADD CONSTRAINT fk_matches_tournament
  FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE SET NULL;

CREATE TABLE tournament_participants (
  tournament_id  UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  registered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  final_rank     INTEGER,
  final_score    INTEGER,
  PRIMARY KEY (tournament_id, user_id)
);
CREATE INDEX idx_tourney_participants_rank ON tournament_participants(tournament_id, final_rank);

-- =====================================================================
-- 8. STATISTICS (denormalized, updated async after each match/session)
-- =====================================================================

CREATE TABLE user_statistics (
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
  overall_accuracy          NUMERIC(5,2) NOT NULL DEFAULT 0,   -- percentage
  avg_solving_time_ms       INTEGER,
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, exam_category_id)
);

-- Per-topic breakdown -> powers "weak topics / strong topics" recommendations.
CREATE TABLE user_topic_statistics (
  user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  topic_id            UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
  questions_attempted INTEGER NOT NULL DEFAULT 0,
  questions_correct   INTEGER NOT NULL DEFAULT 0,
  avg_time_ms         INTEGER,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, topic_id)
);

-- =====================================================================
-- 9. ACHIEVEMENTS & GAMIFICATION
-- =====================================================================

CREATE TABLE achievements (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         VARCHAR(50) UNIQUE NOT NULL,      -- 'FIRST_VICTORY', '10_WINS'...
  name         VARCHAR(100) NOT NULL,
  description  TEXT,
  icon_url     TEXT,
  criteria     JSONB NOT NULL DEFAULT '{}'        -- e.g. {"type":"win_count","value":10}
);

CREATE TABLE user_achievements (
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
  unlocked_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, achievement_id)
);

-- =====================================================================
-- 10. SOCIAL — friends & reporting
-- =====================================================================

CREATE TABLE friendships (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  addressee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status        friend_req_status NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  responded_at  TIMESTAMPTZ,
  CHECK (requester_id <> addressee_id),
  UNIQUE (requester_id, addressee_id)
);
CREATE INDEX idx_friendships_addressee ON friendships(addressee_id, status);

CREATE TABLE reports (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id  UUID NOT NULL REFERENCES users(id),
  target_type  report_target NOT NULL,
  target_id    UUID NOT NULL,               -- polymorphic: question_id / user_id / other
  reason       VARCHAR(100) NOT NULL,
  description  TEXT,
  status       report_status NOT NULL DEFAULT 'open',
  reviewed_by  UUID REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  resolved_at  TIMESTAMPTZ
);
CREATE INDEX idx_reports_status ON reports(status);

-- =====================================================================
-- 11. NOTIFICATIONS
-- =====================================================================
-- Source of truth for notification history stays in Postgres.
-- [REDIS-READY] real-time DELIVERY (push to an open socket) should go
-- through Redis Pub/Sub later; this table remains the durable record.

CREATE TABLE notifications (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type        notification_type NOT NULL,
  payload     JSONB NOT NULL DEFAULT '{}',
  is_read     BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);

-- =====================================================================
-- 12. MATCHMAKING QUEUE  [REDIS-READY — highest priority migration]
-- =====================================================================
-- This table works fine at low concurrency, but it is the #1 candidate
-- to move to Redis (sorted set keyed by rating) once you need sub-second,
-- high-concurrency matchmaking. Kept schema-identical to a Redis ZSET
-- member+score model on purpose: (user_id) -> score = rating.

CREATE TABLE matchmaking_queue (
  user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id  UUID NOT NULL REFERENCES exam_categories(id),
  match_type        match_type NOT NULL,
  rating            INTEGER NOT NULL,
  queued_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_matchmaking_lookup ON matchmaking_queue(exam_category_id, match_type, rating);

-- =====================================================================
-- 13. ADMIN AUDIT LOG
-- =====================================================================

CREATE TABLE admin_audit_logs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  admin_id    UUID NOT NULL REFERENCES users(id),
  action      VARCHAR(100) NOT NULL,           -- 'question.publish', 'user.ban', etc.
  target_type VARCHAR(50) NOT NULL,
  target_id   UUID,
  metadata    JSONB NOT NULL DEFAULT '{}',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_admin ON admin_audit_logs(admin_id, created_at DESC);

-- =====================================================================
-- 14. UPDATED_AT TRIGGER (applied to the tables that need it)
-- =====================================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_questions_updated_at BEFORE UPDATE ON questions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_user_settings_updated_at BEFORE UPDATE ON user_settings
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
