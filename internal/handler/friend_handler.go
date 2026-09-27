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

type FriendHandler struct {
	friendService *service.FriendService
}

func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

// writeFriendError maps friend-service errors to HTTP statuses.
func writeFriendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrRequestNotFound),
		errors.Is(err, service.ErrNotFriends),
		errors.Is(err, service.ErrCategoryNotFound),
		errors.Is(err, service.ErrChallengeNotFound):
		utils.JSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyFriends),
		errors.Is(err, service.ErrRequestPending),
		errors.Is(err, service.ErrFriendOffline),
		errors.Is(err, service.ErrNotEnoughQuestions):
		utils.JSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrCannotFriendSelf):
		utils.JSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrRequestUnavailable):
		utils.JSONError(w, http.StatusForbidden, err.Error())
	default:
		slog.Error("friends request failed", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ListFriends godoc
// GET /api/v1/friends
// Response: { "friends": [...], "incoming": [...], "outgoing": [...] }
func (h *FriendHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	list, err := h.friendService.List(r.Context(), middleware.GetUserID(r))
	if err != nil {
		writeFriendError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, list)
}

// SendRequest godoc
// POST /api/v1/friends/requests
// Body: { "username": "alice" }
//
// If alice already sent you a pending request, it is accepted instead.
func (h *FriendHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" {
		utils.JSONError(w, http.StatusBadRequest, "username is required")
		return
	}

	f, accepted, err := h.friendService.SendRequest(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), req.Username)
	if err != nil {
		writeFriendError(w, err)
		return
	}

	if accepted {
		utils.JSON(w, http.StatusOK, map[string]interface{}{
			"friendship": f,
			"message":    "they had already sent you a request — you are now friends",
		})
		return
	}
	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"friendship": f,
		"message":    "friend request sent",
	})
}

// AcceptRequest godoc
// POST /api/v1/friends/requests/{id}/accept
func (h *FriendHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, true)
}

// DeclineRequest godoc
// POST /api/v1/friends/requests/{id}/decline
func (h *FriendHandler) DeclineRequest(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, false)
}

func (h *FriendHandler) respond(w http.ResponseWriter, r *http.Request, accept bool) {
	f, err := h.friendService.Respond(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), r.PathValue("id"), accept)
	if err != nil {
		writeFriendError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"friendship": f})
}

// RemoveFriend godoc
// DELETE /api/v1/friends/{userId}
//
// Unfriends, withdraws an outgoing request, or dismisses an incoming one.
func (h *FriendHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	if err := h.friendService.Remove(r.Context(), middleware.GetUserID(r), r.PathValue("userId")); err != nil {
		writeFriendError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ChallengeFriend godoc
// POST /api/v1/friends/{userId}/challenge
// Body: { "exam_category_id": "<uuid>" }
//
// Opens a friend match room and sends the friend a "friend_challenge"
// WebSocket message. The friend accepts by joining the room code.
func (h *FriendHandler) ChallengeFriend(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.friendService.Challenge(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), r.PathValue("userId"), req.ExamCategoryID)
	if err != nil {
		writeFriendError(w, err)
		return
	}
	utils.JSON(w, http.StatusCreated, res)
}

// CloseChallenge godoc
// DELETE /api/v1/friends/challenges/{matchId}
//
// The invited friend declines, or the challenger cancels, an open challenge.
func (h *FriendHandler) CloseChallenge(w http.ResponseWriter, r *http.Request) {
	if err := h.friendService.DeclineOrCancelChallenge(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), r.PathValue("matchId")); err != nil {
		writeFriendError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"status": "closed"})
}
