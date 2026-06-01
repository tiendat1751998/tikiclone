-- ============================================================
-- TIKI PRODUCTS CRAWLER DATABASE SCHEMA
-- Database: tiki_platform (MySQL 8.0)
-- ============================================================

USE tiki_platform;

-- ============================================================
-- Tiki Categories (crawled from tiki.vn)
-- ============================================================
CREATE TABLE IF NOT EXISTS tiki_categories (
    id VARCHAR(36) PRIMARY KEY,
    tiki_category_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Original Tiki category ID (e.g., 1789)',
    parent_id VARCHAR(36) DEFAULT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    url_path VARCHAR(500) NOT NULL COMMENT 'Full URL path like /dien-thoai-may-tinh-bang/c1789',
    image_url VARCHAR(500) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    crawled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tiki_cat_tiki_id (tiki_category_id),
    INDEX idx_tiki_cat_parent (parent_id),
    INDEX idx_tiki_cat_slug (slug),
    FOREIGN KEY (parent_id) REFERENCES tiki_categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Tiki Products (crawled from tiki.vn product listing pages)
-- ============================================================
CREATE TABLE IF NOT EXISTS tiki_products (
    id VARCHAR(36) PRIMARY KEY,
    tiki_product_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Original Tiki product ID (e.g., 278600678)',
    category_id VARCHAR(36) DEFAULT NULL COMMENT 'FK to tiki_categories',
    category_name VARCHAR(255) DEFAULT NULL COMMENT 'Denormalized category name from crawl',
    name VARCHAR(500) NOT NULL,
    url VARCHAR(1000) NOT NULL COMMENT 'Full product URL',
    image_url VARCHAR(1000) NOT NULL,
    thumbnail_url VARCHAR(1000) DEFAULT NULL,
    brand VARCHAR(255) DEFAULT NULL,
    
    -- Price info (in VND, whole number)
    price BIGINT NOT NULL DEFAULT 0 COMMENT 'Current/sale price in VND',
    original_price BIGINT DEFAULT NULL COMMENT 'Original/list price in VND',
    discount_percent INT DEFAULT NULL COMMENT 'Discount percentage',
    
    -- Rating & reviews
    rating_average DECIMAL(3,2) DEFAULT NULL COMMENT 'Average rating (0-5)',
    rating_count INT DEFAULT NULL COMMENT 'Total number of ratings',
    review_count INT DEFAULT NULL COMMENT 'Total number of reviews (text reviews)',
    -- Sales
    sold_count INT DEFAULT NULL COMMENT 'Number of items sold (from "Đã bán X")',
    quantity_sold_text VARCHAR(100) DEFAULT NULL COMMENT 'Raw text like "Đã bán 1.3k"',
    
    -- Seller info
    seller_name VARCHAR(255) DEFAULT NULL COMMENT 'Seller/shop name',
    seller_avatar_url VARCHAR(1000) DEFAULT NULL,
    is_tiki_trading BOOLEAN DEFAULT FALSE COMMENT 'Whether sold by Tiki Trading (official)',
    is_official BOOLEAN DEFAULT FALSE COMMENT 'Whether this is an official shop product',
    is_sponsored BOOLEAN DEFAULT FALSE COMMENT 'Whether this is a sponsored/ad listing',
    
    -- Shipping & badges
    badge_text VARCHAR(255) DEFAULT NULL COMMENT 'Badge text like "Freeship", "Installment", etc.',
    shipping_info VARCHAR(255) DEFAULT NULL,
    freeship BOOLEAN DEFAULT FALSE,
    installment BOOLEAN DEFAULT FALSE COMMENT 'Has 0% installment option',
    
    -- Status & timestamps
    status ENUM('active','inactive','out_of_stock','unknown') NOT NULL DEFAULT 'active',
    crawl_page_num INT DEFAULT NULL COMMENT 'Page number where this product was found',
    crawled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_tiki_prod_tiki_id (tiki_product_id),
    INDEX idx_tiki_prod_category (category_id),
    INDEX idx_tiki_prod_price (price),
    INDEX idx_tiki_prod_brand (brand),
    INDEX idx_tiki_prod_seller (seller_name),
    INDEX idx_tiki_prod_sold (sold_count),
    INDEX idx_tiki_prod_rating (rating_average),
    INDEX idx_tiki_prod_discount (discount_percent),
    INDEX idx_tiki_prod_status (status),
    FULLTEXT INDEX idx_tiki_prod_search (name),
    
    FOREIGN KEY (category_id) REFERENCES tiki_categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Tiki Product Images (additional images per product)
-- ============================================================
CREATE TABLE IF NOT EXISTS tiki_product_images (
    id VARCHAR(36) PRIMARY KEY,
    tiki_product_id VARCHAR(50) NOT NULL,
    image_url VARCHAR(1000) NOT NULL,
    thumbnail_url VARCHAR(1000) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    crawled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tiki_img_product (tiki_product_id),
    FOREIGN KEY (tiki_product_id) REFERENCES tiki_products(tiki_product_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Crawl Jobs (track crawl progress)
-- ============================================================
CREATE TABLE IF NOT EXISTS tiki_crawl_jobs (
    id VARCHAR(36) PRIMARY KEY,
    category_url VARCHAR(500) NOT NULL,
    category_name VARCHAR(255) DEFAULT NULL,
    status ENUM('pending','running','completed','failed') NOT NULL DEFAULT 'pending',
    products_found INT NOT NULL DEFAULT 0,
    products_stored INT NOT NULL DEFAULT 0,
    pages_crawled INT NOT NULL DEFAULT 0,
    error_message TEXT DEFAULT NULL,
    started_at TIMESTAMP NULL DEFAULT NULL,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tiki_crawl_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
