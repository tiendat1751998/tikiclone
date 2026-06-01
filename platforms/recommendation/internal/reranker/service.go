package reranker

import (
	"context"
	"sort"
	"time"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/types"
)

type Service interface {
	ReRank(ctx context.Context, recs []types.ProductRecommendation) ([]types.ProductRecommendation, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ReRank(ctx context.Context, recs []types.ProductRecommendation) ([]types.ProductRecommendation, error) {
	if len(recs) == 0 {
		return recs, nil
	}

	cfg, err := s.repo.GetConfig(ctx)
	if err != nil {
		cfg = DefaultReRankConfig()
	}

	type item struct {
		rec      types.ProductRecommendation
		adjusted float64
		category string
		exposure int
		isNew    bool
	}

	items := make([]item, len(recs))
	for i, r := range recs {
		adjusted := r.Score
		isNew := time.Since(r.CreatedAt).Hours() < float64(cfg.NewItemHours)
		if isNew {
			adjusted += cfg.NewItemBoost
		}

		expCount, _ := s.repo.GetExposureCount(ctx, r.ProductID)
		adjusted -= float64(expCount) * cfg.ExposureDownrank
		if adjusted < 0 {
			adjusted = 0
		}

		items[i] = item{
			rec:      r,
			adjusted: adjusted,
			category: r.Category,
			exposure: expCount,
			isNew:    isNew,
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].adjusted > items[j].adjusted
	})

	categoryCount := make(map[string]int)
	var final []types.ProductRecommendation

	for _, it := range items {
		if categoryCount[it.category] >= cfg.MaxPerCategory {
			continue
		}
		categoryCount[it.category]++
		it.rec.Score = it.adjusted
		final = append(final, it.rec)

		s.repo.IncrementExposure(ctx, it.rec.ProductID)
	}

	sort.Slice(final, func(i, j int) bool {
		return final[i].Score > final[j].Score
	})

	return final, nil
}
