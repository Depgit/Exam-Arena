
# Complete API Catalog - Exam Arena

## REST APIs (Complete)

---

### Authentication

| Method | Endpoint                         | Description               | Auth Required      |
| ------ | -------------------------------- | ------------------------- | ------------------ |
| POST   | /api/v1/auth/register            | Register user             | ❌                 |
| POST   | /api/v1/auth/login               | Login                     | ❌                 |
| POST   | /api/v1/auth/refresh             | Refresh access token      | ❌ (refresh token) |
| POST   | /api/v1/auth/logout              | Logout + invalidate token | ✅                 |
| GET    | /api/v1/auth/me                  | Current user info         | ✅                 |
| POST   | /api/v1/auth/change-password     | Change password           | ✅                 |
| POST   | /api/v1/auth/forgot-password     | Send reset email          | ❌                 |
| POST   | /api/v1/auth/reset-password      | Reset with token          | ❌                 |
| POST   | /api/v1/auth/verify-email        | Verify email OTP          | ❌                 |
| POST   | /api/v1/auth/resend-verification | Resend verify email       | ✅                 |

---

### Users

| Method | Endpoint                        | Description          | Auth Required |
| ------ | ------------------------------- | -------------------- | ------------- |
| GET    | /api/v1/users/me                | My full profile      | ✅            |
| PATCH  | /api/v1/users/me                | Update profile       | ✅            |
| POST   | /api/v1/users/me/avatar         | Upload avatar        | ✅            |
| DELETE | /api/v1/users/me/avatar         | Remove avatar        | ✅            |
| GET    | /api/v1/users/{username}        | Public profile       | ✅            |
| GET    | /api/v1/users                   | Search users         | ✅            |
| GET    | /api/v1/users/me/stats          | My overall stats     | ✅            |
| GET    | /api/v1/users/me/stats/category | Stats per category   | ✅            |
| GET    | /api/v1/users/me/elo-history    | ELO graph data       | ✅            |
| GET    | /api/v1/users/{username}/stats  | Public user stats    | ✅            |
| GET    | /api/v1/users/me/activity       | Recent activity feed | ✅            |
| GET    | /api/v1/users/me/achievements   | My achievements      | ✅            |

---

### Subjects & Sections

| Method | Endpoint                       | Description          | Auth Required |
| ------ | ------------------------------ | -------------------- | ------------- |
| GET    | /api/v1/subjects               | List all subjects    | ✅            |
| GET    | /api/v1/subjects/{id}          | Subject detail       | ✅            |
| GET    | /api/v1/subjects/{id}/sections | Sections in subject  | ✅            |
| GET    | /api/v1/sections/{id}          | Section detail       | ✅            |
| GET    | /api/v1/subjects/{id}/stats    | My stats for subject | ✅            |
| GET    | /api/v1/sections/{id}/stats    | My stats for section | ✅            |

---

### Practice

| Method | Endpoint                                | Description                     | Auth Required |
| ------ | --------------------------------------- | ------------------------------- | ------------- |
| POST   | /api/v1/practice/sessions               | Create session                  | ✅            |
| POST   | /api/v1/practice/sessions/{id}/start    | Start session                   | ✅            |
| POST   | /api/v1/practice/sessions/{id}/answer   | Submit answer                   | ✅            |
| POST   | /api/v1/practice/sessions/{id}/skip     | Skip question                   | ✅            |
| POST   | /api/v1/practice/sessions/{id}/bookmark | Bookmark question               | ✅            |
| POST   | /api/v1/practice/sessions/{id}/pause    | Pause session                   | ✅            |
| POST   | /api/v1/practice/sessions/{id}/resume   | Resume session                  | ✅            |
| POST   | /api/v1/practice/sessions/{id}/finish   | Finish early                    | ✅            |
| GET    | /api/v1/practice/sessions/{id}          | Session state                   | ✅            |
| GET    | /api/v1/practice/sessions/{id}/result   | Final result                    | ✅            |
| GET    | /api/v1/practice/sessions/{id}/review   | Full review                     | ✅            |
| GET    | /api/v1/practice/history                | All past sessions               | ✅            |
| GET    | /api/v1/practice/history/summary        | Aggregated summary              | ✅            |
| GET    | /api/v1/practice/bookmarks              | Bookmarked questions            | ✅            |
| DELETE | /api/v1/practice/bookmarks/{questionId} | Remove bookmark                 | ✅            |
| GET    | /api/v1/practice/wrong-answers          | Questions I got wrong           | ✅            |
| POST   | /api/v1/practice/sessions               | Create retry session from wrong | ✅            |

