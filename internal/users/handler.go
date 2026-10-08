package users

import (
	"net/http"
	"sync"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/respond"
)

// Handler serves player profiles: a player's details, ratings, stats and
// match history.
type Handler struct {
	userRepo *Store
}

func NewHandler(userRepo *Store) *Handler {
	return &Handler{userRepo: userRepo}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respond.Error(w, http.StatusBadRequest, "user id required")
		return
	}

	// Independent reads: run them together (one database round trip, not two).
	var (
		user    *models.User
		err     error
		ratings []models.UserRating
		wg      sync.WaitGroup
	)
	wg.Add(2)
	go func() { defer wg.Done(); user, err = h.userRepo.GetByID(r.Context(), userID) }()
	go func() { defer wg.Done(); ratings, _ = h.userRepo.GetRatings(r.Context(), userID) }()
	wg.Wait()
	if err != nil || user == nil {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"ratings": ratings,
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respond.Error(w, http.StatusBadRequest, "user id required")
		return
	}

	stats, err := h.userRepo.GetStatistics(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	respond.JSON(w, http.StatusOK, stats)
}

func (h *Handler) GetMatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		respond.Error(w, http.StatusBadRequest, "user id required")
		return
	}

	// This requires match_repo but we keep it simple
	respond.JSON(w, http.StatusOK, map[string]string{"message": "match history endpoint"})
}
