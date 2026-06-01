package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/email"
	"github.com/tikiclone/tiki/platforms/notification/internal/events"
)

func TestSendEmail(t *testing.T) {
	repo := email.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := email.NewService(repo, pub, "noreply@tiki-clone.com")
	ctx := context.Background()

	req := &email.SendEmailRequest{
		To:        []string{"user@example.com"},
		Subject:   "Welcome!",
		PlainText: "Welcome to Tiki Clone",
		HTML:      "<h1>Welcome!</h1>",
	}

	msg, err := svc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}
	if msg.ID == "" {
		t.Error("expected email ID")
	}
	if msg.Subject != "Welcome!" {
		t.Errorf("expected subject 'Welcome!', got %s", msg.Subject)
	}
	if msg.Status != email.EmailStatusSent {
		t.Errorf("expected status sent, got %s", msg.Status)
	}
}

func TestSendEmailWithCCBCC(t *testing.T) {
	repo := email.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := email.NewService(repo, pub, "noreply@tiki-clone.com")
	ctx := context.Background()

	req := &email.SendEmailRequest{
		To:      []string{"user@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		Subject: "Test",
		HTML:    "<p>Test</p>",
	}

	msg, err := svc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}
	if len(msg.CC) != 1 || msg.CC[0] != "cc@example.com" {
		t.Error("CC not set correctly")
	}
	if len(msg.BCC) != 1 || msg.BCC[0] != "bcc@example.com" {
		t.Error("BCC not set correctly")
	}
}

func TestBulkEmail(t *testing.T) {
	repo := email.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := email.NewService(repo, pub, "noreply@tiki-clone.com")
	ctx := context.Background()

	reqs := []*email.SendEmailRequest{
		{To: []string{"a@example.com"}, Subject: "A", HTML: "<p>A</p>"},
		{To: []string{"b@example.com"}, Subject: "B", HTML: "<p>B</p>"},
	}

	results, err := svc.BulkEmail(ctx, reqs)
	if err != nil {
		t.Fatalf("BulkEmail failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestEmailStatusTracking(t *testing.T) {
	repo := email.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := email.NewService(repo, pub, "noreply@tiki-clone.com")
	ctx := context.Background()

	msg, _ := svc.SendEmail(ctx, &email.SendEmailRequest{
		To:      []string{"user@example.com"},
		Subject: "Track",
		HTML:    "<p>Track</p>",
	})

	if err := svc.TrackBounce(ctx, msg.ID); err != nil {
		t.Fatalf("TrackBounce failed: %v", err)
	}

	updated, _ := svc.GetEmail(ctx, msg.ID)
	if updated.Status != email.EmailStatusBounced {
		t.Errorf("expected bounced, got %s", updated.Status)
	}

	if err := svc.TrackOpen(ctx, msg.ID); err != nil {
		t.Fatalf("TrackOpen failed: %v", err)
	}

	updated, _ = svc.GetEmail(ctx, msg.ID)
	if updated.Status != email.EmailStatusOpened {
		t.Errorf("expected opened, got %s", updated.Status)
	}
}

func TestTemplatedEmail(t *testing.T) {
	repo := email.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := email.NewService(repo, pub, "noreply@tiki-clone.com")
	ctx := context.Background()

	req := &email.SendEmailRequest{
		To:      []string{"user@example.com"},
		Subject: "Hello {{.Name}}",
		HTML:    "<p>Welcome {{.Name}}!</p>",
	}

	msg, err := svc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
}
