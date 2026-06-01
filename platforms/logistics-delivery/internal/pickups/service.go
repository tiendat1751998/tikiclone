package pickups

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Repository interface {
	Create(ctx context.Context, p *Pickup) error
	GetByID(ctx context.Context, id string) (*Pickup, error)
	GetByShipment(ctx context.Context, shipmentID string) (*Pickup, error)
	Update(ctx context.Context, p *Pickup) error
	MarkCompleted(ctx context.Context, id string, pickedUpAt time.Time) error
	MarkFailed(ctx context.Context, id string, reason string) error
}

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) Create(ctx context.Context, p *Pickup) error {
	p.Status = PickupScheduled
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	return s.repo.Create(ctx, p)
}

func (s *Service) MarkCompleted(ctx context.Context, id string) error {
	now := time.Now().UTC()
	if err := s.repo.MarkCompleted(ctx, id, now); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.PickupCompleted,
			Source:    "logistics.pickups",
			Payload:   map[string]string{"pickup_id": id},
			Timestamp: now,
		})
	}
	return nil
}

func (s *Service) MarkFailed(ctx context.Context, id string, reason string) error {
	if err := s.repo.MarkFailed(ctx, id, reason); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.PickupFailed,
			Source:    "logistics.pickups",
			Payload:   map[string]string{"pickup_id": id, "reason": reason},
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) GetByShipment(ctx context.Context, shipmentID string) (*Pickup, error) {
	return s.repo.GetByShipment(ctx, shipmentID)
}
