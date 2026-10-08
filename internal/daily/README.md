# daily

**Job:** the daily challenge — the same 10 questions for everyone, one
attempt, 3 minutes. Most correct wins; ties go to the fastest.

| Endpoint | Gives back |
|---|---|
| `GET /api/v1/daily-challenge` | today's info, your attempt (if any), top players |
| `POST /api/v1/daily-challenge/start` | the questions and your deadline (resuming keeps the same clock and order) |
| `POST /api/v1/daily-challenge/submit` | your score, rank and an answer review |

A new challenge starts at midnight IST. Today's challenge and questions are
kept in memory once loaded.

**Uses:** `questions` (pick and load questions, `ForPlayers`), `models`.
