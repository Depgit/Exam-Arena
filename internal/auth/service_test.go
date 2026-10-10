package auth

import (
	"context"
	"testing"
)

// Validation runs before any database call, so a Service with no store is
// enough to check which usernames are rejected.
func TestRegisterRejectsBadUsernames(t *testing.T) {
	s := &Service{}
	for name, want := range map[string]string{
		"chu nnu":   "username can't contain spaces",
		"chu\tnnu":  "username can't contain spaces",
		"  ab  ":    "username must be between 3 and 30 characters", // trimmed to "ab"
		"bot.evil":  "that username is reserved",
		" guest_1 ": "that username is reserved",
	} {
		_, err := s.Register(context.Background(), RegisterRequest{Username: name, Email: "a@b.co", Password: "password1"})
		if err == nil || err.Error() != want {
			t.Errorf("username %q: got %v, want %q", name, err, want)
		}
	}
}
