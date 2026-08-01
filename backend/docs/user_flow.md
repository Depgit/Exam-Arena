
# User Flow Specification

Project: Exam Arena
Version: 1.0
Status: Draft
Author: Deepak Mishra and Vivek Rai
Last Updated: 2026-07-22

---

# 1. Purpose

This document describes how users interact with Exam Arena.

It defines:

- Primary user journeys
- Alternative paths
- Error scenarios
- Recovery flows

This document intentionally contains no implementation details.

---

# 2. User Types

Guest

Registered User

Administrator

---

# 3. Guest Journey

Landing Page

↓

Explore Features

↓

View Leaderboard

↓

Register

↓

Login

↓

Dashboard

---

Alternative Flow

Landing Page

↓

Login

↓

Dashboard

---

Failure Cases

Invalid credentials

↓

Display error

↓

Retry login

---

Forgot password

↓

Reset password

↓

Login

---

# 4. New User Onboarding

Register

↓

Email Verification (optional)

↓

Complete Profile

↓

Choose Preferred Exam

↓

Interactive Tutorial (future)

↓

Dashboard

---

# 5. Dashboard Flow

Dashboard

↓

User selects

Practice

OR

Ranked Battle

OR

Friend Battle

OR

Leaderboard

OR

Profile

OR

Settings

---

# 6. Practice Flow

Dashboard

↓

Practice

↓

Choose

Exam

Topic

Difficulty

Question Count

↓

Start Practice

↓

Answer Questions

↓

View Explanation

↓

Practice Summary

↓

Return Dashboard

OR

Retry Incorrect Questions

---

Alternative

User exits practice

↓

Progress saved

↓

Dashboard

---

# 7. Ranked Battle Flow

Dashboard

↓

Play Ranked

↓

Join Queue

↓

Searching...

↓

Opponent Found

↓

Accept Match

↓

Countdown

↓

Battle Begins

↓

Answer Questions

↓

Battle Ends

↓

Result Screen

↓

Rating Updated

↓

Return Dashboard

OR

Play Again

---

Failure

No opponent found

↓

Queue timeout

↓

Retry

OR

Cancel

---

Failure

User cancels queue

↓

Dashboard

---

Failure

Opponent declines

↓

Return to queue

---

Failure

Opponent disconnects before match starts

↓

Return to queue

---

# 8. Friend Battle Flow

Dashboard

↓

Create Room

↓

Generate Room Code

↓

Share Code

↓

Friend Joins

↓

Host Starts Match

↓

Battle

↓

Results

↓

Dashboard

---

Alternative

Join Existing Room

↓

Enter Code

↓

Wait Host

↓

Battle

---

Failure

Invalid room code

↓

Display error

↓

Retry

---

Failure

Host leaves room

↓

Room closed

↓

Dashboard

---

# 9. Match Flow

Match Starts

↓

Question 1

↓

Submit Answer

↓

Question 2

↓

Submit Answer

↓

...

↓

Final Question

↓

Submit Answer

↓

Server Calculates Winner

↓

Results

---

# 10. Disconnect Flow

Player Disconnects

↓

Reconnect Timer Starts

↓

Reconnect Successful

↓

Restore Match

↓

Continue Playing

---

Failure

Reconnect timeout

↓

Match forfeited

↓

Opponent wins

↓

Result screen

---

# 11. Browser Refresh Flow

Browser refreshed

↓

Reconnect WebSocket

↓

Validate Session

↓

Restore Match State

↓

Continue Match

---

Failure

Session expired

↓

Login

↓

Dashboard

---

# 12. Matchmaking Flow

Join Queue

↓

Waiting Players

↓

Suitable Opponent Found

↓

Create Match

↓

Notify Both Players

↓

Acceptance Window

↓

Accepted?

YES

↓

Start Match

NO

↓

Return Remaining Player To Queue

---

# 13. Leaderboard Flow

Dashboard

↓

Leaderboard

↓

Choose

Global

Country

Season

↓

View Rankings

↓

Open Player Profile

↓

Back

---

# 14. Match History Flow

Dashboard

↓

History

↓

Select Match

↓

Detailed Review

↓

Questions

↓

Answers

↓

Explanations

↓

Back

---

# 15. Statistics Flow

Dashboard

↓

Statistics

↓

Overview

↓

Topic Analysis

↓

Rating History

↓

Performance Trends

↓

Recommended Practice

---

# 16. Settings Flow

Dashboard

↓

Settings

↓

Change

Password

Theme

Language

Notifications

↓

Save

↓

Confirmation

---

# 17. Logout Flow

Dashboard

↓

Logout

↓

Invalidate Session

↓

Landing Page

---

# 18. Admin Flow

Admin Login

↓

Dashboard

↓

Questions

↓

Create/Edit/Delete

↓

Publish

↓

Available For Matches

---

Alternative

Admin

↓

Users

↓

Search

↓

Review

↓

Suspend

↓

Save

---

# 19. Question Reporting Flow

Question

↓

Report

↓

Choose Reason

↓

Submit

↓

Confirmation

↓

Admin Review

↓

Resolved

---

# 20. Tournament Flow (Future)

Dashboard

↓

Tournament

↓

Register

↓

Waiting Room

↓

Tournament Starts

↓

Rounds

↓

Leaderboard

↓

Rewards

---

# 21. Common Error Flows

Authentication Expired

↓

Login

↓

Continue

---

Network Lost

↓

Reconnect

↓

Restore Session

---

Server Maintenance

↓

Maintenance Screen

↓

Retry Later

---

Question Loading Failed

↓

Retry

↓

Continue

---

# 22. Session Recovery

Whenever possible the platform should recover the user's progress.

Recoverable

- Practice Session
- Ranked Match
- Friend Match
- User Preferences

Not Recoverable

- Expired Authentication
- Deleted Room
- Finished Match

---

# 23. UX Principles

The user should never lose progress because of:

- Browser Refresh
- Temporary Internet Failure
- WebSocket Reconnection
- Server Restart (future)

The platform should always provide:

- Clear Feedback
- Loading Indicators
- Error Messages
- Recovery ActionsMs
