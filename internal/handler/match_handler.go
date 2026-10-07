package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
)

type MatchHandler struct {
	matchService       *service.MatchService
	matchmakingService *service.MatchmakingService
}

func NewMatchHandler(
	matchService *service.MatchService,
	matchmakingService *service.MatchmakingService,
) *MatchHandler {
	return &MatchHandler{
		matchService:       matchService,
		matchmakingService: matchmakingService,
	}
}

// ── Matchmaking ───────────────────────────────────────────────────────────

// JoinQueue godoc
// POST /api/v1/matches/queue
// Body: { "exam_category_id": "<uuid>", "match_type": "ranked" }
func (h *MatchHandler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req service.JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.matchmakingService.JoinQueue(r.Context(), userID, username, req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "queued",
		"message": "waiting for an opponent",
	})
}

// LeaveQueue godoc
// DELETE /api/v1/matches/queue
func (h *MatchHandler) LeaveQueue(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	if err := h.matchmakingService.LeaveQueue(r.Context(), userID); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "removed from queue"})
}

// QueueStats godoc
// GET /api/v1/matches/queue/stats   (requires auth — shows live queue depth)
func (h *MatchHandler) QueueStats(w http.ResponseWriter, r *http.Request) {
	stats := h.matchmakingService.QueueStats()
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"pools": stats,
	})
}

// ── Match Lookup ──────────────────────────────────────────────────────────

// GetMatch godoc
// GET /api/v1/matches/{id}
// Returns full match details including live scores if the match is in progress.
func (h *MatchHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		utils.JSONError(w, http.StatusBadRequest, "match id is required")
		return
	}

	details, err := h.matchService.GetMatchDetails(r.Context(), matchID, middleware.GetUserID(r))
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, details)
}

// ── Friend Matches ────────────────────────────────────────────────────────

// CreateFriendMatch godoc
// POST /api/v1/matches/friend
// Body: { "exam_category_id": "<uuid>" }
// Response: { "match_id": "...", "room_code": "ABC123" }
//
// The creator receives a 6-character room code which they share with the
// friend. The match does not start until the friend calls JoinFriendMatch.
func (h *MatchHandler) CreateFriendMatch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req struct {
		ExamCategoryID string `json:"exam_category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ExamCategoryID == "" {
		utils.JSONError(w, http.StatusBadRequest, "exam_category_id is required")
		return
	}

	match, roomCode, err := h.matchService.CreateFriendMatch(r.Context(), userID, username, req.ExamCategoryID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"match_id":  match.ID,
		"room_code": roomCode,
		"status":    match.Status,
		"message":   "share the room code with your friend",
	})
}

// JoinFriendMatch godoc
// POST /api/v1/matches/friend/join
// Body: { "room_code": "ABC123" }
// Response: { "match_id": "...", "status": "in_progress" }
//
// When this player joins, the room is full and the match starts.
// Both players receive a "match_start" WebSocket message.
// PlayBot starts an unrated match against a bot, for when nobody else is
// searching. The match itself arrives over WebSocket as match_start.
//
//	POST /api/v1/matches/bot  {"exam_category_id": "..."}
func (h *MatchHandler) PlayBot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExamCategoryID string `json:"exam_category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExamCategoryID == "" {
		utils.JSONError(w, http.StatusBadRequest, "exam_category_id is required")
		return
	}
	matchID, err := h.matchService.StartBotMatch(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), req.ExamCategoryID)
	if errors.Is(err, service.ErrNotEnoughQuestions) {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		slog.Error("bot match failed", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "could not start a bot match")
		return
	}
	utils.JSON(w, http.StatusCreated, map[string]interface{}{"match_id": matchID})
}

func (h *MatchHandler) JoinFriendMatch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req struct {
		RoomCode string `json:"room_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RoomCode == "" {
		utils.JSONError(w, http.StatusBadRequest, "room_code is required")
		return
	}

	match, err := h.matchService.JoinFriendMatch(r.Context(), userID, username, req.RoomCode)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"match_id": match.ID,
		"status":   match.Status,
		"message":  "match is starting — listen on your WebSocket connection",
	})
}
