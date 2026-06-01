/**
 * Tiki Crawler Pass 3 - More pages from existing categories
 * + different sort orders to get unique products
 * Target: reach 20,000+ total
 */

import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';
import { execSync } from 'child_process';
import https from 'https';
import http from 'http';
import path from 'path';
import { URL } from 'url';

const BASE_URL = 'https://tiki.vn';
const IMAGE_DIR = '/home/datdt/tikiclone/public/images/products';
const PRODUCTS_FILE = '/tmp/crawled_products_v3.json';
const CATEGORIES_FILE = '/tmp/crawled_categories_v3.json';

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }
function extractPrice(v) {
  if (!v && v !== 0) return 0;
  if (typeof v === 'number') return v;
  return parseInt(String(v).replace(/[^0-9]/g, '')) || 0;
}
function sanitize(s) { return (s || '').replace(/'/g, "''").trim(); }
function cleanHtml(s) { return (s || '').replace(/<[^>]+>/g, '').trim(); }

function localImgDest(imageUrl, productId) {
  try {
    const ext = (path.extname(new URL(imageUrl).pathname).split('?')[0] || '.jpg').toLowerCase();
    return path.join(IMAGE_DIR, `${productId}${ext}`);
  } catch { return path.join(IMAGE_DIR, `${productId}.jpg`); }
}
function localImgUrl(imageUrl, productId) {
  return `/images/products/${path.basename(localImgDest(imageUrl, productId))}`;
}

async function downloadImage(imageUrl, destPath, retries = 2) {
  if (fs.existsSync(destPath)) return true;
  for (let i = 0; i <= retries; i++) {
    try {
      await new Promise((resolve, reject) => {
        const mod = imageUrl.startsWith('https') ? https : http;
        const req = mod.get(imageUrl, {
          headers: {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0 Safari/537.36',
            'Referer': 'https://tiki.vn/',
          },
          timeout: 15000,
        }, (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            downloadImage(res.headers.location, destPath, 0).then(resolve).catch(reject);
            return;
          }
          if (res.statusCode !== 200) { reject(new Error(`HTTP ${res.statusCode}`)); return; }
          const file = fs.createWriteStream(destPath);
          res.pipe(file);
          file.on('finish', () => { file.close(); resolve(); });
        });
        req.on('error', reject);
        req.on('timeout', () => { req.destroy(); reject(new Error('timeout')); });
      });
      return true;
    } catch (e) {
      if (i === retries) return false;
      await sleep(300);
    }
  }
  return false;
}

// Categories that worked well - crawl more pages
const CATEGORIES_EXTENDED = [
  // Re-crawl top categories with more pages and different sort orders
  { id: '1789', slug: 'dien-thoai-may-tinh-bang', name: 'Điện Thoại - Máy Tính Bảng' },
  { id: '1846', slug: 'laptop-may-vi-tinh-linh-kien', name: 'Laptop - Máy Vi Tính - Linh Kiện' },
  { id: '1815', slug: 'thiet-bi-kts-phu-kien-so', name: 'Phụ Kiện Số' },
  { id: '4221', slug: 'dien-tu-dien-lanh', name: 'Điện Tử - Điện Lạnh' },
  { id: '1882', slug: 'dien-gia-dung', name: 'Điện Gia Dụng' },
  { id: '915',  slug: 'thoi-trang-nu', name: 'Thời Trang Nữ' },
  { id: '931',  slug: 'thoi-trang-nam', name: 'Thời Trang Nam' },
  { id: '1686', slug: 'giay-dep-nam', name: 'Giày - Dép Nam' },
  { id: '1703', slug: 'giay-dep-nu', name: 'Giày - Dép Nữ' },
  { id: '27498',slug: 'phu-kien-thoi-trang', name: 'Phụ Kiện Thời Trang' },
  { id: '6000', slug: 'balo-va-vali', name: 'Balo và Vali' },
  { id: '1520', slug: 'lam-dep-suc-khoe', name: 'Làm Đẹp - Sức Khỏe' },
  { id: '1883', slug: 'nha-cua-doi-song', name: 'Nhà Cửa - Đời Sống' },
  { id: '2549', slug: 'do-choi-me-be', name: 'Đồ Chơi - Mẹ & Bé' },
  { id: '1975', slug: 'the-thao-da-ngoai', name: 'Thể Thao - Dã Ngoại' },
  { id: '8322', slug: 'nha-sach-tiki', name: 'Nhà Sách Tiki' },
  { id: '4384', slug: 'bach-hoa-online', name: 'Bách Hóa Online' },
  { id: '8594', slug: 'o-to-xe-may-xe-dap', name: 'Ô Tô - Xe Máy - Xe Đạp' },
  { id: '15078',slug: 'cham-soc-da-mat', name: 'Chăm Sóc Da Mặt' },
];

async function crawlCategoryMorePages(browser, cat, startPage = 26, maxPages = 75) {
  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
    viewport: { width: 1280, height: 900 },
  });
  const page = await ctx.newPage();
  const products = [];
  const seenInThisCrawl = new Set();

  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/api/personalish/v1/blocks/listings') && url.includes(`category=${cat.id}`)) {
      try {
        const data = await response.json();
        const items = data.data || [];
        for (const item of items) {
          const pid = String(item.id || item.sku || '');
          if (!pid || seenInThisCrawl.has(pid)) continue;
          seenInThisCrawl.add(pid);

          const images = [];
          if (item.thumbnail_url) images.push(item.thumbnail_url);
          if (item.images && Array.isArray(item.images)) {
            for (const img of item.images) {
              const u = typeof img === 'string' ? img : (img?.url || img?.thumbnail_url || '');
              if (u?.startsWith('http') && !images.includes(u)) images.push(u);
            }
          }

          const price = extractPrice(item.price || item.final_price);
          const discount = item.discount_rate || null;

          products.push({
            tiki_product_id: pid,
            name: (item.name || '').substring(0, 500),
            description: cleanHtml(item.short_description || '').substring(0, 2000),
            price,
            original_price: (discount && price) ? Math.round(price / (1 - discount / 100)) : null,
            discount_percent: discount,
            brand: (item.brand_name || '').substring(0, 255),
            images: images.filter(u => u.startsWith('http')).slice(0, 8),
            rating_average: item.rating_average || null,
            review_count: item.review_count || null,
            sold_count: (item.quantity_sold?.value) ?? item.order_count ?? null,
            seller_name: String(item.seller_product_id || ''),
            category_id: cat.id,
            category_name: cat.name,
            category_slug: cat.slug,
            url: item.url_path ? `${BASE_URL}${item.url_path}` : `${BASE_URL}/${cat.slug}/c${cat.id}`,
          });
        }
      } catch (e) { /* ignore */ }
    }
  });

  try {
    for (let pageNum = startPage; pageNum <= startPage + maxPages; pageNum++) {
      const url = `${BASE_URL}/${cat.slug}/c${cat.id}?page=${pageNum}`;
      const beforeCount = products.length;
      
      try {
        await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 20000 });
        await sleep(2500 + Math.random() * 1500);
        await page.evaluate(() => window.scrollBy(0, 800));
        await sleep(1000);
      } catch (e) { break; }

      if (products.length === beforeCount) {
        break; // No more products
      }

      await sleep(600 + Math.random() * 800);
    }
  } finally {
    await ctx.close();
  }

  return products;
}

