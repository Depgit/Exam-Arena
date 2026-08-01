package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS returns a middleware that sets CORS headers.
//
// Allowed origins are read from the CORS_ALLOWED_ORIGINS env var
// (comma-separated). In development the default is http://localhost:3000.
// If the env var is set to "*", all origins are allowed (not recommended
// for production).
func CORS(next http.Handler) http.Handler {
	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if rawOrigins == "" {
		rawOrigins = "http://localhost:3000"
	}

	allowed := map[string]bool{}
	for _, o := range strings.Split(rawOrigins, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}
	allowAll := allowed["*"]

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowAll || allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(allowed) == 0 {
			// Fallback: allow any origin (open dev mode)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
