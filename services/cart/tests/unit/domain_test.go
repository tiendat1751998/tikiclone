package domain_test

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/cart/internal/domain"
)

func TestNewCart(t *testing.T) {
	cart := domain.NewCart("USER-001", "SESSION-001", "SGD", 7*24*time.Hour)
	if cart.UserID != "USER-001" {
		t.Errorf("expected USER-001, got %s", cart.UserID)
	}
	if cart.SessionID != "SESSION-001" {
		t.Errorf("expected SESSION-001, got %s", cart.SessionID)
	}
	if cart.Status != domain.CartStatusActive {
		t.Errorf("expected active status, got %s", cart.Status)
	}
	if cart.Currency != "SGD" {
		t.Errorf("expected SGD, got %s", cart.Currency)
	}
	if cart.IsExpired() {
		t.Error("new cart should not be expired")
	}
	if !cart.IsActive() {
		t.Error("new cart should be active")
	}
}

func TestCart_IsExpired(t *testing.T) {
	cart := domain.NewCart("USER-001", "", "SGD", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if !cart.IsExpired() {
		t.Error("cart should be expired")
	}
	if cart.IsActive() {
		t.Error("expired cart should not be active")
	}
}

func TestCart_MarkMerged(t *testing.T) {
	cart := domain.NewCart("USER-001", "", "SGD", 7*24*time.Hour)
	cart.MarkMerged()
	if cart.Status != domain.CartStatusMerged {
		t.Errorf("expected merged status, got %s", cart.Status)
	}
}

func TestCart_MarkCheckout(t *testing.T) {
	cart := domain.NewCart("USER-001", "", "SGD", 7*24*time.Hour)
	if err := cart.MarkCheckout(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cart.Status != domain.CartStatusCheckout {
		t.Errorf("expected checkout status, got %s", cart.Status)
	}
}

func TestCart_MarkCheckout_InvalidState(t *testing.T) {
	cart := domain.NewCart("USER-001", "", "SGD", 7*24*time.Hour)
	cart.MarkMerged()
	err := cart.MarkCheckout()
	if err == nil {
		t.Error("expected error for marking merged cart as checkout")
	}
}

func TestCart_UpdateTotals(t *testing.T) {
	cart := domain.NewCart("USER-001", "", "SGD", 7*24*time.Hour)
	cart.UpdateTotals(5, 15000)
	if cart.ItemCount != 5 {
		t.Errorf("expected item count 5, got %d", cart.ItemCount)
	}
	if cart.Subtotal != 15000 {
		t.Errorf("expected subtotal 15000, got %d", cart.Subtotal)
	}
}

func TestNewCartItem(t *testing.T) {
	item := domain.NewCartItem("CART-001", "SKU-001", "Test Product", "SHOP-001", "Test Shop", 3, 5000, "http://img.url", "color:red")
	if item.CartID != "CART-001" {
		t.Errorf("expected CART-001, got %s", item.CartID)
	}
	if item.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", item.Quantity)
	}
	if item.TotalPrice != 15000 {
		t.Errorf("expected total price 15000, got %d", item.TotalPrice)
	}
	if !item.IsSelected {
		t.Error("new item should be selected")
	}
	if !item.IsAvailable {
		t.Error("new item should be available")
	}
}

func TestCartItem_UpdateQuantity(t *testing.T) {
	item := domain.NewCartItem("CART-001", "SKU-001", "Test", "SHOP-001", "Shop", 1, 5000, "", "")
	item.UpdateQuantity(5)
	if item.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", item.Quantity)
	}
	if item.TotalPrice != 25000 {
		t.Errorf("expected total price 25000, got %d", item.TotalPrice)
	}
}

func TestNewCartSnapshot(t *testing.T) {
	snap := domain.NewCartSnapshot("CART-001", "USER-001", `[{"sku":"SKU-001"}]`, `[{"shop_id":"SHOP-001"}]`, 15000, 3, "SGD", "idem-key-001", 15*time.Minute)
	if snap.CartID != "CART-001" {
		t.Errorf("expected CART-001, got %s", snap.CartID)
	}
	if snap.Subtotal != 15000 {
		t.Errorf("expected subtotal 15000, got %d", snap.Subtotal)
	}
	if snap.IdempotencyKey != "idem-key-001" {
		t.Errorf("expected idem key, got %s", snap.IdempotencyKey)
	}
}

func TestNewCartMergeHistory(t *testing.T) {
	history := &domain.CartMergeHistory{
		ID:           "MERGE-001",
		SourceCartID: "CART-SRC",
		TargetCartID: "CART-TGT",
		UserID:       "USER-001",
		MergeType:    domain.MergeTypeGuestToUser,
		ItemsMerged:  5,
		CreatedAt:    time.Now(),
	}
	if history.MergeType != domain.MergeTypeGuestToUser {
		t.Errorf("expected guest_to_user merge type, got %s", history.MergeType)
	}
}
