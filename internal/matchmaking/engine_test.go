package matchmaking

import (
	"context"
	"testing"
	"time"
)

func TestArenaPairsAcrossAnyRatingGap(t *testing.T) {
	pairs := make(chan PairedPlayers, 4)
	q := NewQueue()
	e := NewEngine(q, func(_ context.Context, p PairedPlayers) { pairs <- p })

	now := time.Now()
	q.Push(Entry{UserID: "low", ExamCategoryID: "c", MatchType: MatchTypeArena, Rating: 800, QueuedAt: now})
	q.Push(Entry{UserID: "high", ExamCategoryID: "c", MatchType: MatchTypeArena, Rating: 2400, QueuedAt: now})
	// Same gap in ranked must not pair.
	q.Push(Entry{UserID: "rlow", ExamCategoryID: "c", MatchType: "ranked", Rating: 800, QueuedAt: now})
	q.Push(Entry{UserID: "rhigh", ExamCategoryID: "c", MatchType: "ranked", Rating: 2400, QueuedAt: now})

	e.tick(context.Background())

	select {
	case p := <-pairs:
		if p.MatchType != MatchTypeArena {
			t.Fatalf("want the arena pair, got %+v", p)
		}
	case <-time.After(time.Second):
		t.Fatal("arena players were not paired")
	}
	select {
	case p := <-pairs:
		t.Fatalf("ranked players 1600 apart must not pair, got %+v", p)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestLongestWaitingPair(t *testing.T) {
	now := time.Now()
	pool := []Entry{
		{UserID: "new", QueuedAt: now},
		{UserID: "oldest", QueuedAt: now.Add(-30 * time.Second)},
		{UserID: "older", QueuedAt: now.Add(-10 * time.Second)},
	}
	a, b, ok := longestWaitingPair(pool)
	if !ok || a.UserID != "oldest" || b.UserID != "older" {
		t.Fatalf("got %s, %s, %v", a.UserID, b.UserID, ok)
	}
}
