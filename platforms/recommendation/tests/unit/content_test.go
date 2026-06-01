package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/content"
)

func TestContentCategoryFullMatch(t *testing.T) {
	svc := content.NewService(content.NewInMemoryRepository())
	ctx := context.Background()

	score, err := svc.ScoreProduct(ctx,
		&content.ProductFeatures{ProductID: "a", Category: "Electronics", Tags: []string{"mobile"}, Price: 100},
		&content.ProductFeatures{ProductID: "b", Category: "Electronics", Tags: []string{"mobile"}, Price: 100},
	)
	if err != nil {
		t.Fatalf("ScoreProduct failed: %v", err)
	}

	if score <= 0 {
		t.Errorf("Expected positive score for same category, got %f", score)
	}
}

func TestContentCategoryParentMatch(t *testing.T) {
	svc := content.NewService(content.NewInMemoryRepository())
	ctx := context.Background()

	score, err := svc.ScoreProduct(ctx,
		&content.ProductFeatures{ProductID: "a", Category: "Smartphones", ParentCategory: "Electronics", Tags: []string{"mobile"}, Price: 100},
		&content.ProductFeatures{ProductID: "b", Category: "Laptops", ParentCategory: "Electronics", Tags: []string{"laptop"}, Price: 200},
	)
	if err != nil {
		t.Fatalf("ScoreProduct failed: %v", err)
	}

	if score <= 0 {
		t.Errorf("Expected positive score for matching parent category, got %f", score)
	}
}

func TestContentTagJaccardSimilarity(t *testing.T) {
	svc := content.NewService(content.NewInMemoryRepository())
	ctx := context.Background()

	score, err := svc.ScoreProduct(ctx,
		&content.ProductFeatures{ProductID: "a", Category: "A", Tags: []string{"red", "blue", "green"}, Price: 100},
		&content.ProductFeatures{ProductID: "b", Category: "B", Tags: []string{"red", "blue", "yellow"}, Price: 100},
	)
	if err != nil {
		t.Fatalf("ScoreProduct failed: %v", err)
	}

	if score <= 0 {
		t.Errorf("Expected positive score for overlapping tags, got %f", score)
	}
}

func TestContentDifferentCategories(t *testing.T) {
	svc := content.NewService(content.NewInMemoryRepository())
	ctx := context.Background()

	score, err := svc.ScoreProduct(ctx,
		&content.ProductFeatures{ProductID: "a", Category: "Electronics", Tags: []string{"mobile"}, Price: 100},
		&content.ProductFeatures{ProductID: "b", Category: "Fashion", Tags: []string{"shoes"}, Price: 50},
	)
	if err != nil {
		t.Fatalf("ScoreProduct failed: %v", err)
	}

	if score > 0.3 {
		t.Errorf("Expected low score for different categories, got %f", score)
	}
}

func TestContentSimilarByContent(t *testing.T) {
	repo := content.NewInMemoryRepository()
	svc := content.NewService(repo)
	ctx := context.Background()

	repo.StoreProductFeatures(ctx, &content.ProductFeatures{ProductID: "p1", Category: "Electronics", Tags: []string{"mobile", "apple"}, Price: 999})
	repo.StoreProductFeatures(ctx, &content.ProductFeatures{ProductID: "p2", Category: "Electronics", Tags: []string{"mobile", "samsung"}, Price: 899})
	repo.StoreProductFeatures(ctx, &content.ProductFeatures{ProductID: "p3", Category: "Fashion", Tags: []string{"shoes"}, Price: 100})

	results, err := svc.SimilarByContent(ctx, "p1", 5)
	if err != nil {
		t.Fatalf("SimilarByContent failed: %v", err)
	}

	foundP2 := false
	for _, id := range results {
		if id == "p2" {
			foundP2 = true
			break
		}
	}

	if !foundP2 {
		t.Error("Expected p2 to be similar to p1")
	}

	p2pos := -1
	p3pos := -1
	for i, id := range results {
		if id == "p2" {
			p2pos = i
		}
		if id == "p3" {
			p3pos = i
		}
	}
	if p2pos < 0 {
		t.Error("Expected p2 in results")
	}
	if p3pos >= 0 && p2pos > 0 && p3pos < p2pos {
		t.Error("p2 should score higher than p3")
	}
}
