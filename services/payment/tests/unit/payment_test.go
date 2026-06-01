package unit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/tikiclone/tiki/services/payment/internal/domain"
)

func TestNewPayment(t *testing.T) {
	p := domain.NewPayment("order-1", "user-1", 5000, "SGD", domain.PaymentMethodCreditCard, "stripe", "idem-1")
	if p.OrderID != "order-1" {
		t.Errorf("expected order_id order-1, got %s", p.OrderID)
	}
	if p.Status != domain.PaymentStatusPending {
		t.Errorf("expected status pending, got %s", p.Status)
	}
	if p.Amount != 5000 {
		t.Errorf("expected amount 5000, got %d", p.Amount)
	}
}

func TestPayment_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from, to domain.PaymentStatus
		expected bool
	}{
		{domain.PaymentStatusPending, domain.PaymentStatusAuthorized, true},
		{domain.PaymentStatusPending, domain.PaymentStatusCaptured, false},
		{domain.PaymentStatusAuthorized, domain.PaymentStatusCaptured, true},
		{domain.PaymentStatusCaptured, domain.PaymentStatusRefunded, true},
		{domain.PaymentStatusRefunded, domain.PaymentStatusCaptured, false},
	}
	for _, tt := range tests {
		p := &domain.Payment{Status: tt.from}
		if p.CanTransitionTo(tt.to) != tt.expected {
			t.Errorf("CanTransitionTo(%s,%s)=%v want %v", tt.from, tt.to, p.CanTransitionTo(tt.to), tt.expected)
		}
	}
}

func TestPayment_RemainingAmount(t *testing.T) {
	p := &domain.Payment{Amount: 10000, AmountRefunded: 3000}
	if p.RemainingAmount() != 7000 {
		t.Errorf("expected 7000, got %d", p.RemainingAmount())
	}
}

func TestWebhookSignature(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	secret := "whsec_test"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	if !domain.VerifyWebhookSignature(payload, sig, secret) {
		t.Error("expected valid signature")
	}
	if domain.VerifyWebhookSignature(payload, "bad", secret) {
		t.Error("expected invalid signature")
	}
}
