package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/exam-arena/internal/platform/middleware"
)

type fakeUsers map[string]bool // user id → is guest

func (f fakeUsers) IsGuest(_ context.Context, id string) (bool, error) {
	guest, ok := f[id]
	return guest || !ok, nil
}

func TestRegisteredOnlyBlocksGuests(t *testing.T) {
	reached := false
	h := RegisteredOnly(fakeUsers{"real": false, "demo": true})(func(w http.ResponseWriter, r *http.Request) {
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
		h := RegisteredOnly(fakeUsers{"real": false, "demo": true})(func(w http.ResponseWriter, r *http.Request) {
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
