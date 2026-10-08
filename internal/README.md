# Backend map

Every folder here has **one job**. Read the folder name to know what it does;
open its `README.md` for what you give it and what you get back.

## The folders

| Folder | Its one job |
|---|---|
| `auth/` | Log in, sign up, demo (guest) accounts, login tokens |
| `users/` | Player profiles, ratings, stats, match history |
| `questions/` | Categories, topics, every question (stored + in-memory bank), question reports |
| `questiongen/` | Invents Math & Reasoning questions from templates (pure code, no database) |
| `questionpool/` | Keeps a rotating stock of generated questions in the database |
| `matchmaking/` | The waiting queue and the engine that pairs players |
| `match/` | A live match: accept window, start, answers, timer, results, rating (ELO), bots, private rooms |
| `friends/` | Friend requests and challenging a friend |
| `practice/` | Solo practice sessions |
| `daily/` | The daily challenge (same 10 questions for everyone) |
| `leaderboard/` | Rankings per category |
| `admin/` | Admin console: stats, create/publish questions, category order |
| `models/` | Shared data shapes (User, Question, Match…) — no logic |
| `platform/` | Shared plumbing every feature uses (see `platform/README.md`) |

## Inside a feature folder

Most feature folders hold the same three kinds of file:

| File | Job | Talks to |
|---|---|---|
| `handler.go` | HTTP endpoints: read the request, call the service, write JSON | the service |
| `service.go` | The rules of the feature | stores, other features |
| `store.go` | SQL queries — the only code that touches Postgres | the database |

So a request always flows **handler → service → store → database**.

## Who uses whom

Arrows point at what a folder depends on. Nothing points back up, so there
are no import cycles.

```
friends, admin
      │
    match ───────────────► matchmaking
      │                        │
auth, practice, daily, leaderboard, questionpool ─► questiongen
      │                        │
   users, questions ◄──────────┘
      │
   models
      │
  platform/*   (config, database, cache, respond, middleware, tokens, passwords, realtime)
```

## Where to start reading

1. `cmd/server/main.go` — how everything is assembled, step by step.
2. `cmd/server/routes.go` — every URL, which folder serves it, and whether
   it needs a login or admin.
3. The README of the folder you care about.

## Things to know

- **One server only.** Live matches, the queue and the WebSocket hub live in
  memory, so the backend runs as a single process. Scaling out needs that
  state moved to something shared (e.g. Redis).
- **`/debug/state` and `/debug/queue` are public.** They're handy locally but
  expose who is queued; protect or remove them in production.
