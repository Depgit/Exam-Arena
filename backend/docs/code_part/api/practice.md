
# Practice API

Project: Exam Arena
Module: Practice
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

The Practice module allows users to practice questions independently without affecting their competitive rating.

Responsibilities

- Create practice sessions
- Select questions
- Deliver questions one at a time
- Accept answers
- Calculate score
- Store history
- Generate analytics

Practice sessions are private and belong to a single user.

---

# 2. Practice Flow

```
Create Session

↓

Session Created

↓

Get Current Question

↓

Submit Answer

↓

Get Next Question

↓

...

↓

Finish Session

↓

Review Results
```

---

# 3. Session States

```
CREATED

STARTED

ACTIVE

PAUSED

FINISHED

EXPIRED

ABANDONED
```

---

# 4. Create Practice Session

POST

```
/api/v1/practice/sessions
```

Authentication

Required

---

## Purpose

Creates a new practice session.

---

## Request

```json
{
    "subjectId":"uuid",
    "sectionId":"uuid",
    "difficulty":"medium",
    "questionCount":20
}
```

---

## Validation

subjectId

Required

sectionId

Required

difficulty

easy

medium

hard

mixed

questionCount

5

10

20

50

---

## Business Rules

1. Validate subject.
2. Validate section.
3. Randomly select questions.
4. Questions must not repeat inside the session.
5. Store selected question IDs.
6. Create session.

---

## Database

```
INSERT practice_sessions

↓

INSERT practice_session_questions
```

---

## Response

201 Created

```json
{
    "success":true,
    "data":{
        "sessionId":"uuid",
        "status":"CREATED"
    }
}
```

---

# 5. Start Session

POST

```
/api/v1/practice/sessions/{sessionId}/start
```

Purpose

Moves session to STARTED.

Returns first question.

---

## Response

```json
{
    "questionNumber":1,
    "totalQuestions":20,
    "question":{
        "id":"uuid",
        "statement":"....",
        "options":[]
    }
}
```

---

# 6. Get Current Question

GET

```
/api/v1/practice/sessions/{sessionId}/current-question
```

Purpose

Returns the active question.

---

## Business Rules

Only current question is returned.

Future questions are never exposed.

---

# 7. Submit Answer

POST

```
/api/v1/practice/sessions/{sessionId}/answer
```

Request

```json
{
    "optionId":"uuid",
    "timeTakenMs":18300
}
```

---

## Business Rules

Server

- verifies option
- records answer
- calculates correctness
- stores response time

If more questions remain

Advance pointer.

Otherwise

Finish session.

---

## Response

```json
{
    "correct":true,
    "score":1,
    "remainingQuestions":13,
    "nextQuestionAvailable":true
}
```

---

# 8. Get Next Question

GET

```
/api/v1/practice/sessions/{sessionId}/next-question
```

Returns

Next unanswered question.

---

# 9. Pause Session

POST

```
/api/v1/practice/sessions/{sessionId}/pause
```

Business Rule

Allowed only if timed mode is disabled.

---

# 10. Resume Session

POST

```
/api/v1/practice/sessions/{sessionId}/resume
```

Returns

Current unanswered question.

---

# 11. Finish Session

POST

```
/api/v1/practice/sessions/{sessionId}/finish
```

Purpose

Ends practice immediately.

Unanswered questions become skipped.

---

# 12. Get Result

GET

```
/api/v1/practice/sessions/{sessionId}/result
```

Response

```json
{
    "score":17,
    "correct":17,
    "wrong":3,
    "accuracy":85,
    "averageTimeMs":12800,
    "durationSeconds":425
}
```

---

# 13. Review Session

GET

```
/api/v1/practice/sessions/{sessionId}/review
```

Returns

Every question

User answer

Correct answer

Explanation

Time spent

---

# 14. Practice History

GET

```
/api/v1/practice/history
```

Query

```
page

pageSize

subjectId

sectionId
```

---

# 15. Delete Session

DELETE

```
/api/v1/practice/sessions/{sessionId}
```

Soft delete only.

---

# 16. Business Rules

A session

- belongs to one user
- cannot change subject
- cannot change section
- cannot add questions after creation
- cannot restart after finishing

---

# 17. Events

PracticeSessionCreated

PracticeStarted

QuestionAnswered

PracticeFinished

---

# 18. Rate Limits

Create Session

30/minute

Answer

Unlimited

History

60/minute

---

# 19. Security

Clients never submit

Correct answer

Score

Accuracy

Server computes everything.

---

# 20. Future

Adaptive Practice

Bookmarks

Mistake Notebook

AI Tutor

Hints

Timed Practice

Daily Goal

Review Weak Areas

Spaced Repetition
