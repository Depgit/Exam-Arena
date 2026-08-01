
# Internal Module Design

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the internal structure of each module.

Every module should consist of small, focused components.

A component should have one responsibility.

---

# 2. Design Principles

Each component should

- Have one responsibility
- Be independently testable
- Hide internal implementation
- Avoid global state
- Communicate through interfaces when necessary

---

# 3. Authentication Module

```
Authentication

├── Auth Service
├── Token Manager
├── Password Manager
├── Session Manager
└── Auth Repository
```

### Auth Service

Coordinates authentication workflows.

Examples

- Register
- Login
- Logout

---

### Token Manager

Responsible for

- JWT creation
- JWT validation
- Token expiration

---

### Password Manager

Responsible for

- Password hashing
- Password verification

---

### Session Manager

Tracks authenticated sessions.

Future

- Device management
- Multiple sessions

---

# 4. User Module

```
User

├── Profile Service
├── Preference Service
├── Avatar Service
└── User Repository
```

Profile Service

Owns profile updates.

Preference Service

Owns language and settings.

Avatar Service

Owns avatar uploads.

---

# 5. Question Module

```
Question

├── Question Service
├── Topic Manager
├── Difficulty Manager
├── Validator
├── Question Selector
└── Question Repository
```

Question Service

CRUD operations.

Topic Manager

Topic hierarchy.

Difficulty Manager

Difficulty metadata.

Validator

Question quality.

Question Selector

Returns fair question sets.

---

# 6. Practice Module

```
Practice

├── Practice Service
├── Session Manager
├── Progress Tracker
├── Review Generator
└── Practice Repository
```

Practice Service

Coordinates practice.

Session Manager

Tracks active sessions.

Progress Tracker

Records answers.

Review Generator

Builds post-practice report.

---

# 7. Matchmaking Module

```
Matchmaking

├── Queue Manager
├── Match Finder
├── Queue Scheduler
├── Queue Cleaner
└── Queue Repository
```

Queue Manager

Join and leave queue.

Match Finder

Finds suitable opponent.

Queue Scheduler

Processes waiting users.

Queue Cleaner

Removes expired entries.

---

# 8. Battle Module

```
Battle

├── Match Service
├── Match Manager
├── Session Manager
├── Timer Manager
├── Question Engine
├── Answer Validator
├── Score Calculator
├── Result Calculator
├── Reconnect Manager
├── Event Dispatcher
└── Match Repository
```

### Match Service

Coordinates the entire battle lifecycle.

### Match Manager

Owns active matches.

### Session Manager

Tracks player connections.

### Timer Manager

Maintains match timers.

### Question Engine

Serves next question.

### Answer Validator

Checks submitted answers.

### Score Calculator

Computes score.

### Result Calculator

Determines winner.

### Reconnect Manager

Restores disconnected players.

### Event Dispatcher

Broadcasts WebSocket events.

---

# 9. Rating Module

```
Rating

├── Elo Calculator
├── Rating Service
├── Rating History
└── Rating Repository
```

Elo Calculator

Pure business logic.

Rating Service

Coordinates updates.

Rating History

Stores changes.

---

# 10. Statistics Module

```
Statistics

├── Statistics Service
├── Accuracy Calculator
├── Trend Analyzer
├── Topic Analyzer
└── Statistics Repository
```

Accuracy Calculator

Calculates accuracy.

Trend Analyzer

Tracks performance over time.

Topic Analyzer

Detects strengths and weaknesses.

---

# 11. Leaderboard Module

```
Leaderboard

├── Ranking Service
├── Ranking Generator
├── Cache Manager (Future)
└── Leaderboard Repository
```

Ranking Generator

Builds leaderboard.

Cache Manager

Redis integration later.

---

# 12. Friend Module

```
Friends

├── Friend Service
├── Invitation Manager
├── Room Manager
└── Friend Repository
```

Invitation Manager

Friend requests.

Room Manager

Friend battle rooms.

---

# 13. Notification Module

```
Notifications

├── Notification Service
├── Email Sender
├── In-App Sender
├── Template Engine
└── Notification Repository
```

Template Engine

Builds notification messages.

---

# 14. Administration Module

```
Administration

├── User Admin
├── Question Admin
├── Report Manager
├── Dashboard Service
└── Audit Logger
```

Audit Logger

Records sensitive admin actions.

---

# 15. Shared Components

Some functionality is shared across modules.

```
Shared

├── Logger
├── Configuration
├── Validation
├── Clock
├── UUID Generator
├── Pagination
├── Error Types
└── Middleware
```

Shared code must remain generic.

Never place business logic here.

---

# 16. Internal Communication

Within a module

```
Service

↓

Manager

↓

Repository

↓

Database
```

Managers should not call unrelated managers directly.

---

# 17. Business Rule Placement

Business rules belong in calculators and managers.

Examples

Correct

```
Rating Service

↓

Elo Calculator
```

Incorrect

```
Handler

↓

Calculate Elo
```

---

# 18. Testing Strategy

Every component should be testable independently.

Priority

- Pure business logic
- Managers
- Services
- Repositories
- Handlers

---

# 19. Growth Strategy

When complexity increases

Split components.

Never create massive services.

Prefer

```
Battle

├── Timer Manager
├── Session Manager
```

Instead of

```
Battle Service

4,000 lines
```

---

# 20. Internal Design Rules

One responsibility per component.

One owner for every business rule.

Prefer composition over inheritance.

Keep components small.

Hide implementation details.

Favour explicit dependencies.

Avoid cyclic imports.

Every component should be understandable in isolation.
