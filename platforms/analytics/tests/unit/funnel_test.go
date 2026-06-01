package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
	"github.com/tikiclone/tiki/platforms/analytics/internal/funnel"
)

func setupFunnelTest(t *testing.T) (*funnel.Service, *events.Service) {
	t.Helper()
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	funnelRepo := funnel.NewInMemoryRepository()
	funnelSvc := funnel.NewService(funnelRepo, eventSvc)

	now := time.Now()
	events_list := []events.AnalyticsEvent{
		{EventID: "f1", EventType: events.EventPageview, UserID: "u1", Timestamp: now},
		{EventID: "f2", EventType: events.EventAddToCart, UserID: "u1", Timestamp: now},
		{EventID: "f3", EventType: events.EventCheckout, UserID: "u1", Timestamp: now},
		{EventID: "f4", EventType: events.EventPurchase, UserID: "u1", Timestamp: now},
		{EventID: "f5", EventType: events.EventPageview, UserID: "u2", Timestamp: now},
		{EventID: "f6", EventType: events.EventAddToCart, UserID: "u2", Timestamp: now},
		{EventID: "f7", EventType: events.EventCheckout, UserID: "u2", Timestamp: now},
		{EventID: "f8", EventType: events.EventPageview, UserID: "u3", Timestamp: now},
	}
	for i := range events_list {
		eventSvc.IngestEvent(context.Background(), &events_list[i])
	}

	return funnelSvc, eventSvc
}

func TestFunnelConversion(t *testing.T) {
	svc, _ := setupFunnelTest(t)

	def := &funnel.FunnelDefinition{
		Name: "purchase_funnel",
		Steps: []funnel.FunnelStep{
			{Name: "Visit", EventType: "pageview", Order: 1},
			{Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
			{Name: "Checkout", EventType: "checkout", Order: 3},
			{Name: "Purchase", EventType: "purchase", Order: 4},
		},
	}

	result, err := svc.BuildFunnel(context.Background(), def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(result.Steps))
	}
	if result.StartCount == 0 {
		t.Error("expected non-zero start count")
	}
}

func TestFunnelDropoff(t *testing.T) {
	svc, _ := setupFunnelTest(t)

	def := &funnel.FunnelDefinition{
		Name: "purchase_funnel",
		Steps: []funnel.FunnelStep{
			{Name: "Visit", EventType: "pageview", Order: 1},
			{Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
			{Name: "Purchase", EventType: "purchase", Order: 3},
		},
	}

	result, err := svc.BuildFunnel(context.Background(), def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dropoffSteps, err := svc.AnalyzeDropoff(context.Background(), result.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dropoffSteps) == 0 {
		t.Error("expected dropoff analysis")
	}
}

func TestFunnelInvalidDefinition(t *testing.T) {
	svc, _ := setupFunnelTest(t)

	_, err := svc.BuildFunnel(context.Background(), &funnel.FunnelDefinition{
		Name:  "invalid",
		Steps: []funnel.FunnelStep{{Name: "Only Step", EventType: "pageview", Order: 1}},
	})
	if err == nil {
		t.Fatal("expected error for funnel with < 2 steps")
	}
}

func TestFunnelEmptyData(t *testing.T) {
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	funnelRepo := funnel.NewInMemoryRepository()
	svc := funnel.NewService(funnelRepo, eventSvc)

	def := &funnel.FunnelDefinition{
		Name: "empty_funnel",
		Steps: []funnel.FunnelStep{
			{Name: "Step 1", EventType: "pageview", Order: 1},
			{Name: "Step 2", EventType: "purchase", Order: 2},
		},
	}

	result, err := svc.BuildFunnel(context.Background(), def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StartCount != 0 {
		t.Errorf("expected 0 start count, got %d", result.StartCount)
	}
}
