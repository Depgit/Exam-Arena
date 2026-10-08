# questionpool

**Job:** keep a stock of generated questions in the database, next to the
hand-written ones, and keep it fresh.

- On start: tops each category up to its target (default 600, split across
  easy/medium/hard).
- Every 6 hours: archives the oldest 25 % and generates replacements; prunes
  archived questions nothing refers to any more.
- **Cap:** total published questions per category ≤ players × 100 ÷
  categories; beyond that the oldest are archived (hand-written included).
- Safe with several servers on one database (a transaction lock per pool).

| Endpoint (admin) | Gives back |
|---|---|
| `GET /api/v1/admin/generator/preview?category=MATH&difficulty=hard&n=20` | sample questions with answers — nothing saved |
| `POST /api/v1/admin/generator/rotate` | how many were archived, generated and pruned |

**Settings:** `QUESTION_GEN_ENABLED`, `QUESTION_GEN_POOL_SIZE`,
`QUESTION_GEN_ROTATE_HOURS`, `QUESTION_GEN_PER_USER`.

**Uses:** `questiongen` (to invent), `questions` (to store and to refresh the bank).
