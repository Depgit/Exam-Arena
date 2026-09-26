package handler

import (
	"net/http"
	"strconv"

	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
)

type LeaderboardHandler struct {
	leaderboardService *service.LeaderboardService
}

func NewLeaderboardHandler(leaderboardService *service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")
	if category == "" {
		utils.JSONError(w, http.StatusBadRequest, "category is required")
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
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONWithMeta(w, http.StatusOK, entries, map[string]interface{}{
		"limit":    limit,
		"offset":   offset,
		"category": category,
	})
}
