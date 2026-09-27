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

	appCache "github.com/exam-arena/internal/cache"
	"github.com/exam-arena/internal/config"
	"github.com/exam-arena/internal/database"
	"github.com/exam-arena/internal/handler"
	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
	"github.com/exam-arena/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

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

	userRepo := repository.NewUserRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	questionRepo := repository.NewQuestionRepo(db)
	topicRepo := repository.NewTopicRepo(db)

	memCache := appCache.NewMemoryCache()

	questionBank := appCache.NewQuestionBank(
		questionRepo.GetAllPublished,
		15*time.Minute,
	)

	catIDs, err := questionRepo.GetActiveCategoryIDs(context.Background())
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		os.Exit(1)
	}
	questionBank.Warm(context.Background(), catIDs)

	bgCtx, bgCancel := context.WithCancel(context.Background())

	questionBank.StartAutoRefresh(bgCtx, questionRepo.GetActiveCategoryIDs)

	hub := ws.NewHub()
	// Record activity on connect and disconnect for the admin "active
	// users" counts.
	hub.SetPresenceHook(func(userID string, _ bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := userRepo.TouchLastActive(ctx, userID); err != nil {
			slog.Warn("failed to record user activity", "user_id", userID, "error", err)
		}
	})

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)

	matchService := service.NewMatchService(
		matchRepo,
		userRepo,
		hub,
		memCache,
		questionBank,
	)

	mmQueue := matchmaking.NewQueue()
	mmEngine := matchmaking.NewEngine(mmQueue, matchService.StartMatchForPair)

	mmService := service.NewMatchmakingService(matchRepo, userRepo, mmQueue)
	if err := mmService.RestoreFromDB(context.Background()); err != nil {
		slog.Warn("queue restore had errors", "error", err)
	}

	go mmEngine.Run(bgCtx)

	practiceService := service.NewPracticeService(repository.NewPracticeRepo(db), questionRepo)
	leaderboardService := service.NewLeaderboardService(repository.NewLeaderboardRepo(db), memCache)
	flagRepo := repository.NewFlagRepo(db)
	dailyService := service.NewDailyService(repository.NewDailyRepo(db), questionRepo)
	friendService := service.NewFriendService(
		repository.NewFriendRepo(db),
		userRepo,
		matchRepo,
		questionRepo,
		matchService,
		hub,
		memCache,
	)

	// ── Handlers ──────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authService)
	matchHandler := handler.NewMatchHandler(matchService, mmService)
	userHandler := handler.NewUserHandler(userRepo, leaderboardService)
	practiceHandler := handler.NewPracticeHandler(practiceService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)
	subjectHandler := handler.NewSubjectHandler(questionRepo)
	topicHandler := handler.NewTopicHandler(topicRepo)
	adminHandler := handler.NewAdminHandler(questionRepo, userRepo, matchRepo, flagRepo, hub)
	flagHandler := handler.NewFlagHandler(flagRepo, questionRepo, questionBank)
	dailyHandler := handler.NewDailyHandler(dailyService)
	friendHandler := handler.NewFriendHandler(friendService)
	wsHandler := handler.NewWSHandler(hub, cfg.JWTSecret, matchService)

	// Start the hub only after every message handler is registered so the
	// handler map is never written while the hub goroutine reads it.
	go hub.Run()

	// ── Router (same as before) ───────────────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","online":%d,"queue":%v}`,
			hub.OnlineCount(), mmQueue.Stats())
	})

	// Auth
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/v1/auth/me", middleware.Auth(cfg.JWTSecret, authHandler.Me))

	// Users
	mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetProfile)
	mux.HandleFunc("GET /api/v1/users/{id}/stats", userHandler.GetStats)
	mux.HandleFunc("GET /api/v1/users/{id}/matches", userHandler.GetMatchHistory)

	// Matches
	mux.HandleFunc("POST /api/v1/matches/queue", middleware.Auth(cfg.JWTSecret, matchHandler.JoinQueue))
	mux.HandleFunc("DELETE /api/v1/matches/queue", middleware.Auth(cfg.JWTSecret, matchHandler.LeaveQueue))
	mux.HandleFunc("GET /api/v1/matches/{id}", middleware.Auth(cfg.JWTSecret, matchHandler.GetMatch))
	mux.HandleFunc("POST /api/v1/matches/friend", middleware.Auth(cfg.JWTSecret, matchHandler.CreateFriendMatch))
	mux.HandleFunc("POST /api/v1/matches/friend/join", middleware.Auth(cfg.JWTSecret, matchHandler.JoinFriendMatch))
	mux.HandleFunc("GET /api/v1/matches/queue/stats", middleware.Auth(cfg.JWTSecret, matchHandler.QueueStats))

	// Friends
	mux.HandleFunc("GET /api/v1/friends", middleware.Auth(cfg.JWTSecret, friendHandler.ListFriends))
	mux.HandleFunc("POST /api/v1/friends/requests", middleware.Auth(cfg.JWTSecret, friendHandler.SendRequest))
	mux.HandleFunc("POST /api/v1/friends/requests/{id}/accept", middleware.Auth(cfg.JWTSecret, friendHandler.AcceptRequest))
	mux.HandleFunc("POST /api/v1/friends/requests/{id}/decline", middleware.Auth(cfg.JWTSecret, friendHandler.DeclineRequest))
	mux.HandleFunc("DELETE /api/v1/friends/{userId}", middleware.Auth(cfg.JWTSecret, friendHandler.RemoveFriend))
	mux.HandleFunc("POST /api/v1/friends/{userId}/challenge", middleware.Auth(cfg.JWTSecret, friendHandler.ChallengeFriend))
	mux.HandleFunc("DELETE /api/v1/friends/challenges/{matchId}", middleware.Auth(cfg.JWTSecret, friendHandler.CloseChallenge))

	// Subjects
	mux.HandleFunc("GET /api/v1/subjects", middleware.Auth(cfg.JWTSecret, subjectHandler.GetAllSubjects))
	mux.HandleFunc("GET /api/v1/subjects/{id}/topics", middleware.Auth(cfg.JWTSecret, topicHandler.ListTopics))

	// Topics
	mux.HandleFunc("GET /api/v1/topics", middleware.Auth(cfg.JWTSecret, topicHandler.ListTopics))
	mux.HandleFunc("GET /api/v1/topics/{id}", middleware.Auth(cfg.JWTSecret, topicHandler.GetTopic))

	// Daily challenge
	mux.HandleFunc("GET /api/v1/daily-challenge", middleware.Auth(cfg.JWTSecret, dailyHandler.Overview))
	mux.HandleFunc("POST /api/v1/daily-challenge/start", middleware.Auth(cfg.JWTSecret, dailyHandler.Start))
	mux.HandleFunc("POST /api/v1/daily-challenge/submit", middleware.Auth(cfg.JWTSecret, dailyHandler.Submit))

	// Question flags
	mux.HandleFunc("POST /api/v1/questions/{id}/flag", middleware.Auth(cfg.JWTSecret, flagHandler.FlagQuestion))

	// Practice
	mux.HandleFunc("POST /api/v1/practice/start", middleware.Auth(cfg.JWTSecret, practiceHandler.StartSession))
	mux.HandleFunc("POST /api/v1/practice/{id}/answer", middleware.Auth(cfg.JWTSecret, practiceHandler.SubmitAnswer))
	mux.HandleFunc("POST /api/v1/practice/{id}/end", middleware.Auth(cfg.JWTSecret, practiceHandler.EndSession))
	mux.HandleFunc("GET /api/v1/practice/{id}", middleware.Auth(cfg.JWTSecret, practiceHandler.GetSession))

	// Leaderboard
	mux.HandleFunc("GET /api/v1/leaderboard/{category}", leaderboardHandler.GetLeaderboard)

	// Admin
	mux.HandleFunc("POST /api/v1/admin/topics", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", topicHandler.CreateTopic)))
	mux.HandleFunc("POST /api/v1/admin/questions", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.CreateQuestion)))
	mux.HandleFunc("PUT /api/v1/admin/questions/{id}/publish", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.PublishQuestion)))
	mux.HandleFunc("GET /api/v1/admin/stats", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.GetSystemStats)))
	mux.HandleFunc("PUT /api/v1/admin/categories/{id}/sort-order", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.SetCategorySortOrder)))
	mux.HandleFunc("GET /api/v1/admin/flags", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", flagHandler.ListFlags)))
	mux.HandleFunc("PUT /api/v1/admin/flags/{questionId}", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", flagHandler.ReviewFlags)))

	// WebSocket
	mux.HandleFunc("GET /ws", wsHandler.HandleConnection)

	// ── Debug (local only) ─────────────────────────────────────────────────
	mux.HandleFunc("GET /debug/state", func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, http.StatusOK, map[string]interface{}{
			"queue":  mmQueue.Stats(),
			"online": hub.OnlineCount(),
		})
	})

	// Full queue snapshot (shows entries per pool)
	mux.HandleFunc("GET /debug/queue", func(w http.ResponseWriter, r *http.Request) {
		snap := mmQueue.Snapshot()
		utils.JSON(w, http.StatusOK, snap)
	})

	// ── Middleware ────────────────────────────────────────────────────
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst, h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	// ── Server ────────────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
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
		slog.Info("server started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down...")
	bgCancel()
	hub.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("stopped")
}
