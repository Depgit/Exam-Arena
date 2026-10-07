package service

import "testing"

func TestCapBudget(t *testing.T) {
	cases := []struct {
		name                                 string
		perLevel, players, perUser, cats, lv int
		wantLevel, wantCategoryCap           int
	}{
		// Today: 13 players → 1300 total → 650 per category; the generated
		// pool (200/level = 600) fits, the oldest 50+ others get archived.
		{"current data", 200, 13, 100, 2, 3, 200, 650},
		{"few players: generated pool shrinks to fit", 200, 3, 100, 2, 3, 50, 150},
		{"no players", 200, 0, 100, 2, 3, 0, 0},
		{"cap disabled", 200, 13, 0, 2, 3, 200, 0},
		{"no active categories", 200, 13, 100, 0, 3, 200, 0},
	}
	for _, c := range cases {
		level, cap := capBudget(c.perLevel, c.players, c.perUser, c.cats, c.lv)
		if level != c.wantLevel || cap != c.wantCategoryCap {
			t.Errorf("%s: got level %d cap %d, want %d / %d", c.name, level, cap, c.wantLevel, c.wantCategoryCap)
		}
	}
}
