package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/exam-arena/internal/platform/middleware"
	"github.com/exam-arena/internal/platform/respond"
)

// GuestCannotPlay is the message a demo account gets from any play endpoint.
const GuestCannotPlay = "Demo accounts can look around, but playing and chatting need a free account."

// guestChecker is the one store call RegisteredOnly needs.
type guestChecker interface {
	IsGuest(ctx context.Context, userID string) (bool, error)
}

// RegisteredOnly returns a wrapper that lets only real (non-guest) accounts
// through. Use it inside middleware.Auth, on every endpoint that starts or
// plays a match, practice session or daily challenge.
//
// It checks the database, not the login token, so guest tokens issued
// before this rule existed are blocked too.
func RegisteredOnly(users guestChecker) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			guest, err := users.IsGuest(r.Context(), middleware.GetUserID(r))
			if err != nil {
				slog.Error("guest check failed", "error", err)
				respond.Error(w, http.StatusInternalServerError, "please try again")
				return
			}
			if guest {
				respond.Error(w, http.StatusForbidden, GuestCannotPlay)
				return
			}
			next(w, r)
		}
	}
}
