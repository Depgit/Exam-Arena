package questions

import (
	"fmt"
	"testing"

	"github.com/exam-arena/internal/models"
)

func sampleQuestion() models.Question {
	q := models.Question{ID: "q-1"}
	for i, text := range []string{"A text", "B text", "C text", "D text"} {
		q.Options = append(q.Options, models.QuestionOption{
			ID: fmt.Sprintf("opt-%d", i), OptionText: text, IsCorrect: i == 1, OrderIndex: i + 1,
		})
	}
	return q
}

func TestPlayerOptionsKeepsEveryOptionAndRenumbers(t *testing.T) {
	opts := PlayerOptions(sampleQuestion(), "match:m1:u1")
	seen := map[string]bool{}
	for i, o := range opts {
		if o.OrderIndex != i+1 {
			t.Fatalf("order_index %d at position %d", o.OrderIndex, i)
		}
		seen[o.ID] = true
	}
	if len(opts) != 4 || len(seen) != 4 {
		t.Fatalf("options lost or duplicated: %+v", opts)
	}
}

func TestPlayerOptionsIsStableForAKey(t *testing.T) {
	q := sampleQuestion()
	a := PlayerOptions(q, "match:m1:u1")
	b := PlayerOptions(q, "match:m1:u1")
	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatal("same key must give the same order (page reload must not reshuffle)")
		}
	}
}

// The option that is stored first should land in every position about
// equally often across sessions — i.e. "the answer is always B" can't happen.
func TestPlayerOptionsSpreadsPositions(t *testing.T) {
	q := sampleQuestion()
	counts := make([]int, 4)
	const n = 8000
	for i := 0; i < n; i++ {
		opts := PlayerOptions(q, fmt.Sprintf("practice:session-%d", i))
		for pos, o := range opts {
			if o.ID == "opt-1" { // the correct one, stored as "B"
				counts[pos]++
			}
		}
	}
	for pos, c := range counts {
		if share := float64(c) / n; share < 0.22 || share > 0.28 {
			t.Errorf("correct option at position %d in %.1f%% of sessions", pos+1, share*100)
		}
	}
}

func TestPlayerOptionsWithoutKeyKeepsStoredOrder(t *testing.T) {
	opts := PlayerOptions(sampleQuestion(), "")
	for i, o := range opts {
		if o.ID != fmt.Sprintf("opt-%d", i) {
			t.Fatalf("no key should keep stored order, got %+v", opts)
		}
	}
}
