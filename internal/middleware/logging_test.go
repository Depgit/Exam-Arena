package middleware

import (
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestRequestLogLevel(t *testing.T) {
	cases := []struct {
		method, path string
		status       int
		want         slog.Level
	}{
		{"GET", "/api/v1/subjects", 200, slog.LevelInfo},
		{"GET", "/health", 200, slog.LevelDebug},
		{"OPTIONS", "/api/v1/friends", 204, slog.LevelDebug},
		{"POST", "/api/v1/auth/login", 401, slog.LevelWarn},
		{"POST", "/api/v1/daily-challenge/submit", 500, slog.LevelError},
		// A failing health check must not be hidden at debug.
		{"GET", "/health", 503, slog.LevelError},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.path, nil)
		if got := requestLogLevel(r, c.status); got != c.want {
			t.Errorf("%s %s %d: got %v, want %v", c.method, c.path, c.status, got, c.want)
		}
	}
}
