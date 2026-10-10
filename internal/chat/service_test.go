package chat

import (
	"testing"
	"time"
)

func TestClean(t *testing.T) {
	cases := map[string]struct {
		want string
		err  error
	}{
		"  hello  ":          {"hello", nil},
		"hi\n\nthere\tfolks": {"hi there folks", nil},
		"   ":                {"", ErrEmpty},
		"\x00\x07":           {"", ErrEmpty},
	}
	for in, c := range cases {
		got, err := Clean(in)
		if got != c.want || err != c.err {
			t.Errorf("Clean(%q) = %q, %v; want %q, %v", in, got, err, c.want, c.err)
		}
	}
	long := make([]rune, MaxLength+1)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := Clean(string(long)); err != ErrTooLong {
		t.Errorf("301 characters: got %v, want ErrTooLong", err)
	}
	if _, err := Clean(string(long[:MaxLength])); err != nil {
		t.Errorf("300 characters should be allowed, got %v", err)
	}
}

func TestFloodControl(t *testing.T) {
	s := &Service{senders: map[string]*senderLog{}}
	now := time.Now()
	for i := 0; i < burstMax; i++ {
		if err := s.allow("u", string(rune('a'+i)), now.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatalf("message %d should pass: %v", i+1, err)
		}
	}
	if err := s.allow("u", "zz", now.Add(5*time.Second)); err != ErrTooFast {
		t.Fatalf("6th message within 10s: got %v, want ErrTooFast", err)
	}
	if err := s.allow("u", "later", now.Add(15*time.Second)); err != nil {
		t.Fatalf("after the window passes: %v", err)
	}
	if err := s.allow("u", "LATER", now.Add(16*time.Second)); err != ErrRepeated {
		t.Fatalf("same text again: got %v, want ErrRepeated", err)
	}
	if err := s.allow("other", "later", now.Add(16*time.Second)); err != nil {
		t.Fatalf("another player is limited separately: %v", err)
	}
}
