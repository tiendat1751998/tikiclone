package domain

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/checkout/internal/domain"
)

func TestNewCheckout(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "idem-key-001", 30*time.Minute)
	if c.UserID != "USER-001" {
		t.Errorf("expected USER-001, got %s", c.UserID)
	}
	if c.CartID != "CART-001" {
		t.Errorf("expected CART-001, got %s", c.CartID)
	}
	if c.Status != domain.CheckoutStatusPending {
		t.Errorf("expected pending status, got %s", c.Status)
	}
	if c.CurrentStep != domain.StepInit {
		t.Errorf("expected init step, got %s", c.CurrentStep)
	}
	if c.IsExpired() {
		t.Error("new checkout should not be expired")
	}
}

func TestCheckout_IsExpired(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if !c.IsExpired() {
		t.Error("checkout should be expired")
	}
}

func TestCheckout_AdvanceStep(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 30*time.Minute)
	c.AdvanceStep(domain.StepValidate)
	if c.CurrentStep != domain.StepValidate {
		t.Errorf("expected validate step, got %s", c.CurrentStep)
	}
}

func TestCheckout_MarkCompleted(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 30*time.Minute)
	c.MarkCompleted("ORD-001")
	if c.Status != domain.CheckoutStatusCompleted {
		t.Errorf("expected completed status, got %s", c.Status)
	}
	if c.OrderID != "ORD-001" {
		t.Errorf("expected ORD-001, got %s", c.OrderID)
	}
	if c.CompletedAt == nil {
		t.Error("completed_at should be set")
	}
}

func TestCheckout_MarkFailed(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 30*time.Minute)
	c.MarkFailed("test failure")
	if c.Status != domain.CheckoutStatusFailed {
		t.Errorf("expected failed status, got %s", c.Status)
	}
	if c.FailureReason != "test failure" {
		t.Errorf("expected 'test failure', got %s", c.FailureReason)
	}
}

func TestCheckout_CanRetry(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 30*time.Minute)
	if !c.CanRetry() {
		t.Error("new checkout should be retryable")
	}

	c.IncrementAttempt()
	c.IncrementAttempt()
	c.IncrementAttempt()
	if c.CanRetry() {
		t.Error("checkout with 3 attempts should not be retryable")
	}
}

func TestCheckout_MarkRollingBack(t *testing.T) {
	c := domain.NewCheckout("USER-001", "CART-001", "", 30*time.Minute)
	c.MarkRollingBack()
	if c.Status != domain.CheckoutStatusRollingBack {
		t.Errorf("expected rolling_back status, got %s", c.Status)
	}
	c.MarkRolledBack()
	if c.Status != domain.CheckoutStatusRolledBack {
		t.Errorf("expected rolled_back status, got %s", c.Status)
	}
}

func TestNewCheckoutStepLog(t *testing.T) {
	log := domain.NewCheckoutStepLog("CHK-001", "validate", "success", 150, "", "")
	if log.CheckoutID != "CHK-001" {
		t.Errorf("expected CHK-001, got %s", log.CheckoutID)
	}
	if log.Step != "validate" {
		t.Errorf("expected validate, got %s", log.Step)
	}
	if log.DurationMs != 150 {
		t.Errorf("expected 150ms, got %d", log.DurationMs)
	}
}

func TestNewPricingSnapshot(t *testing.T) {
	snap := &domain.PricingSnapshot{
		ID: "SNAP-001", CheckoutID: "CHK-001",
		Subtotal: 100000, GrandTotal: 90000, Currency: "SGD",
		CreatedAt: time.Now(),
	}
	if snap.GrandTotal != 90000 {
		t.Errorf("expected 90000, got %d", snap.GrandTotal)
	}
}

func TestNewReconciliationJob(t *testing.T) {
	job := &domain.ReconciliationJob{
		ID: "JOB-001", CheckoutID: "CHK-001",
		JobType: domain.JobTypeReleaseReservation,
		Status:  domain.JobStatusPending,
		MaxAttempts: 3,
	}
	if job.JobType != domain.JobTypeReleaseReservation {
		t.Errorf("expected release_reservation, got %s", job.JobType)
	}
}
