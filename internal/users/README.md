# users

**Job:** everything stored about a player — account, ratings, statistics.

| Endpoint | Gives back |
|---|---|
| `GET /api/v1/users/{id}` | profile + rating per category |
| `GET /api/v1/users/{id}/stats` | wins/losses/accuracy per category |
| `GET /api/v1/users/{id}/matches` | recent matches |

`Store` is used by other features to read and update players: create
accounts, guests and bots, ratings (`EnsureRating`, `UpdateRating`), stats
(`UpdateStatistics` with `OutcomeWin/Loss/Draw`), and deleting stale guests.

**Uses:** `models`, `platform/respond`.
