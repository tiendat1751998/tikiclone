-- ============================================================
-- PERFORMANCE OPTIMIZATION: Additional indexes
-- Created: 2025-06-04
-- ============================================================

-- Cart snapshots: composite index for cart_id + created_at (used in ORDER BY created_at DESC LIMIT 1)
CREATE INDEX IF NOT EXISTS idx_cs_cart_created ON cart_snapshots (cart_id, created_at);

-- Products: index for brand-based filtering
CREATE INDEX IF NOT EXISTS idx_prod_brand ON products (brand);

-- Orders: composite index for user order history queries
CREATE INDEX IF NOT EXISTS idx_ord_user_created ON orders (user_id, created_at DESC);

-- Order items: composite index for order lookups
CREATE INDEX IF NOT EXISTS idx_oi_order_created ON order_items (order_id, created_at);

-- SKUs: index for price range queries
CREATE INDEX IF NOT EXISTS idx_sku_price ON skus (price);

-- Categories: composite index for active category listing
CREATE INDEX IF NOT EXISTS idx_cat_active_parent ON categories (is_active, parent_id, sort_order);
