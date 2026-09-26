# Exam Arena — Frontend

React + Vite frontend for the Exam Arena backend, built from the API docs and LLD you provided.

## What's included

- **Auth**: register/login, JWT stored in `localStorage`, auto-attached to every request.
- **Role-based routing**: after login, `user` accounts land on `/app/*`, `admin` accounts land on `/admin/*`. Each area is route-guarded — a `user` can't reach `/admin/*` and vice versa.
- **User side** (`/app/*`):
  - Dashboard — subjects + your career stats
  - Ranked matchmaking — join/leave queue, live queue depth, auto-redirect into the match on `match_start`
  - Friend match — create a room (get a code) or join one by code
  - Live match — WebSocket-driven gameplay: questions, answer submission, live scoreboard, countdown timer, results screen
  - Practice mode — solo questions with instant right/wrong feedback and explanations
  - Leaderboard — by exam category
  - Profile — ratings + full statistics
- **Admin side** (`/admin/*`):
  - Overview
  - Create question (options editor, single/multi/integer types) → create as draft → publish
  - System stats

## One WebSocket connection

The backend enforces a single WS session per user. `WebSocketProvider` (`src/context/WebSocketContext.jsx`) opens exactly one socket app-wide once you're logged in, auto-reconnects on drop, and exposes a `subscribe(type, callback)` pub/sub so any page can listen for `match_start`, `score_update`, `time_update`, `match_end`, `match_failed`, `error` without opening a second connection.

## Setup

```bash
npm install
cp .env.example .env   # point at your backend if it's not on localhost:8080
npm run dev
```

Edit `.env`:
```
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
```

## Notes / things to wire up as your backend evolves

- **Admin question creation** requires a `topic_id` UUID per the API doc, but no "list topics" endpoint exists yet — the form takes it as a free-text UUID field. Once a `/api/v1/topics` (or similar) endpoint exists, swap that input for a `<select>`.
- `GET /api/v1/users/{id}/matches` is documented as a stub, so match history isn't wired into the Profile page yet.
- Admin accounts aren't created via the register form (the API doesn't expose a role field there) — seed/promote admin users on the backend, then just log in with those credentials here.
