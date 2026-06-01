package push

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

type Service interface {
	RegisterDevice(ctx context.Context, userID string, token string, platform Platform) (*PushDevice, error)
	SendPush(ctx context.Context, req *PushNotificationRequest) (*PushResult, error)
	SendBulkPush(ctx context.Context, req *BulkPushRequest) ([]*PushResult, error)
	GetDevices(ctx context.Context, userID string) ([]*PushDevice, error)
}

type service struct {
	repo Repository
	pub  events.Publisher
}

func NewService(repo Repository, pub events.Publisher) Service {
	return &service{repo: repo, pub: pub}
}

func (s *service) RegisterDevice(ctx context.Context, userID string, token string, platform Platform) (*PushDevice, error) {
	device := &PushDevice{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}

	if err := s.repo.RegisterDevice(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *service) SendPush(ctx context.Context, req *PushNotificationRequest) (*PushResult, error) {
	devices, err := s.repo.ListActiveDevicesByUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if len(devices) == 0 {
		return &PushResult{Success: false, Error: "no active devices"}, nil
	}

	device := devices[0]
	result := &PushResult{
		Success: true,
		Token:   device.Token,
	}

	if result.Success {
		s.pub.Publish(ctx, events.EventNotificationSent, &events.NotificationSentEvent{
			NotificationID: "",
			UserID:         req.UserID,
			Channel:        "push",
			Type:           "push_notification",
			SentAt:         time.Now(),
		})
	} else {
		s.repo.MarkDeviceInactive(ctx, device.ID)
		s.pub.Publish(ctx, events.EventPushFailed, &events.PushFailedEvent{
			UserID:      req.UserID,
			DeviceToken: device.Token,
			Platform:    string(device.Platform),
			Error:       result.Error,
			FailedAt:    time.Now(),
		})
	}

	return result, nil
}

func (s *service) SendBulkPush(ctx context.Context, req *BulkPushRequest) ([]*PushResult, error) {
	var results []*PushResult
	for _, userID := range req.UserIDs {
		subReq := &PushNotificationRequest{
			UserID: userID,
			Title:  req.Title,
			Body:   req.Body,
			Data:   req.Data,
		}
		result, err := s.SendPush(ctx, subReq)
		if err != nil {
			results = append(results, &PushResult{Success: false, Token: "", Error: err.Error()})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *service) GetDevices(ctx context.Context, userID string) ([]*PushDevice, error) {
	return s.repo.ListDevicesByUser(ctx, userID)
}
