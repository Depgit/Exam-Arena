package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	appCache "github.com/exam-arena/internal/cache"
	"github.com/exam-arena/internal/config"
	"github.com/exam-arena/internal/database"
	"github.com/exam-arena/internal/handler"
	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/service"
)

var adapter *httpadapter.HandlerAdapter

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config loading error", "error", err)
		os.Exit(1)
	}

	db, err := database.InitDB(context.Background(), cfg)
	if err != nil {
		slog.Error("database connection error", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(context.Background(), db); err != nil {
		slog.Error("database migrations error", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	questionRepo := repository.NewQuestionRepo(db)
	practiceRepo := repository.NewPracticeRepo(db)
	leaderboardRepo := repository.NewLeaderboardRepo(db)

	memCache := appCache.NewMemoryCache()
	questionBank := appCache.NewQuestionBank(questionRepo.GetAllPublished, 15*time.Minute)

	catIDs, err := questionRepo.GetActiveCategoryIDs(context.Background())
	if err == nil {
		questionBank.Warm(context.Background(), catIDs)
	}

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	practiceService := service.NewPracticeService(practiceRepo, questionRepo)
	leaderboardService := service.NewLeaderboardService(leaderboardRepo, memCache)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo, leaderboardService)
	practiceHandler := handler.NewPracticeHandler(practiceService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)
	adminHandler := handler.NewAdminHandler(questionRepo, userRepo, matchRepo)

	mux := http.NewServeMux()

	// Health check for Lambda API Gateway
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","environment":"aws_lambda"}`)
	})

	// Auth API
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/firebase", authHandler.FirebaseLogin)
	mux.HandleFunc("GET /api/v1/auth/me", middleware.Auth(cfg.JWTSecret, authHandler.Me))

	// User Profile API
	mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetProfile)
	mux.HandleFunc("GET /api/v1/users/{id}/stats", userHandler.GetStats)
	mux.HandleFunc("GET /api/v1/users/{id}/matches", userHandler.GetMatchHistory)

	// Practice Session API
	mux.HandleFunc("POST /api/v1/practice/start", middleware.Auth(cfg.JWTSecret, practiceHandler.StartSession))
	mux.HandleFunc("POST /api/v1/practice/{id}/answer", middleware.Auth(cfg.JWTSecret, practiceHandler.SubmitAnswer))
	mux.HandleFunc("POST /api/v1/practice/{id}/end", middleware.Auth(cfg.JWTSecret, practiceHandler.EndSession))
	mux.HandleFunc("GET /api/v1/practice/{id}", middleware.Auth(cfg.JWTSecret, practiceHandler.GetSession))

	// Leaderboard API
	mux.HandleFunc("GET /api/v1/leaderboard/{category}", leaderboardHandler.GetLeaderboard)

	// Admin API
	mux.HandleFunc("POST /api/v1/admin/questions", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.CreateQuestion)))
	mux.HandleFunc("PUT /api/v1/admin/questions/{id}/publish", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.PublishQuestion)))
	mux.HandleFunc("GET /api/v1/admin/stats", middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", adminHandler.GetSystemStats)))

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	adapter = httpadapter.New(h)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
