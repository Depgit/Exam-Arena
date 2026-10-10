# chat

**Job:** the global chat — one room everyone shares.

| Endpoint | Does |
|---|---|
| `GET /api/v1/chat` | the latest 50 messages, oldest first (from memory, no database call) |
| `POST /api/v1/chat` `{"body": "…"}` | post a message (real accounts only — demo guests can read but not post) |

Every new message is pushed live to all connected players as a
`chat_message` WebSocket event.

**Rules** (`service.go`): up to 300 characters; line breaks and control
characters become spaces; at most 5 messages per 10 seconds per player; the
same text can't be repeated within 30 seconds.

**Storage** (`store.go`): messages are saved in `chat_messages` so a restart
doesn't wipe the chat; only the newest 500 are kept. A deleted account takes
its messages with it.

**Uses:** `platform/realtime` (broadcast), `platform/respond`, `platform/middleware`.
