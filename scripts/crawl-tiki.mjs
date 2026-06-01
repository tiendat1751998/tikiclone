import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';
import { execSync } from 'child_process';

const BASE_URL = 'https://tiki.vn';
const DEFAULT_SELLER_ID = 'usr-002';

const CATEGORIES = [
  { id: '1789', slug: 'dien-thoai-may-tinh-bang', name: 'Điện Thoại - Máy Tính Bảng' },
  { id: '1846', slug: 'laptop-may-vi-tinh-linh-kien', name: 'Laptop - Máy Vi Tính - Linh kiện' },
  { id: '1815', slug: 'thiet-bi-kts-phu-kien-so', name: 'Phụ Kiện Số' },
  { id: '4221', slug: 'dien-tu-dien-lanh', name: 'Điện Tử - Điện Lạnh' },
  { id: '1882', slug: 'dien-gia-dung', name: 'Điện Gia Dụng' },
  { id: '1520', slug: 'lam-dep-suc-khoe', name: 'Làm Đẹp - Sức Khỏe' },
  { id: '1975', slug: 'the-thao-da-ngoai', name: 'Thể Thao - Dã Ngoại' },
  { id: '1883', slug: 'nha-cua-doi-song', name: 'Nhà Cửa - Đời Sống' },
  { id: '4384', slug: 'bach-hoa-online', name: 'Bách Hóa Online' },
  { id: '2549', slug: 'do-choi-me-be', name: 'Đồ Chơi - Mẹ & Bé' },
];

async function fetchCategoryProducts(browser, cat) {
  const ctx = await browser.newContext({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0 Safari/537.36' });
  const page = await ctx.newPage();
  
  const products = await new Promise((resolve) => {
    const allItems = [];
    let resolved = false;
    
    page.on('response', async (response) => {
      const url = response.url();
      if (resolved) return;
      if (url.includes('blocks/listings') && url.includes(`category=${cat.id}`)) {
        try {
          const data = await response.json();
          const items = data.data || [];
          for (const item of items) {
            const p = item.product || item;
            const name = p.name || '';
            const price = p.price || p.final_price || p.list_price || 0;
            const images = (p.images || []).filter(i => i.startsWith('http'));
            const thumb = p.thumbnail_url || '';
            const brand = p.brand?.name || p.brand || '';
            if (name) {
              allItems.push({
                name,
                price: typeof price === 'number' ? price : (parseInt(String(price).replace(/[^0-9]/g,'')) || 0),
                images: images.length > 0 ? images : (thumb ? [thumb] : []),
                description: (p.short_description || p.description || '').substring(0, 2000),
                brand,
                category_name: p.categories?.[0]?.name || cat.name,
                category_slug: cat.slug,
              });
            }
          }
          resolved = true;
          resolve(allItems);
        } catch(e) { resolved = true; resolve(allItems); }
      }
    });
    
    page.goto(`${BASE_URL}/${cat.slug}/c${cat.id}`, { waitUntil: 'domcontentloaded', timeout: 30000 });
    setTimeout(() => { if (!resolved) { resolved = true; resolve(allItems); } }, 10000);
  });
  
  await ctx.close();
  return products;
}

async function main() {
  console.log('=== Tiki.vn API Crawler ===\n');
  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
  const allProducts = [];
  
  try {
    for (const cat of CATEGORIES.slice(0, 5)) {
      console.log(`Fetching ${cat.name}...`);
      const items = await fetchCategoryProducts(browser, cat);
      console.log(`  Got ${items.length} products`);
      for (const p of items) {
        p.category_slug = cat.slug;
        allProducts.push(p);
      }
    }
    
    console.log(`\nTotal: ${allProducts.length}. Selecting first 50...`);
    const selected = allProducts.slice(0, 50);
    
    // Build MongoDB seed script
    const now = new Date().toISOString();
    const storeCatId = uuidv4();
    const catMap = {};
    
    let script = `use tiki_catalog;
db.products.deleteMany({});
db.categories.deleteMany({});
db.categories.insertOne({ category_id: '${storeCatId}', name: 'Tiki', slug: 'tiki', parent_id: '', level: 1, sort_order: 1, children: [] });\n`;
    
    const catSlugs = [...new Set(selected.map(p => p.category_slug))];
    let order = 1;
    for (const slug of catSlugs) {
      const cat = CATEGORIES.find(c => c.slug === slug) || { name: slug, slug };
      const catId = uuidv4();
      catMap[slug] = catId;
      const catName = JSON.stringify(cat.name);
      script += `db.categories.insertOne({ category_id: '${catId}', name: ${catName}, slug: '${slug}', parent_id: '${storeCatId}', level: 2, sort_order: ${order++}, children: [] });\n`;
    }
    
    for (const cat of CATEGORIES) {
      if (catMap[cat.slug]) continue;
      const catId = uuidv4();
      catMap[cat.slug] = catId;
      const catName = JSON.stringify(cat.name);
      script += `db.categories.insertOne({ category_id: '${catId}', name: ${catName}, slug: '${cat.slug}', parent_id: '${storeCatId}', level: 2, sort_order: ${order++}, children: [] });\n`;
    }
    
    script += '\n// Products\n';
    for (const p of selected) {
      const productId = 'spu-' + uuidv4().slice(0, 8);
      const skuId = 'sku-' + uuidv4().slice(0, 8);
      const catId = catMap[p.category_slug] || storeCatId;
      const images = JSON.stringify(p.images.filter(u => u.startsWith('http')));
      const attrs = JSON.stringify({ brand: p.brand || '' });
      const title = JSON.stringify(p.name);
      const desc = JSON.stringify(p.description);
      
      script += `db.products.insertOne({
  spu_id: '${productId}',
  title: ${title},
  description: ${desc},
  category_id: '${catId}',
  seller_id: '${DEFAULT_SELLER_ID}',
  status: 'ACTIVE',
  attributes: ${attrs},
  images: ${images},
  skus: [{
    sku_id: '${skuId}',
    spu_id: '${productId}',
    price: ${p.price},
    stock: 100,
    status: 'ACTIVE',
    variations: [{ name: 'Mặc định', value: 'default' }]
  }],
  created_at: ISODate('${now}'),
  updated_at: ISODate('${now}')
});\n`;
    }
    
    fs.writeFileSync('/tmp/seed-tiki.js', script);
    console.log(`Seed script saved to /tmp/seed-tiki.js (${selected.length} products)`);
    
    console.log('Seeding MongoDB...');
    const out = execSync('docker compose exec -T mongodb mongosh tiki_catalog < /tmp/seed-tiki.js', {
      cwd: '/home/datdt/tikiclone', encoding: 'utf8', timeout: 60000, shell: '/bin/bash',
    });
    console.log('MongoDB seeded successfully!');
    
  } finally {
    await browser.close();
  }
  console.log('\nDone!');
}

main().catch(err => { console.error(err); process.exit(1); });
