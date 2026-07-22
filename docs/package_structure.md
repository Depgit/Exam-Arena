
# Package Structure

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the physical Go project structure.

The package structure reflects the business domain,
not the framework.

Every package should have one clear owner.

---

# 2. Repository Layout

```

exam-arena/

├── cmd/
│   └── server/
│       └── main.go
│
├── configs/
│
├── deployments/
│
├── docs/
│
├── migrations/
│
├── scripts/
│
├── web/
│
├── test/
│
├── internal/
│
├── pkg/
│
├── go.mod
└── README.md

```

---

# 3. cmd/

Contains application entry points.

```

cmd/

└── server/

└── main.go

```

Responsibilities

- Load configuration
- Build dependencies
- Start HTTP server
- Start WebSocket server
- Graceful shutdown

Business logic is forbidden.

---

# 4. configs/

Contains

- Development configuration
- Production configuration
- Environment templates

No secrets committed.

---

# 5. migrations/

Contains SQL migrations.

Example

```

000001_create_users.up.sql

000001_create_users.down.sql

```

---

# 6. deployments/

Deployment resources.

Examples

Dockerfile

docker-compose.yml

nginx.conf

systemd

future Kubernetes manifests

---

# 7. scripts/

Utility scripts.

Examples

Seed questions

Generate test users

Backup database

Import questions

---

# 8. docs/

Project documentation.

Contains every design document created before development.

---

# 9. test/

Integration and end-to-end tests.

Examples

Authentication flow

Match flow

Leaderboard flow

---

# 10. pkg/

Reusable generic packages.

Allowed

Logger

Configuration loader

JWT helper

Pagination

Validator

Clock

UUID

Forbidden

Business logic

Match logic

Question selection

Rating calculation

Anything specific to Exam Arena

---

# 11. internal/

Contains all business modules.

```

internal/

├── auth/
├── user/
├── question/
├── practice/
├── matchmaking/
├── battle/
├── rating/
├── statistics/
├── leaderboard/
├── friend/
├── notification/
├── admin/
│
├── shared/
│
└── platform/

```

Every folder owns one business capability.

---

# 12. Module Layout

Every business module follows the same structure.

Example

```

battle/

├── handler/
├── service/
├── domain/
├── repository/
├── websocket/
├── dto/
├── mapper/
├── events/
└── errors/

```

---

# 13. handler/

Responsibilities

- HTTP handlers
- Request parsing
- Validation
- Response serialization

Must never contain business logic.

---

# 14. websocket/

Contains

WebSocket handlers

Connection events

Realtime messages

Battle events

No business logic.

---

# 15. service/

Coordinates use cases.

Examples

StartMatch

SubmitAnswer

FinishMatch

Services orchestrate work.

They should delegate calculations to the domain layer.

---

# 16. domain/

Contains business rules.

Examples

Match

Answer

Score

Timer

Validation

Elo

Pure business logic.

No HTTP.

No SQL.

No framework.

---

# 17. repository/

Database access only.

Responsibilities

Load

Save

Delete

Search

Repositories never calculate business rules.

---

# 18. dto/

Request and response objects.

Examples

CreateMatchRequest

LoginResponse

LeaderboardDTO

DTOs never contain business logic.

---

# 19. mapper/

Converts between

Database

↓

Domain

↓

DTO

Avoid exposing database models directly.

---

# 20. events/

Domain events.

Examples

MatchStarted

MatchFinished

RatingUpdated

QuestionAnswered

Initially

Simple in-process events.

Future

Redis Pub/Sub

Kafka

---

# 21. errors/

Module-specific errors.

Example

```

ErrMatchNotFound

ErrPlayerOffline

ErrQueueTimeout

```

Typed errors.

---

# 22. shared/

Contains shared code.

Examples

Logger

Middleware

Validation

Context

Response helpers

Shared must never depend on business modules.

---

# 23. platform/

Infrastructure integrations.

Examples

PostgreSQL

SMTP

Redis (future)

Object Storage

Configuration

External APIs

The rest of the application talks to abstractions, not vendor SDKs.

---

# 24. Dependency Rules

Allowed

```

handler

↓

service

↓

domain

↓

repository

↓

database

```

Forbidden

```

repository

↓

handler

```

Forbidden

```

question

↓

battle

↓

question

```

No circular dependencies.

---

# 25. Import Rules

Business modules communicate through exported interfaces.

Never import another module's internal repository.

Correct

```

Battle Service

↓

Rating Service

```

Incorrect

```

Battle Repository

↓

Rating Repository

```

Repositories are private implementation details.

---

# 26. Naming Rules

Prefer

```

question/

battle/

rating/

```

Avoid

```

questions/

battle_service/

rating_module/

```

Package names should be short, singular, and meaningful.

---

# 27. File Naming

Examples

```

service.go

repository.go

handler.go

dto.go

mapper.go

errors.go

events.go

```

Or split by feature when files grow.

Example

```

submit_answer.go

start_match.go

finish_match.go

```

Prefer many small files over one giant file.

---

# 28. Growth Strategy

When a module becomes too large

Split internally.

Example

```

battle/

service/

start_match.go

submit_answer.go

finish_match.go

```

Instead of

```

battle/

service.go

5200 lines

```

---

# 29. Architecture Enforcement

Every pull request should verify

- No circular dependencies
- No business logic in handlers
- No SQL in services
- No framework inside domain
- No direct module coupling

---

# 30. Package Philosophy

Organize code by business capability.

Keep dependencies explicit.

Keep packages small.

Prefer composition.

Avoid utility dumping grounds.

The package structure should make it obvious where every new file belongs.
