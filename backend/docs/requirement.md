
# Product Requirements Document (PRD)

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>Deepak Mishra and Vivek Rai
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the functional and non-functional requirements of
Exam Arena.

It serves as the source of truth for product development.

This document intentionally avoids implementation details.

---

# 2. Goals

The platform should allow students to:

• Practice questions

• Compete against real users

• Improve through analytics

• Track skill using a rating system

• Stay motivated through gamification

---

# 3. User Roles

## 3.1 Guest

A visitor who has not logged in.

Capabilities

• Browse landing page
• View public leaderboard
• Read about the platform
• Register
• Login

Restrictions

• Cannot play
• Cannot practice
• Cannot join tournaments

---

## 3.2 Registered User

A logged-in student.

Capabilities

• Practice questions
• Play ranked matches
• Play with friends
• Join tournaments
• View analytics
• Maintain profile
• View history
• Earn achievements

---

## 3.3 Administrator

Capabilities

• Manage questions
• Manage users
• Manage tournaments
• Moderate reports
• View analytics
• Configure system settings

---

# 4. Functional Requirements

## 4.1 Authentication

The system shall allow users to:

• Register
• Login
• Logout
• Reset password
• Change password
• Verify email (optional in V1)

---

## 4.2 User Profile

Each user shall have:

• Username
• Display Name
• Avatar
• Rating
• Country
• Preferred Language
• Statistics
• Achievements
• Account Settings

---

## 4.3 Question Bank

The system shall support:

• Multiple exam categories
• Multiple topics
• Multiple languages
• Difficulty levels
• Explanations
• Images
• Tags
• Estimated solving time

Question Types

• MCQ
• Multiple Correct (future)
• Integer Answer (future)

---

## 4.4 Practice Mode

Users shall be able to:

Choose

• Exam
• Topic
• Difficulty
• Number of Questions

Practice mode shall:

• Not affect rating
• Show explanations
• Record statistics

---

## 4.5 Ranked Battle

Users shall be able to:

• Join matchmaking
• Cancel matchmaking
• Play real-time matches

Battle requirements

• Same questions
• Same order
• Same timer
• Rating changes
• Match history stored

---

## 4.6 Friend Battle

Users shall be able to:

• Create room
• Join room
• Share room code
• Start match

Rating shall not change.

---

## 4.7 Arena Mode

Arena is a continuous competition.

Requirements

• Time limited
• Unlimited matches
• Continuous matchmaking
• Final leaderboard

---

## 4.8 Tournament

Requirements

Admin can:

• Create tournament
• Schedule tournament
• Configure rules

Users can:

• Register
• Participate
• View leaderboard

---

## 4.9 Leaderboards

System shall provide:

Global Leaderboard

Country Leaderboard

Friends Leaderboard

Season Leaderboard

---

## 4.10 Match History

Every completed match shall store:

• Opponent
• Result
• Rating change
• Time
• Questions
• Answers
• Accuracy
• Duration

---

## 4.11 Statistics

The system shall calculate:

Overall Accuracy

Average Solving Time

Topic Accuracy

Question Attempt Count

Rating History

Longest Win Streak

Longest Losing Streak

Total Matches

Practice Sessions

---

## 4.12 Achievements

Examples

• First Victory

• 10 Wins

• 100 Wins

• 5 Win Streak

• 1000 Questions Solved

• Tournament Winner

---

# 5. Match Requirements

Each ranked match shall satisfy:

Two players

Equal question set

Equal timer

Server authoritative

Rating updated

History stored

Statistics updated

---

# 6. Question Requirements

Each question shall have:

Unique ID

Question

Options

Correct Answer

Difficulty

Estimated Time

Topic

Subtopic

Explanation

Language

Status

Question quality should remain consistent across matches.

---

# 7. Rating Requirements

The system shall:

Assign initial rating

Update rating after every ranked match

Maintain rating history

Support seasonal reset (future)

Support leaderboards

---

# 8. Notifications

Users may receive notifications for:

Friend Invitation

Tournament Start

Match Found

Achievement Unlocked

Daily Challenge

System Announcement

---

# 9. Search

Users shall search:

Topics

Exams

Friends

Past Matches

---

# 10. Reporting

Users shall report:

Incorrect Question

Wrong Answer

Duplicate Question

Offensive Content

Suspicious Player

---

# 11. Admin Requirements

Admin shall:

Create Question

Update Question

Delete Question

Review Reports

Manage Users

Manage Tournaments

View Platform Statistics

---

# 12. Non Functional Requirements

## Performance

Login < 500 ms

Question Loading < 300 ms

Matchmaking < 10 seconds

Battle Latency < 100 ms

Leaderboard < 1 second

---

## Availability

Target uptime

99.9%

---

## Scalability

Initial Target

10,000 registered users

500 concurrent players

Future Target

1,000,000 users

100,000 concurrent users

---

## Security

JWT Authentication

Encrypted Passwords

HTTPS

Rate Limiting

Server-side Validation

Client Never Trusted

---

## Reliability

No rating loss

Match recovery after reconnect

Automatic retries

Graceful server shutdown

---

# 13. Out of Scope (Version 1)

Video Courses

Live Classes

Marketplace

AI Tutor

Chat Rooms

Voice Chat

Offline Mode

---

# 14. Assumptions

Users have internet connectivity.

Users access the platform using modern browsers.

Server maintains authoritative game state.

Questions are curated by administrators.

---

# 15. Future Enhancements

Adaptive Learning

AI Generated Practice

Clan Battles

Team Battles

Season Pass

Premium Membership

Mobile Application

Live Streaming

API for Third Party Integrations

---



# 16. User Flow

Guest
    ↓
Register
    ↓
Login
    ↓
Complete Profile
    ↓
Choose Mode
        ├── Practice
        ├── Ranked Battle
        ├── Friend Battle
        └── Tournament
    ↓
Play Match
    ↓
View Result
    ↓
Rating Updated
    ↓
View Analytics
    ↓
Play Again

---

# 17. Admin Flow

Admin
    ↓
Login
    ↓
Question Dashboard
    ↓
Create Question
    ↓
Review
    ↓
Publish
    ↓
Available for Matchmaking
