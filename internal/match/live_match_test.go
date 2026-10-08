package match

import (
	"fmt"
	"sync"
	"testing"

	"github.com/exam-arena/internal/models"
)

func newTestLiveMatch(questions int, players ...string) *LiveMatch {
	lm := &LiveMatch{
		MatchID:      "m1",
		Questions:    make([]models.Question, questions),
		PlayerStates: make(map[string]*LivePlayer),
		cancelTimer:  func() {},
	}
	for i := range lm.Questions {
		lm.Questions[i].ID = fmt.Sprintf("q%d", i)
	}
	for _, p := range players {
		lm.PlayerStates[p] = &LivePlayer{UserID: p, Username: p, AnsweredIDs: map[string]bool{}}
	}
	return lm
}

func TestRecordAnswerScoringAndCompletion(t *testing.T) {
	lm := newTestLiveMatch(2, "a", "b")

	out, ok := lm.recordAnswer("a", "q0", true, 150)
	if !ok || out.duplicate || out.finished || out.score != 150 {
		t.Fatalf("first answer: ok=%v outcome=%+v", ok, out)
	}

	// A resend (even claiming a different result) changes nothing and
	// reports the original grading so the sender can be re-acknowledged.
	out, _ = lm.recordAnswer("a", "q0", false, 0)
	if !out.duplicate || out.score != 150 || !out.prevCorrect || out.prevPoints != 150 {
		t.Fatalf("duplicate answer should be ignored and report the original, got %+v", out)
	}

	if _, ok := lm.recordAnswer("intruder", "q0", true, 1); ok {
		t.Fatal("a user who is not in the match was accepted")
	}

	lm.recordAnswer("a", "q1", false, 0)
	lm.recordAnswer("b", "q0", false, 0)
	out, _ = lm.recordAnswer("b", "q1", true, 100)
	if !out.finished {
		t.Fatal("expected the match to be finished once every player answered everything")
	}

	res := lm.snapshotResults()
	if len(res) != 2 || res[0].userID != "a" || res[0].score != 150 || res[0].correct != 1 ||
		res[1].userID != "b" || res[1].score != 100 || res[1].answered != 2 {
		t.Fatalf("unexpected results: %+v", res)
	}
}

// The HTTP handler behind GET /matches/{id} reads the scoreboard while the
// hub goroutine records answers. Before LiveMatch had a mutex this was a
// fatal "concurrent map read and map write"; -race must stay clean here.
func TestConcurrentAnswersAndScoreboard(t *testing.T) {
	const n = 50
	lm := newTestLiveMatch(n, "a", "b")

	var wg sync.WaitGroup
	for _, p := range []string{"a", "b"} {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				lm.recordAnswer(p, fmt.Sprintf("q%d", i), i%2 == 0, 100)
			}
		}(p)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			lm.scoreBoard()
			lm.snapshotResults()
			lm.playerIDs()
		}
	}()
	wg.Wait()

	for _, r := range lm.snapshotResults() {
		if r.answered != n || r.correct != n/2 || r.score != n*100 {
			t.Fatalf("player %s ended with %+v", r.userID, r)
		}
	}
}
