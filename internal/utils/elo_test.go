package utils

import "testing"

func TestCalculateEloDrawNeverLowersHigherRated(t *testing.T) {
	cases := []struct {
		name           string
		ratingA        int
		ratingB        int
		wantAUnchanged bool
		wantBUnchanged bool
	}{
		{"A higher", 1500, 1200, true, false},
		{"B higher", 1100, 1400, false, true},
		{"equal", 1200, 1200, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newA, newB := CalculateElo(tc.ratingA, tc.ratingB, 0.5)
			if newA < tc.ratingA || newB < tc.ratingB {
				t.Fatalf("draw lowered a rating: A %d→%d, B %d→%d", tc.ratingA, newA, tc.ratingB, newB)
			}
			if tc.wantAUnchanged != (newA == tc.ratingA) {
				t.Errorf("A %d→%d, want unchanged=%v", tc.ratingA, newA, tc.wantAUnchanged)
			}
			if tc.wantBUnchanged != (newB == tc.ratingB) {
				t.Errorf("B %d→%d, want unchanged=%v", tc.ratingB, newB, tc.wantBUnchanged)
			}
		})
	}
}

func TestCalculateEloWinLossUnchanged(t *testing.T) {
	newA, newB := CalculateElo(1200, 1200, 1.0)
	if newA != 1216 || newB != 1184 {
		t.Fatalf("even win: got A=%d B=%d, want 1216/1184", newA, newB)
	}

	// Upset loss by the favourite still costs rating.
	newA, newB = CalculateElo(1500, 1200, 0.0)
	if newA >= 1500 || newB <= 1200 {
		t.Fatalf("upset: got A=%d B=%d", newA, newB)
	}
}

func TestCalculateEloFloor(t *testing.T) {
	newA, _ := CalculateElo(100, 1200, 0.0)
	if newA != 100 {
		t.Fatalf("rating fell below floor: %d", newA)
	}
}
