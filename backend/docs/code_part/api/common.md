
# Common API Specification

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the common API conventions used throughout the Exam Arena backend.

Every REST endpoint must follow these rules.

---

# 2. Base URL

Development

```
http://localhost:8080/api/v1
```

Production

```
https://api.examarena.com/api/v1
```

All endpoints are versioned.

Future breaking changes will use

```
/api/v2
```

---

# 3. Content Type

Requests

```
Content-Type: application/json
```

Responses

```
Content-Type: application/json
```

Character Encoding

```
UTF-8
```

---

# 4. Authentication

Authentication uses JWT Bearer Tokens.

Example

```
Authorization: Bearer <access_token>
```

Public endpoints

- Register
- Login
- Refresh Token
- Health Check

Every other endpoint requires authentication.

---

# 5. Resource Identifiers

All resources use UUID v7.

Example

```
01982534-2ef0-78db-a645-3c16d80d3b98
```

Client applications must treat IDs as opaque strings.

---

# 6. Timestamp Format

All timestamps are stored in UTC.

Format

```
2026-07-22T15:30:45Z
```

RFC3339

---

# 7. Standard Success Response

```json
{
    "success": true,
    "data": {},
    "meta": null
}
```

---

# 8. Standard Error Response

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

# 9. HTTP Status Codes

| Code | Meaning               |
| ---- | --------------------- |
| 200  | OK                    |
| 201  | Created               |
| 204  | No Content            |
| 400  | Bad Request           |
| 401  | Unauthorized          |
| 403  | Forbidden             |
| 404  | Not Found             |
| 409  | Conflict              |
| 422  | Validation Failed     |
| 429  | Too Many Requests     |
| 500  | Internal Server Error |

---

# 10. Pagination

Request

```
?page=1&pageSize=20
```

Default

```
page = 1
pageSize = 20
```

Maximum

```
pageSize = 100
```

Response

```json
{
    "success": true,
    "data": [],
    "meta": {
        "page": 1,
        "pageSize": 20,
        "totalItems": 540,
        "totalPages": 27
    }
}
```

---

# 11. Sorting

Syntax

```
?sort=rating
```

Descending

```
?sort=-rating
```

Multiple

```
?sort=-rating,username
```

---

# 12. Filtering

Examples

```
?subject=reasoning
```

```
?difficulty=medium
```

```
?language=en
```

Multiple values

```
?difficulty=easy,medium
```

---

# 13. Validation Rules

Strings

Leading and trailing whitespace is trimmed.

Empty strings are rejected unless explicitly allowed.

Maximum request body size

```
1 MB
```

---

# 14. Language Codes

English

```
en
```

Hindi

```
hi
```

Future languages follow ISO 639-1.

---

# 15. Rate Limiting

Authentication

10 requests/minute/IP

General API

120 requests/minute/user

Leaderboard

60 requests/minute/user

Admin APIs

30 requests/minute/user

When exceeded

```
429 Too Many Requests
```

---

# 16. Error Codes

Authentication

AUTH_INVALID_TOKEN

AUTH_EXPIRED_TOKEN

AUTH_INVALID_CREDENTIALS

AUTH_ACCOUNT_DISABLED

Validation

VALIDATION_FAILED

INVALID_REQUEST_BODY

INVALID_QUERY_PARAMETER

Resources

USER_NOT_FOUND

QUESTION_NOT_FOUND

MATCH_NOT_FOUND

SECTION_NOT_FOUND

SUBJECT_NOT_FOUND

Queue

ALREADY_IN_QUEUE

QUEUE_TIMEOUT

MATCH_NOT_READY

Gameplay

MATCH_FINISHED

QUESTION_ALREADY_ANSWERED

INVALID_OPTION

Infrastructure

DATABASE_ERROR

CACHE_ERROR

INTERNAL_SERVER_ERROR

---

# 17. Idempotency

Safe

GET

HEAD

OPTIONS

DELETE

Future

POST endpoints may support

```
Idempotency-Key
```

for retries.

---

# 18. Request Headers

Required

```
Authorization
```

Optional

```
Accept-Language
```

Example

```
Accept-Language: en
```

---

# 19. Response Headers

Examples

```
Content-Type

Cache-Control

X-Request-ID
```

---

# 20. API Naming Rules

Resources use plural nouns.

Correct

```
/users

/subjects

/matches
```

Actions use verbs only when necessary.

Correct

```
POST /practice/start

POST /matchmaking/join
```

Avoid

```
/getUsers

/createMatch

/deleteQuestion
```

---

# 21. Security

Clients never send

- Rating
- Score
- Winner
- Match Result

These are calculated by the server.

Never trust client timestamps for scoring.

---

# 22. Deprecation

Deprecated endpoints include

```
Deprecation: true
```

and a sunset date in the response headers before removal.

---

# 23. Logging

Every request receives

```
X-Request-ID
```

This ID must appear in server logs for tracing.

---

# 24. Versioning Policy

Breaking changes

New API version

Non-breaking changes

Same version

Fields are added but never silently removed.

---

# 25. Design Principles

The server is the source of truth.

Clients request actions.

The server validates all inputs.

The server computes all game outcomes.

APIs are stable, explicit, and backward-compatible whenever possible.
