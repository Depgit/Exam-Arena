
# Design System

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>Deepak Mishra and Vivek Rai
Last Updated: 2026-07-22

---

# 1. Purpose

The Design System defines the visual language and reusable UI components
used across Exam Arena.

Goals

- Consistency
- Simplicity
- Accessibility
- Reusability
- Fast Development

Every screen should look like it belongs to the same product.

---

# 2. Design Principles

## Fast

Users should immediately understand what to do.

Avoid unnecessary clicks.

---

## Competitive

The interface should feel energetic and game-like,
not like a traditional coaching website.

---

## Minimal

Every screen should focus on one primary action.

Avoid clutter.

---

## Responsive

The application must work well on

- Desktop
- Tablet
- Mobile

---

## Accessible

Support

- Keyboard navigation
- Screen readers
- Color contrast
- Focus indicators

---

# 3. Color Palette

## Primary

Used for

- Main buttons
- Links
- Selected navigation
- Highlights

---

## Secondary

Used for

- Secondary actions
- Less important buttons

---

## Success

Used for

- Victory
- Correct Answer
- Completed Tasks

---

## Error

Used for

- Wrong Answer
- Validation Errors
- Failed Requests

---

## Warning

Used for

- Countdown
- Queue Timeout
- Important Alerts

---

## Neutral

Used for

- Background
- Cards
- Borders
- Dividers

---

# 4. Typography

## Headings

H1

Page Title

H2

Section Title

H3

Card Title

---

## Body

Primary Text

Secondary Text

Caption

Helper Text

---

## Numbers

Rating

Timer

Score

Statistics

Always use tabular numbers where possible.

---

# 5. Icons

Use one icon library consistently.

Icons should be simple.

Examples

Play

Practice

Leaderboard

Profile

Settings

Friends

History

Notifications

Search

Logout

Avoid mixing multiple icon libraries.

---

# 6. Spacing

Use a fixed spacing scale.

Examples

XS

Small

Medium

Large

XL

2XL

All layouts should follow this spacing scale.

---

# 7. Border Radius

Small

Inputs

Medium

Cards

Large

Modals

Extra Large

Profile Images

Avoid random border radius values.

---

# 8. Shadows

Level 1

Cards

Level 2

Dropdowns

Level 3

Modals

Avoid excessive shadows.

---

# 9. Buttons

Primary

Main action

Example

Play Ranked

---

Secondary

Alternative action

Example

Practice

---

Danger

Destructive action

Example

Delete Account

---

Success

Positive action

Example

Accept Match

---

Disabled

Waiting state

---

Loading

Shows spinner.

Button remains fixed.

---

# 10. Input Components

Text Input

Password

Search

Dropdown

Radio Button

Checkbox

Toggle

Textarea

Each input must support

Label

Placeholder

Validation

Helper Text

Error State

Disabled State

---

# 11. Cards

Cards display grouped information.

Examples

User Card

Question Card

Leaderboard Entry

Statistics

Achievements

Every card should contain

Title

Content

Optional Actions

---

# 12. Navigation

Desktop

Top Navigation

Mobile

Bottom Navigation

Navigation should clearly indicate

Current Page

Unread Notifications

Active Match

---

# 13. Tables

Used only for

Admin Dashboard

Reports

Question Management

User Management

Avoid tables in player-facing screens.

---

# 14. Modals

Used for

Match Acceptance

Delete Confirmation

Invite Friend

Settings

Every modal contains

Title

Description

Primary Action

Secondary Action

Close Button

---

# 15. Toast Notifications

Short messages.

Examples

Match Found

Profile Updated

Question Saved

Network Restored

Achievement Unlocked

Auto dismiss after a few seconds.

---

# 16. Loading States

Every API request must have

Loading Placeholder

Skeleton Screen

Spinner (only when appropriate)

Avoid blank pages.

---

# 17. Empty States

Examples

No Matches

No Friends

No Notifications

No Reports

Every empty state should include

Illustration

Helpful Message

Suggested Action

---

# 18. Error States

Examples

404

500

Network Error

Authentication Expired

Each error page should provide

Explanation

Recovery Action

Retry Button

---

# 19. Charts

Used for

Rating History

Accuracy

Topic Analysis

Weekly Progress

Charts should prioritize readability over decoration.

---

# 20. Progress Indicators

Linear Progress

Circular Progress

Countdown Timer

Queue Progress

Match Progress

---

# 21. Animations

Animations should be

Short

Smooth

Purposeful

Examples

Page Transition

Button Hover

Victory Screen

Rating Increase

Achievement Unlock

Avoid long animations.

---

# 22. Sound Effects (Future)

Optional

Match Found

Victory

Defeat

Achievement

Countdown

Mute option required.

---

# 23. Responsive Breakpoints

Mobile

Tablet

Desktop

Large Desktop

Layouts should adapt without changing functionality.

---

# 24. Component Library

Reusable Components

Button

Input

Dropdown

Card

Modal

Avatar

Badge

Chip

Tooltip

Progress Bar

Countdown Timer

Question Card

Leaderboard Row

Player Card

Statistic Card

Rating Badge

Loading Skeleton

Pagination

Search Bar

Filter Panel

Empty State

Error State

---

# 25. Battle Screen Principles

The battle screen is the most important screen.

Priority order

1. Question
2. Timer
3. Answer Options
4. Progress
5. Opponent Status

Everything else is secondary.

The player should never be distracted during a live battle.

---

# 26. Dashboard Principles

The dashboard exists to encourage playing.

The primary action must always be visible.

Priority

Play Ranked

Practice

Daily Challenge

Recent Performance

Statistics

---

# 27. Accessibility Checklist

Keyboard Navigation

Visible Focus States

High Contrast

Responsive Text

Alt Text

ARIA Labels

Readable Font Sizes

Touch Friendly Controls

---

# 28. Design Consistency Rules

Never create duplicate components.

Never use multiple button styles for the same action.

Never change spacing arbitrarily.

Never invent new colors without updating this document.

All new UI components must follow this Design System.

---

# 29. Future Enhancements

Dark Mode

Season Themes

Custom Avatars

Animated Backgrounds

Achievement Effects

Profile Customization

Premium Themes

Accessibility Improvements
