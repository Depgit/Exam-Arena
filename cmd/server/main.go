// Command server runs the Exam Arena backend.
//
// This file only assembles the pieces, in the order they're needed:
//
//  1. settings and logging
//  2. database (connect + migrate)
//  3. stores — one per feature, the only code that talks to Postgres
//  4. shared building blocks — cache, question bank, realtime hub
//  5. features — the game's rules, one package each
//  6. background jobs
//  7. HTTP routes (routes.go)
//  8. start, then shut down cleanly on Ctrl-C / SIGTERM
//
// To see what a feature does, open its folder under internal/ and read its
// README.md. internal/README.md is the map of all of them.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exam-arena/internal/auth"
	"github.com/exam-arena/internal/daily"
	"github.com/exam-arena/internal/friends"
	"github.com/exam-arena/internal/leaderboard"
	"github.com/exam-arena/internal/match"
	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/platform/cache"
	"github.com/exam-arena/internal/platform/config"
	"github.com/exam-arena/internal/platform/database"
	"github.com/exam-arena/internal/platform/realtime"
	"github.com/exam-arena/internal/practice"
	"github.com/exam-arena/internal/questionpool"
	"github.com/exam-arena/internal/questions"
	"github.com/exam-arena/internal/users"
)

func main() {
	// ── 1. Settings and logging ─────────────────────────────────────────
	// LevelVar so LOG_LEVEL can be applied after config.Load has read .env.
	logLevel := new(slog.LevelVar) // zero value = Info
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	cfg, err := config.Load()
	// LOG_LEVEL: debug | info | warn | error (default info). Applied before
	// the error check so a config failure is still logged at that level.
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if lerr := logLevel.UnmarshalText([]byte(v)); lerr != nil {
			slog.Warn("ignoring invalid LOG_LEVEL", "value", v)
		}
	}
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	// ── 2. Database ─────────────────────────────────────────────────────
	db, err := database.NewPostgres(context.Background(), cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(context.Background(), db); err != nil {
		slog.Error("migrations", "error", err)
		os.Exit(1)
	}

	// ── 3. Stores (SQL lives only in these) ────────────────────────────
	userStore := users.NewStore(db)
	questionStore := questions.NewStore(db)
	topicStore := questions.NewTopicStore(db)
	flagStore := questions.NewFlagStore(db)
	matchStore := match.NewStore(db)
	queueStore := matchmaking.NewStore(db)

	// ── 4. Shared building blocks ──────────────────────────────────────
	bgCtx, stopBackground := context.WithCancel(context.Background())

	memCache := cache.NewMemoryCache()

	// Every published question, kept in memory so starting a match needs
	// no database call. Refreshed every 15 minutes.
	questionBank := questions.NewBank(questionStore.GetAllPublished, 15*time.Minute)
	categoryIDs, err := questionStore.GetActiveCategoryIDs(context.Background())
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		os.Exit(1)
	}
	questionBank.Warm(context.Background(), categoryIDs)
	questionBank.StartAutoRefresh(bgCtx, questionStore.GetActiveCategoryIDs)

	// WebSocket hub: every connected player, and the messages they send.
	hub := realtime.NewHub()

	// ── 5. Features ────────────────────────────────────────────────────
	authService := auth.NewService(userStore, cfg.JWTSecret, cfg.JWTExpiryHours)

	// The queue is created before the match service because a match that
	// one player declines puts the other player back in it.
	queue := matchmaking.NewQueue()
	matchService := match.NewService(matchStore, userStore, hub, memCache, questionBank, queue, queueStore)
	matchmakingService := matchmaking.NewService(queueStore, userStore, queue)

	practiceService := practice.NewService(practice.NewStore(db), questionStore)
	leaderboardService := leaderboard.NewService(leaderboard.NewStore(db), memCache)
	dailyService := daily.NewService(daily.NewStore(db), questionStore)
	friendService := friends.NewService(friends.NewStore(db), userStore, matchStore, questionStore, matchService, hub, memCache)

	// Generated questions: fills and rotates a pool in the background, then
	// refreshes the question bank. Off unless QUESTION_GEN_ENABLED=true.
	questionPool := questionpool.New(questionStore, questionBank, questionpool.Config{
		Enabled:     cfg.QuestionGenEnabled,
		PoolSize:    cfg.QuestionGenPoolSize,
		RotateEvery: cfg.QuestionGenRotateEvery,
		PerUser:     cfg.QuestionGenPerUser,
	})

	hub.SetPresenceHook(func(userID string, online bool) {
		// Record activity on connect/disconnect for the admin "active
		// users" counts.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := userStore.TouchLastActive(ctx, userID); err != nil {
			slog.Warn("failed to record user activity", "user_id", userID, "error", err)
		}
		// A player who closes the tab during the accept window should not
		// make their opponent wait out the whole 30 seconds.
		matchService.HandlePresenceChange(userID, online)
	})

	// Game messages over WebSocket (answers, accept/decline). Registered
	// before the hub starts so its handler map is never written while the
	// hub goroutine reads it.
	match.RegisterGameMessages(hub, matchService)
	connectHandler := realtime.NewConnectHandler(hub, cfg.JWTSecret)
	go hub.Run()

	// ── 6. Background jobs ─────────────────────────────────────────────
	// Put players who were waiting before a restart back in the queue.
	if err := matchmakingService.RestoreFromDB(context.Background()); err != nil {
		slog.Warn("queue restore had errors", "error", err)
	}
	// Every 2s: pair waiting players. ProposeMatch offers the match to both
	// and only starts it once both accept.
	go matchmaking.NewEngine(queue, matchService.ProposeMatch).Run(bgCtx)

	questionPool.Start(bgCtx)
	go removeStaleGuests(bgCtx, userStore)

	// ── 7. HTTP routes ─────────────────────────────────────────────────
	handler := routes(cfg, deps{
		hub:            hub,
		queue:          queue,
		connect:        connectHandler,
		auth:           auth.NewHandler(authService),
		users:          users.NewHandler(userStore),
		categories:     questions.NewSubjectHandler(questionStore),
		topics:         questions.NewTopicHandler(topicStore),
		flags:          questions.NewFlagHandler(flagStore, questionStore, questionBank),
		queueEndpoints: matchmaking.NewHandler(matchmakingService),
		match:          match.NewHandler(matchService),
		friends:        friends.NewHandler(friendService),
		practice:       practice.NewHandler(practiceService),
		daily:          daily.NewHandler(dailyService),
		leaderboard:    leaderboard.NewHandler(leaderboardService),
		generator:      questionpool.NewHandler(questionPool),
		adminStores:    adminStores{questionStore, userStore, matchStore, flagStore},
	})

	// ── 8. Start, then shut down cleanly ───────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting down...")
	stopBackground()
	hub.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("stopped")
}

// removeStaleGuests deletes old demo (guest) accounts that nobody else
// depends on, every 6 hours.
func removeStaleGuests(ctx context.Context, userStore *users.Store) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	for {
		n, err := userStore.DeleteStaleGuests(ctx, 72*time.Hour)
		if err != nil {
			slog.Warn("guest cleanup failed", "error", err)
		} else if n > 0 {
			slog.Info("removed stale guest accounts", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
