-- 004: when a user last used the app (login or WebSocket connect/disconnect).
-- Backs the "active users" counts on the admin dashboard.
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ;

UPDATE users SET last_active_at = last_login_at WHERE last_active_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_last_active ON users(last_active_at) WHERE deleted_at IS NULL;
