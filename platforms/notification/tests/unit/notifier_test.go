package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
)

func TestSendNotification(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	req := &notifier.SendNotificationRequest{
		UserID:  "user1",
		Type:    notifier.TypeOrderConfirmation,
		Channel: notifier.ChannelInApp,
		Title:   "Order Confirmed",
		Body:    "Your order #123 has been confirmed",
	}

	n, err := svc.Send(ctx, req)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if n.ID == "" {
		t.Error("expected notification ID to be set")
	}
	if n.UserID != "user1" {
		t.Errorf("expected user1, got %s", n.UserID)
	}
	if n.Status != notifier.StatusSent {
		t.Errorf("expected status sent, got %s", n.Status)
	}
}

func TestGetNotifications(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.Send(ctx, &notifier.SendNotificationRequest{
			UserID:  "user1",
			Type:    notifier.TypeOrderConfirmation,
			Channel: notifier.ChannelInApp,
			Title:   "Notification",
			Body:    "Body",
		})
	}

	notifications, err := svc.GetNotifications(ctx, "user1", 10, 0)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}
	if len(notifications) != 5 {
		t.Errorf("expected 5 notifications, got %d", len(notifications))
	}
}

func TestMarkRead(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	n, _ := svc.Send(ctx, &notifier.SendNotificationRequest{
		UserID:  "user1",
		Type:    notifier.TypeOrderConfirmation,
		Channel: notifier.ChannelInApp,
		Title:   "Test",
		Body:    "Body",
	})

	if err := svc.MarkRead(ctx, n.ID); err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}

	updated, _ := svc.GetByID(ctx, n.ID)
	if updated.Status != notifier.StatusRead {
		t.Errorf("expected status read, got %s", updated.Status)
	}
	if updated.ReadAt == nil {
		t.Error("expected ReadAt to be set")
	}
}

func TestDeleteNotification(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	n, _ := svc.Send(ctx, &notifier.SendNotificationRequest{
		UserID:  "user1",
		Type:    notifier.TypeOrderConfirmation,
		Channel: notifier.ChannelInApp,
		Title:   "Test",
		Body:    "Body",
	})

	if err := svc.Delete(ctx, n.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := svc.GetByID(ctx, n.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestGetUnreadCount(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		svc.Send(ctx, &notifier.SendNotificationRequest{
			UserID:  "user1",
			Type:    notifier.TypeOrderConfirmation,
			Channel: notifier.ChannelInApp,
			Title:   "Test",
			Body:    "Body",
		})
	}

	count, err := svc.GetUnreadCount(ctx, "user1")
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 unread, got %d", count)
	}
}

func TestBatchSend(t *testing.T) {
	repo := notifier.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := notifier.NewService(repo, pub)
	ctx := context.Background()

	reqs := []*notifier.SendNotificationRequest{
		{UserID: "user1", Type: notifier.TypeOrderConfirmation, Channel: notifier.ChannelInApp, Title: "1", Body: "1"},
		{UserID: "user1", Type: notifier.TypeShipmentUpdate, Channel: notifier.ChannelInApp, Title: "2", Body: "2"},
		{UserID: "user2", Type: notifier.TypePromotion, Channel: notifier.ChannelEmail, Title: "3", Body: "3"},
	}

	results, err := svc.BatchSend(ctx, reqs)
	if err != nil {
		t.Fatalf("BatchSend failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
