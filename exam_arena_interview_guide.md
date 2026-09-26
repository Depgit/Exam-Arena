# Exam Arena — Complete Interview Guide
> **How to use this doc:** Each section answers *"Why did you write this line/pattern?"* — the exact questions interviewers ask when they read your code. Memorise the **bold interview answers**.

---

## 📐 Project Architecture Overview

```
Exam Arena
├── cmd/server/main.go          ← Entry point, wires everything together
├── internal/
│   ├── config/                 ← Environment-driven config
│   ├── database/               ← PostgreSQL pool + migrations
│   ├── cache/                  ← In-memory KV cache + question bank
│   ├── models/                 ← Pure data structs (no DB logic)
│   ├── repository/             ← All SQL queries (data layer)
│   ├── service/                ← Business logic
│   ├── handler/                ← HTTP request/response layer
│   ├── middleware/             ← Auth, CORS, rate-limit, logging
│   ├── matchmaking/            ← Queue + pairing engine
│   ├── ws/                     ← WebSocket hub + client pump
│   └── utils/                  ← JWT, bcrypt, ELO, response helpers
```

**Design Pattern:** This is a classic **3-layer architecture** (Handler → Service → Repository) plus cross-cutting concerns (middleware, cache, WebSocket).

**Interview Q: "Why separate handler/service/repository?"**
> Each layer has one job. Handlers only parse HTTP. Services only run business logic. Repos only talk to the DB. This makes each layer independently testable and replaceable (e.g., swap Postgres for MySQL by rewriting only the repo layer).

---

## 📦 go.mod — Dependency Choices

```go
module github.com/exam-arena

go 1.25.0

require (
    github.com/golang-jwt/jwt/v5  // JWT auth tokens
    github.com/gorilla/websocket  // WebSocket protocol
    github.com/jackc/pgx/v5       // PostgreSQL driver
    github.com/joho/godotenv      // Load .env files
    golang.org/x/crypto           // bcrypt for password hashing
)
```

**Why these specific packages?**

| Package | Why chosen |
|---|---|
| `pgx/v5` | Fastest Go Postgres driver; native `pgxpool` for connection pooling |
| `gorilla/websocket` | The de-facto standard WebSocket library; supports Hijacker interface |
| `golang-jwt/jwt/v5` | v5 is the maintained fork after the original was abandoned |
| `godotenv` | Loads `.env` only in dev; in prod, OS env vars override |
| `golang.org/x/crypto` | Official Go crypto — bcrypt implementation |

**Zero framework (no Gin/Echo/Fiber) — Interview Q: "Why plain net/http?"**
> Go 1.22+ added method-based routing (`GET /api/v1/users/{id}`) directly in the stdlib. Using net/http means zero magic, full control, and one less dependency to audit.

---

## 🚀 cmd/server/main.go — The Wiring Layer

### Lines 26–27: Structured Logging Setup
```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
slog.SetDefault(logger)
```
**Interview answer:** `slog` is Go 1.21's official structured logging package. JSON output is machine-parseable — log aggregators (Loki, Splunk, Datadog) can query fields like `"status": 500` directly. Setting it as default means every package that calls `slog.Info(...)` uses this handler automatically.

---

### Lines 29–33: Fail-fast config loading
```go
cfg, err := config.Load()
if err != nil {
    slog.Error("config", "error", err)
    os.Exit(1)
}
```
**Interview answer:** We exit immediately if config is broken. This is the **fail-fast principle** — a server with a missing `DATABASE_URL` or `JWT_SECRET` is a broken server, not a degraded one. Letting it start would silently fail on every request.

---

### Lines 35–40: Connection Pool Setup
```go
db, err := database.NewPostgres(context.Background(), cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
defer db.Close()
```
**Interview answer:** `pgxpool` manages a pool of persistent DB connections. Without a pool, each request would spend ~10ms on TCP handshake + TLS + Postgres authentication. With a pool (max 25, min 5), connections are reused. `defer db.Close()` ensures graceful shutdown releases those connections back to Postgres.

---

### Lines 42–45: Migrations at startup
```go
if err := database.RunMigrations(context.Background(), db); err != nil {
    slog.Error("migrations", "error", err)
    os.Exit(1)
}
```
**Interview answer:** Running migrations at startup is an idempotent operation — it checks a `schema_migrations` table and skips already-applied ones. This is the simplest zero-downtime migration strategy for small services. In production at scale, you'd move this to a separate `migrate` command run before deploying the new binary.

---

