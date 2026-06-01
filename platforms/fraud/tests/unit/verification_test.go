package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/verification"
)

func TestInitiateSMSVerification(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, err := svc.InitiateVerification(context.Background(), "user1", verification.MethodSMS, "+1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ID == "" {
		t.Error("expected non-empty ID")
	}
	if req.Status != verification.StatusPending {
		t.Errorf("expected pending, got %s", req.Status)
	}
}

func TestInitiateEmailVerification(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, err := svc.InitiateVerification(context.Background(), "user2", verification.MethodEmail, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Code == "" {
		t.Error("expected non-empty code")
	}
	if len(req.Code) != 6 {
		t.Errorf("expected 6-digit code, got %s", req.Code)
	}
}

func TestVerifyCorrectCode(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, _ := svc.InitiateVerification(context.Background(), "user3", verification.MethodSMS, "+1111111111")
	// Code is generated, we need to access it directly from repo
	fetched, _ := repo.Get(context.Background(), req.ID)

	result, err := svc.VerifyCode(context.Background(), req.ID, fetched.Code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != verification.StatusVerified {
		t.Errorf("expected verified, got %s", result.Status)
	}
}

func TestVerifyWrongCodeMismatch(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, _ := svc.InitiateVerification(context.Background(), "user4", verification.MethodEmail, "test@test.com")

	_, err := svc.VerifyCode(context.Background(), req.ID, "000000")
	if err == nil {
		t.Fatal("expected error for wrong code")
	}
	if err != verification.ErrCodeMismatch {
		t.Errorf("expected ErrCodeMismatch, got %v", err)
	}
}

func TestVerifyExpiredCode(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 0)

	req, _ := svc.InitiateVerification(context.Background(), "user5", verification.MethodSMS, "+5555555555")

	time.Sleep(2 * time.Millisecond)

	_, err := svc.VerifyCode(context.Background(), req.ID, req.Code)
	if err == nil {
		t.Fatal("expected error for expired code")
	}
}

func TestVerifyMaxAttempts(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, _ := svc.InitiateVerification(context.Background(), "user6", verification.MethodSMS, "+6666666666")

	for i := 0; i < 3; i++ {
		svc.VerifyCode(context.Background(), req.ID, "wrong")
	}

	_, err := svc.VerifyCode(context.Background(), req.ID, "wrong")
	if err != verification.ErrMaxAttempts {
		t.Errorf("expected ErrMaxAttempts after 3 failures, got %v", err)
	}
}

func TestCheckKYCStatusUnverified(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	status, err := svc.CheckKYCStatus(context.Background(), "user-kyc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.IsVerified {
		t.Error("expected unverified KYC")
	}
}

func TestSetKYCStatus(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	now := time.Now()
	svc.SetKYCStatus(context.Background(), &verification.KYCStatus{
		UserID: "user-kyc2", IsVerified: true, Level: "tier2",
		DocumentType: "passport", ApprovedAt: &now,
	})

	status, _ := svc.CheckKYCStatus(context.Background(), "user-kyc2")
	if !status.IsVerified {
		t.Error("expected verified KYC")
	}
	if status.Level != "tier2" {
		t.Errorf("expected tier2, got %s", status.Level)
	}
}

func TestGetNonexistentVerification(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	_, err := svc.VerifyCode(context.Background(), "nonexistent", "123456")
	if err != verification.ErrVerificationNotFound {
		t.Errorf("expected ErrVerificationNotFound, got %v", err)
	}
}

func TestVerifyAlreadyVerified(t *testing.T) {
	repo := verification.NewInMemoryRepository()
	svc := verification.NewService(repo, 10)

	req, _ := svc.InitiateVerification(context.Background(), "user-double", verification.MethodEmail, "d@d.com")
	fetched, _ := repo.Get(context.Background(), req.ID)

	_, err := svc.VerifyCode(context.Background(), req.ID, fetched.Code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify again should succeed (already verified)
	result, err := svc.VerifyCode(context.Background(), req.ID, "000000")
	if err != nil {
		t.Fatalf("expected success for already verified, got %v", err)
	}
	if result.Status != verification.StatusVerified {
		t.Errorf("expected verified, got %s", result.Status)
	}
}
