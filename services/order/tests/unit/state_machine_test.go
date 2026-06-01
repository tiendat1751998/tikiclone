package unit

import (
	"errors"
	"testing"

	"github.com/tikiclone/tiki/services/order/internal/domain"
)

func TestStateMachine_ValidTransitions(t *testing.T) {
	sm := domain.NewStateMachine()

	tests := []struct {
		from     domain.OrderStatus
		to       domain.OrderStatus
		expected bool
	}{
		{domain.OrderStatusPending, domain.OrderStatusAwaitingPayment, true},
		{domain.OrderStatusPending, domain.OrderStatusCancelled, true},
		{domain.OrderStatusPending, domain.OrderStatusPaid, false},
		{domain.OrderStatusAwaitingPayment, domain.OrderStatusPaid, true},
		{domain.OrderStatusAwaitingPayment, domain.OrderStatusCancelled, true},
		{domain.OrderStatusPaid, domain.OrderStatusProcessing, true},
		{domain.OrderStatusPaid, domain.OrderStatusCancelled, true},
		{domain.OrderStatusProcessing, domain.OrderStatusPacked, true},
		{domain.OrderStatusPacked, domain.OrderStatusShipped, true},
		{domain.OrderStatusShipped, domain.OrderStatusDelivered, true},
		{domain.OrderStatusDelivered, domain.OrderStatusCompleted, true},
		{domain.OrderStatusCompleted, domain.OrderStatusRefunded, true},
		{domain.OrderStatusCancelled, domain.OrderStatusPaid, false},
		{domain.OrderStatusRefunded, domain.OrderStatusPaid, false},
		{domain.OrderStatusCompleted, domain.OrderStatusCancelled, false},
		{domain.OrderStatusDelivered, domain.OrderStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"_to_"+string(tt.to), func(t *testing.T) {
			result := sm.CanTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestOrder_CanTransitionTo(t *testing.T) {
	order := &domain.Order{
		Status: domain.OrderStatusPending,
	}

	if !order.CanTransitionTo(domain.OrderStatusAwaitingPayment) {
		t.Error("expected pending -> awaiting_payment to be valid")
	}
	if order.CanTransitionTo(domain.OrderStatusPaid) {
		t.Error("expected pending -> paid to be invalid")
	}
}

func TestOrder_IsCancellable(t *testing.T) {
	tests := []struct {
		status   domain.OrderStatus
		expected bool
	}{
		{domain.OrderStatusPending, true},
		{domain.OrderStatusAwaitingPayment, true},
		{domain.OrderStatusPaid, true},
		{domain.OrderStatusProcessing, true},
		{domain.OrderStatusPacked, true},
		{domain.OrderStatusShipped, false},
		{domain.OrderStatusDelivered, false},
		{domain.OrderStatusCompleted, false},
		{domain.OrderStatusCancelled, false},
		{domain.OrderStatusRefunded, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			order := &domain.Order{Status: tt.status}
			if order.IsCancellable() != tt.expected {
				t.Errorf("IsCancellable(%s) = %v, want %v", tt.status, order.IsCancellable(), tt.expected)
			}
		})
	}
}

func TestOrder_TransitionTo(t *testing.T) {
	order := &domain.Order{
		ID:      "test-order-1",
		Status:  domain.OrderStatusPending,
		Version: 1,
	}

	event, err := order.TransitionTo(domain.OrderStatusAwaitingPayment, "user-1", "user", "checkout_complete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != domain.OrderStatusAwaitingPayment {
		t.Errorf("expected status awaiting_payment, got %s", order.Status)
	}
	if event.FromStatus != domain.OrderStatusPending {
		t.Errorf("expected from_status pending, got %s", event.FromStatus)
	}
	if event.ToStatus != domain.OrderStatusAwaitingPayment {
		t.Errorf("expected to_status awaiting_payment, got %s", event.ToStatus)
	}
	if order.Version != 2 {
		t.Errorf("expected version 2, got %d", order.Version)
	}
}

func TestOrder_InvalidTransition(t *testing.T) {
	order := &domain.Order{
		ID:     "test-order-2",
		Status: domain.OrderStatusPending,
	}

	_, err := order.TransitionTo(domain.OrderStatusPaid, "user-1", "user", "direct_pay")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Errorf("expected ErrInvalidStateTransition, got %v", err)
	}
}

func TestOrderNumberGenerator(t *testing.T) {
	gen := domain.NewOrderNumberGenerator()

	// Generate multiple numbers, ensure they're unique
	numbers := make(map[string]bool)
	for i := 0; i < 100; i++ {
		num := gen.Generate()
		if numbers[num] {
			t.Errorf("duplicate order number generated: %s", num)
		}
		numbers[num] = true
	}
}

func TestOrder_IsTerminal(t *testing.T) {
	tests := []struct {
		status   domain.OrderStatus
		expected bool
	}{
		{domain.OrderStatusPending, false},
		{domain.OrderStatusPaid, false},
		{domain.OrderStatusProcessing, false},
		{domain.OrderStatusCompleted, true},
		{domain.OrderStatusCancelled, true},
		{domain.OrderStatusRefunded, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			order := &domain.Order{Status: tt.status}
			if order.IsTerminal() != tt.expected {
				t.Errorf("IsTerminal(%s) = %v, want %v", tt.status, order.IsTerminal(), tt.expected)
			}
		})
	}
}
