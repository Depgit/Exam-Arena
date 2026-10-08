# matchmaking

**Job:** hold players who are looking for a match, and pair them.

| Piece | Job |
|---|---|
| `Queue` (`queue.go`) | In-memory waiting players, one pool per (category, ranked/arena). `ClaimPair` removes two players atomically so nobody is matched twice |
| `Engine` (`engine.go`) | Every 2 s: **Arena** pairs whoever waited longest; **Ranked** pairs the closest ratings within a tolerance of ±100, growing ±5 per second waited, up to ±400. Hands each pair to a callback (`match.Service.ProposeMatch`) |
| `Service` (`service.go`) | Join / leave the queue; `RestoreFromDB` puts waiting players back after a restart |
| `Store` (`store.go`) | Postgres copy of the queue, only for restart recovery |
| `Handler` (`handler.go`) | `POST/DELETE /api/v1/matches/queue`, `GET /api/v1/matches/queue/stats` |

**Does not know about matches** — it only produces pairs. That's what keeps
`match` → `matchmaking` a one-way dependency.

**Uses:** `users` (rating), `models`, `platform/*`.
