package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exam-arena/internal/platform/middleware"
)

// fakeUsers: user id → "real" (verified), "unverified" or "demo".
type fakeUsers map[string]string

func (f fakeUsers) PlayStatus(_ context.Context, id string) (guest, verified bool, err error) {
	kind, ok := f[id]
	if !ok || kind == "demo" {
		return true, false, nil
	}
	return false, kind == "real", nil
}

func TestRegisteredOnlyBlocksGuests(t *testing.T) {
	reached := false
	h := RegisteredOnly(fakeUsers{"real": "real", "demo": "demo"}, false)(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	})
	// No user id in the request context (as for a missing/unknown user):
	// treated as a guest and refused.
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("POST", "/api/v1/matches/bot", nil))
	if rec.Code != http.StatusForbidden || reached {
		t.Fatalf("unknown user: got %d, reached=%v; want 403 and not reached", rec.Code, reached)
	}
}

func TestRegisteredOnlyLetsRealAccountsThrough(t *testing.T) {
	for user, wantCode := range map[string]int{"real": http.StatusOK, "demo": http.StatusForbidden} {
		reached := false
		h := RegisteredOnly(fakeUsers{"real": "real", "demo": "demo"}, false)(func(w http.ResponseWriter, r *http.Request) {
			reached = true
			w.WriteHeader(http.StatusOK)
		})
		req := httptest.NewRequest("POST", "/api/v1/matches/bot", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, user))
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != wantCode || reached != (wantCode == http.StatusOK) {
			t.Errorf("%s: got %d (reached=%v), want %d", user, rec.Code, reached, wantCode)
		}
	}
}

func TestRegisteredOnlyNeedsVerifiedEmailWhenRequired(t *testing.T) {
	users := fakeUsers{"real": "real", "new": "unverified", "demo": "demo"}
	cases := []struct {
		user     string
		required bool
		want     int
	}{
		{"new", true, http.StatusForbidden},
		{"new", false, http.StatusOK},
		{"real", true, http.StatusOK},
		{"demo", false, http.StatusForbidden},
	}
	for _, c := range cases {
		h := RegisteredOnly(users, c.required)(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		req := httptest.NewRequest("POST", "/api/v1/chat", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, c.user))
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s (required=%v): got %d, want %d", c.user, c.required, rec.Code, c.want)
		}
	}
}
