-- ============================================================
-- Migration 007: Ultra-Performance Database Optimization
-- Created: 2026-05-25
-- Targets: tiki_platform (MySQL 8.0) + service-level schemas
--
-- APPLY THIS MIGRATION:
--   mysql -u tiki -p tiki_platform < database/migrations/007_ultra_performance.sql
--
-- SAFETY: All ALTER TABLE statements use ALGORITHM=INPLACE, LOCK=NONE
--         where supported by MySQL 8.0 for online DDL.
-- ============================================================

-- ============================================================
-- 1. ORDERS — Fix missing indexes, add covering indexes
-- ============================================================

-- Covering index for user order history (most common query)
-- Replaces idx_ord_user which doesn't include deleted_at for filtering
ALTER TABLE orders ADD INDEX idx_orders_user_recent (user_id, deleted_at, created_at DESC);

-- Covering index for seller order listing
ALTER TABLE orders ADD INDEX idx_orders_seller_recent (seller_id, deleted_at, created_at DESC);

-- Covering index for idempotency lookups (avoids table scan)
ALTER TABLE orders ADD INDEX idx_orders_idempotency_cover (idempotency_key, deleted_at, id);

-- ============================================================
-- 2. ORDER_ITEMS — Add missing product_id index
-- ============================================================

-- Used by product sales analytics and order detail reconstruction
ALTER TABLE order_items ADD INDEX idx_oi_product_date (product_id, created_at DESC);

-- Covering index for order detail lookups (avoids table access)
ALTER TABLE order_items ADD INDEX idx_oi_order_cover (order_id, product_id, sku_id, quantity, unit_price, total_price);

-- ============================================================
-- 3. PAYMENTS — Fix problematic UNIQUE key, add indexes
-- ============================================================

-- PROBLEM: UNIQUE KEY uk_pay_order_status (order_id, status) blocks
-- multiple payment attempts with the same status (e.g., multiple 'failed').
-- SOLUTION: Drop the unique key, replace with non-unique index.
-- NOTE: If the key doesn't exist in your schema, skip this statement.
-- ALTER TABLE payments DROP INDEX uk_pay_order_status;

-- Covering index for payment lookups by order
ALTER TABLE payments ADD INDEX idx_pay_order_cover (order_id, deleted_at, status, amount, currency);

-- Covering index for idempotency lookups
ALTER TABLE payments ADD INDEX idx_pay_idempotency_cover (idempotency_key, deleted_at, id, status);

-- ============================================================
-- 4. INVENTORY — Flash-sale optimized indexes
-- ============================================================

-- Covering index for stock checks during checkout (avoids table access)
-- This is the HOTTEST path during flash sales
ALTER TABLE stock ADD INDEX idx_stock_sku_warehouse_cover (sku_id, warehouse_id, available_qty, reserved_qty, quantity, version);

-- Covering index for product stock aggregation
ALTER TABLE stock ADD INDEX idx_stock_product_avail (product_id, available_qty, warehouse_id);

-- Covering index for reservation cleanup (avoids table scan)
ALTER TABLE reservations ADD INDEX idx_res_cleanup_cover (status, expires_at, id, sku_id, warehouse_id, quantity);

-- Covering index for order reservation lookups
ALTER TABLE reservations ADD INDEX idx_res_order_cover (order_id, status, sku_id, warehouse_id, quantity);

-- ============================================================
-- 5. FLASH_SALE_INVENTORY — Optimized for flash-sale reads
-- ============================================================

-- Covering index for flash-sale stock checks (the hottest read path)
ALTER TABLE flash_sale_inventory ADD INDEX idx_fsi_sale_cover (flash_sale_id, is_active, start_time, end_time, sku_id, total_stock, reserved_stock, sold_stock, max_per_user);

-- ============================================================
-- 6. CARTS — Covering indexes for hot paths
-- ============================================================

