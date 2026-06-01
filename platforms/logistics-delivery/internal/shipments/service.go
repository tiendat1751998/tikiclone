package shipments

import (
	"context"
	"fmt"
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

func (s *Service) Create(ctx context.Context, sh *Shipment) error {
	if sh.ID == "" {
		return ErrInvalidShipmentData
	}
	if sh.Status == "" {
		sh.Status = StatusPending
	}
	sh.Version = 1
	sh.CreatedAt = time.Now().UTC()
	sh.UpdatedAt = sh.CreatedAt
	if err := s.repo.Create(ctx, sh); err != nil {
		return fmt.Errorf("create shipment: %w", err)
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.ShipmentCreated,
			Source:    "logistics.shipments",
			Payload:   sh,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Shipment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ShipmentFilter) ([]*Shipment, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) TransitionStatus(ctx context.Context, shipmentID string, toStatus ShipmentStatus, reason, replayID string) error {
	sh, err := s.repo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}
	if !IsValidTransition(sh.Status, toStatus) {
		return ErrInvalidStatusTransition
	}
	if err := s.repo.TransitionStatus(ctx, shipmentID, sh.Status, toStatus, replayID); err != nil {
		return err
	}
	if toStatus == StatusDelivered {
		now := time.Now().UTC()
		if err := s.repo.Update(ctx, &Shipment{ID: shipmentID, ActualDeliveredAt: &now}); err != nil {
			return err
		}
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:   events.ShipmentStatusChanged,
			Source: "logistics.shipments",
			Payload: StatusTransition{
				ShipmentID:     shipmentID,
				FromStatus:     sh.Status,
				ToStatus:       toStatus,
				Reason:         reason,
				ReplayID:       replayID,
				TransitionedAt: time.Now().UTC(),
			},
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) Update(ctx context.Context, sh *Shipment) error {
	existing, err := s.repo.GetByID(ctx, sh.ID)
	if err != nil {
		return err
	}
	sh.Version = existing.Version
	sh.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, sh)
}
