# System Context

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the system boundary for Exam Arena.

It identifies:

- Who uses the platform
- Which systems interact with it
- Which systems are internal
- Which systems are external
- The responsibilities of each

This document intentionally avoids implementation details.

---

# 2. System Overview

Exam Arena is a web-based competitive learning platform.

It allows students to:

- Practice aptitude questions
- Compete against real players
- Participate in tournaments
- Track learning progress
- Improve their competitive rating

Exam Arena is the central system.

Everything else either communicates with it or supports it.

---

# 3. System Boundary

```

                    +----------------------+
                    |      Internet        |
                    +----------+-----------+
                               |
                               |
          +--------------------+--------------------+
          |                                         |
          |                                         |
   +------+-------+                         +--------+--------+
   |   Web App    |                         |   Admin Portal  |
   +------+-------+                         +--------+--------+
          |                                         |
          +--------------------+--------------------+
                               |
                               ▼
                    ========================
                    |     Exam Arena        |
                    |    Backend System     |
                    ========================
                               |
        +-----------+----------+-----------+-----------+
        |           |                      |           |
        ▼           ▼                      ▼           ▼
 PostgreSQL     Redis*              Email Service   Object Storage*
                  (*Future)             (SMTP)         (*Future)
```

---

# 4. Primary Actors

## Student

Primary user.

Responsibilities

- Practice
- Play ranked matches
- Play with friends
- View leaderboard
- Manage profile

---

## Administrator

Responsible for platform management.

Responsibilities

- Manage questions
- Moderate reports
- Manage tournaments
- View analytics
- Manage users

---

## System

Exam Arena Backend

Responsible for

- Authentication
- Matchmaking
- Rating
- Question selection
- Statistics
- Match management
- Notifications

---

# 5. Client Applications

## Web Application

Primary client.

Responsibilities

- User Interface
- WebSocket connection
- API communication
- Local storage
- Rendering

---

## Mobile Application (Future)

Same capabilities as web.

Communicates using identical APIs.

---

# 6. Internal Systems

The following components belong to Exam Arena.

Authentication

Question Engine

Match Engine

Rating Engine

Practice Engine

Statistics Engine

Leaderboard Engine

Friend System

Tournament System (Future)

Notification System

These are logical modules.

They may initially exist inside one Go application.

---

# 7. External Systems

## PostgreSQL

Purpose

Persistent data storage.

Stores

Users

Questions

Matches

Answers

Ratings

Statistics

History

Configuration

---

## Redis (Future)

Purpose

Temporary in-memory storage.

Potential Uses

Matchmaking Queue

Online Users

Session Cache

Leaderboard Cache

Rate Limiting

WebSocket Presence

Redis is optional for Version 1.

---

## SMTP Email Provider

Purpose

Email verification

Password reset

System notifications

Examples

Gmail SMTP

Amazon SES

Mailgun

Resend

---

## Object Storage (Future)

Purpose

Avatar uploads

Question images

Assets

Examples

AWS S3

Cloudflare R2

MinIO

---

# 8. Communication

## Browser → Backend

REST API

HTTPS

Purpose

Authentication

Profile

Practice

History

Statistics

Leaderboard

---

## Browser → Backend

WebSocket

Purpose

Matchmaking

Battle

Live updates

Opponent status

Countdown

Reconnect

---

## Backend → PostgreSQL

Persistent storage.

---

## Backend → SMTP

Send email.

---

## Backend → Redis (Future)

Temporary state.

---

# 9. Data Ownership

Exam Arena owns

Users

Matches

Ratings

Questions

Statistics

History

Friendships

Notifications

External systems never own business data.

---

# 10. Security Boundary

Only the backend is trusted.

Browser

Not Trusted

WebSocket

Authenticated

REST API

Authenticated

Database

Private

Admin Portal

Role Protected

Never trust data sent by clients.

The backend validates every request.

---

# 11. Authentication Boundary

Users authenticate only through Exam Arena.

Authentication is never delegated to:

Google (V1)

Facebook

GitHub

Discord

OAuth may be added later.

---

# 12. Failure Boundaries

Browser Failure

User reconnects.

Backend Failure

Recover using persistent storage.

Database Failure

Application enters degraded mode.

SMTP Failure

Retry later.

Redis Failure

System continues operating in V1.

---

# 13. Scaling Strategy

Version 1

Single Go Server

Single PostgreSQL

Single VPS

Version 2

Multiple Go Servers

Redis

Load Balancer

Object Storage

Version 3

Dedicated Match Server

Dedicated WebSocket Server

Read Replicas

Distributed Cache

---

# 14. System Constraints

Version 1 constraints

No Kubernetes

No Kafka

No RabbitMQ

No Microservices

No Redis dependency

Single deployment

Simple operations

---

# 15. Non Goals

Exam Arena does not manage

Video Streaming

Live Classes

Payment Processing (V1)

Chat Platform

Content Creation Tools

Document Hosting

---

# 16. Context Summary

Exam Arena is the central authority for all gameplay.

The backend owns all business rules.

Clients display data.

External systems provide infrastructure services.

Business logic never leaves the backend.
