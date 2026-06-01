package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/order/internal/application"
	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/order/internal/infrastructure/redis"
)

var (
	testDB    *sql.DB
	testRedis *redisinfra.Store
	testSvc   *application.OrderService
)

func TestMain(m *testing.M) {
	cfg := &config.Config{
		AppName:  "tiki-order-test",
		AppEnv:   "test",
		LogLevel: "error",
		MySQL: config.MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     3306,
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "root"),
			Database: "tiki_orders_test",
			Timeout:  5 * time.Second,
		},
		Redis: config.RedisConfig{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"),
			DB:   15,
		},
		Order: config.OrderConfig{
			DefaultCurrency:   "SGD",
			IdempotencyKeyTTL: 24 * time.Hour,
		},
		Kafka: config.KafkaConfig{
			Brokers: []string{},
		},
	}

	db, err := mysql.NewDB(cfg.MySQL)
	if err != nil {
		fmt.Printf("WARNING: MySQL not available, skipping integration tests: %v\n", err)
		os.Exit(0)
	}

	// Run migrations
	runMigrations(db)
	orderRepo := mysql.NewOrderRepository(db)
	outboxRepo := mysql.NewOutboxRepository(db)
	testRedis = redisinfra.NewStore(nil, cfg.Redis)
	testSvc = application.NewOrderService(cfg, orderRepo, outboxRepo, testRedis, nil)

	code := m.Run()

	db.Close()
	os.Exit(code)
}

