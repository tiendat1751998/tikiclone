// Seed script for MongoDB tiki_catalog database
const db = db.getSiblingDB('tiki_catalog');

// Clear existing data
db.products.drop();
db.categories.drop();

// ===== Categories =====
const categories = [
  {
    category_id: "1883", name: "Nhà Cửa - Đời Sống", slug: "nha-cua-doi-song",
    parent_id: "", level: 0, sort_order: 1, children: []
  },
  {
    category_id: "1951", name: "Dụng cụ nhà bếp", slug: "dung-cu-nha-bep",
    parent_id: "1883", level: 0, sort_order: 2, children: []
  },
  {
    category_id: "1973", name: "Trang trí nhà cửa", slug: "trang-tri-nha-cua",
    parent_id: "1883", level: 0, sort_order: 3, children: []
  },
  {
    category_id: "2150", name: "Nội thất", slug: "noi-that",
    parent_id: "1883", level: 0, sort_order: 4, children: []
  },
  {
    category_id: "2015", name: "Đèn & thiết bị chiếu sáng", slug: "den-thiet-bi-chieu-sang",
    parent_id: "1883", level: 0, sort_order: 5, children: []
  },
  {
    category_id: "1966", name: "Đồ dùng và thiết bị nhà tắm", slug: "do-dung-va-thiet-bi-nha-tam",
    parent_id: "1883", level: 0, sort_order: 6, children: []
  },
  {
    category_id: "dien-thoai", name: "Điện thoại", slug: "dien-thoai",
    parent_id: "", level: 0, sort_order: 7, children: []
  },
  {
    category_id: "laptop", name: "Laptop", slug: "laptop",
    parent_id: "", level: 0, sort_order: 8, children: []
  },
  {
    category_id: "thoi-trang", name: "Thời trang", slug: "thoi-trang",
    parent_id: "", level: 0, sort_order: 9, children: []
  },
  {
    category_id: "do-choi-me-be", name: "Mẹ & Bé", slug: "me-va-be",
    parent_id: "", level: 0, sort_order: 10, children: []
  }
];
db.categories.insertMany(categories);
print(`Inserted ${categories.length} categories`);

