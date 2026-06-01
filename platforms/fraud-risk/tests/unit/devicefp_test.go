package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/devicefp"
)

func TestIdentifyNewDevice(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	fp := &devicefp.Fingerprint{
		UserAgent:    "Mozilla/5.0",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		ColorDepth:   24,
		Platform:     "Win32",
		Language:     "en-US",
		Timezone:     "UTC",
	}

	profile, isNew, err := svc.IdentifyDevice(context.Background(), fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isNew {
		t.Error("expected device to be new")
	}
	if profile.DeviceID == "" {
		t.Error("expected non-empty device ID")
	}
	if len(profile.UserAgents) != 1 {
		t.Errorf("expected 1 user agent, got %d", len(profile.UserAgents))
	}
}

func TestIdentifyExistingDevice(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	fp := &devicefp.Fingerprint{
		UserAgent:    "Mozilla/5.0",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		ColorDepth:   24,
		Platform:     "Win32",
		Language:     "en-US",
		Timezone:     "UTC",
	}

	profile1, _, err := svc.IdentifyDevice(context.Background(), fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile2, isNew, err := svc.IdentifyDevice(context.Background(), fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isNew {
		t.Error("expected device to be existing on second call")
	}
	if profile1.DeviceID != profile2.DeviceID {
		t.Error("expected same device ID for identical fingerprint")
	}
}

func TestGetDeviceHistory(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	fp := &devicefp.Fingerprint{
		UserAgent:    "Chrome/90",
		ScreenWidth:  1366,
		ScreenHeight: 768,
		ColorDepth:   32,
		Platform:     "Linux",
		Language:     "es",
		Timezone:     "America/Mexico_City",
	}

	profile, _, err := svc.IdentifyDevice(context.Background(), fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	history, err := svc.GetDeviceHistory(context.Background(), profile.DeviceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if history.DeviceID != profile.DeviceID {
		t.Error("expected matching device ID")
	}
}

func TestMarkSuspicious(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	fp := &devicefp.Fingerprint{UserAgent: "TestAgent", ScreenWidth: 800, ScreenHeight: 600}
	profile, _, err := svc.IdentifyDevice(context.Background(), fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	suspicious, err := svc.MarkSuspicious(context.Background(), profile.DeviceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !suspicious.IsSuspicious {
		t.Error("expected device to be marked suspicious")
	}
	if suspicious.RiskScore != 100 {
		t.Errorf("expected risk score 100, got %f", suspicious.RiskScore)
	}
}

func TestMarkSuspiciousNonexistent(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	_, err := svc.MarkSuspicious(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestCalculateHash(t *testing.T) {
	repo := devicefp.NewInMemoryRepository()
	svc := devicefp.NewService(repo)

	fp1 := &devicefp.Fingerprint{
		UserAgent: "Same", ScreenWidth: 1920, ScreenHeight: 1080,
		ColorDepth: 24, Platform: "Win32", Language: "en", Timezone: "UTC",
	}
	fp2 := &devicefp.Fingerprint{
		UserAgent: "Same", ScreenWidth: 1920, ScreenHeight: 1080,
		ColorDepth: 24, Platform: "Win32", Language: "en", Timezone: "UTC",
	}

	profile1, _, err := svc.IdentifyDevice(context.Background(), fp1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	profile2, _, err := svc.IdentifyDevice(context.Background(), fp2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile1.DeviceID != profile2.DeviceID {
		t.Error("expected same device ID for identical fingerprints")
	}
}
