package handler

import (
	"encoding/json"
	"net/http"

	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
)

type PracticeHandler struct {
	practiceService *service.PracticeService
}

func NewPracticeHandler(practiceService *service.PracticeService) *PracticeHandler {
	return &PracticeHandler{practiceService: practiceService}
}

func (h *PracticeHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req service.StartPracticeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.practiceService.StartSession(r.Context(), userID, req)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, resp)
}

func (h *PracticeHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	var req service.SubmitPracticeAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.practiceService.SubmitAnswer(r.Context(), userID, sessionID, req)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

func (h *PracticeHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	if err := h.practiceService.EndSession(r.Context(), userID, sessionID); err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *PracticeHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	sessionID := r.PathValue("id")

	session, err := h.practiceService.GetSession(r.Context(), userID, sessionID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, session)
}
