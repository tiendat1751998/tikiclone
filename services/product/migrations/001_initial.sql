-- Product Service Database Migration
-- Version: 001_initial.sql

CREATE DATABASE IF NOT EXISTS tiki_product CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE tiki_product;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id VARCHAR(64) NOT NULL UNIQUE,
    title VARCHAR(512) NOT NULL,
    description TEXT,
    category_id VARCHAR(64) NOT NULL,
    brand_id VARCHAR(64) DEFAULT NULL,
    seller_id VARCHAR(64) NOT NULL,
    status ENUM('DRAFT', 'PENDING_REVIEW', 'ACTIVE', 'INACTIVE', 'REJECTED', 'DELETED') NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_products_category (category_id),
    INDEX idx_products_seller (seller_id),
    INDEX idx_products_brand (brand_id),
    INDEX idx_products_status (status),
    INDEX idx_products_created (created_at),
    INDEX idx_products_deleted (deleted_at),
    FULLTEXT INDEX idx_products_search (title, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- SKUs table
CREATE TABLE IF NOT EXISTS skus (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id VARCHAR(64) NOT NULL UNIQUE,
    spu_id VARCHAR(64) NOT NULL,
    price DECIMAL(12, 2) NOT NULL,
    sale_price DECIMAL(12, 2) DEFAULT NULL,
    stock INT NOT NULL DEFAULT 0,
    weight DECIMAL(8, 2) DEFAULT NULL,
    length DECIMAL(8, 2) DEFAULT NULL,
    width DECIMAL(8, 2) DEFAULT NULL,
    height DECIMAL(8, 2) DEFAULT NULL,
    status ENUM('ACTIVE', 'INACTIVE', 'OUT_OF_STOCK') NOT NULL DEFAULT 'ACTIVE',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_skus_spu (spu_id),
    INDEX idx_skus_status (status),
    INDEX idx_skus_price (price),
    FOREIGN KEY (spu_id) REFERENCES products(spu_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Product images table
CREATE TABLE IF NOT EXISTS product_images (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id VARCHAR(64) NOT NULL,
    url VARCHAR(1024) NOT NULL,
    alt_text VARCHAR(256) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_images_spu (spu_id),
    FOREIGN KEY (spu_id) REFERENCES products(spu_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    slug VARCHAR(256) NOT NULL UNIQUE,
    parent_id VARCHAR(64) DEFAULT NULL,
    level INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    image_url VARCHAR(1024) DEFAULT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_categories_parent (parent_id),
    INDEX idx_categories_level (level),
    INDEX idx_categories_active (is_active),
    INDEX idx_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Attributes table
CREATE TABLE IF NOT EXISTS attributes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    attribute_id VARCHAR(64) NOT NULL UNIQUE,
    category_id VARCHAR(64) NOT NULL,
    name VARCHAR(256) NOT NULL,
    type ENUM('TEXT', 'NUMBER', 'BOOLEAN', 'SELECT', 'MULTI_SELECT', 'COLOR') NOT NULL DEFAULT 'TEXT',
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    is_filterable BOOLEAN NOT NULL DEFAULT FALSE,
    is_searchable BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_attributes_category (category_id),
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Attribute values table
CREATE TABLE IF NOT EXISTS attribute_values (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    attribute_id VARCHAR(64) NOT NULL,
    value VARCHAR(256) NOT NULL,
    display_value VARCHAR(256) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_attr_values_attr (attribute_id),
    FOREIGN KEY (attribute_id) REFERENCES attributes(attribute_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Product attribute values table
CREATE TABLE IF NOT EXISTS product_attribute_values (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id VARCHAR(64) NOT NULL,
    attribute_id VARCHAR(64) NOT NULL,
    value_id VARCHAR(64) DEFAULT NULL,
    custom_value VARCHAR(512) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pav_spu (spu_id),
    INDEX idx_pav_attr (attribute_id),
    FOREIGN KEY (spu_id) REFERENCES products(spu_id) ON DELETE CASCADE,
    FOREIGN KEY (attribute_id) REFERENCES attributes(attribute_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Moderation records table
CREATE TABLE IF NOT EXISTS moderation_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id VARCHAR(64) NOT NULL,
    status ENUM('PENDING', 'APPROVED', 'REJECTED', 'FLAGGED') NOT NULL DEFAULT 'PENDING',
    reason TEXT DEFAULT NULL,
    reviewer_id VARCHAR(64) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_moderation_spu (spu_id),
    INDEX idx_moderation_status (status),
    FOREIGN KEY (spu_id) REFERENCES products(spu_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Product media table
CREATE TABLE IF NOT EXISTS product_media (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id VARCHAR(64) NOT NULL,
    sku_id VARCHAR(64) DEFAULT NULL,
    type ENUM('IMAGE', 'VIDEO', 'DOCUMENT') NOT NULL DEFAULT 'IMAGE',
    url VARCHAR(1024) NOT NULL,
    thumbnail_url VARCHAR(1024) DEFAULT NULL,
    alt_text VARCHAR(256) DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    mime_type VARCHAR(64) DEFAULT NULL,
    file_size INT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_media_spu (spu_id),
    INDEX idx_media_sku (sku_id),
    INDEX idx_media_type (type),
    FOREIGN KEY (spu_id) REFERENCES products(spu_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
