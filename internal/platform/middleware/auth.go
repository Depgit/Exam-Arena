package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/exam-arena/internal/platform/respond"
	"github.com/exam-arena/internal/platform/tokens"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UsernameKey contextKey = "username"
	RoleKey     contextKey = "role"
)

func Auth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respond.Error(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respond.Error(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		claims, err := tokens.Validate(parts[1], secret)
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value(RoleKey).(string)
		if !ok || userRole != role {
			respond.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func GetUserID(r *http.Request) string {
	if v, ok := r.Context().Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func GetUsername(r *http.Request) string {
	if v, ok := r.Context().Value(UsernameKey).(string); ok {
		return v
	}
	return ""
}
