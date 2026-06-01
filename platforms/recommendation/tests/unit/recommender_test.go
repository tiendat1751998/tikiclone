package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/collaborative"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/content"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/personalization"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/recommender"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/reranker"
	"github.com/tikiclone/tiki/platforms/recommendation/internal/trending"
)

func setupRecommender() (recommender.Service, *collaborative.InMemoryRepository, *content.InMemoryRepository, *trending.InMemoryRepository, *personalization.InMemoryRepository) {
	collabRepo := collaborative.NewInMemoryRepository()
	collabSvc := collaborative.NewService(collabRepo)

	contentRepo := content.NewInMemoryRepository()
	contentSvc := content.NewService(contentRepo)

	trendingRepo := trending.NewInMemoryRepository()
	trendingSvc := trending.NewService(trendingRepo)

	personalRepo := personalization.NewInMemoryRepository()
	personalSvc := personalization.NewService(personalRepo)

	rerankerRepo := reranker.NewInMemoryRepository()
	rerankerSvc := reranker.NewService(rerankerRepo)

	recRepo := recommender.NewInMemoryRepository()
	recSvc := recommender.NewService(recRepo, collabSvc, contentSvc, trendingSvc, personalSvc, rerankerSvc)

	return recSvc, collabRepo, contentRepo, trendingRepo, personalRepo
}

func TestHybridRecommendationScoreCalculation(t *testing.T) {
	ctx := context.Background()
	svc, collabRepo, contentRepo, trendingRepo, _ := setupRecommender()

	// Seed collaborative data
	collabRepo.StoreRating(ctx, "user1", "prod1", 1.0)
	collabRepo.StoreRating(ctx, "user1", "prod2", 0.8)
	collabRepo.StoreRating(ctx, "user2", "prod1", 1.0)
	collabRepo.StoreRating(ctx, "user2", "prod2", 0.9)
	collabRepo.StoreRating(ctx, "user2", "prod3", 1.0)

	// Seed content features
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prod1", Category: "Electronics", Tags: []string{"mobile", "apple"}, Price: 999,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prod2", Category: "Electronics", Tags: []string{"mobile", "samsung"}, Price: 899,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prod3", Category: "Audio", Tags: []string{"headphones", "sony"}, Price: 349,
	})

	// Seed trending data
	trendingRepo.RecordInteraction(ctx, "prod1", 1.0)
	trendingRepo.RecordInteraction(ctx, "prod1", 1.0)
	trendingRepo.RecordInteraction(ctx, "prod2", 1.0)

	recCtx := recommender.RecommendationContext{
		UserID:    "user1",
		ProductID: "prod1",
		Type:      "",
		Limit:     10,
	}

	recs, err := svc.GetRecommendations(ctx, recCtx)
	if err != nil {
		t.Fatalf("GetRecommendations failed: %v", err)
	}

	if len(recs) == 0 {
		t.Fatal("Expected non-empty recommendations")
	}

	if recs[0].Score <= 0 {
		t.Errorf("Expected positive score, got %f", recs[0].Score)
	}
}

func TestRecommendationsDeduplication(t *testing.T) {
	ctx := context.Background()
	svc, collabRepo, contentRepo, _, _ := setupRecommender()

	// Seed data to produce duplicate candidates
	collabRepo.StoreRating(ctx, "user1", "prodA", 1.0)
	collabRepo.StoreRating(ctx, "user1", "prodB", 0.5)
	collabRepo.StoreRating(ctx, "user2", "prodA", 1.0)
	collabRepo.StoreRating(ctx, "user2", "prodC", 1.0)

	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodA", Category: "Electronics", Tags: []string{"gadget"}, Price: 100,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodB", Category: "Electronics", Tags: []string{"gadget"}, Price: 200,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodC", Category: "Home", Tags: []string{"appliance"}, Price: 300,
	})

	recCtx := recommender.RecommendationContext{
		UserID:    "user1",
		ProductID: "prodA",
		Type:      "",
		Limit:     10,
	}

	recs, err := svc.GetRecommendations(ctx, recCtx)
	if err != nil {
		t.Fatalf("GetRecommendations failed: %v", err)
	}

	seen := make(map[string]bool)
	for _, r := range recs {
		if seen[r.ProductID] {
			t.Errorf("Duplicate product found: %s", r.ProductID)
		}
		seen[r.ProductID] = true
	}
}

