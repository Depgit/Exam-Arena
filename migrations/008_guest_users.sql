-- 008: one-click demo accounts ("Try the demo" on the login page).
--
-- Each visitor gets their own throwaway account rather than sharing one:
-- the WebSocket hub allows one live session per user, so a shared demo
-- account would have visitors kicking each other off.
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_guest BOOLEAN NOT NULL DEFAULT false;

-- Cleanup scans old guests.
CREATE INDEX IF NOT EXISTS idx_users_guest_created ON users(created_at) WHERE is_guest;
