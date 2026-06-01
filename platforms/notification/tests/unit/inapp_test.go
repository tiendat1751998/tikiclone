package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/inapp"
)

func TestSendInApp(t *testing.T) {
	repo := inapp.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := inapp.NewService(repo, pub)
	ctx := context.Background()

	req := &inapp.SendInAppRequest{
		UserID:   "user1",
		Category: inapp.CategoryOrder,
		Title:    "Order Shipped",
		Body:     "Your order has been shipped!",
	}

	n, err := svc.SendInApp(ctx, req)
	if err != nil {
		t.Fatalf("SendInApp failed: %v", err)
	}
	if n.ID == "" {
		t.Error("expected notification ID")
	}
	if n.Read {
		t.Error("new notification should not be read")
	}
}

func TestGetFeed(t *testing.T) {
	repo := inapp.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := inapp.NewService(repo, pub)
	ctx := context.Background()

	svc.SendInApp(ctx, &inapp.SendInAppRequest{
		UserID:   "user1",
		Category: inapp.CategoryOrder,
		Title:    "Test",
		Body:     "Body",
	})

	feed, err := svc.GetFeed(ctx, "user1", 10, 0)
	if err != nil {
		t.Fatalf("GetFeed failed: %v", err)
	}
	if len(feed) != 1 {
		t.Errorf("expected 1 notification, got %d", len(feed))
	}
}

func TestMarkAllRead(t *testing.T) {
	repo := inapp.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := inapp.NewService(repo, pub)
	ctx := context.Background()

	svc.SendInApp(ctx, &inapp.SendInAppRequest{UserID: "user1", Category: inapp.CategoryOrder, Title: "1", Body: "1"})
	svc.SendInApp(ctx, &inapp.SendInAppRequest{UserID: "user1", Category: inapp.CategoryPayment, Title: "2", Body: "2"})

	if err := svc.MarkAllRead(ctx, "user1"); err != nil {
		t.Fatalf("MarkAllRead failed: %v", err)
	}

	count, _ := svc.GetUnreadCount(ctx, "user1")
	if count != 0 {
		t.Errorf("expected 0 unread after mark all read, got %d", count)
	}
}

func TestDismiss(t *testing.T) {
	repo := inapp.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := inapp.NewService(repo, pub)
	ctx := context.Background()

	n, _ := svc.SendInApp(ctx, &inapp.SendInAppRequest{
		UserID:   "user1",
		Category: inapp.CategorySystem,
		Title:    "Dismiss me",
		Body:     "Body",
	})

	if err := svc.Dismiss(ctx, n.ID); err != nil {
		t.Fatalf("Dismiss failed: %v", err)
	}

	feed, _ := svc.GetFeed(ctx, "user1", 10, 0)
	if len(feed) != 0 {
		t.Error("dismissed notification should not appear in feed")
	}
}
