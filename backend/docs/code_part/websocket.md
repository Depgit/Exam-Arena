
# WebSocket Specification

Project: Exam Arena

Version: 1.0

Status: Draft

Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the realtime communication layer used by Exam Arena.

Responsibilities

- Connection management
- Authentication
- Message routing
- Heartbeats
- Reconnection
- Channel subscription
- Rate limiting

It does NOT define gameplay.

Gameplay is defined in

docs/realtime/battle_protocol.md

---

# 2. Transport

Protocol

WebSocket

Development

ws://localhost:8080/ws

Production

wss://api.examarena.com/ws

---

# 3. Authentication

Authentication occurs immediately after the socket connects.

Connection

↓

AUTHENTICATE

↓

Authenticated

↓

Subscribe Channels

↓

Receive Events

Unauthenticated sockets are disconnected.

---

# 4. Authentication Message

Client → Server

```json
{
    "type":"AUTHENTICATE",
    "requestId":"uuid",
    "payload":{
        "accessToken":"jwt"
    }
}
```

Server

```json
{
    "type":"AUTH_SUCCESS",
    "requestId":"uuid",
    "payload":{
        "userId":"uuid"
    }
}
```

Failure

```json
{
    "type":"AUTH_FAILED",
    "requestId":"uuid",
    "payload":{
        "reason":"INVALID_TOKEN"
    }
}
```

Server immediately closes connection.

---

# 5. Message Envelope

Every message follows the same format.

Client → Server

```json
{
    "type":"SUBMIT_ANSWER",
    "requestId":"uuid",
    "payload":{}
}
```

Server → Client

```json
{
    "type":"QUESTION_STARTED",
    "eventId":"uuid",
    "sequence":104,
    "channel":"battle",
    "timestamp":"2026-07-22T12:00:00Z",
    "payload":{}
}
```

---

# 6. Channels

Each event belongs to one channel.

battle

notifications

friends

system

future

chat

arena

tournament

One WebSocket connection may subscribe to multiple channels.

---

# 7. Subscribe

Client

```json
{
    "type":"SUBSCRIBE",
    "requestId":"uuid",
    "payload":{
        "channels":[
            "notifications",
            "friends"
        ]
    }
}
```

Server

```json
{
    "type":"SUBSCRIBED",
    "requestId":"uuid",
    "payload":{
        "channels":[
            "notifications",
            "friends"
        ]
    }
}
```

Battle channel subscriptions are handled automatically when a battle starts.

---

# 8. Unsubscribe

Client

```json
{
    "type":"UNSUBSCRIBE",
    "requestId":"uuid",
    "payload":{
        "channels":[
            "friends"
        ]
    }
}
```

---

# 9. Heartbeat

Server sends heartbeat every

20 seconds.

Server

```json
{
    "type":"PING"
}
```

Client

```json
{
    "type":"PONG"
}
```

Missing

3 consecutive heartbeats

↓

Connection closed.

---

# 10. Reconnection

Client

↓

Reconnect Socket

↓

Authenticate

↓

Resubscribe Channels

↓

Receive Missed Events

Server uses

sequence

to determine which events the client missed.

---

# 11. Event Ordering

Every event contains

sequence

Example

```
101

102

103

104
```

Client ignores

duplicate

old

out-of-order

events.

---

# 12. Rate Limits

Maximum

50 client messages/second

Exceeded

↓

RATE_LIMIT_EXCEEDED

↓

Connection may be closed.

---

# 13. Request / Response Pattern

Client messages requiring acknowledgement include

requestId

Example

```json
{
    "type":"SUBSCRIBE",
    "requestId":"123",
    "payload":{}
}
```

Server

```json
{
    "type":"ACK",
    "requestId":"123"
}
```

This lets the client match responses to requests.

---

# 14. Error Message

```json
{
    "type":"ERROR",
    "payload":{
        "code":"INVALID_MESSAGE",
        "message":"Unknown event."
    }
}
```

---

# 15. Close Codes

4001

Authentication Failed

4002

Token Expired

4003

Duplicate Connection

4004

Rate Limited

4005

Protocol Error

4006

Server Shutdown

---

# 16. Security Rules

Server is authoritative.

Client commands are validated.

Unknown message types ignored.

Payload size limited to

64 KB

Maximum connection lifetime

24 hours

---

# 17. Logging

Every connection receives

connectionId

Every message logs

connectionId

userId

requestId

eventId

timestamp

---

# 18. Compression

Future

permessage-deflate

Enabled only after performance testing.

---

# 19. Future Features

Binary protocol

Protocol version negotiation

Delta updates

Presence

Typing indicators

Voice

Chat

Live spectators

---

# 20. Design Principles

One socket per authenticated user.

Multiple logical channels.

Server pushes events.

Clients send commands.

The WebSocket layer transports messages only.

Business logic belongs to domain services.
