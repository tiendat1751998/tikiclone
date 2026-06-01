package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/dispatcher"
	"github.com/tikiclone/tiki/platforms/notification/internal/email"
	"github.com/tikiclone/tiki/platforms/notification/internal/events"
	"github.com/tikiclone/tiki/platforms/notification/internal/inapp"
	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
	"github.com/tikiclone/tiki/platforms/notification/internal/preferences"
	"github.com/tikiclone/tiki/platforms/notification/internal/push"
	"github.com/tikiclone/tiki/platforms/notification/internal/sms"
)

func TestDispatchInApp(t *testing.T) {
	dispatchRepo := dispatcher.NewInMemoryRepository()
	notifRepo := notifier.NewInMemoryRepository()
	pushRepo := push.NewInMemoryRepository()
	emailRepo := email.NewInMemoryRepository()
	smsRepo := sms.NewInMemoryRepository()
	inappRepo := inapp.NewInMemoryRepository()
	prefRepo := preferences.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()

	notifSvc := notifier.NewService(notifRepo, pub)
	pushSvc := push.NewService(pushRepo, pub)
	emailSvc := email.NewService(emailRepo, pub, "noreply@test.com")
	smsSvc := sms.NewService(smsRepo, pub, "+15005550006")
	inappSvc := inapp.NewService(inappRepo, pub)
	prefSvc := preferences.NewService(prefRepo)

	dispatchSvc := dispatcher.NewService(dispatchRepo, notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc)
	ctx := context.Background()

	job, err := dispatchSvc.CreateJob(ctx, &notifier.SendNotificationRequest{
		UserID:  "user1",
		Type:    notifier.TypeOrderConfirmation,
		Channel: notifier.ChannelInApp,
		Title:   "Order Confirmed",
		Body:    "Your order has been confirmed",
	})
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.ID == "" {
		t.Error("expected job ID")
	}
}

func TestBatchDispatch(t *testing.T) {
	dispatchRepo := dispatcher.NewInMemoryRepository()
	notifRepo := notifier.NewInMemoryRepository()
	pushRepo := push.NewInMemoryRepository()
	emailRepo := email.NewInMemoryRepository()
	smsRepo := sms.NewInMemoryRepository()
	inappRepo := inapp.NewInMemoryRepository()
	prefRepo := preferences.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()

	notifSvc := notifier.NewService(notifRepo, pub)
	pushSvc := push.NewService(pushRepo, pub)
	emailSvc := email.NewService(emailRepo, pub, "noreply@test.com")
	smsSvc := sms.NewService(smsRepo, pub, "+15005550006")
	inappSvc := inapp.NewService(inappRepo, pub)
	prefSvc := preferences.NewService(prefRepo)

	dispatchSvc := dispatcher.NewService(dispatchRepo, notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc)
	ctx := context.Background()

	jobs := []*dispatcher.DispatchJob{
		{UserID: "u1", Channel: notifier.ChannelInApp, Type: notifier.TypeOrderConfirmation, Title: "T1", Body: "B1"},
		{UserID: "u2", Channel: notifier.ChannelInApp, Type: notifier.TypeShipmentUpdate, Title: "T2", Body: "B2"},
	}

	results, err := dispatchSvc.BatchDispatch(ctx, jobs)
	if err != nil {
		t.Fatalf("BatchDispatch failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestDispatchPreferenceGating(t *testing.T) {
	ctx := context.Background()
	dispatchRepo := dispatcher.NewInMemoryRepository()
	notifRepo := notifier.NewInMemoryRepository()
	pushRepo := push.NewInMemoryRepository()
	emailRepo := email.NewInMemoryRepository()
	smsRepo := sms.NewInMemoryRepository()
	inappRepo := inapp.NewInMemoryRepository()
	prefRepo := preferences.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()

	notifSvc := notifier.NewService(notifRepo, pub)
	pushSvc := push.NewService(pushRepo, pub)
	emailSvc := email.NewService(emailRepo, pub, "noreply@test.com")
	smsSvc := sms.NewService(smsRepo, pub, "+15005550006")
	inappSvc := inapp.NewService(inappRepo, pub)
	prefSvc := preferences.NewService(prefRepo)

	pushEnabled := false
	prefSvc.UpdatePreferences(ctx, "user1", &preferences.UpdatePreferenceRequest{
		PushEnabled: &pushEnabled,
	})

	dispatchSvc := dispatcher.NewService(dispatchRepo, notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc)

	job := &dispatcher.DispatchJob{
		UserID:  "user1",
		Channel: notifier.ChannelPush,
		Type:    notifier.TypePromotion,
		Title:   "Sale!",
		Body:    "Big sale!",
	}

	result, err := dispatchSvc.Dispatch(ctx, job)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if result.Success {
		t.Error("expected dispatch to be blocked by preferences")
	}
}

func TestDispatchFallbackChannel(t *testing.T) {
	dispatchRepo := dispatcher.NewInMemoryRepository()
	notifRepo := notifier.NewInMemoryRepository()
	pushRepo := push.NewInMemoryRepository()
	emailRepo := email.NewInMemoryRepository()
	smsRepo := sms.NewInMemoryRepository()
	inappRepo := inapp.NewInMemoryRepository()
	prefRepo := preferences.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()

	notifSvc := notifier.NewService(notifRepo, pub)
	pushSvc := push.NewService(pushRepo, pub)
	emailSvc := email.NewService(emailRepo, pub, "noreply@test.com")
	smsSvc := sms.NewService(smsRepo, pub, "+15005550006")
	inappSvc := inapp.NewService(inappRepo, pub)
	prefSvc := preferences.NewService(prefRepo)

	dispatchSvc := dispatcher.NewService(dispatchRepo, notifSvc, pushSvc, emailSvc, smsSvc, inappSvc, prefSvc)
	ctx := context.Background()

	job := &dispatcher.DispatchJob{
		UserID:  "user-no-push-devices",
		Channel: notifier.ChannelPush,
		Type:    notifier.TypeOrderConfirmation,
		Title:   "Push Fallback",
		Body:    "This push will fail, but system handles it",
	}

	result, err := dispatchSvc.Dispatch(ctx, job)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if result.Success {
		t.Error("expected push to fail (no devices)")
	}
}
