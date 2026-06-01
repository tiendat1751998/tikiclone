package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/sms"
)

func TestSendSMS(t *testing.T) {
	repo := sms.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := sms.NewService(repo, pub, "+15005550006")
	ctx := context.Background()

	req := &sms.SendSMSRequest{
		To:   "+14085551234",
		Body: "Your OTP code is 123456",
	}

	msg, err := svc.SendSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}
	if msg.ID == "" {
		t.Error("expected SMS ID")
	}
	if msg.To != "+14085551234" {
		t.Errorf("expected to +14085551234, got %s", msg.To)
	}
	if msg.Status != "sent" {
		t.Errorf("expected status sent, got %s", msg.Status)
	}
}

func TestSendBulkSMS(t *testing.T) {
	repo := sms.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := sms.NewService(repo, pub, "+15005550006")
	ctx := context.Background()

	req := &sms.BulkSMSRequest{
		To:   []string{"+14085551234", "+14085551235"},
		Body: "Flash sale!",
	}

	results, err := svc.SendBulkSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendBulkSMS failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestVerifyPhone(t *testing.T) {
	repo := sms.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := sms.NewService(repo, pub, "+15005550006")
	ctx := context.Background()

	resp, err := svc.VerifyPhone(ctx, &sms.VerifyPhoneRequest{
		Phone: "+14085551234",
		Code:  "123456",
	})
	if err != nil {
		t.Fatalf("VerifyPhone failed: %v", err)
	}
	if !resp.Valid {
		t.Error("expected valid response for 6-digit code")
	}

	resp, err = svc.VerifyPhone(ctx, &sms.VerifyPhoneRequest{
		Phone: "+14085551234",
		Code:  "123",
	})
	if err != nil {
		t.Fatalf("VerifyPhone failed: %v", err)
	}
	if resp.Valid {
		t.Error("expected invalid for non-6-digit code")
	}
}
