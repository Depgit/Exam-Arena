
# User API

Project: Exam Arena
Module: User
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

The User module manages player information.

It owns

- Public Profile
- Preferences
- Avatar
- Display Name
- Language
- Country

It does NOT own

- Authentication
- Rating
- Match History
- Statistics
- Friends

---

# 2. Resource

```

User

```

Represents a player.

---

# 3. User Model

```json
{
    "id":"uuid",
    "username":"deepak",
    "displayName":"Deepak Kumar",
    "avatar":"https://...",
    "country":"IN",
    "language":"en",
    "bio":"Backend Engineer",
    "createdAt":"2026-07-22T10:20:30Z"
}
```

---

# 4. Get Current User

GET

```
/api/v1/users/me
```

Authentication

Required

---

## Purpose

Returns profile of authenticated player.

---

## Business Rules

- User must exist.
- Account must be active.

---

## Response

```json
{
    "success":true,
    "data":{
        "id":"uuid",
        "username":"deepak",
        "displayName":"Deepak Kumar",
        "avatar":"https://...",
        "country":"IN",
        "language":"en",
        "bio":"Backend Engineer"
    }
}
```

---

## Errors

USER_NOT_FOUND

AUTH_INVALID_TOKEN

---

# 5. Update Profile

PATCH

```
/api/v1/users/me
```

Authentication

Required

---

## Request

```json
{
    "displayName":"Deepak Kumar",
    "country":"IN",
    "language":"en",
    "bio":"I love reasoning."
}
```

All fields are optional.

---

## Validation

Display Name

- 2-50 characters

Bio

- Maximum 200 characters

Country

- ISO-3166 country code

Language

- ISO-639-1

---

## Business Rules

Only owner can edit profile.

Username cannot be changed.

Email cannot be changed here.

---

## Database

```
UPDATE user_profiles
```

---

## Response

200 OK

```json
{
    "success":true,
    "data":{
        "updated":true
    }
}
```

---

## Errors

VALIDATION_FAILED

USER_NOT_FOUND

---

# 6. Upload Avatar

POST

```
/api/v1/users/me/avatar
```

Authentication

Required

---

## Content-Type

```
multipart/form-data
```

---

## Request

Field

```
avatar
```

---

## Validation

Allowed

- JPG
- PNG
- WEBP

Maximum Size

5 MB

---

## Business Rules

1. Validate image.
2. Upload to object storage.
3. Update avatar URL.
4. Delete old avatar (optional).

---

## Response

```json
{
    "success":true,
    "data":{
        "avatarUrl":"https://cdn.examarena.com/avatar/123.png"
    }
}
```

---

## Errors

VALIDATION_FAILED

USER_NOT_FOUND

---

# 7. Delete Avatar

DELETE

```
/api/v1/users/me/avatar
```

Authentication

Required

---

## Business Rules

Replace avatar with default image.

---

## Response

204 No Content

---

# 8. Get Public Profile

GET

```
/api/v1/users/{username}
```

Authentication

Optional

---

## Purpose

View another player's public profile.

---

## Response

```json
{
    "success":true,
    "data":{
        "username":"deepak",
        "displayName":"Deepak Kumar",
        "avatar":"https://...",
        "country":"IN",
        "joinedAt":"2026-01-10T08:00:00Z"
    }
}
```

Private information is never returned.

---

## Errors

USER_NOT_FOUND

---

# 9. Search Users

GET

```
/api/v1/users
```

Authentication

Required

---

## Query Parameters

```
?q=deep
&page=1
&pageSize=20
```

---

## Purpose

Search users by username.

Used for

- Friend requests
- Viewing profiles

---

## Response

```json
{
    "success":true,
    "data":[
        {
            "username":"deepak",
            "displayName":"Deepak Kumar",
            "avatar":"https://..."
        }
    ]
}
```

---

# 10. User Preferences

GET

```
/api/v1/users/me/preferences
```

Authentication

Required

---

## Response

```json
{
    "success":true,
    "data":{
        "language":"en",
        "theme":"dark",
        "sound":true,
        "notifications":true
    }
}
```

---

# 11. Update Preferences

PATCH

```
/api/v1/users/me/preferences
```

Authentication

Required

---

## Request

```json
{
    "language":"hi",
    "theme":"light",
    "sound":false,
    "notifications":true
}
```

---

## Business Rules

Only owner may update preferences.

---

## Response

200 OK

---

# 12. Deactivate Account

POST

```
/api/v1/users/me/deactivate
```

Authentication

Required

---

## Purpose

Deactivate the user's account without deleting historical data.

---

## Business Rules

- Active matches cannot be interrupted.
- User cannot deactivate while in matchmaking queue.
- Rating, match history, and statistics are preserved.

---

## Response

204 No Content

---

## Errors

ALREADY_IN_QUEUE

MATCH_IN_PROGRESS

---

# 13. Data Ownership

User owns

- Display Name
- Avatar
- Language
- Country
- Bio
- Preferences

Server owns

- Username
- Email
- Rating
- Statistics
- Match History
- Join Date

---

# 14. Events Published

ProfileUpdated

AvatarChanged

PreferencesUpdated

AccountDeactivated

Future subscribers

- Analytics
- Notification Service
- Achievement Service

---

# 15. Security Rules

A user can only modify their own profile.

Public endpoints never expose

- Email
- Password
- Refresh Tokens
- Login History
- Internal IDs

Avatar uploads are virus-scanned (future enhancement).

---

# 16. Rate Limits

Get Profile

120 requests/minute

Update Profile

20 requests/minute

Upload Avatar

10 requests/hour

Search Users

60 requests/minute

---

# 17. Future Features

- Custom profile banner
- Social links
- Verified badge
- Follow players
- Profile themes
- Profile achievements
- Profile badges
- Online status
- Last seen
