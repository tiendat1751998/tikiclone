-- ============================================================
-- Migration 002: Fix Image URLs
-- Replaces external CDN URLs with local image paths
-- ============================================================

-- 1. SHOPEE_PRODUCT - product_images (legacy table)
-- Seed data already uses local paths, but update any external URLs
UPDATE tiki_product.product_images
SET url = '/images/products/product-1.jpg',
    alt_text = 'iPhone 15 Pro Max'
WHERE spu_id = 'spu-001' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-2.jpg',
    alt_text = 'Samsung Galaxy S24 Ultra'
WHERE spu_id = 'spu-002' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-3.jpg',
    alt_text = 'MacBook Pro 14'
WHERE spu_id = 'spu-003' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-4.jpg',
    alt_text = 'Sony WH-1000XM5'
WHERE spu_id = 'spu-004' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-5.jpg',
    alt_text = 'Nike Air Max 270'
WHERE spu_id = 'spu-005' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-6.jpg',
    alt_text = 'Adidas Ultraboost 22'
WHERE spu_id = 'spu-006' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-7.jpg',
    alt_text = 'Ergonomic Office Chair'
WHERE spu_id = 'spu-007' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-8.jpg',
    alt_text = 'Non-Stick Cookware Set'
WHERE spu_id = 'spu-008' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-9.jpg',
    alt_text = 'Cotton Crew Neck T-Shirt'
WHERE spu_id = 'spu-009' AND url LIKE 'http%';

UPDATE tiki_product.product_images
SET url = '/images/products/product-10.jpg',
    alt_text = 'Floral Summer Dress'
WHERE spu_id = 'spu-010' AND url LIKE 'http%';

-- 2. SHOPEE_PRODUCT - product_media
UPDATE tiki_product.product_media
SET url = '/images/products/product-1.jpg',
    thumbnail_url = '/images/products/product-1.jpg'
WHERE spu_id = 'spu-001' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-2.jpg',
    thumbnail_url = '/images/products/product-2.jpg'
WHERE spu_id = 'spu-002' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-3.jpg',
    thumbnail_url = '/images/products/product-3.jpg'
WHERE spu_id = 'spu-003' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-4.jpg',
    thumbnail_url = '/images/products/product-4.jpg'
WHERE spu_id = 'spu-004' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-5.jpg',
    thumbnail_url = '/images/products/product-5.jpg'
WHERE spu_id = 'spu-005' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-6.jpg',
    thumbnail_url = '/images/products/product-6.jpg'
WHERE spu_id = 'spu-006' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-7.jpg',
    thumbnail_url = '/images/products/product-7.jpg'
WHERE spu_id = 'spu-007' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-8.jpg',
    thumbnail_url = '/images/products/product-8.jpg'
WHERE spu_id = 'spu-008' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-9.jpg',
    thumbnail_url = '/images/products/product-9.jpg'
WHERE spu_id = 'spu-009' AND url LIKE 'http%';

UPDATE tiki_product.product_media
SET url = '/images/products/product-10.jpg',
    thumbnail_url = '/images/products/product-10.jpg'
WHERE spu_id = 'spu-010' AND url LIKE 'http%';

-- 3. SHOPEE_PLATFORM - categories (set default category image)
UPDATE tiki_platform.categories
SET image_url = '/images/categories/default-category.png'
WHERE image_url IS NOT NULL AND image_url LIKE 'http%';

-- 4. SHOPEE_PRODUCT - categories (service schema)
UPDATE tiki_product.categories
SET image_url = '/images/categories/default-category.png'
WHERE image_url IS NOT NULL AND image_url LIKE 'http%';

-- 6. SHOPEE_PLATFORM - users (default avatars)
UPDATE tiki_platform.users
SET avatar_url = '/images/avatars/default-avatar.png'
WHERE avatar_url IS NOT NULL AND avatar_url LIKE 'http%';

-- 7. SHOPEE_CART - cart_items (seed already uses local paths, catch any external)
UPDATE tiki_cart.cart_items
SET image_url = '/images/products/product-1.jpg'
WHERE sku = 'sku-001' AND image_url LIKE 'http%';

UPDATE tiki_cart.cart_items
SET image_url = '/images/products/product-4.jpg'
WHERE sku = 'sku-007' AND image_url LIKE 'http%';

UPDATE tiki_cart.cart_items
SET image_url = '/images/products/product-2.jpg'
WHERE sku = 'sku-003' AND image_url LIKE 'http%';

UPDATE tiki_cart.cart_items
SET image_url = '/images/products/product-5.jpg'
WHERE sku = 'sku-008' AND image_url LIKE 'http%';

UPDATE tiki_cart.cart_items
SET image_url = '/images/products/product-9.jpg'
WHERE sku = 'sku-013' AND image_url LIKE 'http%';

-- 10. Catch-all: blanket update for any remaining external URLs in product_media
UPDATE tiki_product.product_media
SET url = '/images/products/default-product.png',
    thumbnail_url = '/images/products/default-product.png'
WHERE url LIKE 'https://salt.tikicdn.com%'
   OR url LIKE 'https://cdn.example.com%'
   OR url LIKE 'https://store.storeimages%'
   OR url LIKE 'https://images.samsung.com%'
   OR url LIKE 'https://hoanghamobile.com%'
   OR url LIKE 'https://tiki.%';

-- 11. Catch-all: product_images
UPDATE tiki_product.product_images
SET url = '/images/products/default-product.png'
WHERE url LIKE 'https://salt.tikicdn.com%'
   OR url LIKE 'https://cdn.example.com%'
   OR url LIKE 'https://store.storeimages%'
   OR url LIKE 'https://images.samsung.com%'
   OR url LIKE 'https://hoanghamobile.com%'
   OR url LIKE 'https://tiki.%';