### Lines 47–49: Repository Initialization
```go
userRepo   := repository.NewUserRepo(db)
matchRepo  := repository.NewMatchRepo(db)
questionRepo := repository.NewQuestionRepo(db)
```
**Interview answer:** Repositories are injected with the DB pool at startup, not created on every request. This is **dependency injection** — each service receives its dependencies rather than creating them. Easier to mock in unit tests (`interface` injection).

---

### Lines 51–63: Question Bank Warm-up
```go
memCache := appCache.NewMemoryCache()

questionBank := appCache.NewQuestionBank(
    questionRepo.GetAllPublished,
    15*time.Minute,
)

catIDs, err := questionRepo.GetActiveCategoryIDs(context.Background())
questionBank.Warm(context.Background(), catIDs)
```
**Interview answer:** At startup we pre-load all published questions into memory, grouped by exam category. This is **cache warming**. During a live match, answer correctness is checked in-memory with zero DB calls — critical for real-time performance. The 15-minute refresh ensures newly published questions appear without a restart.

---

### Lines 65–67: Background Context with Cancel
```go
bgCtx, bgCancel := context.WithCancel(context.Background())
questionBank.StartAutoRefresh(bgCtx, questionRepo.GetActiveCategoryIDs)
```
**Interview answer:** `bgCancel` is called during graceful shutdown (line 195). This signals the refresh goroutine to stop without killing ongoing operations. This is the canonical Go pattern for **cooperative goroutine cancellation** using `context.Done()`.

---

### Lines 69–70: WebSocket Hub
```go
hub := ws.NewHub()
go hub.Run()
```
**Interview answer:** The Hub is a **central message broker** for all WebSocket connections. It runs in its own goroutine with channels for registration, unregistration, and incoming messages. This single-goroutine pattern on the Hub's internal state means zero mutexes for message dispatching — only the clients map uses an `RWMutex`.

---

### Lines 82–90: Matchmaking Engine
```go
mmQueue  := matchmaking.NewQueue()
mmEngine := matchmaking.NewEngine(mmQueue, matchService.StartMatchForPair)

mmService := service.NewMatchmakingService(matchRepo, userRepo, mmQueue)
if err := mmService.RestoreFromDB(context.Background()); err != nil {
    slog.Warn("queue restore had errors", "error", err)
}

go mmEngine.Run(bgCtx)
```
**Interview answer:** Three things happen here:
1. The **Queue** holds players waiting for a match (in-memory, thread-safe)
2. The **Engine** polls the queue every 2 seconds and pairs compatible players, calling `matchService.StartMatchForPair` as a callback
3. **RestoreFromDB** recovers the queue from Postgres after a server restart — players don't lose their queue position on deploy

The `StartMatchForPair` is passed as a function value (Go first-class functions), keeping the Engine decoupled from the MatchService — no import cycle.

---

### Lines 106–163: HTTP Router
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /health", ...)
mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetProfile)
mux.HandleFunc("POST /api/v1/matches/queue", middleware.Auth(cfg.JWTSecret, matchHandler.JoinQueue))
```
**Interview answer:** Go 1.22's enhanced `ServeMux` supports method-prefix routing (`GET /path`) and path parameters (`{id}`). `middleware.Auth(...)` wraps handler functions — this is the **decorator/middleware pattern**. Auth middleware validates JWT and injects `userID` into the request context, so handlers don't repeat that logic.

---

### Lines 165–170: Middleware Stack (Order Matters!)
```go
var h http.Handler = mux
h = middleware.CORS(h)
h = middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst, h)
h = middleware.Logging(h)
h = middleware.Recovery(h)
```
**Interview answer:** Middleware wraps outward-in. The **outermost** runs first on request, last on response. Order here:
1. **Recovery** runs outermost — catches any panic from inner layers
2. **Logging** — logs after response is written (gets the status code)
3. **RateLimit** — drops requests before CORS and routing work
4. **CORS** — sets headers before routing

**Why Recovery is outermost:** If Logging itself panicked, Recovery would catch it. If it were innermost, a panic in the mux would bypass it.

---

### Lines 172–200: HTTP Server + Graceful Shutdown
```go
srv := &http.Server{
    Addr:         addr,
    Handler:      h,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGTERM)

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        os.Exit(1)
    }
}()

