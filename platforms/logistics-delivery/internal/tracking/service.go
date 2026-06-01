package tracking

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) AppendEvent(ctx context.Context, e *TrackingEvent) error {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	e.CreatedAt = time.Now().UTC()
	if err := s.repo.AppendEvent(ctx, e); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.TrackingUpdated,
			Source:    "logistics.tracking",
			Payload:   e,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) GetTimeline(ctx context.Context, shipmentID string) (*TrackingTimeline, error) {
	events, err := s.repo.GetTimeline(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return &TrackingTimeline{
		ShipmentID: shipmentID,
		Events:     events,
		Milestones: milestones,
	}, nil
}

func (s *Service) GetLastEvent(ctx context.Context, shipmentID string) (*TrackingEvent, error) {
	event, err := s.repo.GetLastEvent(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrTrackingEventNotFound
	}
	return event, nil
}

func (s *Service) ListEvents(ctx context.Context, filter TrackingFilter) ([]*TrackingEvent, int64, error) {
	return s.repo.ListEvents(ctx, filter)
}
