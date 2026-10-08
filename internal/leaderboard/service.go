package leaderboard

import (
	"context"
	"fmt"
	"time"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/cache"
)

type Service struct {
	repo  *Store
	cache cache.Cache
}

func NewService(repo *Store, cache cache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) GetLeaderboard(ctx context.Context, categoryCode string, limit, offset int) ([]models.LeaderboardEntry, error) {
	cacheKey := fmt.Sprintf("leaderboard:%s:%d:%d", categoryCode, limit, offset)

	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]models.LeaderboardEntry), nil
	}

	cat, err := s.repo.GetCategoryByCode(ctx, categoryCode)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", categoryCode)
	}

	entries, err := s.repo.GetLeaderboard(ctx, cat.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Adjust ranks for offset
	for i := range entries {
		entries[i].Rank = offset + i + 1
	}

	s.cache.Set(cacheKey, entries, 1*time.Minute) // Short TTL for near-realtime

	return entries, nil
}

func (s *Service) GetCategoryByCode(ctx context.Context, code string) (*models.ExamCategory, error) {
	return s.repo.GetCategoryByCode(ctx, code)
}