<-done                          // block until signal
bgCancel()                      // stop background goroutines
hub.Shutdown()                  // close WebSocket connections
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)               // drain in-flight HTTP requests
```
**Interview answer:** This is **graceful shutdown**. When the OS sends SIGTERM (e.g., `kubectl rollout`), the server:
1. Stops accepting new connections
2. Gives in-flight requests 30 seconds to complete
3. Cancels background goroutines (question refresh, matchmaking engine)
4. Closes WebSocket connections

Without timeouts (`ReadTimeout`, `WriteTimeout`), a slow client could hold a connection forever — a **Slowloris attack** vector. `IdleTimeout` reclaims keep-alive connections that went idle.

---

## ⚙️ internal/config/config.go

```go
func Load() (*Config, error) {
    _ = godotenv.Load()   // ignore error — .env may not exist in prod

    cfg := &Config{
        ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
        ...
    }
    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
}
```
**Interview answer:** `_ = godotenv.Load()` intentionally discards the error. In production, `.env` doesn't exist — OS environment variables are set by the container orchestrator (Docker Compose, Kubernetes). In development, `.env` is loaded. The pattern is: **OS env always wins, .env is a dev convenience**.

Mandatory variables (`DATABASE_URL`, `JWT_SECRET`) are validated and return errors. Optional variables have typed helpers (`getEnvInt`, `getEnvFloat`) with sensible defaults.

---

## 🗄️ internal/database/postgres.go

```go
func NewPostgres(ctx context.Context, connStr string, maxConns, minConns int32) (*pgxpool.Pool, error) {
    cfg, _ := pgxpool.ParseConfig(connStr)
    cfg.MaxConns = maxConns
    cfg.MinConns = minConns

    pool, _ := pgxpool.NewWithConfig(ctx, cfg)
    
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping: %w", err)
    }
    return pool, nil
}
```
**Interview answer:** `Ping` is critical — it verifies the DB is actually reachable at startup. Without it, the server starts successfully but every request fails. `%w` in `fmt.Errorf` wraps the error so callers can use `errors.Is()` to inspect the root cause.

**Connection pool sizing interview Q:**
- `MaxConns = 25`: Postgres's default `max_connections = 100`, so one app instance uses 25 slots, leaving room for 3 instances + admin tools.
- `MinConns = 5`: Keeps 5 connections warm to avoid connection overhead on the first request after idle periods.

---

## 🔐 internal/middleware/auth.go

```go
type contextKey string

const (
    UserIDKey   contextKey = "user_id"
    UsernameKey contextKey = "username"
    RoleKey     contextKey = "role"
)
```
**Interview answer:** Using a **custom type** `contextKey` (not `string`) prevents collisions. If two packages both set `ctx.WithValue("user_id", ...)`, they'd overwrite each other since the key type is just `string`. With `contextKey`, only code that imports this package can read these values.

```go
func Auth(secret string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            utils.JSONError(w, http.StatusUnauthorized, "invalid authorization format")
            return
        }

        claims, err := utils.ValidateToken(parts[1], secret)
        
        ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```
**Interview answer:** `SplitN(..., 2)` splits into at most 2 parts — this handles tokens that contain spaces without panic. After validation, claims are injected into the **request context** (not global state), so each goroutine gets its own user identity. `r.WithContext(ctx)` creates a shallow copy of the request with the new context.

```go
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userRole, ok := r.Context().Value(RoleKey).(string)
        if !ok || userRole != role {
            utils.JSONError(w, http.StatusForbidden, "insufficient permissions")
            return
        }
        next.ServeHTTP(w, r)
    }
}
```
**Usage:** `middleware.Auth(secret, middleware.RequireRole("admin", handler))` — chained, only runs if Auth passes.
**Interview Q: Difference between 401 and 403?** 401 = "who are you?" (not authenticated). 403 = "I know who you are, you can't do this" (not authorized).

---

## 🚦 internal/middleware/ratelimit.go — Token Bucket Algorithm

```go
type visitor struct {
    tokens   float64
    lastSeen time.Time
}
```
**Interview answer:** This implements a **Token Bucket** rate limiter from scratch. Each IP gets a bucket with `burst` capacity. Tokens refill at `rate` tokens/second. Each request costs 1 token. If the bucket is empty → 429 Too Many Requests.

```go
func (rl *rateLimiter) allow(ip string) bool {
    elapsed := time.Since(v.lastSeen).Seconds()
    v.tokens += elapsed * rl.rate         // refill
    if v.tokens > float64(rl.burst) {
        v.tokens = float64(rl.burst)      // cap at burst
    }
    v.lastSeen = time.Now()

    if v.tokens < 1 {
        return false                       // reject
    }
    v.tokens--
    return true
}
```
**Why not a fixed window?** Fixed windows have a "boundary burst" problem — a client can make 2× burst requests by hitting end-of-window + start-of-next. Token bucket smooths this out.

```go
func (rl *rateLimiter) cleanup() {
    for {
        time.Sleep(time.Minute)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(rl.visitors, ip)     // GC inactive visitors
            }
        }
        rl.mu.Unlock()
    }
}
```
**Interview answer:** Without cleanup, the `visitors` map grows unboundedly — a memory leak. The janitor goroutine evicts IPs inactive for 3 minutes.

```go
ip := r.RemoteAddr
if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
    ip = forwarded
}
```
**Interview answer:** Behind a load balancer/reverse proxy, `RemoteAddr` is always the proxy IP. `X-Forwarded-For` contains the real client IP. **Security note:** In production, only trust `X-Forwarded-For` from known proxy IPs — a client can spoof this header to bypass rate limiting.

---

## 📝 internal/middleware/logging.go — wrappedResponseWriter

```go
type wrappedResponseWriter struct {
    http.ResponseWriter
    statusCode  int
    wroteHeader bool
}
```
**Interview answer:** Go's `http.ResponseWriter` doesn't expose the status code after writing it. We wrap it to capture the code for logging. The **embedded `http.ResponseWriter`** (anonymous field) means our wrapper automatically implements the interface — we only override the methods we care about.

```go
func (rw *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
        return hj.Hijack()
    }
    panic("underlying ResponseWriter does not implement http.Hijacker")
}
```
**Interview answer:** WebSocket upgrade requires the `http.Hijacker` interface — it takes over the raw TCP connection from the HTTP server. Without forwarding `Hijack()`, the logging middleware would break WebSocket connections. This is **interface forwarding** — delegating to the underlying implementation.

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic recovered", "error", err, ...)
                http.Error(w, `{"success":false,"error":"internal server error"}`, 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```
