package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/push"
)

func TestRegisterDevice(t *testing.T) {
	repo := push.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := push.NewService(repo, pub)
	ctx := context.Background()

	device, err := svc.RegisterDevice(ctx, "user1", "token-abc-123", push.PlatformIOS)
	if err != nil {
		t.Fatalf("RegisterDevice failed: %v", err)
	}
	if device.ID == "" {
		t.Error("expected device ID")
	}
	if !device.Active {
		t.Error("expected device to be active")
	}
}

func TestSendPush(t *testing.T) {
	repo := push.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := push.NewService(repo, pub)
	ctx := context.Background()

	svc.RegisterDevice(ctx, "user1", "token-abc", push.PlatformAndroid)

	req := &push.PushNotificationRequest{
		UserID: "user1",
		Title:  "Test Push",
		Body:   "This is a test push",
		Data:   map[string]string{"order_id": "123"},
	}

	result, err := svc.SendPush(ctx, req)
	if err != nil {
		t.Fatalf("SendPush failed: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestSendPushNoDevices(t *testing.T) {
	repo := push.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := push.NewService(repo, pub)
	ctx := context.Background()

	req := &push.PushNotificationRequest{
		UserID: "user-no-devices",
		Title:  "Test",
		Body:   "Body",
	}

	result, err := svc.SendPush(ctx, req)
	if err != nil {
		t.Fatalf("SendPush failed: %v", err)
	}
	if result.Success {
		t.Error("expected failure with no devices")
	}
}

func TestBulkPush(t *testing.T) {
	repo := push.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := push.NewService(repo, pub)
	ctx := context.Background()

	svc.RegisterDevice(ctx, "user1", "token1", push.PlatformIOS)
	svc.RegisterDevice(ctx, "user2", "token2", push.PlatformAndroid)

	req := &push.BulkPushRequest{
		UserIDs: []string{"user1", "user2"},
		Title:   "Bulk Push",
		Body:    "Bulk body",
	}

	results, err := svc.SendBulkPush(ctx, req)
	if err != nil {
		t.Fatalf("SendBulkPush failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected all successes, got error: %s", r.Error)
		}
	}
}

func TestGetDevices(t *testing.T) {
	repo := push.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := push.NewService(repo, pub)
	ctx := context.Background()

	svc.RegisterDevice(ctx, "user1", "token1", push.PlatformIOS)
	svc.RegisterDevice(ctx, "user1", "token2", push.PlatformAndroid)

	devices, err := svc.GetDevices(ctx, "user1")
	if err != nil {
		t.Fatalf("GetDevices failed: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}
