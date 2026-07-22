
# Battle Protocol

Project: Exam Arena

Module: Realtime Battle

Version: 1.0

Status: Draft

---

# 1. Purpose

This document defines the realtime protocol between clients and the Battle Service.

The protocol is built on top of WebSocket.

Goals

- Low latency
- Server authoritative
- Recoverable after reconnect
- Deterministic
- Cheat resistant

---

# 2. Connection Flow

Player

↓

WebSocket Connect

↓

Authenticate

↓

Join Battle Channel

↓

Receive Battle Snapshot

↓

Receive Live Events

↓

Disconnect

↓

Reconnect (if necessary)

---

# 3. WebSocket Endpoint

GET

```
/ws
```

Authentication

JWT

Example

```
ws://localhost:8080/ws?token=<JWT>
```

Production

```
wss://api.examarena.com/ws
```

---

# 4. General Message Format

Every message follows the same structure.

Server → Client

```json
{
    "type":"QUESTION_STARTED",
    "battleId":"uuid",
    "timestamp":"2026-07-22T12:30:15Z",
    "payload":{}
}
```

Client → Server

```json
{
    "type":"SUBMIT_ANSWER",
    "payload":{}
}
```

---

# 5. Message Types

Server

CONNECTED

BATTLE_CREATED

COUNTDOWN_STARTED

QUESTION_STARTED

ANSWER_RESULT

QUESTION_FINISHED

BATTLE_FINISHED

PLAYER_DISCONNECTED

PLAYER_RECONNECTED

HEARTBEAT

ERROR

Client

JOIN_BATTLE

SUBMIT_ANSWER

PING

LEAVE_BATTLE

---

# 6. Connection

Server

```json
{
    "type":"CONNECTED",
    "payload":{
        "userId":"uuid",
        "serverTime":"2026-07-22T12:30:00Z"
    }
}
```

---

# 7. Join Battle

Client

```json
{
    "type":"JOIN_BATTLE",
    "payload":{
        "battleId":"uuid"
    }
}
```

Business Rules

- Player must belong to battle.
- Connection becomes bound to battle.
- Server sends current snapshot.

---

# 8. Battle Snapshot

Returned immediately after JOIN_BATTLE.

```json
{
    "type":"BATTLE_SNAPSHOT",
    "payload":{
        "eventId": "01K0M9Y...",
        "sequence": 42,
        "battleId":"uuid",
        "state":"COUNTDOWN",
        "questionNumber":0,
        "remainingTimeMs":3000,
        "yourScore":0,
        "opponentScore":0
    }
}
```

---

# 9. Countdown

Server

```json
{
    "type":"COUNTDOWN_STARTED",
    "payload":{
        "seconds":3
    }
}
```

Client

Displays

```
3

2

1

GO
```

No client-side countdown logic determines game state. The server is authoritative.

---

# 10. Question Started

Server

```json
{
    "type":"QUESTION_STARTED",
    "payload":{
        "questionId":"uuid",
        "questionNumber":1,
        "totalQuestions":10,
        "statement":"What is 25% of 200?",
        "options":[
            {
                "id":"1",
                "text":"25"
            },
            {
                "id":"2",
                "text":"50"
            },
            {
                "id":"3",
                "text":"100"
            },
            {
                "id":"4",
                "text":"75"
            }
        ],
        "timeLimitMs":15000
    }
}
```

---

# 11. Submit Answer

Client

```json
{
    "type":"SUBMIT_ANSWER",
    "payload":{
        "questionId":"uuid",
        "optionId":"2"
    }
}
```

Business Rules

Player can answer only once.

Late answers ignored.

Duplicate answers ignored.

Server timestamp is authoritative.

---

# 12. Answer Result

Server

```json
{
    "type":"ANSWER_RESULT",
    "payload":{
        "correct":true,
        "score":1,
        "totalScore":5
    }
}
```

The correct answer is not revealed yet.

---

# 13. Question Finished

Server

```json
{
    "type":"QUESTION_FINISHED",
    "payload":{
        "correctOptionId":"2",
        "yourScore":5,
        "opponentScore":4
    }
}
```

Immediately followed by the next question or battle completion.

---

# 14. Battle Finished

Server

```json
{
    "type":"BATTLE_FINISHED",
    "payload":{
        "winner":"playerA",
        "myScore":8,
        "opponentScore":7,
        "ratingChange":18
    }
}
```

---

# 15. Heartbeat

Every

```
20 seconds
```

Server

```json
{
    "type":"HEARTBEAT"
}
```

Client

```json
{
    "type":"PING"
}
```

If heartbeat fails three consecutive times, the connection is considered lost.

---

# 16. Disconnect

Server

Marks player

```
DISCONNECTED
```

Reconnect window

```
60 seconds
```

If reconnect succeeds

Resume battle.

Otherwise

Player forfeits.

---

# 17. Reconnect

Client reconnects

↓

Authenticate

↓

JOIN_BATTLE

↓

Receive BATTLE_SNAPSHOT

↓

Continue battle

The server restores the latest state.

---

# 18. Error Messages

Example

```json
{
    "type":"ERROR",
    "payload":{
        "code":"INVALID_OPTION",
        "message":"Option does not belong to question."
    }
}
```

---

# 19. Security Rules

The client cannot

- change timer
- change score
- choose next question
- reveal answers
- modify rating
- modify battle state

The server validates every action.

---

# 20. Protocol Rules

Messages must be processed in the order received.

The server ignores duplicate submissions.

Only the Battle Engine changes game state.

Clients are presentation layers.

---

# 21. Future Events

PLAYER_TYPING

EMOJI

CHAT_MESSAGE

REMATCH_REQUEST

SPECTATOR_JOINED

TOURNAMENT_UPDATE

ACHIEVEMENT_UNLOCKED
