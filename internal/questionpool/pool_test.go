package questionpool

import "testing"

func TestCapBudget(t *testing.T) {
	cases := []struct {
		name                         string
		players, perUser, cats, max_ int
		want                         int
	}{
		{"41 players, 2 categories", 41, 100, 2, 5000, 2050},
		{"grows with players", 80, 100, 2, 5000, 4000},
		{"ceiling", 500, 100, 2, 5000, 5000},
		{"no players", 0, 100, 2, 5000, 0},
		{"sizing disabled", 41, 0, 2, 5000, 0},
		{"no active categories", 41, 100, 0, 5000, 0},
	}
	for _, c := range cases {
		if got := capBudget(c.players, c.perUser, c.cats, c.max_); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestGeneratedPerLevelFillsTheShare(t *testing.T) {
	// 2050 share, 120 hand-written → 1930 generated → 643 per difficulty.
	if got := generatedPerLevel(2050, 120, 3); got != 643 {
		t.Errorf("got %d, want 643", got)
	}
	// Hand-written alone already fill the share: generate nothing.
	if got := generatedPerLevel(100, 150, 3); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestTinyCeilingIsIgnored(t *testing.T) {
	// "10" once archived all but 10 questions per category.
	for _, set := range []int{0, 10, 499} {
		if got := New(nil, nil, Config{MaxPerCategory: set}).cfg.MaxPerCategory; got != 5000 {
			t.Errorf("MaxPerCategory %d: got %d, want the 5000 default", set, got)
		}
	}
	if got := New(nil, nil, Config{MaxPerCategory: 800}).cfg.MaxPerCategory; got != 800 {
		t.Errorf("a sensible ceiling should be kept, got %d", got)
	}
}
