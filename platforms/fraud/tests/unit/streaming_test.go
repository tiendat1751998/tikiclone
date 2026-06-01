package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/streaming"
)

func TestProcessEvent(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	event := &core.FraudEvent{
		UserID: "user1", IP: "192.168.1.1", DeviceID: "dev1",
		Timestamp: time.Now(),
	}

	svc.ProcessEvent(context.Background(), event)

	count := svc.CountInWindow(context.Background(), "user", "user1", 1*time.Hour)
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

func TestCountInWindow(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	for i := 0; i < 5; i++ {
		svc.ProcessEvent(context.Background(), &core.FraudEvent{
			UserID: "user-w", Timestamp: time.Now(),
		})
	}

	count1m := svc.CountInWindow(context.Background(), "user", "user-w", 1*time.Minute)
	if count1m != 5 {
		t.Errorf("expected 5 events in 1min window, got %d", count1m)
	}

	time.Sleep(2 * time.Millisecond)

	count1h := svc.CountInWindow(context.Background(), "user", "user-w", 1*time.Hour)
	if count1h != 5 {
		t.Errorf("expected 5 events in 1hr window, got %d", count1h)
	}
}

func TestDetectBurst(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	for i := 0; i < 50; i++ {
		svc.ProcessEvent(context.Background(), &core.FraudEvent{
			UserID: "burst-user", Timestamp: time.Now(),
		})
	}

	window := svc.DetectBurst(context.Background(), "user", "burst-user")
	if window.Count < 50 {
		t.Errorf("expected count >= 50, got %d", window.Count)
	}
}

func TestDetectBurstNotTriggered(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	svc.ProcessEvent(context.Background(), &core.FraudEvent{
		UserID: "normal-user", Timestamp: time.Now(),
	})

	window := svc.DetectBurst(context.Background(), "user", "normal-user")
	if window.IsBurst {
		t.Error("expected no burst for single event")
	}
}

func TestGetAggregated(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	svc.ProcessEvent(context.Background(), &core.FraudEvent{
		UserID: "agg-user", IP: "10.0.0.1", Timestamp: time.Now(),
	})

	agg := svc.GetAggregated(context.Background(), "user", "agg-user")
	if agg.EntityID != "agg-user" {
		t.Errorf("expected agg-user, got %s", agg.EntityID)
	}
	if agg.Window1Min < 1 {
		t.Errorf("expected >= 1 in 1min window, got %d", agg.Window1Min)
	}
}

func TestMultipleEntities(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	svc.ProcessEvent(context.Background(), &core.FraudEvent{
		UserID: "u1", IP: "10.0.0.1", DeviceID: "d1", Timestamp: time.Now(),
	})
	svc.ProcessEvent(context.Background(), &core.FraudEvent{
		UserID: "u2", IP: "10.0.0.2", DeviceID: "d2", Timestamp: time.Now(),
	})

	if c := svc.CountInWindow(context.Background(), "user", "u1", 1*time.Hour); c != 1 {
		t.Errorf("expected 1 for u1, got %d", c)
	}
	if c := svc.CountInWindow(context.Background(), "user", "u2", 1*time.Hour); c != 1 {
		t.Errorf("expected 1 for u2, got %d", c)
	}
	if c := svc.CountInWindow(context.Background(), "ip", "10.0.0.1", 1*time.Hour); c != 1 {
		t.Errorf("expected 1 for ip, got %d", c)
	}
	if c := svc.CountInWindow(context.Background(), "device", "d1", 1*time.Hour); c != 1 {
		t.Errorf("expected 1 for device, got %d", c)
	}
}

func TestBurstThreshold(t *testing.T) {
	repo := streaming.NewInMemoryRepository()
	svc := streaming.NewService(repo)

	for i := 0; i < 100; i++ {
		svc.ProcessEvent(context.Background(), &core.FraudEvent{
			UserID: "burst-threshold", Timestamp: time.Now(),
		})
	}

	window := svc.DetectBurst(context.Background(), "user", "burst-threshold")
	if !window.IsBurst {
		t.Error("expected burst detection for 100 events")
	}
	if window.Threshold <= 0 {
		t.Error("expected positive threshold")
	}
}
