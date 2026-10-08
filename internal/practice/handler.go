package practice

import (
	"encoding/json"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

type Handler struct {
	practiceService *Service
}

func NewHandler(practiceService *Service) *Handler {
	return &Handler{practiceService: practiceService}
}

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req StartPracticeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.practiceService.StartSession(r.Context(), userID, req)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	respond.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	var req SubmitPracticeAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.practiceService.SubmitAnswer(r.Context(), userID, sessionID, req)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, resp)
}

func (h *Handler) EndSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	if err := h.practiceService.EndSession(r.Context(), userID, sessionID); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	session, err := h.practiceService.GetSession(r.Context(), userID, sessionID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, session)
}
