package notifier

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

type Service interface {
	Send(ctx context.Context, req *SendNotificationRequest) (*Notification, error)
	BatchSend(ctx context.Context, reqs []*SendNotificationRequest) ([]*Notification, error)
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*Notification, error)
	GetByID(ctx context.Context, id string) (*Notification, error)
	MarkRead(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type service struct {
	repo    Repository
	pub     events.Publisher
}

func NewService(repo Repository, pub events.Publisher) Service {
	return &service{repo: repo, pub: pub}
}

func (s *service) Send(ctx context.Context, req *SendNotificationRequest) (*Notification, error) {
	if req.Channel == "" {
		req.Channel = ChannelInApp
	}
	if req.Priority == 0 {
		req.Priority = PriorityNormal
	}

	n := &Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Channel:  req.Channel,
		Title:    req.Title,
		Body:     req.Body,
		Data:     req.Data,
		Priority: req.Priority,
		Status:   StatusSent,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	s.pub.Publish(ctx, events.EventNotificationSent, &events.NotificationSentEvent{
		NotificationID: n.ID,
		UserID:         n.UserID,
		Channel:        string(n.Channel),
		Type:           string(n.Type),
		SentAt:         time.Now(),
	})

	return n, nil
}

func (s *service) BatchSend(ctx context.Context, reqs []*SendNotificationRequest) ([]*Notification, error) {
	results := make([]*Notification, 0, len(reqs))
	for _, req := range reqs {
		n, err := s.Send(ctx, req)
		if err != nil {
			continue
		}
		results = append(results, n)
	}
	return results, nil
}

func (s *service) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*Notification, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *service) GetByID(ctx context.Context, id string) (*Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) MarkRead(ctx context.Context, id string) error {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.MarkRead(ctx, id); err != nil {
		return err
	}

	s.pub.Publish(ctx, events.EventNotificationRead, &events.NotificationReadEvent{
		NotificationID: n.ID,
		UserID:         n.UserID,
		ReadAt:         time.Now(),
	})

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}