---

### Arena (Matchmaking & Contest)

| Method | Endpoint                           | Description              | Auth Required |
| ------ | ---------------------------------- | ------------------------ | ------------- |
| GET    | /api/v1/arena/modes                | Available game modes     | ✅            |
| GET    | /api/v1/arena/queue/status         | My current queue status  | ✅            |
| POST   | /api/v1/arena/matches/{id}/rematch | Request rematch          | ✅            |
| GET    | /api/v1/arena/matches/{id}         | Match detail             | ✅            |
| GET    | /api/v1/arena/matches/{id}/review  | Full match review        | ✅            |
| GET    | /api/v1/arena/history              | My match history         | ✅            |
| GET    | /api/v1/arena/history/summary      | Win/loss summary         | ✅            |
| GET    | /api/v1/arena/active               | My active match (if any) | ✅            |

---

### Friends

| Method | Endpoint                              | Description          | Auth Required |
| ------ | ------------------------------------- | -------------------- | ------------- |
| GET    | /api/v1/friends                       | My friends list      | ✅            |
| GET    | /api/v1/friends/online                | Online friends       | ✅            |
| GET    | /api/v1/friends/requests              | Incoming requests    | ✅            |
| GET    | /api/v1/friends/requests/sent         | Sent requests        | ✅            |
| POST   | /api/v1/friends/requests              | Send friend request  | ✅            |
| POST   | /api/v1/friends/requests/{id}/accept  | Accept request       | ✅            |
| POST   | /api/v1/friends/requests/{id}/decline | Decline request      | ✅            |
| DELETE | /api/v1/friends/{userId}              | Remove friend        | ✅            |
| POST   | /api/v1/friends/{userId}/block        | Block user           | ✅            |
| DELETE | /api/v1/friends/{userId}/block        | Unblock user         | ✅            |
| GET    | /api/v1/friends/blocked               | Blocked users list   | ✅            |
| POST   | /api/v1/friends/{userId}/challenge    | Send arena challenge | ✅            |

---

### Leaderboard

| Method | Endpoint                                | Description         | Auth Required |
| ------ | --------------------------------------- | ------------------- | ------------- |
| GET    | /api/v1/leaderboard/global              | Global ELO ranking  | ✅            |
| GET    | /api/v1/leaderboard/subject/{subjectId} | Subject leaderboard | ✅            |
| GET    | /api/v1/leaderboard/friends             | Friends leaderboard | ✅            |
| GET    | /api/v1/leaderboard/weekly              | Weekly ranking      | ✅            |
| GET    | /api/v1/leaderboard/monthly             | Monthly ranking     | ✅            |
| GET    | /api/v1/leaderboard/me/rank             | My rank info        | ✅            |

---

### Notifications

| Method | Endpoint                           | Description         | Auth Required |
| ------ | ---------------------------------- | ------------------- | ------------- |
| GET    | /api/v1/notifications              | All notifications   | ✅            |
| GET    | /api/v1/notifications/unread/count | Unread count        | ✅            |
| POST   | /api/v1/notifications/{id}/read    | Mark one read       | ✅            |
| POST   | /api/v1/notifications/read-all     | Mark all read       | ✅            |
| DELETE | /api/v1/notifications/{id}         | Delete notification | ✅            |
| GET    | /api/v1/notifications/preferences  | My preferences      | ✅            |
| PATCH  | /api/v1/notifications/preferences  | Update preferences  | ✅            |

---

### Tournaments *(Future)*

| Method | Endpoint                             | Description        | Auth Required |
| ------ | ------------------------------------ | ------------------ | ------------- |
| GET    | /api/v1/tournaments                  | List tournaments   | ✅            |
| GET    | /api/v1/tournaments/{id}             | Tournament detail  | ✅            |
| POST   | /api/v1/tournaments/{id}/register    | Register           | ✅            |
| GET    | /api/v1/tournaments/{id}/bracket     | Bracket state      | ✅            |
| GET    | /api/v1/tournaments/{id}/leaderboard | Tournament ranking | ✅            |
| GET    | /api/v1/tournaments/me               | My tournaments     | ✅            |

---

### Admin *(Internal)*

