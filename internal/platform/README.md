# platform — shared plumbing

Small building blocks that every feature uses. None of them knows anything
about the game.

| Folder | Job | You give it → you get back |
|---|---|---|
| `config/` | Read settings from env / `.env` | — → `*Config` |
| `database/` | Connect to Postgres, run migrations, recognise DB errors | URL → `*pgxpool.Pool` |
| `cache/` | In-memory key/value store with expiry | key → value |
| `respond/` | Write the standard JSON reply `{success, data, error}` | data / error → HTTP response |
| `middleware/` | Wrap every request: login check, admin check, CORS, rate limit, logging, panic recovery | handler → wrapped handler |
| `tokens/` | Create and check login tokens (JWT) | user → token, token → user |
| `passwords/` | Hash and check passwords (bcrypt) | password → hash, (password, hash) → ok |
| `realtime/` | WebSocket: connect players, deliver messages, route incoming ones | user id + message → delivered |
