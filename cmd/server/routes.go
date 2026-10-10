package main

import (
	"fmt"
	"net/http"

	"github.com/exam-arena/internal/admin"
	"github.com/exam-arena/internal/auth"
	"github.com/exam-arena/internal/chat"
	"github.com/exam-arena/internal/daily"
	"github.com/exam-arena/internal/friends"
	"github.com/exam-arena/internal/leaderboard"
	"github.com/exam-arena/internal/match"
	"github.com/exam-arena/internal/matchmaking"
	"github.com/exam-arena/internal/platform/config"
	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/realtime"
	"github.com/exam-arena/internal/platform/respond"
	"github.com/exam-arena/internal/practice"
	"github.com/exam-arena/internal/questionpool"
	"github.com/exam-arena/internal/questions"
	"github.com/exam-arena/internal/users"
)

// deps is everything the routes need, built in main.go.
type deps struct {
	hub     *realtime.Hub
	queue   *matchmaking.Queue
	connect *realtime.ConnectHandler

	auth           *auth.Handler
	users          *users.Handler
	categories     *questions.SubjectHandler
	topics         *questions.TopicHandler
	flags          *questions.FlagHandler
	queueEndpoints *matchmaking.Handler
	match          *match.Handler
	friends        *friends.Handler
	practice       *practice.Handler
	daily          *daily.Handler
	leaderboard    *leaderboard.Handler
	generator      *questionpool.Handler
	chat           *chat.Handler
	adminStores    adminStores
}

type adminStores struct {
	questions *questions.Store
	users     *users.Store
	matches   *match.Store
	flags     *questions.FlagStore
}

