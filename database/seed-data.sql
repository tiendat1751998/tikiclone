-- ============================================================
-- SEED DATA FOR ALL TIKICLONE DATABASES
-- ============================================================

-- ============================================================
-- 1. SHOPEE_AUTH - Users, Roles, Permissions
-- ============================================================
USE tiki_auth;

-- Insert admin user (password: Admin@123)
INSERT IGNORE INTO users (user_id, email, phone, password_hash, full_name, role, is_verified, is_active) VALUES
('usr-001', 'admin@tiki.com', '+6590001001', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Super Admin', 'SUPER_ADMIN', 1, 1),
('usr-002', 'seller@tiki.com', '+6590001002', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Demo Seller', 'SELLER', 1, 1),
('usr-003', 'buyer@tiki.com', '+6590001003', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Demo Buyer', 'BUYER', 1, 1),
('usr-004', 'john.doe@email.com', '+6590001004', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'John Doe', 'BUYER', 1, 1),
('usr-005', 'jane.smith@email.com', '+6590001005', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Jane Smith', 'SELLER', 1, 1);

-- Assign roles to users
INSERT IGNORE INTO user_roles (user_id, role_id) VALUES
('usr-001', (SELECT role_id FROM roles WHERE name = 'SUPER_ADMIN')),
('usr-002', (SELECT role_id FROM roles WHERE name = 'SELLER')),
('usr-003', (SELECT role_id FROM roles WHERE name = 'BUYER')),
('usr-004', (SELECT role_id FROM roles WHERE name = 'BUYER')),
('usr-005', (SELECT role_id FROM roles WHERE name = 'SELLER'));

-- Assign permissions to ADMIN role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM roles r, permissions p
WHERE r.name = 'ADMIN'
  AND p.resource IN ('users', 'products', 'orders', 'payments');

-- Assign permissions to SELLER role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM roles r, permissions p
WHERE r.name = 'SELLER'
  AND ((p.resource = 'products' AND p.action IN ('read', 'write'))
    OR (p.resource = 'orders' AND p.action IN ('read'))
    OR (p.resource = 'inventory' AND p.action IN ('read', 'write')));

-- Assign permissions to BUYER role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.role_id, p.permission_id
FROM roles r, permissions p
WHERE r.name = 'BUYER'
  AND ((p.resource = 'products' AND p.action = 'read')
    OR (p.resource = 'orders' AND p.action IN ('read', 'write', 'cancel')));

-- ============================================================
-- 2. SHOPEE_PLATFORM - Users (mirror of auth for Go services)
-- ============================================================
USE tiki_platform;

INSERT IGNORE INTO users (id, email, phone, username, password_hash, display_name, avatar_url, status, email_verified, phone_verified) VALUES
('usr-001', 'admin@tiki.com', '+6590001001', 'admin', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Super Admin', '/images/avatars/default-avatar.png', 'active', 1, 1),
('usr-002', 'seller@tiki.com', '+6590001002', 'seller1', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Demo Seller', '/images/avatars/default-avatar.png', 'active', 1, 1),
('usr-003', 'buyer@tiki.com', '+6590001003', 'buyer1', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Demo Buyer', '/images/avatars/default-avatar.png', 'active', 1, 1),
('usr-004', 'john.doe@email.com', '+6590001004', 'johndoe', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'John Doe', '/images/avatars/default-avatar.png', 'active', 1, 1),
('usr-005', 'jane.smith@email.com', '+6590001005', 'janesmith', '$2a$12$LJ3m4ys3Lk8nFgQOIc/MNOxHBMGxPJsGK5bHJKzCQvGJmZ3mFJXi', 'Jane Smith', '/images/avatars/default-avatar.png', 'active', 1, 1);

-- ============================================================
-- 3. SHOPEE_PRODUCT - Categories, Products, SKUs
-- ============================================================
USE tiki_product;

-- Categories (Electronics)
INSERT INTO categories (category_id, name, slug, parent_id, level, sort_order, image_url, is_active) VALUES
('cat-001', 'Electronics', 'electronics', NULL, 1, 1, '/images/categories/default-category.png', 1),
('cat-002', 'Mobile Phones', 'mobile-phones', 'cat-001', 2, 1, '/images/categories/default-category.png', 1),
('cat-003', 'Laptops', 'laptops', 'cat-001', 2, 2, '/images/categories/default-category.png', 1),
('cat-004', 'Audio', 'audio', 'cat-001', 2, 3, '/images/categories/default-category.png', 1);

-- Categories (Fashion)
INSERT INTO categories (category_id, name, slug, parent_id, level, sort_order, image_url, is_active) VALUES
('cat-005', 'Fashion', 'fashion', NULL, 1, 2, '/images/categories/default-category.png', 1),
('cat-006', 'Men\'s Clothing', 'mens-clothing', 'cat-005', 2, 1, '/images/categories/default-category.png', 1),
('cat-007', 'Women\'s Clothing', 'womens-clothing', 'cat-005', 2, 2, '/images/categories/default-category.png', 1),
('cat-008', 'Shoes', 'shoes', 'cat-005', 2, 3, '/images/categories/default-category.png', 1);

-- Categories (Home & Living)
INSERT INTO categories (category_id, name, slug, parent_id, level, sort_order, image_url, is_active) VALUES
('cat-009', 'Home & Living', 'home-living', NULL, 1, 3, '/images/categories/default-category.png', 1),
('cat-010', 'Furniture', 'furniture', 'cat-009', 2, 1, '/images/categories/default-category.png', 1),
('cat-011', 'Kitchen', 'kitchen', 'cat-009', 2, 2, '/images/categories/default-category.png', 1);

-- Products
INSERT INTO products (spu_id, title, description, category_id, brand_id, seller_id, status) VALUES
('spu-001', 'iPhone 15 Pro Max 256GB', 'Latest Apple iPhone with A17 Pro chip, titanium design, 48MP camera system', 'cat-002', 'brand-apple', 'usr-002', 'ACTIVE'),
('spu-002', 'Samsung Galaxy S24 Ultra', 'Flagship Samsung phone with S Pen, 200MP camera, AI features', 'cat-002', 'brand-samsung', 'usr-002', 'ACTIVE'),
('spu-003', 'MacBook Pro 14" M3 Pro', 'Apple MacBook Pro with M3 Pro chip, 18GB RAM, 512GB SSD', 'cat-003', 'brand-apple', 'usr-005', 'ACTIVE'),
('spu-004', 'Sony WH-1000XM5 Headphones', 'Industry-leading noise cancellation, 30hr battery, premium sound', 'cat-004', 'brand-sony', 'usr-002', 'ACTIVE'),
('spu-005', 'Nike Air Max 270', 'Men\'s running shoes with Air Max cushioning, breathable mesh upper', 'cat-008', 'brand-nike', 'usr-005', 'ACTIVE'),
('spu-006', 'Adidas Ultraboost 22', 'Women\'s running shoes with Boost midsole, Primeknit upper', 'cat-008', 'brand-adidas', 'usr-002', 'ACTIVE'),
('spu-007', 'Ergonomic Office Chair', 'Adjustable lumbar support, breathable mesh back, 360° swivel', 'cat-010', 'brand-hermanmiller', 'usr-005', 'ACTIVE'),
('spu-008', 'Non-Stick Cookware Set', '10-piece ceramic non-stick cookware set, dishwasher safe', 'cat-011', 'brand-tefal', 'usr-002', 'ACTIVE'),
('spu-009', 'Cotton Crew Neck T-Shirt', '100% organic cotton, pre-shrunk, available in multiple colors', 'cat-006', 'brand-uniqlo', 'usr-005', 'ACTIVE'),
('spu-010', 'Floral Summer Dress', 'Lightweight floral print dress, A-line silhouette, knee length', 'cat-007', 'brand-zara', 'usr-002', 'ACTIVE');

-- SKUs
INSERT INTO skus (sku_id, spu_id, price, sale_price, stock, weight, length, width, height, status) VALUES
-- iPhone 15 Pro Max
('sku-001', 'spu-001', 179900, 169900, 50, 221, 160, 77, 8, 'ACTIVE'),
('sku-002', 'spu-001', 214900, 199900, 30, 221, 160, 77, 8, 'ACTIVE'),
-- Samsung Galaxy S24 Ultra
('sku-003', 'spu-002', 159900, 149900, 75, 233, 163, 80, 9, 'ACTIVE'),
('sku-004', 'spu-002', 189900, 179900, 40, 233, 163, 80, 9, 'ACTIVE'),
-- MacBook Pro
('sku-005', 'spu-003', 249900, 239900, 20, 1600, 315, 220, 16, 'ACTIVE'),
('sku-006', 'spu-003', 309900, 289900, 15, 1600, 315, 220, 16, 'ACTIVE'),
-- Sony Headphones
('sku-007', 'spu-004', 39900, 34900, 100, 250, 200, 180, 80, 'ACTIVE'),
-- Nike Air Max
('sku-008', 'spu-005', 18900, 15900, 200, 350, 320, 120, 110, 'ACTIVE'),
('sku-009', 'spu-005', 18900, 15900, 150, 340, 310, 120, 110, 'ACTIVE'),
-- Adidas Ultraboost
('sku-010', 'spu-006', 21900, 18900, 180, 310, 300, 110, 100, 'ACTIVE'),
-- Office Chair
('sku-011', 'spu-007', 89900, 79900, 25, 18000, 680, 680, 1100, 'ACTIVE'),
-- Cookware Set
('sku-012', 'spu-008', 12900, 9900, 300, 5000, 400, 300, 200, 'ACTIVE'),
-- T-Shirt
('sku-013', 'spu-009', 1990, 1490, 500, 200, 300, 250, 20, 'ACTIVE'),
('sku-014', 'spu-009', 1990, 1490, 400, 200, 300, 250, 20, 'ACTIVE'),
-- Summer Dress
('sku-015', 'spu-010', 4990, 3990, 250, 300, 350, 250, 30, 'ACTIVE'),
('sku-016', 'spu-010', 4990, 3990, 180, 300, 350, 250, 30, 'ACTIVE');

-- Product images
INSERT INTO product_images (spu_id, url, alt_text, sort_order, is_primary) VALUES
('spu-001', '/images/products/product-1.jpg', 'iPhone 15 Pro Max', 1, 1),
('spu-001', '/images/products/product-1.jpg', 'iPhone 15 Pro Max Back', 2, 0),
('spu-002', '/images/products/product-2.jpg', 'Samsung Galaxy S24 Ultra', 1, 1),
('spu-003', '/images/products/product-3.jpg', 'MacBook Pro 14', 1, 1),
('spu-004', '/images/products/product-4.jpg', 'Sony WH-1000XM5', 1, 1),
('spu-005', '/images/products/product-5.jpg', 'Nike Air Max 270', 1, 1),
('spu-006', '/images/products/product-6.jpg', 'Adidas Ultraboost 22', 1, 1),
('spu-007', '/images/products/product-7.jpg', 'Ergonomic Office Chair', 1, 1),
('spu-008', '/images/products/product-8.jpg', 'Non-Stick Cookware Set', 1, 1),
('spu-009', '/images/products/product-9.jpg', 'Cotton Crew Neck T-Shirt', 1, 1),
('spu-010', '/images/products/product-10.jpg', 'Floral Summer Dress', 1, 1);

-- ============================================================
-- 4. SHOPEE_INVENTORY - Stock
-- ============================================================
USE tiki_inventory;

INSERT INTO stock (id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level) VALUES
('inv-001', 'spu-001', 'sku-001', 'wh-001', 50, 0, 50, 'in_stock', 10),
('inv-002', 'spu-001', 'sku-002', 'wh-001', 30, 0, 30, 'in_stock', 10),
('inv-003', 'spu-002', 'sku-003', 'wh-001', 75, 0, 75, 'in_stock', 15),
('inv-004', 'spu-002', 'sku-004', 'wh-001', 40, 0, 40, 'in_stock', 10),
('inv-005', 'spu-003', 'sku-005', 'wh-001', 20, 0, 20, 'in_stock', 5),
('inv-006', 'spu-003', 'sku-006', 'wh-001', 15, 0, 15, 'in_stock', 5),
('inv-007', 'spu-004', 'sku-007', 'wh-001', 100, 0, 100, 'in_stock', 20),
('inv-008', 'spu-005', 'sku-008', 'wh-002', 200, 0, 200, 'in_stock', 30),
('inv-009', 'spu-005', 'sku-009', 'wh-002', 150, 0, 150, 'in_stock', 30),
('inv-010', 'spu-006', 'sku-010', 'wh-002', 180, 0, 180, 'in_stock', 25),
('inv-011', 'spu-007', 'sku-011', 'wh-002', 25, 0, 25, 'in_stock', 5),
('inv-012', 'spu-008', 'sku-012', 'wh-002', 300, 0, 300, 'in_stock', 50),
('inv-013', 'spu-009', 'sku-013', 'wh-002', 500, 0, 500, 'in_stock', 100),
('inv-014', 'spu-009', 'sku-014', 'wh-002', 400, 0, 400, 'in_stock', 80),
('inv-015', 'spu-010', 'sku-015', 'wh-002', 250, 0, 250, 'in_stock', 40),
('inv-016', 'spu-010', 'sku-016', 'wh-002', 180, 0, 180, 'in_stock', 30);

-- ============================================================
-- 5. SHOPEE_CART - Carts & Cart Items
-- ============================================================
USE tiki_cart;

INSERT INTO carts (id, user_id, session_id, status, currency, item_count, subtotal, expires_at) VALUES
('cart-001', 'usr-003', NULL, 'active', 'VND', 3, 173380, DATE_ADD(NOW(), INTERVAL 7 DAY)),
('cart-002', 'usr-004', NULL, 'active', 'VND', 2, 58800, DATE_ADD(NOW(), INTERVAL 7 DAY)),
('cart-003', NULL, 'sess-guest-001', 'active', 'VND', 1, 34900, DATE_ADD(NOW(), INTERVAL 1 DAY));

INSERT INTO cart_items (id, cart_id, sku, product_name, shop_id, shop_name, quantity, unit_price, total_price, image_url, is_selected, is_available) VALUES
('ci-001', 'cart-001', 'sku-001', 'iPhone 15 Pro Max 256GB', 'shop-001', 'Apple Store SG', 1, 169900, 169900, '/images/products/product-1.jpg', 1, 1),
('ci-002', 'cart-001', 'sku-007', 'Sony WH-1000XM5 Headphones', 'shop-001', 'Tech Haven', 1, 34900, 34900, '/images/products/product-4.jpg', 1, 1),
('ci-003', 'cart-001', 'sku-013', 'Cotton Crew Neck T-Shirt', 'shop-002', 'Fashion Hub', 2, 1490, 2980, '/images/products/product-9.jpg', 0, 1),
('ci-004', 'cart-002', 'sku-003', 'Samsung Galaxy S24 Ultra', 'shop-001', 'Samsung Official', 1, 149900, 149900, '/images/products/product-2.jpg', 1, 1),
('ci-005', 'cart-002', 'sku-008', 'Nike Air Max 270', 'shop-002', 'Sports World', 2, 15900, 31800, '/images/products/product-5.jpg', 1, 1),
('ci-006', 'cart-003', 'sku-007', 'Sony WH-1000XM5 Headphones', 'shop-001', 'Tech Haven', 1, 34900, 34900, '/images/products/product-4.jpg', 1, 1);

-- ============================================================
-- 6. SHOPEE_ORDER - Orders & Order Items
-- ============================================================
USE tiki_order;

INSERT INTO orders (id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address) VALUES
('ord-001', 'SP-20260521-00001', 'usr-003', 'usr-002', 'delivered', 173390, 'VND',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}'),
('ord-002', 'SP-20260521-00002', 'usr-003', 'usr-005', 'shipped', 89900, 'VND',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}'),
('ord-003', 'SP-20260521-00003', 'usr-004', 'usr-002', 'processing', 149900, 'VND',
 '{"name":"John Doe","phone":"+6590001004","address":"456 Marina Bay, Singapore 018956","postal":"018956"}',
 '{"name":"John Doe","phone":"+6590001004","address":"456 Marina Bay, Singapore 018956","postal":"018956"}'),
('ord-004', 'SP-20260521-00004', 'usr-004', 'usr-005', 'pending', 47700, 'VND',
 '{"name":"John Doe","phone":"+6590001004","address":"456 Marina Bay, Singapore 018956","postal":"018956"}',
 '{"name":"John Doe","phone":"+6590001004","address":"456 Marina Bay, Singapore 018956","postal":"018956"}');

INSERT INTO order_items (id, order_id, product_id, sku_id, shop_id, quantity, unit_price, total_price) VALUES
('oi-001', 'ord-001', 'spu-001', 'sku-001', 'shop-001', 1, 169900, 169900),
('oi-002', 'ord-001', 'spu-004', 'sku-007', 'shop-001', 1, 3490, 3490),
('oi-003', 'ord-002', 'spu-007', 'sku-011', 'shop-002', 1, 79900, 79900),
('oi-004', 'ord-002', 'spu-009', 'sku-013', 'shop-002', 2, 1490, 2980),
('oi-005', 'ord-003', 'spu-002', 'sku-003', 'shop-001', 1, 149900, 149900),
('oi-006', 'ord-004', 'spu-005', 'sku-008', 'shop-002', 3, 15900, 47700);

-- ============================================================
-- 7. SHOPEE_PAYMENT - Payments
-- ============================================================
USE tiki_payment;

INSERT INTO payments (id, order_id, user_id, amount, currency, status, payment_method, psp_provider, psp_transaction_id, authorized_at, captured_at) VALUES
('pay-001', 'ord-001', 'usr-003', 173390, 'VND', 'captured', 'credit_card', 'stripe', 'pi_3Nxample123', NOW(), NOW()),
('pay-002', 'ord-002', 'usr-003', 89900, 'VND', 'captured', 'paynow', 'stripe', 'pi_3Nxample456', NOW(), NOW()),
('pay-003', 'ord-003', 'usr-004', 149900, 'VND', 'authorized', 'credit_card', 'stripe', 'pi_3Nxample789', NOW(), NULL),
('pay-004', 'ord-004', 'usr-004', 47700, 'VND', 'pending', 'paynow', 'stripe', NULL, NULL, NULL);

-- ============================================================
-- 8. SHOPEE_CHECKOUT - Checkouts
-- ============================================================
USE tiki_checkout;

INSERT INTO checkouts (id, user_id, cart_id, order_id, status, idempotency_key, current_step, subtotal, discount_total, shipping_total, grand_total, currency, expires_at, completed_at) VALUES
('chk-001', 'usr-003', 'cart-001', 'ord-001', 'completed', 'idem-001', 'completed', 173390, 0, 0, 173390, 'VND', DATE_ADD(NOW(), INTERVAL 1 DAY), NOW()),
('chk-002', 'usr-003', 'cart-001', 'ord-002', 'completed', 'idem-002', 'completed', 89900, 0, 0, 89900, 'VND', DATE_ADD(NOW(), INTERVAL 1 DAY), NOW()),
('chk-003', 'usr-004', 'cart-002', NULL, 'pending', NULL, 'init', 58800, 0, 500, 59300, 'VND', DATE_ADD(NOW(), INTERVAL 30 MINUTE), NULL);

-- ============================================================
-- 9. SHOPEE_PLATFORM - Vouchers/Promotions
-- ============================================================
USE tiki_platform;

INSERT INTO vouchers (id, code, title, description, type, discount_value, min_spend, max_discount, usage_limit, per_user_limit, scope, start_time, end_time, status, stackable, priority) VALUES
('vch-001', 'WELCOME10', 'Welcome 10% Off', '10% off for new users, min spend $50', 'percentage', 10, 5000, 2000, 10000, 1, 'platform', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_ADD(NOW(), INTERVAL 30 DAY), 'active', 0, 10),
('vch-002', 'FREESHIP', 'Free Shipping', 'Free shipping on all orders', 'shipping', 0, 0, 0, 50000, 1, 'platform', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_ADD(NOW(), INTERVAL 60 DAY), 'active', 1, 5),
('vch-003', 'SAVE20', '$20 Off $100', '$20 off when you spend $100 or more', 'fixed', 2000, 10000, 2000, 5000, 1, 'platform', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_ADD(NOW(), INTERVAL 15 DAY), 'active', 0, 8),
('vch-004', 'ELECTRO15', 'Electronics 15% Off', '15% off on electronics, max $50', 'percentage', 15, 3000, 5000, 3000, 1, 'category', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_ADD(NOW(), INTERVAL 20 DAY), 'active', 0, 7),
('vch-005', 'FASHION25', 'Fashion 25% Off', '25% off on fashion items, max $30', 'percentage', 25, 2000, 3000, 2000, 1, 'category', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_ADD(NOW(), INTERVAL 25 DAY), 'active', 0, 6);

-- ============================================================
-- 10. SHOPEE_PLATFORM - Shipments
-- ============================================================
USE tiki_platform;

INSERT INTO shipments (id, order_id, carrier, tracking_number, status, shipping_address, estimated_delivery, shipping_fee, weight_grams) VALUES
('shp-001', 'ord-001', 'Ninja Van', 'NV-SG-9876543210', 'delivered',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}',
 DATE_SUB(NOW(), INTERVAL 2 DAY), 0, 471),
('shp-002', 'ord-002', 'J&T Express', 'JT-SG-1234567890', 'in_transit',
 '{"name":"Demo Buyer","phone":"+6590001003","address":"123 Orchard Road, Singapore 238863","postal":"238863"}',
 DATE_ADD(NOW(), INTERVAL 2 DAY), 500, 18200),
('shp-003', 'ord-003', 'DHL Express', 'DHL-SG-5555666677', 'pending',
 '{"name":"John Doe","phone":"+6590001004","address":"456 Marina Bay, Singapore 018956","postal":"018956"}',
 DATE_ADD(NOW(), INTERVAL 3 DAY), 800, 233);

-- ============================================================
-- 11. SHOPEE_PLATFORM - Warehouses
-- ============================================================
USE tiki_platform;

INSERT IGNORE INTO warehouses (id, name, code, address, city, country, is_active) VALUES
('wh-001', 'Singapore Central Warehouse', 'SG-CENTRAL', '10 Pasir Panjang Road, Singapore', 'Singapore', 'SG', 1),
('wh-002', 'Singapore East Warehouse', 'SG-EAST', '50 Changi South Avenue, Singapore', 'Singapore', 'SG', 1);

-- ============================================================
-- 12. SHOPEE_PLATFORM - Categories
-- ============================================================
USE tiki_platform;

INSERT INTO categories (id, name, slug, parent_id, level, sort_order, image_url, is_active) VALUES
('cat-001', 'Electronics', 'electronics', NULL, 1, 1, '/images/categories/default-category.png', 1),
('cat-002', 'Mobile Phones', 'mobile-phones', 'cat-001', 2, 1, '/images/categories/default-category.png', 1),
('cat-003', 'Laptops', 'laptops', 'cat-001', 2, 2, '/images/categories/default-category.png', 1),
('cat-004', 'Audio', 'audio', 'cat-001', 2, 3, '/images/categories/default-category.png', 1),
('cat-005', 'Fashion', 'fashion', NULL, 1, 2, '/images/categories/default-category.png', 1),
('cat-006', 'Men\'s Clothing', 'mens-clothing', 'cat-005', 2, 1, '/images/categories/default-category.png', 1),
('cat-007', 'Women\'s Clothing', 'womens-clothing', 'cat-005', 2, 2, '/images/categories/default-category.png', 1),
('cat-008', 'Shoes', 'shoes', 'cat-005', 2, 3, '/images/categories/default-category.png', 1),
('cat-009', 'Home & Living', 'home-living', NULL, 1, 3, '/images/categories/default-category.png', 1),
('cat-010', 'Furniture', 'furniture', 'cat-009', 2, 1, '/images/categories/default-category.png', 1),
('cat-011', 'Kitchen', 'kitchen', 'cat-009', 2, 2, '/images/categories/default-category.png', 1);

-- ============================================================
-- 13. SHOPEE_PLATFORM - Products
-- ============================================================
USE tiki_platform;

INSERT INTO products (id, shop_id, category_id, name, description, brand, status, currency) VALUES
('spu-001', 'shop-001', 'cat-002', 'iPhone 15 Pro Max 256GB', 'Latest Apple iPhone with A17 Pro chip', 'Apple', 'active', 'VND'),
('spu-002', 'shop-001', 'cat-002', 'Samsung Galaxy S24 Ultra', 'Flagship Samsung phone with S Pen', 'Samsung', 'active', 'VND'),
('spu-003', 'shop-002', 'cat-003', 'MacBook Pro 14" M3 Pro', 'Apple MacBook Pro with M3 Pro chip', 'Apple', 'active', 'VND'),
('spu-004', 'shop-001', 'cat-004', 'Sony WH-1000XM5 Headphones', 'Industry-leading noise cancellation', 'Sony', 'active', 'VND'),
('spu-005', 'shop-002', 'cat-008', 'Nike Air Max 270', 'Men\'s running shoes', 'Nike', 'active', 'VND'),
('spu-006', 'shop-001', 'cat-008', 'Adidas Ultraboost 22', 'Women\'s running shoes', 'Adidas', 'active', 'VND'),
('spu-007', 'shop-002', 'cat-010', 'Ergonomic Office Chair', 'Adjustable lumbar support', 'Herman Miller', 'active', 'VND'),
('spu-008', 'shop-001', 'cat-011', 'Non-Stick Cookware Set', '10-piece ceramic non-stick', 'Tefal', 'active', 'VND'),
('spu-009', 'shop-002', 'cat-006', 'Cotton Crew Neck T-Shirt', '100% organic cotton', 'Uniqlo', 'active', 'VND'),
('spu-010', 'shop-001', 'cat-007', 'Floral Summer Dress', 'Lightweight floral print', 'Zara', 'active', 'VND');

-- ============================================================
-- 14. SHOPEE_PLATFORM - SKUs
-- ============================================================
USE tiki_platform;

INSERT INTO skus (id, product_id, sku_code, price, sale_price, stock, status) VALUES
('sku-001', 'spu-001', 'IP15PM-256-BLK', 179900, 169900, 50, 'active'),
('sku-002', 'spu-001', 'IP15PM-512-WHT', 214900, 199900, 30, 'active'),
('sku-003', 'spu-002', 'S24U-256-TIT', 159900, 149900, 75, 'active'),
('sku-004', 'spu-002', 'S24U-512-BLK', 189900, 179900, 40, 'active'),
('sku-005', 'spu-003', 'MBP14-M3P-18G', 249900, 239900, 20, 'active'),
('sku-006', 'spu-004', 'SONY-XM5-BLK', 39900, 34900, 100, 'active'),
('sku-007', 'spu-005', 'NIKE-AM270-10', 18900, 15900, 200, 'active'),
('sku-008', 'spu-006', 'ADIDAS-UB22-7', 21900, 18900, 180, 'active'),
('sku-009', 'spu-007', 'CHAIR-ERG-01', 89900, 79900, 25, 'active'),
('sku-010', 'spu-008', 'COOK-10PC-WHT', 12900, 9900, 300, 'active'),
('sku-011', 'spu-009', 'TSHIRT-M-BLK-L', 1990, 1490, 500, 'active'),
('sku-012', 'spu-010', 'DRESS-F-FLR-M', 4990, 3990, 250, 'active');

-- ============================================================
-- 15. SHOPEE_PLATFORM - Audit Logs
-- ============================================================
USE tiki_platform;

INSERT INTO audit_logs (id, actor_id, actor_type, action, resource_type, resource_id, new_value) VALUES
('al-001', 'usr-001', 'admin', 'USER_REGISTERED', 'user', 'usr-001', '{"email":"admin@tiki.com"}'),
('al-002', 'usr-002', 'user', 'USER_REGISTERED', 'user', 'usr-002', '{"email":"seller@tiki.com"}'),
('al-003', 'usr-003', 'user', 'USER_REGISTERED', 'user', 'usr-003', '{"email":"buyer@tiki.com"}'),
('al-004', 'usr-002', 'user', 'PRODUCT_CREATED', 'product', 'spu-001', '{"title":"iPhone 15 Pro Max"}'),
('al-005', 'usr-003', 'user', 'ORDER_PLACED', 'order', 'ord-001', '{"total":173390}'),
('al-006', 'usr-002', 'user', 'ORDER_SHIPPED', 'order', 'ord-001', '{"tracking":"NV-SG-9876543210"}'),
('al-007', 'system', 'system', 'ORDER_DELIVERED', 'order', 'ord-001', '{"delivered_at":"2026-05-19"}');
