package dispatch

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Repository interface {
	Create(ctx context.Context, d *Dispatch) error
	GetByID(ctx context.Context, id string) (*Dispatch, error)
	GetByShipment(ctx context.Context, shipmentID string) (*Dispatch, error)
	Update(ctx context.Context, d *Dispatch) error
	List(ctx context.Context, filter DispatchFilter) ([]*Dispatch, int64, error)
}

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) CreateDispatch(ctx context.Context, d *Dispatch) error {
	d.Status = DispatchPending
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = d.CreatedAt
	if err := s.repo.Create(ctx, d); err != nil {
		return err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.DispatchCreated,
			Source:    "logistics.dispatch",
			Payload:   d,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) AssignCourier(ctx context.Context, dispatchID, courierID string) error {
	d, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	if d.Status != DispatchPending {
		return ErrDispatchAlreadyAssigned
	}
	d.CourierID = courierID
	d.Status = DispatchAssigned
	d.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, d)
}

func (s *Service) MarkEnRoute(ctx context.Context, dispatchID string) error {
	d, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	d.Status = DispatchEnRoute
	d.DispatchTime = timePtr(time.Now().UTC())
	d.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, d)
}

func (s *Service) MarkCompleted(ctx context.Context, dispatchID string) error {
	d, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	d.Status = DispatchCompleted
	d.CompletedAt = timePtr(time.Now().UTC())
	d.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, d)
}

func (s *Service) GetByShipment(ctx context.Context, shipmentID string) (*Dispatch, error) {
	return s.repo.GetByShipment(ctx, shipmentID)
}

func timePtr(t time.Time) *time.Time { return &t }
