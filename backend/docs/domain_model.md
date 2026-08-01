
# Domain Model

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the core business entities of Exam Arena and the
relationships between them.

The Domain Model is independent of:

- Database
- APIs
- Programming Language
- Framework

It represents the business itself.

---

# 2. Core Domain

The platform revolves around one central concept.

```
                Competition
                     │
     ┌───────────────┼───────────────┐
     │               │               │
   User            Match         Question
     │               │               │
     │               │               │
 Statistics      Answer         Topic
     │               │
     │               │
  Rating        Match Result
```

---

# 3. Entity: User

Represents a registered player.

Responsibilities

- Authentication
- Owns profile
- Plays matches
- Solves questions
- Has statistics
- Has rating
- Owns history

Relationships

```
User

├── Profile (1)

├── Rating (1)

├── Statistics (1)

├── MatchHistory (N)

├── PracticeSession (N)

├── Friend (N)

└── Notification (N)
```

Lifecycle

Register

↓

Active

↓

Suspended

↓

Deleted

---

# 4. Entity: Profile

Represents public information.

Contains

- Username
- Avatar
- Country
- Preferred Language
- Bio (future)

Relationship

```
User
   │
   │ 1 : 1
   ▼
Profile
```

---

# 5. Entity: Rating

Represents competitive skill.

Contains

- Current Rating
- Highest Rating
- Lowest Rating
- Rating History

Rules

Only ranked matches modify Rating.

Practice never changes Rating.

---

# 6. Entity: Statistics

Represents player performance.

Contains

Accuracy

Win Rate

Matches Played

Questions Solved

Average Time

Weak Topics

Strong Topics

Current Streak

Longest Streak

Relationship

```
User

│

└── Statistics
```

---

# 7. Entity: Match

Represents one competitive game.

Contains

Unique ID

Players

Question Set

Start Time

End Time

Status

Winner

Score

Type

Relationships

```
Match

├── Player A

├── Player B

├── Questions

├── Answers

└── Result
```

Lifecycle

Created

↓

Waiting

↓

Accepted

↓

Running

↓

Completed

↓

Archived

Alternative

Cancelled

Expired

Abandoned

---

# 8. Entity: Question

Represents a single problem.

Contains

Question Text

Options

Correct Answer

Explanation

Difficulty

Estimated Time

Topic

Language

Status

Relationships

```
Topic

│

└── Question

      │

      └── Explanation
```

Lifecycle

Draft

↓

Review

↓

Published

↓

Archived

---

# 9. Entity: Answer

Represents a player's response.

Contains

Player

Question

Selected Option

Submitted Time

Correct

Score

Time Taken

Relationship

```
Match

│

├── Question

│

└── Answer
```

One Answer belongs to exactly one Question in exactly one Match.

---

# 10. Entity: Topic

Represents a learning category.

Examples

Reasoning

Mathematics

English

General Knowledge

Contains

Name

Description

Parent Topic

Subtopics

Questions

Relationship

```
Topic

├── SubTopic

└── Questions
```

---

# 11. Entity: Match Result

Represents the outcome.

Contains

Winner

Loser

Draw

Rating Change

Duration

Accuracy

Relationship

```
Match

│

└── Match Result
```

---

# 12. Entity: Practice Session

Represents self-learning.

Contains

Questions

Answers

Duration

Accuracy

Review

No Rating

Relationship

```
User

│

└── Practice Session
```

---

# 13. Entity: Leaderboard

Represents rankings.

Contains

Player

Rank

Rating

Wins

Country

Season

Types

Global

Country

Friends

Season

---

# 14. Entity: Friendship

Represents a social connection.

Relationship

```
User

│

├── Friend

└── Friend
```

States

Pending

Accepted

Blocked

Removed

---

# 15. Entity: Tournament (Future)

Contains

Name

Schedule

Rules

Participants

Leaderboard

Rounds

Winner

Relationship

```
Tournament

├── Players

├── Matches

└── Leaderboard
```

---

# 16. Entity: Notification

Contains

Title

Message

Type

Status

Created Time

Relationship

```
User

│

└── Notifications
```

---

# 17. Aggregate Boundaries

To keep the domain maintainable, related entities form aggregates.

User Aggregate

```
User

├── Profile

├── Rating

├── Statistics

└── Notifications
```

Question Aggregate

```
Question

├── Explanation

└── Metadata
```

Match Aggregate

```
Match

├── Players

├── Questions

├── Answers

└── Result
```

Tournament Aggregate

```
Tournament

├── Participants

├── Matches

└── Rankings
```

---

# 18. Business Relationships

```
User

plays

↓

Match

contains

↓

Questions

answered by

↓

Answers

produces

↓

Result

updates

↓

Rating

updates

↓

Statistics
```

---

# 19. Domain Events

The business reacts to important events.

Examples

UserRegistered

MatchCreated

MatchStarted

QuestionAnswered

MatchFinished

RatingUpdated

FriendRequestSent

TournamentStarted

AchievementUnlocked

QuestionReported

These events describe business actions,
not implementation details.

---

# 20. Domain Invariants

These rules must always be true.

A Match has exactly two players.

A User cannot join two ranked matches simultaneously.

A Question belongs to one Topic.

An Answer belongs to one Match.

Only Ranked Matches change Rating.

Only Published Questions appear in games.

Server is the source of truth.

No duplicate active usernames.

---

# 21. Future Domain Objects

Clan

Team

Season

Reward

Badge

Store

Premium Subscription

Mission

Daily Challenge

Replay

AI Coach

These should integrate naturally without changing existing domain concepts.