**Interview answer:** In Go, an unrecovered `panic` kills the entire goroutine — but `net/http` already recovers panics at the connection level. We add our own Recovery to log the panic with context (path, method) and return a proper JSON error instead of the default HTML error page.

---

## 🔑 internal/utils/jwt.go

```go
type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims    // embedded: ExpiresAt, IssuedAt, etc.
}
```
**Interview answer:** Embedding `jwt.RegisteredClaims` adds standard JWT fields (exp, iat, iss) while keeping our custom claims. The library handles expiry validation automatically when `ParseWithClaims` is called.

```go
func ValidateToken(tokenStr, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(secret), nil
    })
}
```
**Interview Q: Why check the signing method?**
> **Algorithm confusion attack.** If a client sends a token signed with `alg: none` or `RS256` and we don't check, a vulnerable library might accept it. We explicitly require HMAC (HS256). The key function verifies the algorithm before returning the secret.

---

## 🔒 internal/utils/password.go

```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}
```
**Interview Q: Why bcrypt? Why cost 12?**
> bcrypt is a **slow** hashing algorithm by design. Cost 12 means 2^12 = 4096 rounds — takes ~250ms on modern hardware. This makes brute-force attacks extremely slow. MD5/SHA are fast (billions/sec on a GPU), making them useless for passwords. The salt is embedded in the output — no separate salt column needed.

---

## 📊 internal/utils/elo.go — ELO Rating System

```go
const (
    DefaultRating = 1200
    KFactor       = 32
)

func CalculateElo(ratingA, ratingB int, scoreA float64) (newA, newB int) {
    ea := expectedScore(ratingA, ratingB)    // probability A wins
    eb := 1.0 - ea
    scoreB := 1.0 - scoreA                   // 1=win, 0.5=draw, 0=loss

    newA = ratingA + int(math.Round(KFactor * (scoreA - ea)))
    newB = ratingB + int(math.Round(KFactor * (scoreB - eb)))
}

func expectedScore(ratingA, ratingB int) float64 {
    return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}
```
**Interview answer:** ELO is the rating system used in chess. The **expected score** formula: if A has rating 1600 and B has 1200, A is expected to win with probability ~91%. If A actually wins, their rating barely changes (small surprise). If B wins, both ratings change dramatically (big surprise).

**K-Factor = 32** means max rating change per game is ±32. Lower K (e.g., 16) is used in high-level chess for stability. Higher K is used for new players.

**Floor at 100**: prevents ratings from going negative, which would break the probability formula.

---

## 🌐 internal/ws/ — WebSocket System

