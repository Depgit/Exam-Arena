package questions

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/exam-arena/internal/models"
)

type Bank struct {
	mu              sync.RWMutex
	bank            map[string][]models.Question // categoryID → questions
	RefreshInterval time.Duration
	loader          Loader
}

type Loader func(ctx context.Context, categoryID string) ([]models.Question, error)

type CategoryLister func(ctx context.Context) ([]string, error)

func NewBank(loader Loader, refreshInterval time.Duration) *Bank {
	return &Bank{
		bank:            make(map[string][]models.Question),
		RefreshInterval: refreshInterval,
		loader:          loader,
	}
}

func (qb *Bank) Warm(ctx context.Context, categoryIDs []string) {
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

// RefreshCategory reloads one category, leaving the others untouched.
func (qb *Bank) RefreshCategory(ctx context.Context, categoryID string) error {
	questions, err := qb.loader(ctx, categoryID)
	if err != nil {
		return err
	}
	qb.mu.Lock()
	qb.bank[categoryID] = questions
	qb.mu.Unlock()
	return nil
}

func (qb *Bank) StartAutoRefresh(ctx context.Context, listCategories CategoryLister) {
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

func (qb *Bank) GetRandom(categoryID string, n int) ([]models.Question, error) {
	qb.mu.RLock()
	pool, ok := qb.bank[categoryID]
	poolCopy := make([]models.Question, len(pool))
	copy(poolCopy, pool)
	qb.mu.RUnlock()

	if !ok || len(poolCopy) < n {
		return nil, fmt.Errorf(
			"not enough questions for category %s: have %d, need %d",
			categoryID, len(poolCopy), n,
		)
	}

	selected := partialShuffle(poolCopy, n)
	return selected, nil
}

func (qb *Bank) CategorySize(categoryID string) int {
	qb.mu.RLock()
	defer qb.mu.RUnlock()
	return len(qb.bank[categoryID])
}

func partialShuffle(pool []models.Question, n int) []models.Question {
	seed := uint64(time.Now().UnixNano())

	for i := 0; i < n; i++ {
		seed = lcgNext(seed)
		j := i + int(seed>>33)%(len(pool)-i)
		pool[i], pool[j] = pool[j], pool[i]
	}
	return pool[:n]
}

func lcgNext(state uint64) uint64 {
	return state*6364136223846793005 + 1442695040888963407
}
