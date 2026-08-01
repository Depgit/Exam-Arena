package handler

import (
	"encoding/json"
	"net/http"

	"github.com/exam-arena/internal/middleware"
	"github.com/exam-arena/internal/service"
	"github.com/exam-arena/internal/utils"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil || user == nil {
		utils.JSONError(w, http.StatusNotFound, "user not found")
		return
	}

	utils.JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) FirebaseLogin(w http.ResponseWriter, r *http.Request) {
	var req service.FirebaseLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.LoginWithFirebase(r.Context(), req)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