### message.go
```go
type Message struct {
    Type    string                 `json:"type"`
    Payload map[string]interface{} `json:"payload,omitempty"`
}
```
**Interview answer:** This is a **typed envelope pattern** — all WebSocket messages share the same wrapper. The `type` field drives dispatch (like a discriminated union). `omitempty` means empty payload won't be serialized — saves bytes on the wire.

### hub.go
```go
type Hub struct {
    clients    map[string]*Client   // userID → client (O(1) lookup)
    mu         sync.RWMutex
    register   chan *Client         // buffered: 64
    unregister chan *Client
    incoming   chan *IncomingMessage // buffered: 256
    handlers   map[string]MessageHandler
    done       chan struct{}
}
```
**Interview Q: Why channels for register/unregister instead of directly locking the map?**
> The Hub's `Run()` goroutine serializes all map mutations — zero chance of concurrent writes to `clients`. Channels act as a **message queue** between goroutines. Buffered channels (cap=64) prevent blocking when many clients connect simultaneously.

```go
case client := <-h.register:
    h.mu.Lock()
    if existing, ok := h.clients[client.userID]; ok {
        close(existing.send)          // kick old session
        delete(h.clients, client.userID)
    }
    h.clients[client.userID] = client
    h.mu.Unlock()
```
**Interview answer:** **Single session enforcement** — if the same user opens a second WebSocket tab, the old connection is closed. Closing the `send` channel signals `WritePump` to exit (channel range stops). This prevents ghost connections accumulating.

```go
func (h *Hub) SendToUser(userID string, msg Message) {
    h.mu.RLock()
    client, ok := h.clients[userID]
    h.mu.RUnlock()

    if ok {
        client.SendJSON(msg)
    }
}
```
**Interview Q: Why RLock here but Lock in Run()?**
> `SendToUser` is called from many goroutines concurrently (match timer, answer handler, etc.). `RLock` allows multiple concurrent reads — we're only reading the map, not modifying it. The Run() goroutine is the only writer, so it uses `Lock`.

### client.go
```go
const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10   // = 54 seconds
    maxMessageSize = 4096
)
```
**Interview answer:** 
- `pingPeriod = pongWait × 0.9` ensures we send a PING before the read deadline expires. If the client doesn't PONG within 60s, we close the connection — detects silent disconnections (phone screen locked, WiFi lost).
- `maxMessageSize = 4096` bytes caps incoming messages — prevents malformed/malicious large messages from consuming memory.

```go
func (c *Client) ReadPump() {
    defer func() {
        c.hub.unregister <- c    // auto-unregister on exit
        c.conn.Close()
    }()
    ...
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break   // connection closed
        }
        c.hub.incoming <- &IncomingMessage{Client: c, Payload: message}
    }
}
```
**Interview answer:** ReadPump and WritePump run in **separate goroutines** per client. This is the gorilla/websocket recommended pattern — their docs state a connection supports one concurrent reader and one concurrent writer. The `defer` guarantees cleanup regardless of how the loop exits.

```go
// Drain queued messages
n := len(c.send)
for i := 0; i < n; i++ {
    w.Write([]byte("\n"))
    w.Write(<-c.send)
}
```
**Interview answer:** **Message batching** — if multiple messages are queued, they're written in one WebSocket frame instead of multiple syscalls. Reduces per-message overhead.

---

## 🧠 internal/cache/ — In-Memory Cache

### cache.go — Interface Design
```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(key string)
    DeletePrefix(prefix string)
    DeleteIfPresent(key string) bool
}
```
**Interview Q: Why define an interface?**
> `MatchService` depends on `Cache` the interface, not `MemoryCache` the struct. This means:
> 1. **Testability** — tests can pass a mock cache
> 2. **Swappability** — replace with Redis by implementing the same interface
> 3. **No import cycle** — service package doesn't need to import cache package internals

### memory.go
```go
func (mc *MemoryCache) DeleteIfPresent(key string) bool {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    item, ok := mc.items[key]
    if !ok { return false }
    if time.Now().After(item.expiresAt) {
        delete(mc.items, key)
        return false   // expired = not present
    }
    delete(mc.items, key)
    return true
}
```
**Interview answer:** This is an **atomic check-and-delete** operation. The mutex ensures no two goroutines can both see the key as present and both delete it. Used in `endMatch` to prevent **double-ending** a match when the timer goroutine and the `allAnswered` check race.

