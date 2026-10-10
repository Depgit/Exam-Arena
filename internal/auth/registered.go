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

// VerifyToPlay is the message an unverified player gets from any play endpoint.
const VerifyToPlay = "Verify your email to play and chat — enter the code we emailed you."

// playChecker is the one store call RegisteredOnly needs.
type playChecker interface {
	PlayStatus(ctx context.Context, userID string) (guest, verified bool, err error)
}

// RegisteredOnly returns a wrapper that lets only real (non-guest)
// accounts through — and, when requireVerified is set, only those that
// have verified their email. Use it inside middleware.Auth, on every
// endpoint that starts or plays a match, practice session, daily
// challenge, or posts to chat.
//
// It checks the database, not the login token, so tokens issued before a
// rule existed are covered too.
func RegisteredOnly(users playChecker, requireVerified bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			guest, verified, err := users.PlayStatus(r.Context(), middleware.GetUserID(r))
			if err != nil {
				slog.Error("play check failed", "error", err)
				respond.Error(w, http.StatusInternalServerError, "please try again")
				return
			}
			if guest {
				respond.Error(w, http.StatusForbidden, GuestCannotPlay)
				return
			}
			if requireVerified && !verified {
				respond.Error(w, http.StatusForbidden, VerifyToPlay)
				return
			}
			next(w, r)
		}
	}
}
