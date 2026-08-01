
# Query Design

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the primary database queries used by Exam Arena.

Goals

- Identify read/write patterns
- Design indexes before implementation
- Estimate query frequency
- Detect future scaling bottlenecks

The database should be optimized for user behavior,
not table structure.

---

# 2. Query Categories

High Frequency

- Login
- Matchmaking
- Question Selection
- Submit Answer
- Leaderboard
- Statistics

Medium Frequency

- Practice History
- Match History
- Profile

Low Frequency

- Admin Dashboard
- Question CRUD
- Reports

---

# 3. Authentication

## Login

Business Goal

Authenticate user.

Query

```

SELECT id, password_hash
FROM users
WHERE email = ?

```

Expected Result

One row

Frequency

Very High

Required Index

```

users(email)

```

Target

<10 ms

---

## Load User

```

SELECT *
FROM users
WHERE id = ?

```

Index

Primary Key

---

# 4. Matchmaking

## Find Player Rating

```

SELECT current_rating
FROM ratings
WHERE user_id = ?

```

Frequency

High

---

## Future Redis

When Redis exists

Rating lookup should come from cache.

Database becomes fallback.

---

# 5. Question Selection

Business Goal

Find fair questions.

Example

```

SELECT id
FROM questions
WHERE topic_id = ?
AND difficulty = ?
AND status = 'published'
ORDER BY RANDOM()
LIMIT 10;

```

Problem

ORDER BY RANDOM() becomes slow.

---

Future Strategy

Use

- Random IDs
- Pre-generated pools
- Reservoir sampling

Never use ORDER BY RANDOM() on millions of rows.

---

Index

(topic_id, difficulty, status)

---

# 6. Match Creation

Insert

```

INSERT INTO matches (...)

```

Immediately

Insert

Players

Questions

Single transaction.

---

# 7. Load Match

```

SELECT *
FROM matches
WHERE id = ?

```

Expected

One row

---

## Load Match Players

```

SELECT *
FROM match_players
WHERE match_id = ?

```

Expected

Two rows

---

## Load Match Questions

```

SELECT *
FROM match_questions
WHERE match_id = ?
ORDER BY sequence

```

---

# 8. Submit Answer

Business Goal

Save answer quickly.

```

INSERT INTO match_answers (...)

```

Immediately afterwards

Update in-memory match state.

Future

Database write can be asynchronous.

---

# 9. Match History

Business Goal

Display last matches.

```

SELECT *

FROM matches

JOIN match_players

ON ...

WHERE user_id = ?

ORDER BY started_at DESC

LIMIT 20

```

Index

(user_id, started_at DESC)

Target

<50 ms

---

# 10. Rating History

```

SELECT *

FROM rating_history

WHERE user_id = ?

ORDER BY created_at DESC

LIMIT 20

```

Index

(user_id, created_at DESC)

---

# 11. User Statistics

```

SELECT *

FROM user_statistics

WHERE user_id = ?

```

Primary Key lookup.

---

# 12. Topic Statistics

```

SELECT *

FROM topic_statistics

WHERE user_id = ?

```

Frequency

Medium

---

# 13. Leaderboard

Business Goal

Top players.

```

SELECT user_id,
current_rating

FROM ratings

ORDER BY current_rating DESC

LIMIT 100

```

Index

(current_rating DESC)

Future

Redis cache.

---

## Player Rank

Future

Window function

or cached ranking.

Avoid counting every player.

---

# 14. Practice Questions

```

SELECT *

FROM questions

WHERE topic_id = ?

LIMIT 20

```

Simple indexed lookup.

---

# 15. Friend List

```

SELECT *

FROM friendships

WHERE requester_id = ?

AND status='accepted'

```

Index

(requester_id, status)

---

# 16. Notifications

Unread

```

SELECT *

FROM notifications

WHERE user_id = ?

AND read=false

ORDER BY created_at DESC

```

Index

(user_id, read)

---

# 17. Admin

Question Search

```

SELECT *

FROM questions

WHERE status='pending'

```

Index

(status)

---

Reports

```

SELECT *

FROM reports

WHERE status='open'

```

Index

(status)

---

# 18. Heavy Queries

These require monitoring.

Leaderboard

Question search

Topic statistics

Large match history

Analytics

These may require caching later.

---

# 19. Query Performance Targets

| Query              | Target  |
| ------------------ | ------- |
| Login              | <10 ms  |
| Load Profile       | <20 ms  |
| Question Selection | <100 ms |
| Submit Answer      | <20 ms  |
| Match Creation     | <50 ms  |
| Match History      | <50 ms  |
| Leaderboard        | <100 ms |
| Statistics         | <50 ms  |

---

# 20. Query Evolution

Version 1

Simple SQL

↓

Version 2

Optimized indexes

↓

Version 3

Redis cache

↓

Version 4

Read replicas

↓

Version 5

Materialized views

Only optimize after measurement.

---

# 21. Anti-Patterns

Avoid

SELECT *

on large tables.

Avoid

N+1 queries.

Avoid

ORDER BY RANDOM()

Avoid

Loading unnecessary columns.

Avoid

Multiple queries inside loops.

Avoid

Missing indexes on WHERE clauses.

---

# 22. Monitoring

Track

Slow queries

Execution time

Rows scanned

Rows returned

Index usage

Lock waits

Deadlocks

Review the slow query log regularly.

---

# 23. Caching Strategy

Do NOT cache initially.

Introduce cache only for

- Leaderboard
- Question pools
- User sessions
- Frequently accessed statistics

Database remains the source of truth.

---

# 24. Design Philosophy

Queries define indexes.

Indexes influence schema.

Business requirements define queries.

Always measure before optimizing.
