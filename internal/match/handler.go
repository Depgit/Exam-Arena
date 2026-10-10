package match

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
	"github.com/exam-arena/internal/questions"
)

type Handler struct {
	matchService *Service
}

func NewHandler(matchService *Service) *Handler {
	return &Handler{matchService: matchService}
}

// ── Match Lookup ──────────────────────────────────────────────────────────

// GetMatch godoc
// GET /api/v1/matches/{id}
// Returns full match details including live scores if the match is in progress.
func (h *Handler) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchID := r.PathValue("id")
	if matchID == "" {
		respond.Error(w, http.StatusBadRequest, "match id is required")
		return
	}

	details, err := h.matchService.GetMatchDetails(r.Context(), matchID, middleware.GetUserID(r))
	if err != nil {
		respond.Error(w, http.StatusNotFound, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, details)
}

// CurrentMatch godoc
// GET /api/v1/matches/current
// Response: { "match_id": "<uuid>" } while the player is in a running match,
// otherwise { "match_id": null }. The app uses it to send a player who closed
// the tab back into their match.
func (h *Handler) CurrentMatch(w http.ResponseWriter, r *http.Request) {
	id, err := h.matchService.CurrentMatch(r.Context(), middleware.GetUserID(r))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to look up current match")
		return
	}
	var matchID *string
	if id != "" {
		matchID = &id
	}
	respond.JSON(w, http.StatusOK, map[string]*string{"match_id": matchID})
}

// LeaveMatch godoc
// POST /api/v1/matches/{id}/leave
// Gives up a live match: the other player wins, and in a rated match the
// leaver takes the loss. Everyone gets the usual match_end message.
func (h *Handler) LeaveMatch(w http.ResponseWriter, r *http.Request) {
	err := h.matchService.LeaveMatch(r.Context(), r.PathValue("id"), middleware.GetUserID(r))
	switch {
	case errors.Is(err, ErrMatchOver):
		respond.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotAPlayer):
		respond.Error(w, http.StatusForbidden, err.Error())
	case err != nil:
		respond.Error(w, http.StatusInternalServerError, "failed to leave match")
	default:
		respond.JSON(w, http.StatusOK, map[string]string{"status": "left"})
	}
}

// NotInMatch guards the routes that start a new match: a player already
// in a running match gets 409 instead of a second, overlapping game.
func (h *Handler) NotInMatch(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, err := h.matchService.CurrentMatch(r.Context(), middleware.GetUserID(r))
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "failed to check for a running match")
			return
		}
		if current != "" {
			respond.Error(w, http.StatusConflict, ErrInAMatch.Error())
			return
		}
		next(w, r)
	}
}

// ── Friend Matches ────────────────────────────────────────────────────────

// CreateFriendMatch godoc
// POST /api/v1/matches/friend
// Body: { "exam_category_id": "<uuid>" }
// Response: { "match_id": "...", "room_code": "ABC123" }
//
// The creator receives a 6-character room code which they share with the
// friend. The match does not start until the friend calls JoinFriendMatch.
func (h *Handler) CreateFriendMatch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req struct {
		ExamCategoryID string `json:"exam_category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ExamCategoryID == "" {
		respond.Error(w, http.StatusBadRequest, "exam_category_id is required")
		return
	}

	match, roomCode, err := h.matchService.CreateFriendMatch(r.Context(), userID, username, req.ExamCategoryID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]interface{}{
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
func (h *Handler) PlayBot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExamCategoryID string `json:"exam_category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExamCategoryID == "" {
		respond.Error(w, http.StatusBadRequest, "exam_category_id is required")
		return
	}
	matchID, err := h.matchService.StartBotMatch(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), req.ExamCategoryID)
	if errors.Is(err, questions.ErrNotEnoughQuestions) {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		slog.Error("bot match failed", "error", err)
		respond.Error(w, http.StatusInternalServerError, "could not start a bot match")
		return
	}
	respond.JSON(w, http.StatusCreated, map[string]interface{}{"match_id": matchID})
}

func (h *Handler) JoinFriendMatch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req struct {
		RoomCode string `json:"room_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RoomCode == "" {
		respond.Error(w, http.StatusBadRequest, "room_code is required")
		return
	}

	match, err := h.matchService.JoinFriendMatch(r.Context(), userID, username, req.RoomCode)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"match_id": match.ID,
		"status":   match.Status,
		"message":  "match is starting — listen on your WebSocket connection",
	})
}
