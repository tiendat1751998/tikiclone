package fulfillment

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Repository interface {
	Create(ctx context.Context, f *Fulfillment) error
	GetByID(ctx context.Context, id string) (*Fulfillment, error)
	GetByShipment(ctx context.Context, shipmentID string) (*Fulfillment, error)
	Update(ctx context.Context, f *Fulfillment) error
	MarkPacked(ctx context.Context, id string) error
	MarkShipped(ctx context.Context, id string) error
	MarkCompleted(ctx context.Context, id string) error
}

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) Create(ctx context.Context, f *Fulfillment) error {
	f.Status = FulfillmentPending
	f.CreatedAt = time.Now().UTC()
	f.UpdatedAt = f.CreatedAt
	return s.repo.Create(ctx, f)
}

func (s *Service) MarkPacked(ctx context.Context, id string) error {
	if err := s.repo.MarkPacked(ctx, id); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.FulfillmentPacked,
			Source:    "logistics.fulfillment",
			Payload:   map[string]string{"fulfillment_id": id},
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) MarkShipped(ctx context.Context, id string) error {
	if err := s.repo.MarkShipped(ctx, id); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.FulfillmentShipped,
			Source:    "logistics.fulfillment",
			Payload:   map[string]string{"fulfillment_id": id},
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) GetByShipment(ctx context.Context, shipmentID string) (*Fulfillment, error) {
	return s.repo.GetByShipment(ctx, shipmentID)
}
