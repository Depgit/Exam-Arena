
# Database Design

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the logical database design for Exam Arena.

Goals

- Normalize business data
- Support efficient reads
- Maintain data integrity
- Scale from MVP to millions of users
- Keep future migrations simple

Database

PostgreSQL 16+

---

# 2. Design Principles

- UUID primary keys
- UTC timestamps
- Soft delete only where necessary
- Explicit foreign keys
- Indexed search fields
- Avoid storing derived data unless justified
- Business rules live in Go, not SQL procedures

---

# 3. Database Overview

```
User
 ├── Profile
 ├── Rating
 ├── Statistics
 ├── Friendships
 └── Sessions

Question
 ├── Topic
 ├── Option
 └── Explanation

Match
 ├── Players
 ├── Questions
 ├── Answers
 └── Result

Leaderboard (derived)

Notifications
Reports
```

---

# 4. User Aggregate

## users

Stores account identity.

Columns

- id
- email
- username
- password_hash
- status
- created_at
- updated_at

Indexes

- email (unique)
- username (unique)

---

## user_profiles

Stores public profile information.

Columns

- id
- user_id
- avatar_url
- country
- preferred_language
- display_name

Relationship

User (1) → Profile (1)

---

## user_settings

Stores user preferences.

Examples

- Theme
- Notification settings
- Language
- Privacy

---

# 5. Rating Aggregate

## ratings

Current rating.

Columns

- user_id
- current_rating
- highest_rating
- lowest_rating
- games_played
- updated_at

---

## rating_history

Immutable history.

Columns

- id
- user_id
- match_id
- previous_rating
- new_rating
- delta
- created_at

Never update rows.

Append only.

---

# 6. Statistics Aggregate

## user_statistics

Columns

- user_id
- matches_played
- wins
- losses
- draws
- questions_answered
- correct_answers
- average_response_ms
- current_streak
- longest_streak

---

## topic_statistics

One row per

User × Topic

Tracks

- attempts
- correct
- average_time

---

# 7. Question Aggregate

## topics

Columns

- id
- name
- parent_id
- description

Supports hierarchy.

---

## questions

Columns

- id
- topic_id
- exam_type
- difficulty
- language
- statement
- explanation
- estimated_time_sec
- status
- created_by
- created_at

Indexes

(topic_id, difficulty)

(status)

(language)

---

## question_options

Columns

- id
- question_id
- option_order
- option_text
- is_correct

One-to-many

Question → Options

---

# 8. Match Aggregate

## matches

Stores one battle.

Columns

- id
- type
- status
- duration_seconds
- started_at
- finished_at
- winner_user_id
- created_at

Indexes

status

started_at

---

## match_players

One row per player.

Columns

- id
- match_id
- user_id
- final_score
- accuracy
- rating_before
- rating_after
- disconnected

Relationship

Match (1)

↓

Players (2)

---

## match_questions

Stores generated question set.

Columns

- id
- match_id
- question_id
- sequence

Allows replay.

---

## match_answers

Stores submitted answers.

Columns

- id
- match_id
- question_id
- user_id
- selected_option_id
- is_correct
- response_time_ms
- score

Append only.

---

# 9. Practice Aggregate

## practice_sessions

Columns

- id
- user_id
- started_at
- finished_at
- mode

---

## practice_answers

Columns

- id
- session_id
- question_id
- selected_option_id
- response_time_ms
- is_correct

---

# 10. Friendship Aggregate

## friendships

Columns

- id
- requester_id
- receiver_id
- status
- created_at

Statuses

Pending

Accepted

Blocked

Removed

---

# 11. Notifications

## notifications

Columns

- id
- user_id
- type
- title
- body
- read
- created_at

---

# 12. Reports

## reports

Columns

- id
- reporter_id
- question_id
- reason
- status
- reviewed_by
- reviewed_at

---

# 13. Admin Audit

## audit_logs

Append-only.

Columns

- id
- admin_id
- action
- resource
- resource_id
- created_at

---

# 14. Relationships

```
users

├── user_profiles

├── ratings

├── statistics

├── friendships

├── practice_sessions

└── matches

matches

├── players

├── questions

└── answers

questions

├── topic

└── options
```

---

# 15. Index Strategy

Unique

- email
- username

Lookup

- topic_id
- difficulty
- status
- language

Sorting

- rating
- created_at
- started_at

Composite

(topic_id, difficulty)

(match_id, sequence)

(user_id, created_at)

---

# 16. Constraints

- Username unique
- Email unique
- Match has exactly two players (enforced by application)
- Question must belong to a topic
- Rating must exist for every user
- Foreign keys enabled

---

# 17. Soft Delete Strategy

Soft delete

- users
- questions

Hard delete

- notifications
- sessions
- temporary records

Match history is never deleted.

---

# 18. Transactions

Use transactions for

- User registration
- Match completion
- Rating update
- Question publishing

Avoid long-running transactions.

---

# 19. Future Tables

- tournaments
- tournament_matches
- clans
- clan_members
- achievements
- seasons
- daily_challenges
- replay_events
- subscriptions

---

# 20. Migration Strategy

Rules

- Never edit applied migrations
- Always create new migrations
- Forward-compatible changes
- Backward-compatible deployments when possible

Migration naming

000001_create_users

000002_create_questions

000003_create_matches

...

---

# 21. Performance Targets

- Login lookup < 10 ms
- Match creation < 50 ms
- Question selection < 100 ms
- Leaderboard query < 100 ms
- Player profile < 50 ms

Indexes should be reviewed as query patterns evolve.

---

# 22. Data Retention

Keep indefinitely

- Users
- Matches
- Ratings
- Questions
- Statistics

Cleanup periodically

- Expired sessions
- Temporary matchmaking data
- Old notification deliveries