```go
func (mc *MemoryCache) janitor() {
    ticker := time.NewTicker(1 * time.Minute)
    for {
        select {
        case <-ticker.C:
            mc.mu.Lock()
            now := time.Now()
            for k, v := range mc.items {
                if now.After(v.expiresAt) {
                    delete(mc.items, k)
                }
            }
            mc.mu.Unlock()
        case <-mc.stop:
            return
        }
    }
}
```
**Interview answer:** Lazy expiry (checking TTL in `Get`) alone is insufficient — expired keys accumulate in memory. The janitor goroutine actively evicts them every minute. This is the same pattern Redis uses internally.

### question_cache.go — QuestionBank
```go
func (qb *QuestionBank) GetRandom(categoryID string, n int) ([]models.Question, error) {
    qb.mu.RLock()
    pool, ok := qb.bank[categoryID]
    poolCopy := make([]models.Question, len(pool))
    copy(poolCopy, pool)
    qb.mu.RUnlock()    // ← release before CPU-intensive shuffle
    
    selected := partialShuffle(poolCopy, n)
    return selected, nil
}
```
**Interview Q: Why copy before releasing the lock?**
> The shuffle happens on the copy, so the original pool is never mutated. The lock is held for the minimum time (just the copy), not for the entire shuffle operation. This maximizes read throughput when many matches start simultaneously.

```go
func partialShuffle(pool []models.Question, n int) []models.Question {
    seed := uint64(time.Now().UnixNano())
    for i := 0; i < n; i++ {
        seed = lcgNext(seed)
        j := i + int(seed>>33)%(len(pool)-i)
        pool[i], pool[j] = pool[j], pool[i]
    }
    return pool[:n]
}

func lcgNext(state uint64) uint64 {
    return state*6364136223846793005 + 1442695040888963407  // Knuth's LCG
}
```
**Interview answer:** **Fisher-Yates partial shuffle** — O(n) time, only shuffles the first `n` elements needed. LCG (Linear Congruential Generator) is fast but not cryptographically secure. For exam questions (not cryptography), this is acceptable. If question order must be unpredictable (anti-cheat), use `crypto/rand`.

---

## 🎯 internal/matchmaking/ — Matchmaking System

### queue.go
```go
type Key struct {
    CategoryID string
    MatchType  string
}
```
**Interview answer:** Two players only match if they share the same `Key` (same exam category AND match type). Using a struct as a map key is valid in Go as long as all fields are comparable — strings are comparable.

```go
func (q *Queue) Push(e Entry) {
    q.mu.Lock()
    defer q.mu.Unlock()

    // Remove stale entry first (player re-queued)
    for k, pool := range q.pools {
        q.pools[k] = removeByUserID(pool, e.UserID)
    }
    
    key := Key{CategoryID: e.ExamCategoryID, MatchType: e.MatchType}
    q.pools[key] = append(q.pools[key], e)
}
```
**Interview answer:** Ensuring a player can only be in one pool at a time prevents matching them twice. The cleanup loop is O(total_queue_size) — fine for small queues. At scale (thousands of concurrent players), you'd use a secondary index `userID → Key` for O(1) removal.

```go
func removeByUserID(pool []Entry, userID string) []Entry {
    out := pool[:0]    // reuse underlying array — zero allocation
    for _, e := range pool {
        if e.UserID != userID {
            out = append(out, e)
        }
    }
    return out
}
```
**Interview Q: Why `pool[:0]`?**
> This is a **filter-in-place** trick. `pool[:0]` creates a zero-length slice backed by the same array. `append` reuses the array without allocating new memory. Equivalent to Python's `list.remove()` but zero-alloc.

```go
func (q *Queue) Snapshot() map[Key][]Entry {
    q.mu.Lock()
    defer q.mu.Unlock()

    snap := make(map[Key][]Entry, len(q.pools))
    for k, pool := range q.pools {
        copied := make([]Entry, len(pool))
        copy(copied, pool)
        snap[k] = copied
    }
    return snap
}
```
**Interview answer:** The Engine works on a **snapshot** — a deep copy of the queue taken under lock. After the copy, the lock is released and the Engine does its O(n²) pairing logic without blocking new queue operations. This is the **copy-on-read** pattern.

### engine.go
```go
type Engine struct {
    baseRatingRange    int     // ±100 initially
    ratingExpandPerSec float64 // +5/second
    maxRatingRange     int     // caps at ±400
}
```
**Interview answer:** **Expanding tolerance** — a player waiting 0 seconds only matches within ±100 rating. After 60 seconds, the window expands to ±100 + (60 × 5) = ±400. This prevents infinite queue times while still preferring close-rated matches.

