package chat

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

// Handler serves the global chat over HTTP. New messages are pushed live to
// everyone as "chat_message" WebSocket events.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Recent returns the latest messages, oldest first.
//
//	GET /api/v1/chat
func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, h.service.Recent())
}

// Send posts a message to the global chat.
//
//	POST /api/v1/chat  {"body": "hello"}
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	m, err := h.service.Send(r.Context(), middleware.GetUserID(r), middleware.GetUsername(r), req.Body)
	switch {
	case errors.Is(err, ErrTooFast):
		respond.Error(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, ErrEmpty), errors.Is(err, ErrTooLong), errors.Is(err, ErrRepeated):
		respond.Error(w, http.StatusBadRequest, err.Error())
	case err != nil:
		slog.Error("chat send failed", "error", err)
		respond.Error(w, http.StatusInternalServerError, "message not sent, please try again")
	default:
		respond.JSON(w, http.StatusCreated, m)
	}
}
