-- ============================================================
-- SHOPEE CLONE — MASTER DATABASE SCHEMA
-- Database: tiki_platform (MySQL 8.0)
-- Charset: utf8mb4
-- Collation: utf8mb4_unicode_ci
-- Engine: InnoDB
-- ============================================================

-- ============================================================
-- 1. IDENTITY & AUTHENTICATION
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) DEFAULT NULL,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) DEFAULT NULL,
    avatar_url VARCHAR(500) DEFAULT NULL,
    status ENUM('pending','active','inactive','locked','suspended') NOT NULL DEFAULT 'pending',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    twofa_secret VARCHAR(255) DEFAULT NULL,
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP NULL DEFAULT NULL,
    last_login_at TIMESTAMP NULL DEFAULT NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_users_email (email),
    INDEX idx_users_phone (phone),
    INDEX idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255) DEFAULT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL,
    action ENUM('create','read','update','delete','manage') NOT NULL,
    description VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_perm_resource_action (resource, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by VARCHAR(36) DEFAULT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    session_id VARCHAR(36) NOT NULL,
    device_id VARCHAR(100) DEFAULT NULL,
    ip_address VARCHAR(45) DEFAULT NULL,
    user_agent TEXT DEFAULT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_rt_user (user_id),
    INDEX idx_rt_token (token_hash),
    INDEX idx_rt_session (session_id),
    INDEX idx_rt_expires (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS failed_login_attempts (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    attempted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_fla_email (email),
    INDEX idx_fla_ip (ip_address),
    INDEX idx_fla_time (attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS idempotency_keys (
    `key` VARCHAR(255) PRIMARY KEY,
    response_body JSON DEFAULT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ik_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS outbox_events (
    event_id VARCHAR(36) PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at TIMESTAMP NULL DEFAULT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT DEFAULT NULL,
    INDEX idx_oe_processed (processed, created_at),
    INDEX idx_oe_aggregate (aggregate_type, aggregate_id),
    INDEX idx_oe_type (event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 2. INVENTORY SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS warehouses (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    address TEXT DEFAULT NULL,
    city VARCHAR(100) DEFAULT NULL,
    region VARCHAR(100) DEFAULT NULL,
    country VARCHAR(3) DEFAULT 'SG',
    priority INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_wh_code (code),
    INDEX idx_wh_active (is_active, deleted_at),
    INDEX idx_wh_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    sku_id VARCHAR(36) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    reserved_qty INT NOT NULL DEFAULT 0,
    available_qty INT NOT NULL DEFAULT 0,
    status ENUM('in_stock','low_stock','out_of_stock','reserved') NOT NULL DEFAULT 'in_stock',
    reorder_level INT NOT NULL DEFAULT 10,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_sku_warehouse (sku_id, warehouse_id),
    INDEX idx_stock_product (product_id),
    INDEX idx_stock_status (status),
    INDEX idx_stock_available (available_qty),
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reservations (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    sku_id VARCHAR(36) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    status ENUM('active','committed','released','expired') NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP NOT NULL,
    idempotency_key VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_res_order (order_id),
    INDEX idx_res_sku (sku_id),
    INDEX idx_res_status (status),
    INDEX idx_res_expires (expires_at),
    INDEX idx_res_idempotency (idempotency_key),
    INDEX idx_res_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_movements (
    id VARCHAR(36) PRIMARY KEY,
    sku_id VARCHAR(36) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    movement_type ENUM('increase','decrease','reserve','release','confirm','adjust','flash_sale') NOT NULL,
    quantity INT NOT NULL,
    before_qty INT NOT NULL,
    after_qty INT NOT NULL,
    reference_id VARCHAR(36) DEFAULT NULL,
    reason TEXT DEFAULT NULL,
    operator_id VARCHAR(36) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sm_sku (sku_id),
    INDEX idx_sm_warehouse (warehouse_id),
    INDEX idx_sm_type (movement_type),
    INDEX idx_sm_reference (reference_id),
    INDEX idx_sm_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS flash_sale_inventory (
    id VARCHAR(36) PRIMARY KEY,
    flash_sale_id VARCHAR(36) NOT NULL,
    sku_id VARCHAR(36) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    total_stock BIGINT NOT NULL DEFAULT 0,
    reserved_stock BIGINT NOT NULL DEFAULT 0,
    sold_stock BIGINT NOT NULL DEFAULT 0,
    max_per_user BIGINT NOT NULL DEFAULT 1,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_fsi_sku (flash_sale_id, sku_id),
    INDEX idx_fsi_sale (flash_sale_id),
    INDEX idx_fsi_active (is_active, start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 3. CART SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS carts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) DEFAULT NULL,
    session_id VARCHAR(100) DEFAULT NULL,
    status ENUM('active','merged','expired','checkout') NOT NULL DEFAULT 'active',
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    item_count INT NOT NULL DEFAULT 0,
    subtotal BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_cart_user (user_id, status, deleted_at),
    INDEX idx_cart_session (session_id, status, deleted_at),
    INDEX idx_cart_expires (expires_at, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cart_items (
    id VARCHAR(36) PRIMARY KEY,
    cart_id VARCHAR(36) NOT NULL,
    sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(500) NOT NULL,
    shop_id VARCHAR(36) NOT NULL,
    shop_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    image_url VARCHAR(500) DEFAULT NULL,
    attributes TEXT DEFAULT NULL,
    is_selected BOOLEAN NOT NULL DEFAULT TRUE,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ci_cart (cart_id),
    INDEX idx_ci_sku (sku),
    INDEX idx_ci_shop (shop_id),
    UNIQUE KEY uk_cart_sku (cart_id, sku),
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cart_snapshots (
    id VARCHAR(36) PRIMARY KEY,
    cart_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    items JSON NOT NULL,
    seller_groups JSON NOT NULL,
    subtotal BIGINT NOT NULL,
    item_count INT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_cs_cart (cart_id),
    INDEX idx_cs_idempotency (idempotency_key),
    INDEX idx_cs_expires (expires_at),
    FOREIGN KEY (cart_id) REFERENCES carts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cart_merge_history (
    id VARCHAR(36) PRIMARY KEY,
    source_cart_id VARCHAR(36) NOT NULL,
    target_cart_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    merge_type ENUM('guest_to_user','session','conflict_resolution') NOT NULL,
    items_merged INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_cmh_user (user_id),
    INDEX idx_cmh_target (target_cart_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 4. PRODUCT CATALOG SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) DEFAULT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    level INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    image_url VARCHAR(500) DEFAULT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cat_parent (parent_id, is_active),
    INDEX idx_cat_level (level, sort_order),
    INDEX idx_cat_slug (slug),
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    shop_id VARCHAR(36) NOT NULL,
    category_id VARCHAR(36) NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT DEFAULT NULL,
    brand VARCHAR(255) DEFAULT NULL,
    status ENUM('draft','pending_review','active','inactive','archived','rejected') NOT NULL DEFAULT 'draft',
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_prod_shop (shop_id, status, deleted_at),
    INDEX idx_prod_category (category_id, status, deleted_at),
    INDEX idx_prod_status (status),
    FULLTEXT INDEX idx_prod_search (name, description),
    FOREIGN KEY (category_id) REFERENCES categories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS skus (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    sku_code VARCHAR(100) NOT NULL,
    attributes JSON DEFAULT NULL,
    price BIGINT NOT NULL,
    sale_price BIGINT DEFAULT NULL,
    stock INT NOT NULL DEFAULT 0,
    status ENUM('active','inactive','out_of_stock') NOT NULL DEFAULT 'active',
    weight_grams INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_sku_code (sku_code),
    INDEX idx_sku_product (product_id),
    INDEX idx_sku_status (status),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS product_media (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    media_type ENUM('image','video') NOT NULL,
    url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500) DEFAULT NULL,
    alt_text VARCHAR(255) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pm_product (product_id),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS product_attributes (
    id VARCHAR(36) PRIMARY KEY,
    category_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    attr_type ENUM('text','number','select','multi_select','boolean','color') NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    options JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pa_category (category_id, is_active),
    FOREIGN KEY (category_id) REFERENCES categories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS product_moderation (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    status ENUM('pending','approved','rejected','flagged') NOT NULL DEFAULT 'pending',
    reason TEXT DEFAULT NULL,
    reviewer_id VARCHAR(36) DEFAULT NULL,
    reviewed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pmod_product (product_id),
    INDEX idx_pmod_status (status),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 5. PROMOTION SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS vouchers (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT NULL,
    type ENUM('percentage','fixed','shipping') NOT NULL,
    discount_value BIGINT NOT NULL,
    min_spend BIGINT NOT NULL DEFAULT 0,
    max_discount BIGINT NOT NULL DEFAULT 0,
    usage_limit BIGINT NOT NULL DEFAULT 10000,
    usage_count BIGINT NOT NULL DEFAULT 0,
    per_user_limit INT NOT NULL DEFAULT 1,
    scope ENUM('platform','shop','category','sku') NOT NULL DEFAULT 'platform',
    shop_id VARCHAR(36) DEFAULT NULL,
    category_id VARCHAR(36) DEFAULT NULL,
    sku VARCHAR(100) DEFAULT NULL,
    region VARCHAR(50) DEFAULT NULL,
    payment_method VARCHAR(50) DEFAULT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status ENUM('active','inactive','expired','exhausted') NOT NULL DEFAULT 'active',
    stackable BOOLEAN NOT NULL DEFAULT FALSE,
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_vch_code (code),
    INDEX idx_vch_status (status, start_time, end_time),
    INDEX idx_vch_scope (scope, shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS voucher_redemptions (
    id VARCHAR(36) PRIMARY KEY,
    voucher_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36) NOT NULL,
    discount_amount BIGINT NOT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vr_voucher (voucher_id),
    INDEX idx_vr_user (user_id, voucher_id),
    INDEX idx_vr_order (order_id),
    INDEX idx_vr_idempotency (idempotency_key),
    FOREIGN KEY (voucher_id) REFERENCES vouchers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS campaigns (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type ENUM('flash_sale','scheduled','seasonal','seller','category') NOT NULL,
    description TEXT DEFAULT NULL,
    rules JSON DEFAULT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status ENUM('draft','active','paused','ended') NOT NULL DEFAULT 'draft',
    priority INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cmp_type (type, status),
    INDEX idx_cmp_time (start_time, end_time, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS eligibility_rules (
    id VARCHAR(36) PRIMARY KEY,
    promotion_id VARCHAR(36) NOT NULL,
    target_type ENUM('user','region','payment','seller','product') NOT NULL,
    target_value VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    INDEX idx_er_promo (promotion_id),
    FOREIGN KEY (promotion_id) REFERENCES vouchers(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stacking_rules (
    id VARCHAR(36) PRIMARY KEY,
    promotion_type VARCHAR(50) NOT NULL,
    can_stack_with VARCHAR(50) NOT NULL,
    max_stack_count INT NOT NULL DEFAULT 1,
    priority INT NOT NULL DEFAULT 0,
    INDEX idx_sr_type (promotion_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 6. ORDER SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id VARCHAR(36) NOT NULL,
    shop_id VARCHAR(36) NOT NULL,
    status ENUM('pending','confirmed','processing','shipped','delivered','cancelled','refunded','failed') NOT NULL DEFAULT 'pending',
    subtotal BIGINT NOT NULL,
    discount_total BIGINT NOT NULL DEFAULT 0,
    shipping_fee BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    shipping_address JSON NOT NULL,
    billing_address JSON DEFAULT NULL,
    notes TEXT DEFAULT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    placed_at TIMESTAMP NULL DEFAULT NULL,
    confirmed_at TIMESTAMP NULL DEFAULT NULL,
    shipped_at TIMESTAMP NULL DEFAULT NULL,
    delivered_at TIMESTAMP NULL DEFAULT NULL,
    cancelled_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ord_user (user_id, status, created_at),
    INDEX idx_ord_shop (shop_id, status, created_at),
    INDEX idx_ord_status (status),
    INDEX idx_ord_number (order_number),
    INDEX idx_ord_idempotency (idempotency_key),
    INDEX idx_ord_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    sku_id VARCHAR(36) NOT NULL,
    product_name VARCHAR(500) NOT NULL,
    quantity INT NOT NULL,
    unit_price BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_oi_order (order_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_status_history (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    from_status VARCHAR(50) DEFAULT NULL,
    to_status VARCHAR(50) NOT NULL,
    actor_id VARCHAR(36) DEFAULT NULL,
    actor_type ENUM('user','system','admin','webhook') NOT NULL DEFAULT 'system',
    reason TEXT DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_osh_order (order_id, created_at),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_cancellations (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    reason TEXT NOT NULL,
    cancelled_by VARCHAR(36) NOT NULL,
    cancellation_type ENUM('user','system','admin','fraud','payment_failed','out_of_stock') NOT NULL,
    refund_amount BIGINT NOT NULL DEFAULT 0,
    compensation_status ENUM('pending','in_progress','completed','failed') DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_oc_order (order_id),
    FOREIGN KEY (order_id) REFERENCES orders(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 7. PAYMENT SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    payment_method ENUM('credit_card','debit_card','wallet','bank_transfer','cod','paylater') NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_transaction_id VARCHAR(255) DEFAULT NULL,
    status ENUM('pending','authorized','captured','failed','refunded','partially_refunded','cancelled') NOT NULL DEFAULT 'pending',
    amount_refunded BIGINT NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    authorized_at TIMESTAMP NULL DEFAULT NULL,
    captured_at TIMESTAMP NULL DEFAULT NULL,
    failed_at TIMESTAMP NULL DEFAULT NULL,
    failure_reason TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_pay_order_status (order_id, status),
    INDEX idx_pay_user (user_id, created_at),
    INDEX idx_pay_status (status),
    INDEX idx_pay_provider_tx (provider_transaction_id),
    INDEX idx_pay_idempotency (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refunds (
    id VARCHAR(36) PRIMARY KEY,
    payment_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reason TEXT NOT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    status ENUM('pending','processing','completed','failed') NOT NULL DEFAULT 'pending',
    provider_refund_id VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ref_payment (payment_id),
    INDEX idx_ref_order (order_id),
    INDEX idx_ref_status (status),
    FOREIGN KEY (payment_id) REFERENCES payments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS fraud_checks (
    id VARCHAR(36) PRIMARY KEY,
    payment_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    risk_score INT NOT NULL,
    is_fraudulent BOOLEAN NOT NULL DEFAULT FALSE,
    signals JSON DEFAULT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_fc_payment (payment_id),
    INDEX idx_fc_user (user_id),
    INDEX idx_fc_score (risk_score),
    FOREIGN KEY (payment_id) REFERENCES payments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS webhook_events (
    id VARCHAR(36) PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSON NOT NULL,
    signature VARCHAR(500) NOT NULL,
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_we_provider (provider, event_type),
    INDEX idx_we_processed (processed)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 8. CHECKOUT SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS checkouts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    cart_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36) DEFAULT NULL,
    status ENUM('pending','validating','pricing_frozen','reserving_inventory','inventory_reserved','processing_payment','completed','failed','rolling_back','rolled_back','expired') NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(100) DEFAULT NULL,
    current_step VARCHAR(50) NOT NULL DEFAULT 'init',
    failure_reason TEXT DEFAULT NULL,
    attempt_count INT NOT NULL DEFAULT 0,
    reservation_keys TEXT DEFAULT NULL,
    pricing_snapshot_id VARCHAR(36) DEFAULT NULL,
    promotion_results TEXT DEFAULT NULL,
    subtotal BIGINT NOT NULL DEFAULT 0,
    discount_total BIGINT NOT NULL DEFAULT 0,
    shipping_total BIGINT NOT NULL DEFAULT 0,
    grand_total BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    expires_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ck_user (user_id, status),
    INDEX idx_ck_cart (cart_id),
    INDEX idx_ck_status (status, expires_at),
    INDEX idx_ck_idempotency (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS checkout_step_logs (
    id VARCHAR(36) PRIMARY KEY,
    checkout_id VARCHAR(36) NOT NULL,
    step VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT DEFAULT NULL,
    metadata TEXT DEFAULT NULL,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_csl_checkout (checkout_id),
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pricing_snapshots (
    id VARCHAR(36) PRIMARY KEY,
    checkout_id VARCHAR(36) NOT NULL,
    items JSON NOT NULL,
    seller_groups JSON NOT NULL,
    subtotal BIGINT NOT NULL,
    discount_total BIGINT NOT NULL DEFAULT 0,
    shipping_total BIGINT NOT NULL DEFAULT 0,
    grand_total BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    promotions_applied JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ps_checkout (checkout_id),
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reservation_orchestrations (
    id VARCHAR(36) PRIMARY KEY,
    checkout_id VARCHAR(36) NOT NULL,
    reservation_key VARCHAR(100) NOT NULL,
    sku VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    quantity BIGINT NOT NULL,
    status ENUM('pending','reserved','released','failed') NOT NULL DEFAULT 'pending',
    error_message TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ro_checkout (checkout_id),
    INDEX idx_ro_key (reservation_key),
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reconciliation_jobs (
    id VARCHAR(36) PRIMARY KEY,
    checkout_id VARCHAR(36) NOT NULL,
    job_type ENUM('release_reservation','confirm_reservation','update_order_status') NOT NULL,
    status ENUM('pending','running','completed','failed') NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP NOT NULL,
    metadata TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rj_status (status, next_retry_at),
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 9. SHIPMENT SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS shipments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    carrier VARCHAR(100) NOT NULL,
    tracking_number VARCHAR(100) DEFAULT NULL,
    status ENUM('pending','picked_up','in_transit','out_for_delivery','delivered','failed','returned') NOT NULL DEFAULT 'pending',
    shipping_address JSON NOT NULL,
    estimated_delivery TIMESTAMP NULL DEFAULT NULL,
    actual_delivery TIMESTAMP NULL DEFAULT NULL,
    shipping_fee BIGINT NOT NULL DEFAULT 0,
    weight_grams INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ship_order (order_id),
    INDEX idx_ship_tracking (tracking_number),
    INDEX idx_ship_status (status),
    INDEX idx_ship_carrier (carrier)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_tracking_events (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    location VARCHAR(255) DEFAULT NULL,
    description TEXT DEFAULT NULL,
    event_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ste_shipment (shipment_id, event_time),
    FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 10. NOTIFICATION SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    body TEXT DEFAULT NULL,
    data JSON DEFAULT NULL,
    channel ENUM('push','email','sms','inapp') NOT NULL,
    status ENUM('pending','sent','delivered','failed','read') NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMP NULL DEFAULT NULL,
    delivered_at TIMESTAMP NULL DEFAULT NULL,
    read_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notif_user (user_id, status, created_at),
    INDEX idx_notif_type (type),
    INDEX idx_notif_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS notification_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(500) DEFAULT NULL,
    body TEXT NOT NULL,
    variables JSON DEFAULT NULL,
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_nt_type (type, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_notification_preferences (
    user_id VARCHAR(36) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    quiet_hours VARCHAR(50) DEFAULT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, channel, notification_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS push_devices (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    device_token VARCHAR(500) NOT NULL,
    platform ENUM('ios','android','web') NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pd_user (user_id, is_active),
    INDEX idx_pd_token (device_token),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 11. SEARCH INDEXING SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS search_index_tasks (
    id VARCHAR(36) PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36) NOT NULL,
    action ENUM('index','update','delete') NOT NULL,
    status ENUM('pending','processing','completed','failed') NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 0,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    error_message TEXT DEFAULT NULL,
    scheduled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sit_status (status, scheduled_at),
    INDEX idx_sit_entity (entity_type, entity_id),
    INDEX idx_sit_priority (priority, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS search_query_logs (
    id VARCHAR(36) PRIMARY KEY,
    query_text VARCHAR(500) NOT NULL,
    user_id VARCHAR(36) DEFAULT NULL,
    session_id VARCHAR(100) DEFAULT NULL,
    result_count INT NOT NULL DEFAULT 0,
    response_time_ms INT NOT NULL DEFAULT 0,
    filters JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sql_query (query_text),
    INDEX idx_sql_time (created_at),
    INDEX idx_sql_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 12. RECOMMENDATION SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS user_embeddings (
    user_id VARCHAR(36) PRIMARY KEY,
    embedding JSON NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ue_model (model_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS product_embeddings (
    product_id VARCHAR(36) PRIMARY KEY,
    embedding JSON NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pe_model (model_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS recommendation_events (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    event_type ENUM('impression','click','add_to_cart','purchase') NOT NULL,
    recommendation_type VARCHAR(50) NOT NULL,
    position INT DEFAULT NULL,
    score FLOAT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_re_user (user_id, event_type, created_at),
    INDEX idx_re_product (product_id, event_type),
    INDEX idx_re_type (recommendation_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 13. USER BEHAVIOR SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS click_events (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    product_id VARCHAR(36) DEFAULT NULL,
    page_url VARCHAR(500) NOT NULL,
    referrer VARCHAR(500) DEFAULT NULL,
    device_type VARCHAR(20) DEFAULT NULL,
    ip_address VARCHAR(45) DEFAULT NULL,
    user_agent TEXT DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ce_user (user_id, event_time),
    INDEX idx_ce_session (session_id),
    INDEX idx_ce_type (event_type, event_time),
    INDEX idx_ce_product (product_id, event_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(100) NOT NULL UNIQUE,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP NULL DEFAULT NULL,
    page_views INT NOT NULL DEFAULT 0,
    events INT NOT NULL DEFAULT 0,
    device_type VARCHAR(20) DEFAULT NULL,
    ip_address VARCHAR(45) DEFAULT NULL,
    INDEX idx_us_user (user_id, start_time),
    INDEX idx_us_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 14. FRAUD DETECTION SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS fraud_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    rule_type ENUM('velocity','amount','geolocation','device','behavior','custom') NOT NULL,
    conditions JSON NOT NULL,
    action ENUM('block','review','flag','allow') NOT NULL DEFAULT 'review',
    priority INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_fr_type (rule_type, is_active),
    INDEX idx_fr_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS fraud_cases (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36) DEFAULT NULL,
    payment_id VARCHAR(36) DEFAULT NULL,
    status ENUM('open','reviewing','confirmed_fraud','dismissed','escalated') NOT NULL DEFAULT 'open',
    risk_score FLOAT NOT NULL DEFAULT 0,
    evidence JSON DEFAULT NULL,
    assigned_to VARCHAR(36) DEFAULT NULL,
    resolution_notes TEXT DEFAULT NULL,
    resolved_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_fc_user (user_id, status),
    INDEX idx_fc_status (status),
    INDEX idx_fc_score (risk_score),
    INDEX idx_fc_assigned (assigned_to)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS device_fingerprints (
    id VARCHAR(36) PRIMARY KEY,
    fingerprint_hash VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(36) DEFAULT NULL,
    device_info JSON DEFAULT NULL,
    risk_score FLOAT NOT NULL DEFAULT 0,
    first_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_df_user (user_id),
    INDEX idx_df_hash (fingerprint_hash),
    INDEX idx_df_risk (risk_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 15. ADVERTISING SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS ad_campaigns (
    id VARCHAR(36) PRIMARY KEY,
    advertiser_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status ENUM('draft','active','paused','ended') NOT NULL DEFAULT 'draft',
    budget BIGINT NOT NULL,
    daily_budget BIGINT NOT NULL,
    spend BIGINT NOT NULL DEFAULT 0,
    bid_strategy ENUM('cpc','cpm','cpa') NOT NULL DEFAULT 'cpc',
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    targeting JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ac_advertiser (advertiser_id, status),
    INDEX idx_ac_status (status, start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_groups (
    id VARCHAR(36) PRIMARY KEY,
    campaign_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status ENUM('active','paused') NOT NULL DEFAULT 'active',
    bid_amount BIGINT NOT NULL,
    targeting JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ag_campaign (campaign_id, status),
    FOREIGN KEY (campaign_id) REFERENCES ad_campaigns(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_impressions (
    id VARCHAR(36) PRIMARY KEY,
    ad_group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) DEFAULT NULL,
    query VARCHAR(500) DEFAULT NULL,
    position INT DEFAULT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ai_adgroup (ad_group_id, event_time),
    INDEX idx_ai_user (user_id, event_time),
    FOREIGN KEY (ad_group_id) REFERENCES ad_groups(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ad_clicks (
    id VARCHAR(36) PRIMARY KEY,
    ad_group_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    cost BIGINT NOT NULL DEFAULT 0,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ac_adgroup (ad_group_id, event_time),
    INDEX idx_ac_user (user_id, event_time),
    FOREIGN KEY (ad_group_id) REFERENCES ad_groups(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 16. LIVE COMMERCE SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS livestreams (
    id VARCHAR(36) PRIMARY KEY,
    seller_id VARCHAR(36) NOT NULL,
    title VARCHAR(500) NOT NULL,
    status ENUM('scheduled','live','ended') NOT NULL DEFAULT 'scheduled',
    viewer_count BIGINT NOT NULL DEFAULT 0,
    peak_viewers BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMP NULL DEFAULT NULL,
    ended_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ls_seller (seller_id, status),
    INDEX idx_ls_status (status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS livestream_chat_messages (
    id VARCHAR(36) PRIMARY KEY,
    livestream_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    message_type ENUM('text','gift','reaction','system') NOT NULL DEFAULT 'text',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_lcm_stream (livestream_id, created_at),
    FOREIGN KEY (livestream_id) REFERENCES livestreams(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS livestream_pinned_products (
    id VARCHAR(36) PRIMARY KEY,
    livestream_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    pinned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_lpp_stream (livestream_id, pinned_at),
    FOREIGN KEY (livestream_id) REFERENCES livestreams(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS livestream_gifts (
    id VARCHAR(36) PRIMARY KEY,
    livestream_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    gift_type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_lg_stream (livestream_id, created_at),
    FOREIGN KEY (livestream_id) REFERENCES livestreams(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 17. ANALYTICS SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS analytics_events (
    id VARCHAR(36) PRIMARY KEY,
    event_name VARCHAR(100) NOT NULL,
    user_id VARCHAR(36) DEFAULT NULL,
    session_id VARCHAR(100) DEFAULT NULL,
    properties JSON DEFAULT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ae_name (event_name, event_time),
    INDEX idx_ae_user (user_id, event_time),
    INDEX idx_ae_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS aggregated_metrics (
    id VARCHAR(36) PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    dimension VARCHAR(100) DEFAULT NULL,
    dimension_value VARCHAR(255) DEFAULT NULL,
    value BIGINT NOT NULL DEFAULT 0,
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_am_metric_dim_window (metric_name, dimension, dimension_value, window_start),
    INDEX idx_am_name_window (metric_name, window_start, window_end)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 18. BILLING SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS billing_accounts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    account_type ENUM('user','seller','platform') NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    status ENUM('active','suspended','closed') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_ba_user_type (user_id, account_type),
    INDEX idx_ba_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS billing_transactions (
    id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(36) NOT NULL,
    transaction_type ENUM('credit','debit','refund','adjustment','fee') NOT NULL,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_type VARCHAR(50) DEFAULT NULL,
    reference_id VARCHAR(36) DEFAULT NULL,
    description TEXT DEFAULT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_bt_account (account_id, created_at),
    INDEX idx_bt_reference (reference_type, reference_id),
    INDEX idx_bt_idempotency (idempotency_key),
    FOREIGN KEY (account_id) REFERENCES billing_accounts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 19. LOGISTICS DELIVERY SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS delivery_routes (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL,
    origin_lat DECIMAL(10, 8) NOT NULL,
    origin_lng DECIMAL(11, 8) NOT NULL,
    dest_lat DECIMAL(10, 8) NOT NULL,
    dest_lng DECIMAL(11, 8) NOT NULL,
    distance_km FLOAT NOT NULL DEFAULT 0,
    estimated_duration_min INT NOT NULL DEFAULT 0,
    status ENUM('planned','active','completed','cancelled') NOT NULL DEFAULT 'planned',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dr_shipment (shipment_id),
    INDEX idx_dr_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS couriers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) DEFAULT NULL,
    vehicle_type VARCHAR(50) DEFAULT NULL,
    status ENUM('available','busy','offline') NOT NULL DEFAULT 'available',
    current_lat DECIMAL(10, 8) DEFAULT NULL,
    current_lng DECIMAL(11, 8) DEFAULT NULL,
    last_location_update TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_courier_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS delivery_dispatches (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL,
    courier_id VARCHAR(36) NOT NULL,
    route_id VARCHAR(36) DEFAULT NULL,
    status ENUM('assigned','picked_up','in_transit','delivered','failed') NOT NULL DEFAULT 'assigned',
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    picked_up_at TIMESTAMP NULL DEFAULT NULL,
    delivered_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_dd_shipment (shipment_id),
    INDEX idx_dd_courier (courier_id, status),
    FOREIGN KEY (courier_id) REFERENCES couriers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 20. OMS FULFILLMENT SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS fulfillment_orders (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    warehouse_id VARCHAR(36) NOT NULL,
    status ENUM('pending','picking','packed','shipped','delivered','returned','cancelled') NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 0,
    assigned_to VARCHAR(36) DEFAULT NULL,
    picked_at TIMESTAMP NULL DEFAULT NULL,
    packed_at TIMESTAMP NULL DEFAULT NULL,
    shipped_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_fo_order (order_id),
    INDEX idx_fo_warehouse (warehouse_id, status),
    INDEX idx_fo_status (status, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS fulfillment_items (
    id VARCHAR(36) PRIMARY KEY,
    fulfillment_order_id VARCHAR(36) NOT NULL,
    sku_id VARCHAR(36) NOT NULL,
    quantity INT NOT NULL,
    picked_quantity INT NOT NULL DEFAULT 0,
    status ENUM('pending','picked','packed','shortage') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_fi_fulfillment (fulfillment_order_id),
    FOREIGN KEY (fulfillment_order_id) REFERENCES fulfillment_orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 21. PAYMENT LEDGER SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS ledger_entries (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    entry_type ENUM('debit','credit') NOT NULL,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    reference_type VARCHAR(50) DEFAULT NULL,
    reference_id VARCHAR(36) DEFAULT NULL,
    description TEXT DEFAULT NULL,
    idempotency_key VARCHAR(100) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_le_account (account_id, created_at),
    INDEX idx_le_transaction (transaction_id),
    INDEX idx_le_reference (reference_type, reference_id),
    INDEX idx_le_idempotency (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ledger_transactions (
    id VARCHAR(36) PRIMARY KEY,
    transaction_type ENUM('payment','refund','settlement','adjustment','fee') NOT NULL,
    status ENUM('pending','completed','failed','reversed') NOT NULL DEFAULT 'pending',
    total_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'VND',
    idempotency_key VARCHAR(100) DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_lt_type (transaction_type, status),
    INDEX idx_lt_idempotency (idempotency_key),
    INDEX idx_lt_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 22. NOTIFICATION CAMPAIGN SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS notification_campaigns (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type ENUM('promotional','transactional','system') NOT NULL,
    channel ENUM('push','email','sms','inapp') NOT NULL,
    template_id VARCHAR(36) DEFAULT NULL,
    target_audience JSON DEFAULT NULL,
    status ENUM('draft','scheduled','running','paused','completed') NOT NULL DEFAULT 'draft',
    scheduled_at TIMESTAMP NULL DEFAULT NULL,
    started_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    total_recipients INT NOT NULL DEFAULT 0,
    sent_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_nc_status (status, scheduled_at),
    INDEX idx_nc_type (type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 23. SERVICE MESH / API GATEWAY
-- ============================================================

CREATE TABLE IF NOT EXISTS api_routes (
    id VARCHAR(36) PRIMARY KEY,
    path VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    strip_path BOOLEAN NOT NULL DEFAULT FALSE,
    preserve_host BOOLEAN NOT NULL DEFAULT FALSE,
    timeout_ms INT NOT NULL DEFAULT 30000,
    retry_count INT NOT NULL DEFAULT 3,
    circuit_breaker_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_per_minute INT NOT NULL DEFAULT 1000,
    auth_required BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_ar_path_method (path, method),
    INDEX idx_ar_service (service_name, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS service_health_checks (
    id VARCHAR(36) PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    check_interval_sec INT NOT NULL DEFAULT 30,
    timeout_sec INT NOT NULL DEFAULT 5,
    healthy_threshold INT NOT NULL DEFAULT 2,
    unhealthy_threshold INT NOT NULL DEFAULT 3,
    last_check_at TIMESTAMP NULL DEFAULT NULL,
    last_status ENUM('healthy','unhealthy','unknown') NOT NULL DEFAULT 'unknown',
    consecutive_failures INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_shc_service (service_name),
    INDEX idx_shc_status (last_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 24. GLOBAL INFRASTRUCTURE
-- ============================================================

CREATE TABLE IF NOT EXISTS system_config (
    id VARCHAR(36) PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    value_type ENUM('string','int','float','json','bool') NOT NULL DEFAULT 'string',
    description TEXT DEFAULT NULL,
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sc_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    actor_id VARCHAR(36) DEFAULT NULL,
    actor_type ENUM('user','system','admin','service') NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(36) DEFAULT NULL,
    old_value JSON DEFAULT NULL,
    new_value JSON DEFAULT NULL,
    ip_address VARCHAR(45) DEFAULT NULL,
    user_agent TEXT DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_al_actor (actor_id, actor_type, created_at),
    INDEX idx_al_resource (resource_type, resource_id, created_at),
    INDEX idx_al_action (action, created_at),
    INDEX idx_al_time (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 25. SRE / OPERATIONS
-- ============================================================

CREATE TABLE IF NOT EXISTS deployment_history (
    id VARCHAR(36) PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    environment ENUM('development','staging','production') NOT NULL,
    deployed_by VARCHAR(36) DEFAULT NULL,
    deployment_strategy ENUM('rolling','blue_green','canary','recreate') NOT NULL DEFAULT 'rolling',
    status ENUM('pending','in_progress','completed','failed','rolled_back') NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    rollback_reason TEXT DEFAULT NULL,
    metadata JSON DEFAULT NULL,
    INDEX idx_dh_service (service_name, environment, started_at),
    INDEX idx_dh_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS incident_reports (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    severity ENUM('critical','high','medium','low') NOT NULL,
    status ENUM('open','investigating','mitigating','resolved','postmortem') NOT NULL DEFAULT 'open',
    affected_services JSON NOT NULL,
    root_cause TEXT DEFAULT NULL,
    resolution TEXT DEFAULT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL DEFAULT NULL,
    created_by VARCHAR(36) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ir_severity (severity, status),
    INDEX idx_ir_status (status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 26. AIML SERVICE
-- ============================================================

CREATE TABLE IF NOT EXISTS ml_models (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    model_type ENUM('recommendation','fraud_detection','search_ranking','pricing','nlp','image') NOT NULL,
    version VARCHAR(50) NOT NULL,
    status ENUM('training','active','deprecated','failed') NOT NULL DEFAULT 'training',
    artifact_path VARCHAR(500) DEFAULT NULL,
    metrics JSON DEFAULT NULL,
    feature_schema JSON DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_ml_name_version (name, version),
    INDEX idx_ml_type (model_type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ml_training_jobs (
    id VARCHAR(36) PRIMARY KEY,
    model_id VARCHAR(36) NOT NULL,
    status ENUM('pending','running','completed','failed','cancelled') NOT NULL DEFAULT 'pending',
    training_data_path VARCHAR(500) DEFAULT NULL,
    hyperparameters JSON DEFAULT NULL,
    metrics JSON DEFAULT NULL,
    started_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    error_message TEXT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_mtj_model (model_id, status),
    INDEX idx_mtj_status (status),
    FOREIGN KEY (model_id) REFERENCES ml_models(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 27. DEVELOPER PORTAL
-- ============================================================

CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(36) PRIMARY KEY,
    developer_id VARCHAR(36) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    scopes JSON NOT NULL,
    rate_limit_per_minute INT NOT NULL DEFAULT 100,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    last_used_at TIMESTAMP NULL DEFAULT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ak_developer (developer_id, is_active),
    INDEX idx_ak_hash (key_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS developer_applications (
    id VARCHAR(36) PRIMARY KEY,
    developer_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT NULL,
    redirect_uris JSON DEFAULT NULL,
    client_id VARCHAR(100) NOT NULL UNIQUE,
    status ENUM('active','suspended','revoked') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_da_developer (developer_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- INITIAL DATA
-- ============================================================

-- Default roles
INSERT IGNORE INTO roles (id, name, description) VALUES
    ('role_admin', 'admin', 'Platform administrator with full access'),
    ('role_seller', 'seller', 'Seller with shop management access'),
    ('role_buyer', 'buyer', 'Regular buyer account'),
    ('role_support', 'support', 'Customer support agent'),
    ('role_moderator', 'moderator', 'Content moderator');

-- Default permissions
INSERT IGNORE INTO permissions (id, name, resource, action) VALUES
    ('perm_product_create', 'product:create', 'product', 'create'),
    ('perm_product_read', 'product:read', 'product', 'read'),
    ('perm_product_update', 'product:update', 'product', 'update'),
    ('perm_product_delete', 'product:delete', 'product', 'delete'),
    ('perm_order_manage', 'order:manage', 'order', 'manage'),
    ('perm_user_manage', 'user:manage', 'user', 'manage'),
    ('perm_admin_full', 'admin:full', '*', 'manage');

-- Default admin role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES
    ('role_admin', 'perm_admin_full'),
    ('role_seller', 'perm_product_create'),
    ('role_seller', 'perm_product_read'),
    ('role_seller', 'perm_product_update'),
    ('role_seller', 'perm_order_manage'),
    ('role_buyer', 'perm_product_read'),
    ('role_moderator', 'perm_product_read'),
    ('role_moderator', 'perm_product_update');

-- Default system config
INSERT IGNORE INTO system_config (id, config_key, config_value, value_type, description) VALUES
    ('cfg_001', 'platform.name', 'Tiki Clone', 'string', 'Platform display name'),
    ('cfg_002', 'platform.currency', 'VND', 'string', 'Default currency'),
    ('cfg_003', 'inventory.reservation_ttl_minutes', '15', 'int', 'Reservation expiration time'),
    ('cfg_004', 'payment.idempotency_ttl_hours', '24', 'int', 'Payment idempotency key TTL'),
    ('cfg_005', 'order.max_items_per_order', '50', 'int', 'Maximum items per order'),
    ('cfg_006', 'search.max_results', '100', 'int', 'Maximum search results'),
    ('cfg_007', 'notification.batch_size', '1000', 'int', 'Notification batch send size'),
    ('cfg_008', 'fraud.max_risk_score', '80', 'int', 'Maximum risk score before blocking');
