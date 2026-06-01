package couriers

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Repository interface {
	Create(ctx context.Context, c *Courier) error
	GetByID(ctx context.Context, id string) (*Courier, error)
	Update(ctx context.Context, c *Courier) error
	ListAvailable(ctx context.Context, zoneID string) ([]*Courier, error)
	List(ctx context.Context, offset, limit int) ([]*Courier, int64, error)
	UpdateLocation(ctx context.Context, courierID string, lat, lng float64) error
}

type WebhookStore interface {
	IsDuplicate(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type Service struct {
	repo        Repository
	webhookRepo WebhookStore
	producer    events.Producer
}

func NewService(repo Repository, webhookRepo WebhookStore, producer events.Producer) *Service {
	return &Service{repo: repo, webhookRepo: webhookRepo, producer: producer}
}

func (s *Service) Create(ctx context.Context, c *Courier) error {
	c.CreatedAt = time.Now().UTC()
	c.UpdatedAt = c.CreatedAt
	return s.repo.Create(ctx, c)
}

func (s *Service) UpdateLocation(ctx context.Context, courierID string, lat, lng float64) error {
	return s.repo.UpdateLocation(ctx, courierID, lat, lng)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Courier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) FindAvailable(ctx context.Context, zoneID string) (*Courier, error) {
	couriers, err := s.repo.ListAvailable(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	for _, c := range couriers {
		if c.CurrentLoad < c.MaxCapacity {
			return c, nil
		}
	}
	return nil, ErrCourierNotAvailable
}

func (s *Service) ProcessWebhook(ctx context.Context, payload *WebhookPayload) error {
	if payload.ReceivedAt.IsZero() {
		payload.ReceivedAt = time.Now().UTC()
	}
	dup, err := s.webhookRepo.IsDuplicate(ctx, payload.CourierID+"-"+payload.EventType+"-"+payload.ShipmentID)
	if err != nil {
		return err
	}
	if dup {
		return ErrDuplicateWebhookEvent
	}
	if payload.Data != nil {
		if lat, ok := payload.Data["latitude"].(float64); ok {
			if lng, ok := payload.Data["longitude"].(float64); ok {
				_ = s.repo.UpdateLocation(ctx, payload.CourierID, lat, lng)
			}
		}
		if status, ok := payload.Data["status"].(string); ok {
			c, err := s.repo.GetByID(ctx, payload.CourierID)
			if err == nil {
				c.Status = CourierStatus(status)
				c.UpdatedAt = time.Now().UTC()
				_ = s.repo.Update(ctx, c)
			}
		}
	}
	_ = s.webhookRepo.MarkProcessed(ctx, payload.CourierID+"-"+payload.EventType+"-"+payload.ShipmentID)
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.CourierWebhookReceived,
			Source:    "logistics.couriers",
			Payload:   payload,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}
