/**
 * Fetch 10 products from each Tiki.vn category using their public API
 */

// Tiki category IDs for API
const CATEGORIES = [
  { id: '1789', slug: 'dien-thoai-may-tinh-bang', name: 'Điện Thoại - Máy Tính Bảng' },
  { id: '1846', slug: 'laptop-may-vi-tinh-linh-kien', name: 'Laptop - Máy Vi Tính - Linh kiện' },
  { id: '915', slug: 'thiet-bi-kts-phu-kien-so', name: 'Thiết Bị Số - Phụ Kiện Số' },
  { id: '918', slug: 'dien-tu-dien-lanh', name: 'Điện Tử - Điện Lạnh' },
  { id: '1520', slug: 'lam-dep-suc-khoe', name: 'Làm Đẹp - Sức Khỏe' },
  { id: '926', slug: 'thoi-trang-nam', name: 'Thời trang nam' },
  { id: '931', slug: 'thoi-trang-nu', name: 'Thời trang nữ' },
  { id: '17079', slug: 'giay-dep-nam', name: 'Giày - Dép nam' },
  { id: '17078', slug: 'giay-dep-nu', name: 'Giày - Dép nữ' },
  { id: '981', slug: 'dong-ho-va-trang-suc', name: 'Đồng hồ và Trang sức' },
  { id: '4384', slug: 'bach-hoa-online', name: 'Bách Hóa Online' },
  { id: '1810', slug: 'the-thao-da-ngoai', name: 'Thể Thao - Dã Ngoại' },
  { id: '1728', slug: 'nha-cua-doi-song', name: 'Nhà Cửa - Đời Sống' },
  { id: '2169', slug: 'do-choi-me-be', name: 'Đồ Chơi - Mẹ & Bé' },
  { id: '1966', slug: 'oto-xe-may-xe-dap', name: 'Ô Tô - Xe Máy - Xe Đạp' },
  { id: '2526', slug: 'may-anh-may-quay-phim', name: 'Máy Ảnh - Máy Quay Phim' },
  { id: '17080', slug: 'phu-kien-thoi-trang', name: 'Phụ kiện thời trang' },
  { id: '975', slug: 'ngon', name: 'NGON' },
  { id: '5684', slug: 'balo-va-vali', name: 'Balo và Vali' },
];

async function fetchProducts(cat, page = 1) {
  const url = `https://tiki.vn/api/personalish/v1/blocks/listings?limit=10&category=${cat.id}&page=${page}`;
  
  try {
    const res = await fetch(url, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
      },
    });
    const data = await res.json();
    return data.data || [];
  } catch (e) {
    console.error(`Error fetching ${cat.name}:`, e.message);
    return [];
  }
}

function extractPrice(v) {
  if (!v && v !== 0) return 0;
  if (typeof v === 'number') return v;
  return parseInt(String(v).replace(/[^0-9]/g, '')) || 0;
}

function transformProduct(item, cat) {
  return {
    tiki_product_id: String(item.id),
    name: (item.name || '').substring(0, 500),
    description: (item.short_description || '').replace(/<[^>]+>/g, '').substring(0, 2000),
    price: extractPrice(item.final_price || item.price),
    original_price: extractPrice(item.list_price || item.original_price || item.price),
    brand: item.brand_name || '',
    image_url: item.thumbnail_url || item.image_url || '',
    rating_average: item.rating_average || 0,
    review_count: item.review_count || 0,
    sold_count: item.quantity_sold?.value || 0,
    seller_name: item.seller_name || '',
    category_slug: cat.slug,
    category_name: cat.name,
  };
}

async function main() {
  console.log('Fetching 10 products from each Tiki category...\n');
  
  const results = {};
  
  for (const cat of CATEGORIES) {
    const items = await fetchProducts(cat);
    const products = items.slice(0, 10).map((item) => transformProduct(item, cat));
    results[cat.slug] = products;
    console.log(`${cat.name}: ${products.length} products`);
  }
  
  // Save to file
  const fs = require('fs');
  fs.writeFileSync('/tmp/tiki-products-by-category.json', JSON.stringify(results, null, 2));
  console.log('\nSaved to /tmp/tiki-products-by-category.json');
}

main().catch(e => { console.error(e); process.exit(1); });