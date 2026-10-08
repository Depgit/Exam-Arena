package matchmaking

import (
	"encoding/json"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

// Handler serves the matchmaking queue endpoints: join, leave, and
// how many players are waiting.
type Handler struct {
	matchmakingService *Service
}

func NewHandler(matchmakingService *Service) *Handler {
	return &Handler{matchmakingService: matchmakingService}
}

// JoinQueue godoc
// POST /api/v1/matches/queue
// Body: { "exam_category_id": "<uuid>", "match_type": "ranked" }
func (h *Handler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	username := middleware.GetUsername(r)

	var req JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.matchmakingService.JoinQueue(r.Context(), userID, username, req); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "queued",
		"message": "waiting for an opponent",
	})
}

// LeaveQueue godoc
// DELETE /api/v1/matches/queue
func (h *Handler) LeaveQueue(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	if err := h.matchmakingService.LeaveQueue(r.Context(), userID); err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "removed from queue"})
}

// QueueStats godoc
// GET /api/v1/matches/queue/stats   (requires auth — shows live queue depth)
func (h *Handler) QueueStats(w http.ResponseWriter, r *http.Request) {
	stats := h.matchmakingService.QueueStats()
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"pools": stats,
	})
}
