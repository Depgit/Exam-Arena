
# Battle API

Project: Exam Arena
Module: Battle
Version: 1.0
Status: Draft

---

# 1. Purpose

The Battle module manages live multiplayer matches.

Responsibilities

- Battle lifecycle
- Question progression
- Answer validation
- Timer management
- Score calculation
- Winner determination
- ELO updates
- Match history

It does NOT

- Authenticate users
- Match players
- Manage questions

---

# 2. Battle Object

```json
{
    "id":"uuid",
    "status":"PLAYING",
    "subject":"reasoning",
    "timeControl":"2m",
    "questionCount":10,
    "currentQuestion":4
}
```

---

# 3. Battle States

CREATED

WAITING_FOR_PLAYERS

COUNTDOWN

PLAYING

FINISHED

Failure States

PLAYER_DISCONNECTED

ABANDONED

EXPIRED

---

# 4. Get Battle

GET

```
/api/v1/battles/{battleId}
```

Authentication

Required

---

## Purpose

Returns battle information.

Used

- reconnect
- refresh page

---

## Response

```json
{
    "success":true,
    "data":{
        "battleId":"uuid",
        "status":"PLAYING",
        "currentQuestion":4,
        "questionCount":10,
        "remainingTime":42
    }
}
```

---

# 5. Get Battle Result

GET

```
/api/v1/battles/{battleId}/result
```

Authentication

Required

---

## Purpose

Returns final battle result.

---

## Response

```json
{
    "winner":"playerA",
    "myScore":8,
    "opponentScore":6,
    "ratingChange":18
}
```

---

# 6. Battle History

GET

```
/api/v1/battles/history
```

Authentication

Required

---

Query

page

pageSize

subject

---

# 7. Battle Details

GET

```
/api/v1/battles/{battleId}/review
```

Returns

Questions

Correct answers

Your answers

Opponent answers (future)

Time spent

Explanation

---

# 8. Leave Battle

POST

```
/api/v1/battles/{battleId}/leave
```

Business Rules

Leaving counts as surrender.

Opponent immediately wins.

---

# 9. Reconnect

POST

```
/api/v1/battles/{battleId}/reconnect
```

Authentication

Required

---

Purpose

Resume active battle.

Returns

Current battle snapshot.

---

# 10. ELO Update

Performed after battle finishes.

Formula

```
New Rating

=

Old Rating

+

K × (Actual − Expected)
```

Calculated only by server.

---

# 11. Match Completion

When last question finishes

↓

Calculate score

↓

Determine winner

↓

Update ratings

↓

Store history

↓

Publish BattleFinished event

---

# 12. Events

BattleCreated

CountdownStarted

BattleStarted

QuestionStarted

QuestionFinished

PlayerAnswered

BattleFinished

RatingUpdated

---

# 13. Security

Clients cannot

Choose questions

Modify answers

Modify score

Modify timer

Modify winner

Everything is computed by server.

---

# 14. Rate Limits

Reconnect

10/minute

History

60/minute

Result

60/minute

---

# 15. Future

Spectator Mode

Replay

Share Battle

Tournament

Voice Chat

Battle Chat

Rematch
