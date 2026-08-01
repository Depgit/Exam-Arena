
# Matchmaking API

Project: Exam Arena
Module: Matchmaking
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

The Matchmaking module pairs players with opponents for real-time competitive matches.

Responsibilities

- Join matchmaking queue
- Leave queue
- Find suitable opponent
- Create match
- Handle queue timeout
- Support reconnects

It does NOT

- Deliver questions
- Calculate scores
- Store match history

Those belong to the Battle module.

---

# 2. Queue Session

Every matchmaking attempt creates one Queue Session.

Queue Session represents one search attempt.

Example

```
User

↓

Queue Session

↓

Searching

↓

Matched

↓

Battle
```

---

# 3. Queue States

```
CREATED

SEARCHING

MATCH_FOUND

MATCH_CREATED

COMPLETED
```

Failure states

```
CANCELLED

TIMED_OUT

EXPIRED
```

---

# 4. Join Queue

POST

```
/api/v1/matchmaking/queue
```

Authentication

Required

---

## Purpose

Creates a new matchmaking session.

---

## Request

```json
{
    "subject":"reasoning",
    "timeControl":"2m"
}
```

---

## Validation

subject

Required

Allowed

- reasoning
- quant
- english
- mixed

timeControl

Allowed

30s

1m

2m

5m

---

## Business Rules

1. User cannot already be searching.
2. User cannot already be in a match.
3. Snapshot current rating.
4. Create Queue Session.
5. Put player into matchmaking pool.
6. Start searching.

---

## Database

```
INSERT queue_sessions
```

---

## Response

201 Created

```json
{
    "success":true,
    "data":{
        "queueSessionId":"uuid",
        "status":"SEARCHING",
        "joinedAt":"2026-07-22T12:30:40Z"
    }
}
```

---

## Errors

ALREADY_IN_QUEUE

MATCH_ALREADY_STARTED

VALIDATION_FAILED

---

# 5. Queue Status

GET

```
/api/v1/matchmaking/queue/{queueSessionId}
```

Authentication

Required

---

## Purpose

Returns current queue status.

---

## Response (Searching)

```json
{
    "status":"SEARCHING",
    "elapsedSeconds":18
}
```

---

## Response (Matched)

```json
{
    "status":"MATCH_FOUND",
    "matchId":"uuid"
}
```

---

## Response (Timeout)

```json
{
    "status":"TIMED_OUT"
}
```

---

# 6. Cancel Queue

DELETE

```
/api/v1/matchmaking/queue/{queueSessionId}
```

Authentication

Required

---

## Business Rules

Allowed only while

```
SEARCHING
```

Not allowed after

```
MATCH_FOUND
```

---

## Response

204 No Content

---

## Errors

MATCH_FOUND_ALREADY

QUEUE_NOT_FOUND

---

# 7. Queue History

GET

```
/api/v1/matchmaking/history
```

Authentication

Required

---

Returns

Previous queue attempts.

Useful for debugging.

---

# 8. Match Creation

Internal Endpoint

Not public.

Called by Matchmaking Service.

```
Create Match

↓

Generate Match ID

↓

Select Questions

↓

Reserve Players

↓

Publish Match Created Event
```

---

# 9. Matching Algorithm

Primary

Rating (ELO)

Secondary

Waiting Time

Example

```
Player A

Rating

1500

↓

Search

±50

↓

10 sec later

±100

↓

20 sec later

±150

↓

30 sec later

Unlimited
```

---

# 10. Queue Timeout

Default

60 seconds

If timeout occurs

Queue Session

↓

TIMED_OUT

Client may join again.

---

# 11. Events

QueueJoined

QueueCancelled

OpponentMatched

MatchCreated

QueueTimeout

---

# 12. Security

Players cannot

Choose opponent

Modify rating

Modify queue state

Everything is server controlled.

---

# 13. Rate Limits

Join Queue

5/minute

Cancel Queue

10/minute

Status

60/minute

---

# 14. Future Features

Friend Queue

Private Match

Ranked Queue

Casual Queue

Tournament Queue

Region Selection

Cross-region Matchmaking

Party Queue

Skill Buckets

Bot Match

---

# 15. Queue Lifecycle

```
Join Queue

↓

SEARCHING

↓

Opponent Found

↓

Create Match

↓

Move Players

↓

Battle Module
```