```go
func (e *Engine) findBestPair(pool []Entry) (Entry, Entry, bool) {
    bestDiff := math.MaxFloat64
    
    for i := 0; i < len(pool); i++ {
        for j := i + 1; j < len(pool); j++ {
            tolA := e.tolerance(now.Sub(a.QueuedAt))
            tolB := e.tolerance(now.Sub(b.QueuedAt))
            effectiveTol := math.Max(tolA, tolB)    // most lenient player wins
            
            diff := math.Abs(float64(a.Rating - b.Rating))
            if diff <= effectiveTol && diff < bestDiff {
                bestDiff = diff
                bestI, bestJ = i, j
            }
        }
    }
}
```
**Interview answer:** O(n²) **greedy best-match**. The effective tolerance uses `math.Max` — if Player A has waited 60s (±400 tol) and Player B just joined (±100 tol), the pair is eligible. The player who waited longer sets the window.

**Interview Q: How to scale this to millions of players?**
> - Partition queues by rating band (e.g., 1000-1200, 1200-1400)
> - Use Redis Sorted Sets (ZADD by rating, ZRANGEBYSCORE for range queries)
> - Run multiple Engine instances, one per partition

---

## 🎮 internal/service/match_service.go — The Core Game Logic

### LiveMatch — In-Memory Game State
```go
type LiveMatch struct {
    MatchID      string
    Questions    []models.Question      // WITH correct answers (server-side only)
    PlayerStates map[string]*LivePlayer
    StartedAt    time.Time
    cancelTimer  context.CancelFunc
}
```
**Interview answer:** `LiveMatch` lives in the cache only — it's never written to Postgres. This is the **write-behind / read-through** pattern for hot data. Only the final results are persisted. Key insight: correct answers are in `LiveMatch` but **never sent to clients** — the server is the single source of truth.

### startMatchForPair — 7-Step Flow
```go
// Step 1: Get questions (zero DB call)
questions, err := s.questionBank.GetRandom(pair.ExamCategoryID, QuestionsPerMatch)

// Step 2: Persist match skeleton
match, err := s.matchRepo.CreateMatch(ctx, ...)

// Step 4: Store LiveMatch in cache
s.cache.Set(liveMatchKey(match.ID), liveMatch, 35*time.Minute)

// Step 5: Build player-safe questions (NO correct answers)
playerQuestions := toPlayerQuestions(questions)

// Step 6: Send match_start to both players via WebSocket
s.hub.SendToUser(p.UserID, startMsg)

// Step 7: Start countdown timer
go s.runMatchTimer(timerCtx, match.ID, ...)
```
**Interview answer:** Questions are fetched from memory (Step 1) before touching the DB (Step 2). This means match creation latency is dominated by one DB write, not by question fetching. The cache TTL is 35 minutes > match timer (2 min) — ensures the match doesn't expire from cache mid-game.

### SubmitAnswer — Correctness Without a DB Call
```go
func (s *MatchService) SubmitAnswer(ctx context.Context, req AnswerRequest) error {
    // Load from cache (not DB)
    raw, ok := s.cache.Get(liveMatchKey(req.MatchID))
    lm := raw.(*LiveMatch)

    // Idempotent: ignore duplicate submissions
    if _, alreadyAnswered := player.AnsweredIDs[req.QuestionID]; alreadyAnswered {
        return nil
    }

    // Check correctness in memory
    isCorrect = s.isCorrectOption(lm.Questions, req.QuestionID, req.OptionID)

    // Time bonus scoring
    if isCorrect {
        pointsEarned = PointsPerCorrect    // 100 pts base
        if req.TimeTakenMs < 10_000 { pointsEarned += TimeBonusFull }  // +50
        if req.TimeTakenMs < 30_000 { pointsEarned += TimeBonusHalf }  // +25
    }

    // Async DB write (non-blocking)
    go func() {
        s.matchRepo.SaveAnswer(...)
        s.matchRepo.UpdatePlayerScore(...)
    }()

    // If all answered, end the match
    if lm.allAnswered() {
        lm.cancelTimer()
        go s.endMatch(context.Background(), lm)
    }
}
```
**Interview Q: "Why async DB write in answer submission?"**
> During gameplay, latency must be sub-millisecond. Writing to Postgres (network + disk) would add ~10-50ms per answer. We update the in-memory state immediately (synchronous) and fire-and-forget the DB write (async). The `LiveMatch` cache is the source of truth until `endMatch` runs.

