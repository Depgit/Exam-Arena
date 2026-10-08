# auth

**Job:** let people in — sign up, log in, demo (guest) accounts — and hand
out login tokens.

| Endpoint | Gives back |
|---|---|
| `POST /api/v1/auth/register` | token + user (usernames starting `bot.` / `guest_` are reserved) |
| `POST /api/v1/auth/login` | token + user (log in with username or email) |
| `POST /api/v1/auth/guest` | a brand-new demo account + token (max 10 per IP per hour) |
| `GET /api/v1/auth/me` | the logged-in user |

**Files:** `service.go` (rules), `handler.go` (endpoints, guest rate limit).

**Uses:** `users` (to create/find accounts), `platform/tokens`, `platform/passwords`.
