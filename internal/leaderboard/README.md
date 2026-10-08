# leaderboard

**Job:** rankings by rating for one category.

**Endpoint:** `GET /api/v1/leaderboard/{category}?limit=&offset=` → ranked
players. Guests and bots are left out.

Results are cached in memory and cleared whenever a match ends.

**Uses:** `platform/cache`, `models`.
