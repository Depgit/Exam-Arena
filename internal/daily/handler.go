package daily

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

type Handler struct {
	dailyService *Service
}

func NewHandler(dailyService *Service) *Handler {
	return &Handler{dailyService: dailyService}
}

func writeDailyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoDailyChallenge):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDailyNotStarted):
		respond.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrDailyAlreadyPlayed):
		respond.Error(w, http.StatusConflict, err.Error())
	default:
		slog.Error("daily challenge request failed", "error", err)
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// Overview godoc
// GET /api/v1/daily-challenge
func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	ov, err := h.dailyService.Overview(r.Context(), middleware.GetUserID(r))
	if err != nil {
		writeDailyError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, ov)
}

// Start godoc
// POST /api/v1/daily-challenge/start
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	start, err := h.dailyService.Start(r.Context(), middleware.GetUserID(r))
	if err != nil {
		writeDailyError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, start)
}

// Submit godoc
// POST /api/v1/daily-challenge/submit
// Body: { "answers": [ { "question_id": "...", "option_id": "..." } ] }
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answers []DailyAnswer `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	res, err := h.dailyService.Submit(r.Context(), middleware.GetUserID(r), req.Answers)
	if err != nil {
		writeDailyError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, res)
}
