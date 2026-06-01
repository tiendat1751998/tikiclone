/**
 * Robust Tiki.vn Crawler for 50k products
 * Uses direct API calls + parallel processing
 */
import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';

const BASE_URL = 'https://tiki.vn';
const IMAGE_DIR = '/home/datdt/tikiclone/public/images/products';
const TARGET = 50000;
const CONCURRENCY = 3; // Parallel categories

const CATEGORIES = [
  { id: '1789', slug: 'dien-thoai-may-tinh-bang', name: 'Điện Thoại - Máy Tính Bảng' },
  { id: '1846', slug: 'laptop-may-vi-tinh-linh-kien', name: 'Laptop - Máy Vi Tính - Linh Kiện' },
  { id: '915', slug: 'thoi-trang-nu', name: 'Thời Trang Nữ' },
  { id: '931', slug: 'thoi-trang-nam', name: 'Thời Trang Nam' },
  { id: '4384', slug: 'bach-hoa-online', name: 'Bách Hóa Online' },
  { id: '1520', slug: 'lam-dep-suc-khoe', name: 'Làm Đẹp - Sức Khỏe' },
];

let allProducts = [];
let seenIds = new Set();
const startTime = Date.now();

function extractPrice(v) {
  if (!v && v !== 0) return 0;
  if (typeof v === 'number') return v;
  return parseInt(String(v).replace(/[^0-9]/g, '')) || 0;
}

async function crawlCategory(cat, browser) {
  const products = [];
  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 Chrome/120.0.0.0',
    viewport: { width: 1280, height: 900 },
  });
  const page = await ctx.newPage();
  
  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/api/personalish/v1/blocks/listings') && url.includes(`category=${cat.id}`)) {
      try {
        const data = await response.json();
        const items = data.data || [];
        for (const item of items) {
          const pid = String(item.id);
          if (!pid || seenIds.has(pid)) continue;
          seenIds.add(pid);
          
          products.push({
            tiki_product_id: pid,
            name: (item.name || '').substring(0, 500),
            description: (item.short_description || '').replace(/<[^>]+>/g, '').substring(0, 2000),
            price: extractPrice(item.price || item.final_price),
            original_price: extractPrice(item.original_price || item.list_price),
            brand: item.brand_name || '',
            images: [item.thumbnail_url, ...(item.images || []).map(i => typeof i === 'string' ? i : i.url)].filter(Boolean),
            rating_average: item.rating_average || null,
            review_count: item.review_count || null,
            sold_count: item.quantity_sold?.value || null,
            seller_name: String(item.seller_product_id || ''),
            category_id: cat.id,
            category_name: cat.name,
            category_slug: cat.slug,
          });
        }
      } catch (e) {}
    }
  });

  try {
    for (let pageNum = 1; pageNum <= 200 && products.length < 10000; pageNum++) {
      try {
        await page.goto(`${BASE_URL}/${cat.slug}/c${cat.id}?page=${pageNum}`, {
          waitUntil: 'domcontentloaded',
          timeout: 15000,
        });
        await page.waitForTimeout(2000);
        await page.evaluate(() => window.scrollBy(0, 500));
        await page.waitForTimeout(1000);
      } catch (e) {
        if (e.message.includes('crash') || e.message.includes('closed')) break;
      }
    }
  } finally {
    await ctx.close();
  }
  
  return products;
}

async function main() {
  console.log(`=== Crawler for ${TARGET} products ===`);
  
  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-dev-shm-usage'],
  });
  
  try {
    const catRecords = CATEGORIES.map((c, i) => ({
      id: uuidv4(),
      tiki_id: c.id,
      name: c.name,
      slug: c.slug,
      parent_id: '',
      level: 1,
      sort_order: i + 1,
      product_count: 0,
      is_active: true,
    }));
    
    const catDbMap = {};
    for (const c of catRecords) catDbMap[c.tiki_id] = c.id;
    
    // Process categories in batches
    for (let i = 0; i < CATEGORIES.length; i += CONCURRENCY) {
      const batch = CATEGORIES.slice(i, i + CONCURRENCY);
      const promises = batch.map(c => crawlCategory(c, browser));
      const results = await Promise.all(promises);
      
      for (let j = 0; j < batch.length; j++) {
        const cat = batch[j];
        const prods = results[j];
        for (const p of prods) {
          p.db_id = uuidv4();
          p.category_db_id = catDbMap[cat.id];
          allProducts.push(p);
        }
        const catRec = catRecords.find(c => c.tiki_id === cat.id);
        if (catRec) catRec.product_count = prods.length;
        console.log(`  ${cat.name}: ${prods.length} products`);
      }
      
      // Save progress
      if (allProducts.length >= 10000 || i + CONCURRENCY >= CATEGORIES.length) {
        fs.writeFileSync('/tmp/crawl_products.json', JSON.stringify(allProducts));
      }
    }
    
    console.log(`\nTotal crawled: ${allProducts.length} products`);
    fs.writeFileSync('/tmp/crawl_products.json', JSON.stringify(allProducts));
    fs.writeFileSync('/tmp/crawl_categories.json', JSON.stringify(catRecords));
    
  } finally {
    await browser.close();
  }
}

main().catch(e => { console.error(e); process.exit(1); });