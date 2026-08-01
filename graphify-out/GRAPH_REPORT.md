# Graph Report - .  (2026-08-01)

## Corpus Check
- 118 files · ~93,306 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 808 nodes · 1497 edges · 68 communities (58 shown, 10 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 78 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Community 0: page.tsx, Home()
- Community 1: auth_handler.go, Request
- Community 2: question_cache.go, Context
- Community 3: question.go, Time
- Community 4: admin_handler.go, Request
- Community 5: main.go, main()
- Community 6: ws_handler.go, RawMessage
- Community 7: engine.go, abs()
- Community 8: user.go, Time
- Community 9: exam_arena_schema.sql, achievements
- Community 10: autoprefixer, canvas-confetti
- Community 11: useBackendAuth.ts, AuthMode
- Community 12: eslint, eslint-config-next
- Community 13: tsconfig.json, compilerOptions
- Community 14: test_ws_match.go, checkLeaderboard()
- Community 15: schema.sql, exam_categories
- Community 16: 001_initial_schema.sql, exam_categories
- Community 17: logging.go, Conn
- Community 18: memory.go, Duration
- Community 19: AuthPage.tsx, AuthPage()
- Community 20: SoundEngine, .constructor()
- Community 21: ratelimit.go, Handler
- Community 22: layout.tsx, inter
- Community 23: NotificationEngine, .constructor()
- Community 24: game_room.go, RWMutex
- Community 25: postgres.go, Context
- Community 26: notification.go, Time
- Community 27: eslint.config.mjs, __dirname
- Community 28: test_friend_match.sh, test_friend_match.sh script
- Community 29: test_match.sh, test_match.sh script
- Community 34: next-env.d.ts, NOTE: This file should not be edited
- Community 35: postcss.config.mjs, config
- Community 66: github.com/exam-arena
- Community 67: test_ws

## God Nodes (most connected - your core abstractions)
1. `main()` - 32 edges
2. `MatchRepo` - 29 edges
3. `UserRepo` - 26 edges
4. `JSONError()` - 23 edges
5. `MatchService` - 22 edges
6. `users` - 20 edges
7. `Hub` - 20 edges
8. `compilerOptions` - 17 edges
9. `QuestionRepo` - 16 edges
10. `PlayerProfile` - 16 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewMemoryCache()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/cache/memory.go
- `main()` --calls--> `NewQuestionBank()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/cache/question_cache.go
- `main()` --calls--> `NewAdminHandler()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/handler/admin_handler.go
- `main()` --calls--> `NewAuthHandler()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/handler/auth_handler.go
- `main()` --calls--> `NewMatchHandler()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/handler/match_handler.go

## Import Cycles
- None detected.

## Communities (68 total, 10 thin omitted)

### Community 0 - "Community 0: page.tsx, Home()"
Cohesion: 0.06
Nodes (65): Home(), DualExamArena(), DualExamArenaProps, getNow(), FriendsAndChatView(), FriendsAndChatViewProps, LeaderboardView(), LeaderboardViewProps (+57 more)

### Community 1 - "Community 1: auth_handler.go, Request"
Cohesion: 0.08
Nodes (30): Request, ResponseWriter, NewAuthHandler(), Request, ResponseWriter, Request, ResponseWriter, NewMatchHandler() (+22 more)

### Community 2 - "Community 2: question_cache.go, Context"
Cohesion: 0.09
Nodes (28): Context, Duration, Question, RWMutex, lcgNext(), NewQuestionBank(), partialShuffle(), generateRoomCode() (+20 more)

### Community 3 - "Community 3: question.go, Time"
Cohesion: 0.08
Nodes (24): Time, Context, DB, NewPracticeRepo(), Context, DB, Question, NewQuestionRepo() (+16 more)

### Community 4 - "Community 4: admin_handler.go, Request"
Cohesion: 0.09
Nodes (19): Request, ResponseWriter, NewAdminHandler(), Time, Context, DB, MatchPlayer, NewMatchRepo() (+11 more)

### Community 5 - "Community 5: main.go, main()"
Cohesion: 0.07
Nodes (27): main(), getEnv(), getEnvFloat(), getEnvInt(), Load(), Context, DB, RunMigrations() (+19 more)

### Community 6 - "Community 6: ws_handler.go, RawMessage"
Cohesion: 0.08
Nodes (17): RawMessage, Request, ResponseWriter, NewWSHandler(), ValidateToken(), Conn, NewClient(), RWMutex (+9 more)

### Community 7 - "Community 7: engine.go, abs()"
Cohesion: 0.10
Nodes (17): Context, Duration, NewEngine(), Mutex, Time, NewQueue(), removeByUserID(), Context (+9 more)

### Community 8 - "Community 8: user.go, Time"
Cohesion: 0.10
Nodes (18): Time, Context, DB, NewUserRepo(), Context, NewAuthService(), GenerateToken(), CheckPassword() (+10 more)

### Community 9 - "Community 9: exam_arena_schema.sql, achievements"
Cohesion: 0.14
Nodes (35): achievements, admin_audit_logs, exam_categories, friendships, match_answers, match_players, match_questions, matches (+27 more)

### Community 10 - "Community 10: autoprefixer, canvas-confetti"
Cohesion: 0.06
Nodes (33): autoprefixer, canvas-confetti, class-variance-authority, clsx, firebase, NotificationCenter(), useIsMobile(), dependencies (+25 more)

### Community 11 - "Community 11: useBackendAuth.ts, AuthMode"
Cohesion: 0.13
Nodes (31): AuthMode, useBackendAuth(), UseBackendAuthResult, BackendMatchHookOptions, useBackendMatch(), UseBackendMatchResult, apiCreateFriendMatch(), apiGetLeaderboard() (+23 more)

### Community 12 - "Community 12: eslint, eslint-config-next"
Cohesion: 0.06
Nodes (32): eslint, eslint-config-next, firebase-tools, devDependencies, eslint, eslint-config-next, firebase-tools, tailwindcss (+24 more)

### Community 13 - "Community 13: tsconfig.json, compilerOptions"
Cohesion: 0.07
Nodes (27): compilerOptions, allowJs, baseUrl, esModuleInterop, incremental, isolatedModules, jsx, lib (+19 more)

### Community 14 - "Community 14: test_ws_match.go, checkLeaderboard()"
Cohesion: 0.21
Nodes (19): checkLeaderboard(), checkPlayerStats(), connectWS(), drainScores(), Conn, Duration, RawMessage, httpGet() (+11 more)

### Community 15 - "Community 15: schema.sql, exam_categories"
Cohesion: 0.29
Nodes (18): exam_categories, match_answers, match_players, match_questions, matches, matchmaking_queue, practice_session_questions, practice_sessions (+10 more)

### Community 16 - "Community 16: 001_initial_schema.sql, exam_categories"
Cohesion: 0.29
Nodes (18): exam_categories, match_answers, match_players, match_questions, matches, matchmaking_queue, practice_session_questions, practice_sessions (+10 more)

### Community 17 - "Community 17: logging.go, Conn"
Cohesion: 0.16
Nodes (10): Conn, Handler, ResponseWriter, Logging(), newWrappedResponseWriter(), Recovery(), wrappedResponseWriter, PushOptions (+2 more)

### Community 18 - "Community 18: memory.go, Duration"
Cohesion: 0.18
Nodes (6): Duration, RWMutex, Time, NewMemoryCache(), cacheItem, MemoryCache

### Community 19 - "Community 19: AuthPage.tsx, AuthPage()"
Cohesion: 0.15
Nodes (8): AuthPage(), AuthPageProps, AuthView, getPasswordStrength(), InputFieldProps, PlayerCardProps, StatBadgeProps, ThemeMode

### Community 21 - "Community 21: ratelimit.go, Handler"
Cohesion: 0.31
Nodes (7): Handler, Mutex, Time, newRateLimiter(), RateLimit(), rateLimiter, visitor

### Community 22 - "Community 22: layout.tsx, inter"
Cohesion: 0.22
Nodes (5): inter, metadata, extends, nextConfig, next

### Community 24 - "Community 24: game_room.go, RWMutex"
Cohesion: 0.33
Nodes (3): RWMutex, NewGameRoom(), GameRoom

### Community 25 - "Community 25: postgres.go, Context"
Cohesion: 0.50
Nodes (3): Context, NewPostgres(), Pool

## Knowledge Gaps
- **91 isolated node(s):** `github.com/exam-arena`, `tags`, `contextKey`, `Topic`, `LeaderboardEntry` (+86 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `Community 5: main.go, main()` to `Community 1: auth_handler.go, Request`, `Community 2: question_cache.go, Context`, `Community 3: question.go, Time`, `Community 4: admin_handler.go, Request`, `Community 6: ws_handler.go, RawMessage`, `Community 7: engine.go, abs()`, `Community 8: user.go, Time`, `Community 17: logging.go, Conn`, `Community 18: memory.go, Duration`, `Community 21: ratelimit.go, Handler`?**
  _High betweenness centrality (0.110) - this node is a cross-community bridge._
- **Why does `MatchService` connect `Community 2: question_cache.go, Context` to `Community 1: auth_handler.go, Request`, `Community 4: admin_handler.go, Request`, `Community 5: main.go, main()`, `Community 6: ws_handler.go, RawMessage`, `Community 8: user.go, Time`?**
  _High betweenness centrality (0.048) - this node is a cross-community bridge._
- **Why does `UserRepo` connect `Community 8: user.go, Time` to `Community 1: auth_handler.go, Request`, `Community 2: question_cache.go, Context`, `Community 4: admin_handler.go, Request`, `Community 7: engine.go, abs()`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Are the 31 inferred relationships involving `main()` (e.g. with `NewMemoryCache()` and `NewQuestionBank()`) actually correct?**
  _`main()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **Are the 21 inferred relationships involving `JSONError()` (e.g. with `Auth()` and `RequireRole()`) actually correct?**
  _`JSONError()` has 21 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/exam-arena`, `tags`, `contextKey` to the rest of the system?**
  _91 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0: page.tsx, Home()` be split into smaller, more focused modules?**
  _Cohesion score 0.06414565826330532 - nodes in this community are weakly interconnected._