package users

import (
	"strings"
	"testing"
	"time"

	"github.com/exam-arena/internal/models"
)

func TestCleanDisplayName(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"  Priya   Sharma ":     {"Priya Sharma", true},
		"Line\nbreak\ttab":      {"Line break tab", true},
		"":                      {"", true},
		"   ":                   {"", true},
		"A":                     {"", false},
		"राहुल":                 {"राहुल", true},
		strings.Repeat("x", 30): {strings.Repeat("x", 30), true},
		strings.Repeat("x", 31): {"", false},
	}
	for in, c := range cases {
		got, err := CleanDisplayName(in)
		if (err == nil) != c.ok || (c.ok && got != c.want) {
			t.Errorf("%q: got %q, err %v; want %q ok=%v", in, got, err, c.want, c.ok)
		}
	}
}

func TestPublicViewHidesPrivateFields(t *testing.T) {
	now := time.Now()
	u := &models.User{Username: "priya", Email: "priya@gmail.com", PasswordHash: "hash", EmailVerifiedAt: &now, LastLoginAt: &now}
	p := publicView(u)
	if p.Email != "" || p.PasswordHash != "" || p.EmailVerifiedAt != nil || p.LastLoginAt != nil {
		t.Fatalf("private fields leaked: %+v", p)
	}
	if p.Username != "priya" || u.Email != "priya@gmail.com" {
		t.Fatal("publicView must keep public fields and not modify the original")
	}
}
