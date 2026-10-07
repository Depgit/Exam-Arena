-- 009: bot opponents ("Play vs Bot" when nobody else is searching).
--
-- Bots are real user rows because matches store their players, but they
-- are flagged so leaderboards, the player count and similar can skip them.
-- Their accounts are created on demand by the server (internal/service/bot.go).
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_bot BOOLEAN NOT NULL DEFAULT false;
