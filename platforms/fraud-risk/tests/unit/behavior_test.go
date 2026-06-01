package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/behavior"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
)

func TestBuildProfile(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	profile, err := analyzer.BuildProfile(context.Background(), "user1", 14, "192.168.", "device-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != "user1" {
		t.Errorf("expected user1, got %s", profile.UserID)
	}
	if profile.TypicalLoginHour != 14 {
		t.Errorf("expected login hour 14, got %d", profile.TypicalLoginHour)
	}
}

func TestAnalyzeBehaviorNoDeviation(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	analyzer.BuildProfile(context.Background(), "user2", 10, "10.0.", "device-xyz")

	ev := &core.Event{
		Type:      core.EventLogin,
		UserID:    "user2",
		IP:        "10.0.1.1",
		DeviceID:  "device-xyz",
		Timestamp: time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
	}

	result, err := analyzer.AnalyzeBehavior(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasDeviation {
		t.Errorf("expected no deviation, got reasons: %v", result.Reasons)
	}
}

func TestAnalyzeBehaviorLoginHourDeviation(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	analyzer.BuildProfile(context.Background(), "user3", 8, "", "")

	ev := &core.Event{
		Type:      core.EventLogin,
		UserID:    "user3",
		Timestamp: time.Date(2025, 1, 1, 20, 0, 0, 0, time.UTC),
	}

	result, err := analyzer.AnalyzeBehavior(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasDeviation {
		t.Error("expected deviation for login hour change")
	}
}

func TestAnalyzeBehaviorIPDeviation(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	analyzer.BuildProfile(context.Background(), "user4", 12, "192.168.", "")

	ev := &core.Event{
		Type:      core.EventLogin,
		UserID:    "user4",
		IP:        "10.0.0.1",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	result, err := analyzer.AnalyzeBehavior(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasDeviation {
		t.Error("expected deviation for IP outside typical range")
	}
}

func TestAnalyzeBehaviorDeviceDeviation(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	analyzer.BuildProfile(context.Background(), "user5", 9, "", "device-old")

	ev := &core.Event{
		Type:     core.EventLogin,
		UserID:   "user5",
		DeviceID: "device-new",
		Timestamp: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC),
	}

	result, err := analyzer.AnalyzeBehavior(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasDeviation {
		t.Error("expected deviation for different device")
	}
}

func TestAnalyzeBehaviorUnknownUser(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	ev := &core.Event{
		Type:   core.EventLogin,
		UserID: "unknown",
	}

	result, err := analyzer.AnalyzeBehavior(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasDeviation {
		t.Error("expected no deviation for unknown user")
	}
}

func TestDetectDeviation(t *testing.T) {
	profileRepo := behavior.NewInMemoryProfileRepository()
	ruleRepo := behavior.NewInMemoryRuleRepository()
	analyzer := behavior.NewAnalyzer(profileRepo, ruleRepo)

	profile := &behavior.UserBehaviorProfile{
		UserID:           "u1",
		TypicalLoginHour: 10,
		TypicalIPRange:   "192.168.",
		TypicalDevice:    "device-1",
	}

	ev := &core.Event{
		Type:     core.EventLogin,
		UserID:   "u1",
		IP:       "10.0.0.1",
		DeviceID: "device-2",
		Timestamp: time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC),
	}

	result := analyzer.DetectDeviation(context.Background(), profile, ev)
	if !result.HasDeviation {
		t.Error("expected deviation detected")
	}
	if result.Score <= 0 {
		t.Errorf("expected positive score, got %f", result.Score)
	}
}
