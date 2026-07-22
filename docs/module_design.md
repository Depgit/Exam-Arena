
# Module Design

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the logical modules inside the Exam Arena backend.

It describes:

- Module responsibilities
- Ownership
- Public interfaces
- Internal data
- Allowed dependencies

Each module should own a specific business capability.

---

# 2. Module Overview

The backend consists of the following modules.

```

Authentication
Users
Questions
Practice
Matchmaking
Battle
Rating
Statistics
Leaderboard
Friends
Notifications
Administration

```

Each module owns its own business logic.

---

# 3. Dependency Rules

Allowed

```

Presentation Layer

↓

Application Layer

↓

Domain Modules

↓

Infrastructure

```

Modules should communicate through exported services/interfaces.

Forbidden

- Circular dependencies
- Shared mutable state
- Direct database access across modules

---

# 4. Authentication Module

## Purpose

Identity and access management.

Responsibilities

- Register user
- Login
- JWT generation
- Password hashing
- Password reset
- Token validation

Owns

- Authentication workflow
- Credentials

Does NOT own

- Profile
- Statistics
- Rating

Public Operations

- Register
- Login
- ValidateToken
- RefreshToken
- ResetPassword

---

# 5. User Module

## Purpose

Manage user information.

Responsibilities

- Profile
- Preferences
- Avatar
- Account settings

Owns

- Public profile
- User preferences

Depends On

Authentication

Does NOT own

- Rating
- Statistics
- Match history

Public Operations

- GetProfile
- UpdateProfile
- UpdatePreferences

---

# 6. Question Module

## Purpose

Manage question bank.

Responsibilities

- CRUD questions
- Topics
- Difficulty
- Explanations
- Validation

Owns

- Questions
- Topics
- Metadata

Public Operations

- CreateQuestion
- UpdateQuestion
- DeleteQuestion
- GetQuestion
- SearchQuestions
- SelectQuestions

Only published questions may be selected.

---

# 7. Practice Module

## Purpose

Single-player learning.

Responsibilities

- Start practice
- Save progress
- Practice review
- Retry incorrect questions

Depends On

Question Module

Statistics Module

Does NOT modify

Rating

Public Operations

- StartPractice
- SubmitAnswer
- FinishPractice
- ReviewPractice

---

# 8. Matchmaking Module

## Purpose

Find suitable opponents.

Responsibilities

- Queue management
- Opponent selection
- Queue timeout
- Match creation request

Owns

- Waiting queue

Does NOT own

- Battle
- Rating

Public Operations

- JoinQueue
- LeaveQueue
- FindOpponent
- CancelQueue

Future

Redis-backed queue.

---

# 9. Battle Module

## Purpose

Run live matches.

Responsibilities

- Match lifecycle
- Timer
- Question progression
- Answer validation
- Winner calculation

Owns

- Active matches

Depends On

Question Module

Rating Module

Statistics Module

Public Operations

- StartMatch
- SubmitAnswer
- FinishMatch
- ReconnectPlayer

---

# 10. Rating Module

## Purpose

Competitive ranking.

Responsibilities

- Elo calculation
- Rating history
- Seasonal reset (future)

Owns

- Rating rules

Public Operations

- CalculateRating
- UpdateRating
- RatingHistory

Only Battle Module may request rating updates.

---

# 11. Statistics Module

## Purpose

Performance analytics.

Responsibilities

- Accuracy
- Win rate
- Topic analysis
- Trends
- Weak topics

Owns

Player statistics.

Public Operations

- UpdateStats
- GetStats
- TopicAnalysis

Statistics are append-only where possible.

---

# 12. Leaderboard Module

## Purpose

Ranking display.

Responsibilities

- Global rankings
- Country rankings
- Seasonal rankings

Depends On

Rating Module

Statistics Module

Public Operations

- GlobalLeaderboard
- CountryLeaderboard
- PlayerRank

Leaderboard is read-only.

---

# 13. Friend Module

## Purpose

Social interactions.

Responsibilities

- Friend requests
- Friend list
- Friend battles

Public Operations

- SendRequest
- AcceptRequest
- RemoveFriend
- ListFriends

---

# 14. Notification Module

## Purpose

User notifications.

Responsibilities

- Match found
- Password reset email
- Tournament reminders
- Achievement notifications

Supports

In-app

Email

Push (future)

Public Operations

- Send
- MarkRead
- ListNotifications

---

# 15. Administration Module

## Purpose

Platform management.

Responsibilities

- Manage users
- Manage questions
- Moderate reports
- Dashboard analytics

Depends On

Question

User

Statistics

Public Operations

- BanUser
- PublishQuestion
- ResolveReport

---

# 16. Ownership Matrix

| Module         | Owns               |
| -------------- | ------------------ |
| Authentication | Credentials        |
| User           | Profile            |
| Questions      | Question Bank      |
| Practice       | Practice Sessions  |
| Matchmaking    | Waiting Queue      |
| Battle         | Active Matches     |
| Rating         | Elo                |
| Statistics     | Performance Data   |
| Leaderboard    | Rankings           |
| Friends        | Relationships      |
| Notifications  | User Notifications |
| Administration | Moderation         |

---

# 17. Cross-Module Communication

Modules never manipulate another module's data directly.

Correct

```

Battle
↓

Rating.Update()

```

Incorrect

```

Battle

↓

UPDATE ratings SET ...

```

Business rules stay inside the owning module.

---

# 18. Module Lifecycle

Request

↓

Handler

↓

Application Service

↓

Target Module

↓

Repository

↓

Database

---

# 19. Events Between Modules

Examples

MatchFinished

↓

Rating Module

↓

Statistics Module

↓

Leaderboard Module

Another example

UserRegistered

↓

Statistics Module

↓

Notification Module

Modules react to business events instead of tightly coupling logic.

---

# 20. Future Modules

Achievements

Daily Challenges

Arena

Tournament

Clans

Store

Premium

AI Coach

Replay

Each future module should follow the same ownership principles.

---

# 21. Module Design Principles

Every module has one clear responsibility.

Every business rule has one owner.

No module may bypass another module's rules.

Dependencies should always point toward the owning module.

The architecture should encourage adding new modules rather than modifying unrelated ones.
