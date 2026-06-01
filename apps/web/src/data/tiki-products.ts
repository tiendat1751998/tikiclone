import { Product } from "@/types";

function createProduct(id: string, shop_id: string, category_id: string, name: string, price: number, original_price: number, discount_percent: number, sold_count: number, rating_average: number, rating_count: number, review_count: number, seller_name: string, image_url: string, is_official = false): Product {
  return {
    id, shop_id, category_id, name, description: "", price, original_price, discount_percent, stock: 100, sold_count, quantity_sold_text: `Đã bán ${sold_count}`, rating_average, rating_count, review_count, seller_name, is_official, status: "active", condition: "new", created_at: "2024-01-01", updated_at: "2024-01-01", image_url
  };
}

export const CATEGORY_PRODUCTS: Record<string, Product[]> = {
  "nha-sach-tiki": [
    createProduct("sp-ns1", "shop-001", "nha-sach-tiki", "Sách Bán Chạy Nhất 2024", 89000, 120000, 25, 1250, 4.8, 156, 67, "Nhà Sách Tiki", "/images/products/72709918.png", false),
    createProduct("sp-ns2", "shop-001", "nha-sach-tiki", "Trạng Tính - Paulo Coelho", 62000, 78000, 20, 890, 4.7, 112, 45, "Tiki Trading", "/images/products/72712371.png", false),
    createProduct("sp-ns3", "shop-001", "nha-sach-tiki", "Tôi Thì Từ Tính", 75000, 75000, 0, 2100, 4.6, 98, 67, "Nhà Sách Tiki", "/images/products/72723335.jpg", false),
    createProduct("sp-ns4", "shop-001", "nha-sach-tiki", "Doraemon - Tập Truyện", 45000, 65000, 30, 345, 4.9, 89, 34, "Nhà Sách Tiki", "/images/products/72725402.png", false),
    createProduct("sp-ns5", "shop-001", "nha-sach-tiki", "Đắc Nhân Tâm", 89000, 120000, 25, 2100, 4.8, 145, 56, "Nhà Sách Tiki", "/images/products/72727057.jpg", false),
  ],
  "dien-thoai-may-tinh-bang": [
    createProduct("sp-101", "shop-001", "dien-thoai-may-tinh-bang", "iPhone 15 Pro Max 256GB", 28900000, 32000000, 9, 542, 4.9, 128, 45, "Tiki Trading", "/images/products/72723335.jpg", true),
    createProduct("sp-102", "shop-002", "dien-thoai-may-tinh-bang", "Samsung Galaxy S24 Ultra", 25900000, 29900000, 13, 321, 4.7, 98, 32, "TechStore", "/images/products/72725402.png", true),
    createProduct("sp-103", "shop-002", "dien-thoai-may-tinh-bang", "Xiaomi Redmi Note 13", 4990000, 5990000, 16, 890, 4.5, 156, 67, "TechStore", "/images/products/72727057.jpg", true),
    createProduct("sp-104", "shop-002", "dien-thoai-may-tinh-bang", "Oppo Reno 11 Pro", 9990000, 12990000, 23, 156, 4.6, 67, 28, "TechStore", "/images/products/72746857.jpg", false),
    createProduct("sp-105", "shop-001", "dien-thoai-may-tinh-bang", "iPad Air 5 Gen", 12990000, 15990000, 18, 342, 4.7, 89, 34, "Tiki Trading", "/images/products/72791018.jpg", true),
  ],
  "laptop-may-vi-tinh-linh-kien": [
    createProduct("sp-201", "shop-001", "laptop-may-vi-tinh-linh-kien", "MacBook Air M2 2024", 28900000, 32000000, 9, 234, 4.8, 89, 34, "Tiki Trading", "/images/products/72746857.jpg", true),
    createProduct("sp-202", "shop-002", "laptop-may-vi-tinh-linh-kien", "Dell Inspiron 15 5000", 18900000, 21000000, 10, 156, 4.6, 78, 23, "TechStore", "/images/products/72791018.jpg", false),
    createProduct("sp-203", "shop-002", "laptop-may-vi-tinh-linh-kien", "ASUS VivoBook 14", 15900000, 18500000, 14, 342, 4.5, 67, 28, "TechStore", "/images/products/72800659.jpg", false),
    createProduct("sp-204", "shop-002", "laptop-may-vi-tinh-linh-kien", "Lenovo ThinkPad X1", 29900000, 34900000, 14, 89, 4.7, 45, 18, "TechStore", "/images/products/7280161.jpg", false),
  ],
  "thoi-trang-nu": [
    createProduct("sp-701", "shop-005", "thoi-trang-nu", "Váy Liên Buộc Vạt Mảnh", 399000, 599000, 33, 345, 4.6, 67, 28, "FashionHub", "/images/products/72870951.png", false),
    createProduct("sp-702", "shop-005", "thoi-trang-nu", "Áo Khoác Nữ windbreaker", 599000, 899000, 33, 189, 4.5, 56, 21, "FashionHub", "/images/products/72870957.jpg", false),
    createProduct("sp-703", "shop-005", "thoi-trang-nu", "Áo Sơ Mi Nữ Tay Dài", 299000, 450000, 33, 423, 4.4, 78, 31, "FashionHub", "/images/products/72875812.jpg", false),
    createProduct("sp-704", "shop-005", "thoi-trang-nu", "Chân Váy Bút Mòn", 349000, 499000, 30, 267, 4.5, 65, 26, "FashionHub", "/images/products/72880242.jpg", false),
    createProduct("sp-705", "shop-005", "thoi-trang-nu", "Set Đồ Nữ Thể Thao", 549000, 849000, 35, 189, 4.6, 56, 22, "FashionHub", "/images/products/72881856.jpg", false),
  ],
  "thoi-trang-nam": [
    createProduct("sp-601", "shop-005", "thoi-trang-nam", "Áo Thun Nam Cotton Basic", 199000, 299000, 33, 456, 4.5, 78, 32, "FashionHub", "/images/products/72870951.png", false),
    createProduct("sp-602", "shop-005", "thoi-trang-nam", "Quần Jeans Nam Slim Fit", 499000, 799000, 37, 234, 4.4, 65, 24, "FashionHub", "/images/products/72870957.jpg", false),
    createProduct("sp-603", "shop-005", "thoi-trang-nam", "Áo Sơ Mi Trắng Tay Ngắn", 249000, 399000, 37, 345, 4.5, 78, 29, "FashionHub", "/images/products/72875812.jpg", false),
    createProduct("sp-604", "shop-005", "thoi-trang-nam", "Áo Khoác Nam Gió", 599000, 899000, 33, 189, 4.4, 56, 23, "FashionHub", "/images/products/72880242.jpg", false),
    createProduct("sp-605", "shop-005", "thoi-trang-nam", "Quần Tây Nam Slim", 449000, 699000, 35, 267, 4.6, 67, 28, "FashionHub", "/images/products/72881856.jpg", false),
  ],
  "giay-dep-nu": [
    createProduct("sp-901", "shop-006", "giay-dep-nu", "Giày Sandal Đụp Nữ", 399000, 599000, 33, 423, 4.4, 89, 34, "ShoeStore", "/images/products/72890268.jpg", false),
    createProduct("sp-902", "shop-006", "giay-dep-nu", "Boot Nữ Cao ấn tượng", 799000, 1199000, 33, 267, 4.5, 67, 26, "ShoeStore", "/images/products/72895953.jpg", false),
    createProduct("sp-903", "shop-006", "giay-dep-nu", "Giày Thể Thao Nữ", 599000, 899000, 33, 345, 4.6, 78, 31, "ShoeStore", "/images/products/72899999.jpg", false),
    createProduct("sp-904", "shop-006", "giay-dep-nu", "Dép Lê Nữ Đế Phẳng", 199000, 299000, 33, 234, 4.4, 56, 22, "ShoeStore", "/images/products/72901279.jpg", false),
  ],
  "giay-dep-nam": [
    createProduct("sp-801", "shop-006", "giay-dep-nam", "Nike Air Force 1 Low", 2490000, 3200000, 22, 234, 4.7, 78, 31, "ShoeStore", "/images/products/72881856.jpg", true),
    createProduct("sp-802", "shop-006", "giay-dep-nam", "Adidas Ultraboost 22", 2990000, 3500000, 17, 156, 4.6, 67, 23, "ShoeStore", "/images/products/72885997.jpg", true),
    createProduct("sp-803", "shop-006", "giay-dep-nam", "Giày Sneaker Nam Classic", 1299000, 1799000, 27, 345, 4.5, 78, 34, "ShoeStore", "/images/products/72890268.jpg", false),
    createProduct("sp-804", "shop-006", "giay-dep-nam", "Dép Nam Đế Dày", 299000, 399000, 25, 234, 4.4, 56, 21, "ShoeStore", "/images/products/72895953.jpg", false),
  ],
  "bach-hoa-online": [
    createProduct("sp-1101", "shop-001", "bach-hoa-online", "Trà Trí Nhân Tương Sữa Ong", 89000, 120000, 25, 890, 4.7, 98, 45, "TikiMart", "/images/products/72901510.jpeg", false),
    createProduct("sp-1102", "shop-001", "bach-hoa-online", "Gạo ST Basmati Hữu Nghị", 129000, 169000, 23, 567, 4.6, 89, 34, "TikiMart", "/images/products/72902732.jpg", false),
    createProduct("sp-1103", "shop-001", "bach-hoa-online", "Nước Ép Gừng Nguyên Chất", 99000, 149000, 33, 423, 4.5, 67, 26, "TikiMart", "/images/products/72850517.jpeg", false),
    createProduct("sp-1104", "shop-001", "bach-hoa-online", "Bột Ngũ Cốc Ăn Kiêng", 149000, 199000, 25, 267, 4.4, 56, 22, "TikiMart", "/images/products/72850637.jpeg", false),
  ],
  "the-thao-da-ngoai": [
    createProduct("sp-1201", "shop-009", "the-thao-da-ngoai", "Bóng Đá Cao Su TForce", 199000, 299000, 33, 345, 4.5, 67, 23, "SportShop", "/images/products/72957200.jpg", false),
    createProduct("sp-1202", "shop-009", "the-thao-da-ngoai", "Dép Chống Trượt Thể Thao", 149000, 199000, 25, 234, 4.4, 56, 22, "SportShop", "/images/products/72968489.png", false),
    createProduct("sp-1203", "shop-009", "the-thao-da-ngoai", "Bơi Lôi Đai Bụng", 179000, 249000, 28, 156, 4.6, 45, 18, "SportShop", "/images/products/72825272.jpg", false),
    createProduct("sp-1204", "shop-009", "the-thao-da-ngoai", "Gậy Tập Chống Đẩy", 399000, 599000, 33, 89, 4.5, 34, 12, "SportShop", "/images/products/72850088.jpg", false),
  ],
  "dong-ho-va-trang-suc": [
    createProduct("sp-1001", "shop-007", "dong-ho-va-trang-suc", "Đồng Hồ Casio Edifice", 1890000, 2490000, 24, 178, 4.6, 56, 21, "WatchStore", "/images/products/72899999.jpg", true),
    createProduct("sp-1002", "shop-008", "dong-ho-va-trang-suc", "Vòng Tay Nam Thép Không Gỉ", 299000, 399000, 25, 345, 4.5, 67, 26, "JewelStore", "/images/products/72901279.jpg", false),
  ],
  "dien-tu-dien-lanh": [],
  "thiet-bi-kts-phu-kien-so": [],
};