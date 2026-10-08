# realtime

**Job:** keep a live WebSocket connection to every online player and move
messages in both directions.

| Piece | Job |
|---|---|
| `Hub` | Knows who is online; `SendToUser(id, msg)`, `IsOnline`, `OnlineCount`; routes incoming messages by `type` to the handler registered for it |
| `Client` | One player's connection: reads and writes, with ping/pong keep-alive |
| `ConnectHandler` | `GET /wss?token=…` — checks the login token and opens the connection; answers `ping` |
| `Message` | `{"type": "...", "payload": {...}}` — the shape of every message |

**One session per player:** a new connection for the same user closes the
old one.

**Features register their own message types** — e.g. `match` registers
`submit_answer`, `accept_match`, `decline_match`.

**Uses:** `tokens`.
