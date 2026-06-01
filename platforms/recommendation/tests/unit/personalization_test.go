package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/recommendation/internal/personalization"
)

func TestBuildUserProfile(t *testing.T) {
	svc := personalization.NewService(personalization.NewInMemoryRepository())
	ctx := context.Background()

	events := []personalization.InteractionEvent{
		{UserID: "user1", Category: "Electronics", Brand: "Apple", Price: 999, Tags: []string{"premium", "mobile"}, Weight: 1.0, Timestamp: time.Now()},
		{UserID: "user1", Category: "Electronics", Brand: "Samsung", Price: 899, Tags: []string{"mobile", "flagship"}, Weight: 0.5, Timestamp: time.Now().Add(-1 * time.Hour)},
		{UserID: "user1", Category: "Fashion", Brand: "Nike", Price: 129, Tags: []string{"shoes"}, Weight: 0.3, Timestamp: time.Now().Add(-2 * time.Hour)},
	}

	profile, err := svc.BuildProfile(ctx, events)
	if err != nil {
		t.Fatalf("BuildProfile failed: %v", err)
	}

	if profile == nil {
		t.Fatal("Expected non-nil profile")
	}

	if profile.UserID != "user1" {
		t.Errorf("Expected user1, got %s", profile.UserID)
	}

	if len(profile.CategoryWeights) == 0 {
		t.Error("Expected category weights")
	}

	if profile.TotalInteractions != 3 {
		t.Errorf("Expected 3 interactions, got %d", profile.TotalInteractions)
	}
}

func TestScoreItemAgainstProfile(t *testing.T) {
	svc := personalization.NewService(personalization.NewInMemoryRepository())
	ctx := context.Background()

	profile := &personalization.UserProfile{
		UserID:          "user1",
		CategoryWeights: map[string]float64{"Electronics": 0.8, "Fashion": 0.2},
		PreferredBrands: map[string]float64{"Apple": 0.6, "Samsung": 0.3},
		PreferredPriceMid: 500,
		InterestVector:  map[string]float64{"premium": 0.7},
		TotalInteractions: 5,
	}

	score := svc.ScoreItem(ctx, profile, "Electronics", "Apple", 999, []string{"premium"})
	if score <= 0 {
		t.Errorf("Expected positive score for matching item, got %f", score)
	}

	lowScore := svc.ScoreItem(ctx, profile, "Home", "Unknown", 10, []string{"cheap"})
	if lowScore >= score {
		t.Errorf("Expected lower score for non-matching item, got %f vs %f", lowScore, score)
	}
}

func TestScoreItemNilProfile(t *testing.T) {
	svc := personalization.NewService(personalization.NewInMemoryRepository())
	ctx := context.Background()

	score := svc.ScoreItem(ctx, nil, "Electronics", "Apple", 999, []string{"premium"})
	if score != 0 {
		t.Errorf("Expected 0 for nil profile, got %f", score)
	}
}
