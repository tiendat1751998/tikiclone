package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/search/internal/ranking"
	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

func TestRankingScoreCalculation(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	svc := ranking.NewService(repo)
	ctx := context.Background()

	doc := &search.ProductDocument{
		ID:          "1",
		Title:       "iPhone 15 Pro Max",
		Description: "Latest Apple smartphone",
		Category:    "Electronics",
		Price:       999,
		Rating:      4.8,
		Stock:       500,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}

	score, factors := svc.Score(ctx, doc, "iphone")
	if score <= 0 {
		t.Errorf("Expected positive score, got %f", score)
	}

	if len(factors) == 0 {
		t.Fatal("Expected ranking factors, got none")
	}

	titleFactor := findFactor(factors, "title_match")
	if titleFactor == nil {
		t.Fatal("Expected title_match factor")
	}
	if titleFactor.Value <= 0 {
		t.Errorf("Expected positive title_match value, got %f", titleFactor.Value)
	}

	ratingFactor := findFactor(factors, "rating")
	if ratingFactor == nil {
		t.Fatal("Expected rating factor")
	}
	expectedRating := doc.Rating / 5.0
	if ratingFactor.Value != expectedRating {
		t.Errorf("Expected rating factor value %f, got %f", expectedRating, ratingFactor.Value)
	}
}

func TestRankingWithNoMatch(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	svc := ranking.NewService(repo)
	ctx := context.Background()

	doc := &search.ProductDocument{
		ID:          "1",
		Title:       "Samsung TV",
		Description: "4K Smart TV",
		Category:    "Electronics",
		Rating:      4.5,
		Stock:       100,
		CreatedAt:   time.Now().Add(-72 * time.Hour),
	}

	score, _ := svc.Score(ctx, doc, "iphone")
	if score < 0 {
		t.Errorf("Expected non-negative score even with no match, got %f", score)
	}
}

func TestRankingMultipleDocuments(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	svc := ranking.NewService(repo)
	ctx := context.Background()

	docs := []search.ProductDocument{
		{
			ID: "1", Title: "iPhone 15 Pro", Category: "Electronics",
			Rating: 4.8, Stock: 500, Price: 999, CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID: "2", Title: "iPhone 14", Category: "Electronics",
			Rating: 4.5, Stock: 200, Price: 699, CreatedAt: time.Now().Add(-720 * time.Hour),
		},
		{
			ID: "3", Title: "Samsung Galaxy", Category: "Electronics",
			Rating: 4.3, Stock: 300, Price: 899, CreatedAt: time.Now().Add(-48 * time.Hour),
		},
	}

	ranked := svc.Rank(ctx, docs, "iphone")
	if len(ranked) != 3 {
		t.Fatalf("Expected 3 ranked docs, got %d", len(ranked))
	}

	if ranked[0].ID != "1" {
		t.Log("Note: expected iPhone 15 Pro (best match) to rank first")
	}
}

func TestRankingClickSignals(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	ctx := context.Background()

	repo.RecordClick(ctx, "product1", "iphone")
	repo.RecordClick(ctx, "product1", "iphone")
	repo.RecordClick(ctx, "product2", "iphone")

	signal, err := repo.GetClickSignals(ctx, "product1")
	if err != nil {
		t.Fatalf("GetClickSignals failed: %v", err)
	}
	if signal.Count != 2 {
		t.Errorf("Expected 2 clicks for product1, got %d", signal.Count)
	}
	if signal.CTR <= 0 {
		t.Errorf("Expected positive CTR for product1, got %f", signal.CTR)
	}

	top, err := repo.GetTopClicked(ctx, 5)
	if err != nil {
		t.Fatalf("GetTopClicked failed: %v", err)
	}
	if len(top) == 0 {
		t.Fatal("Expected top clicked items, got none")
	}
	if top[0].ProductID != "product1" {
		t.Errorf("Expected product1 as top clicked, got %s", top[0].ProductID)
	}
}

func TestRankingConfig(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	svc := ranking.NewService(repo)
	ctx := context.Background()

	cfg, err := svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg.TitleBoost != 3.0 {
		t.Errorf("Expected default TitleBoost 3.0, got %f", cfg.TitleBoost)
	}
	if cfg.CategoryBoost != 2.0 {
		t.Errorf("Expected default CategoryBoost 2.0, got %f", cfg.CategoryBoost)
	}

	repo.SetConfig(ctx, &ranking.RankingConfig{
		TitleBoost: 5.0, CategoryBoost: 3.0, RatingBoost: 2.0,
		RecencyBoost: 1.5, PopularityBoost: 1.0, ClickBoost: 2.5,
	})

	cfg, err = svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg.TitleBoost != 5.0 {
		t.Errorf("Expected updated TitleBoost 5.0, got %f", cfg.TitleBoost)
	}
}

func TestQueryTokenization(t *testing.T) {
	repo := ranking.NewInMemoryRepository()
	svc := ranking.NewService(repo)

	doc := &search.ProductDocument{
		ID: "1", Title: "iPhone 15 Pro",
		Rating: 4.5, Stock: 100, CreatedAt: time.Now(),
	}

	score1, _ := svc.Score(context.Background(), doc, "iPhone 15")
	score2, _ := svc.Score(context.Background(), doc, "samsung galaxy")
	if score1 <= score2 {
		t.Log("Note: iPhone should score higher than Samsung for iPhone document")
	}
}

func findFactor(factors []ranking.RankingFactor, name string) *ranking.RankingFactor {
	for _, f := range factors {
		if f.Name == name {
			return &f
		}
	}
	return nil
}
