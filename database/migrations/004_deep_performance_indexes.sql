-- ============================================================
-- Migration 004: Deep Performance Indexes & Query Optimizations
-- Created: 2026-02-24
-- Targets: tiki_platform (MySQL 8.0)
--
-- APPLY THIS MIGRATION:
--   mysql -u tiki -p tiki_platform < database/migrations/004_deep_performance_indexes.sql
--
-- VERIFY INDEX USAGE:
--   EXPLAIN SELECT ... (check that type=ref/range, Extra=Using index)
-- ============================================================

-- ============================================================
-- 1. PRODUCTS — Hot path: search, listing, category filter
-- ============================================================

-- Covers: WHERE status='active' AND deleted_at IS NULL ORDER BY created_at DESC
-- Used by: /api/products/featured, /api/products/search
-- This replaces the need for MySQL to scan the full table
ALTER TABLE products ADD INDEX idx_prod_active_created (status, deleted_at, created_at DESC);

-- Covers: WHERE category_id=? AND status='active' AND deleted_at IS NULL ORDER BY created_at DESC  
-- Used by: category product listing page
ALTER TABLE products ADD INDEX idx_prod_cat_created (category_id, status, deleted_at, created_at DESC);

-- Covers: WHERE shop_id=? AND deleted_at IS NULL ORDER BY created_at DESC
-- Used by: shop page product listing
ALTER TABLE products ADD INDEX idx_prod_shop_created (shop_id, deleted_at, created_at DESC);

-- Covers: WHERE status='active' AND deleted_at IS NULL AND category_id=? (for category filter)
-- Smaller than idx_prod_cat_created when you don't need sorting
ALTER TABLE products ADD INDEX idx_prod_active_cat (status, deleted_at, category_id);

-- ============================================================
-- 2. SKUS — Price filtering, sorting, stock checks
-- ============================================================

-- Covers: WHERE product_id=? ORDER BY price ASC
-- Used by: product detail page (get cheapest SKU first)
ALTER TABLE skus ADD INDEX idx_sku_product_price (product_id, price);

-- Covers: price range queries with product reference
-- Used by: search with min_price/max_price filter
ALTER TABLE skus ADD INDEX idx_sku_price_range (sale_price, product_id);

-- Covers: WHERE product_id=? AND status='active' (stock check during checkout)
ALTER TABLE skus ADD INDEX idx_sku_product_status (product_id, status, stock);

-- ============================================================
-- 3. PRODUCT_MEDIA — Image loading
-- ============================================================

-- Covers: WHERE product_id=? AND is_primary=1 ORDER BY sort_order
-- Used by: product listings (only need 1 primary image per product)
ALTER TABLE product_media ADD INDEX idx_pm_primary_image (product_id, is_primary, sort_order);

-- ============================================================
-- 4. CATEGORIES — Tree navigation
-- ============================================================

-- Covers: WHERE is_active=1 ORDER BY level, sort_order, name
-- Used by: full category tree fetch (categories API endpoint)
ALTER TABLE categories ADD INDEX idx_cat_tree (is_active, level, sort_order);

-- Covers: WHERE parent_id=? AND is_active=1 ORDER BY sort_order
-- Used by: subcategories of a parent
ALTER TABLE categories ADD INDEX idx_cat_parent_sort (parent_id, is_active, sort_order);

-- ============================================================
-- 5. CARTS — Cart lookups
-- ============================================================

-- Covering index for: WHERE user_id=? AND status='active' AND deleted_at IS NULL
-- Avoids table lookup entirely
ALTER TABLE carts ADD INDEX idx_cart_user_lookup (user_id, status, deleted_at);

-- Covering index for: WHERE session_id=? AND status='active' AND deleted_at IS NULL  
ALTER TABLE carts ADD INDEX idx_cart_session_lookup (session_id, status, deleted_at);

-- ============================================================
-- 6. CART_ITEMS — Cart item queries
-- ============================================================

-- Covers: WHERE cart_id=? ORDER BY added_at DESC
ALTER TABLE cart_items ADD INDEX idx_ci_cart_added (cart_id, added_at DESC);

-- ============================================================
-- 7. ORDERS — Order history & lookups
-- ============================================================

-- Covers: WHERE user_id=? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?
ALTER TABLE orders ADD INDEX idx_orders_user_recent (user_id, deleted_at, created_at DESC);

-- Covers: WHERE seller_id=? AND deleted_at IS NULL ORDER BY created_at DESC
ALTER TABLE orders ADD INDEX idx_orders_seller_recent (seller_id, deleted_at, created_at DESC);

-- ============================================================
-- 8. ORDER_ITEMS — Order detail lookups
-- ============================================================

-- Covers: WHERE product_id=? (product sales analytics)
ALTER TABLE order_items ADD INDEX idx_oi_product_date (product_id, created_at DESC);

-- ============================================================
-- 9. CHECKOUTS — Checkout flow
-- ============================================================

-- Covers: WHERE idempotency_key=? (idempotency check)
ALTER TABLE checkouts ADD INDEX idx_ck_idem (idempotency_key);

-- Covers: WHERE status IN (...) AND expires_at < ? (expired cleanup)
ALTER TABLE checkouts ADD INDEX idx_ck_expiry (status, expires_at);

-- ============================================================
-- 10. INVENTORY — Stock & reservation queries
-- ============================================================

-- Covers: WHERE product_id=? (stock check during checkout)
ALTER TABLE stock ADD INDEX idx_stock_product_avail (product_id, available_qty);

-- Covers: WHERE status='active' AND expires_at < NOW() (cleanup)
ALTER TABLE reservations ADD INDEX idx_res_cleanup (status, expires_at);

-- ============================================================
-- 11. REVIEWS — Product reviews
-- ============================================================

-- Covers: WHERE product_id=? AND status='approved' ORDER BY created_at DESC
ALTER TABLE tiki_reviews ADD INDEX idx_rev_product_created (product_id, status, created_at DESC);

-- ============================================================
-- 12. DEALS — Flash sale lookups
-- ============================================================

-- Covers: WHERE deal_id=? AND is_active=1 ORDER BY deal_price ASC
ALTER TABLE tiki_deal_products ADD INDEX idx_tdp_deal_price (deal_id, is_active, deal_price);

-- ============================================================
-- 13. RECOMMENDATIONS — Fast recommendation lookups
-- ============================================================

-- Covers: WHERE user_id=? ORDER BY created_at DESC LIMIT ?
ALTER TABLE recommendation_events ADD INDEX idx_re_user_recent (user_id, created_at DESC);

-- ============================================================
-- 14. NOTIFICATIONS — User notification feed
-- ============================================================

-- Covers: WHERE user_id=? ORDER BY created_at DESC LIMIT ?
ALTER TABLE notifications ADD INDEX idx_notif_user_recent (user_id, created_at DESC);

-- ============================================================
-- 15. ANALYZE — Update optimizer statistics
-- ============================================================

ANALYZE TABLE products;
ANALYZE TABLE skus;
ANALYZE TABLE product_media;
ANALYZE TABLE categories;
ANALYZE TABLE carts;
ANALYZE TABLE cart_items;
ANALYZE TABLE orders;
ANALYZE TABLE order_items;
ANALYZE TABLE checkouts;
ANALYZE TABLE stock;
ANALYZE TABLE reservations;
ANALYZE TABLE tiki_reviews;
ANALYZE TABLE tiki_deals;
ANALYZE TABLE tiki_deal_products;
ANALYZE TABLE notifications;
