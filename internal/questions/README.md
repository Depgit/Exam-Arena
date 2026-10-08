# questions

**Job:** everything about questions — categories, topics, the questions
themselves, getting them ready for players, and players reporting bad ones.

| Piece | Job |
|---|---|
| `Store` (`store.go`) | SQL for questions & categories; `store_generated.go` adds the generated-question pool queries |
| `TopicStore`, `FlagStore` | SQL for topics and question reports |
| `Bank` (`bank.go`) | All published questions **in memory**, refreshed every 15 min, so a match starts with no DB call; `GetRandom(category, n)` |
| `ForPlayers(questions, key)` (`for_player.go`) | Removes the answers and shuffles each question's options (same key → same order, so a reload doesn't reshuffle) |
| `CategoryHandler` | `GET /api/v1/subjects` — active categories |
| `TopicHandler` | topic list/details; admins create topics |
| `FlagHandler` | players report a question; admins review reports |
| `ErrNotEnoughQuestions` | returned when a category can't fill a match |

**Uses:** `models`, `platform/database`, `platform/middleware`, `platform/respond`.
