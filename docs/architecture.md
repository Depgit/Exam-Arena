
# Engineering Architecture Guide

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the software architecture of Exam Arena.

It explains:

- System architecture
- Design principles
- Module boundaries
- Dependency rules
- Communication patterns
- Scaling strategy

It intentionally avoids implementation-specific code.

---

# 2. Architectural Goals

The architecture should be:

- Simple
- Modular
- Maintainable
- Testable
- Scalable
- Observable
- Secure

The MVP should optimize for developer productivity.

Future versions should scale without major rewrites.

---

# 3. Why a Modular Monolith?

Version 1 will use a Modular Monolith.

Reasons

- Small team
- Lower infrastructure cost
- Easier debugging
- Single deployment
- Simpler local development
- Faster feature delivery

A modular monolith is not a "big ball of mud."

Each module has clear ownership and boundaries.

---

# 4. Evolution Strategy

Phase 1

Single VPS

↓

Single Go Process

↓

PostgreSQL

↓

Nginx

---

Phase 2

Multiple Go Instances

↓

Redis

↓

Load Balancer

↓

Background Workers

---

Phase 3

Dedicated Services

↓

Match Service

↓

Question Service

↓

Notification Service

↓

Analytics Service

Microservices are introduced only when operational needs justify them.

---

# 5. High-Level Architecture

```
                        Internet
                            │
                            ▼
                        Nginx
                            │
            ┌───────────────┴───────────────┐
            │                               │
            ▼                               ▼
     REST API Layer                  WebSocket Layer
            │                               │
            └───────────────┬───────────────┘
                            ▼
                    Application Layer
                            │
        ┌──────────┬────────┼──────────┬─────────┐
        │          │        │          │         │
        ▼          ▼        ▼          ▼         ▼
      Auth     Match    Question    Rating    Practice
        │          │        │          │         │
        └──────────┴────────┼──────────┴─────────┘
                            ▼
                      Domain Layer
                            │
                            ▼
                    Persistence Layer
                            │
                            ▼
                      PostgreSQL
```

---

# 6. Architectural Layers

## Presentation Layer

Responsibilities

- REST API
- WebSocket
- Validation
- Authentication
- Serialization

Must NOT contain business logic.

---

## Application Layer

Coordinates use cases.

Examples

- Register User
- Start Match
- Submit Answer
- Calculate Rating
- Practice Session

This layer orchestrates workflows.

---

## Domain Layer

Contains business rules.

Examples

- Match rules
- Elo calculation
- Question selection
- Scoring
- Validation

The Domain Layer should have no knowledge of HTTP, SQL, or WebSockets.

---

## Infrastructure Layer

Responsibilities

- Database
- Logging
- Configuration
- Email
- Cache
- External services

Infrastructure supports the domain but never defines business rules.

---

# 7. Module Boundaries

The application is divided into business modules.

```
Authentication

Question Bank

Practice

Matchmaking

Battle Engine

Rating

Statistics

Leaderboard

Friends

Tournament

Notifications

Administration
```

Each module owns its own business logic.

Modules communicate through well-defined interfaces.

---

# 8. Dependency Rules

Dependencies always point inward.

```
HTTP

↓

Application

↓

Domain

↓

Infrastructure
```

Forbidden

- Domain importing HTTP
- Domain importing SQL
- Domain importing WebSocket
- Domain importing Gin/Fiber

---

# 9. Communication Patterns

## REST

Used for

Authentication

Profile

History

Statistics

Question Management

Settings

Leaderboard

---

## WebSocket

Used for

Matchmaking

Live Battle

Countdown

Opponent Status

Timer

Reconnect

Score Updates

---

## Internal Calls

Modules communicate through interfaces and services.

No module should directly manipulate another module's data.

---

# 10. Request Lifecycle

Client

↓

HTTP Request

↓

Router

↓

Middleware

↓

Handler

↓

Application Service

↓

Domain Logic

↓

Repository

↓

Database

↓

Response

---

# 11. WebSocket Lifecycle

Client Connects

↓

Authentication

↓

Session Created

↓

Join Match Queue

↓

Receive Events

↓

Battle

↓

Disconnect

↓

Reconnect

↓

Close

---

# 12. Concurrency Model

Go routines are used for:

- Match timers
- Queue processing
- Background jobs
- Email sending
- Statistics updates

Shared state must be protected.

Avoid global mutable state.

---

# 13. Persistence Strategy

Persistent Data

- Users
- Questions
- Matches
- Ratings
- Statistics

Temporary Data

Initially

- In memory

Future

- Redis

Business logic must not depend directly on Redis.

---

# 14. Configuration

Configuration comes from:

Environment Variables

Configuration File

Secrets

No hardcoded configuration.

---

# 15. Logging

Structured logging only.

Every request should include:

- Request ID
- User ID (if authenticated)
- Match ID (if applicable)
- Duration
- Status

Sensitive information must never be logged.

---

# 16. Error Handling

Every layer returns typed errors.

Presentation layer converts errors into HTTP responses or WebSocket events.

Never expose internal implementation details to clients.

---

# 17. Security

JWT Authentication

Password Hashing

HTTPS

Rate Limiting

Input Validation

Role-Based Access

Server Authoritative Game State

Never trust client input.

---

# 18. Observability

Collect

- Request latency
- Error rates
- Active users
- Active matches
- Queue wait time
- Database latency

Logs, metrics, and health checks should be built in from the beginning.

---

# 19. Testing Strategy

Unit Tests

- Domain logic
- Elo calculation
- Question selection

Integration Tests

- Database
- API
- Match flow

End-to-End Tests

- Registration
- Practice
- Ranked match
- Friend match

Business rules should be tested independently of HTTP.

---

# 20. Deployment Architecture

Version 1

```
Internet
    │
    ▼
Nginx
    │
    ▼
Go Application
    │
    ▼
PostgreSQL
```

Version 2

```
Internet
    │
Load Balancer
    │
 ┌──┴─────┐
 │        │
 ▼        ▼
Go      Go
 │        │
 └──┬─────┘
    ▼
 Redis
    │
    ▼
PostgreSQL
```

---

# 21. Scalability Principles

Scale vertically first.

Scale horizontally only when necessary.

Introduce Redis only when:

- Matchmaking requires shared state
- Multiple Go instances are deployed
- Cache becomes beneficial

Introduce message queues only when background work becomes significant.

---

# 22. Coding Principles

- Small packages
- Single responsibility
- Explicit dependencies
- Dependency injection
- Context propagation
- No hidden global state
- Prefer composition over inheritance
- Keep interfaces close to where they are used

---

# 23. Anti-Patterns to Avoid

- God services
- Massive utility packages
- Circular dependencies
- Business logic inside handlers
- SQL inside handlers
- Global variables
- Premature microservices
- Copy-paste validation
- Large transactions without need

---

# 24. Architecture Evolution

The architecture should evolve by adding capabilities, not by rewriting the system.

Examples

V1
In-memory matchmaking

↓

V2
Redis-backed matchmaking

↓

V3
Dedicated matchmaking service

Without changing the public APIs or core domain rules.

---

# 25. Architecture Principles

The backend is the source of truth.

The domain owns business rules.

Infrastructure supports the domain.

Modules communicate through clear contracts.

Every architectural decision should optimize for long-term maintainability while keeping the MVP simple enough to ship quickly.
