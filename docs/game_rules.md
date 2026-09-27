# Game Rules

Project: Exam Arena
Version: 1.0
Status: Draft
Author: <Your Name></your>Deepak Mishra and Vivek Rai
Last Updated: 2026-07-22

---

# 1. Purpose

This document defines the official game mechanics used by Exam Arena.

Every ranked battle, friend battle, tournament and arena match follows these
rules unless explicitly overridden by the game mode.

This document is the source of truth for gameplay.

---

# 2. Objectives

The goals of the game are:

• Make learning enjoyable.

• Reward knowledge and speed.

• Keep every match fair.

• Minimize luck.

• Prevent cheating.

---

# 3. Match Types

The platform supports multiple match types.

## Practice

Single player.

No rating changes.

Statistics recorded.

Explanations shown.

---

## Ranked Battle

Two players.

Real-time.

Rating affected.

Statistics recorded.

Match history stored.

---

## Friend Battle

Two invited players.

Private room.

No rating changes.

Statistics recorded.

---

## Arena

Continuous ranked matches.

Fixed duration event.

Highest score wins.

Open to everyone: players are paired with the longest-waiting player in the pool regardless of rating. Ratings still update.

---

## Daily Challenge

Solo, played from the home page (not a matchmaking queue).

Same 10 questions for every player each day; the day resets at midnight IST.

One attempt per day, 180 seconds, answers can be changed until submitted.

Ranked by most correct, then fastest time. No rating change.

---

## Tournament

Scheduled competition.

Admin controlled.

Special leaderboard.

Rewards.

---

# 4. Match Flow

A ranked match follows this sequence.

Player enters queue

↓

Matchmaking

↓

Opponent found

↓

Question set generated

↓

30 second acceptance window

↓

Countdown (3...2...1)

↓

Battle starts

↓

Questions answered

↓

Battle ends

↓

Winner calculated

↓

Rating updated

↓

Statistics updated

↓

Match stored

---

# 5. Match Duration

Supported queues

90 Seconds

2 Minutes

3 Minutes

Future

5 Minutes

Custom Rooms

---

# 6. Question Selection

Questions are selected dynamically.

Selection is based on

Difficulty

Topic

Estimated solving time

Question quality

Previous attempts

Duplicate avoidance

Questions continue being selected until the total estimated solving time reaches
the target match duration.

Example

Target = 120 seconds

Question A = 20 sec

Question B = 15 sec

Question C = 30 sec

Question D = 25 sec

Question E = 30 sec

Total = 120 sec

Both players receive the exact same set.

---

# 7. Difficulty Distribution

Default distribution

40% Easy

40% Medium

20% Hard

This distribution may vary depending on game mode.

---

# 8. Question Order

Questions appear sequentially.

Players cannot skip questions.

Players cannot return to previous questions.

Server controls question order.

---

# 9. Timer Rules

Each match has:

Overall Match Timer

Each question also has an estimated solving time used only for question
selection.

There is NO per-question timeout.

Players may spend time strategically.

---

# 10. Scoring

Question Score = Base Score × Difficulty Multiplier × Accuracy Multiplier

Speed Bonus = Remaining Match Time × Bonus Factor

Wrong Answer = 0

Skipped Question = 0

OR

Difficulty 1 = 50 points
Difficulty 2 = 75 points
Difficulty 3 = 100 points
Difficulty 4 = 150 points
Difficulty 5 = 200 points

---

# 11. Winner Determination

Winner priority

1. Highest Score
2. More Correct Answers
3. Less Total Time Used
4. Draw

---

# 12. Rating Rules

Only Ranked Battles modify rating.

Winner gains rating.

Loser loses rating.

Draw never lowers the higher-rated player's rating: it stays unchanged, while the lower-rated player gains according to Elo rules.

Friend Battles never modify rating.

Practice never modifies rating.

---

# 13. Disconnect Rules

If player disconnects

Server keeps match alive for

60 seconds.

Player may reconnect.

If timeout expires

Disconnected player forfeits.

Opponent wins.

---

# 14. Reconnection

Player reconnects

↓

Server restores

Current Question

Remaining Match Time

Current Score

Connection resumes.

---

# 15. Anti-Cheat Rules

Server is authoritative.

Client never decides:

Score

Winner

Correct Answer

Timer

Rating

Question Selection

Client only submits answers.

Everything else is validated by the server.

---

# 16. Question Rules

Every question must contain

Unique ID

Topic

Subtopic

Difficulty

Estimated Solving Time

Language

Correct Answer

Explanation

Status

Questions marked inactive shall never appear.

---

# 17. Match Integrity

Both players receive

Same questions

Same order

Same timer

Same scoring rules

No player receives additional information.

---

# 18. Statistics

Every match updates

Accuracy

Average solving speed

Topic accuracy

Rating history

Win streak

Loss streak

Questions solved

Matches played

---

# 19. Practice Rules

Practice mode allows

Unlimited attempts

Immediate explanations

Bookmarks

Retry incorrect questions

Adaptive recommendations (future)

No rating changes.

---

# 20. Tournament Rules

Tournament configuration

Match duration

Question pool

Rating enabled/disabled

Number of rounds

Scoring method

Tie-breaking rules

Defined by administrator.

---

# 21. Friend Match Rules

Host creates room.

Room code generated.

Friend joins.

Host starts game.

No rating changes.

Statistics still recorded.

---

# 22. Draw Rules

A draw occurs when

Score equal

Correct answers equal

Time equal

The higher-rated player's rating stays unchanged.

The lower-rated player gains rating according to the Elo draw calculation.

With equal ratings, neither rating changes.

A draw counts as a draw in statistics (not a win or a loss) and ends the current win streak.

---

# 23. Match Completion

A match ends when

Timer expires

OR

Both players submit all questions

Server immediately calculates

Winner

Rating

Statistics

History

---

# 24. Future Enhancements

Power Play Mode

Team Battles

Clan Wars

Seasonal Rules

Sudden Death Round

Adaptive Question Selection

AI Difficulty Adjustment

Custom Tournament Rules
