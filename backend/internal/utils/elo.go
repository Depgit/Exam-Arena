package utils

import "math"

const (
	DefaultRating = 1200
	KFactor       = 32
)

// CalculateElo returns new ratings for player A and B given the outcome.
// scoreA = 1.0 for win, 0.5 for draw, 0.0 for loss.
func CalculateElo(ratingA, ratingB int, scoreA float64) (newA, newB int) {
	ea := expectedScore(ratingA, ratingB)
	eb := 1.0 - ea
	scoreB := 1.0 - scoreA

	newA = ratingA + int(math.Round(KFactor*(scoreA-ea)))
	newB = ratingB + int(math.Round(KFactor*(scoreB-eb)))

	if newA < 100 {
		newA = 100
	}
	if newB < 100 {
		newB = 100
	}

	return newA, newB
}

func expectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}
