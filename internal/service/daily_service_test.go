package service

import (
	"testing"
	"time"
)

func TestChallengeDateRollsOverAtISTMidnight(t *testing.T) {
	// 18:29 UTC = 23:59 IST (still the 27th); 18:30 UTC = 00:00 IST on the 28th.
	before := time.Date(2026, 9, 27, 18, 29, 0, 0, time.UTC)
	after := time.Date(2026, 9, 27, 18, 30, 0, 0, time.UTC)

	if got := challengeDate(before).Format("2006-01-02"); got != "2026-09-27" {
		t.Errorf("23:59 IST: got %s, want 2026-09-27", got)
	}
	if got := challengeDate(after).Format("2006-01-02"); got != "2026-09-28" {
		t.Errorf("00:00 IST: got %s, want 2026-09-28", got)
	}
	if got := nextReset(before); !got.Equal(after) {
		t.Errorf("next reset: got %s, want %s", got, after)
	}
}
