-- SQLite Schema for Exam Arena

CREATE TABLE IF NOT EXISTS users (
  id                 TEXT PRIMARY KEY,
  username           TEXT UNIQUE NOT NULL,
  email              TEXT UNIQUE NOT NULL,
  password_hash      TEXT NOT NULL,
  display_name       TEXT,
  avatar_url         TEXT,
  country_code       TEXT,
  preferred_language TEXT DEFAULT 'en',
  role               TEXT DEFAULT 'user',
  status             TEXT DEFAULT 'active',
  email_verified_at  DATETIME,
  last_login_at      DATETIME,
  last_login_ip      TEXT,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at         DATETIME
);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE TABLE IF NOT EXISTS user_settings (
  user_id    TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  settings   TEXT NOT NULL DEFAULT '{}',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_auth_tokens (
  id          TEXT PRIMARY KEY,
  user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL,
  purpose     TEXT NOT NULL,
  expires_at  DATETIME NOT NULL,
  used_at     DATETIME,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user ON user_auth_tokens(user_id, purpose);

CREATE TABLE IF NOT EXISTS exam_categories (
  id          TEXT PRIMARY KEY,
  code        TEXT UNIQUE NOT NULL,
  name        TEXT NOT NULL,
  description TEXT,
  is_active   INTEGER NOT NULL DEFAULT 1,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS topics (
  id               TEXT PRIMARY KEY,
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id) ON DELETE CASCADE,
  parent_topic_id  TEXT REFERENCES topics(id) ON DELETE CASCADE,
  name             TEXT NOT NULL,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_topics_category ON topics(exam_category_id);
CREATE INDEX IF NOT EXISTS idx_topics_parent ON topics(parent_topic_id);

CREATE TABLE IF NOT EXISTS tags (
  id   TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS questions (
  id                     TEXT PRIMARY KEY,
  exam_category_id       TEXT NOT NULL REFERENCES exam_categories(id),
  topic_id               TEXT NOT NULL REFERENCES topics(id),
  question_type          TEXT NOT NULL DEFAULT 'mcq_single',
  difficulty             TEXT NOT NULL,
  language               TEXT NOT NULL DEFAULT 'en',
  body                   TEXT NOT NULL,
  explanation            TEXT,
  estimated_time_seconds INTEGER NOT NULL DEFAULT 60,
  status                 TEXT NOT NULL DEFAULT 'draft',
  created_by             TEXT REFERENCES users(id),
  reviewed_by            TEXT REFERENCES users(id),
  published_at           DATETIME,
  created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_questions_selection ON questions(exam_category_id, topic_id, difficulty, status);

CREATE TABLE IF NOT EXISTS question_options (
  id          TEXT PRIMARY KEY,
  question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  option_text TEXT NOT NULL,
  is_correct  INTEGER NOT NULL DEFAULT 0,
  order_index INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_options_question ON question_options(question_id);

CREATE TABLE IF NOT EXISTS practice_sessions (
  id               TEXT PRIMARY KEY,
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id),
  topic_id         TEXT REFERENCES topics(id),
  difficulty       TEXT,
  question_count   INTEGER NOT NULL,
  started_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ended_at         DATETIME,
  status           TEXT NOT NULL DEFAULT 'in_progress'
);
CREATE INDEX IF NOT EXISTS idx_practice_user ON practice_sessions(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS practice_session_questions (
  session_id         TEXT NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  question_id        TEXT NOT NULL REFERENCES questions(id),
  order_index        INTEGER NOT NULL,
  selected_option_id TEXT REFERENCES question_options(id),
  is_correct         INTEGER,
  time_taken_ms      INTEGER,
  answered_at        DATETIME,
  PRIMARY KEY (session_id, question_id)
);

CREATE TABLE IF NOT EXISTS matches (
  id               TEXT PRIMARY KEY,
  match_type       TEXT NOT NULL,
  status           TEXT NOT NULL DEFAULT 'waiting',
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id),
  room_code        TEXT,
  tournament_id    TEXT,
  timer_seconds    INTEGER NOT NULL,
  max_players      INTEGER NOT NULL DEFAULT 2,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at       DATETIME,
  ended_at         DATETIME
);
CREATE INDEX IF NOT EXISTS idx_matches_room_code ON matches(room_code);
CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status);

CREATE TABLE IF NOT EXISTS match_players (
  id                TEXT PRIMARY KEY,
  match_id          TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  user_id           TEXT NOT NULL REFERENCES users(id),
  score             INTEGER NOT NULL DEFAULT 0,
  final_rank        INTEGER,
  rating_before     INTEGER,
  rating_after      INTEGER,
  rating_delta      INTEGER,
  connection_status TEXT NOT NULL DEFAULT 'connected',
  joined_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at           DATETIME,
  UNIQUE (match_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_match_players_user ON match_players(user_id, joined_at DESC);

CREATE TABLE IF NOT EXISTS match_questions (
  match_id    TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  question_id TEXT NOT NULL REFERENCES questions(id),
  order_index INTEGER NOT NULL,
  PRIMARY KEY (match_id, order_index)
);

CREATE TABLE IF NOT EXISTS match_answers (
  id                 TEXT PRIMARY KEY,
  match_id           TEXT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  user_id            TEXT NOT NULL REFERENCES users(id),
  question_id        TEXT NOT NULL REFERENCES questions(id),
  selected_option_id TEXT REFERENCES question_options(id),
  is_correct         INTEGER NOT NULL,
  time_taken_ms      INTEGER NOT NULL,
  answered_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (match_id, user_id, question_id)
);
CREATE INDEX IF NOT EXISTS idx_match_answers_match ON match_answers(match_id);

CREATE TABLE IF NOT EXISTS user_ratings (
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id),
  rating           INTEGER NOT NULL DEFAULT 1200,
  matches_played   INTEGER NOT NULL DEFAULT 0,
  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, exam_category_id)
);
CREATE INDEX IF NOT EXISTS idx_user_ratings_lookup ON user_ratings(exam_category_id, rating);

CREATE TABLE IF NOT EXISTS rating_history (
  id               TEXT PRIMARY KEY,
  user_id          TEXT NOT NULL REFERENCES users(id),
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id),
  match_id         TEXT REFERENCES matches(id),
  old_rating       INTEGER NOT NULL,
  new_rating       INTEGER NOT NULL,
  delta            INTEGER NOT NULL,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rating_history_user ON rating_history(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_statistics (
  user_id                 TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id        TEXT NOT NULL REFERENCES exam_categories(id),
  total_matches           INTEGER NOT NULL DEFAULT 0,
  wins                    INTEGER NOT NULL DEFAULT 0,
  losses                  INTEGER NOT NULL DEFAULT 0,
  draws                   INTEGER NOT NULL DEFAULT 0,
  current_win_streak      INTEGER NOT NULL DEFAULT 0,
  longest_win_streak      INTEGER NOT NULL DEFAULT 0,
  longest_losing_streak   INTEGER NOT NULL DEFAULT 0,
  total_questions_solved  INTEGER NOT NULL DEFAULT 0,
  total_practice_sessions INTEGER NOT NULL DEFAULT 0,
  overall_accuracy        REAL NOT NULL DEFAULT 0,
  avg_solving_time_ms     INTEGER,
  updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, exam_category_id)
);

CREATE TABLE IF NOT EXISTS matchmaking_queue (
  user_id          TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  exam_category_id TEXT NOT NULL REFERENCES exam_categories(id),
  match_type       TEXT NOT NULL,
  rating           INTEGER NOT NULL,
  queued_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_matchmaking_lookup ON matchmaking_queue(exam_category_id, match_type, rating);

-- Seed initial exam categories
INSERT OR IGNORE INTO exam_categories (id, code, name, description, is_active) VALUES
  ('cat-ssc-cgl', 'SSC', 'SSC CGL', 'Staff Selection Commission exams', 1),
  ('cat-banking', 'BANK', 'Banking / IBPS', 'IBPS, SBI and other banking exams', 1),
  ('cat-railways', 'RRB', 'Railways', 'RRB and railway recruitment exams', 1),
  ('cat-upsc', 'UPSC', 'UPSC Prelims', 'Union Public Service Commission exams', 1),
  ('cat-cat', 'CAT', 'CAT / MBA', 'Common Admission Test and MBA entrance exams', 1),
  ('cat-state-psc', 'STATE_PSC', 'State PSC', 'State Public Service Commission exams', 1);

-- Seed initial topic & questions so practice and battle mode work immediately out-of-the-box!
INSERT OR IGNORE INTO topics (id, exam_category_id, name) VALUES
  ('topic-ssc-qa', 'cat-ssc-cgl', 'Quantitative Aptitude'),
  ('topic-[#1]', 'cat-ssc-cgl', 'General Intelligence & Reasoning');

INSERT OR IGNORE INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status) VALUES
  ('q-ssc-1', 'cat-ssc-cgl', 'topic-ssc-qa', 'mcq_single', 'easy', 'If 20% of a number is 50, then what is 40% of that number?', 'If 20% is 50, then 40% (which is double 20%) is 50 * 2 = 100.', 30, 'published'),
  ('q-ssc-2', 'cat-ssc-cgl', 'topic-ssc-qa', 'mcq_single', 'medium', 'A train running at 72 km/h crosses a pole in 9 seconds. What is the length of the train in meters?', 'Speed = 72 * (5/18) = 20 m/s. Length = Speed * Time = 20 * 9 = 180 meters.', 45, 'published'),
  ('q-ssc-3', 'cat-ssc-cgl', 'topic-ssc-qa', 'mcq_single', 'hard', 'Two pipes A and B can fill a tank in 12 and 16 hours respectively. If both pipes are opened together, after how much time should pipe B be closed so that the tank is full in 9 hours?', 'A fills 1/12 per hour. In 9 hours A fills 9/12 = 3/4. Remaining 1/4 is filled by B. B rate is 1/16. Time for B = (1/4)/(1/16) = 4 hours.', 60, 'published');

INSERT OR IGNORE INTO question_options (id, question_id, option_text, is_correct, order_index) VALUES
  ('opt-1-a', 'q-ssc-1', '75', 0, 1),
  ('opt-1-b', 'q-ssc-1', '100', 1, 2),
  ('opt-1-c', 'q-ssc-1', '120', 0, 3),
  ('opt-1-d', 'q-ssc-1', '150', 0, 4),
  ('opt-2-a', 'q-ssc-1', '150 meters', 0, 1),
  ('opt-2-b', 'q-ssc-2', '180 meters', 1, 2),
  ('opt-2-c', 'q-ssc-2', '200 meters', 0, 3),
  ('opt-2-d', 'q-ssc-2', '240 meters', 0, 4),
  ('opt-3-a', 'q-ssc-3', '3 hours', 0, 1),
  ('opt-3-b', 'q-ssc-3', '4 hours', 1, 2),
  ('opt-3-c', 'q-ssc-3', '4.5 hours', 0, 3),
  ('opt-3-d', 'q-ssc-3', '6 hours', 0, 4);