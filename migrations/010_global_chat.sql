-- 010: the global chat (internal/chat).
--
-- The server keeps the most recent messages in memory and only reads this
-- table at startup, so a restart doesn't wipe the chat. Old rows are
-- trimmed by the server; a deleted account takes its messages with it.
CREATE TABLE IF NOT EXISTS chat_messages (
  id          BIGSERIAL PRIMARY KEY,
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body        TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_created ON chat_messages(created_at DESC);
