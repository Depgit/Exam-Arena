# practice

**Job:** solo practice — pick a category and difficulty, answer questions,
get instant feedback. Never affects rating.

| Endpoint | Gives back |
|---|---|
| `POST /api/v1/practice/start` | session id + questions (answers removed, options shuffled) |
| `POST /api/v1/practice/{id}/answer` | correct or not, plus the explanation |
| `POST /api/v1/practice/{id}/end` | — |
| `GET /api/v1/practice/{id}` | the session so far |

**Uses:** `questions` (random questions, `PlayerOptions`), `models`.
