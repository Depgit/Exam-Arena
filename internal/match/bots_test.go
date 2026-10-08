package match

import (
	randv2 "math/rand/v2"
	"testing"

	"github.com/exam-arena/internal/models"
)

func TestBotForPicksByRating(t *testing.T) {
	for rating, want := range map[int]string{900: "bot.rookie", 1149: "bot.rookie", 1150: "bot.ace", 1300: "bot.ace", 1350: "bot.master", 1900: "bot.master"} {
		if got := botFor(rating).Username; got != want {
			t.Errorf("rating %d: got %s, want %s", rating, got, want)
		}
	}
}

func TestBotProfilesFinishInsideTheMatchClock(t *testing.T) {
	for _, b := range botProfiles {
		// Even answering every question at its slowest, the bot must finish
		// before the match timer, or the early finish could never trigger.
		if worst := b.MaxDelay * QuestionsPerMatch; worst.Seconds() >= MatchTimerSeconds {
			t.Errorf("%s worst case %v exceeds the %ds match", b.Username, worst, MatchTimerSeconds)
		}
	}
}

func TestBotPickAccuracy(t *testing.T) {
	q := models.Question{Options: []models.QuestionOption{
		{ID: "a"}, {ID: "b", IsCorrect: true}, {ID: "c"}, {ID: "d"},
	}}
	r := randv2.New(randv2.NewPCG(1, 2))
	for _, acc := range []float64{0.5, 0.7, 0.88} {
		correct := 0
		const n = 20000
		for i := 0; i < n; i++ {
			if botPick(r, q, acc) == "b" {
				correct++
			}
		}
		if got := float64(correct) / n; got < acc-0.02 || got > acc+0.02 {
			t.Errorf("accuracy %.2f: bot was right %.3f of the time", acc, got)
		}
	}
}
