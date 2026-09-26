# Low-Level Design — Exam Arena Backend

> **Language:** Go · **DB:** PostgreSQL (pgx/v5 pool) · **Transport:** HTTP/1.1 + WebSocket (gorilla/websocket)  
> **Architecture:** Layered monolith — Handler → Service → Repository → Postgres, with an in-process matchmaking engine and WebSocket hub running as long-lived goroutines.

---

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Package Dependency Graph](#2-package-dependency-graph)
3. [Data Models (Domain Layer)](#3-data-models-domain-layer)
4. [Repository Layer](#4-repository-layer)
5. [Cache Layer](#5-cache-layer)
6. [Matchmaking Engine](#6-matchmaking-engine)
7. [Service Layer](#7-service-layer)
8. [WebSocket Layer](#8-websocket-layer)
9. [Handler Layer (HTTP + WS)](#9-handler-layer)
10. [Middleware](#10-middleware)
11. [Server Bootstrap & Lifecycle](#11-server-bootstrap--lifecycle)
12. [Critical Data Flows](#12-critical-data-flows)
13. [Concurrency Model & Invariants](#13-concurrency-model--invariants)
14. [Scoring & Elo Rating System](#14-scoring--elo-rating-system)
15. [API Surface (Route Map)](#15-api-surface-route-map)
16. [Design Decisions & Trade-offs](#16-design-decisions--trade-offs)

---

## 1. System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Client (Browser / App)                      │
│                 HTTP REST  ◄───►  WebSocket (ws://)                  │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────▼──────────────┐
                    │  Global Middleware Chain     │
                    │  Recovery → Logging →        │
                    │  RateLimit → CORS            │
                    └─────────────┬──────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
       ┌──────▼──────┐    ┌───────▼──────┐   ┌───────▼──────┐
       │  AuthHandler │    │ MatchHandler │   │  WSHandler   │
       │  UserHandler │    │ AdminHandler │   │              │
       │  PracticeH.  │    │ LeaderboardH │   │              │
       │  SubjectH.   │    │              │   │              │
       └──────┬──────┘    └───────┬──────┘   └───────┬──────┘
              │                   │                   │
       ┌──────▼──────────────────▼──────────────────▼──────┐
       │                   Service Layer                     │
       │  AuthService · MatchService · MatchmakingService   │
       │  PracticeService · LeaderboardService              │
       └──────┬──────────────────────────────────┬──────────┘
              │                                  │
    ┌─────────▼─────────┐              ┌─────────▼──────────┐
    │   Repository Layer │              │   In-Memory Layer   │
    │   UserRepo         │              │   MemoryCache       │
    │   MatchRepo        │              │   QuestionBank      │
    │   QuestionRepo     │              │   ws.Hub            │
    │   PracticeRepo     │              │   matchmaking.Queue │
    │   LeaderboardRepo  │              │   matchmaking.Engine│
    └─────────┬─────────┘              └─────────────────────┘
              │
    ┌─────────▼─────────┐
    │     PostgreSQL     │
    │  (pgxpool.Pool)    │
    └───────────────────┘
```

---

## 2. Package Dependency Graph

```
cmd/server/main
    ├── internal/config
    ├── internal/database
    ├── internal/repository      ← uses pgxpool.Pool
    │       ├── UserRepo
    │       ├── MatchRepo
    │       ├── QuestionRepo
    │       ├── PracticeRepo
    │       └── LeaderboardRepo
    ├── internal/cache
    │       ├── Cache (interface)
    │       ├── MemoryCache
    │       └── QuestionBank
    ├── internal/matchmaking
    │       ├── Queue
    │       └── Engine
    ├── internal/service
    │       ├── AuthService       → UserRepo, utils
    │       ├── MatchService      → MatchRepo, UserRepo, ws.Hub, Cache, QuestionBank
    │       ├── MatchmakingService→ MatchRepo, UserRepo, Queue
    │       ├── PracticeService   → PracticeRepo, QuestionRepo
    │       └── LeaderboardService→ LeaderboardRepo, Cache
    ├── internal/ws
    │       ├── Hub
    │       ├── Client
    │       ├── GameRoom
    │       └── Message
    ├── internal/handler
    │       ├── AuthHandler       → AuthService
    │       ├── MatchHandler      → MatchService, MatchmakingService
    │       ├── UserHandler       → UserRepo, LeaderboardService
    │       ├── PracticeHandler   → PracticeService
    │       ├── LeaderboardHandler→ LeaderboardService
    │       ├── SubjectHandler    → QuestionRepo
    │       ├── AdminHandler      → QuestionRepo, UserRepo, MatchRepo
    │       └── WSHandler         → ws.Hub, AuthService, MatchService
    ├── internal/middleware
    │       ├── Auth(JWT)
    │       ├── RequireRole
    │       ├── CORS
    │       ├── RateLimit
    │       └── Logging/Recovery
    └── internal/utils
            ├── JWT (GenerateToken / ValidateToken)
            ├── Password (HashPassword / CheckPassword)
            └── Elo (CalculateElo)
```

---

## 3. Data Models (Domain Layer)

All models live in `internal/models/`.

### 3.1 User

```go
type User struct {
    ID                string     // UUID PK
    Username          string     // unique, 3–30 chars
    Email             string     // unique, hidden in JSON with omitempty
    PasswordHash      string     // bcrypt, never serialised (json:"-")
    DisplayName       *string
    AvatarURL         *string
    CountryCode       *string
    PreferredLanguage string
    Role              string     // "user" | "admin"
    Status            string     // "active" | "suspended" | ...
    EmailVerifiedAt   *time.Time
    LastLoginAt       *time.Time
    CreatedAt, UpdatedAt time.Time
}
```

### 3.2 UserRating

Per-user, per-category Elo rating:

```go
type UserRating struct {
    UserID         string
    ExamCategoryID string
    Rating         int       // Elo, default 1200
    MatchesPlayed  int
    UpdatedAt      time.Time
}
```

### 3.3 UserStatistics

Aggregate career stats, per-user per-category:

```go
type UserStatistics struct {
    UserID, ExamCategoryID  string
    TotalMatches, Wins, Losses, Draws int
    CurrentWinStreak, LongestWinStreak, LongestLosingStreak int
    TotalQuestionsSolved, TotalPracticeSessions int
    OverallAccuracy   float64
    AvgSolvingTimeMs  *int
}
```

### 3.4 LeaderboardEntry (read model)

```go
type LeaderboardEntry struct {
    Rank          int
    UserID, Username string
    DisplayName   *string
    AvatarURL     *string
    Rating        int
    MatchesPlayed int
}
```

### 3.5 Match

```go
type Match struct {
    ID             string     // UUID PK
    MatchType      string     // "ranked" | "friend" | "tournament"
    Status         string     // "waiting" | "starting" | "in_progress" | "completed" | "abandoned"
    ExamCategoryID string
    RoomCode       *string    // only for friend matches (6-char alpha-numeric)
    TournamentID   *string
    TimerSeconds   int        // 120 default
    MaxPlayers     int        // 2
    CreatedAt      time.Time
    StartedAt      *time.Time
    EndedAt        *time.Time
}
```

### 3.6 MatchPlayer

```go
type MatchPlayer struct {
    ID               string
    MatchID, UserID  string
    Username         string    // denormalised for read convenience
    Score            int
    FinalRank        *int
    RatingBefore     *int
    RatingAfter      *int
    RatingDelta      *int
    ConnectionStatus string    // "connected" | "disconnected"
    JoinedAt         time.Time
}
```

### 3.7 MatchQuestion / MatchAnswer

```go
type MatchQuestion struct {
    MatchID, QuestionID string
    OrderIndex          int
}

type MatchAnswer struct {
    ID               string
    MatchID, UserID  string
    QuestionID       string
    SelectedOptionID *string   // nil = skipped / timed out
    IsCorrect        bool
    TimeTakenMs      int
    AnsweredAt       time.Time
}
```

### 3.8 MatchResult / MatchPlayerResult (response DTOs)

```go
type MatchResult struct {
    MatchID  string
    Status   string
    Players  []MatchPlayerResult
    Duration int
}

type MatchPlayerResult struct {
    UserID, Username string
    Score, Rank, RatingDelta, Correct, Total int
}
```

### 3.9 QueueEntry

```go
type QueueEntry struct {
    UserID, ExamCategoryID, MatchType string
    Rating   int
    QueuedAt time.Time
}
```

### 3.10 Question / QuestionOption

```go
type Question struct {
    ID                   string
    ExamCategoryID       string
    TopicID              string
    QuestionType         string     // "mcq_single" | ...
    Difficulty           string     // "easy" | "medium" | "hard"
    Language             string
    Body                 string
    Explanation          *string
    EstimatedTimeSeconds int
    Status               string     // "draft" | "published"
    Options              []QuestionOption
    CreatedAt            time.Time
}

type QuestionOption struct {
    ID, QuestionID string
    OptionText     string
    IsCorrect      bool    // hidden from players during match
    OrderIndex     int
}
```

### 3.11 Player-safe projections

During a live match, the server strips correctness information:

```go
type QuestionForPlayer struct {
    ID, QuestionType, Difficulty, Body string
    EstimatedTimeSeconds int
    Options []OptionForPlayer
    OrderIndex int
}

type OptionForPlayer struct {
    ID, OptionText string
    OrderIndex int
}
```

### 3.12 ExamCategory / Topic

```go
type ExamCategory struct {
    ID, Code, Name, Description string
    IsActive bool
}

type Topic struct {
    ID, ExamCategoryID string
    ParentTopicID *string
    Name string
}
```

### 3.13 Notification

```go
type Notification struct {
    ID      string
    UserID  string
    Type    string
    Payload map[string]interface{}
    IsRead  bool
    CreatedAt time.Time
}
```

---

## 4. Repository Layer

All repositories hold a `*pgxpool.Pool` and expose typed methods. No ORM is used — raw SQL with `pgx` scans.

### 4.1 UserRepo (`internal/repository/user_repo.go`)

| Method | SQL Operation | Notes |
|--------|---------------|-------|
| `Create(ctx, username, email, hash)` | `INSERT INTO users … RETURNING …` | Sets `display_name = username` by default |
| `GetByID(ctx, id)` | `SELECT … WHERE id=$1 AND deleted_at IS NULL` | Soft-delete aware |
| `GetByUsername(ctx, u)` | `SELECT … WHERE username=$1 AND deleted_at IS NULL` | |
| `GetByEmail(ctx, e)` | `SELECT … WHERE email=$1 AND deleted_at IS NULL` | |
| `UpdateLastLogin(ctx, id, ip)` | `UPDATE users SET last_login_at=now(), last_login_ip=$2` | |
| `GetStatistics(ctx, id)` | `SELECT … FROM user_statistics WHERE user_id=$1` | Returns slice (one per category) |
| `GetRatings(ctx, id)` | `SELECT … FROM user_ratings WHERE user_id=$1` | Returns slice |
| `GetRating(ctx, id, catID)` | Single row lookup | Returns `nil` if not found |
| `EnsureRating(ctx, id, catID)` | `INSERT … ON CONFLICT DO NOTHING RETURNING …` | Upserts initial Elo=1200; re-reads on conflict |
| `UpdateRating(ctx, id, catID, newRating, played)` | `UPDATE user_ratings SET rating=$3 …` | |
| `InsertRatingHistory(ctx, …)` | `INSERT INTO rating_history …` | Audit log per match |
| `UpdateStatistics(ctx, id, catID, won, correct, total)` | `INSERT … ON CONFLICT DO UPDATE SET …` | Complex UPSERT for all streak/accuracy fields |
| `GetUserCount(ctx)` | `COUNT(*)` aggregate | Used by admin stats |

### 4.2 MatchRepo (`internal/repository/match_repo.go`)

| Method | SQL Operation | Notes |
|--------|---------------|-------|
| `CreateMatch(ctx, type, catID, timer, maxPlayers, roomCode)` | `INSERT INTO matches … RETURNING …` | |
| `GetMatch(ctx, id)` | Full match row by PK | Returns `nil` if not found |
| `GetMatchByRoomCode(ctx, code)` | `WHERE room_code=$1 AND status IN ('waiting','starting')` | |
| `UpdateMatchStatus(ctx, id, status)` | Dynamically appends `started_at` or `ended_at` | |
| `AddPlayer(ctx, matchID, userID, ratingBefore)` | `INSERT INTO match_players … RETURNING …` | |
| `GetMatchPlayers(ctx, matchID)` | JOIN with `users`, ORDER BY score DESC | |
| `GetPlayerCount(ctx, matchID)` | `COUNT(*)` | |
| `AddMatchQuestions(ctx, matchID, qIDs[])` | pgx Batch INSERT | Uses `pgx.Batch` for bulk insert |
| `GetMatchQuestions(ctx, matchID)` | `SELECT … ORDER BY order_index` | |
| `SaveAnswer(ctx, *MatchAnswer)` | `INSERT … ON CONFLICT DO NOTHING` | Idempotent — duplicate answer ignored |
| `UpdatePlayerScore(ctx, matchID, userID, score)` | `UPDATE match_players SET score=$3` | |
| `UpdatePlayerRating(ctx, matchID, userID, after, delta, rank)` | `UPDATE match_players SET rating_after, rating_delta, final_rank` | |
| `GetMatchAnswers(ctx, matchID, userID)` | `SELECT … WHERE match_id AND user_id` | |
| `GetUserMatchHistory(ctx, userID, limit, offset)` | JOIN matches+players, completed only | Paginated |
| `AddToQueue(ctx, *QueueEntry)` | `INSERT … ON CONFLICT(user_id) DO UPDATE …` | Upsert, mirror of in-memory queue |
| `RemoveFromQueue(ctx, userID)` | `DELETE FROM matchmaking_queue` | |
| `FindMatch(ctx, catID, type, rating, range)` | `WHERE rating BETWEEN … LIMIT 2` | Legacy/backup — Engine uses in-memory queue |
| `GetQueueEntries(ctx)` | `SELECT … ORDER BY queued_at ASC` | Used for startup restore |
| `GetActiveMatchCount(ctx)` | `COUNT(*) WHERE status IN ('waiting','starting','in_progress')` | |
| `GetTotalMatchCount(ctx)` | `COUNT(*)` | |

### 4.3 QuestionRepo (`internal/repository/question_repo.go`)

| Method | Notes |
|--------|-------|
| `GetRandomQuestions(ctx, catID, count)` | `ORDER BY RANDOM() LIMIT $2` — used by PracticeService only |
| `GetByID(ctx, id)` | Loads options via `GetOptions` |
| `GetOptions(ctx, questionID)` | `ORDER BY order_index` |
| `Create(ctx, *Question, options[])` | Transaction: inserts question then options |
| `GetAllPublished(ctx, catID)` | **Bulk-optimised**: loads all questions then all options via `WHERE question_id = ANY($1)` — no N+1 |
| `GetActiveCategoryIDs(ctx)` | Used by QuestionBank to know which categories to warm |
| `GetActiveCategories(ctx)` | Full `ExamCategory` slice |
| `Publish(ctx, questionID)` | `UPDATE questions SET status='published', published_at=now()` |
| `GetQuestionCount(ctx)` | Count of published questions |
| `IsOptionCorrect(ctx, optionID)` | Lookup correctness from DB — used **only** by PracticeService (not live matches) |

### 4.4 PracticeRepo (`internal/repository/practice_repo.go`)

| Method | Notes |
|--------|-------|
| `CreateSession(ctx, userID, catID, topicID, difficulty, count)` | INSERT into `practice_sessions` |
| `AddSessionQuestion(ctx, sessionID, qID, order)` | INSERT into `practice_session_questions` |
| `AnswerQuestion(ctx, sessionID, qID, optionID, correct, timeMs)` | UPDATE the specific question row |
| `EndSession(ctx, sessionID)` | `SET status='completed', ended_at=now()` |
| `GetSession(ctx, sessionID)` | Single row lookup |

**Internal DTO** exposed from this package (not `models/`):

```go
type PracticeSession struct {
    ID, UserID, ExamCategoryID string
    TopicID    *string
    Difficulty *string
    QuestionCount int
    Status     string
}
```

### 4.5 LeaderboardRepo (`internal/repository/leaderboard_repo.go`)

| Method | Notes |
|--------|-------|
| `GetLeaderboard(ctx, catID, limit, offset)` | JOIN `user_ratings` + `users`, uses `ROW_NUMBER() OVER (ORDER BY rating DESC)` window function |
| `GetCategoryByCode(ctx, code)` | Lookup by short code (e.g. `"upsc"`) |

---

## 5. Cache Layer

### 5.1 Cache Interface (`internal/cache/cache.go`)

```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(key string)
    DeletePrefix(prefix string)
    DeleteIfPresent(key string) bool  // atomic check-and-delete
}
```

`DeleteIfPresent` is the critical primitive used to prevent double-ending a match when both the timer goroutine and `allAnswered()` fire concurrently.

### 5.2 MemoryCache (`internal/cache/memory.go`)

| Field | Purpose |
|-------|---------|
| `mu sync.RWMutex` | Read-write lock — readers don't block each other |
| `items map[string]cacheItem` | Key → `{value interface{}, expiresAt time.Time}` |
| `stop chan struct{}` | Used to cleanly stop the background janitor goroutine |

**Janitor goroutine**: Runs every 1 minute, acquires a write lock and evicts all expired entries.

**Key behaviours:**
- `Get`: Returns `nil, false` if key missing **or** expired (lazy expiry check)
- `DeleteIfPresent`: Acquires write lock atomically; returns `false` if expired (treated as "not present")
- `DeletePrefix`: Full O(n) scan — acceptable because the only prefix-deleted key space is `"leaderboard:"` after match end

### 5.3 QuestionBank (`internal/cache/question_cache.go`)

```go
type QuestionBank struct {
    mu              sync.RWMutex
    bank            map[string][]models.Question  // categoryID → questions (with correct answers)
    RefreshInterval time.Duration
    loader          QuestionLoader   // func(ctx, categoryID) ([]Question, error)
}
```

**Warming sequence:**
1. `Warm(ctx, catIDs[])` — called at startup; loads all published questions per category via `QuestionRepo.GetAllPublished` (bulk, no N+1)
2. `StartAutoRefresh(ctx, listCategories)` — spawns goroutine that re-warms every `RefreshInterval` (15 min default)

**GetRandom(categoryID, n):**
1. RLock, copy the pool for the category, RUnlock
2. `partialShuffle(poolCopy, n)` — Fisher-Yates partial shuffle using a simple LCG PRNG seeded by `time.Now().UnixNano()`
3. Returns `pool[:n]`

> **Design note:** The bank holds full `Question` objects including `IsCorrect` on each option. These are used server-side for correctness checking. Players only receive `QuestionForPlayer` (no `IsCorrect`).

---

## 6. Matchmaking Engine

### 6.1 Queue (`internal/matchmaking/queue.go`)

```go
type Key struct { CategoryID, MatchType string }

type Queue struct {
    mu    sync.Mutex
    pools map[Key][]Entry   // one FIFO slice per (category, matchType) pair
}
```

**Entry:**
```go
type Entry struct {
    UserID, Username, ExamCategoryID, MatchType string
    Rating   int
    QueuedAt time.Time
}
```

| Method | Behaviour |
|--------|-----------|
| `Push(e Entry)` | Removes stale entries for same UserID across all pools, then appends to correct pool |
| `Remove(userID)` | Removes from whichever pool the user is in |
| `IsQueued(userID)` | O(n) scan all pools |
| `Snapshot()` | Returns a deep copy of all pools — lock is held only during the copy |
| `ClaimPair(a, b Entry)` | **Atomic** — acquires lock, verifies both are still present, removes them; returns `false` if either is gone |
| `Stats()` | Returns `map[poolKey]count` for monitoring |

### 6.2 Engine (`internal/matchmaking/engine.go`)

```go
type Engine struct {
    queue           *Queue
    onMatchFound    OnMatchFound  // callback = MatchService.StartMatchForPair
    tickInterval    time.Duration // 2s
    baseRatingRange    int        // ±100
    ratingExpandPerSec float64    // 5 per second
    maxRatingRange     int        // ±400 cap
}
```

**Run loop** (single goroutine):
```
every 2 seconds:
    snap = queue.Snapshot()
    for each pool in snap:
        if pool has < 2 players → skip
        processPool(pool)
```

**processPool:**
```
remaining = copy of pool
while len(remaining) >= 2:
    (a, b) = findBestPair(remaining)   // O(n²) scan
    if not found → break
    if !queue.ClaimPair(a, b) → remove both from remaining, retry
    remove a, b from remaining
    go onMatchFound(ctx, PairedPlayers{a, b, ...})
```

**findBestPair logic:**
- For every pair `(i, j)` in pool, compute `effectiveTol = max(tol(a.waitTime), tol(b.waitTime))`
- `tol(waited) = min(baseRatingRange + waited.Seconds() * ratingExpandPerSec, maxRatingRange)`
- Select pair with smallest `|a.Rating - b.Rating|` that still `<= effectiveTol`

**Result:** A player waiting 60 seconds has tolerance ±400 (the cap). A player who just joined has ±100.

---

## 7. Service Layer

### 7.1 AuthService (`internal/service/auth_service.go`)

```go
type AuthService struct {
    userRepo  *repository.UserRepo
    jwtSecret string
    jwtExpiry int  // hours
}
```

| Method | Behaviour |
|--------|-----------|
| `Register(ctx, req)` | Validates username length (3-30), password length (≥8), email format; checks uniqueness; bcrypt-hashes password; creates user; issues JWT |
| `Login(ctx, req)` | Accepts username **or** email (detected by `@`); verifies password hash; checks `status == "active"`; issues JWT |
| `GetUser(ctx, userID)` | Thin pass-through to `userRepo.GetByID` |

**Request/Response types:**
```go
type RegisterRequest  { Username, Email, Password string }
type LoginRequest     { Login, Password string }    // Login = username OR email
type AuthResponse     { Token string; User *models.User }
```

### 7.2 MatchService (`internal/service/match_service.go`)

The most complex service — orchestrates the entire match lifecycle.

```go
type MatchService struct {
    matchRepo    *repository.MatchRepo
    userRepo     *repository.UserRepo
    hub          *ws.Hub
    cache        cache.Cache         // live match states: "live_match:<id>"
    questionBank *cache.QuestionBank // pre-warmed questions with correct answers
}
```

**Constants:**
```go
QuestionsPerMatch = 10
MatchTimerSeconds = 120
PointsPerCorrect  = 100
TimeBonusFull     = 50   // answered in < 10s
TimeBonusHalf     = 25   // answered in < 30s
```

**In-memory LiveMatch state:**
```go
type LiveMatch struct {
    MatchID, CategoryID, MatchType string
    Questions    []models.Question              // includes IsCorrect — never sent to client
    PlayerStates map[string]*LivePlayer         // userID → state
    StartedAt    time.Time
    TimerSeconds int
    cancelTimer  context.CancelFunc             // cancels the timer goroutine
}

type LivePlayer struct {
    UserID, Username string
    Score            int
    AnsweredIDs      map[string]bool            // questionID → wasCorrect
}
```

**Cache key:** `"live_match:" + matchID`  
**Cache TTL:** 35 minutes (matches are max 2 minutes, extra buffer for post-match reads)

#### 7.2.1 StartMatchForPair (ranked match flow)

Called by Engine's callback. Runs in its own goroutine.

```
1. questionBank.GetRandom(categoryID, 10)  ← zero DB calls
2. matchRepo.CreateMatch(...)              ← persist skeleton
3. matchRepo.AddMatchQuestions(...)        ← batch insert question IDs
4. for each player: matchRepo.AddPlayer(...)
5. matchRepo.UpdateMatchStatus("in_progress")
6. Build LiveMatch, store in cache (35 min TTL)
7. toPlayerQuestions(questions)            ← strip IsCorrect
8. hub.SendToUser(each player, "match_start" WS msg)
9. go runMatchTimer(120s)
```

#### 7.2.2 SubmitAnswer

Called from WebSocket handler, synchronous.

```
1. cache.Get("live_match:<id>") → LiveMatch
2. Validate player is in match
3. Idempotency: if question already answered → return nil
4. isCorrect = isCorrectOption(lm.Questions, questionID, optionID)  ← in-memory, no DB
5. Calculate points:
   - Correct: 100 base
   + 50 if timeTakenMs < 10_000
   + 25 if timeTakenMs < 30_000
6. Update player.Score and player.AnsweredIDs in memory
7. go { matchRepo.SaveAnswer(...); matchRepo.UpdatePlayerScore(...) }  ← async persist
8. hub.SendToUser(all players, "score_update" with full scoreboard)
9. if lm.allAnswered() → lm.cancelTimer(); go endMatch(lm)
```

#### 7.2.3 endMatch

```
1. cache.DeleteIfPresent("live_match:<id>") → if false, another goroutine already ended it → return
2. matchRepo.UpdateMatchStatus("completed")
3. Sort players by score descending
4. if 2-player match:
   a. userRepo.GetRating(each player, categoryID)
   b. utils.CalculateElo(ratingA, ratingB, scoreA)  ← scoreA = 1.0/0.5/0.0
   c. go {
       userRepo.UpdateRating(A, B)
       userRepo.InsertRatingHistory(A, B)
       matchRepo.UpdatePlayerRating(A, B)
       userRepo.UpdateStatistics(A, B)
       cache.DeletePrefix("leaderboard:")   ← invalidate leaderboard cache
   }
5. hub.SendToUser(all players, "match_end" msg with ranked results)
```

#### 7.2.4 runMatchTimer

```
ticker(10s)  → broadcasts "time_update" to all players with remaining seconds
deadline(120s) → if fires, calls go endMatch(lm)
ctx.Done()   → cancelled by cancelTimer() when match ends naturally → returns
```

#### 7.2.5 CreateFriendMatch

```
1. Generate 6-char room code (chars: ABCDEFGHJKLMNPQRSTUVWXYZ23456789)
2. matchRepo.CreateMatch("friend", catID, 120, 2, &roomCode)
3. userRepo.EnsureRating(creator, catID)
4. matchRepo.AddPlayer(matchID, creator, ratingBefore)
5. Return (match, roomCode, nil)
```

#### 7.2.6 JoinFriendMatch

```
1. matchRepo.GetMatchByRoomCode(roomCode)      ← only "waiting"/"starting" matches
2. matchRepo.GetMatchPlayers(matchID)
3. Guard: creator can't join own room
4. Guard: room must not be full
5. userRepo.EnsureRating(joiner, catID)
6. matchRepo.AddPlayer(matchID, joiner, ratingBefore)
7. if newPlayerCount >= maxPlayers → go startFriendMatchGame(match, allPlayers)
8. Return match
```

`startFriendMatchGame` is identical to `StartMatchForPair` but operates on already-persisted players.

#### 7.2.7 GetMatchDetails

```
1. matchRepo.GetMatch(matchID)
2. matchRepo.GetMatchPlayers(matchID)
3. if status == "in_progress" AND cache.Get("live_match:<id>") exists:
   → attach LiveScores (from cache scoreboard) and Questions (player-safe)
4. Return MatchDetails
```

### 7.3 MatchmakingService (`internal/service/matchmaking_service.go`)

```go
type MatchmakingService struct {
    matchRepo *repository.MatchRepo
    userRepo  *repository.UserRepo
    queue     *matchmaking.Queue
}
```

| Method | Behaviour |
|--------|-----------|
| `JoinQueue(ctx, userID, username, req)` | `userRepo.EnsureRating` → build `Entry` → `queue.Push` → `matchRepo.AddToQueue` (best-effort mirror) |
| `LeaveQueue(ctx, userID)` | `queue.Remove` → `matchRepo.RemoveFromQueue` |
| `IsQueued(userID)` | `queue.IsQueued` (in-memory only) |
| `QueueStats()` | `queue.Stats()` — returns pool depth map |
| `RestoreFromDB(ctx)` | Reads `matchmaking_queue` table → fetches username per user → calls `queue.Push` for each |

### 7.4 PracticeService (`internal/service/practice_service.go`)

```go
type PracticeService struct {
    practiceRepo *repository.PracticeRepo
    questionRepo *repository.QuestionRepo
}
```

| Method | Behaviour |
|--------|-----------|
| `StartSession(ctx, userID, req)` | Clamps `QuestionCount` to [1,50]; `practiceRepo.CreateSession`; `questionRepo.GetRandomQuestions` (DB random); adds each question to session; returns player-safe questions |
| `SubmitAnswer(ctx, userID, sessionID, req)` | Validates session ownership; `questionRepo.IsOptionCorrect` (DB lookup — unlike live match); `practiceRepo.AnswerQuestion`; fetches explanation |
| `EndSession(ctx, userID, sessionID)` | Validates ownership; `practiceRepo.EndSession` |
| `GetSession(ctx, userID, sessionID)` | Validates ownership; returns `PracticeSession` |

> **Key difference from live match:** Practice mode hits the DB per answer (`IsOptionCorrect`) because there's no in-memory `LiveMatch` state. This is acceptable since practice is single-player.

### 7.5 LeaderboardService (`internal/service/leaderboard_service.go`)

```go
type LeaderboardService struct {
    repo  *repository.LeaderboardRepo
    cache cache.Cache
}
```

| Method | Behaviour |
|--------|-----------|
| `GetLeaderboard(ctx, categoryCode, limit, offset)` | Cache key = `"leaderboard:<code>:<limit>:<offset>"`; TTL = 1 minute; on miss: `repo.GetCategoryByCode` then `repo.GetLeaderboard`; adjusts rank for offset |
| `GetCategoryByCode(ctx, code)` | Pass-through to repo |

**Cache invalidation:** `endMatch` calls `cache.DeletePrefix("leaderboard:")` after updating ratings, ensuring stale leaderboard data is evicted within 1 minute at most.

---

## 8. WebSocket Layer

### 8.1 Message (`internal/ws/message.go`)

```go
type Message struct {
    Type    string                 `json:"type"`
    Payload map[string]interface{} `json:"payload,omitempty"`
}
```

All WebSocket messages (both client→server and server→client) use this envelope.

### 8.2 Client (`internal/ws/client.go`)

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte  // buffered(256) — outbound message queue
    userID   string
    username string
}
```

**Constants:**
```
writeWait      = 10s
pongWait       = 60s
pingPeriod     = 54s   (9/10 of pongWait)
maxMessageSize = 4096 bytes
```

**ReadPump** (goroutine): Reads messages from WebSocket, forwards to `hub.incoming` channel. Deregisters client on disconnect.

**WritePump** (goroutine): Drains `client.send` channel and writes to WebSocket. Sends periodic pings. Closes connection on channel close.

**SendJSON**: Marshals `Message` to JSON, non-blocking send to `client.send`; drops message if buffer full (logged as warning).

### 8.3 Hub (`internal/ws/hub.go`)

```go
type Hub struct {
    clients    map[string]*Client   // userID → client (O(1) lookup)
    mu         sync.RWMutex
    register   chan *Client          // buffered(64)
    unregister chan *Client          // buffered(64)
    incoming   chan *IncomingMessage // buffered(256)
    handlers   map[string]MessageHandler
    done       chan struct{}
}

type MessageHandler func(client *Client, payload json.RawMessage)
```

**Run loop** (single goroutine — serialises all client state mutations):
```
select:
  case <-register:   close existing connection for user (single-session), add new client
  case <-unregister: remove client from map, close send channel
  case <-incoming:   handleMessage(msg)
  case <-done:       close all client send channels, return
```

**Key methods:**
- `SendToUser(userID, msg)` — RLock, lookup, RUnlock, then `client.SendJSON` (non-blocking)
- `Broadcast(msg)` — RLock, marshal once, send to all clients' channels
- `IsOnline(userID)` / `OnlineCount()` — RLock queries

**Single-session enforcement:** When a new connection for an already-connected `userID` arrives, the old client's `send` channel is closed (causing its WritePump to exit and disconnect the WebSocket).

### 8.4 GameRoom (`internal/ws/game_room.go`)

```go
type GameRoom struct {
    mu        sync.RWMutex
    MatchID   string
    PlayerIDs []string
    Scores    map[string]int
    Active    bool
}
```

> **Note:** `GameRoom` is defined but not wired into the main flow. The live match state is tracked by `LiveMatch` in `MemoryCache`. `GameRoom` may be reserved for a future multi-room broadcasting abstraction.

---

## 9. Handler Layer

### 9.1 AuthHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/auth/register` | POST | `Register` | None |
| `/api/v1/auth/login` | POST | `Login` | None |
| `/api/v1/auth/me` | GET | `Me` | JWT |

### 9.2 UserHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/users/{id}` | GET | `GetProfile` | None |
| `/api/v1/users/{id}/stats` | GET | `GetStats` | None |
| `/api/v1/users/{id}/matches` | GET | `GetMatchHistory` | None |

### 9.3 MatchHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/matches/queue` | POST | `JoinQueue` | JWT |
| `/api/v1/matches/queue` | DELETE | `LeaveQueue` | JWT |
| `/api/v1/matches/queue/stats` | GET | `QueueStats` | JWT |
| `/api/v1/matches/{id}` | GET | `GetMatch` | JWT |
| `/api/v1/matches/friend` | POST | `CreateFriendMatch` | JWT |
| `/api/v1/matches/friend/join` | POST | `JoinFriendMatch` | JWT |

### 9.4 PracticeHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/practice/start` | POST | `StartSession` | JWT |
| `/api/v1/practice/{id}/answer` | POST | `SubmitAnswer` | JWT |
| `/api/v1/practice/{id}/end` | POST | `EndSession` | JWT |
| `/api/v1/practice/{id}` | GET | `GetSession` | JWT |

### 9.5 LeaderboardHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/leaderboard/{category}` | GET | `GetLeaderboard` | None |

### 9.6 SubjectHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/subjects` | GET | `GetAllSubjects` | JWT |

### 9.7 AdminHandler

| Route | Method | Handler | Auth |
|-------|--------|---------|------|
| `/api/v1/admin/questions` | POST | `CreateQuestion` | JWT + Role=admin |
| `/api/v1/admin/questions/{id}/publish` | PUT | `PublishQuestion` | JWT + Role=admin |
| `/api/v1/admin/stats` | GET | `GetSystemStats` | JWT + Role=admin |

`CreateQuestion` request body:
```json
{
  "exam_category_id": "uuid",
  "topic_id":         "uuid",
  "question_type":    "mcq_single",
  "difficulty":       "medium",
  "body":             "Question text",
  "explanation":      "optional",
  "estimated_time_seconds": 60,
  "options": [
    { "option_text": "A", "is_correct": false },
    { "option_text": "B", "is_correct": true }
  ]
}
```

### 9.8 WSHandler

```go
type WSHandler struct {
    hub          *ws.Hub
    jwtSecret    string
    matchService *service.MatchService
}
```

**Upgrade flow** (`GET /ws?token=<JWT>`):
1. Extract `token` from query param (browsers can't set headers on WS handshake)
2. `utils.ValidateToken(token, secret)`
3. `upgrader.Upgrade(w, r, nil)` → gorilla WebSocket conn
4. `ws.NewClient(hub, conn, claims.UserID, claims.Username)`
5. `hub.Register(client)`
6. `go client.WritePump()` and `go client.ReadPump()`
7. Send `connected` message back to client

**Registered message handlers:**

| `type` field | Handler | Description |
|-------------|---------|-------------|
| `submit_answer` | `handleSubmitAnswer` | Routes to `MatchService.SubmitAnswer` |
| `ping` | `handlePing` | Returns `pong` response |

**submit_answer payload:**
```json
{
  "match_id":     "uuid",
  "question_id":  "uuid",
  "option_id":    "uuid",
  "time_taken_ms": 4321
}
```

---

## 10. Middleware

All middleware wraps `http.Handler` using the standard adapter pattern.

| Middleware | File | Description |
|-----------|------|-------------|
| `Auth(secret, next)` | `middleware/auth.go` | Validates `Authorization: Bearer <JWT>`; injects `user_id`, `username`, `role` into request context |
| `RequireRole(role, next)` | `middleware/auth.go` | Reads `role` from context, returns 403 if mismatch |
| `CORS(next)` | `middleware/cors.go` | Sets CORS headers |
| `RateLimit(rps, burst, next)` | `middleware/ratelimit.go` | Token bucket rate limiter per IP |
| `Logging(next)` | `middleware/logging.go` | Structured request logging with duration |
| `Recovery(next)` | `middleware/logging.go` | Panic recovery, returns 500 |

**Middleware application order (outermost first):**
```
Recovery → Logging → RateLimit → CORS → Router (mux)
```

**Context keys:**
```go
const (
    UserIDKey   contextKey = "user_id"
    UsernameKey contextKey = "username"
    RoleKey     contextKey = "role"
)
```

---

## 11. Server Bootstrap & Lifecycle

**Startup sequence in `main.go`:**

```
1. Load config (env vars via config.Load())
2. Connect to PostgreSQL (pgxpool with min/max conns)
3. Run database migrations
4. Initialise repositories (UserRepo, MatchRepo, QuestionRepo)
5. Create MemoryCache
6. Create QuestionBank
7. Warm QuestionBank with all active categories (blocks until done)
8. Start QuestionBank auto-refresh goroutine (every 15 min)
9. Start ws.Hub goroutine
10. Initialise services (Auth, Match, Matchmaking, Practice, Leaderboard)
11. Create matchmaking Queue and Engine
12. RestoreFromDB — reload queue entries from Postgres
13. Start Engine goroutine (ticks every 2s)
14. Register all HTTP handlers and routes
15. Apply middleware chain
16. Start HTTP server
17. Wait for SIGINT/SIGTERM
```

**Graceful shutdown:**
```
1. bgCancel()       — stops QuestionBank refresh, Engine, any ctx-aware goroutines
2. hub.Shutdown()   — closes all client send channels
3. srv.Shutdown(30s timeout) — drains in-flight HTTP requests
```

---

## 12. Critical Data Flows

### 12.1 Ranked Match Flow (End-to-End)

```
Player A                  Player B             Server
   │  POST /matches/queue     │                   │
   │─────────────────────────────────────────────►│
   │                          │                   │ EnsureRating, queue.Push, DB mirror
   │  {status: "queued"}      │                   │
   │◄─────────────────────────────────────────────│
   │                          │  POST /matches/queue │
   │                          │──────────────────►│
   │                          │                   │ queue.Push
   │                          │  {status: queued} │
   │                          │◄──────────────────│
   │                          │                   │
   │  GET /ws?token=...       │                   │  Engine tick (every 2s)
   │─────────────────────────────────────────────►│  findBestPair → ClaimPair
   │  connected               │                   │  go StartMatchForPair
   │◄─────────────────────────────────────────────│
   │                          │  GET /ws?token=...│
   │                          │──────────────────►│
   │                          │  connected        │
   │                          │◄──────────────────│
   │  WS: match_start         │  WS: match_start  │
   │◄─────────────────────────│◄──────────────────│ hub.SendToUser × 2
   │                          │                   │ timer goroutine starts
   │  WS: submit_answer       │                   │
   │─────────────────────────────────────────────►│ isCorrectOption (in-memory)
   │                          │                   │ update LivePlayer score
   │  WS: score_update        │  WS: score_update │ go save to DB (async)
   │◄─────────────────────────│◄──────────────────│ broadcastScoreUpdate
   │                          │                   │
   │  (after all answered or timer fires)         │
   │  WS: match_end           │  WS: match_end    │ endMatch:
   │◄─────────────────────────│◄──────────────────│   Elo calc → async DB persist
                                                      leaderboard cache invalidate
```

### 12.2 Friend Match Flow

```
Player A                  Player B
   │  POST /matches/friend                    → CreateFriendMatch
   │  ← { match_id, room_code: "ABC123" }
   │
   │  (share "ABC123" out-of-band)
   │
   │  GET /ws?token=...   (Player A connects WS)
   │
   │                       POST /matches/friend/join { room_code: "ABC123" }
   │                                                 → JoinFriendMatch
   │                                                 → go startFriendMatchGame
   │  WS: match_start      WS: match_start   ← hub.SendToUser × 2
```

### 12.3 Practice Session Flow

```
POST /practice/start → CreateSession (DB) + GetRandomQuestions (DB RANDOM()) 
                     → Returns session_id + QuestionForPlayer[]
POST /practice/{id}/answer → Validate session ownership
                           → IsOptionCorrect (DB) 
                           → AnswerQuestion (DB UPDATE)
                           → Returns {is_correct, explanation}
POST /practice/{id}/end → EndSession (DB UPDATE status=completed)
```

---

## 13. Concurrency Model & Invariants

### Goroutines in play at steady state

| Goroutine | Count | Purpose |
|-----------|-------|---------|
| `hub.Run()` | 1 | Serialises all client registration/message routing |
| `client.ReadPump()` | 1 per connected user | Reads from WebSocket |
| `client.WritePump()` | 1 per connected user | Writes to WebSocket |
| `Engine.Run()` | 1 | Matchmaking tick loop |
| `QuestionBank.StartAutoRefresh` | 1 | Periodic question reload |
| `MemoryCache.janitor` | 1 | Cache eviction |
| `runMatchTimer` | 1 per active match | Timer + time_update broadcasts |
| `endMatch` goroutine | 0 or 1 per match | Post-match Elo/stats writes |
| Answer save goroutines | 0–N | Async answer persistence |

### Synchronisation points

| Resource | Synchronisation |
|----------|----------------|
| `ws.Hub.clients` | `sync.RWMutex` — multiple readers (SendToUser, IsOnline) concurrent with one writer (register/unregister) |
| `matchmaking.Queue.pools` | `sync.Mutex` — single lock for Push/Remove/ClaimPair/Snapshot |
| `cache.MemoryCache.items` | `sync.RWMutex` |
| `cache.QuestionBank.bank` | `sync.RWMutex` |
| `ws.GameRoom.Scores` | `sync.RWMutex` (not currently used in main flow) |

### Critical invariant: Double-end prevention

When a match ends, two goroutines could trigger `endMatch` concurrently:
1. `runMatchTimer` deadline fires
2. `SubmitAnswer` detects `allAnswered()`

Both call `endMatch(lm)`. The first line is:
```go
if !s.cache.DeleteIfPresent(liveMatchKey(lm.MatchID)) {
    return  // already ended
}
```
`DeleteIfPresent` holds the write lock atomically — only one caller gets `true`. The other returns immediately. This prevents double Elo updates, double WebSocket messages, etc.

---

## 14. Scoring & Elo Rating System

### 14.1 Per-question scoring

```
Base: 100 points (if correct)
+ 50 bonus if timeTakenMs < 10,000  (answered in < 10 seconds)
+ 25 bonus if timeTakenMs < 30,000  (answered in < 30 seconds)

Wrong or skipped: 0 points
```

Maximum possible score per question: **150 points** (correct + time bonus full)  
Maximum possible match score (10 questions): **1,500 points**

### 14.2 Elo calculation

Uses `utils.CalculateElo(ratingA, ratingB, scoreA float64) (newA, newB int)`.

`scoreA`:
- `1.0` — player A has higher score (wins)
- `0.5` — tied score (draw)
- `0.0` — player A has lower score (loses)

Standard Elo formula with K-factor:
```
expected_A = 1 / (1 + 10^((ratingB - ratingA) / 400))
delta_A = K * (scoreA - expected_A)
newRatingA = ratingA + delta_A
```

Elo updates are persisted **asynchronously** after `match_end` is sent to players.

### 14.3 Statistics update (UPSERT)

`userRepo.UpdateStatistics` uses a complex SQL UPSERT that atomically updates:
- `total_matches += 1`
- `wins` or `losses += 1`
- `current_win_streak` (reset to 0 on loss)
- `longest_win_streak = GREATEST(current, new_streak)`
- `longest_losing_streak` (approximated)
- `total_questions_solved += correctCount`
- `overall_accuracy` (rolling weighted average)

---

## 15. API Surface (Route Map)

| Method | Path | Auth | Handler |
|--------|------|------|---------|
| GET | `/health` | None | inline — returns JSON with online count + queue stats |
| GET | `/debug/state` | None | inline — queue stats + online count |
| GET | `/debug/queue` | None | inline — full queue snapshot |
| GET | `/ws` | JWT via query param | `wsHandler.HandleConnection` |
| POST | `/api/v1/auth/register` | None | `authHandler.Register` |
| POST | `/api/v1/auth/login` | None | `authHandler.Login` |
| GET | `/api/v1/auth/me` | JWT | `authHandler.Me` |
| GET | `/api/v1/users/{id}` | None | `userHandler.GetProfile` |
| GET | `/api/v1/users/{id}/stats` | None | `userHandler.GetStats` |
| GET | `/api/v1/users/{id}/matches` | None | `userHandler.GetMatchHistory` |
| POST | `/api/v1/matches/queue` | JWT | `matchHandler.JoinQueue` |
| DELETE | `/api/v1/matches/queue` | JWT | `matchHandler.LeaveQueue` |
| GET | `/api/v1/matches/queue/stats` | JWT | `matchHandler.QueueStats` |
| GET | `/api/v1/matches/{id}` | JWT | `matchHandler.GetMatch` |
| POST | `/api/v1/matches/friend` | JWT | `matchHandler.CreateFriendMatch` |
| POST | `/api/v1/matches/friend/join` | JWT | `matchHandler.JoinFriendMatch` |
| GET | `/api/v1/subjects` | JWT | `subjectHandler.GetAllSubjects` |
| POST | `/api/v1/practice/start` | JWT | `practiceHandler.StartSession` |
| POST | `/api/v1/practice/{id}/answer` | JWT | `practiceHandler.SubmitAnswer` |
| POST | `/api/v1/practice/{id}/end` | JWT | `practiceHandler.EndSession` |
| GET | `/api/v1/practice/{id}` | JWT | `practiceHandler.GetSession` |
| GET | `/api/v1/leaderboard/{category}` | None | `leaderboardHandler.GetLeaderboard` |
| POST | `/api/v1/admin/questions` | JWT + admin | `adminHandler.CreateQuestion` |
| PUT | `/api/v1/admin/questions/{id}/publish` | JWT + admin | `adminHandler.PublishQuestion` |
| GET | `/api/v1/admin/stats` | JWT + admin | `adminHandler.GetSystemStats` |

---

## 16. Design Decisions & Trade-offs

| Decision | Rationale | Trade-off |
|----------|-----------|-----------|
| **In-memory LiveMatch** instead of DB per-answer lookup | Zero DB calls during gameplay = sub-millisecond answer processing | State lost on server crash — match would be abandoned |
| **Single Engine goroutine** for matchmaking | No mutex needed inside Engine; queue provides its own concurrency | Maximum throughput = 1 pairing cycle per tick (2s); acceptable for current scale |
| **QuestionBank pre-warms all published questions** | GetRandom is instant (in-memory shuffle) vs `ORDER BY RANDOM()` which is slow on large tables | Memory usage grows with question count; 15-min refresh may serve stale data briefly |
| **Async answer persistence** | Handler returns immediately; DB write doesn't block the player experience | In a crash window between answer accepted and DB write, the answer would be lost from Postgres (LiveMatch is still the source of truth until match ends) |
| **Async Elo/stats write after match_end WS send** | Players see results immediately; DB writes are fire-and-forget | Rating update failure is not surfaced to players; requires monitoring |
| **`DeleteIfPresent` for double-end guard** | Atomic check-and-delete prevents duplicate Elo calculations without requiring explicit match-level locks | Requires cache to be the single source of truth for liveness |
| **DB mirror of matchmaking queue** | Allows queue restore after server restart | Queue can diverge from DB if server crashes mid-operation |
| **pgx Batch for match questions** | Bulk insert of 10 question IDs in a single round-trip | None significant |
| **Leaderboard TTL = 1 minute** | Near-realtime without hammering DB on every request | Leaderboard can be up to 1 minute stale; invalidated immediately after each match end |
| **No ORM** | Full SQL control, predictable query plans | More verbose, schema changes require manual query updates |

---

*Generated from source analysis of commit at 2026-09-26. All file paths relative to `/Users/batman/Documents/backend/Exam_Arena/`.*