// routes is the whole HTTP API on one page: every URL, who serves it, and
// whether you must be logged in (loggedIn) or an admin (adminOnly).
func routes(cfg *config.Config, d deps) http.Handler {
	loggedIn := func(h http.HandlerFunc) http.HandlerFunc { return middleware.Auth(cfg.JWTSecret, h) }
	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(cfg.JWTSecret, middleware.RequireRole("admin", h))
	}
	// Playing (matches, bots, friend challenges, practice, daily) needs a real
	// account; demo (guest) accounts can only look around.
	registered := auth.RegisteredOnly(d.adminStores.users)
	canPlay := func(h http.HandlerFunc) http.HandlerFunc { return loggedIn(registered(h)) }
	adminHandler := admin.NewHandler(d.adminStores.questions, d.adminStores.users, d.adminStores.matches, d.adminStores.flags, d.hub)

	mux := http.NewServeMux()

	// Health check (uptime monitors, Render)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","online":%d,"queue":%v}`, d.hub.OnlineCount(), d.queue.Stats())
	})

	// Login & accounts — internal/auth
	mux.HandleFunc("POST /api/v1/auth/register", d.auth.Register)
	mux.HandleFunc("POST /api/v1/auth/login", d.auth.Login)
	mux.HandleFunc("POST /api/v1/auth/guest", d.auth.Guest)
	mux.HandleFunc("GET /api/v1/auth/me", loggedIn(d.auth.Me))

	// Player profiles — internal/users
	mux.HandleFunc("GET /api/v1/users/search", loggedIn(d.users.SearchPlayers))
	mux.HandleFunc("GET /api/v1/users/{id}", d.users.GetProfile)
	mux.HandleFunc("GET /api/v1/users/{id}/stats", d.users.GetStats)
	mux.HandleFunc("GET /api/v1/users/{id}/matches", d.users.GetMatchHistory)

	// Matchmaking queue — internal/matchmaking
	mux.HandleFunc("POST /api/v1/matches/queue", canPlay(d.queueEndpoints.JoinQueue))
	mux.HandleFunc("DELETE /api/v1/matches/queue", loggedIn(d.queueEndpoints.LeaveQueue))
	mux.HandleFunc("GET /api/v1/matches/queue/stats", loggedIn(d.queueEndpoints.QueueStats))

	// Matches: details, private rooms, bots — internal/match
	mux.HandleFunc("GET /api/v1/matches/current", loggedIn(d.match.CurrentMatch))
	mux.HandleFunc("GET /api/v1/matches/{id}", loggedIn(d.match.GetMatch))
	mux.HandleFunc("POST /api/v1/matches/friend", canPlay(d.match.CreateFriendMatch))
	mux.HandleFunc("POST /api/v1/matches/friend/join", canPlay(d.match.JoinFriendMatch))
	mux.HandleFunc("POST /api/v1/matches/bot", canPlay(d.match.PlayBot))

	// Friends & challenges — internal/friends
	mux.HandleFunc("GET /api/v1/friends", loggedIn(d.friends.ListFriends))
	mux.HandleFunc("POST /api/v1/friends/requests", loggedIn(d.friends.SendRequest))
	mux.HandleFunc("POST /api/v1/friends/requests/{id}/accept", loggedIn(d.friends.AcceptRequest))
	mux.HandleFunc("POST /api/v1/friends/requests/{id}/decline", loggedIn(d.friends.DeclineRequest))
	mux.HandleFunc("DELETE /api/v1/friends/{userId}", loggedIn(d.friends.RemoveFriend))
	mux.HandleFunc("POST /api/v1/friends/{userId}/challenge", canPlay(d.friends.ChallengeFriend))
	mux.HandleFunc("DELETE /api/v1/friends/challenges/{matchId}", loggedIn(d.friends.CloseChallenge))

	// Categories, topics, reporting a question — internal/questions
	mux.HandleFunc("GET /api/v1/subjects", loggedIn(d.categories.GetAllSubjects))
	mux.HandleFunc("GET /api/v1/subjects/{id}/topics", loggedIn(d.topics.ListTopics))
	mux.HandleFunc("GET /api/v1/topics", loggedIn(d.topics.ListTopics))
	mux.HandleFunc("GET /api/v1/topics/{id}", loggedIn(d.topics.GetTopic))
	mux.HandleFunc("POST /api/v1/questions/{id}/flag", loggedIn(d.flags.FlagQuestion))

	// Daily challenge — internal/daily
	mux.HandleFunc("GET /api/v1/daily-challenge", loggedIn(d.daily.Overview))
	mux.HandleFunc("POST /api/v1/daily-challenge/start", canPlay(d.daily.Start))
	mux.HandleFunc("POST /api/v1/daily-challenge/submit", canPlay(d.daily.Submit))

	// Solo practice — internal/practice
	mux.HandleFunc("POST /api/v1/practice/start", canPlay(d.practice.StartSession))
	mux.HandleFunc("POST /api/v1/practice/{id}/answer", canPlay(d.practice.SubmitAnswer))
	mux.HandleFunc("POST /api/v1/practice/{id}/end", loggedIn(d.practice.EndSession))
	mux.HandleFunc("GET /api/v1/practice/{id}", loggedIn(d.practice.GetSession))

	// Global chat — internal/chat (everyone logged in reads; real accounts post)
	mux.HandleFunc("GET /api/v1/chat", loggedIn(d.chat.Recent))
	mux.HandleFunc("POST /api/v1/chat", canPlay(d.chat.Send))

	// Rankings — internal/leaderboard
	mux.HandleFunc("GET /api/v1/leaderboard/{category}", d.leaderboard.GetLeaderboard)

	// Admin console — internal/admin, internal/questions, internal/questionpool
	mux.HandleFunc("POST /api/v1/admin/topics", adminOnly(d.topics.CreateTopic))
	mux.HandleFunc("POST /api/v1/admin/questions", adminOnly(adminHandler.CreateQuestion))
	mux.HandleFunc("PUT /api/v1/admin/questions/{id}/publish", adminOnly(adminHandler.PublishQuestion))
	mux.HandleFunc("GET /api/v1/admin/stats", adminOnly(adminHandler.GetSystemStats))
	mux.HandleFunc("PUT /api/v1/admin/categories/{id}/sort-order", adminOnly(adminHandler.SetCategorySortOrder))
	mux.HandleFunc("GET /api/v1/admin/generator/preview", adminOnly(d.generator.Preview))
	mux.HandleFunc("POST /api/v1/admin/generator/rotate", adminOnly(d.generator.Rotate))
	mux.HandleFunc("GET /api/v1/admin/flags", adminOnly(d.flags.ListFlags))
	mux.HandleFunc("PUT /api/v1/admin/flags/{questionId}", adminOnly(d.flags.ReviewFlags))

	// Live connection for matches, notifications — internal/platform/realtime
	mux.HandleFunc("GET /wss", d.connect.HandleConnection)

	// Debug (unauthenticated — see internal/README.md)
	mux.HandleFunc("GET /debug/state", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]interface{}{"queue": d.queue.Stats(), "online": d.hub.OnlineCount()})
	})
	mux.HandleFunc("GET /debug/queue", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, d.queue.Snapshot())
	})

	// Every request passes through these, outermost last:
	// recover from panics → log → rate-limit → CORS → route.
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst, h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)
	return h
}
