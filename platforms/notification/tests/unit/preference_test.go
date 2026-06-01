package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/notification/internal/preferences"
)

func TestGetPreferencesDefaults(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	pref, err := svc.GetPreferences(ctx, "user1")
	if err != nil {
		t.Fatalf("GetPreferences failed: %v", err)
	}
	if !pref.ChannelOptIn.Push {
		t.Error("push should be opted in by default")
	}
	if !pref.ChannelOptIn.InApp {
		t.Error("inapp should be opted in by default")
	}
}

func TestUpdatePreferences(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	optIn := false
	pushEnabled := false

	updated, err := svc.UpdatePreferences(ctx, "user1", &preferences.UpdatePreferenceRequest{
		ChannelOptIn: &preferences.ChannelOptIn{Push: false, Email: true, SMS: true, InApp: true},
		PushEnabled:  &pushEnabled,
		EmailDigest:  &optIn,
	})
	if err != nil {
		t.Fatalf("UpdatePreferences failed: %v", err)
	}
	if updated.ChannelOptIn.Push {
		t.Error("push should be opted out")
	}
	if updated.PushEnabled {
		t.Error("push should be disabled")
	}
	if updated.EmailDigest {
		t.Error("email digest should be disabled")
	}
}

func TestShouldSend(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	ok, err := svc.ShouldSend(ctx, "user1", "push", "promotion")
	if err != nil {
		t.Fatalf("ShouldSend failed: %v", err)
	}
	if !ok {
		t.Error("should allow sending by default")
	}
}

func TestShouldSendChannelOptedOut(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	pushEnabled := false
	svc.UpdatePreferences(ctx, "user1", &preferences.UpdatePreferenceRequest{
		PushEnabled: &pushEnabled,
	})

	ok, err := svc.ShouldSend(ctx, "user1", "push", "promotion")
	if err != nil {
		t.Fatalf("ShouldSend failed: %v", err)
	}
	if ok {
		t.Error("should not allow sending when push is disabled")
	}
}

func TestSuppressionList(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	svc.AddSuppression(ctx, "user1", "", "", "unsubscribed")

	suppressed, err := svc.IsSuppressed(ctx, "user1", "", "")
	if err != nil {
		t.Fatalf("IsSuppressed failed: %v", err)
	}
	if !suppressed {
		t.Error("user1 should be suppressed")
	}
}

func TestQuietHoursEnforcement(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	svc.UpdatePreferences(ctx, "user1", &preferences.UpdatePreferenceRequest{
		QuietHours: &preferences.QuietHours{
			Enabled: true,
			Start:   "00:00",
			End:     "23:59",
			Timezone: "UTC",
		},
	})

	ok, err := svc.ShouldSend(ctx, "user1", "push", "promotion")
	if err != nil {
		t.Fatalf("ShouldSend failed: %v", err)
	}
	if ok {
		t.Error("should not send during quiet hours")
	}
}

func TestPreferenceCRUD(t *testing.T) {
	repo := preferences.NewInMemoryRepository()
	svc := preferences.NewService(repo)
	ctx := context.Background()

	pref, err := svc.GetPreferences(ctx, "user2")
	if err != nil {
		t.Fatalf("GetPreferences failed: %v", err)
	}
	if pref.UserID != "user2" {
		t.Errorf("expected user2, got %s", pref.UserID)
	}

	digest := true
	updated, err := svc.UpdatePreferences(ctx, "user2", &preferences.UpdatePreferenceRequest{
		EmailDigest: &digest,
	})
	if err != nil {
		t.Fatalf("UpdatePreferences failed: %v", err)
	}
	if !updated.EmailDigest {
		t.Error("email digest should be enabled")
	}

	fetched, _ := svc.GetPreferences(ctx, "user2")
	if !fetched.EmailDigest {
		t.Error("fetched preference should have email digest enabled")
	}
}
