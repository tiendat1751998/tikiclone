package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/trending"
)

func TestTrendingVelocityScoring(t *testing.T) {
	repo := trending.NewInMemoryRepository()
	svc := trending.NewService(repo)
	ctx := context.Background()

	repo.RecordInteraction(ctx, "hot1", 1.0)
	repo.RecordInteraction(ctx, "hot1", 1.0)
	repo.RecordInteraction(ctx, "hot1", 1.0)
	repo.RecordInteraction(ctx, "hot2", 1.0)
	repo.RecordInteraction(ctx, "hot3", 0.5)

	results, err := svc.GetTrending(ctx, 5)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected trending results")
	}

	if results[0].ProductID != "hot1" {
		t.Errorf("Expected hot1 to be most trending, got %s", results[0].ProductID)
	}

	if results[0].Score <= 0 {
		t.Errorf("Expected positive score, got %f", results[0].Score)
	}

	if results[0].Score > 1.0 {
		t.Errorf("Expected normalized score <= 1.0, got %f", results[0].Score)
	}
}

func TestTrendingEmptyData(t *testing.T) {
	svc := trending.NewService(trending.NewInMemoryRepository())
	ctx := context.Background()

	results, err := svc.GetTrending(ctx, 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected empty results, got %d", len(results))
	}
}

func TestTrendingRecordInteraction(t *testing.T) {
	svc := trending.NewService(trending.NewInMemoryRepository())
	ctx := context.Background()

	err := svc.RecordInteraction(ctx, "prod1")
	if err != nil {
		t.Fatalf("RecordInteraction failed: %v", err)
	}

	results, err := svc.GetTrending(ctx, 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	found := false
	for _, r := range results {
		if r.ProductID == "prod1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected prod1 in trending after recording interaction")
	}
}