**Interview Q: "What if the async write fails?"**
> The final match result (scores, ELO delta) is written synchronously in `endMatch`. Per-answer audit data could be lost, but the match outcome is preserved. For stricter guarantees, use a retry queue or write-ahead log.

### endMatch — Double-End Prevention
```go
func (s *MatchService) endMatch(ctx context.Context, lm *LiveMatch) {
    // Atomic check-and-delete — only one goroutine proceeds
    if !s.cache.DeleteIfPresent(liveMatchKey(lm.MatchID)) {
        return  // another goroutine already ended it
    }
    ...
}
```
**Interview answer:** Two goroutines can trigger `endMatch` simultaneously: the **timer** (time's up) and the **allAnswered** check (everyone answered). `DeleteIfPresent` is an atomic operation — exactly one goroutine gets `true` and proceeds. The other gets `false` and returns immediately. Without this guard, ELO would be calculated and DB writes would run twice.

### ELO Update Flow
```go
var scoreA float64
switch {
case a.player.Score > b.player.Score: scoreA = 1.0    // A wins
case a.player.Score == b.player.Score: scoreA = 0.5   // draw
default: scoreA = 0.0                                   // A loses
}

newRA, newRB := utils.CalculateElo(ratingA.Rating, ratingB.Rating, scoreA)

// Persist ratings asynchronously
go func() {
    s.userRepo.UpdateRating(bgCtx, ...)
    s.userRepo.InsertRatingHistory(bgCtx, ...)
    s.cache.DeletePrefix("leaderboard:")    // invalidate leaderboard cache
}()
```
**Interview answer:** Leaderboard cache is invalidated after every ranked match (`DeletePrefix("leaderboard:")`) — ensures the next leaderboard request reflects the new ratings. This is **cache invalidation** — one of the famously hard problems in CS.

---

## 🏆 Key Design Decisions — Interview Summary Table

| Decision | Why |
|---|---|
| No web framework (plain net/http) | Go 1.22 stdlib is sufficient; less magic |
| pgxpool connection pool | Reuse TCP connections; tune min/max per load |
| In-memory question bank | Zero DB calls during gameplay; sub-ms latency |
| LiveMatch in cache, not DB | Hot write path; only final results persisted |
| Async DB writes for answers | Non-blocking game loop; ~50ms DB vs ~0ms cache |
| `DeleteIfPresent` for match end | Prevents double-ELO calculation (race condition) |
| Single Engine goroutine for pairing | Zero races inside Engine; Queue handles its own locking |
| RestoreFromDB on startup | Queue survives server restarts / deploys |
| Expanding ELO tolerance | Balances match quality vs queue time |
| Fisher-Yates partial shuffle | O(n) random question selection |
| Middleware chain (outer-in) | Recovery must be outermost to catch all panics |
| Custom contextKey type | Prevents context value collisions across packages |
| `X-Forwarded-For` for rate limiting | Correct IP behind load balancer |
| JSON structured logging (slog) | Machine-parseable; queryable by log aggregators |
| Graceful shutdown (30s) | Drain in-flight requests before process exits |

---

## 🔥 Common Interview Questions & Answers

**Q: How does your WebSocket auth work?**
> The `GET /ws` endpoint checks a JWT token from the `?token=` query param (not headers, since browser WebSocket API can't set headers). The Hub enforces single-session — connecting again kicks the old tab.

**Q: How would you scale this to 10,000 concurrent users?**
> 1. Replace in-memory Queue with Redis Sorted Sets (ZADD/ZRANGEBYSCORE)
> 2. Replace MemoryCache with Redis (distributed cache, survives restarts)
> 3. Replace Hub with Redis Pub/Sub (multiple server instances)
> 4. Horizontal scale app servers behind a load balancer
> 5. Read replicas for leaderboard queries

**Q: What's your biggest bottleneck?**
> The `findBestPair` function is O(n²) per pool per tick. Fine for hundreds of users, but at thousands you'd need a more efficient data structure (sorted array + binary search) or partition the queue by rating band.

**Q: How do you prevent cheating?**
> Correct answers are never sent to clients. `QuestionForPlayer` strips `IsCorrect`. Answer validation happens server-side using the `LiveMatch` struct. A cheating client can submit any option ID, but only the server knows if it's correct.

**Q: What happens if the server crashes mid-match?**
> The match is lost from the cache. DB has the partial state (players, some answers). Players reconnect, get no `live_match` from cache, and the match is effectively abandoned. To fix: use Redis as the cache (survives restarts) and add a match-recovery endpoint.

---

*This document covers every major design decision in Exam Arena from an interview perspective.*
