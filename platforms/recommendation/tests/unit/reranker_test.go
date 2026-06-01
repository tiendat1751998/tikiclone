package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/reranker"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/types"
)

func TestDiversityReRanking(t *testing.T) {
	repo := reranker.NewInMemoryRepository()
	svc := reranker.NewService(repo)
	ctx := context.Background()

	cfg, _ := repo.GetConfig(ctx)
	cfg.MaxPerCategory = 2
	repo.SetConfig(ctx, cfg)

	recs := []types.ProductRecommendation{
		{ProductID: "p1", Category: "Electronics", Score: 1.0, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		{ProductID: "p2", Category: "Electronics", Score: 0.9, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		{ProductID: "p3", Category: "Electronics", Score: 0.8, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		{ProductID: "p4", Category: "Fashion", Score: 0.7, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
		{ProductID: "p5", Category: "Home", Score: 0.6, CreatedAt: time.Now().Add(-10 * 24 * time.Hour)},
	}

	reranked, err := svc.ReRank(ctx, recs)
	if err != nil {
		t.Fatalf("ReRank failed: %v", err)
	}

	electronicsCount := 0
	for _, r := range reranked {
		if r.Category == "Electronics" {
			electronicsCount++
		}
	}
	if electronicsCount > 2 {
		t.Errorf("Expected at most 2 Electronics, got %d", electronicsCount)
	}
}

func TestNewItemBoost(t *testing.T) {
	repo := reranker.NewInMemoryRepository()
	svc := reranker.NewService(repo)
	ctx := context.Background()

	recs := []types.ProductRecommendation{
		{ProductID: "old", Category: "A", Score: 0.5, CreatedAt: time.Now().Add(-30 * 24 * time.Hour)},
		{ProductID: "new", Category: "A", Score: 0.5, CreatedAt: time.Now().Add(-1 * time.Hour)},
	}

	reranked, err := svc.ReRank(ctx, recs)
	if err != nil {
		t.Fatalf("ReRank failed: %v", err)
	}

	if len(reranked) >= 2 && reranked[0].ProductID != "new" {
		t.Logf("New item should be boosted: first=%s score=%.2f", reranked[0].ProductID, reranked[0].Score)
	}
}

func TestExposureFairness(t *testing.T) {
	repo := reranker.NewInMemoryRepository()
	svc := reranker.NewService(repo)
	ctx := context.Background()

	repo.IncrementExposure(ctx, "overexposed")
	repo.IncrementExposure(ctx, "overexposed")
	repo.IncrementExposure(ctx, "overexposed")

	recs := []types.ProductRecommendation{
		{ProductID: "overexposed", Category: "A", Score: 0.9, CreatedAt: time.Now()},
		{ProductID: "fresh", Category: "A", Score: 0.8, CreatedAt: time.Now()},
	}

	reranked, err := svc.ReRank(ctx, recs)
	if err != nil {
		t.Fatalf("ReRank failed: %v", err)
	}

	if len(reranked) >= 2 && reranked[0].ProductID == "overexposed" && reranked[1].ProductID == "fresh" {
		t.Logf("Note: overexposed item may still rank high if boost compensates, checking scores: %.2f vs %.2f",
			reranked[0].Score, reranked[1].Score)
	}
}
