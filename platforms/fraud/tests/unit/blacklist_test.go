package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
)

func TestAddBlacklistEntry(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	entry := &blacklist.BlacklistEntry{
		Type:  blacklist.BlacklistIP,
		Value: "192.168.1.1",
		Reason: blacklist.ReasonFraudulentActivity,
	}

	if err := svc.Add(context.Background(), entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCheckBlacklistedIP(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	svc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistIP, Value: "10.0.0.5",
		Reason: blacklist.ReasonSuspiciousBehavior,
	})

	resp, err := svc.Check(context.Background(), &blacklist.CheckRequest{IP: "10.0.0.5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Blocked {
		t.Error("expected blocked for blacklisted IP")
	}
	if len(resp.Reasons) == 0 {
		t.Error("expected at least 1 reason")
	}
}

func TestCheckNotBlacklisted(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	resp, err := svc.Check(context.Background(), &blacklist.CheckRequest{IP: "1.2.3.4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Blocked {
		t.Error("expected not blocked for clean IP")
	}
}

func TestCheckBlacklistedUser(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	svc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistUser, Value: "bad-user",
		Reason: blacklist.ReasonChargeback,
	})

	resp, _ := svc.Check(context.Background(), &blacklist.CheckRequest{UserID: "bad-user"})
	if !resp.Blocked {
		t.Error("expected blocked for blacklisted user")
	}
}

func TestRemoveBlacklistEntry(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	entry := &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistIP, Value: "10.0.0.99",
		Reason: blacklist.ReasonManualReview,
	}
	svc.Add(context.Background(), entry)

	if err := svc.Remove(context.Background(), entry.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, _ := svc.Check(context.Background(), &blacklist.CheckRequest{IP: "10.0.0.99"})
	if resp.Blocked {
		t.Error("expected not blocked after removal")
	}
}

func TestRemoveByValue(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	svc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistCard, Value: "4111-1111-1111-1111",
		Reason: blacklist.ReasonFraudulentActivity,
	})

	if err := svc.RemoveByValue(context.Background(), blacklist.BlacklistCard, "4111-1111-1111-1111"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, _ := svc.Check(context.Background(), &blacklist.CheckRequest{CardNumber: "4111-1111-1111-1111"})
	if resp.Blocked {
		t.Error("expected not blocked after removal")
	}
}

func TestBulkImport(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	entries := []blacklist.BlacklistEntry{
		{Type: blacklist.BlacklistIP, Value: "10.0.0.1", Reason: blacklist.ReasonHighRisk},
		{Type: blacklist.BlacklistIP, Value: "10.0.0.2", Reason: blacklist.ReasonHighRisk},
		{Type: blacklist.BlacklistIP, Value: "10.0.0.3", Reason: blacklist.ReasonHighRisk},
	}

	added, err := svc.BulkImport(context.Background(), entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added != 3 {
		t.Errorf("expected 3 added, got %d", added)
	}
}

func TestTTLExpiry(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	exp := time.Now().Add(-1 * time.Hour)
	svc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistIP, Value: "10.0.0.100",
		Reason: blacklist.ReasonManualReview, ExpiresAt: &exp,
	})

	expired, err := svc.ExpireEntries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expired != 1 {
		t.Errorf("expected 1 expired, got %d", expired)
	}

	entry, err := repo.GetByTypeAndValue(context.Background(), blacklist.BlacklistIP, "10.0.0.100")
	if err == nil {
		t.Errorf("expected entry to be expired, got %+v", entry)
	}
}

func TestDeviceBlacklist(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	svc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistDevice, Value: "stolen-device-id",
		Reason: blacklist.ReasonFraudulentActivity,
	})

	resp, _ := svc.Check(context.Background(), &blacklist.CheckRequest{DeviceID: "stolen-device-id"})
	if !resp.Blocked {
		t.Error("expected blocked for blacklisted device")
	}
}

func TestMultipleBlacklistHits(t *testing.T) {
	repo := blacklist.NewInMemoryRepository()
	svc := blacklist.NewService(repo)

	svc.Add(context.Background(), &blacklist.BlacklistEntry{Type: blacklist.BlacklistUser, Value: "bad-user", Reason: blacklist.ReasonChargeback})
	svc.Add(context.Background(), &blacklist.BlacklistEntry{Type: blacklist.BlacklistIP, Value: "10.0.0.5", Reason: blacklist.ReasonFraudulentActivity})

	resp, _ := svc.Check(context.Background(), &blacklist.CheckRequest{UserID: "bad-user", IP: "10.0.0.5"})
	if !resp.Blocked {
		t.Error("expected blocked for multiple blacklist matches")
	}
	if len(resp.Reasons) != 2 {
		t.Errorf("expected 2 reasons, got %d", len(resp.Reasons))
	}
}