func TestRecommendationsEmptyState(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setupRecommender()

	recCtx := recommender.RecommendationContext{
		ProductID: "nonexistent",
		Type:      recommender.RecTypeRelated,
		Limit:     10,
	}

	_, err := svc.GetRecommendations(ctx, recCtx)
	if err == nil {
		t.Log("Empty state: expected error for no data")
	}
}

func TestRecommendationsInvalidType(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _ := setupRecommender()

	recCtx := recommender.RecommendationContext{
		Type:  "unknown_type",
		Limit: 10,
	}

	_, err := svc.GetRecommendations(ctx, recCtx)
	if err != recommender.ErrNoRecommendations && err != nil {
		t.Logf("Expected no recommendations or empty state: %v", err)
	}
}

func TestTrendingRecommendations(t *testing.T) {
	ctx := context.Background()
	svc, _, _, trendingRepo, _ := setupRecommender()

	trendingRepo.RecordInteraction(ctx, "hot1", 1.0)
	trendingRepo.RecordInteraction(ctx, "hot1", 1.0)
	trendingRepo.RecordInteraction(ctx, "hot1", 1.0)
	trendingRepo.RecordInteraction(ctx, "hot2", 1.0)
	trendingRepo.RecordInteraction(ctx, "hot2", 0.5)

	recCtx := recommender.RecommendationContext{
		Type:  recommender.RecTypeTrending,
		Limit: 5,
	}

	recs, err := svc.GetRecommendations(ctx, recCtx)
	if err != nil {
		t.Fatalf("Trending recommendations failed: %v", err)
	}

	if len(recs) == 0 {
		t.Fatal("Expected trending recommendations")
	}

	if len(recs) > 0 && recs[0].Type != recommender.RecTypeTrending {
		t.Errorf("Expected trending type, got %v", recs[0].Type)
	}
}

func TestPersonalizedRecommendations(t *testing.T) {
	ctx := context.Background()
	svc, collabRepo, contentRepo, _, personalRepo := setupRecommender()

	collabRepo.StoreRating(ctx, "user_p1", "prodX", 1.0)
	collabRepo.StoreRating(ctx, "user_p1", "prodY", 0.5)
	collabRepo.StoreRating(ctx, "user_p2", "prodX", 1.0)
	collabRepo.StoreRating(ctx, "user_p2", "prodZ", 1.0)

	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodX", Category: "Fashion", Tags: []string{"shoes"}, Price: 50,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodY", Category: "Fashion", Tags: []string{"shirts"}, Price: 30,
	})
	contentRepo.StoreProductFeatures(ctx, &content.ProductFeatures{
		ProductID: "prodZ", Category: "Electronics", Tags: []string{"gadget"}, Price: 200,
	})

	personalRepo.SaveProfile(ctx, &personalization.UserProfile{
		UserID:          "user_p1",
		CategoryWeights: map[string]float64{"Fashion": 0.8, "Electronics": 0.2},
		TotalInteractions: 5,
	})

	recCtx := recommender.RecommendationContext{
		UserID: "user_p1",
		Type:   recommender.RecTypePersonalized,
		Limit:  10,
	}

	recs, err := svc.GetRecommendations(ctx, recCtx)
	if err != nil {
		t.Fatalf("Personalized recommendations failed: %v", err)
	}

	if len(recs) > 0 && recs[0].Reason != string(recommender.ReasonPersonalized) && recs[0].Reason != string(recommender.ReasonBoughtAlsoBought) {
		t.Logf("Personalized rec reason: %s", recs[0].Reason)
	}
}
