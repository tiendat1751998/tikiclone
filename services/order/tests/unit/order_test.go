package unit

import (
	"testing"

	"github.com/tikiclone/tiki/services/order/internal/domain"
)

func TestNewOrder(t *testing.T) {
	items := []domain.OrderItem{
		*domain.NewOrderItem("", "prod-1", "sku-1", "shop-1", 2, 1000, nil),
		*domain.NewOrderItem("", "prod-2", "sku-2", "shop-1", 1, 5000, nil),
	}

	shipping := domain.Address{
		Street1:    "123 Main St",
		City:       "Singapore",
		PostalCode: "123456",
		Country:    "SG",
	}

	order := domain.NewOrder("user-1", "shop-1", "SGD", "idem-key-1", shipping, shipping, items)

	if order.UserID != "user-1" {
		t.Errorf("expected user_id user-1, got %s", order.UserID)
	}
	if order.Status != domain.OrderStatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
	if order.TotalAmount != 7000 {
		t.Errorf("expected total 7000, got %d", order.TotalAmount)
	}
	if order.Currency != "SGD" {
		t.Errorf("expected currency SGD, got %s", order.Currency)
	}
	if len(order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(order.Items))
	}
	if order.Version != 1 {
		t.Errorf("expected version 1, got %d", order.Version)
	}
}

func TestNewOrderItem(t *testing.T) {
	item := domain.NewOrderItem("order-1", "prod-1", "sku-1", "shop-1", 3, 1000, nil)

	if item.OrderID != "order-1" {
		t.Errorf("expected order_id order-1, got %s", item.OrderID)
	}
	if item.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", item.Quantity)
	}
	if item.TotalPrice != 3000 {
		t.Errorf("expected total_price 3000, got %d", item.TotalPrice)
	}
}

func TestSplitOrderBySellers(t *testing.T) {
	items := []domain.OrderItem{
		*domain.NewOrderItem("", "prod-1", "sku-1", "shop-1", 1, 1000, nil),
		*domain.NewOrderItem("", "prod-2", "sku-2", "shop-1", 1, 2000, nil),
		*domain.NewOrderItem("", "prod-3", "sku-3", "shop-2", 1, 3000, nil),
	}

	order := &domain.Order{Items: items}
	splits := domain.SplitOrderBySellers(order)

	if len(splits) != 2 {
		t.Errorf("expected 2 seller groups, got %d", len(splits))
	}
	if len(splits["shop-1"]) != 2 {
		t.Errorf("expected 2 items for shop-1, got %d", len(splits["shop-1"]))
	}
	if len(splits["shop-2"]) != 1 {
		t.Errorf("expected 1 item for shop-2, got %d", len(splits["shop-2"]))
	}
}

func TestOrderSnapshot(t *testing.T) {
	cart := &domain.CartSnapshot{
		Items: []domain.SnapshotItem{
			{ProductID: "prod-1", SkuID: "sku-1", ShopID: "shop-1", Name: "Test Product", Quantity: 2, UnitPrice: 1000},
		},
		TotalAmount: 2000,
		Currency:    "SGD",
	}

	snapshot, err := domain.NewOrderSnapshot("order-1", cart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !snapshot.VerifyChecksum() {
		t.Error("expected checksum to verify")
	}

	restored, err := snapshot.CartSnapshot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored.TotalAmount != 2000 {
		t.Errorf("expected total 2000, got %d", restored.TotalAmount)
	}
}

func TestOrderCancellation(t *testing.T) {
	cancel := domain.NewOrderCancellation("order-1", "changed mind", "user-1", domain.CancellationTypeUser, 5000)

	if cancel.CompensationStatus != domain.CompensationPending {
		t.Errorf("expected compensation status pending, got %s", cancel.CompensationStatus)
	}
	if cancel.RefundAmount != 5000 {
		t.Errorf("expected refund amount 5000, got %d", cancel.RefundAmount)
	}
}

func TestOrderReconciliation(t *testing.T) {
	rec := domain.NewOrderReconciliation("order-1", domain.ReconciliationTypePayment)

	if rec.Status != domain.ReconciliationStatusPending {
		t.Errorf("expected status pending, got %s", rec.Status)
	}
	if !rec.CanRetry() {
		t.Error("expected to be able to retry")
	}

	rec.IncrementRetry()
	if rec.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", rec.RetryCount)
	}

	// Max retries is 3
	rec.RetryCount = 3
	if rec.CanRetry() {
		t.Error("expected to not be able to retry after max retries")
	}
}

func TestIdempotencyRecord(t *testing.T) {
	record := domain.NewIdempotencyRecord("order-1", 24*3600000000000) // 24h in nanoseconds

	if record.OrderID != "order-1" {
		t.Errorf("expected order_id order-1, got %s", record.OrderID)
	}
	if record.IsExpired() {
		t.Error("expected record to not be expired")
	}
}
