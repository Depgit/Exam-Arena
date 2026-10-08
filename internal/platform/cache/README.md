# cache

**Job:** a small in-memory key/value store whose entries expire.

**You call:** `NewMemoryCache()`, then `Set(key, value, ttl)`, `Get(key)`,
`Delete`, `DeletePrefix`, `DeleteIfPresent`.

**Used for:** live matches (`live_match:<id>`), cached leaderboards, friend
challenges. Everything is lost on restart — it's not a database.
