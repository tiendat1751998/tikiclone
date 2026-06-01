package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
	"github.com/tikiclone/tiki/platforms/analytics/internal/session"
)

func TestSessionTracking(t *testing.T) {
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	sessionRepo := session.NewInMemoryRepository()
	sessionSvc := session.NewService(sessionRepo, eventSvc, 30)

	now := time.Now()
	event := &events.AnalyticsEvent{
		EventID:   "sess-e1",
		EventType: events.EventPageview,
		UserID:    "user-sess-1",
		SessionID: "generated",
		Timestamp: now,
		Device:    "desktop",
		Source:    "web",
		Country:   "US",
	}

	s, err := sessionSvc.TrackSession(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil session")
	}
	if s.UserID != "user-sess-1" {
		t.Errorf("expected user-sess-1, got %s", s.UserID)
	}
	if !s.IsActive {
		t.Error("expected active session")
	}
	if s.Pageviews != 1 {
		t.Errorf("expected 1 pageview, got %d", s.Pageviews)
	}
}

func TestSessionTimeout(t *testing.T) {
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	sessionRepo := session.NewInMemoryRepository()
	sessionSvc := session.NewService(sessionRepo, eventSvc, 1)

	now := time.Now()
	event1 := &events.AnalyticsEvent{
		EventID: "st-e1", EventType: events.EventPageview, UserID: "user-timeout",
		Timestamp: now.Add(-2 * time.Minute), Device: "mobile",
	}
	sessionSvc.TrackSession(context.Background(), event1)

	event2 := &events.AnalyticsEvent{
		EventID: "st-e2", EventType: events.EventPageview, UserID: "user-timeout",
		Timestamp: now, Device: "mobile",
	}
	s2, err := sessionSvc.TrackSession(context.Background(), event2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s2 == nil {
		t.Fatal("expected new session after timeout")
	}
}

func TestSessionMetrics(t *testing.T) {
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	sessionRepo := session.NewInMemoryRepository()
	sessionSvc := session.NewService(sessionRepo, eventSvc, 30)

	now := time.Now()
	events_list := []*events.AnalyticsEvent{
		{EventID: "sm1", EventType: events.EventPageview, UserID: "u1", Timestamp: now, Source: "web"},
		{EventID: "sm2", EventType: events.EventPurchase, UserID: "u1", Timestamp: now, Revenue: 100, Source: "web"},
		{EventID: "sm3", EventType: events.EventPageview, UserID: "u2", Timestamp: now, Source: "mobile"},
	}

	for _, e := range events_list {
		sessionSvc.TrackSession(context.Background(), e)
	}

	filter := &session.SessionFilter{}
	metrics, err := sessionSvc.CalculateSessionMetrics(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics == nil {
		t.Fatal("expected metrics")
	}
	if metrics.TotalSessions != 2 {
		t.Errorf("expected 2 sessions, got %d", metrics.TotalSessions)
	}
}

func TestSessionEmptyData(t *testing.T) {
	sessionRepo := session.NewInMemoryRepository()
	sessionSvc := session.NewService(sessionRepo, events.NewService(events.NewInMemoryRepository()), 30)

	sessions, total, err := sessionSvc.GetSessions(context.Background(), &session.SessionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 total, got %d", total)
	}
	if len(sessions) != 0 {
		t.Errorf("expected empty sessions, got %d", len(sessions))
	}
}