| Method | Endpoint                           | Description           |
| ------ | ---------------------------------- | --------------------- |
| POST   | /api/v1/admin/questions            | Create question       |
| PATCH  | /api/v1/admin/questions/{id}       | Update question       |
| DELETE | /api/v1/admin/questions/{id}       | Delete question       |
| POST   | /api/v1/admin/questions/bulk       | Bulk import           |
| GET    | /api/v1/admin/questions            | List all with filters |
| GET    | /api/v1/admin/questions/{id}/stats | Question analytics    |
| GET    | /api/v1/admin/users                | List users            |
| PATCH  | /api/v1/admin/users/{id}/ban       | Ban user              |
| GET    | /api/v1/admin/matches              | All matches log       |
| GET    | /api/v1/admin/stats/platform       | Platform metrics      |

---

---

## WebSocket API

### Connection

```
wss://api.examarena.com/ws?token=<JWT>
```

---

### Message Envelope

```
Every message in both directions uses this shape:

{
  "type":    "string",     ← event name
  "payload": {},           ← event data (varies per type)
  "req_id":  "string"      ← optional, client sets for request tracking
}
```

---

### Arena WebSocket Events

#### CLIENT → SERVER (you send these)

| Type                        | When to Send                    | Payload                                                  |
| --------------------------- | ------------------------------- | -------------------------------------------------------- |
| `arena.queue.join`        | User clicks Find Match          | `{ subject_id, section_id, time_control, match_type }` |
| `arena.queue.leave`       | User cancels search             | `{}`                                                   |
| `arena.match.ready`       | User confirms ready (countdown) | `{ match_id }`                                         |
| `arena.match.answer`      | User picks an option            | `{ match_id, question_id, option, time_ms }`           |
| `arena.match.skip`        | User skips question             | `{ match_id, question_id }`                            |
| `arena.challenge.send`    | Challenge a friend              | `{ to_user_id, subject_id, time_control }`             |
| `arena.challenge.accept`  | Accept challenge                | `{ challenge_id }`                                     |
| `arena.challenge.decline` | Decline challenge               | `{ challenge_id }`                                     |
| `arena.rematch.request`   | Request rematch after match     | `{ match_id }`                                         |
| `arena.rematch.accept`    | Accept rematch                  | `{ match_id }`                                         |
| `arena.rematch.decline`   | Decline rematch                 | `{ match_id }`                                         |

---

#### SERVER → CLIENT (you receive these)

**Queue Events**

| Type                   | When                         | Payload                              |
| ---------------------- | ---------------------------- | ------------------------------------ |
| `arena.queue.joined` | Confirmed in queue           | `{ position, estimated_wait_sec }` |
| `arena.queue.tick`   | Every 5s while waiting       | `{ wait_sec, elo_range_current }`  |
| `arena.queue.left`   | Confirmed removed from queue | `{}`                               |
| `arena.queue.failed` | Could not join queue         | `{ reason }`                       |

**Match Lifecycle**

| Type                                  | When                         | Payload                                                                                                       |
| ------------------------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `arena.match.found`                 | Opponent found               | `{ match_id, opponent, subject, time_control, countdown_sec }`                                              |
| `arena.match.opponent_ready`        | Opponent confirmed ready     | `{ match_id }`                                                                                              |
| `arena.match.cancelled`             | Opponent did not ready up    | `{ match_id, reason }`                                                                                      |
| `arena.match.started`               | Match officially begins      | `{ match_id, started_at }`                                                                                  |
| `arena.match.question`              | New question pushed          | `{ match_id, question, question_number, total, time_limit_sec, points }`                                    |
| `arena.match.tick`                  | Every 1s during question     | `{ match_id, question_id, time_remaining_sec }`                                                             |
| `arena.match.answer_result`         | Your answer graded           | `{ match_id, question_id, is_correct, correct_option, points_earned, time_ms, your_score, opponent_score }` |
| `arena.match.opponent_answered`     | Opponent answered            | `{ match_id, question_id, is_correct, time_ms }`                                                            |
| `arena.match.question_timeout`      | Time ran out                 | `{ match_id, question_id, correct_option }`                                                                 |
| `arena.match.next_question_in`      | Buffer between questions     | `{ match_id, next_in_sec }`                                                                                 |
| `arena.match.opponent_disconnected` | Opponent lost connection     | `{ match_id, grace_period_sec }`                                                                            |
| `arena.match.opponent_reconnected`  | Opponent came back           | `{ match_id }`                                                                                              |
| `arena.match.abandoned`             | Opponent gave up / timed out | `{ match_id, winner_id }`                                                                                   |
| `arena.match.ended`                 | Match complete               | `{ match_id, winner, your_score, opponent_score, elo_change, new_elo, breakdown }`                          |

