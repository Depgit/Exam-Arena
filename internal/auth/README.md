# auth

**Job:** let people in — sign up, log in, demo (guest) accounts — and hand
out login tokens.

| Endpoint | Gives back |
|---|---|
| `POST /api/v1/auth/register` | token + user (usernames starting `bot.` / `guest_` are reserved) |
| `POST /api/v1/auth/login` | token + user (log in with username or email) |
| `POST /api/v1/auth/guest` | a brand-new demo account + token (max 10 per IP per hour) |
| `POST /api/v1/auth/google` | token + user, or `needs_username` for a first-time Google player (send the same `credential` again with a `username`) |
| `GET /api/v1/auth/me` | the logged-in user (`needs_email_verification` tells the app to ask for the code) |
| `POST /api/v1/auth/verify-email/send` | emails a 6-digit code (optional `email` fixes a typo first; one per 60 s) |
| `POST /api/v1/auth/verify-email` | checks the code (15 min, 5 tries) and returns the verified user |

**Google sign-in:** the app's Google button hands us an ID token; `google.go`
checks it against Google's public keys (no client secret). A Google account
is matched by its id, then by email (linking the existing player), else a new
player is created after they pick a username.

**Email verification:** on when `RESEND_API_KEY` is set (or
`EMAIL_VERIFICATION=required`). Unverified players can look around but not
play or chat — `RegisteredOnly` blocks them like demo accounts. Google
players are verified automatically. Throwaway inboxes are refused.

**Files:** `service.go` (rules), `handler.go` (endpoints, guest rate limit),
`google.go` + `google_login.go` (Google sign-in), `verify_email.go` + `codes.go`
(email codes), `registered.go` (who may play).

**Uses:** `users` (to create/find accounts), `platform/tokens`, `platform/passwords`,
`platform/mailer` (sends the codes through resend.com).

**Settings:** `GOOGLE_CLIENT_ID`, `RESEND_API_KEY`, `MAIL_FROM`
(default `Mind Race <noreply@mindrace.in>`), `EMAIL_VERIFICATION` (`required` / `off`).
