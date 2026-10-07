package handler

import (
	"net/http/httptest"
	"testing"
)

func TestGuestLimiter(t *testing.T) {
	l := newGuestLimiter(3)
	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4th guest in an hour from one IP should be refused")
	}
	if !l.allow("5.6.7.8") {
		t.Fatal("another IP has its own allowance")
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/v1/auth/guest", nil)
	r.RemoteAddr = "10.0.0.9:51234"
	if got := clientIP(r); got != "10.0.0.9" {
		t.Errorf("direct: got %q", got)
	}
	r.Header.Set("X-Forwarded-For", "203.0.113.7, 10.1.2.3")
	if got := clientIP(r); got != "203.0.113.7" {
		t.Errorf("behind proxy: got %q, want the first (client) address", got)
	}
}
