package unit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/tracking"
)

type memTrackingRepo struct {
	mu     sync.Mutex
	events []*tracking.TrackingEvent
}

func newMemTrackingRepo() *memTrackingRepo {
	return &memTrackingRepo{}
}

func (r *memTrackingRepo) AppendEvent(_ context.Context, e *tracking.TrackingEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}

func (r *memTrackingRepo) GetTimeline(_ context.Context, shipmentID string) ([]*tracking.TrackingEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*tracking.TrackingEvent
	for _, e := range r.events {
		if e.ShipmentID == shipmentID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *memTrackingRepo) GetMilestones(_ context.Context, shipmentID string) ([]tracking.Milestone, error) {
	return nil, nil
}

func (r *memTrackingRepo) ListEvents(_ context.Context, filter tracking.TrackingFilter) ([]*tracking.TrackingEvent, int64, error) {
	return nil, 0, nil
}

func (r *memTrackingRepo) GetLastEvent(_ context.Context, shipmentID string) (*tracking.TrackingEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var last *tracking.TrackingEvent
	for _, e := range r.events {
		if e.ShipmentID == shipmentID {
			if last == nil || e.OccurredAt.After(last.OccurredAt) {
				last = e
			}
		}
	}
	return last, nil
}

func (r *memTrackingRepo) DeleteEvents(_ context.Context, shipmentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var kept []*tracking.TrackingEvent
	for _, e := range r.events {
		if e.ShipmentID != shipmentID {
			kept = append(kept, e)
		}
	}
	r.events = kept
	return nil
}

func TestTrackingEvent(t *testing.T) {
	svc := tracking.NewService(newMemTrackingRepo(), nil)
	e := &tracking.TrackingEvent{
		ID:         "evt-001",
		ShipmentID: "ship-001",
		EventType:  tracking.EventPickedUp,
		Description: "Package picked up from warehouse",
	}
	if err := svc.AppendEvent(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrackingTimeline(t *testing.T) {
	svc := tracking.NewService(newMemTrackingRepo(), nil)
	events := []*tracking.TrackingEvent{
		{ID: "e1", ShipmentID: "ship-002", EventType: tracking.EventPickedUp},
		{ID: "e2", ShipmentID: "ship-002", EventType: tracking.EventArrivedAtHub},
		{ID: "e3", ShipmentID: "ship-002", EventType: tracking.EventOutForDelivery},
	}
	for _, e := range events {
		if err := svc.AppendEvent(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	timeline, err := svc.GetTimeline(context.Background(), "ship-002")
	if err != nil {
		t.Fatal(err)
	}
	if len(timeline.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(timeline.Events))
	}
}

func TestTrackingLastEvent(t *testing.T) {
	svc := tracking.NewService(newMemTrackingRepo(), nil)
	svc.AppendEvent(context.Background(), &tracking.TrackingEvent{
		ID: "e1", ShipmentID: "ship-003", EventType: tracking.EventPickedUp,
		OccurredAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
	})
	svc.AppendEvent(context.Background(), &tracking.TrackingEvent{
		ID: "e2", ShipmentID: "ship-003", EventType: tracking.EventDelivered,
		OccurredAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	})
	last, err := svc.GetLastEvent(context.Background(), "ship-003")
	if err != nil {
		t.Fatal(err)
	}
	if last.EventType != tracking.EventDelivered {
		t.Errorf("expected delivered, got %v", last.EventType)
	}
}

func TestTrackingNoEvents(t *testing.T) {
	svc := tracking.NewService(newMemTrackingRepo(), nil)
	_, err := svc.GetLastEvent(context.Background(), "ship-none")
	if err == nil {
		t.Error("expected error for no events")
	}
}
