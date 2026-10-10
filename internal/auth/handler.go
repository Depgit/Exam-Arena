package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

type Handler struct {
	authService *Service
	guests      *guestLimiter
}

func NewHandler(authService *Service) *Handler {
	return &Handler{authService: authService, guests: newGuestLimiter(guestsPerHour)}
}

// guestsPerHour caps demo accounts per visitor IP, so the button can't be
// scripted to fill the users table.
const guestsPerHour = 10

// Guest signs the visitor in to a brand-new demo account.
//
//	POST /api/v1/auth/guest
func (h *Handler) Guest(w http.ResponseWriter, r *http.Request) {
	if !h.guests.allow(clientIP(r)) {
		respond.Error(w, http.StatusTooManyRequests, "too many demo accounts from this network — try again later")
		return
	}
	resp, err := h.authService.Guest(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not start a demo session")
		return
	}
	respond.JSON(w, http.StatusCreated, resp)
}

// clientIP is the visitor's address. Behind a proxy (Render) the first
// X-Forwarded-For entry is the client; later entries are proxies.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i > 0 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}

// guestLimiter allows at most max guest accounts per IP per rolling hour.
type guestLimiter struct {
	mu   sync.Mutex
	max  int
	seen map[string][]time.Time
}

func newGuestLimiter(max int) *guestLimiter {
	return &guestLimiter{max: max, seen: map[string][]time.Time{}}
}

func (l *guestLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-time.Hour)
	recent := l.seen[ip][:0]
	for _, t := range l.seen[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= l.max {
		l.seen[ip] = recent
		return false
	}
	l.seen[ip] = append(recent, time.Now())
	// Drop idle IPs now and then so the map can't grow without bound.
	if len(l.seen) > 10000 {
		for k, ts := range l.seen {
			if len(ts) == 0 || ts[len(ts)-1].Before(cutoff) {
				delete(l.seen, k)
			}
		}
	}
	return true
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, resp)
}

// Google signs in with a "Sign in with Google" ID token.
//
//	POST /api/v1/auth/google  { "credential": "<id token>", "username": "optional" }
//
// A first-time player gets { needs_username, suggested_username } back and
// sends the same credential again with the username they picked.
func (h *Handler) Google(w http.ResponseWriter, r *http.Request) {
	var req GoogleRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil || req.Credential == "" {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	res, err := h.authService.Google(r.Context(), req)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

// SendVerificationCode emails a fresh 6-digit code.
//
//	POST /api/v1/auth/verify-email/send  { "email": "optional new address" }
func (h *Handler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}
	err := h.authService.SendVerificationCode(r.Context(), middleware.GetUserID(r), req.Email)
	var tooSoon ErrTooSoon
	switch {
	case errors.As(err, &tooSoon):
		respond.Error(w, http.StatusTooManyRequests, err.Error())
	case errors.Is(err, ErrAlreadyVerified), errors.Is(err, ErrEmailTaken):
		respond.Error(w, http.StatusConflict, err.Error())
	case err != nil && strings.Contains(err.Error(), "send email"):
		slog.Error("verification email failed", "error", err)
		respond.Error(w, http.StatusBadGateway, "couldn't send the email right now — try again in a minute")
	case err != nil:
		respond.Error(w, http.StatusBadRequest, err.Error())
	default:
		respond.JSON(w, http.StatusOK, map[string]string{"status": "sent"})
	}
}

// VerifyEmail checks the code and returns the updated user.
//
//	POST /api/v1/auth/verify-email  { "code": "123456" }
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.authService.VerifyEmail(r.Context(), middleware.GetUserID(r), req.Code)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	respond.JSON(w, http.StatusOK, user)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil || user == nil {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}

	respond.JSON(w, http.StatusOK, user)
}
