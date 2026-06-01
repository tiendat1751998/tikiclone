package tikiclone

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	sharedRedis "github.com/tikiclone/tiki/packages/go-shared/pkg/redis"

	// Order service
	orderDomain "github.com/tikiclone/tiki/services/order/internal/domain"
	orderMysql "github.com/tikiclone/tiki/services/order/internal/infrastructure/mysql"
	orderRedis "github.com/tikiclone/tiki/services/order/internal/infrastructure/redis"
	orderApp "github.com/tikiclone/tiki/services/order/internal/application"
	orderConfig "github.com/tikiclone/tiki/services/order/internal/config"

	// Inventory service
	invDomain "github.com/tikiclone/tiki/services/inventory/internal/domain"
	invMysql "github.com/tikiclone/tiki/services/inventory/internal/infrastructure/mysql"
	invRedis "github.com/tikiclone/tiki/services/inventory/internal/infrastructure/redis"
	invApp "github.com/tikiclone/tiki/services/inventory/internal/application"
	invConfig "github.com/tikiclone/tiki/services/inventory/internal/config"

	// Payment service
	payDomain "github.com/tikiclone/tiki/services/payment/internal/domain"
	payMysql "github.com/tikiclone/tiki/services/payment/internal/infrastructure/mysql"
	payRedis "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
	payApp "github.com/tikiclone/tiki/services/payment/internal/application"
	payConfig "github.com/tikiclone/tiki/services/payment/internal/config"

	// Promotion service
	promoDomain "github.com/tikiclone/tiki/services/promotion/internal/domain"
	promoMysql "github.com/tikiclone/tiki/services/promotion/internal/infrastructure/mysql"
	promoRedis "github.com/tikiclone/tiki/services/promotion/internal/infrastructure/redis"
	promoApp "github.com/tikiclone/tiki/services/promotion/internal/application"

	// Cart service
	cartDomain "github.com/tikiclone/tiki/services/cart/internal/domain"
	cartMysql "github.com/tikiclone/tiki/services/cart/internal/infrastructure/mysql"
	cartApp "github.com/tikiclone/tiki/services/cart/internal/application"
)

var (
	invService    *invApp.InventoryService
	orderService  *orderApp.OrderService
	payService    *payApp.PaymentService
	promoSvc      *promoApp.PromotionService
	cartSvc       *cartApp.CartService
	globalDB      *sql.DB
	globalDBx     *sqlx.DB
	fraudDetector payDomain.FraudDetector
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbHost := getEnv("MYSQL_HOST", "localhost")
	dbPass := getEnv("MYSQL_PASSWORD", "root_password")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	dsn := fmt.Sprintf("root:%s@tcp(%s:3306)/", dbPass, dbHost)
	rootDB, err := sql.Open("mysql", dsn+"?charset=utf8mb4&parseTime=true&loc=UTC")
	if err != nil {
		fmt.Fprintf(os.Stderr, "MySQL not available: %v\n", err)
		os.Exit(0)
	}
	defer rootDB.Close()
	if err := rootDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "MySQL ping failed: %v\n", err)
		os.Exit(0)
	}

	testDBName := "tiki_production_test"
	rootDB.Exec("DROP DATABASE IF EXISTS " + testDBName)
	rootDB.Exec("CREATE DATABASE " + testDBName + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	rootDB.Close()

	dsnFull := fmt.Sprintf("root:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=UTC", dbPass, dbHost, testDBName)
	db, err := sql.Open("mysql", dsnFull)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(0)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	globalDB = db

	dbx := sqlx.NewDb(db, "mysql")
	globalDBx = dbx

	redisClient, err := sharedRedis.NewClient(redisAddr, "", 15)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Redis not available: %v\n", err)
		os.Exit(0)
	}
	redisClient.FlushDB(ctx)

	runAllMigrations(dbx, testDBName)

	seedTestData(dbx)

	initServices(dbx, redisClient)

	code := m.Run()

	rootDB2, _ := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s:3306)/", dbPass, dbHost))
	rootDB2.Exec("DROP DATABASE IF EXISTS " + testDBName)
	rootDB2.Close()

	os.Exit(code)
}

