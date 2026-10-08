package leaderboard

import (
	"net/http"
	"strconv"

	"github.com/exam-arena/internal/platform/respond"
)

type Handler struct {
	leaderboardService *Service
}

func NewHandler(leaderboardService *Service) *Handler {
	return &Handler{leaderboardService: leaderboardService}
}

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")
	if category == "" {
		respond.Error(w, http.StatusBadRequest, "category is required")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	entries, err := h.leaderboardService.GetLeaderboard(r.Context(), category, limit, offset)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSONWithMeta(w, http.StatusOK, entries, map[string]interface{}{
		"limit":    limit,
		"offset":   offset,
		"category": category,
	})
}