// ===== Products =====
const products = [
  {
    spu_id: "spu-dien-thoai-001", title: "iPhone 15 Pro Max 256GB", description: "Màn hình Super Retina XDR 6.7 inch, Chip A17 Pro, Camera 48MP",
    category_id: "dien-thoai", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Apple", color: "Titan Natural" },
    images: ["/images/products/spu-dien-thoai-001.jpg"],
    skus: [{ sku_id: "sku-iphone-15pm-256", price: 30990000, compare_price: 34990000, stock: 50, status: "ACTIVE", variations: [{ name: "Màu", value: "Titan Natural" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 1250
  },
  {
    spu_id: "spu-dien-thoai-002", title: "Samsung Galaxy S24 Ultra", description: "Màn hình Dynamic AMOLED 2x 6.8 inch, Chip Snapdragon 8 Gen 3, Camera 200MP",
    category_id: "dien-thoai", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Samsung", color: "Titanium Gray" },
    images: ["/images/products/spu-dien-thoai-002.jpg"],
    skus: [{ sku_id: "sku-s24u-256", price: 24990000, compare_price: 28990000, stock: 35, status: "ACTIVE", variations: [{ name: "Màu", value: "Titanium Gray" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 890
  },
  {
    spu_id: "spu-dien-thoai-003", title: "Xiaomi 14 Pro", description: "Màn hình LTPO AMOLED 6.73 inch, Chip Snapdragon 8 Gen 3, Camera Leica 50MP",
    category_id: "dien-thoai", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Xiaomi", color: "Black" },
    images: ["/images/products/spu-dien-thoai-003.jpg"],
    skus: [{ sku_id: "sku-xiaomi14p-256", price: 12990000, compare_price: 15990000, stock: 100, status: "ACTIVE", variations: [{ name: "Màu", value: "Black" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 2100
  },
  {
    spu_id: "spu-laptop-001", title: "MacBook Pro 14 inch M3 Pro", description: "Chip M3 Pro 18GB RAM, 512GB SSD, Màn hình Liquid Retina XDR",
    category_id: "laptop", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Apple", color: "Space Gray" },
    images: ["/images/products/spu-laptop-001.jpg"],
    skus: [{ sku_id: "sku-mbp14-m3pro", price: 39990000, compare_price: 45990000, stock: 20, status: "ACTIVE", variations: [{ name: "Màu", value: "Space Gray" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 567
  },
  {
    spu_id: "spu-laptop-002", title: "Dell XPS 15", description: "Intel Core i7-13700H, 16GB RAM, 512GB SSD, Màn hình OLED 3.5K",
    category_id: "laptop", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Dell", color: "Platinum Silver" },
    images: ["/images/products/spu-laptop-002.jpg"],
    skus: [{ sku_id: "sku-xps15-i7", price: 32990000, stock: 15, status: "ACTIVE", variations: [{ name: "Màu", value: "Platinum Silver" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 345
  },
  {
    spu_id: "spu-laptop-003", title: "Laptop Lenovo ThinkPad X1 Carbon Gen 11", description: "Intel Core i7-1365U, 16GB RAM, 512GB SSD, Màn hình 14 inch WUXGA",
    category_id: "laptop", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Lenovo", color: "Black" },
    images: ["/images/products/spu-laptop-003.jpg"],
    skus: [{ sku_id: "sku-x1c-i7", price: 39990000, stock: 10, status: "ACTIVE", variations: [{ name: "Màu", value: "Black" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 234
  },
  {
    spu_id: "spu-nha-bep-001", title: "Bộ nồi chống dính cao cấp 5 món", description: "Chất liệu nhôm cao cấp phủ ceramic, chống dính hoàn hảo, an toàn sức khỏe",
    category_id: "1951", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "LocknLock", material: "Ceramic" },
    images: ["/images/products/spu-nha-bep-001.jpg"],
    skus: [{ sku_id: "sku-noid-5mon", price: 699000, compare_price: 899000, stock: 200, status: "ACTIVE", variations: [{ name: "Màu", value: "Xanh" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 3450
  },
  {
    spu_id: "spu-nha-bep-002", title: "Máy xay sinh tố đa năng Philips", description: "Công suất 900W, 2 cối xay, lưỡi dao thép không gỉ",
    category_id: "1951", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Philips" },
    images: ["/images/products/spu-nha-bep-002.jpg"],
    skus: [{ sku_id: "sku-mayxay-philips", price: 999000, compare_price: 1299000, stock: 80, status: "ACTIVE", variations: [{ name: "Màu", value: "Trắng" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 5678
  },
  {
    spu_id: "spu-noi-that-001", title: "Bàn làm việc thông minh chỉnh chiều cao", description: "Kích thước 120x60cm, chống xước, chịu lực tốt, điều chỉnh điện 3 cấp độ nhớ",
    category_id: "2150", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "TikiHome", material: "Gỗ công nghiệp" },
    images: ["/images/products/spu-noi-that-001.jpg"],
    skus: [{ sku_id: "sku-banlamviec", price: 5890000, stock: 25, status: "ACTIVE", variations: [{ name: "Màu", value: "Trắng" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 890
  },
  {
    spu_id: "spu-thoi-trang-001", title: "Áo thun nam cotton cao cấp", description: "Chất liệu 100% cotton, form regular fit, thoáng mát",
    category_id: "thoi-trang", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "TikiStyle", material: "Cotton" },
    images: ["/images/products/spu-thoi-trang-001.jpg"],
    skus: [
      { sku_id: "sku-aothun-s", price: 199000, stock: 500, status: "ACTIVE", variations: [{ name: "Size", value: "S" }], image: "" },
      { sku_id: "sku-aothun-m", price: 199000, stock: 500, status: "ACTIVE", variations: [{ name: "Size", value: "M" }], image: "" },
      { sku_id: "sku-aothun-l", price: 199000, stock: 500, status: "ACTIVE", variations: [{ name: "Size", value: "L" }], image: "" },
    ],
    created_at: new Date(), updated_at: new Date(), sold_count: 12345
  },
  {
    spu_id: "spu-do-choi-me-be-001", title: "Tã quần cao cấp size M 60 miếng", description: "Khô thoáng, chống tràn, an toàn cho da bé",
    category_id: "do-choi-me-be", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Merries" },
    images: ["/images/products/spu-do-choi-me-be-001.jpg"],
    skus: [{ sku_id: "sku-ta-merries-m", price: 249000, stock: 1000, status: "ACTIVE", variations: [{ name: "Size", value: "M" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 23456
  },
  {
    spu_id: "spu-trang-tri-001", title: "Kệ trang trí treo tường 3 tầng", description: "Chất liệu gỗ MDF, sơn tĩnh điện, chịu lực tốt, phù hợp trang trí nhà cửa",
    category_id: "1973", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "TikiHome", material: "Gỗ MDF" },
    images: ["/images/products/spu-trang-tri-001.png"],
    skus: [{ sku_id: "sku-ketrangtri", price: 459000, stock: 60, status: "ACTIVE", variations: [{ name: "Màu", value: "Trắng" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 4567
  },
  {
    spu_id: "spu-den-001", title: "Đèn bàn LED thông minh", description: "Điều chỉnh độ sáng, chống mỏi mắt, cổng sạc USB",
    category_id: "2015", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Rạng Đông" },
    images: ["/images/products/spu-den-001.jpg"],
    skus: [{ sku_id: "sku-denled", price: 329000, stock: 150, status: "ACTIVE", variations: [{ name: "Màu", value: "Trắng" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 7890
  },
  {
    spu_id: "spu-nha-tam-001", title: "Sen vòi nóng lạnh cao cấp", description: "Chất liệu đồng thau mạ crôm, chống rỉ sét, van gốm cao cấp",
    category_id: "1966", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "Inax" },
    images: ["/images/products/spu-nha-tam-001.jpg"],
    skus: [{ sku_id: "sku-senvoi", price: 2890000, stock: 40, status: "ACTIVE", variations: [{ name: "Màu", value: "Crôm" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 1234
  },
  {
    spu_id: "spu-dien-thoai-004", title: "OPPO Find N3 Fold", description: "Màn hình gập 7.8 inch, Chip Snapdragon 8 Gen 2, Camera 48MP",
    category_id: "dien-thoai", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "OPPO", color: "Gold" },
    images: ["/images/products/spu-dien-thoai-004.jpg"],
    skus: [{ sku_id: "sku-oppo-n3", price: 39990000, stock: 10, status: "ACTIVE", variations: [{ name: "Màu", value: "Gold" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 456
  },
  {
    spu_id: "spu-thoi-trang-002", title: "Túi xách nữ cao cấp", description: "Chất liệu da bò thật, khóa kéo cao cấp, phong cách thanh lịch",
    category_id: "thoi-trang", seller_id: "usr-002", status: "ACTIVE",
    attributes: { brand: "TikiStyle", material: "Da bò" },
    images: ["/images/products/spu-thoi-trang-002.png"],
    skus: [{ sku_id: "sku-tuixach", price: 1599000, stock: 80, status: "ACTIVE", variations: [{ name: "Màu", value: "Đen" }], image: "" }],
    created_at: new Date(), updated_at: new Date(), sold_count: 3456
  },
];

db.products.insertMany(products);
print(`Inserted ${products.length} products`);

// Indexes
db.products.createIndex({ spu_id: 1 }, { unique: true });
db.products.createIndex({ category_id: 1 });
db.products.createIndex({ status: 1 });
db.products.createIndex({ title: "text" });
db.products.createIndex({ price: 1 });
db.products.createIndex({ created_at: -1 });

db.categories.createIndex({ category_id: 1 }, { unique: true });
db.categories.createIndex({ slug: 1 });

print("Indexes created successfully");
