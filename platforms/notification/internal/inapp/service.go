package inapp

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

type Service interface {
	SendInApp(ctx context.Context, req *SendInAppRequest) (*InAppNotification, error)
	GetFeed(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context, userID string) error
	Dismiss(ctx context.Context, id string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type service struct {
	repo Repository
	pub  events.Publisher
}

func NewService(repo Repository, pub events.Publisher) Service {
	return &service{repo: repo, pub: pub}
}

func (s *service) SendInApp(ctx context.Context, req *SendInAppRequest) (*InAppNotification, error) {
	n := &InAppNotification{
		UserID:   req.UserID,
		Category: req.Category,
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Action:   req.Action,
		Read:     false,
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	s.pub.Publish(ctx, events.EventInAppSent, &events.InAppSentEvent{
		NotificationID: n.ID,
		UserID:         n.UserID,
		Category:       string(n.Category),
		SentAt:         time.Now(),
	})

	return n, nil
}

func (s *service) GetFeed(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *service) MarkRead(ctx context.Context, id string) error {
	return s.repo.MarkRead(ctx, id)
}

func (s *service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *service) Dismiss(ctx context.Context, id string) error {
	return s.repo.Dismiss(ctx, id)
}

func (s *service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}
