package unit

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/inventory/internal/domain"
)

func TestNewStock(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 100, 10)
	if s.Quantity != 100 {
		t.Errorf("expected 100, got %d", s.Quantity)
	}
	if s.AvailableQty != 100 {
		t.Errorf("expected 100, got %d", s.AvailableQty)
	}
	if s.Status != domain.StockStatusInStock {
		t.Errorf("expected in_stock, got %s", s.Status)
	}
}

func TestStock_Reserve(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 100, 10)
	if err := s.Reserve(30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ReservedQty != 30 {
		t.Errorf("expected reserved 30, got %d", s.ReservedQty)
	}
	if s.AvailableQty != 70 {
		t.Errorf("expected available 70, got %d", s.AvailableQty)
	}
}

func TestStock_ReserveInsufficient(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 10, 5)
	if err := s.Reserve(20); err != domain.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestStock_Release(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 100, 10)
	s.Reserve(30)
	if err := s.Release(20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ReservedQty != 10 {
		t.Errorf("expected reserved 10, got %d", s.ReservedQty)
	}
	if s.AvailableQty != 90 {
		t.Errorf("expected available 90, got %d", s.AvailableQty)
	}
}

func TestStock_Deduct(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 100, 10)
	s.Reserve(30)
	if err := s.Deduct(30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Quantity != 70 {
		t.Errorf("expected quantity 70, got %d", s.Quantity)
	}
	if s.ReservedQty != 0 {
		t.Errorf("expected reserved 0, got %d", s.ReservedQty)
	}
}

func TestStock_Replenish(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 50, 10)
	s.Replenish(50)
	if s.Quantity != 100 {
		t.Errorf("expected 100, got %d", s.Quantity)
	}
	if s.AvailableQty != 100 {
		t.Errorf("expected 100, got %d", s.AvailableQty)
	}
}

func TestStock_StatusTransitions(t *testing.T) {
	s := domain.NewStock("prod-1", "sku-1", "wh-1", 10, 10)
	if s.Status != domain.StockStatusLowStock {
		t.Errorf("expected low_stock, got %s", s.Status)
	}
	s.Replenish(100)
	if s.Status != domain.StockStatusInStock {
		t.Errorf("expected in_stock, got %s", s.Status)
	}
	s.Reserve(109)
	if s.Status != domain.StockStatusLowStock {
		t.Errorf("expected low_stock, got %s", s.Status)
	}
}

func TestNewReservation(t *testing.T) {
	res := domain.NewReservation("order-1", "user-1", "prod-1", "sku-1", "wh-1", 5, 30*time.Minute, "idem-1")
	if res.OrderID != "order-1" {
		t.Errorf("expected order-1, got %s", res.OrderID)
	}
	if res.Quantity != 5 {
		t.Errorf("expected 5, got %d", res.Quantity)
	}
	if res.Status != domain.ReservationStatusActive {
		t.Errorf("expected active, got %s", res.Status)
	}
	if res.IsExpired() {
		t.Error("expected not expired")
	}
}

func TestReservation_Commit(t *testing.T) {
	res := domain.NewReservation("order-1", "user-1", "prod-1", "sku-1", "wh-1", 1, 30*time.Minute, "idem-1")
	if err := res.Commit(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != domain.ReservationStatusCommitted {
		t.Errorf("expected committed, got %s", res.Status)
	}
}

func TestReservation_Release(t *testing.T) {
	res := domain.NewReservation("order-1", "user-1", "prod-1", "sku-1", "wh-1", 1, 30*time.Minute, "idem-1")
	if err := res.Release(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != domain.ReservationStatusReleased {
		t.Errorf("expected released, got %s", res.Status)
	}
}
