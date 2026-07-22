
# Authentication API

Project: Exam Arena
Module: Authentication
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

The Authentication module is responsible for:

- User registration
- Login
- JWT authentication
- Refresh tokens
- Logout
- Password management

It does NOT manage:

- User profile
- Rating
- Statistics
- Friends

---

# 2. Authentication Strategy

Authentication uses:

- Access Token (JWT)
- Refresh Token

Access Token

Purpose

Authenticate API requests.

Lifetime

15 minutes

Refresh Token

Purpose

Generate a new access token.

Lifetime

30 days

Stored

- Database
- HttpOnly cookie (future mobile/web option)

---

# 3. Password Policy

Minimum

8 characters

Maximum

64 characters

Must contain

- Uppercase
- Lowercase
- Number

Recommended

Special character

Passwords are hashed using

Argon2id

Passwords are never stored in plaintext.

---

# 4. Register User

POST

/api/v1/auth/register

Authentication

Not Required

---

## Request

```json
{
    "email":"john@example.com",
    "username":"john123",
    "password":"StrongPassword123"
}
```

---

## Validation

Email

- Required
- Valid format
- Unique

Username

- Required
- 3-20 characters
- Letters, numbers, underscore

Password

- Required
- Minimum 8 characters

---

## Business Rules

1. Email must be unique.
2. Username must be unique.
3. Create user.
4. Create default profile.
5. Create default rating.

Initial Rating

```
1200
```

6. Create statistics record.
7. Generate tokens.
8. Return authenticated user.

---

## Database Operations

Transaction

```
INSERT users

↓

INSERT profiles

↓

INSERT ratings

↓

INSERT statistics

↓

INSERT refresh_tokens

COMMIT
```

If any step fails

Rollback.

---

## Events

Published

```
UserRegistered
```

Future consumers

- Welcome email
- Analytics
- Achievement system

---

## Success

201 Created

```json
{
    "success":true,
    "data":{
        "user":{
            "id":"uuid",
            "username":"john123",
            "rating":1200
        },
        "tokens":{
            "accessToken":"...",
            "refreshToken":"..."
        }
    }
}
```

---

## Errors

AUTH_EMAIL_ALREADY_EXISTS

AUTH_USERNAME_ALREADY_EXISTS

VALIDATION_FAILED

DATABASE_ERROR

---

# 5. Login

POST

/api/v1/auth/login

Authentication

Not Required

---

## Request

```json
{
    "email":"john@example.com",
    "password":"StrongPassword123"
}
```

---

## Validation

Email required.

Password required.

---

## Business Rules

1. Find user.
2. Verify password.
3. Check account status.
4. Generate access token.
5. Rotate refresh token.
6. Update last login timestamp.

---

## Database Operations

```
SELECT users

↓

Verify password

↓

UPDATE refresh token

↓

UPDATE last_login_at
```

---

## Response

200 OK

```json
{
    "success":true,
    "data":{
        "user":{
            "id":"uuid",
            "username":"john123",
            "rating":1425
        },
        "tokens":{
            "accessToken":"...",
            "refreshToken":"..."
        }
    }
}
```

---

## Errors

AUTH_INVALID_CREDENTIALS

AUTH_ACCOUNT_DISABLED

AUTH_INVALID_TOKEN

DATABASE_ERROR

---

# 6. Refresh Token

POST

/api/v1/auth/refresh

Authentication

Refresh Token

---

## Request

```json
{
    "refreshToken":"..."
}
```

---

## Business Rules

1. Verify refresh token.
2. Verify not expired.
3. Verify not revoked.
4. Rotate refresh token.
5. Issue new access token.

---

## Response

```json
{
    "success":true,
    "data":{
        "accessToken":"...",
        "refreshToken":"..."
    }
}
```

---

## Errors

AUTH_REFRESH_TOKEN_EXPIRED

AUTH_INVALID_TOKEN

---

# 7. Logout

POST

/api/v1/auth/logout

Authentication

Required

---

## Request

No body.

---

## Business Rules

1. Verify access token.
2. Revoke refresh token.
3. Remove active session.

---

## Response

204 No Content

---

## Errors

AUTH_INVALID_TOKEN

---

# 8. Get Current User

GET

/api/v1/auth/me

Authentication

Required

---

## Purpose

Returns authenticated user's identity.

---

## Response

```json
{
    "success":true,
    "data":{
        "id":"uuid",
        "username":"john123",
        "email":"john@example.com"
    }
}
```

---

# 9. Change Password

POST

/api/v1/auth/change-password

Authentication

Required

---

## Request

```json
{
    "currentPassword":"OldPassword",
    "newPassword":"NewPassword123"
}
```

---

## Business Rules

1. Verify current password.
2. Validate new password.
3. Hash password.
4. Replace password.
5. Revoke all refresh tokens.
6. Force login on all devices.

---

## Response

204 No Content

---

## Errors

AUTH_INVALID_CREDENTIALS

VALIDATION_FAILED

---

# 10. Forgot Password (Future)

POST

/api/v1/auth/forgot-password

Creates password reset token.

Not implemented in MVP.

---

# 11. Reset Password (Future)

POST

/api/v1/auth/reset-password

Uses password reset token.

Not implemented in MVP.

---

# 12. JWT Claims

Access Token contains

```json
{
    "sub":"user_uuid",
    "username":"john123",
    "iat":123456,
    "exp":123456,
    "iss":"exam-arena"
}
```

Never include

- Rating
- Statistics
- Friends
- Profile

---

# 13. Rate Limits

Register

5/minute/IP

Login

10/minute/IP

Refresh

30/minute/user

Change Password

5/hour/user

---

# 14. Audit Logs

Log

- Registration
- Login
- Logout
- Password change
- Refresh token usage

Never log

- Password
- JWT
- Refresh token

---

# 15. Security Rules

Passwords are hashed using Argon2id.

JWTs are signed using Ed25519.

Refresh tokens are random 256-bit values.

Refresh tokens are stored hashed.

Tokens are rotated on every refresh.

HTTPS required in production.

---

# 16. Sequence Diagram

Register

```
Client

↓

POST /register

↓

Validation

↓

Hash Password

↓

Create User
        │
        ▼
Publish UserRegistered Event
        │
        ├────────► User Module → Create Profile
        ├────────► Rating Module → Create Rating (1200)
        ├────────► Statistics Module → Create Statistics
        └────────► Notification Module → Welcome Notification

↓

Generate JWT

↓

Return Response
```

---

# 17. Future Enhancements

- Email verification
- Google login
- GitHub login
- Apple login
- Multi-device sessions
- Two-factor authentication (TOTP)
- Login history
- Trusted devices
- Suspicious login detection
