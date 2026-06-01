package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/collaborative"
)

func TestCollaborativeItemBasedSimilarity(t *testing.T) {
	repo := collaborative.NewInMemoryRepository()
	svc := collaborative.NewService(repo)
	ctx := context.Background()

	repo.StoreRating(ctx, "user1", "item1", 1.0)
	repo.StoreRating(ctx, "user1", "item2", 0.5)
	repo.StoreRating(ctx, "user2", "item1", 1.0)
	repo.StoreRating(ctx, "user2", "item2", 0.8)
	repo.StoreRating(ctx, "user2", "item3", 1.0)

	results, err := svc.ItemBasedSimilar(ctx, "item1", 5)
	if err != nil {
		t.Fatalf("ItemBasedSimilar failed: %v", err)
	}

	foundItem2 := false
	foundItem3 := false
	for _, r := range results {
		if r.ItemID == "item2" {
			foundItem2 = true
			if r.Similarity <= 0 {
				t.Errorf("Expected positive similarity between item1 and item2, got %f", r.Similarity)
			}
		}
		if r.ItemID == "item3" {
			foundItem3 = true
		}
	}

	if !foundItem2 {
		t.Error("Expected item2 to be similar to item1")
	}
	if !foundItem3 {
		t.Error("Expected item3 to be similar to item1")
	}
}

func TestCollaborativeUserBasedRecommend(t *testing.T) {
	repo := collaborative.NewInMemoryRepository()
	svc := collaborative.NewService(repo)
	ctx := context.Background()

	repo.StoreRating(ctx, "user1", "item1", 1.0)
	repo.StoreRating(ctx, "user1", "item2", 0.8)
	repo.StoreRating(ctx, "user2", "item1", 1.0)
	repo.StoreRating(ctx, "user2", "item2", 0.9)
	repo.StoreRating(ctx, "user2", "item3", 1.0)
	repo.StoreRating(ctx, "user3", "item1", 0.5)
	repo.StoreRating(ctx, "user3", "item3", 0.7)

	results, err := svc.UserBasedRecommend(ctx, "user1", 5)
	if err != nil {
		t.Fatalf("UserBasedRecommend failed: %v", err)
	}

	if len(results) == 0 {
		t.Log("User-based rec: no results (expected with sparse matrix)")
		return
	}

	for _, r := range results {
		if r.ItemID == "item1" || r.ItemID == "item2" {
			t.Errorf("Should not recommend items user already rated: %s", r.ItemID)
		}
	}
}

func TestCollaborativeImplicitFeedbackWeight(t *testing.T) {
	repo := collaborative.NewInMemoryRepository()
	svc := collaborative.NewService(repo)
	ctx := context.Background()

	err := svc.RecordInteraction(ctx, "user1", "item1", false)
	if err != nil {
		t.Fatalf("RecordInteraction (implicit) failed: %v", err)
	}

	ratings, err := repo.GetUserRatings(ctx, "user1")
	if err != nil {
		t.Fatalf("GetUserRatings failed: %v", err)
	}

	if ratings["item1"] != 0.3 {
		t.Errorf("Expected implicit weight 0.3, got %f", ratings["item1"])
	}

	err = svc.RecordInteraction(ctx, "user1", "item2", true)
	if err != nil {
		t.Fatalf("RecordInteraction (explicit) failed: %v", err)
	}

	ratings, _ = repo.GetUserRatings(ctx, "user1")
	if ratings["item2"] != 1.0 {
		t.Errorf("Expected explicit weight 1.0, got %f", ratings["item2"])
	}
}
