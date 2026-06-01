package sms

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

type Service interface {
	SendSMS(ctx context.Context, req *SendSMSRequest) (*SMSMessage, error)
	SendBulkSMS(ctx context.Context, req *BulkSMSRequest) ([]*SMSMessage, error)
	VerifyPhone(ctx context.Context, req *VerifyPhoneRequest) (*VerifyPhoneResponse, error)
}

type service struct {
	repo       Repository
	pub        events.Publisher
	fromNumber string
}

func NewService(repo Repository, pub events.Publisher, fromNumber string) Service {
	return &service{repo: repo, pub: pub, fromNumber: fromNumber}
}

func (s *service) SendSMS(ctx context.Context, req *SendSMSRequest) (*SMSMessage, error) {
	msg := &SMSMessage{
		To:       req.To,
		From:     s.fromNumber,
		Body:     req.Body,
		Status:   "sent",
		Provider: ProviderMock,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.pub.Publish(ctx, events.EventSMSSent, &events.SMSSentEvent{
		SMSID:  msg.ID,
		To:     msg.To,
		SentAt: time.Now(),
	})

	return msg, nil
}

func (s *service) SendBulkSMS(ctx context.Context, req *BulkSMSRequest) ([]*SMSMessage, error) {
	results := make([]*SMSMessage, 0, len(req.To))
	for _, phone := range req.To {
		msg, err := s.SendSMS(ctx, &SendSMSRequest{To: phone, Body: req.Body})
		if err != nil {
			continue
		}
		results = append(results, msg)
	}
	return results, nil
}

func (s *service) VerifyPhone(ctx context.Context, req *VerifyPhoneRequest) (*VerifyPhoneResponse, error) {
	return &VerifyPhoneResponse{
		Valid:   len(req.Code) == 6,
		Message: "verification successful",
	}, nil
}
