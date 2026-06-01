package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

func TestEventIngestion(t *testing.T) {
	repo := events.NewInMemoryRepository()
	svc := events.NewService(repo)

	event := &events.AnalyticsEvent{
		EventID:   "evt-1",
		EventType: events.EventPageview,
		UserID:    "user1",
		SessionID: "session1",
		Timestamp: time.Now(),
		Source:    "web",
		Device:    "desktop",
		Country:   "US",
	}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, err := svc.GetEvent(context.Background(), "evt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored == nil {
		t.Fatal("expected stored event")
	}
	if stored.EventType != events.EventPageview {
		t.Errorf("expected pageview, got %s", stored.EventType)
	}
}

func TestEventDeduplication(t *testing.T) {
	repo := events.NewInMemoryRepository()
	svc := events.NewService(repo)

	event := &events.AnalyticsEvent{
		EventID:   "evt-dedup",
		EventType: events.EventPageview,
		UserID:    "user1",
		Timestamp: time.Now(),
	}

	err := svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.IngestEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error on duplicate: %v", err)
	}

	count, _, _ := repo.ListEvents(context.Background(), events.EventPageview, time.Time{}, time.Now(), 0, 100)
	if count != nil && len(count) != 1 {
		t.Errorf("expected 1 event after dedup, got %d", len(count))
	}
}

func TestBatchIngestion(t *testing.T) {
	repo := events.NewInMemoryRepository()
	svc := events.NewService(repo)

	events_list := []events.AnalyticsEvent{
		{EventID: "b1", EventType: events.EventPageview, UserID: "u1", Timestamp: time.Now()},
		{EventID: "b2", EventType: events.EventPageview, UserID: "u2", Timestamp: time.Now()},
		{EventID: "b3", EventType: events.EventPurchase, UserID: "u1", Revenue: 100, Timestamp: time.Now()},
	}

	ingested, err := svc.BatchIngest(context.Background(), events_list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ingested != 3 {
		t.Errorf("expected 3 ingested, got %d", ingested)
	}
}

func TestEventEmptyData(t *testing.T) {
	svc := events.NewService(events.NewInMemoryRepository())

	event, err := svc.GetEvent(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != nil {
		t.Error("expected nil for non-existent event")
	}
}

func TestEventCount(t *testing.T) {
	repo := events.NewInMemoryRepository()
	svc := events.NewService(repo)

	now := time.Now()
	for i := 0; i < 5; i++ {
		svc.IngestEvent(context.Background(), &events.AnalyticsEvent{
			EventID: "cnt-" + string(rune('0'+i)), EventType: events.EventPageview, UserID: "u1", Timestamp: now,
		})
	}
	for i := 0; i < 3; i++ {
		svc.IngestEvent(context.Background(), &events.AnalyticsEvent{
			EventID: "cnt-p-" + string(rune('0'+i)), EventType: events.EventPurchase, UserID: "u1", Revenue: 10, Timestamp: now,
		})
	}

	count, _ := svc.GetEventCount(context.Background(), events.EventPageview, now.Add(-time.Hour), now.Add(time.Hour))
	if count != 5 {
		t.Errorf("expected 5 pageviews, got %d", count)
	}

	revenue, _ := svc.GetRevenue(context.Background(), now.Add(-time.Hour), now.Add(time.Hour))
	if revenue != 30 {
		t.Errorf("expected 30 revenue, got %f", revenue)
	}
}
