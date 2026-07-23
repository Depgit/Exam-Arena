# Exam Arena — Schema Design Notes

Companion to `exam_arena_schema.sql`. Read this before you implement anything.

## 1. The core idea: Postgres is truth, Redis is speed

Every table in the schema is designed so **Postgres never depends on Redis
existing**. Redis, when you add it, only ever does two things:

1. **Cache** data that Postgres already has (rebuildable at any time).
2. **Hold ephemeral state** that doesn't need to survive a restart (queues, live timers).

This means you can add or remove Redis at any point without a schema
migration or data-loss risk — if Redis goes down, the app degrades to
"slower" instead of "broken."

## 2. What to move to Redis, in priority order

| Priority | Data | Postgres table today | Redis structure later | Why |
|---|---|---|---|---|
| 1 | Matchmaking queue | `matchmaking_queue` | Sorted set (`ZADD`, score = rating) per `exam_category_id + match_type` | Needs sub-second reads/writes at high concurrency; polling a table doesn't scale past a few hundred concurrent players |
| 2 | Live match state (current question index, per-player timer, connection status) | in-memory on the match's server process today | Redis hash + Pub/Sub, keyed `match:{id}` | Needs to be shared across server instances once you horizontally scale match servers |
| 3 | Leaderboards (global/country/friends/season) | `user_ratings`, `season_ratings` | Sorted set `ZADD leaderboard:{season}:{category} rating user_id` | `ZRANGE`/`ZREVRANK` give O(log N) rank lookups instead of a full `ORDER BY` scan |
| 4 | Auth tokens (password reset, email verify) | `user_auth_tokens` | `SETEX` with TTL | Native expiry, no cron cleanup job needed |
| 5 | Rate limiting / anti-cheat throttling | (not persisted — app-layer only) | `INCR` + `EXPIRE` per user/IP | This was never going to live in Postgres anyway |
| 6 | Session/JWT blacklist | (not persisted) | `SET` with TTL = token remaining life | Needed once you support "logout everywhere" |
| 7 | Hot reads: user profile, published question by id | `users`, `questions` | Cache-aside `GET`/`SETEX` | Pure read accelerator, not required for correctness |

Notice items 1–3 are **structurally present in the schema already** as
normal tables — so today the app works correctly on Postgres alone, and the
migration later is "point the read/write for this table at Redis instead,"
not "redesign the data model."

## 3. Why some choices were made

**UUID primary keys everywhere.** No auto-increment integers. This matters
for three production reasons: (1) IDs can be generated client-side or by
any server before insert, which matters once you have multiple match
servers; (2) UUIDs don't leak growth-rate information (a competitor can't
tell how many users you have from an ID); (3) they map cleanly to Redis key
names (`user:{uuid}`) with zero translation layer.

**`matches` is one table for ranked/friend/arena/tournament**, not four
separate tables. All four modes share >90% of their shape (players,
questions, timer, status). A `match_type` enum plus nullable
`tournament_id`/`room_code` avoids duplicating logic in four places, and
your leaderboard/history/analytics queries don't need `UNION ALL` across
four tables.

**`rating_history` and `rating_history`-style append-only logs.** Current
rating lives in `user_ratings` (fast read), but every change is also logged
immutably. This is what lets you rebuild `user_ratings` or a Redis
leaderboard from scratch if either gets corrupted or you change your Elo
formula — replay the log.

**JSONB for `settings`, `rules`, `criteria`, `payload`.** These are exactly
the fields the PRD says will evolve (tournament rule types, achievement
criteria, notification payloads differ per type). JSONB means adding a new
tournament format or achievement type is a data change, not a migration.

**Soft delete on `users` (`deleted_at`), not on most other tables.**
Users need "delete my account" while keeping match history integrity
(foreign keys don't orphan). Questions, matches, etc. use a `status`/
`archived` field instead where the PRD calls for a lifecycle, and hard
delete elsewhere is fine since nothing else needs to survive a delete.

**Separate `question_options` table instead of a JSON array on
`questions`.** You need to reference "the option the player selected" by ID
from `match_answers` — a normalized FK is more reliable for scoring/
auditing than indexing into a JSON array, and it's what lets you support
`mcq_multiple` and future question types without changing the answers table.

## 4. Indexing notes

- `idx_questions_selection` and `idx_user_ratings_lookup` are the two
  indexes matchmaking will hit hardest — question selection and
  rating-based opponent search. Watch these first under load.
- `match_answers` and `rating_history` are append-only and will grow
  fastest. Consider partitioning both by `created_at` (monthly) once
  you're past ~10M rows — the schema doesn't need to change for this,
  just the table's storage layout.
- Add a `pg_stat_statements`-driven review after your first real load test
  rather than pre-guessing every index; the ones included cover the
  query patterns explicit in the PRD (matchmaking, leaderboards, history).

## 5. Things worth doing before you call this "production ready"

- **Row-Level Security or an application-layer guard** so a user can only
  read their own `match_answers`/`practice_session_questions` rows —
  important since answers include `is_correct`, which must never leak to
  an opponent mid-match.
- **A read replica** for leaderboard/analytics queries once traffic grows,
  so heavy `SELECT`s never compete with match-write latency (the PRD's
  <100ms battle latency target depends on this).
- **Connection pooling** (PgBouncer) — matches generate many short-lived
  connections at high concurrency.
- **Background job runner** (e.g. a queue, not necessarily Redis-based) to
  update `user_statistics` / `user_topic_statistics` asynchronously after
  each match, rather than in the request path.
- **Migration tool** (e.g. `sqlx`, `Prisma Migrate`, `Flyway`, `Alembic`) —
  this file is the *initial* schema; treat every future change as a
  versioned migration from day one.

## 6. Explicitly not modeled (matches PRD's "Out of Scope")

No tables for video content, live classes, marketplace, AI-generated
questions, or chat/voice — consistent with Section 8/13 of the vision and
PRD documents. Adding these later is additive (new tables), not a rework
of what's here.