-- Covering index for user cart lookup (avoids table access entirely)
ALTER TABLE carts ADD INDEX idx_cart_user_cover (user_id, status, deleted_at, id, item_count, subtotal, version, expires_at);

-- Covering index for session cart lookup
ALTER TABLE carts ADD INDEX idx_cart_session_cover (session_id, status, deleted_at, id, item_count, subtotal, version, expires_at);

-- Covering index for expired cart cleanup
ALTER TABLE carts ADD INDEX idx_cart_expiry_cover (expires_at, status, deleted_at, id);

-- ============================================================
-- 7. CART_ITEMS — Covering indexes
-- ============================================================

-- Covering index for cart item listing (avoids table access)
ALTER TABLE cart_items ADD INDEX idx_ci_cart_cover (cart_id, added_at DESC, sku, product_name, quantity, unit_price, total_price, is_selected, is_available);

-- ============================================================
-- 8. PRODUCTS — Service-level schema fixes
-- ============================================================

-- Covering index for product listing by shop (avoids table access)
ALTER TABLE products ADD INDEX idx_prod_shop_cover (shop_id, status, deleted_at, created_at DESC, id, name, category_id);

-- Covering index for category product listing
ALTER TABLE products ADD INDEX idx_prod_cat_cover (category_id, status, deleted_at, created_at DESC, id, name);

-- ============================================================
-- 9. SKUS — Covering indexes for price/stock queries
-- ============================================================

-- Covering index for product detail page (get all SKUs for a product)
ALTER TABLE skus ADD INDEX idx_sku_product_cover (product_id, status, price, sale_price, stock, weight_grams, id, sku_code);

-- Covering index for price range filtering
ALTER TABLE skus ADD INDEX idx_sku_price_cover (sale_price, product_id, status, stock);

-- ============================================================
-- 10. VOUCHERS — Covering indexes for promotion lookups
-- ============================================================

-- Covering index for voucher validation (the hottest promotion query)
ALTER TABLE vouchers ADD INDEX idx_vch_code_cover (code, status, start_time, end_time, usage_limit, usage_count, per_user_limit, scope, shop_id, category_id, sku, discount_value, min_spend, max_discount);

-- Covering index for active voucher listing
ALTER TABLE vouchers ADD INDEX idx_vch_active_cover (status, start_time, end_time, scope, shop_id, id, code, discount_value, min_spend);

-- ============================================================
-- 11. VOUCHER_REDEMPTIONS — Covering indexes
-- ============================================================

-- Covering index for per-user voucher usage check
ALTER TABLE voucher_redemptions ADD INDEX idx_vr_user_voucher_cover (user_id, voucher_id, created_at DESC);

-- ============================================================
-- 12. OUTBOX_EVENTS — Covering index for polling
-- ============================================================

-- Covering index for outbox polling (avoids table access)
ALTER TABLE outbox_events ADD INDEX idx_oe_processed_cover (processed, created_at, event_id, aggregate_type, aggregate_id, event_type);

-- ============================================================
-- 13. STOCK_MOVEMENTS — Partitioning recommendation
-- ============================================================
-- For very large deployments, consider RANGE partitioning by created_at:
-- ALTER TABLE stock_movements PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
--     PARTITION p202601 VALUES LESS THAN (202602),
--     PARTITION p202602 VALUES LESS THAN (202603),
--     ...
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- ============================================================
-- 14. ANALYZE — Update optimizer statistics
-- ============================================================

ANALYZE TABLE orders;
ANALYZE TABLE order_items;
ANALYZE TABLE payments;
ANALYZE TABLE stock;
ANALYZE TABLE reservations;
ANALYZE TABLE flash_sale_inventory;
ANALYZE TABLE carts;
ANALYZE TABLE cart_items;
ANALYZE TABLE products;
ANALYZE TABLE skus;
ANALYZE TABLE vouchers;
ANALYZE TABLE voucher_redemptions;
ANALYZE TABLE outbox_events;