async function main() {
  console.log('=== Tiki Crawler Pass 3: More Pages ===');
  
  let existingProducts = [];
  let existingCategories = [];
  try { existingProducts = JSON.parse(fs.readFileSync(PRODUCTS_FILE, 'utf8')); } catch (e) {}
  try { existingCategories = JSON.parse(fs.readFileSync(CATEGORIES_FILE, 'utf8')); } catch (e) {}
  
  const seenProductIds = new Set(existingProducts.map(p => p.tiki_product_id));
  const catDbIdMap = {};
  for (const c of existingCategories) catDbIdMap[c.tiki_id] = c.id;
  
  console.log(`Existing: ${existingProducts.length} products`);

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage'],
  });

  const newProducts = [];

  try {
    for (let i = 0; i < CATEGORIES_EXTENDED.length; i++) {
      const cat = CATEGORIES_EXTENDED[i];
      console.log(`[${i + 1}/${CATEGORIES_EXTENDED.length}] ${cat.name} - pages 26+`);
      
      const products = await crawlCategoryMorePages(browser, cat, 26, 50);
      
      let newCount = 0;
      for (const p of products) {
        if (!seenProductIds.has(p.tiki_product_id)) {
          seenProductIds.add(p.tiki_product_id);
          p.db_id = uuidv4();
          p.category_db_id = catDbIdMap[cat.id] || uuidv4();
          newProducts.push(p);
          existingProducts.push(p);
          newCount++;
        }
      }
      
      console.log(`  New: ${newCount}. Total: ${existingProducts.length}`);
      
      if (i % 5 === 0) {
        fs.writeFileSync(PRODUCTS_FILE, JSON.stringify(existingProducts));
      }
      
      if (existingProducts.length >= 22000) {
        console.log('Reached 22,000!');
        break;
      }
      
      await sleep(1000 + Math.random() * 1500);
    }
  } finally {
    await browser.close();
  }

  console.log(`\nNew products: ${newProducts.length}`);
  console.log(`Total: ${existingProducts.length}`);

  // Download images for new products
  if (newProducts.length > 0) {
    console.log('=== Downloading Images ===');
    let dlOk = 0, dlFail = 0;
    for (let i = 0; i < newProducts.length; i++) {
      const p = newProducts[i];
      if (!p.images || p.images.length === 0) continue;
      const dest = localImgDest(p.images[0], p.tiki_product_id);
      const ok = await downloadImage(p.images[0], dest);
      if (ok) {
        p.local_image_url = localImgUrl(p.images[0], p.tiki_product_id);
        p.image_url = p.local_image_url;
        dlOk++;
      } else {
        p.local_image_url = '';
        p.image_url = p.images[0];
        dlFail++;
      }
      if ((i + 1) % 500 === 0) console.log(`  ${i + 1}/${newProducts.length}`);
      if (i % 100 === 0) sleep(100);
    }
    console.log(`Images: ${dlOk} ok, ${dlFail} fail`);
  }

  // Save to MySQL and MongoDB
  console.log('=== Updating MySQL ===');
  await saveToMySQL(newProducts);
  
  console.log('=== Updating MongoDB ===');
  await saveToMongoDB(newProducts);

  fs.writeFileSync(PRODUCTS_FILE, JSON.stringify(existingProducts));
  fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(existingCategories));
  
  console.log(`\n========== PASS 3 COMPLETE ==========`);
  console.log(`Total products: ${existingProducts.length}`);
  console.log(`Total categories: ${existingCategories.length}`);
}