**Challenge Events**

| Type                         | When                    | Payload                                                                |
| ---------------------------- | ----------------------- | ---------------------------------------------------------------------- |
| `arena.challenge.received` | Someone challenged you  | `{ challenge_id, from_user, subject, time_control, expires_in_sec }` |
| `arena.challenge.accepted` | Your challenge accepted | `{ challenge_id, match_id }`                                         |
| `arena.challenge.declined` | Your challenge declined | `{ challenge_id }`                                                   |
| `arena.challenge.expired`  | No response in time     | `{ challenge_id }`                                                   |
| `arena.rematch.received`   | Opponent wants rematch  | `{ match_id }`                                                       |
| `arena.rematch.accepted`   | Rematch confirmed       | `{ new_match_id }`                                                   |
| `arena.rematch.declined`   | Rematch refused         | `{ match_id }`                                                       |
| `arena.rematch.expired`    | Rematch window closed   | `{ match_id }`                                                       |

---

### Presence WebSocket Events

#### CLIENT → SERVER

| Type              | Payload                        |
| ----------------- | ------------------------------ |
| `presence.ping` | `{}` (keep alive, every 30s) |

#### SERVER → CLIENT

| Type                           | When                | Payload                   |
| ------------------------------ | ------------------- | ------------------------- |
| `presence.friend_online`     | Friend connects     | `{ user_id, username }` |
| `presence.friend_offline`    | Friend disconnects  | `{ user_id, username }` |
| `presence.friend_in_match`   | Friend enters match | `{ user_id }`           |
| `presence.friend_left_match` | Friend exits match  | `{ user_id }`           |

---

### Notification WebSocket Events

#### SERVER → CLIENT only

| Type                          | When                         | Payload                                         |
| ----------------------------- | ---------------------------- | ----------------------------------------------- |
| `notification.new`          | Any new notification arrives | `{ id, type, title, body, data, created_at }` |
| `notification.unread_count` | Count changes                | `{ count }`                                   |

---

### System WebSocket Events

#### SERVER → CLIENT only

| Type                        | When                       | Payload                        |
| --------------------------- | -------------------------- | ------------------------------ |
| `system.connected`        | Connection established     | `{ user_id, server_time }`   |
| `system.error`            | Any server error           | `{ code, message, req_id? }` |
| `system.maintenance`      | Scheduled downtime warning | `{ starts_at, message }`     |
| `system.force_disconnect` | Server kicks client        | `{ reason }`                 |
| `system.pong`             | Response to client ping    | `{ server_time }`            |

---

### WebSocket Error Codes

```
4001 - Unauthorized (bad/expired token)
4002 - Already in queue
4003 - Already in match
4004 - Match not found
4005 - Question already answered
4006 - Invalid option (not a/b/c/d)
4007 - Time expired for question
4008 - Opponent not found
4009 - Challenge expired
4010 - Cannot challenge yourself
4011 - User is blocked
4012 - Rate limit exceeded
4013 - Match already ended
5001 - Internal server error
```

---

### Full WebSocket Session Flow

```
CLIENT                          SERVER
  │                               │
  │── WSS connect (?token=xxx) ──▶│
  │◀── system.connected ──────────│
  │                               │
  │── arena.queue.join ──────────▶│
  │◀── arena.queue.joined ────────│
  │◀── arena.queue.tick (x N) ────│  ← every 5s while waiting
  │                               │
  │◀── arena.match.found ─────────│
  │── arena.match.ready ─────────▶│
  │◀── arena.match.started ───────│
  │                               │
  │◀── arena.match.question ──────│  ← Q1 pushed
  │◀── arena.match.tick (x N) ────│  ← every 1s
  │── arena.match.answer ────────▶│
  │◀── arena.match.answer_result ─│
  │◀── arena.match.opponent_answered│
  │◀── arena.match.next_question_in│ ← 2s buffer
  │◀── arena.match.question ──────│  ← Q2 pushed
  │         ... repeats ...       │
  │◀── arena.match.ended ─────────│
  │                               │
  │── arena.rematch.request ─────▶│
  │◀── arena.rematch.received ────│  (opponent gets this)
  │◀── arena.rematch.accepted ────│
```
