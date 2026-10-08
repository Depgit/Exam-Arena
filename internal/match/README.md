# match

**Job:** run a live match from "both players found" to "results and new ratings".

| File | Job |
|---|---|
| `accept.go` | The 30-second accept window: `match_found` → both `accept_match` → start. Decline/timeout/disconnect puts the other player back in the queue, keeping their place |
| `service.go` | Start a match (questions from the in-memory bank, rows in Postgres, live state in memory), take answers (`SubmitAnswer`, idempotent), timer, end the match early once everyone has answered, results, rating updates |
| `elo.go` | `CalculateElo` — new ratings after a match |
| `bots.go` | Bot opponents (Rookie / Ace / Master) for when nobody is searching; bot matches are unrated |
| `game_messages.go` | WebSocket messages: `submit_answer`, `accept_match`, `decline_match` |
| `handler.go` | `GET /api/v1/matches/{id}`, private rooms (`/matches/friend`, `/friend/join`), `POST /matches/bot` |
| `store.go` | SQL for matches, players, answers |

**Messages it sends:** `match_found`, `opponent_accepted`, `match_cancelled`,
`match_start`, `score_update`, `time_update`, `match_end`, `match_failed`.

**Live state is in memory** (`cache` key `live_match:<id>`), so a server
restart loses matches in progress.

**Uses:** `matchmaking` (queue), `questions` (bank, `ForPlayers`), `users`
(ratings, stats, bot accounts), `platform/realtime`, `platform/cache`.