async function saveToMySQL(newProducts) {
  const batchSize = 100;
  let inserted = 0;
  for (let i = 0; i < newProducts.length; i += batchSize) {
    const batch = newProducts.slice(i, i + batchSize);
    const values = batch.map(p =>
      `('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average || 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`
    ).join(',');
    const sql = `INSERT IGNORE INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ${values}`;
    try {
      execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${sql.replace(/"/g, '\\"')}" 2>/dev/null`, { timeout: 60000, encoding: 'utf8' });
      inserted += batch.length;
    } catch (e) {
      for (const p of batch) {
        try {
          const s = `INSERT IGNORE INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average || 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`;
          execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${s.replace(/"/g, '\\"')}" 2>/dev/null`, { timeout: 10000, encoding: 'utf8' });
          inserted++;
        } catch (e2) { /* skip */ }
      }
    }
    if ((i + batchSize) % 1000 === 0 || i + batchSize >= newProducts.length) {
      console.log(`  ${inserted}/${newProducts.length}`);
    }
  }
  try {
    const cnt = execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -N -e "SELECT COUNT(*) FROM tiki_products" 2>/dev/null`, { timeout: 10000, encoding: 'utf8' });
    console.log(`  MySQL total: ${cnt.trim()}`);
  } catch (e) { /* */ }
}

async function saveToMongoDB(newProducts) {
  const now = new Date().toISOString();
  const chunkSize = 2000;
  for (let i = 0; i < newProducts.length; i += chunkSize) {
    const chunk = newProducts.slice(i, i + chunkSize);
    let script = '';
    for (const p of chunk) {
      const imgs = JSON.stringify(p.images || []);
      const attrs = JSON.stringify({ brand: p.brand || '' });
      const skuId = 'sku-' + uuidv4().slice(0, 8);
      script += `db.products.insertOne({spu_id:'${p.db_id}',title:${JSON.stringify(p.name)},description:${JSON.stringify(p.description || '')},category_id:'${p.category_db_id}',seller_id:'usr-002',status:'ACTIVE',attributes:${attrs},images:${imgs},local_image_url:${JSON.stringify(p.local_image_url || '')},skus:[{sku_id:'${skuId}',spu_id:'${p.db_id}',price:${p.price},stock:${Math.floor(Math.random()*200)+10},status:'ACTIVE',variations:[{name:'Mặc định',value:'default'}]}],rating_average:${p.rating_average || 'null'},review_count:${p.review_count || 0},sold_count:${p.sold_count || 0},tiki_product_id:'${p.tiki_product_id}',created_at:ISODate('${now}'),updated_at:ISODate('${now}')});\n`;
    }
    fs.writeFileSync(`/tmp/seed_mongo_p3_chunk_${Math.floor(i/chunkSize)}.js`, script);
    try {
      execSync(`docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog < /tmp/seed_mongo_p3_chunk_${Math.floor(i/chunkSize)}.js 2>/dev/null`, { timeout: 120000, encoding: 'utf8' });
      console.log(`  Chunk ${Math.floor(i/chunkSize)} OK`);
    } catch (e) { console.log(`  Chunk ${Math.floor(i/chunkSize)} error`); }
  }
}

main().catch(err => {
  console.error('FATAL:', err);
  process.exit(1);
});
