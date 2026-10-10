-- 011: "Continue with Google" and email verification.

-- Google's stable account id ("sub" in its sign-in token). Set when a player
-- signs in with Google, so later sign-ins find them even if they change the
-- email on their Google account.
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_sub TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_sub ON users(google_sub) WHERE google_sub IS NOT NULL;

-- One pending 6-digit code per player. Only a hash of the code is stored.
CREATE TABLE IF NOT EXISTS email_verification_codes (
  user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  code_hash  TEXT NOT NULL,
  email      CITEXT NOT NULL,          -- the address the code was sent to
  expires_at TIMESTAMPTZ NOT NULL,
  attempts   INT NOT NULL DEFAULT 0,   -- wrong guesses so far
  sent_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
