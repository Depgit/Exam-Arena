# questionpool

**Job:** keep a stock of generated questions in the database, next to the
hand-written ones, and keep it fresh.

- **Size grows with players:** each category gets players × 100 ÷
  categories published questions (at most 5000). The generator makes what
  the hand-written questions leave of that, split across easy/medium/hard.
  "Players" means real accounts — demo accounts, bots and admins don't count.
- Beyond that share the oldest are archived (hand-written included), so the
  stock also shrinks if players are removed.
- On start and every 6 hours it re-checks the player count and fills up.
  Every 6 hours it also archives the oldest 25 % and generates replacements,
  and prunes archived questions nothing refers to any more.
- With `QUESTION_GEN_PER_USER=0` the generated pool is a fixed
  `QUESTION_GEN_POOL_SIZE` per category instead.
- Safe with several servers on one database (a transaction lock per pool).

| Endpoint (admin) | Gives back |
|---|---|
| `GET /api/v1/admin/generator/preview?category=MATH&difficulty=hard&n=20` | sample questions with answers — nothing saved |
| `POST /api/v1/admin/generator/rotate` | how many were archived, generated and pruned |

**Settings:** `QUESTION_GEN_ENABLED`, `QUESTION_GEN_POOL_SIZE`,
`QUESTION_GEN_ROTATE_HOURS`, `QUESTION_GEN_PER_USER`,
`QUESTION_GEN_MAX_PER_CATEGORY` (default 5000; values under 500 are ignored,
because the ceiling archives everything above it).

**Uses:** `questiongen` (to invent), `questions` (to store and to refresh the bank).
