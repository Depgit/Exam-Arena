package matchmaking

import (
	"fmt"
	"sync"
	"time"
)

// Entry is one player sitting in the matchmaking queue.
// The layout mirrors a Redis Sorted Set: member = UserID, score = Rating.
// When you move to Redis, replace this struct with a Redis ZADD call and
// store the rest (CategoryID, MatchType, QueuedAt) in a Redis Hash keyed
// by UserID. The Engine interface below stays identical.
type Entry struct {
	UserID         string
	Username       string
	ExamCategoryID string
	MatchType      string
	Rating         int
	QueuedAt       time.Time
}

// Key uniquely identifies a matchmaking pool.
// Two players only compete against each other if they share the same key.
type Key struct {
	CategoryID string
	MatchType  string
}

func (k Key) String() string {
	return fmt.Sprintf("%s:%s", k.CategoryID, k.MatchType)
}

// Queue is a thread-safe, in-memory matchmaking queue.
//
// Internal structure:
//
//	pools map[Key][]Entry   — one slice per (category, matchType) pair
//
// Each slice is kept insertion-ordered (FIFO). The Engine reads from it
// and pairs players; it never modifies it directly — all mutations go
// through Push / Remove so the mutex is always held correctly.
type Queue struct {
	mu    sync.Mutex
	pools map[Key][]Entry
}

func NewQueue() *Queue {
	return &Queue{
		pools: make(map[Key][]Entry),
	}
}

// Push adds a player to the right pool.
// If the player is already queued (same UserID in any pool), the old
// entry is removed first so a player can never appear twice.
func (q *Queue) Push(e Entry) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Remove stale entry across all pools (player re-queued with new settings)
	for k, pool := range q.pools {
		q.pools[k] = removeByUserID(pool, e.UserID)
	}

	key := Key{CategoryID: e.ExamCategoryID, MatchType: e.MatchType}
	q.pools[key] = append(q.pools[key], e)
}

// Remove pulls a specific player out of whichever pool they're in.
// Safe to call if the player is not in the queue (no-op).
func (q *Queue) Remove(userID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for k, pool := range q.pools {
		q.pools[k] = removeByUserID(pool, userID)
	}
}

// IsQueued returns true if the player is currently waiting.
func (q *Queue) IsQueued(userID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, pool := range q.pools {
		for _, e := range pool {
			if e.UserID == userID {
				return true
			}
		}
	}
	return false
}

// Snapshot returns a copy of all pools so the Engine can read without
// holding the lock during the (potentially slow) pairing logic.
func (q *Queue) Snapshot() map[Key][]Entry {
	q.mu.Lock()
	defer q.mu.Unlock()

	snap := make(map[Key][]Entry, len(q.pools))
	for k, pool := range q.pools {
		copied := make([]Entry, len(pool))
		copy(copied, pool)
		snap[k] = copied
	}
	return snap
}

// ClaimPair atomically removes exactly these two users from the queue.
// Returns false if either user is no longer present (they left or were
// already claimed by a concurrent call — impossible with a single Engine
// goroutine, but this makes the code safe for future parallelism).
func (q *Queue) ClaimPair(a, b Entry) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := Key{CategoryID: a.ExamCategoryID, MatchType: a.MatchType}
	pool, ok := q.pools[key]
	if !ok {
		return false
	}

	foundA, foundB := false, false
	for _, e := range pool {
		if e.UserID == a.UserID {
			foundA = true
		}
		if e.UserID == b.UserID {
			foundB = true
		}
	}
	if !foundA || !foundB {
		return false // one of them left between snapshot and claim
	}

	pool = removeByUserID(pool, a.UserID)
	pool = removeByUserID(pool, b.UserID)
	q.pools[key] = pool
	return true
}

// Stats returns queue depth per pool — useful for an admin endpoint.
func (q *Queue) Stats() map[string]int {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make(map[string]int, len(q.pools))
	for k, pool := range q.pools {
		out[k.String()] = len(pool)
	}
	return out
}

// ── helpers ──────────────────────────────────────────────────────────

func removeByUserID(pool []Entry, userID string) []Entry {
	// Build in-place without allocation when the entry is absent
	out := pool[:0]
	for _, e := range pool {
		if e.UserID != userID {
			out = append(out, e)
		}
	}
	return out
}
