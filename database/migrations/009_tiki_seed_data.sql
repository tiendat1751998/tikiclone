-- ============================================================
-- TIKI PRODUCTS SEED DATA
-- Sản phẩm crawl từ tiki.vn + mock data cho các danh mục khác
-- Database: tiki_platform
-- ============================================================

USE tiki_platform;

-- ============================================================
-- 1. TIKI CATEGORIES
-- ============================================================
INSERT INTO tiki_categories (id, tiki_category_id, name, slug, url_path, sort_order) VALUES
('tcat-001', '1789', 'Điện Thoại - Máy Tính Bảng', 'dien-thoai-may-tinh-bang', '/dien-thoai-may-tinh-bang/c1789', 1),
('tcat-002', '1846', 'Laptop - Máy Vi Tính - Linh Kiện', 'laptop-may-vi-tinh-linh-kien', '/laptop-may-vi-tinh-linh-kien/c1846', 2),
('tcat-003', '1882', 'Điện Gia Dụng', 'dien-gia-dung', '/dien-gia-dung/c1882', 3),
('tcat-004', '1883', 'Nhà Cửa - Đời Sống', 'nha-cua-doi-song', '/nha-cua-doi-song/c1883', 4),
('tcat-005', '1520', 'Làm Đẹp - Sức Khỏe', 'lam-dep-suc-khoe', '/lam-dep-suc-khoe/c1520', 5),
('tcat-006', '2549', 'Mẹ & Bé', 'me-be', '/me-be/c2549', 6),
('tcat-007', '915', 'Thời Trang Nữ', 'thoi-trang-nu', '/thoi-trang-nu/c915', 7),
('tcat-008', '931', 'Thời Trang Nam', 'thoi-trang-nam', '/thoi-trang-nam/c931', 8),
('tcat-009', '1686', 'Giày - Dép Nam', 'giay-dep-nam', '/giay-dep-nam/c1686', 9),
('tcat-010', '1703', 'Giày - Dép Nữ', 'giay-dep-nu', '/giay-dep-nu/c1703', 10),
('tcat-011', '4221', 'Điện Tử - Điện Lạnh', 'dien-tu-dien-lanh', '/dien-tu-dien-lanh/c4221', 11),
('tcat-012', '8322', 'Nhà Sách Tiki', 'nha-sach-tiki', '/nha-sach-tiki/c8322', 12),
('tcat-013', '1975', 'Thể Thao - Dã Ngoại', 'the-thao-da-ngoai', '/the-thao-da-ngoai/c1975', 13),
('tcat-014', '8594', 'Ô Tô - Xe Máy - Xe Đạp', 'o-to-xe-may-xe-dap', '/o-to-xe-may-xe-dap/c8594', 14),
('tcat-015', '4384', 'Bách Hóa Online', 'bach-hoa-online', '/bach-hoa-online/c4384', 15),
('tcat-016', '27497', 'Đồng Hồ và Trang Sức', 'dong-ho-va-trang-suc', '/dong-ho-va-trang-suc/c27497', 16),
('tcat-017', '6000', 'Balo và Vali', 'balo-va-vali', '/balo-va-vali/c6000', 17),
('tcat-018', '27498', 'Phụ Kiện Thời Trang', 'phu-kien-thoi-trang', '/phu-kien-thoi-trang/c27498', 18),
('tcat-019', '914', 'Túi Thời Trang Nữ', 'tui-thoi-trang-nu', '/tui-thoi-trang-nu/c914', 19),
('tcat-020', '28856', 'Máy Đọc Sách', 'may-doc-sach', '/may-doc-sach/c28856', 20)
ON DUPLICATE KEY UPDATE name=VALUES(name), updated_at=NOW();

