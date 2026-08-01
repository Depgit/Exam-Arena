
# Subject API

Project: Exam Arena
Module: Subjects
Version: 1.0
Status: Draft
Author: <Your Name></your>
Last Updated: 2026-07-22

---

# 1. Purpose

The Subject module provides the content hierarchy used throughout the platform.

Responsibilities

- List available subjects
- List sections inside a subject
- Provide metadata for practice and ranked modes

It does NOT own

- Questions
- Statistics
- Practice sessions
- Matches

---

# 2. Domain Model

```

Subject

↓

Section

```

Example

```

Reasoning

↓

Blood Relation

```

Every Section belongs to exactly one Subject.

---

# 3. Subject Object

```json
{
    "id":"uuid",
    "slug":"reasoning",
    "name":"Reasoning",
    "description":"Logical reasoning and aptitude questions.",
    "icon":"brain",
    "color":"#4F46E5",
    "isEnabled":true
}
```

---

# 4. Section Object

```json
{
    "id":"uuid",
    "subjectId":"uuid",
    "slug":"blood-relation",
    "name":"Blood Relation",
    "description":"Family relationship problems.",
    "estimatedDurationMinutes": 45,
    "isEnabled":true
}
```

---

# 5. Get All Subjects

GET

```
/api/v1/subjects
```

Authentication

Not Required

---

## Purpose

Returns every enabled subject.

---

## Business Rules

Only enabled subjects are returned.

Order is controlled by administrator.

---

## Response

```json
{
    "success":true,
    "data":[
        {
            "id":"uuid",
            "slug":"reasoning",
            "name":"Reasoning",
            "icon":"brain"
        },
        {
            "id":"uuid",
            "slug":"quant",
            "name":"Quantitative Aptitude",
            "icon":"calculator"
        },
        {
            "id":"uuid",
            "slug":"english",
            "name":"English",
            "icon":"book"
        },
        {
            "id":"uuid",
            "slug":"general-awareness",
            "name":"General Awareness",
            "icon":"globe"
        }
    ]
}
```

---

# 6. Get Subject Details

GET

```
/api/v1/subjects/{subjectId}
```

Authentication

Not Required

---

## Purpose

Returns complete information about one subject.

---

## Response

```json
{
    "success":true,
    "data":{
        "id":"uuid",
        "slug":"reasoning",
        "name":"Reasoning",
        "description":"Logical aptitude problems.",
        "sectionCount":12
    }
}
```

---

## Errors

SUBJECT_NOT_FOUND

---

# 7. Get Sections

GET

```
/api/v1/subjects/{subjectId}/sections
```

Authentication

Not Required

---

## Purpose

Returns every enabled section under a subject.

---

## Response

```json
{
    "success":true,
    "data":[
        {
            "id":"uuid",
            "slug":"blood-relation",
            "name":"Blood Relation"
        },
        {
            "id":"uuid",
            "slug":"coding-decoding",
            "name":"Coding Decoding"
        },
        {
            "id":"uuid",
            "slug":"direction-sense",
            "name":"Direction Sense"
        }
    ]
}
```

---

## Errors

SUBJECT_NOT_FOUND

---

# 8. Get Section Details

GET

```
/api/v1/sections/{sectionId}
```

Authentication

Not Required

---

## Purpose

Returns metadata for one section.

---

## Response

```json
{
    "success":true,
    "data":{
        "id":"uuid",
        "subjectId":"uuid",
        "slug":"blood-relation",
        "name":"Blood Relation",
        "description":"Questions based on family relationships.",
        "estimatedDurationMinutes": 45
    }
}
```

---

## Errors

SECTION_NOT_FOUND

---

# 9. Search Sections

GET

```
/api/v1/sections
```

Query Parameters

```
?q=blood
```

Authentication

Not Required

---

## Purpose

Search sections by name.

---

## Response

```json
{
    "success":true,
    "data":[
        {
            "id":"uuid",
            "name":"Blood Relation"
        }
    ]
}
```

---

# 10. Subject Statistics

GET

```
/api/v1/subjects/{subjectId}/summary
```

Authentication

Not Required

---

## Purpose

Returns public metadata.

Example

```json
{
    "success":true,
    "data":{
        "sections":14,
        "questions":12840
    }
}
```

This endpoint does NOT return player statistics.

---

# 11. Business Rules

A Subject

- Has many Sections
- Cannot be deleted if questions exist
- May be disabled
- Has a unique slug

A Section

- Belongs to exactly one Subject
- Has a unique slug within its Subject
- Cannot be deleted if questions exist
- May be disabled

---

# 12. Events

SubjectCreated

SubjectUpdated

SectionCreated

SectionUpdated

Future

SubjectDisabled

SectionDisabled

---

# 13. Security

Read operations are public.

Create, update and delete operations belong only to the Admin module.

---

# 14. Caching

Subjects change very rarely.

Recommended Cache

24 hours

Future

Redis

Current

In-memory cache

---

# 15. Rate Limits

List Subjects

120 requests/minute

List Sections

120 requests/minute

Search Sections

60 requests/minute

---

# 16. Future Extensions

- Exam mapping
- Chapter hierarchy
- Tags
- Difficulty distribution
- AI-generated learning paths
- Subject images
- Localization
