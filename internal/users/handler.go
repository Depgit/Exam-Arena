package users

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/middleware"
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
		"user":    publicView(user),
		"ratings": ratings,
	})
}

// publicView is what anyone may see of a player. Profiles are public, so
// the email address and login details stay out; a player sees their own
// through GET /auth/me.
func publicView(u *models.User) *models.User {
	p := *u
	p.Email = ""
	p.PasswordHash = ""
	p.EmailVerifiedAt = nil
	p.LastLoginAt = nil
	return &p
}

// UpdateMe changes the logged-in player's own details. Only the display
// name can change: the username is permanent (friends and links use it).
//
//	PATCH /api/v1/users/me  { "display_name": "Priya S" }
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DisplayName *string `json:"display_name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil || req.DisplayName == nil {
		respond.Error(w, http.StatusBadRequest, "send a display_name")
		return
	}
	userID := middleware.GetUserID(r)
	me, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil || me == nil {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if me.IsGuest {
		respond.Error(w, http.StatusForbidden, "demo accounts can't change their name — create a free account")
		return
	}
	name, err := CleanDisplayName(*req.DisplayName)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if name == "" {
		name = me.Username // cleared: fall back to the username
	}
	// Don't let a name pose as someone else's username.
	if !strings.EqualFold(name, me.Username) {
		if other, _ := h.userRepo.GetByUsername(r.Context(), name); other != nil {
			respond.Error(w, http.StatusBadRequest, "that name is another player's username — pick something else")
			return
		}
	}
	if err := h.userRepo.SetDisplayName(r.Context(), userID, name); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to save")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"display_name": name})
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

// SearchPlayers suggests players as you type a name to add as a friend.
//
//	GET /api/v1/users/search?q=chu   (at least 2 characters; up to 8 results)
func (h *Handler) SearchPlayers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(q)) < 2 {
		respond.JSON(w, http.StatusOK, []PlayerMatch{})
		return
	}
	if len([]rune(q)) > 30 {
		q = string([]rune(q)[:30])
	}
	matches, err := h.userRepo.SearchPlayers(r.Context(), q, middleware.GetUserID(r), 8)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "search failed")
		return
	}
	respond.JSON(w, http.StatusOK, matches)
}
