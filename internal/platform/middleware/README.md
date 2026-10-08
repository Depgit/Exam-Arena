# middleware

**Job:** code that runs around *every* request, before the feature sees it.

| Function | What it does |
|---|---|
| `Auth(secret, h)` | Requires a valid login token; puts the user id/name on the request |
| `RequireRole("admin", h)` | Only lets admins through (use inside `Auth`) |
| `GetUserID(r)` / `GetUsername(r)` | Who is making this request |
| `CORS` | Lets the website (Netlify) call this API |
| `RateLimit(rps, burst, h)` | Limits requests per IP |
| `Logging` | One log line per request; errors stand out, health checks are quiet |
| `Recovery` | A crash inside a request returns a 500 instead of killing the server |

**Uses:** `tokens` (to check logins), `respond` (to reply with errors).
