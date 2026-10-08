# friends

**Job:** friend lists and challenging a friend to a match.

| Endpoint | Does |
|---|---|
| `GET /api/v1/friends` | friends (with online status), incoming and outgoing requests |
| `POST /api/v1/friends/requests` | send a request by username |
| `POST …/requests/{id}/accept` · `/decline` | answer a request |
| `DELETE /api/v1/friends/{userId}` | remove a friend |
| `POST /api/v1/friends/{userId}/challenge` | challenge an online friend (expires after 10 min) |
| `DELETE /api/v1/friends/challenges/{matchId}` | cancel or decline a challenge |

**Messages it sends:** `friend_request`, `friend_request_accepted`,
`friend_challenge`, `friend_challenge_cancelled`, `friend_challenge_declined`.

**Uses:** `match` (to create the challenge match), `users`, `questions`,
`platform/realtime`, `platform/cache`.
