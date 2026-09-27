package matchmaking

import (
	"context"
	"log/slog"
	"math"
	"sort"
	"time"
)

// PairedPlayers is what the Engine hands off to the MatchService.
type PairedPlayers struct {
	PlayerA        Entry
	PlayerB        Entry
	ExamCategoryID string
	MatchType      string
}

// OnMatchFound is the callback the Engine calls when it pairs two players.
// The MatchService registers itself here. This keeps the Engine decoupled
// from the MatchService (no import cycle).
type OnMatchFound func(ctx context.Context, pair PairedPlayers)

// Engine is the matchmaking brain. It runs a single goroutine that:
//  1. Takes a snapshot of the queue (no lock held during pairing)
//  2. Finds the best pair in each pool
//  3. Atomically claims those two players
//  4. Calls OnMatchFound (which starts the match in a new goroutine)
//
// Single-goroutine design means zero race conditions with zero mutexes
// inside the Engine itself. The Queue handles its own concurrency.
type Engine struct {
	queue        *Queue
	onMatchFound OnMatchFound
	tickInterval time.Duration

	// Rating tolerance starts low and expands the longer a player waits.
	baseRatingRange    int     // initial ±rating window
	ratingExpandPerSec float64 // how much the window grows per second waited
	maxRatingRange     int     // cap so we don't match 800 vs 2400
}

func NewEngine(q *Queue, onMatchFound OnMatchFound) *Engine {
	return &Engine{
		queue:              q,
		onMatchFound:       onMatchFound,
		tickInterval:       2 * time.Second,
		baseRatingRange:    100,
		ratingExpandPerSec: 5, // ±5 rating per second → ±100 after 20s, ±300 after 60s
		maxRatingRange:     400,
	}
}

// Run starts the matchmaking loop. Cancel ctx to stop it cleanly.
func (e *Engine) Run(ctx context.Context) {
	slog.Info("matchmaking engine started",
		"tick", e.tickInterval,
		"base_range", e.baseRatingRange,
		"max_range", e.maxRatingRange,
	)

	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("matchmaking engine stopped")
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}

// tick processes every pool once.
func (e *Engine) tick(ctx context.Context) {
	snap := e.queue.Snapshot()

	for _, pool := range snap {
		if len(pool) < 2 {
			continue
		}
		e.processPool(ctx, pool)
	}
}

// processPool greedily pairs the best matches from one pool.
// It keeps pairing until no eligible pair remains.
func (e *Engine) processPool(ctx context.Context, pool []Entry) {
	// Work on a local copy; after each claim we shrink it
	remaining := make([]Entry, len(pool))
	copy(remaining, pool)

	// Every entry in a pool shares its match type.
	openToAll := pool[0].MatchType == MatchTypeArena

	for len(remaining) >= 2 {
		var a, b Entry
		var found bool
		if openToAll {
			a, b, found = longestWaitingPair(remaining)
		} else {
			a, b, found = e.findBestPair(remaining)
		}
		if !found {
			break
		}

		// Atomically claim — if either left between snapshot and now, skip
		if !e.queue.ClaimPair(a, b) {
			// Remove whichever is gone from our local slice and retry
			remaining = removeByUserID(remaining, a.UserID)
			remaining = removeByUserID(remaining, b.UserID)
			continue
		}

		// Remove from local slice so we don't re-pair them
		remaining = removeByUserID(remaining, a.UserID)
		remaining = removeByUserID(remaining, b.UserID)

		pair := PairedPlayers{
			PlayerA:        a,
			PlayerB:        b,
			ExamCategoryID: a.ExamCategoryID,
			MatchType:      a.MatchType,
		}

		slog.Info("match found",
			"player_a", a.Username,
			"rating_a", a.Rating,
			"player_b", b.Username,
			"rating_b", b.Rating,
			"delta", abs(a.Rating-b.Rating),
			"category", a.ExamCategoryID,
		)

		// Hand off to MatchService in its own goroutine so the engine
		// loop never blocks on DB/WS work.
		go e.onMatchFound(ctx, pair)
	}
}

// MatchTypeArena is open to everyone: players are paired regardless of
// rating, longest-waiting first.
const MatchTypeArena = "arena"

// longestWaitingPair returns the two players who have waited longest.
func longestWaitingPair(pool []Entry) (Entry, Entry, bool) {
	if len(pool) < 2 {
		return Entry{}, Entry{}, false
	}
	sorted := make([]Entry, len(pool))
	copy(sorted, pool)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].QueuedAt.Before(sorted[j].QueuedAt) })
	return sorted[0], sorted[1], true
}

// findBestPair scans all O(n²) pairs and returns the one with the smallest
// effective rating difference that still falls within each player's tolerance.
//
// "Effective tolerance" grows with wait time, so a player who has been
// waiting 60 seconds accepts a much wider range than one who just joined.
func (e *Engine) findBestPair(pool []Entry) (Entry, Entry, bool) {
	bestDiff := math.MaxFloat64
	bestI, bestJ := -1, -1

	now := time.Now()

	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			a, b := pool[i], pool[j]

			// Each player gets their own tolerance based on their own wait time
			tolA := e.tolerance(now.Sub(a.QueuedAt))
			tolB := e.tolerance(now.Sub(b.QueuedAt))
			// The effective window is the larger of the two
			effectiveTol := math.Max(tolA, tolB)

			diff := math.Abs(float64(a.Rating - b.Rating))
			if diff <= effectiveTol && diff < bestDiff {
				bestDiff = diff
				bestI, bestJ = i, j
			}
		}
	}

	if bestI == -1 {
		return Entry{}, Entry{}, false
	}
	return pool[bestI], pool[bestJ], true
}

// tolerance returns the acceptable ±rating window for a player who has
// been waiting for `waited` duration.
func (e *Engine) tolerance(waited time.Duration) float64 {
	expanded := float64(e.baseRatingRange) + waited.Seconds()*e.ratingExpandPerSec
	return math.Min(expanded, float64(e.maxRatingRange))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