func runMigrations(db *sqlx.DB) {
	migration := `
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(36) PRIMARY KEY,
		order_number VARCHAR(64) NOT NULL UNIQUE,
		user_id VARCHAR(36) NOT NULL,
		seller_id VARCHAR(36) NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'pending',
		total_amount BIGINT NOT NULL DEFAULT 0,
		currency VARCHAR(3) NOT NULL DEFAULT 'SGD',
		shipping_address JSON,
		billing_address JSON,
		idempotency_key VARCHAR(255) DEFAULT '',
		snapshot_id VARCHAR(36) DEFAULT '',
		parent_order_id VARCHAR(36) DEFAULT NULL,
		metadata JSON,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL,
		INDEX idx_orders_user_id (user_id),
		INDEX idx_orders_status (status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS order_items (
		id VARCHAR(36) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		product_id VARCHAR(36) NOT NULL,
		sku_id VARCHAR(36) NOT NULL,
		shop_id VARCHAR(36) NOT NULL,
		quantity INT NOT NULL DEFAULT 1,
		unit_price BIGINT NOT NULL DEFAULT 0,
		total_price BIGINT NOT NULL DEFAULT 0,
		snapshot JSON,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_order_items_order_id (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS order_snapshots (
		id VARCHAR(36) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		snapshot_data JSON NOT NULL,
		checksum VARCHAR(64) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS order_lifecycle_history (
		id VARCHAR(36) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		from_state VARCHAR(32) NOT NULL,
		to_state VARCHAR(32) NOT NULL,
		transition_reason VARCHAR(255) DEFAULT '',
		actor_id VARCHAR(36) DEFAULT '',
		actor_type VARCHAR(32) DEFAULT '',
		metadata JSON,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_lifecycle_order_id (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS order_cancellations (
		id VARCHAR(36) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		reason TEXT NOT NULL,
		cancelled_by VARCHAR(36) NOT NULL,
		cancelled_by_type VARCHAR(32) NOT NULL,
		compensation_status VARCHAR(32) NOT NULL DEFAULT 'pending',
		refund_amount BIGINT NOT NULL DEFAULT 0,
		metadata JSON,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS order_reconciliation (
		id VARCHAR(36) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		reconciliation_type VARCHAR(32) NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'pending',
		last_checked_at TIMESTAMP NULL,
		retry_count INT NOT NULL DEFAULT 0,
		max_retries INT NOT NULL DEFAULT 3,
		metadata JSON,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS idempotency_keys (
		` + "`key`" + ` VARCHAR(255) PRIMARY KEY,
		order_id VARCHAR(36) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS outbox_events (
		event_id VARCHAR(36) PRIMARY KEY,
		aggregate_type VARCHAR(100) NOT NULL,
		aggregate_id VARCHAR(100) NOT NULL,
		event_type VARCHAR(100) NOT NULL,
		payload JSON NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		processed BOOLEAN NOT NULL DEFAULT FALSE,
		INDEX idx_outbox_processed (processed, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	db.Exec(migration)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func cleanupOrders(db *sql.DB) {
	db.Exec("DELETE FROM outbox_events")
	db.Exec("DELETE FROM idempotency_keys")
	db.Exec("DELETE FROM order_lifecycle_history")
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM order_snapshots")
	db.Exec("DELETE FROM order_cancellations")
	db.Exec("DELETE FROM order_reconciliation")
	db.Exec("DELETE FROM orders")
}

func TestIntegration_CreateOrder(t *testing.T) {
	if testDB == nil {
		t.Skip("MySQL not available")
	}
	defer cleanupOrders(testDB)

	ctx := context.Background()
	req := &application.CreateOrderRequest{
		UserID:         "user-int-1",
		SellerID:       "shop-int-1",
		Currency:       "SGD",
		IdempotencyKey: "idem-int-1",
		ShippingAddress: domain.Address{
			Street1:    "123 Test St",
			City:       "Singapore",
			PostalCode: "123456",
			Country:    "SG",
		},
		BillingAddress: domain.Address{
			Street1:    "123 Test St",
			City:       "Singapore",
			PostalCode: "123456",
			Country:    "SG",
		},
		Items: []domain.SnapshotItem{
			{ProductID: "prod-1", SkuID: "sku-1", ShopID: "shop-1", Name: "Test Product", Quantity: 2, UnitPrice: 1000},
		},
	}

	order, err := testSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	if order.Status != domain.OrderStatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
	if order.TotalAmount != 2000 {
		t.Errorf("expected total 2000, got %d", order.TotalAmount)
	}
	if order.OrderNumber == "" {
		t.Error("expected order number to be set")
	}

	// Verify order can be retrieved
	fetched, err := testSvc.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to get order: %v", err)
	}
	if fetched.ID != order.ID {
		t.Errorf("expected order ID %s, got %s", order.ID, fetched.ID)
	}
}

func TestIntegration_Idempotency(t *testing.T) {
	if testDB == nil {
		t.Skip("MySQL not available")
	}
	defer cleanupOrders(testDB)

	ctx := context.Background()
	req := &application.CreateOrderRequest{
		UserID:          "user-idem-1",
		SellerID:        "shop-idem-1",
		IdempotencyKey:  "idem-same-key-123",
		ShippingAddress: domain.Address{Street1: "St", City: "SG", Country: "SG"},
		BillingAddress:  domain.Address{Street1: "St", City: "SG", Country: "SG"},
		Items: []domain.SnapshotItem{
			{ProductID: "p1", SkuID: "s1", ShopID: "sh1", Name: "Item", Quantity: 1, UnitPrice: 100},
		},
	}

	// First creation
	order1, err := testSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Second creation with same idempotency key should return same order
	order2, err := testSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}

	if order1.ID != order2.ID {
		t.Errorf("expected same order ID, got %s vs %s", order1.ID, order2.ID)
	}
}

func TestIntegration_CancelOrder(t *testing.T) {
	if testDB == nil {
		t.Skip("MySQL not available")
	}
	defer cleanupOrders(testDB)

	ctx := context.Background()
	req := &application.CreateOrderRequest{
		UserID:          "user-cancel-1",
		SellerID:        "shop-cancel-1",
		IdempotencyKey:  "idem-cancel-1",
		ShippingAddress: domain.Address{Street1: "St", City: "SG", Country: "SG"},
		BillingAddress:  domain.Address{Street1: "St", City: "SG", Country: "SG"},
		Items: []domain.SnapshotItem{
			{ProductID: "p1", SkuID: "s1", ShopID: "sh1", Name: "Item", Quantity: 1, UnitPrice: 100},
		},
	}

	order, err := testSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("create order failed: %v", err)
	}

	cancelReq := &application.CancelOrderRequest{
		OrderID:       order.ID,
		Reason:        "changed mind",
		CancelledBy:   "user-cancel-1",
		CancelledType: domain.CancellationTypeUser,
	}

	cancelled, err := testSvc.CancelOrder(ctx, cancelReq)
	if err != nil {
		t.Fatalf("cancel order failed: %v", err)
	}

	if cancelled.Status != domain.OrderStatusCancelled {
		t.Errorf("expected status cancelled, got %s", cancelled.Status)
	}
}

func TestIntegration_OrderLifecycle(t *testing.T) {
	if testDB == nil {
		t.Skip("MySQL not available")
	}
	defer cleanupOrders(testDB)

	ctx := context.Background()
	req := &application.CreateOrderRequest{
		UserID:          "user-lifecycle-1",
		SellerID:        "shop-lifecycle-1",
		IdempotencyKey:  "idem-lifecycle-1",
		ShippingAddress: domain.Address{Street1: "St", City: "SG", Country: "SG"},
		BillingAddress:  domain.Address{Street1: "St", City: "SG", Country: "SG"},
		Items: []domain.SnapshotItem{
			{ProductID: "p1", SkuID: "s1", ShopID: "sh1", Name: "Item", Quantity: 1, UnitPrice: 100},
		},
	}

	order, err := testSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("create order failed: %v", err)
	}

	// Transition through lifecycle
	transitions := []domain.OrderStatus{
		domain.OrderStatusAwaitingPayment,
		domain.OrderStatusPaid,
		domain.OrderStatusProcessing,
		domain.OrderStatusPacked,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusCompleted,
	}

	for _, target := range transitions {
		order, err = testSvc.TransitionStatus(ctx, order.ID, target, "user-lifecycle-1", "user", "test transition")
		if err != nil {
			t.Fatalf("transition to %s failed: %v", target, err)
		}
		if order.Status != target {
			t.Errorf("expected status %s, got %s", target, order.Status)
		}
	}

	// Verify history
	history, err := testSvc.GetOrderHistory(ctx, order.ID)
	if err != nil {
		t.Fatalf("get history failed: %v", err)
	}
	if len(transitions) != len(history)-1 { // -1 for initial creation event
		t.Errorf("expected %d history entries, got %d", len(transitions), len(history))
	}
}
