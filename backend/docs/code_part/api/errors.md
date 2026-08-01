
# API Error Catalog

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines every API error returned by the Exam Arena backend.

Goals

- Consistent error responses
- Predictable client behavior
- Easy frontend handling
- Standard logging

Every error has

- Error Code
- HTTP Status
- Message
- Description
- Recovery Action

The error code is stable.

The message may change.

---

# 2. Standard Error Response

```json
{
    "success": false,
    "error": {
        "code": "USER_NOT_FOUND",
        "message": "User not found.",
        "details": null
    }
}
```

---

# 3. Error Categories

AUTH

VALIDATION

USER

SUBJECT

SECTION

QUESTION

PRACTICE

MATCHMAKING

MATCH

STATISTICS

FRIENDS

NOTIFICATION

ADMIN

SYSTEM

---

# 4. Authentication Errors

## AUTH_INVALID_CREDENTIALS

HTTP

401 Unauthorized

Meaning

Email or password is incorrect.

Recovery

Prompt user to login again.

---

## AUTH_INVALID_TOKEN

HTTP

401 Unauthorized

Meaning

JWT token is malformed or invalid.

Recovery

Refresh token or login.

---

## AUTH_EXPIRED_TOKEN

HTTP

401 Unauthorized

Meaning

Access token expired.

Recovery

Call refresh endpoint.

---

## AUTH_REFRESH_TOKEN_EXPIRED

HTTP

401 Unauthorized

Meaning

Refresh token expired.

Recovery

User must login again.

---

## AUTH_ACCOUNT_DISABLED

HTTP

403 Forbidden

Meaning

Administrator disabled account.

---

## AUTH_EMAIL_ALREADY_EXISTS

HTTP

409 Conflict

Meaning

Email already registered.

---

## AUTH_USERNAME_ALREADY_EXISTS

HTTP

409 Conflict

Meaning

Username already exists.

---

# 5. Validation Errors

## VALIDATION_FAILED

HTTP

422 Unprocessable Entity

Meaning

One or more fields failed validation.

Example

```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_FAILED",
        "message": "Validation failed.",
        "details": {
            "username": "minimum length is 3",
            "password": "minimum length is 8"
        }
    }
}
```

---

## INVALID_REQUEST_BODY

HTTP

400 Bad Request

Meaning

Malformed JSON.

---

## INVALID_QUERY_PARAMETER

HTTP

400 Bad Request

Meaning

Query parameter is invalid.

---

## INVALID_PATH_PARAMETER

HTTP

400 Bad Request

Meaning

Invalid UUID or path value.

---

# 6. User Errors

## USER_NOT_FOUND

404

User does not exist.

---

## USER_ALREADY_EXISTS

409

User already exists.

---

## USER_PROFILE_NOT_FOUND

404

Profile missing.

---

## USER_NOT_ACTIVE

403

Account inactive.

---

# 7. Subject Errors

## SUBJECT_NOT_FOUND

404

Subject does not exist.

---

# 8. Section Errors

## SECTION_NOT_FOUND

404

Section does not exist.

---

# 9. Question Errors

## QUESTION_NOT_FOUND

404

Question not found.

---

## QUESTION_ALREADY_ANSWERED

409

Answer already submitted.

---

## INVALID_OPTION

400

Selected option does not belong to question.

---

## QUESTION_NOT_AVAILABLE

409

Question unavailable.

---

# 10. Practice Errors

## PRACTICE_SESSION_NOT_FOUND

404

Practice session not found.

---

## PRACTICE_ALREADY_FINISHED

409

Session completed.

---

## PRACTICE_NOT_STARTED

409

Practice not started.

---

# 11. Matchmaking Errors

## ALREADY_IN_QUEUE

409

Player already waiting.

---

## QUEUE_TIMEOUT

408

No opponent found.

---

## MATCH_FOUND_ALREADY

409

Player already matched.

---

## INVALID_QUEUE_SELECTION

400

Unsupported game mode.

---

# 12. Match Errors

## MATCH_NOT_FOUND

404

Match does not exist.

---

## MATCH_ALREADY_STARTED

409

Match already running.

---

## MATCH_ALREADY_FINISHED

409

Match completed.

---

## MATCH_NOT_STARTED

409

Match has not started.

---

## PLAYER_NOT_IN_MATCH

403

User does not belong to match.

---

## OPPONENT_DISCONNECTED

409

Opponent disconnected.

---

## ANSWER_TIMEOUT

408

Answer time expired.

---

## RECONNECT_WINDOW_EXPIRED

410

Reconnect period ended.

---

# 13. Statistics Errors

## STATISTICS_NOT_FOUND

404

Statistics unavailable.

---

# 14. Friend Errors

## FRIEND_NOT_FOUND

404

Friend relationship missing.

---

## FRIEND_REQUEST_ALREADY_SENT

409

Duplicate request.

---

## FRIEND_REQUEST_NOT_FOUND

404

Request missing.

---

## CANNOT_ADD_SELF

400

Cannot add yourself.

---

# 15. Notification Errors

## NOTIFICATION_NOT_FOUND

404

Notification missing.

---

# 16. Admin Errors

## ADMIN_PERMISSION_REQUIRED

403

Admin access required.

---

## QUESTION_NOT_APPROVED

409

Question pending review.

---

## DUPLICATE_QUESTION

409

Duplicate question detected.

---

# 17. Rate Limiting

## RATE_LIMIT_EXCEEDED

429

Too many requests.

Recovery

Retry later.

---

# 18. Infrastructure Errors

## DATABASE_ERROR

500

Unexpected database error.

---

## CACHE_ERROR

500

Cache unavailable.

---

## WEBSOCKET_ERROR

500

Realtime communication failed.

---

## INTERNAL_SERVER_ERROR

500

Unexpected server error.

---

# 19. Error Design Rules

Errors should

- Have stable codes
- Never expose database details
- Never expose stack traces
- Never expose SQL queries
- Be safe for clients

---

# 20. Logging

Every server error logs

- Request ID
- User ID (if authenticated)
- Endpoint
- Error Code
- Stack Trace (internal only)

Clients never receive stack traces.

---

# 21. Localization

Error messages may be translated.

Clients should rely on

error.code

not

error.message

for application logic.

---

# 22. Future Errors

As new modules are added

- Tournament
- Clan
- Arena
- AI Coach
- Daily Challenge

new error groups should follow the same naming convention.
