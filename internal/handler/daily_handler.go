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

type DailyHandler struct {
	dailyService *service.DailyService
}

func NewDailyHandler(dailyService *service.DailyService) *DailyHandler {
	return &DailyHandler{dailyService: dailyService}
}

func writeDailyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNoDailyChallenge):
		utils.JSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrDailyNotStarted):
		utils.JSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrDailyAlreadyPlayed):
		utils.JSONError(w, http.StatusConflict, err.Error())
	default:
		slog.Error("daily challenge request failed", "error", err)
		utils.JSONError(w, http.StatusInternalServerError, "internal server error")
	}
}

// Overview godoc
// GET /api/v1/daily-challenge
func (h *DailyHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ov, err := h.dailyService.Overview(r.Context(), middleware.GetUserID(r))
	if err != nil {
		writeDailyError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, ov)
}

// Start godoc
// POST /api/v1/daily-challenge/start
func (h *DailyHandler) Start(w http.ResponseWriter, r *http.Request) {
	start, err := h.dailyService.Start(r.Context(), middleware.GetUserID(r))
	if err != nil {
		writeDailyError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, start)
}

// Submit godoc
// POST /api/v1/daily-challenge/submit
// Body: { "answers": [ { "question_id": "...", "option_id": "..." } ] }
func (h *DailyHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answers []service.DailyAnswer `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	res, err := h.dailyService.Submit(r.Context(), middleware.GetUserID(r), req.Answers)
	if err != nil {
		writeDailyError(w, err)
		return
	}
	utils.JSON(w, http.StatusOK, res)
}