-- ============================================================
-- 2. TIKI PRODUCTS - Điện Thoại - Máy Tính Bảng (tcat-001)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, is_sponsored, status) VALUES
('tp-001', '278600678', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy S25 FE (8/128GB), Concert Camera 50MP, Pin bền bỉ, Trợ thủ AI thông minh - Hàng chính hãng - Xanh Navy', 'https://tiki.vn/dien-thoai-samsung-galaxy-s25-fe-8-128gb-concert-camera-50mp-pin-ben-bi-tro-thu-ai-thong-minh-hang-chinh-hang-p278600678.html', '/images/tiki/278600678.PNG', 11990000, 16900000, 29, 5.00, 263, '263', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-002', '278505394', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy A17 LTE (8/128GB), Kính Cường Lực Gorilla Victus, Camera 50MP & IOS, AI Gemini - Hàng Chính Hãng - Xanh', 'https://tiki.vn/dien-thoai-samsung-galaxy-a17-lte-8-128gb-kinh-cuong-luc-gorilla-victus-camera-50mp-ios-ai-gemini-hang-chinh-hang-p278505394.html', '/images/tiki/278505394.jpg', 4190000, 5240000, 20, 5.00, 169, '169', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-003', '279121902', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy A07 5G - Hàng Chính Hãng - Xanh (4GB/128GB)', 'https://tiki.vn/dien-thoai-samsung-galaxy-a07-5g-hang-chinh-hang-p279121902.html', '/images/tiki/279121902.jpg', 3290000, 4440000, 26, 5.00, 22, '22', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-004', '278890829', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy A17 5G, Camera 50MP & IOS, Kính Cường Lực Gorilla Victus, AI Gemini - Hàng Chính Hãng - Xanh (8GB/128GB)', 'https://tiki.vn/dien-thoai-samsung-galaxy-a17-5g-camera-50mp-ios-kinh-cuong-luc-gorilla-victus-ai-gemini-hang-chinh-hang-p278890829.html', '/images/tiki/278890829.jpg', 5090000, 6210000, 18, 5.00, 19, '19', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-005', '279185974', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Tai nghe nhét tai Samsung EO-IC100 Type C - Hàng Chính Hãng', 'https://tiki.vn/tai-nghe-nhet-tai-samsung-eo-ic100-type-c-hang-chinh-hang-p279185974.html', '/images/tiki/279185974.png', 190000, 560000, 66, 4.00, 116, '116', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-006', '279185842', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy S26+ (12GB/256GB) - Hàng Chính Hãng - Xanh', 'https://tiki.vn/dien-thoai-samsung-galaxy-s26-12gb-256gb-hang-chinh-hang-p279185842.html', '/images/tiki/279185842.jpg', 26990000, NULL, NULL, NULL, NULL, NULL, 'Tiki Trading', 1, 1, 1, 'active'),
('tp-007', '279185812', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy S26 (12GB/256GB) - Hàng Chính Hãng - Trắng', 'https://tiki.vn/dien-thoai-samsung-galaxy-s26-12gb-256gb-hang-chinh-hang-p279185812.html', '/images/tiki/279185812.jpg', 21990000, 26200000, 16, NULL, NULL, NULL, 'Tiki Trading', 1, 1, 1, 'active'),
('tp-008', '279185853', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy S26 Ultra (12GB/256GB) - Hàng Chính Hãng - Trắng', 'https://tiki.vn/dien-thoai-samsung-galaxy-s26-ultra-12gb-256gb-hang-chinh-hang-p279185853.html', '/images/tiki/279185853.jpg', 32990000, NULL, NULL, NULL, 4, 'Đã bán 4', 'Tiki Trading', 1, 1, 1, 'active'),
('tp-009', '279257510', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Điện Thoại Samsung Galaxy A57 5G (8GB/128GB) - Hàng Chính Hãng - Gray', 'https://tiki.vn/dien-thoai-samsung-galaxy-a57-5g-8gb-128gb-hang-chinh-hang-p279257510.html', '/images/tiki/279257510.jpg', 10990000, NULL, NULL, 5.00, NULL, NULL, 'Tiki Trading', 1, 1, 1, 'active'),
('tp-010', '279188180', 'tcat-001', 'Điện Thoại - Máy Tính Bảng', 'Máy Tính Bảng Galaxy Tab S10 Lite Wifi (8GB/256GB) - Hàng Chính Hãng - Xám', 'https://tiki.vn/may-tinh-bang-galaxy-tab-s10-lite-wifi-8gb-256gb-hang-chinh-hang-p279188180.html', '/images/tiki/279188180.png', 8490000, NULL, NULL, NULL, NULL, NULL, 'Tiki Trading', 1, 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 3. TIKI PRODUCTS - Laptop (tcat-002)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-011', '278100001', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'MacBook Air 13 inch M3 2024 8GB/256GB - Hàng Chính Hãng', 'https://tiki.vn/macbook-air-13-inch-m3-2024-p278100001.html', '/images/tiki/placeholder.svg', 28990000, 32990000, 12, 5.00, 1250, '1.2k', 'Tiki Trading', 1, 1, 'active'),
('tp-012', '278100002', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'MacBook Pro 14 inch M3 Pro 18GB/512GB - Hàng Chính Hãng', 'https://tiki.vn/macbook-pro-14-inch-m3-pro-p278100002.html', '/images/tiki/placeholder.svg', 49990000, 54990000, 9, 5.00, 340, '340', 'Tiki Trading', 1, 1, 'active'),
('tp-013', '278100003', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop ASUS Vivobook 15 OLED A1505VA - i5 13500H/16GB/512GB', 'https://tiki.vn/laptop-asus-vivobook-15-oled-a1505va-p278100003.html', '/images/tiki/placeholder.svg', 18990000, 22990000, 17, 4.50, 890, '890', 'Tiki Trading', 1, 1, 'active'),
('tp-014', '278100004', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop Dell Inspiron 15 3520 - i5 1235U/8GB/512GB', 'https://tiki.vn/laptop-dell-inspiron-15-3520-p278100004.html', '/images/tiki/placeholder.svg', 14990000, 17990000, 17, 4.00, 560, '560', 'Tiki Trading', 1, 1, 'active'),
('tp-015', '278100005', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop Lenovo IdeaPad Slim 3 15IAH8 - i5 12450H/8GB/512GB', 'https://tiki.vn/laptop-lenovo-ideapad-slim-3-p278100005.html', '/images/tiki/placeholder.svg', 13490000, 16990000, 21, 4.50, 1100, '1.1k', 'Tiki Trading', 1, 1, 'active'),
('tp-016', '278100006', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop Acer Aspire 5 15 - i5 1335U/16GB/512GB', 'https://tiki.vn/laptop-acer-aspire-5-p278100006.html', '/images/tiki/placeholder.svg', 15990000, 19990000, 20, 4.00, 420, '420', 'Tiki Trading', 1, 1, 'active'),
('tp-017', '278100007', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop HP Pavilion 15 - i7 1355U/16GB/512GB', 'https://tiki.vn/laptop-hp-pavilion-15-p278100007.html', '/images/tiki/placeholder.svg', 19990000, 23990000, 17, 4.50, 280, '280', 'Tiki Trading', 1, 1, 'active'),
('tp-018', '278100008', 'tcat-002', 'Laptop - Máy Vi Tính - Linh Kiện', 'Laptop MSI Gaming GF63 Thin - i5 12450H/8GB/512GB/RTX 2050', 'https://tiki.vn/laptop-msi-gaming-gf63-thin-p278100008.html', '/images/tiki/placeholder.svg', 17990000, 22990000, 22, 4.50, 670, '670', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 4. TIKI PRODUCTS - Điện Gia Dụng (tcat-003)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-019', '278200001', 'tcat-003', 'Điện Gia Dụng', 'Máy lọc không khí Xiaomi Air Purifier 4 Compact', 'https://tiki.vn/may-loc-khong-khi-xiaomi-air-purifier-4-compact-p278200001.html', '/images/tiki/placeholder.svg', 2490000, 3490000, 29, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-020', '278200002', 'tcat-003', 'Điện Gia Dụng', 'Nồi cơm điện tử Toshiba 1.8L RC-18DRM', 'https://tiki.vn/noi-com-dien-tu-toshiba-1-8l-rc-18drm-p278200002.html', '/images/tiki/placeholder.svg', 890000, 1290000, 31, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-021', '278200003', 'tcat-003', 'Điện Gia Dụng', 'Bàn ủi hơi nước Philips Azur 8000 Series DST8041/80', 'https://tiki.vn/ban-ui-hoi-nuoc-philips-azur-8000-series-p278200003.html', '/images/tiki/placeholder.svg', 1590000, 2190000, 27, 4.00, 1800, '1.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-022', '278200004', 'tcat-003', 'Điện Gia Dụng', 'Máy xay sinh tố Vitamix E310', 'https://tiki.vn/may-xay-sinh-tto-vitamix-e310-p278200004.html', '/images/tiki/placeholder.svg', 8990000, 11990000, 25, 5.00, 340, '340', 'Tiki Trading', 1, 1, 'active'),
('tp-023', '278200005', 'tcat-003', 'Điện Gia Dụng', 'Robot hút bụi Xiaomi Robot Vacuum S10+', 'https://tiki.vn/robot-hut-bui-xiaomi-robot-vacuum-s10-p278200005.html', '/images/tiki/placeholder.svg', 9990000, 13990000, 29, 4.50, 890, '890', 'Tiki Trading', 1, 1, 'active'),
('tp-024', '278200006', 'tcat-003', 'Điện Gia Dụng', 'Lò vi sóng Samsung 23L MS23F300EEV/SV', 'https://tiki.vn/lo-vi-song-samsung-23l-ms23f300eev-p278200006.html', '/images/tiki/placeholder.svg', 1890000, 2490000, 24, 4.00, 3200, '3.2k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 5. TIKI PRODUCTS - Nhà Cửa - Đời Sống (tcat-004)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-025', '278300001', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Bộ chén sứ 12 miếng Minh Long I - Hàng Chính Hãng', 'https://tiki.vn/bo-chen-su-12-mieng-minh-long-i-p278300001.html', '/images/tiki/placeholder.svg', 390000, 590000, 34, 4.50, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active'),
('tp-026', '278300002', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Nệm cao su thiên nhiên Vạn Thành 160x200x15cm', 'https://tiki.vn/nem-cao-su-thien-nhien-van-thanh-p278300002.html', '/images/tiki/placeholder.svg', 3990000, 5990000, 33, 4.50, 1200, '1.2k', 'Tiki Trading', 1, 1, 'active'),
('tp-027', '278300003', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Bàn làm việc gỗ MDF Hiện đại 120x60x75cm', 'https://tiki.vn/ban-lam-viec-go-mdf-hien-dai-p278300003.html', '/images/tiki/placeholder.svg', 1290000, 1890000, 32, 4.00, 890, '890', 'Tiki Trading', 1, 1, 'active'),
('tp-028', '278300004', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Tủ quần áo 2 cánh gỗ MDF phong cách Bắc Âu', 'https://tiki.vn/tu-quan-ao-2-canh-go-mdf-p278300004.html', '/images/tiki/placeholder.svg', 2990000, 4490000, 33, 4.50, 560, '560', 'Tiki Trading', 1, 1, 'active'),
('tp-029', '278300005', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Bộ khăn tắm cao cấp 3 miếng 70x140cm', 'https://tiki.vn/bo-khan-tam-cao-cap-3-mieng-p278300005.html', '/images/tiki/placeholder.svg', 199000, 349000, 43, 4.00, 7800, '7.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-030', '278300006', 'tcat-004', 'Nhà Cửa - Đời Sống', 'Đèn bàn LED đọc sách chống cận thị Baseus', 'https://tiki.vn/den-ban-led-doc-sach-chong-can-thi-baseus-p278300006.html', '/images/tiki/placeholder.svg', 490000, 790000, 38, 4.50, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 6. TIKI PRODUCTS - Làm Đẹp - Sức Khỏe (tcat-005)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-031', '278400001', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Kem chống nắng La Roche-Posay Anthelios UV Mune 400 SPF50+ 50ml', 'https://tiki.vn/kem-chong-nang-la-roche-posay-anthelios-p278400001.html', '/images/tiki/placeholder.svg', 429000, 549000, 22, 4.50, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-032', '278400002', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Serum Vitamin C The Ordinary 30ml', 'https://tiki.vn/serum-vitamin-c-the-ordinary-30ml-p278400002.html', '/images/tiki/placeholder.svg', 299000, 449000, 33, 4.50, 12000, '12k', 'Tiki Trading', 1, 1, 'active'),
('tp-033', '278400003', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Sữa rửa mặt Cetaphil Gentle Skin Cleanser 250ml', 'https://tiki.vn/sua-rua-mat-cetaphil-gentle-skin-cleanser-p278400003.html', '/images/tiki/placeholder.svg', 249000, 349000, 29, 4.50, 15000, '15k', 'Tiki Trading', 1, 1, 'active'),
('tp-034', '278400004', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Mặt nạ dưỡng da Mediheal Tea Tree Essential Mask 10 miếng', 'https://tiki.vn/mat-na-duong-da-mediheal-tea-tree-p278400004.html', '/images/tiki/placeholder.svg', 129000, 199000, 35, 4.00, 25000, '25k', 'Tiki Trading', 1, 1, 'active'),
('tp-035', '278400005', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Nước tẩy trang Bioderma Sensibio H2O 500ml', 'https://tiki.vn/nuoc-tay-trang-bioderma-sensibio-h2o-p278400005.html', '/images/tiki/placeholder.svg', 349000, 499000, 30, 4.50, 9800, '9.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-036', '278400006', 'tcat-005', 'Làm Đẹp - Sức Khỏe', 'Kem dưỡng ẩt CeraVe Moisturizing Cream 50ml', 'https://tiki.vn/kem-duong-am-cerave-moisturizing-cream-p278400006.html', '/images/tiki/placeholder.svg', 199000, 299000, 33, 4.50, 18000, '18k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 7. TIKI PRODUCTS - Mẹ & Bé (tcat-006)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-037', '278500001', 'tcat-006', 'Mẹ & Bé', 'Tã quần Pampers Newborn 1 84 miếng (dưới 5kg)', 'https://tiki.vn/ta-quan-pampers-newborn-1-84-mieng-p278500001.html', '/images/tiki/placeholder.svg', 249000, 349000, 29, 5.00, 45000, '45k', 'Tiki Trading', 1, 1, 'active'),
('tp-038', '278500002', 'tcat-006', 'Mẹ & Bé', 'Sữa bột Enfamil A+ Neuro Pro 1 850g (0-6 tháng)', 'https://tiki.vn/sua-bot-enfamil-a-neuro-pro-1-850g-p278500002.html', '/images/tiki/placeholder.svg', 599000, 749000, 20, 5.00, 12000, '12k', 'Tiki Trading', 1, 1, 'active'),
('tp-039', '278500003', 'tcat-006', 'Mẹ & Bé', 'Xe đẩy em bé Aprica FLiora Plus - Hàng Nhập Khẩu', 'https://tiki.vn/xe-day-em-be-aprica-fliora-plus-p278500003.html', '/images/tiki/placeholder.svg', 4990000, 6990000, 29, 4.50, 890, '890', 'Tiki Trading', 1, 1, 'active'),
('tp-040', '278500004', 'tcat-006', 'Mẹ & Bé', 'Ghế ô tô cho bé Joie i-Spin 360 (0-4 tuổi)', 'https://tiki.vn/ghe-o-to-cho-be-joie-i-spin-360-p278500004.html', '/images/tiki/placeholder.svg', 3990000, 5490000, 27, 4.50, 560, '560', 'Tiki Trading', 1, 1, 'active'),
('tp-041', '278500005', 'tcat-006', 'Mẹ & Bé', 'Bình sức Philips Avent Natural 260ml', 'https://tiki.vn/binh-sua-philips-avent-natural-260ml-p278500005.html', '/images/tiki/placeholder.svg', 199000, 299000, 33, 4.50, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-042', '278500006', 'tcat-006', 'Mẹ & Bé', 'Máy hút sữa điện đôi Medela Freestyle Flex', 'https://tiki.vn/may-hut-sua-dien-doi-medela-freestyle-flex-p278500006.html', '/images/tiki/placeholder.svg', 5990000, 7990000, 25, 5.00, 340, '340', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 8. TIKI PRODUCTS - Thời Trang Nữ (tcat-007)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-043', '278600001', 'tcat-007', 'Thời Trang Nữ', 'Đầm hoa nhí cổ vuông tay lỡ phong cách Hàn Quốc', 'https://tiki.vn/dam-hoa-ni-co-vuong-tay-lo-p278600001.html', '/images/tiki/placeholder.svg', 199000, 349000, 43, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-044', '278600002', 'tcat-007', 'Thời Trang Nữ', 'Áo sơ mi nữ cổ bẻ tay dài vải lụa mềm mại', 'https://tiki.vn/ao-so-mi-nu-co-be-tay-dai-vai-lua-p278600002.html', '/images/tiki/placeholder.svg', 249000, 399000, 38, 4.00, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-045', '278600003', 'tcat-007', 'Thời Trang Nữ', 'Quần jean nữ ống rộng cạp cao co giãn tốt', 'https://tiki.vn/quan-jean-nu-ong-rong-cao-cap-p278600003.html', '/images/tiki/placeholder.svg', 299000, 499000, 40, 4.50, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-046', '278600004', 'tcat-007', 'Thời Trang Nữ', 'Áo khoác blazer nữ dáng dài thanh lịch', 'https://tiki.vn/ao-khoac-blazer-nu-dang-dai-thanh-lich-p278600004.html', '/images/tiki/placeholder.svg', 449000, 699000, 36, 4.00, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-047', '278600005', 'tcat-007', 'Thời Trang Nữ', 'Chân váy chữ A ngắn phong cách Nhật Bản', 'https://tiki.vn/chan-vay-chu-a-ngan-phong-cach-nhat-ban-p278600005.html', '/images/tiki/placeholder.svg', 179000, 299000, 40, 4.50, 6700, '6.7k', 'Tiki Trading', 1, 1, 'active'),
('tp-048', '278600006', 'tcat-007', 'Thời Trang Nữ', 'Set đồ thể thao nữ áo bra + quần short', 'https://tiki.vn/set-do-the-thao-nu-ao-bra-quan-short-p278600006.html', '/images/tiki/placeholder.svg', 159000, 279000, 43, 4.00, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 9. TIKI PRODUCTS - Thời Trang Nam (tcat-008)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-049', '278700001', 'tcat-008', 'Thời Trang Nam', 'Áo thun nam cổ tròn cotton 100% dáng oversize', 'https://tiki.vn/ao-thun-nam-co-tron-cotton-100-p278700001.html', '/images/tiki/placeholder.svg', 149000, 249000, 40, 4.50, 12000, '12k', 'Tiki Trading', 1, 1, 'active'),
('tp-050', '278700002', 'tcat-008', 'Thời Trang Nam', 'Quần jean nam ống đứng wash nhẹ co giãn', 'https://tiki.vn/quan-jean-nam-ong-dung-wash-nhe-p278700002.html', '/images/tiki/placeholder.svg', 349000, 549000, 36, 4.00, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-051', '278700003', 'tcat-008', 'Thời Trang Nam', 'Áo sơ mi nam dài tay vải kate thoáng mát', 'https://tiki.vn/ao-so-mi-nam-dai-tay-vai-kate-p278700003.html', '/images/tiki/placeholder.svg', 249000, 399000, 38, 4.50, 6700, '6.7k', 'Tiki Trading', 1, 1, 'active'),
('tp-052', '278700004', 'tcat-008', 'Thời Trang Nam', 'Áo khoác bomber nam phong cách streetwear', 'https://tiki.vn/ao-khoac-bomber-nam-phong-cach-streetwear-p278700004.html', '/images/tiki/placeholder.svg', 399000, 649000, 38, 4.00, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-053', '278700005', 'tcat-008', 'Thời Trang Nam', 'Quần short nam kaki dáng slimfit', 'https://tiki.vn/quan-short-nam-kaki-dang-slimfit-p278700005.html', '/images/tiki/placeholder.svg', 199000, 349000, 43, 4.00, 9800, '9.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-054', '278700006', 'tcat-008', 'Thời Trang Nam', 'Áo polo nam pique cotton cổ bẻ', 'https://tiki.vn/ao-polo-nam-pique-cotton-co-be-p278700006.html', '/images/tiki/placeholder.svg', 229000, 379000, 40, 4.50, 7600, '7.6k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 10. TIKI PRODUCTS - Giày Dép Nam (tcat-009)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-055', '278800001', 'tcat-009', 'Giày - Dép Nam', 'Giày thể thao Nike Air Max 270 - Hàng Chính Hãng', 'https://tiki.vn/giay-the-thao-nike-air-max-270-p278800001.html', '/images/tiki/placeholder.svg', 3490000, 4990000, 30, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-056', '278800002', 'tcat-009', 'Giày - Dép Nam', 'Giày chạy bộ Adidas Ultraboost Light - Hàng Chính Hãng', 'https://tiki.vn/giay-chay-bo-adidas-ultraboost-light-p278800002.html', '/images/tiki/placeholder.svg', 4290000, 5990000, 28, 4.50, 1800, '1.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-057', '278800003', 'tcat-009', 'Giày - Dép Nam', 'Giày lười nam da bò thật ECCO - Hàng Nhập Khẩu', 'https://tiki.vn/giay-luoi-nam-da-bo-that-ecco-p278800003.html', '/images/tiki/placeholder.svg', 2990000, 4490000, 33, 4.00, 890, '890', 'Tiki Trading', 1, 1, 'active'),
('tp-058', '278800004', 'tcat-009', 'Giày - Dép Nam', 'Dép nam quai ngang da thật Porsches Design', 'https://tiki.vn/dep-nam-quai-ngang-da-that-porsches-design-p278800004.html', '/images/tiki/placeholder.svg', 599000, 999000, 40, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-059', '278800005', 'tcat-009', 'Giày - Dép Nam', 'Giày sneaker New Balance 574 - Hàng Chính Hãng', 'https://tiki.vn/giay-sneaker-new-balance-574-p278800005.html', '/images/tiki/placeholder.svg', 2190000, 2990000, 27, 4.00, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-060', '278800006', 'tcat-009', 'Giày - Dép Nam', 'Giày tây nam da bò Oxford - Hàng Chính Hãng', 'https://tiki.vn/giay-tay-nam-da-bo-oxford-p278800006.html', '/images/tiki/placeholder.svg', 1890000, 2790000, 32, 4.50, 1200, '1.2k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 11. TIKI PRODUCTS - Giày Dép Nữ (tcat-010)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-061', '278900001', 'tcat-010', 'Giày - Dép Nữ', 'Giày cao gót mũi nhọn 7cm da thật', 'https://tiki.vn/giay-cao-got-mui-nhon-7cm-da-that-p278900001.html', '/images/tiki/placeholder.svg', 499000, 899000, 44, 4.50, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-062', '278900002', 'tcat-010', 'Giày - Dép Nữ', 'Giày sneaker nữ Nike Air Force 1 - Hàng Chính Hãng', 'https://tiki.vn/giay-sneaker-nu-nike-air-force-1-p278900002.html', '/images/tiki/placeholder.svg', 2490000, 3490000, 29, 4.50, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active'),
('tp-063', '278900003', 'tcat-010', 'Giày - Dép Nữ', 'Giày búp bê nữ da mềm quai hậu', 'https://tiki.vn/giay-bup-be-nu-da-mem-quai-hau-p278900003.html', '/images/tiki/placeholder.svg', 349000, 599000, 42, 4.00, 6700, '6.7k', 'Tiki Trading', 1, 1, 'active'),
('tp-064', '278900004', 'tcat-010', 'Giày - Dép Nữ', 'Dép nữ xỏ ngón quai ngang da thật', 'https://tiki.vn/dep-nu-xo-ngon-quai-ngang-da-that-p278900004.html', '/images/tiki/placeholder.svg', 299000, 499000, 40, 4.50, 12000, '12k', 'Tiki Trading', 1, 1, 'active'),
('tp-065', '278900005', 'tcat-010', 'Giày - Dép Nữ', 'Giày sandal nữ cao gót 5cm đế bệt', 'https://tiki.vn/giay-sandal-nu-cao-got-5cm-de-bet-p278900005.html', '/images/tiki/placeholder.svg', 399000, 699000, 43, 4.00, 9800, '9.8k', 'Tiki Trading', 1, 1, 'active'),
('tp-066', '278900006', 'tcat-010', 'Giày - Dép Nữ', 'Giày thể thao nữ Adidas Stan Smith - Hàng Chính Hãng', 'https://tiki.vn/giay-the-thao-nu-adidas-stan-smith-p278900006.html', '/images/tiki/placeholder.svg', 1990000, 2790000, 29, 4.50, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 12. TIKI PRODUCTS - Sách (tcat-012)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-067', '279000001', 'tcat-012', 'Nhà Sách Tiki', 'Nhà Giả Kim - Paulo Coelho (Tái bản 2024)', 'https://tiki.vn/nha-gia-kim-paulo-coelho-p279000001.html', '/images/tiki/placeholder.svg', 79000, 129000, 39, 5.00, 45000, '45k', 'Tiki Trading', 1, 1, 'active'),
('tp-068', '279000002', 'tcat-012', 'Nhà Sách Tiki', 'Đắc Nhân Tâm - Dale Carnegie (Bìa cứng)', 'https://tiki.vn/dac-nhan-tam-dale-carnegie-p279000002.html', '/images/tiki/placeholder.svg', 99000, 169000, 41, 5.00, 67000, '67k', 'Tiki Trading', 1, 1, 'active'),
('tp-069', '279000003', 'tcat-012', 'Nhà Sách Tiki', 'Tôi Thấy Hoa Vàng Trên Cỏ Xanh - Nguyễn Nhật Ánh', 'https://tiki.vn/toi-thay-hoa-vang-tren-co-xanh-p279000003.html', '/images/tiki/placeholder.svg', 89000, 139000, 36, 5.00, 34000, '34k', 'Tiki Trading', 1, 1, 'active'),
('tp-070', '279000004', 'tcat-012', 'Nhà Sách Tiki', 'Cà Phê Công Papa Và Bà Tôi - Nguyễn Nhật Ánh', 'https://tiki.vn/ca-phe-cung-papa-va-ba-toi-p279000004.html', '/images/tiki/placeholder.svg', 109000, 159000, 31, 5.00, 23000, '23k', 'Tiki Trading', 1, 1, 'active'),
('tp-071', '279000005', 'tcat-012', 'Nhà Sách Tiki', 'Dám Bị Ghét - Koga Fumitake', 'https://tiki.vn/dam-bi-ghet-koga-fumitake-p279000005.html', '/images/tiki/placeholder.svg', 119000, 189000, 37, 4.50, 18000, '18k', 'Tiki Trading', 1, 1, 'active'),
('tp-072', '279000006', 'tcat-012', 'Nhà Sách Tiki', 'Atomic Habits - James Clear (Bìa cứng)', 'https://tiki.vn/atomic-habits-james-clear-p279000006.html', '/images/tiki/placeholder.svg', 199000, 299000, 33, 5.00, 12000, '12k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 13. TIKI PRODUCTS - Thể Thao (tcat-013)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-073', '279100001', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Xạp tập yoga 6mm chống trượt Lululemon', 'https://tiki.vn/xap-tap-yoga-6mm-chong-truot-lululemon-p279100001.html', '/images/tiki/placeholder.svg', 129000, 199000, 35, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-074', '279100002', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Bóng đá Nike Premier League Flight - Size 5', 'https://tiki.vn/bong-da-nike-premier-league-flight-p279100002.html', '/images/tiki/placeholder.svg', 499000, 799000, 38, 4.50, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-075', '279100003', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Vợt cầu lông Yonex Nanoray 10F - Hàng Chính Hãng', 'https://tiki.vn/vot-cau-long-yonex-nanoray-10f-p279100003.html', '/images/tiki/placeholder.svg', 890000, 1290000, 31, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-076', '279100004', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Balo leo núi 40L The North Face Borealis', 'https://tiki.vn/balo-leo-nui-40l-the-north-face-borealis-p279100004.html', '/images/tiki/placeholder.svg', 2490000, 3490000, 29, 4.50, 1200, '1.2k', 'Tiki Trading', 1, 1, 'active'),
('tp-077', '279100005', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Xe đạp thể thao Giant Escape 3 - Khung nhôm', 'https://tiki.vn/xe-dap-the-thao-giant-escape-3-p279100005.html', '/images/tiki/placeholder.svg', 7990000, 10990000, 27, 4.00, 560, '560', 'Tiki Trading', 1, 1, 'active'),
('tp-078', '279100006', 'tcat-013', 'Thể Thao - Dã Ngoại', 'Tạ tay điều chỉnh 2x10kg Bowflex SelectTech', 'https://tiki.vn/ta-tay-dieu-chinh-2x10kg-bowflex-selecttech-p279100006.html', '/images/tiki/placeholder.svg', 2990000, 4490000, 33, 4.50, 890, '890', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 14. TIKI PRODUCTS - Đồng Hồ (tcat-016)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-079', '279200001', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Đồng hồ nam Casio MTP-V002L-7BUDF - Nhôm/Mặt số', 'https://tiki.vn/dong-ho-nam-casio-mtp-v002l-7budf-p279200001.html', '/images/tiki/placeholder.svg', 599000, 899000, 33, 4.50, 12000, '12k', 'Tiki Trading', 1, 1, 'active'),
('tp-080', '279200002', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Đồng hồ nữ Daniel Wellington Classic Petite 32mm', 'https://tiki.vn/dong-ho-nu-daniel-wellington-classic-petite-p279200002.html', '/images/tiki/placeholder.svg', 2490000, 3490000, 29, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-081', '279200003', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Đồng hồ nam Orient RA-AC0M01B10B - Automatic', 'https://tiki.vn/dong-ho-nam-orient-ra-ac0m01b10b-p279200003.html', '/images/tiki/placeholder.svg', 5990000, 7990000, 25, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-082', '279200004', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Nhẫn bạc Ý 925 đính đá Swarovski', 'https://tiki.vn/nhan-bac-y-925-dinh-da-swarovski-p279200004.html', '/images/tiki/placeholder.svg', 499000, 799000, 38, 4.00, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-083', '279200005', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Vòng tay Pandora Moments dây da', 'https://tiki.vn/vong-tay-pandora-moments-day-da-p279200005.html', '/images/tiki/placeholder.svg', 1290000, 1890000, 32, 4.50, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-084', '279200006', 'tcat-016', 'Đồng Hồ và Trang Sức', 'Đồng hồ Apple Watch SE 2024 40mm GPS', 'https://tiki.vn/dong-ho-apple-watch-se-2024-40mm-gps-p279200006.html', '/images/tiki/placeholder.svg', 5990000, 7490000, 20, 5.00, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 15. TIKI PRODUCTS - Bách Hóa Online (tcat-015)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-085', '279300001', 'tcat-015', 'Bách Hóa Online', 'Mì ăn liền Hảo Hảo tôm chua cay 75g x 30 gói', 'https://tiki.vn/mi-an-lien-hao-hao-tom-chua-cay-75g-x-30-goi-p279300001.html', '/images/tiki/placeholder.svg', 89000, 129000, 31, 4.50, 67000, '67k', 'Tiki Trading', 1, 1, 'active'),
('tp-086', '279300002', 'tcat-015', 'Bách Hóa Online', 'Nước mắm Nam Ngư 500ml chai nhựa', 'https://tiki.vn/nuoc-mam-nam-ngu-500ml-chai-nhua-p279300002.html', '/images/tiki/placeholder.svg', 32000, 45000, 29, 4.50, 45000, '45k', 'Tiki Trading', 1, 1, 'active'),
('tp-087', '279300003', 'tcat-015', 'Bách Hóa Online', 'Dầu ăn Simply 1L - Hàng Việt Nam', 'https://tiki.vn/dau-an-simply-1l-p279300003.html', '/images/tiki/placeholder.svg', 49000, 69000, 29, 4.50, 34000, '34k', 'Tiki Trading', 1, 1, 'active'),
('tp-088', '279300004', 'tcat-015', 'Bách Hóa Online', 'Sữa tươi tiệt trùng Vinamilk 100% 1L x 6 hộp', 'https://tiki.vn/sua-tuoi-tiet-trung-vinamilk-100-1l-x-6-hop-p279300004.html', '/images/tiki/placeholder.svg', 169000, 229000, 26, 5.00, 23000, '23k', 'Tiki Trading', 1, 1, 'active'),
('tp-089', '279300005', 'tcat-015', 'Bách Hóa Online', 'Bột giặt Omo Matic 3.8kg - Hương nước xả Downy', 'https://tiki.vn/bot-giat-omo-matic-3-8kg-p279300005.html', '/images/tiki/placeholder.svg', 149000, 199000, 25, 4.50, 18000, '18k', 'Tiki Trading', 1, 1, 'active'),
('tp-090', '279300006', 'tcat-015', 'Bách Hóa Online', 'Khăn giấy Bless you gấp 3 lớp 250 tờ x 6 cuộn', 'https://tiki.vn/khan-giay-bless-you-gap-3-lop-250-to-x-6-cuon-p279300006.html', '/images/tiki/placeholder.svg', 79000, 119000, 34, 4.00, 12000, '12k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 16. TIKI PRODUCTS - Điện Tử - Điện Lạnh (tcat-011)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-091', '279400001', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Tivi Samsung 4K Crystal UHD 55 inch UA55DU8000', 'https://tiki.vn/tivi-samsung-4k-crystal-uhd-55-inch-ua55du8000-p279400001.html', '/images/tiki/placeholder.svg', 10990000, 14990000, 27, 4.50, 3400, '3.4k', 'Tiki Trading', 1, 1, 'active'),
('tp-092', '279400002', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Máy lạnh Daikin 12000BTU FTKA25UAVMV inverter 1 chiều', 'https://tiki.vn/may-lanh-daikin-12000btu-ftka25uavmv-p279400002.html', '/images/tiki/placeholder.svg', 8990000, 11990000, 25, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-093', '279400003', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Loa Bluetooth JBL Charge 5 - Chống nước IP67', 'https://tiki.vn/loa-bluetooth-jbl-charge-5-p279400003.html', '/images/tiki/placeholder.svg', 3490000, 4990000, 30, 4.50, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-094', '279400004', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Tai nghe Sony WH-1000XM5 - Chống ồn chủ động', 'https://tiki.vn/tai-nghe-sony-wh-1000xm5-p279400004.html', '/images/tiki/placeholder.svg', 6990000, 8990000, 22, 5.00, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-095', '279400005', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Máy tính bảng iPad Air M2 11 inch 128GB WiFi', 'https://tiki.vn/may-tinh-bang-ipad-air-m2-11-inch-p279400005.html', '/images/tiki/placeholder.svg', 14990000, 17990000, 17, 5.00, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active'),
('tp-096', '279400006', 'tcat-011', 'Điện Tử - Điện Lạnh', 'Ổ cứng di động Samsung T7 Shield 1TB', 'https://tiki.vn/o-cung-di-dong-samsung-t7-shield-1tb-p279400006.html', '/images/tiki/placeholder.svg', 1990000, 2790000, 29, 4.50, 6700, '6.7k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

-- ============================================================
-- 17. TIKI PRODUCTS - Balo và Vali (tcat-017)
-- ============================================================
INSERT INTO tiki_products (id, tiki_product_id, category_id, category_name, name, url, image_url, price, original_price, discount_percent, rating_average, sold_count, quantity_sold_text, seller_name, is_tiki_trading, is_official, status) VALUES
('tp-097', '279500001', 'tcat-017', 'Balo và Vali', 'Balo laptop 15.6 inch Samsonite Classic - Chống sốc', 'https://tiki.vn/balo-laptop-15-6-inch-samsonite-classic-p279500001.html', '/images/tiki/placeholder.svg', 1290000, 1890000, 32, 4.50, 5600, '5.6k', 'Tiki Trading', 1, 1, 'active'),
('tp-098', '279500002', 'tcat-017', 'Balo và Vali', 'Vali kéo 24 inch American Tourister Curio - ABS+PC', 'https://tiki.vn/vali-keo-24-inch-american-tourister-curio-p279500002.html', '/images/tiki/placeholder.svg', 3490000, 4990000, 30, 4.50, 2300, '2.3k', 'Tiki Trading', 1, 1, 'active'),
('tp-099', '279500003', 'tcat-017', 'Balo và Vali', 'Balo thể thao Adidas Classic 3S 25L', 'https://tiki.vn/balo-the-thao-adidas-classic-3s-25l-p279500003.html', '/images/tiki/placeholder.svg', 599000, 899000, 33, 4.00, 8900, '8.9k', 'Tiki Trading', 1, 1, 'active'),
('tp-100', '279500004', 'tcat-017', 'Balo và Vali', 'Vali kéo 20 inch Xiaomi City Carry-On - Nhôm-magiê', 'https://tiki.vn/vali-keo-20-inch-xiaomi-city-carry-on-p279500004.html', '/images/tiki/placeholder.svg', 2490000, 3490000, 29, 4.50, 4500, '4.5k', 'Tiki Trading', 1, 1, 'active')
ON DUPLICATE KEY UPDATE name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price), discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count), rating_average=VALUES(rating_average), updated_at=NOW();

SELECT 'Seed data loaded successfully!' AS status;
SELECT COUNT(*) AS total_products FROM tiki_products;
SELECT COUNT(*) AS total_categories FROM tiki_categories;