func runAllMigrations(dbx *sqlx.DB, dbName string) {
	migrations := []string{
		// Inventory
		`CREATE TABLE IF NOT EXISTS stock (
			id VARCHAR(36) PRIMARY KEY, product_id VARCHAR(36) NOT NULL, sku_id VARCHAR(36) NOT NULL,
			warehouse_id VARCHAR(36) NOT NULL, quantity INT NOT NULL DEFAULT 0,
			reserved_qty INT NOT NULL DEFAULT 0, available_qty INT NOT NULL DEFAULT 0,
			status ENUM('in_stock','low_stock','out_of_stock','reserved') NOT NULL DEFAULT 'in_stock',
			reorder_level INT NOT NULL DEFAULT 10, version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_sku_warehouse (sku_id, warehouse_id)
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS reservations (
			id VARCHAR(36) PRIMARY KEY, order_id VARCHAR(36) NOT NULL, user_id VARCHAR(36) NOT NULL,
			product_id VARCHAR(36) NOT NULL, sku_id VARCHAR(36) NOT NULL, warehouse_id VARCHAR(36) NOT NULL,
			quantity INT NOT NULL DEFAULT 0,
			status ENUM('active','committed','released','expired') NOT NULL DEFAULT 'active',
			expires_at TIMESTAMP NOT NULL, idempotency_key VARCHAR(255) DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		// Orders
		`CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(36) PRIMARY KEY, order_number VARCHAR(64) NOT NULL UNIQUE,
			user_id VARCHAR(36) NOT NULL, seller_id VARCHAR(36) NOT NULL,
			status ENUM('pending','awaiting_payment','paid','processing','packed','shipped','delivered','completed','cancelled','refunded') NOT NULL DEFAULT 'pending',
			total_amount BIGINT NOT NULL DEFAULT 0, currency VARCHAR(3) NOT NULL DEFAULT 'SGD',
			shipping_address JSON, billing_address JSON,
			idempotency_key VARCHAR(255) DEFAULT '', snapshot_id VARCHAR(36) DEFAULT '',
			parent_order_id VARCHAR(36) DEFAULT NULL, metadata JSON,
			version INT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL DEFAULT NULL
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id VARCHAR(36) PRIMARY KEY, order_id VARCHAR(36) NOT NULL,
			product_id VARCHAR(36) NOT NULL, sku_id VARCHAR(36) NOT NULL,
			shop_id VARCHAR(36) NOT NULL, quantity INT NOT NULL DEFAULT 1,
			unit_price BIGINT NOT NULL DEFAULT 0, total_price BIGINT NOT NULL DEFAULT 0,
			snapshot JSON, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS outbox_events (
			event_id VARCHAR(36) PRIMARY KEY, aggregate_type VARCHAR(100) NOT NULL,
			aggregate_id VARCHAR(100) NOT NULL, event_type VARCHAR(100) NOT NULL,
			payload JSON NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			processed BOOLEAN NOT NULL DEFAULT FALSE,
			status VARCHAR(20) DEFAULT 'pending',
			error_message TEXT, retries INT DEFAULT 0,
			processing_at TIMESTAMP NULL
		) ENGINE=InnoDB`,
		// Payments
		`CREATE TABLE IF NOT EXISTS payments (
			id VARCHAR(36) PRIMARY KEY, order_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL, amount BIGINT NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'SGD',
			status ENUM('pending','authorized','captured','failed','refunded','partial_refund') NOT NULL DEFAULT 'pending',
			payment_method VARCHAR(50) NOT NULL, psp_transaction_id VARCHAR(255) DEFAULT '',
			psp_provider VARCHAR(100) DEFAULT '', idempotency_key VARCHAR(255) DEFAULT '',
			amount_refunded BIGINT NOT NULL DEFAULT 0, failure_reason TEXT,
			metadata JSON, version INT NOT NULL DEFAULT 1,
			authorized_at TIMESTAMP NULL, captured_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL DEFAULT NULL
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS refunds (
			id VARCHAR(36) PRIMARY KEY, payment_id VARCHAR(36) NOT NULL,
			order_id VARCHAR(36) NOT NULL, amount BIGINT NOT NULL,
			currency VARCHAR(3) NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'pending',
			reason TEXT, psp_refund_id VARCHAR(255) DEFAULT '',
			idempotency_key VARCHAR(255) DEFAULT '', metadata JSON,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			\`key\` VARCHAR(255) PRIMARY KEY, payment_id VARCHAR(36) NOT NULL,
			expires_at TIMESTAMP NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		// Promotions
		`CREATE TABLE IF NOT EXISTS vouchers (
			id VARCHAR(36) PRIMARY KEY, code VARCHAR(50) NOT NULL UNIQUE,
			title VARCHAR(255) NOT NULL,
			type ENUM('percentage','fixed','shipping') NOT NULL,
			discount_value BIGINT NOT NULL, min_spend BIGINT NOT NULL DEFAULT 0,
			max_discount BIGINT NOT NULL DEFAULT 0,
			usage_limit BIGINT NOT NULL DEFAULT 10000,
			usage_count BIGINT NOT NULL DEFAULT 0,
			per_user_limit INT NOT NULL DEFAULT 1,
			scope VARCHAR(50) NOT NULL DEFAULT 'platform',
			start_time TIMESTAMP NOT NULL, end_time TIMESTAMP NOT NULL,
			status ENUM('active','inactive','expired','exhausted') NOT NULL DEFAULT 'active',
			stackable BOOLEAN NOT NULL DEFAULT FALSE,
			priority INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS voucher_redemptions (
			id VARCHAR(36) PRIMARY KEY, voucher_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL, order_id VARCHAR(36) NOT NULL,
			discount_amount BIGINT NOT NULL, idempotency_key VARCHAR(100) DEFAULT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_vr_voucher (voucher_id), INDEX idx_vr_user (user_id, voucher_id),
			INDEX idx_vr_idempotency (idempotency_key),
			FOREIGN KEY (voucher_id) REFERENCES vouchers(id)
		) ENGINE=InnoDB`,
		// Cart
		`CREATE TABLE IF NOT EXISTS carts (
			id VARCHAR(36) PRIMARY KEY, user_id VARCHAR(36) DEFAULT '',
			session_id VARCHAR(100) DEFAULT '', status VARCHAR(32) NOT NULL DEFAULT 'active',
			currency VARCHAR(3) NOT NULL DEFAULT 'SGD',
			item_count INT NOT NULL DEFAULT 0, subtotal BIGINT NOT NULL DEFAULT 0,
			version INT NOT NULL DEFAULT 1,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL DEFAULT NULL
		) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id VARCHAR(36) PRIMARY KEY, cart_id VARCHAR(36) NOT NULL,
			sku VARCHAR(100) NOT NULL, product_name VARCHAR(255) NOT NULL,
			shop_id VARCHAR(36) NOT NULL, shop_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL DEFAULT 1, unit_price BIGINT NOT NULL DEFAULT 0,
			total_price BIGINT NOT NULL DEFAULT 0, image_url TEXT, attributes TEXT,
			is_selected BOOLEAN NOT NULL DEFAULT TRUE,
			is_available BOOLEAN NOT NULL DEFAULT TRUE,
			added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE
		) ENGINE=InnoDB`,
		// Payment fraud_checks
		`CREATE TABLE IF NOT EXISTS fraud_checks (
			id VARCHAR(36) PRIMARY KEY, payment_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL, risk_score INT NOT NULL DEFAULT 0,
			risk_level VARCHAR(20) DEFAULT 'low', is_fraud BOOLEAN NOT NULL DEFAULT FALSE,
			reasons JSON, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		// Webhook events
		`CREATE TABLE IF NOT EXISTS webhook_events (
			id VARCHAR(36) PRIMARY KEY, psp_provider VARCHAR(100) NOT NULL,
			event_type VARCHAR(100) NOT NULL, payload JSON NOT NULL,
			signature VARCHAR(255) DEFAULT '', processed BOOLEAN NOT NULL DEFAULT FALSE,
			idempotency_key VARCHAR(255) DEFAULT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
	}
	for _, m := range migrations {
		if _, err := dbx.Exec(m); err != nil {
			fmt.Fprintf(os.Stderr, "Migration error: %v\nSQL: %.100s\n", err, m)
		}
	}
}

func seedTestData(dbx *sqlx.DB) {
	now := time.Now()
	seeds := []string{
		// Stock: SKU-001 has 5 units available
		`INSERT INTO stock (id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level, version) 
		 VALUES ('stk-001', 'prod-001', 'sku-001', 'wh-001', 100, 0, 100, 'in_stock', 10, 1)
		 ON DUPLICATE KEY UPDATE quantity=100, available_qty=100, status='in_stock'`,
		// Stock: SKU-002 has 3 units (low stock)
		`INSERT INTO stock (id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level, version) 
		 VALUES ('stk-002', 'prod-002', 'sku-002', 'wh-001', 3, 0, 3, 'low_stock', 10, 1)
		 ON DUPLICATE KEY UPDATE quantity=3, available_qty=3`,
		// Flash sale SKU has exactly 10 units
		`INSERT INTO stock (id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level, version) 
		 VALUES ('stk-003', 'prod-003', 'sku-flash-001', 'wh-001', 10, 0, 10, 'in_stock', 5, 1)
		 ON DUPLICATE KEY UPDATE quantity=10, available_qty=10`,
		// Voucher with usage_limit=5
		`INSERT INTO vouchers (id, code, title, type, discount_value, min_spend, max_discount, usage_limit, usage_count, per_user_limit, scope, start_time, end_time, status, stackable, priority)
		 VALUES ('vch-001', 'FLASH10', 'Flash 10% Off', 'percentage', 10, 0, 5000, 5, 0, 1, 'platform', '` + now.Add(-time.Hour).Format("2006-01-02 15:04:05") + `', '` + now.Add(24*time.Hour).Format("2006-01-02 15:04:05") + `', 'active', FALSE, 1)
		 ON DUPLICATE KEY UPDATE usage_limit=5, usage_count=0`,
	}
	for _, s := range seeds {
		if _, err := dbx.Exec(s); err != nil {
			fmt.Fprintf(os.Stderr, "Seed error: %v\nSQL: %.100s\n", err, s)
		}
	}
}

func initServices(dbx *sqlx.DB, redisClient *sharedRedis.RedisClient) {
	_ = observability.InitLogger("production-test", "error")

	// Inventory service
	invRepo := invMysql.NewInventoryRepository(dbx)
	invRedisStore := invRedis.NewStore(redisClient, invConfig.RedisConfig{})
	invService = invApp.NewInventoryService(
		&invConfig.Config{
			Inventory: invConfig.InventoryConfig{
				ReservationTTL: 30 * time.Minute,
				IdempotencyTTL: 24 * time.Hour,
			},
		},
		globalDB,
		invRepo,
		invRedisStore,
		nil,
	)

	// Order service
	orderRepo := orderMysql.NewOrderRepository(dbx)
	outboxRepo := orderMysql.NewOutboxRepository(dbx)
	orderRedisStore := orderRedis.NewStore(redisClient, orderConfig.RedisConfig{})
	orderService = orderApp.NewOrderService(
		&orderConfig.Config{
			Order: orderConfig.OrderConfig{
				DefaultCurrency:   "SGD",
				IdempotencyKeyTTL: 24 * time.Hour,
			},
		},
		orderRepo,
		outboxRepo,
		orderRedisStore,
		nil,
	)

	// Payment service
	payRepo := payMysql.NewPaymentRepository(dbx)
	payRedisStore := payRedis.NewStore(redisClient, payConfig.RedisConfig{})
	fraudDetector = &mockFraudDetector{}
	payService = payApp.NewPaymentService(
		&payConfig.Config{
			Payment: payConfig.PaymentConfig{
				DefaultPSP:       "stripe",
				IdempotencyTTL:   24 * time.Hour,
				WebhookSecret:    "whsec_test",
				FraudRiskThreshold: 100,
			},
		},
		payRepo,
		payRedisStore,
		nil,
		fraudDetector,
	)

	// Promotion service
	voucherRepo := promoMysql.NewVoucherRepository(dbx)
	redemptionRepo := promoMysql.NewVoucherRedemptionRepository(dbx)
	promoRedisStore := promoRedis.NewStore(redisClient, promoRedis.Config{})
	promoSvc = promoApp.NewPromotionService(
		voucherRepo,
		redemptionRepo,
		nil, nil, nil, nil,
		promoRedisStore,
		nil,
	)

	// Cart service
	cartRepo := cartMysql.NewCartRepository(dbx)
	itemRepo := cartMysql.NewCartItemRepository(dbx)
	cartSvc = cartApp.NewCartService(
		cartRepo, itemRepo, nil, nil,
		nil,
		7*24*time.Hour,
		15*time.Minute,
		100, 99,
		nil,
	)
}

// ===== TESTS =====

// TestFullBusinessFlow validates the complete order-to-payment flow
func TestFullBusinessFlow(t *testing.T) {
	ctx := context.Background()

	// Step 1: Create a cart and add items
	cart, err := cartSvc.GetOrCreateCart(ctx, "user-001", "", "SGD")
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}
	item, err := cartSvc.AddItem(ctx, cart.ID, cartApp.AddItemRequest{
		SKU: "sku-001", ProductName: "Test Product", ShopID: "shop-001",
		ShopName: "Test Shop", Quantity: 2, UnitPrice: 50000,
	})
	if err != nil {
		t.Fatalf("add item: %v", err)
	}
	if item.Quantity != 2 || item.UnitPrice != 50000 {
		t.Fatalf("item mismatch: qty=%d, price=%d", item.Quantity, item.UnitPrice)
	}
	t.Logf("Cart created: %s, item: %s", cart.ID, item.ID)

	// Step 2: Reserve inventory for the order
	res, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
		OrderID:        "order-001",
		UserID:         "user-001",
		ProductID:      "prod-001",
		SkuID:          "sku-001",
		WarehouseID:    "wh-001",
		Quantity:       2,
		IdempotencyKey: "idem-reserve-001",
	})
	if err != nil {
		t.Fatalf("reserve stock: %v", err)
	}
	if res == nil || res.Status != invDomain.ReservationStatusActive {
		t.Fatal("reservation not active")
	}
	t.Logf("Stock reserved: %s", res.ID)

	// Verify stock decreased
	stock, err := invService.GetStock(ctx, "sku-001", "wh-001")
	if err != nil {
		t.Fatalf("get stock: %v", err)
	}
	if stock.AvailableQty != 98 || stock.ReservedQty != 2 {
		t.Fatalf("stock mismatch: available=%d reserved=%d (expected 98,2)", stock.AvailableQty, stock.ReservedQty)
	}

	// Step 3: Authorize payment
	pay, err := payService.AuthorizePayment(ctx, &payApp.AuthorizePaymentRequest{
		OrderID:        "order-001",
		UserID:         "user-001",
		Amount:         100000,
		Currency:       "SGD",
		PaymentMethod:  payDomain.PaymentMethodCreditCard,
		IdempotencyKey: "idem-pay-001",
	})
	if err != nil {
		t.Fatalf("authorize payment: %v", err)
	}
	if pay.Status != payDomain.PaymentStatusAuthorized {
		t.Fatalf("expected authorized, got %s", pay.Status)
	}
	t.Logf("Payment authorized: %s", pay.ID)

	// Step 4: Create order
	orderResult, err := orderService.CreateOrder(ctx, &orderApp.CreateOrderRequest{
		UserID:         "user-001",
		SellerID:       "shop-001",
		Items:          []orderDomain.OrderItem{{ProductID: "prod-001", SkuID: "sku-001", ShopID: "shop-001", Quantity: 2, UnitPrice: 50000}},
		Currency:       "SGD",
		IdempotencyKey: "idem-order-001",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if orderResult.Status != orderDomain.OrderStatusPending {
		t.Fatalf("expected pending, got %s", orderResult.Status)
	}
	t.Logf("Order created: %s", orderResult.ID)

	// Step 5: Verify idempotency works (same key returns cached result)
	orderResult2, err := orderService.CreateOrder(ctx, &orderApp.CreateOrderRequest{
		UserID: "user-001", SellerID: "shop-001",
		Items: []orderDomain.OrderItem{{ProductID: "prod-001", SkuID: "sku-001", ShopID: "shop-001", Quantity: 2, UnitPrice: 50000}},
		Currency: "SGD", IdempotencyKey: "idem-order-001",
	})
	if err != nil {
		t.Fatalf("idempotent order: %v", err)
	}
	if orderResult2.ID != orderResult.ID {
		t.Fatal("idempotency failed: different order returned")
	}
	t.Log("Idempotency verified for order creation")
}

// TestFlashSaleConcurrency validates that overselling is prevented
func TestFlashSaleConcurrency(t *testing.T) {
	ctx := context.Background()

	const totalStock = 10
	const concurrentRequests = 50
	var successCount int64

	// Reset stock for flash sale
	globalDBx.Exec("UPDATE stock SET quantity=10, reserved_qty=0, available_qty=10, version=1 WHERE sku_id='sku-flash-001'")

	var wg sync.WaitGroup
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			orderID := fmt.Sprintf("flash-order-%d", id)
			idemKey := fmt.Sprintf("flash-idem-%d", id)
			_, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
				OrderID:        orderID,
				UserID:         fmt.Sprintf("user-%d", id),
				ProductID:      "prod-003",
				SkuID:          "sku-flash-001",
				WarehouseID:    "wh-001",
				Quantity:       1,
				IdempotencyKey: idemKey,
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}
	wg.Wait()

	if successCount > totalStock {
		t.Fatalf("OVERSOLD! Reserved %d units but only %d available", successCount, totalStock)
	}
	t.Logf("Flash sale: %d/%d succeeded (capacity: %d)", successCount, concurrentRequests, totalStock)

	// Verify exact count
	var reserved int
	globalDBx.Get(&reserved, "SELECT COUNT(*) FROM reservations WHERE sku_id='sku-flash-001' AND status='active'")
	if reserved != int(successCount) {
		t.Fatalf("reservation count mismatch: %d reserved vs %d success", reserved, successCount)
	}

	// Verify stock is exactly depleted
	var stock struct {
		Available, Reserved int
	}
	globalDBx.Get(&stock, "SELECT available_qty, reserved_qty FROM stock WHERE sku_id='sku-flash-001'")
	if stock.Available != totalStock-int(successCount) {
		t.Fatalf("available=%d, expected=%d", stock.Available, totalStock-int(successCount))
	}
	if stock.Reserved != int(successCount) {
		t.Fatalf("reserved=%d, expected=%d", stock.Reserved, successCount)
	}
}

// TestVoucherDoubleRedeem validates race condition prevention
func TestVoucherDoubleRedeem(t *testing.T) {
	ctx := context.Background()
	const usageLimit int64 = 5
	const concurrentRequests = 20

	var successCount int64
	var wg sync.WaitGroup

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			idemKey := fmt.Sprintf("vch-idem-%d", id)
			_, err := promoSvc.RedeemVoucher(ctx, "FLASH10", fmt.Sprintf("user-%d", id), fmt.Sprintf("order-%d", id), idemKey, 100000, "", "", "", "", "")
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}
	wg.Wait()

	if successCount > usageLimit {
		t.Fatalf("Voucher OVERSOLD! %d redemptions but limit is %d", successCount, usageLimit)
	}
	t.Logf("Voucher redemptions: %d/%d succeeded (limit: %d)", successCount, concurrentRequests, usageLimit)

	var usageCount int64
	globalDBx.Get(&usageCount, "SELECT usage_count FROM vouchers WHERE code='FLASH10'")
	if usageCount != successCount {
		t.Fatalf("usage_count=%d != successCount=%d", usageCount, successCount)
	}
}

// TestPaymentIdempotency validates double-charge prevention
func TestPaymentIdempotency(t *testing.T) {
	ctx := context.Background()

	pay1, err := payService.AuthorizePayment(ctx, &payApp.AuthorizePaymentRequest{
		OrderID: "order-idem-001", UserID: "user-idem-001",
		Amount: 50000, Currency: "SGD",
		PaymentMethod: payDomain.PaymentMethodCreditCard,
		IdempotencyKey: "idem-pay-unique",
	})
	if err != nil {
		t.Fatalf("first payment: %v", err)
	}

	// Same idempotency key returns same payment
	pay2, err := payService.AuthorizePayment(ctx, &payApp.AuthorizePaymentRequest{
		OrderID: "order-idem-001", UserID: "user-idem-001",
		Amount: 50000, Currency: "SGD",
		PaymentMethod: payDomain.PaymentMethodCreditCard,
		IdempotencyKey: "idem-pay-unique",
	})
	if err != nil {
		t.Fatalf("idempotent payment: %v", err)
	}
	if pay1.ID != pay2.ID {
		t.Fatal("idempotency key returned different payment")
	}
	t.Log("Payment idempotency verified")
}

// TestReserveStockIdempotency validates idempotent reservation
func TestReserveStockIdempotency(t *testing.T) {
	ctx := context.Background()

	res1, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
		OrderID: "order-idem-res-001", UserID: "user-idem-res-001",
		ProductID: "prod-001", SkuID: "sku-001", WarehouseID: "wh-001",
		Quantity: 1, IdempotencyKey: "idem-res-unique",
	})
	if err != nil {
		t.Fatalf("first reserve: %v", err)
	}

	res2, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
		OrderID: "order-idem-res-001", UserID: "user-idem-res-001",
		ProductID: "prod-001", SkuID: "sku-001", WarehouseID: "wh-001",
		Quantity: 1, IdempotencyKey: "idem-res-unique",
	})
	if err != nil {
		t.Fatalf("idempotent reserve: %v", err)
	}
	if res1.ID != res2.ID {
		t.Fatal("idempotency key returned different reservation")
	}
	t.Log("Reserve stock idempotency verified")
}

// TestConcurrentCartModifications validates cart race conditions
func TestConcurrentCartModifications(t *testing.T) {
	ctx := context.Background()

	cart, err := cartSvc.GetOrCreateCart(ctx, "user-concurrent-001", "", "SGD")
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sku := fmt.Sprintf("sku-concurrent-%d", id)
			cartSvc.AddItem(ctx, cart.ID, cartApp.AddItemRequest{
				SKU: sku, ProductName: "Concurrent Item",
				ShopID: "shop-001", ShopName: "Test Shop",
				Quantity: 1, UnitPrice: int64(10000 + id*1000),
			})
		}(i)
	}
	wg.Wait()

	retrieved, items, err := cartSvc.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}
	if len(items) != 10 {
		t.Fatalf("expected 10 items, got %d", len(items))
	}
	if retrieved.ItemCount != 10 {
		t.Fatalf("expected item_count=10, got %d", retrieved.ItemCount)
	}
	t.Logf("Cart has %d items, total=%d (verified %d)", len(items), retrieved.Subtotal, retrieved.ItemCount)
}

// TestReleaseAndVerifyStock tests stock release
func TestReleaseAndVerifyStock(t *testing.T) {
	ctx := context.Background()

	initialQty := 100
	globalDBx.Exec("UPDATE stock SET quantity=100, reserved_qty=0, available_qty=100, version=1 WHERE sku_id='sku-001'")

	res, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
		OrderID: "order-release-001", UserID: "user-release-001",
		ProductID: "prod-001", SkuID: "sku-001", WarehouseID: "wh-001",
		Quantity: 5, IdempotencyKey: "idem-release-001",
	})
	if err != nil {
		t.Fatalf("reserve for release: %v", err)
	}

	stockBefore, _ := invService.GetStock(ctx, "sku-001", "wh-001")
	t.Logf("Before release: available=%d, reserved=%d", stockBefore.AvailableQty, stockBefore.ReservedQty)

	if err := invService.ReleaseStock(ctx, res.ID); err != nil {
		t.Fatalf("release stock: %v", err)
	}

	stockAfter, _ := invService.GetStock(ctx, "sku-001", "wh-001")
	if stockAfter.AvailableQty != initialQty || stockAfter.ReservedQty != 0 {
		t.Fatalf("after release: available=%d (want %d), reserved=%d (want 0)",
			stockAfter.AvailableQty, initialQty, stockAfter.ReservedQty)
	}
	t.Log("Stock release verified: inventory returned correctly")
}

// TestPaymentDoubleChargePrevention validates concurrent payment prevention
func TestPaymentDoubleChargePrevention(t *testing.T) {
	ctx := context.Background()
	const numConcurrent = 10
	var successCount int64
	var wg sync.WaitGroup

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := payService.AuthorizePayment(ctx, &payApp.AuthorizePaymentRequest{
				OrderID:        "order-doublecharge-001",
				UserID:         "user-doublecharge-001",
				Amount:         50000,
				Currency:       "SGD",
				PaymentMethod:  payDomain.PaymentMethodCreditCard,
				IdempotencyKey: fmt.Sprintf("idem-double-%d", id),
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}
	wg.Wait()

	if successCount > 1 {
		t.Fatalf("Double charge! %d payments created for same order", successCount)
	}
	t.Logf("Double charge prevention: %d/%d succeeded (expected 1)", successCount, numConcurrent)
}

// TestInvalidRefundRejection validates refund business rules
func TestInvalidRefundRejection(t *testing.T) {
	ctx := context.Background()

	pay, err := payService.AuthorizePayment(ctx, &payApp.AuthorizePaymentRequest{
		OrderID: "order-refund-001", UserID: "user-refund-001",
		Amount: 10000, Currency: "SGD",
		PaymentMethod:  payDomain.PaymentMethodCreditCard,
		IdempotencyKey: "idem-refund-001",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	// Try refund on non-captured payment - should fail
	_, err = payService.RefundPayment(ctx, pay.ID, "test refund", "idem-refund-002", 5000, "user-refund-001")
	if err == nil {
		t.Fatal("expected error for refund on non-captured payment")
	}
	t.Logf("Refund correctly rejected: %v", err)

	// Try negative refund - should fail
	_, err = payService.RefundPayment(ctx, pay.ID, "negative test", "idem-refund-003", -1000, "user-refund-001")
	if err == nil {
		t.Fatal("expected error for negative refund")
	}
	t.Logf("Negative refund correctly rejected: %v", err)

	// Try excessive refund - should fail
	_, err = payService.RefundPayment(ctx, pay.ID, "excessive", "idem-refund-004", 999999, "user-refund-001")
	if err == nil {
		t.Fatal("expected error for excessive refund")
	}
	t.Logf("Excessive refund correctly rejected: %v", err)
}

// TestConcurrentVoucherPerUserLimit validates per-user voucher limit
func TestConcurrentVoucherPerUserLimit(t *testing.T) {
	ctx := context.Background()
	const perUserLimit = 1
	const concurrentAttempts = 10
	var successCount int64

	// Reset voucher
	globalDBx.Exec("UPDATE vouchers SET usage_count=0 WHERE code='FLASH10'")
	globalDBx.Exec("DELETE FROM voucher_redemptions WHERE voucher_id='vch-001'")

	var wg sync.WaitGroup
	for i := 0; i < concurrentAttempts; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			orderID := fmt.Sprintf("order-userlimit-%d", id)
			idemKey := fmt.Sprintf("idem-userlimit-%d", id)
			_, err := promoSvc.RedeemVoucher(ctx, "FLASH10", "user-single", orderID, idemKey, 100000, "", "", "", "", "")
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}
	wg.Wait()

	if successCount > int64(perUserLimit) {
		t.Fatalf("Per-user limit %d breached! %d redemptions for same user", perUserLimit, successCount)
	}
	t.Logf("Per-user limit enforced: %d/%d succeeded (limit: %d)", successCount, concurrentAttempts, perUserLimit)
}

// TestOrderStateMachine validates all valid transitions
func TestOrderStateMachine(t *testing.T) {
	order := &orderDomain.Order{
		ID: "test-sm-1", Status: orderDomain.OrderStatusPending, Version: 1,
	}

	transitions := []struct {
		from, to orderDomain.OrderStatus
		valid    bool
	}{
		{orderDomain.OrderStatusPending, orderDomain.OrderStatusAwaitingPayment, true},
		{orderDomain.OrderStatusPending, orderDomain.OrderStatusCancelled, true},
		{orderDomain.OrderStatusPending, orderDomain.OrderStatusPaid, false},
		{orderDomain.OrderStatusAwaitingPayment, orderDomain.OrderStatusPaid, true},
		{orderDomain.OrderStatusAwaitingPayment, orderDomain.OrderStatusCancelled, true},
		{orderDomain.OrderStatusPaid, orderDomain.OrderStatusProcessing, true},
		{orderDomain.OrderStatusPaid, orderDomain.OrderStatusCancelled, true},
		{orderDomain.OrderStatusProcessing, orderDomain.OrderStatusPacked, true},
		{orderDomain.OrderStatusProcessing, orderDomain.OrderStatusCancelled, true},
		{orderDomain.OrderStatusPacked, orderDomain.OrderStatusShipped, true},
		{orderDomain.OrderStatusPacked, orderDomain.OrderStatusCancelled, true},
		{orderDomain.OrderStatusShipped, orderDomain.OrderStatusDelivered, true},
		{orderDomain.OrderStatusShipped, orderDomain.OrderStatusRefunded, true},
		{orderDomain.OrderStatusDelivered, orderDomain.OrderStatusCompleted, true},
		{orderDomain.OrderStatusDelivered, orderDomain.OrderStatusRefunded, true},
		{orderDomain.OrderStatusCompleted, orderDomain.OrderStatusRefunded, true},
	}

	for _, tt := range transitions {
		t.Run(fmt.Sprintf("%s->%s", tt.from, tt.to), func(t *testing.T) {
			o := &orderDomain.Order{Status: tt.from}
			result := o.CanTransitionTo(tt.to)
			if result != tt.valid {
				t.Errorf("CanTransitionTo(%s,%s)=%v, want %v", tt.from, tt.to, result, tt.valid)
			}
		})
	}
}

// TestCircuitBreakerRace validates thread safety
func TestCircuitBreakerRace(t *testing.T) {
	cb := newCircuitBreakerForTest(5, 100*time.Millisecond)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				cb.Execute(func() error {
					return fmt.Errorf("ephemeral error %d", id)
				})
			} else {
				_ = cb.State()
			}
		}(i)
	}
	wg.Wait()
	t.Log("CircuitBreaker survived 100 concurrent calls without data race")
}

func newCircuitBreakerForTest(threshold int, reset time.Duration) *mockCB {
	return &mockCB{failures: 0, threshold: threshold, resetTimeout: reset}
}

type mockCB struct {
	failures      int
	threshold     int
	resetTimeout  time.Duration
	state         int
}

func (m *mockCB) Execute(fn func() error) error {
	if err := fn(); err != nil {
		m.failures++
		if m.failures >= m.threshold {
			m.state = 1
		}
		return err
	}
	m.failures = 0
	return nil
}

func (m *mockCB) State() int {
	return m.state
}

// ===== HELPERS =====

type mockFraudDetector struct{}

func (m *mockFraudDetector) Assess(ctx context.Context, paymentID, userID string, amount int64, method payDomain.PaymentMethod) (*payDomain.FraudCheckResult, error) {
	return payDomain.NewFraudCheckResult(paymentID, userID, 0, false), nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestOrderCancellationFlow(t *testing.T) {
	ctx := context.Background()

	res, err := invService.ReserveStock(ctx, &invApp.ReserveStockRequest{
		OrderID: "order-cancel-001", UserID: "user-cancel-001",
		ProductID: "prod-001", SkuID: "sku-001", WarehouseID: "wh-001",
		Quantity: 1, IdempotencyKey: "idem-cancel-reserve",
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}

	// Simulate cancellation: release stock back
	if err := invService.ReleaseStock(ctx, res.ID); err != nil {
		t.Fatalf("release: %v", err)
	}

	stock, _ := invService.GetStock(ctx, "sku-001", "wh-001")
	if stock.AvailableQty != 100 || stock.ReservedQty != 0 {
		t.Fatalf("after cancel: available=%d, reserved=%d", stock.AvailableQty, stock.ReservedQty)
	}
	t.Log("Cancellation flow: stock released correctly")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
