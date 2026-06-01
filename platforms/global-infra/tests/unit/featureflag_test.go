package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/featureflag"
)

func TestFeatureFlagCreate(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	flag := &featureflag.FeatureFlag{
		Name:              "test-flag",
		Enabled:           true,
		PercentageRollout: 50,
		UserSegment:       featureflag.SegmentAll,
	}

	created, err := svc.Create(context.Background(), flag)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Name != "test-flag" {
		t.Errorf("expected test-flag, got %s", created.Name)
	}
}

func TestFeatureFlagGet(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	flag := &featureflag.FeatureFlag{Name: "get-flag", Enabled: true, PercentageRollout: 100, UserSegment: featureflag.SegmentAll}
	svc.Create(context.Background(), flag)

	got, err := svc.Get(context.Background(), "get-flag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected flag to exist")
	}
	if got.Name != "get-flag" {
		t.Errorf("expected get-flag, got %s", got.Name)
	}
}

func TestFeatureFlagList(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{Name: "flag-a", Enabled: true, PercentageRollout: 100, UserSegment: featureflag.SegmentAll})
	svc.Create(context.Background(), &featureflag.FeatureFlag{Name: "flag-b", Enabled: false, PercentageRollout: 0, UserSegment: featureflag.SegmentAll})

	flags, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}
}

func TestFeatureFlagEvaluateEnabled(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{
		Name:              "active-flag",
		Enabled:           true,
		PercentageRollout: 100,
		UserSegment:       featureflag.SegmentAll,
	})

	result, err := svc.Evaluate(context.Background(), "active-flag", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Enabled {
		t.Errorf("expected flag to be enabled, got disabled: %s", result.Reason)
	}
}

func TestFeatureFlagEvaluateDisabled(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{
		Name:              "disabled-flag",
		Enabled:           false,
		PercentageRollout: 100,
		UserSegment:       featureflag.SegmentAll,
	})

	result, err := svc.Evaluate(context.Background(), "disabled-flag", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Enabled {
		t.Error("expected flag to be disabled")
	}
}

func TestFeatureFlagEvaluatePercentage(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{
		Name:              "pct-flag",
		Enabled:           true,
		PercentageRollout: 0,
		UserSegment:       featureflag.SegmentAll,
	})

	result, err := svc.Evaluate(context.Background(), "pct-flag", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Enabled {
		t.Error("expected flag to be disabled with 0%% rollout")
	}
}

func TestFeatureFlagEvaluateNotFound(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	result, err := svc.Evaluate(context.Background(), "nonexistent", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Enabled {
		t.Error("expected nonexistent flag to be disabled")
	}
}

func TestFeatureFlagValidateRollout(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	_, err := svc.Create(context.Background(), &featureflag.FeatureFlag{
		Name:              "bad-flag",
		Enabled:           true,
		PercentageRollout: 150,
		UserSegment:       featureflag.SegmentAll,
	})
	if err == nil {
		t.Error("expected error for invalid percentage")
	}
}

func TestFeatureFlagUpdate(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{Name: "update-flag", Enabled: false, PercentageRollout: 50, UserSegment: featureflag.SegmentAll})
	err := svc.Update(context.Background(), &featureflag.FeatureFlag{Name: "update-flag", Enabled: true, PercentageRollout: 100, UserSegment: featureflag.SegmentAll})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.Get(context.Background(), "update-flag")
	if got == nil || !got.Enabled {
		t.Error("expected flag to be enabled after update")
	}
}

func TestFeatureFlagSegmentFilter(t *testing.T) {
	repo := featureflag.NewInMemoryRepository()
	svc := featureflag.NewService(repo)

	svc.Create(context.Background(), &featureflag.FeatureFlag{
		Name:              "staff-flag",
		Enabled:           true,
		PercentageRollout: 100,
		UserSegment:       featureflag.SegmentStaff,
	})

	result, err := svc.Evaluate(context.Background(), "staff-flag", "regular-user-99999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Enabled {
		t.Error("expected regular user to not be in staff segment")
	}
}


