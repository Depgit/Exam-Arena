package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/exam-arena/internal/models"
)

// QuestionBank pre-warms a pool of published questions per exam category
// so that matchmaking never blocks on a cold DB query.
//
// Layout:
//
//	bank map[categoryID] []Question   — full question objects with options
//
// Refresh strategy:
//   - Warm() fetches all published questions for all active categories at startup.
//   - A background goroutine calls Warm() every RefreshInterval so stale
//     questions (newly published / archived) are picked up automatically.
//
// Thread safety:
//   - A sync.RWMutex protects the bank. Reads (GetRandom) hold RLock;
//     the refresh goroutine holds full Lock only for the swap.
//
// When you move to Redis:
//   - Store each question as a Redis Hash (HSET question:<id> ...)
//   - Keep a Redis Set per category (SADD category:<id>:questions <qid>)
//   - SRANDMEMBER gives O(1) random selection without loading everything
type QuestionBank struct {
	mu              sync.RWMutex
	bank            map[string][]models.Question // categoryID → questions
	RefreshInterval time.Duration
	loader          QuestionLoader
}

// QuestionLoader is the function the bank calls to reload.
// Pass repository.QuestionRepo.GetAllPublished as the implementation.
type QuestionLoader func(ctx context.Context, categoryID string) ([]models.Question, error)

// CategoryLister returns the IDs of all active exam categories.
type CategoryLister func(ctx context.Context) ([]string, error)

func NewQuestionBank(loader QuestionLoader, refreshInterval time.Duration) *QuestionBank {
	return &QuestionBank{
		bank:            make(map[string][]models.Question),
		RefreshInterval: refreshInterval,
		loader:          loader,
	}
}

// Warm fetches all questions for the given category IDs and replaces the bank atomically.
func (qb *QuestionBank) Warm(ctx context.Context, categoryIDs []string) {
	slog.Info("warming question bank", "categories", len(categoryIDs))

	newBank := make(map[string][]models.Question, len(categoryIDs))
	for _, catID := range categoryIDs {
		questions, err := qb.loader(ctx, catID)
		if err != nil {
			slog.Error("failed to load questions for category", "category", catID, "error", err)
			continue
		}
		newBank[catID] = questions
		slog.Info("questions loaded", "category", catID, "count", len(questions))
	}

	qb.mu.Lock()
	qb.bank = newBank
	qb.mu.Unlock()
}

// StartAutoRefresh runs Warm every RefreshInterval in a background goroutine.
func (qb *QuestionBank) StartAutoRefresh(ctx context.Context, listCategories CategoryLister) {
	go func() {
		ticker := time.NewTicker(qb.RefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				catIDs, err := listCategories(ctx)
				if err != nil {
					slog.Error("failed to list categories", "error", err)
					continue
				}
				qb.Warm(ctx, catIDs)
			}
		}
	}()
}

// GetRandom returns exactly n questions for the given category, chosen
// by a Fisher-Yates partial shuffle so every call gets a different set.
//
// Returns an error if fewer than n questions are available.
func (qb *QuestionBank) GetRandom(categoryID string, n int) ([]models.Question, error) {
	qb.mu.RLock()
	pool, ok := qb.bank[categoryID]
	// Copy the slice header so we can release the read lock before shuffling.
	// We do NOT copy the underlying Question structs — they are read-only.
	poolCopy := make([]models.Question, len(pool))
	copy(poolCopy, pool)
	qb.mu.RUnlock()

	if !ok || len(poolCopy) < n {
		return nil, fmt.Errorf(
			"not enough questions for category %s: have %d, need %d",
			categoryID, len(poolCopy), n,
		)
	}

	// Partial Fisher-Yates: shuffle only the first n elements.
	// Uses the current nanosecond as a cheap seed — good enough for
	// game question selection, not for cryptography.
	selected := partialShuffle(poolCopy, n)
	return selected, nil
}

// CategorySize returns how many questions are cached for a category.
func (qb *QuestionBank) CategorySize(categoryID string) int {
	qb.mu.RLock()
	defer qb.mu.RUnlock()
	return len(qb.bank[categoryID])
}

// ── Fisher-Yates partial shuffle ─────────────────────────────────────
// We only need the first n elements, so we stop after n swaps.
// Time complexity: O(n) — independent of the pool size.

func partialShuffle(pool []models.Question, n int) []models.Question {
	// Use a simple LCG seeded from current time for speed.
	// Replace with crypto/rand if question order must be unpredictable.
	seed := uint64(time.Now().UnixNano())

	for i := 0; i < n; i++ {
		seed = lcgNext(seed)
		j := i + int(seed>>33)%(len(pool)-i)
		pool[i], pool[j] = pool[j], pool[i]
	}
	return pool[:n]
}

func lcgNext(state uint64) uint64 {
	// Knuth's LCG constants
	return state*6364136223846793005 + 1442695040888963407
}
