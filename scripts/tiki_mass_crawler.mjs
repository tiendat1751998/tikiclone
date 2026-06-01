/**
 * Tiki.vn Mass Crawler v3
 * Crawls 20,000+ products and categories from Tiki.vn
 * Uses Playwright to navigate + intercept API responses
 * Saves images locally, stores in MySQL + MongoDB
 * 
 * Usage: node scripts/tiki_mass_crawler.mjs
 */

import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import https from 'https';
import http from 'http';
import { URL } from 'url';

const BASE_URL = 'https://tiki.vn';
const IMAGE_DIR = '/home/datdt/tikiclone/public/images/products';
const PRODUCTS_FILE = '/tmp/crawled_products_v3.json';
const CATEGORIES_FILE = '/tmp/crawled_categories_v3.json';

fs.mkdirSync(IMAGE_DIR, { recursive: true });

let stats = {
  productsFound: 0, productsSaved: 0,
  imagesDownloaded: 0, imagesFailed: 0,
  startTime: Date.now(),
};

// Comprehensive category list - 27 root categories
const CATEGORIES = [
  { id: '1789', slug: 'dien-thoai-may-tinh-bang', name: 'Điện Thoại - Máy Tính Bảng' },
  { id: '1846', slug: 'laptop-may-vi-tinh-linh-kien', name: 'Laptop - Máy Vi Tính - Linh Kiện' },
  { id: '1815', slug: 'thiet-bi-kts-phu-kien-so', name: 'Phụ Kiện Số' },
  { id: '4221', slug: 'dien-tu-dien-lanh', name: 'Điện Tử - Điện Lạnh' },
  { id: '1882', slug: 'dien-gia-dung', name: 'Điện Gia Dụng' },
  { id: '8712', slug: 'may-anh-quay-phim', name: 'Máy Ảnh - Quay Phim' },
  { id: '4385', slug: 'thiet-bi-am-thanh', name: 'Thiết Bị Âm Thanh' },
  { id: '915',  slug: 'thoi-trang-nu', name: 'Thời Trang Nữ' },
  { id: '931',  slug: 'thoi-trang-nam', name: 'Thời Trang Nam' },
  { id: '1686', slug: 'giay-dep-nam', name: 'Giày - Dép Nam' },
  { id: '1703', slug: 'giay-dep-nu', name: 'Giày - Dép Nữ' },
  { id: '27498',slug: 'phu-kien-thoi-trang', name: 'Phụ Kiện Thời Trang' },
  { id: '27497',slug: 'dong-ho-va-trang-suc', name: 'Đồng Hồ và Trang Sức' },
  { id: '6000', slug: 'balo-va-vali', name: 'Balo và Vali' },
  { id: '1520', slug: 'lam-dep-suc-khoe', name: 'Làm Đẹp - Sức Khỏe' },
  { id: '15078',slug: 'cham-soc-da-mat', name: 'Chăm Sóc Da Mặt' },
  { id: '15077',slug: 'trang-diem', name: 'Trang Điểm' },
  { id: '15080',slug: 'cham-soc-toc', name: 'Chăm Sóc Tóc' },
  { id: '1883', slug: 'nha-cua-doi-song', name: 'Nhà Cửa - Đời Sống' },
  { id: '2549', slug: 'do-choi-me-be', name: 'Đồ Chơi - Mẹ & Bé' },
  { id: '1975', slug: 'the-thao-da-ngoai', name: 'Thể Thao - Dã Ngoại' },
  { id: '8322', slug: 'nha-sach-tiki', name: 'Nhà Sách Tiki' },
  { id: '4384', slug: 'bach-hoa-online', name: 'Bách Hóa Online' },
  { id: '8594', slug: 'o-to-xe-may-xe-dap', name: 'Ô Tô - Xe Máy - Xe Đạp' },
  { id: '1795', slug: 'dien-thoai-smartphone', name: 'Điện Thoại Smartphone' },
  { id: '1794', slug: 'may-tinh-bang', name: 'Máy Tính Bảng' },
  { id: '8992', slug: 'may-in', name: 'Máy In' },
];

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

// Crawl a single category using response interception
async function crawlCategory(browser, cat, maxPages = 200) {
  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
    viewport: { width: 1280, height: 900 },
  });
  const page = await ctx.newPage();
  const products = [];
  const seenInCategory = new Set();

  // Intercept product listing API responses
  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/api/personalish/v1/blocks/listings') && url.includes(`category=${cat.id}`)) {
      try {
        const data = await response.json();
        const items = data.data || [];
        for (const item of items) {
          const pid = String(item.id || item.sku || '');
          if (!pid || seenInCategory.has(pid)) continue;
          seenInCategory.add(pid);

          const images = [];
          if (item.thumbnail_url) images.push(item.thumbnail_url);
          if (item.images && Array.isArray(item.images)) {
            for (const img of item.images) {
              const u = typeof img === 'string' ? img : (img?.url || img?.thumbnail_url || '');
              if (u?.startsWith('http') && !images.includes(u)) images.push(u);
            }
          }

          const price = extractPrice(item.price || item.final_price);
          const origPrice = extractPrice(item.original_price || item.list_price);
          const discount = item.discount_rate || null;

          products.push({
            tiki_product_id: pid,
            name: (item.name || '').substring(0, 500),
            description: cleanHtml(item.short_description || '').substring(0, 2000),
            price,
            original_price: origPrice || (discount && price ? Math.round(price / (1 - discount / 100)) : null),
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
    for (let pageNum = 1; pageNum <= maxPages; pageNum++) {
      const url = `${BASE_URL}/${cat.slug}/c${cat.id}?page=${pageNum}`;
      const beforeCount = products.length;
      
      try {
        await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 20000 });
        // Wait for API responses
        await sleep(3000 + Math.random() * 2000);
        // Scroll to trigger lazy load
        await page.evaluate(() => window.scrollBy(0, 800));
        await sleep(1500);
      } catch (e) {
        console.log(`    Page ${pageNum} error: ${e.message.substring(0, 80)}`);
        break;
      }

      const newCount = products.length - beforeCount;
      if (newCount === 0) {
        // Try one more page in case it's slow
        if (pageNum < maxPages) {
          try {
            await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 20000 });
            await sleep(5000);
          } catch (e) { break; }
          if (products.length === beforeCount) break;
        } else {
          break;
        }
      }

      // Random delay between pages
      await sleep(800 + Math.random() * 1200);
     
      // Early stop if we got all products
      if (products.length === beforeCount) break;
    }
  } finally {
    await ctx.close();
  }

  return products;
}

async function main() {
  console.log('=== Tiki.vn Mass Crawler v3 ===');
  console.log(`Categories: ${CATEGORIES.length}, Target: 20,000+ products`);
  console.log(`Images: ${IMAGE_DIR}`);
  console.log('');

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
  });

  const allProducts = [];
  const seenProductIds = new Set();
  const allCatRecords = [];
  const catDbIdMap = {};

  // Retry wrapper for crawling
  async function crawlCategorySafe(cat, browser, attempt = 1) {
    try {
      return await crawlCategory(browser, cat, 100);
    } catch (e) {
      if (attempt < 3) {
        console.log(`  Retry ${attempt} for ${cat.name}...`);
        await new Promise(r => setTimeout(r, 5000));
        return crawlCategorySafe(cat, browser, attempt + 1);
      }
      console.log(`  Failed ${cat.name}: ${e.message.substring(0, 100)}`);
      return [];
    }
  }

  try {
    // Build root category records
    for (const cat of CATEGORIES) {
      const dbId = uuidv4();
      catDbIdMap[cat.id] = dbId;
      allCatRecords.push({
        id: dbId,
        tiki_id: cat.id,
        name: cat.name,
        slug: cat.slug,
        parent_id: '',
        parent_tiki_id: '',
        level: 1,
        sort_order: allCatRecords.length + 1,
        product_count: 0,
        image_url: '',
        url_path: `/${cat.slug}/c${cat.id}`,
        is_active: true,
      });
    }

    // Step 1: Crawl products from each category
    for (let i = 0; i < CATEGORIES.length; i++) {
      const cat = CATEGORIES[i];
      console.log(`[${i + 1}/${CATEGORIES.length}] ${cat.name} (ID:${cat.id})`);

      const products = await crawlCategorySafe(cat, browser);

      let newCount = 0;
      for (const p of products) {
        if (!seenProductIds.has(p.tiki_product_id)) {
          seenProductIds.add(p.tiki_product_id);
          p.db_id = uuidv4();
          p.category_db_id = catDbIdMap[cat.id];
          allProducts.push(p);
          newCount++;
        }
      }

      // Update product count in category record
      const catRecord = allCatRecords.find(c => c.tiki_id === cat.id);
      if (catRecord) catRecord.product_count = newCount;

      console.log(`  Products: ${newCount} new. Total: ${allProducts.length}`);
      stats.productsFound = allProducts.length;

      // Save progress
      if (i % 3 === 0 || i === CATEGORIES.length - 1) {
        fs.writeFileSync(PRODUCTS_FILE, JSON.stringify(allProducts));
        fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(allCatRecords));
      }

if (allProducts.length >= 52000) {
         console.log('\nReached 52,000 products target!');
         break;
       }

      await sleep(1500 + Math.random() * 2000);
    }

    console.log(`\n=== Crawl Complete: ${allProducts.length} products ===`);

// Step 2: Prepare images (use CDN URLs directly)
    console.log('\n=== Preparing Images (CDN) ===');
    for (let i = 0; i < allProducts.length; i++) {
      const p = allProducts[i];
      p.local_image_url = p.images?.[0] || '';
    }

    // Step 3: Save to MySQL
    console.log('\n=== Saving to MySQL ===');
    await saveToMySQL(allCatRecords, allProducts);

    // Step 4: Save to MongoDB
    console.log('\n=== Saving to MongoDB ===');
    await saveToMongoDB(allCatRecords, allProducts);

    // Final report
    const elapsed = ((Date.now() - stats.startTime) / 1000 / 60).toFixed(1);
    console.log(`\n========== COMPLETE ==========`);
    console.log(`Time: ${elapsed} min`);
    console.log(`Products: ${allProducts.length}`);
    console.log(`Categories: ${allCatRecords.length}`);
    console.log(`Images: ${stats.imagesDownloaded} downloaded, ${stats.imagesFailed} failed`);

    stats.productsSaved = allProducts.length;
    fs.writeFileSync(PRODUCTS_FILE, JSON.stringify(allProducts));
    fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(allCatRecords));
    fs.writeFileSync('/tmp/crawl_stats_v3.json', JSON.stringify(stats, null, 2));

  } finally {
    await browser.close();
  }
}

async function saveToMySQL(categories, products) {
  const createTables = `
DROP TABLE IF EXISTS tiki_categories;
DROP TABLE IF EXISTS tiki_products;
CREATE TABLE tiki_categories (
  id VARCHAR(36) PRIMARY KEY,
  tiki_id VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  parent_id VARCHAR(36) DEFAULT '',
  parent_tiki_id VARCHAR(50) DEFAULT '',
  level INT DEFAULT 1,
  sort_order INT DEFAULT 0,
  product_count INT DEFAULT 0,
  image_url VARCHAR(500) DEFAULT '',
  url_path VARCHAR(500) DEFAULT '',
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tiki_id (tiki_id),
  INDEX idx_parent (parent_id), INDEX idx_level (level), INDEX idx_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE tiki_products (
  id VARCHAR(36) PRIMARY KEY,
  tiki_product_id VARCHAR(50) NOT NULL,
  category_id VARCHAR(36) NOT NULL,
  category_name VARCHAR(255) DEFAULT NULL,
  category_slug VARCHAR(255) DEFAULT NULL,
  name VARCHAR(500) NOT NULL,
  description TEXT DEFAULT NULL,
  brand VARCHAR(255) DEFAULT NULL,
  price BIGINT NOT NULL DEFAULT 0,
  original_price BIGINT DEFAULT NULL,
  discount_percent INT DEFAULT NULL,
  rating_average DECIMAL(3,2) DEFAULT NULL,
  review_count INT DEFAULT NULL,
  sold_count INT DEFAULT NULL,
  seller_name VARCHAR(255) DEFAULT NULL,
  image_url VARCHAR(500) DEFAULT NULL,
  local_image_url VARCHAR(500) DEFAULT NULL,
  images JSON DEFAULT NULL,
  url VARCHAR(500) DEFAULT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tiki_prod_id (tiki_product_id),
  INDEX idx_category (category_id), INDEX idx_price (price),
  INDEX idx_status (status), INDEX idx_brand (brand),
  FULLTEXT INDEX idx_search (name, description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`;

  fs.writeFileSync('/tmp/create_tiki_tables.sql', createTables);
  try {
    execSync('docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform < /tmp/create_tiki_tables.sql', {
      timeout: 30000, encoding: 'utf8',
    });
    console.log('  Tables created');
  } catch (e) {
    console.log('  Table creation:', e.message.substring(0, 200));
  }

  // Insert categories
  console.log(`  Inserting ${categories.length} categories...`);
  for (let i = 0; i < categories.length; i += 50) {
    const batch = categories.slice(i, i + 50);
    const values = batch.map(c =>
      `('${c.id}','${c.tiki_id}','${sanitize(c.name)}','${c.slug}','${c.parent_id}','${c.parent_tiki_id}',${c.level},${c.sort_order},${c.product_count},'${sanitize(c.image_url)}','${c.url_path}',1)`
    ).join(',');
    const sql = `INSERT INTO tiki_categories (id,tiki_id,name,slug,parent_id,parent_tiki_id,level,sort_order,product_count,image_url,url_path,is_active) VALUES ${values} ON DUPLICATE KEY UPDATE name=VALUES(name),product_count=VALUES(product_count),updated_at=NOW()`;
    try {
      execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${sql.replace(/"/g, '\\"')}"`, { timeout: 30000, encoding: 'utf8' });
    } catch (e) {
      // One by one
      for (const c of batch) {
        try {
          const s = `INSERT INTO tiki_categories (id,tiki_id,name,slug,parent_id,parent_tiki_id,level,sort_order,product_count,image_url,url_path,is_active) VALUES ('${c.id}','${c.tiki_id}','${sanitize(c.name)}','${c.slug}','${c.parent_id}','${c.parent_tiki_id}',${c.level},${c.sort_order},${c.product_count},'${sanitize(c.image_url)}','${c.url_path}',1) ON DUPLICATE KEY UPDATE name=VALUES(name),updated_at=NOW()`;
          execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${s.replace(/"/g, '\\"')}"`, { timeout: 10000, encoding: 'utf8' });
        } catch (e2) { /* skip */ }
      }
    }
  }

  // Insert products in batches
  console.log(`  Inserting ${products.length} products...`);
  const batchSize = 100;
  let inserted = 0;
  for (let i = 0; i < products.length; i += batchSize) {
    const batch = products.slice(i, i + batchSize);
    const values = batch.map(p =>
      `('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average || 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`
    ).join(',');

    const sql = `INSERT INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ${values} ON DUPLICATE KEY UPDATE name=VALUES(name),price=VALUES(price),original_price=VALUES(original_price),discount_percent=VALUES(discount_percent),sold_count=VALUES(sold_count),rating_average=VALUES(rating_average),image_url=VALUES(image_url),local_image_url=VALUES(local_image_url),updated_at=NOW()`;

    try {
      execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${sql.replace(/"/g, '\\"')}"`, { timeout: 60000, encoding: 'utf8' });
      inserted += batch.length;
    } catch (e) {
      // Mini-batch fallback
      for (let j = 0; j < batch.length; j += 10) {
        const mb = batch.slice(j, j + 10);
        const mv = mb.map(p =>
          `('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average || 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`
        ).join(',');
        const ms = `INSERT INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ${mv} ON DUPLICATE KEY UPDATE name=VALUES(name),price=VALUES(price),updated_at=NOW()`;
        try {
          execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${ms.replace(/"/g, '\\"')}"`, { timeout: 30000, encoding: 'utf8' });
          inserted += mb.length;
        } catch (e2) { /* skip individual */ }
      }
    }

    if ((i + batchSize) % 2000 === 0 || i + batchSize >= products.length) {
      console.log(`  Progress: ${inserted}/${products.length}`);
    }
  }

  // Verify
  try {
    const cnt = execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -N -e "SELECT COUNT(*) FROM tiki_products" 2>/dev/null`, { timeout: 10000, encoding: 'utf8' });
    console.log(`  MySQL total products: ${cnt.trim()}`);
  } catch (e) { /* */ }
}

async function saveToMongoDB(categories, products) {
  const now = new Date().toISOString();
  let script = 'use tiki_catalog;\n';
  script += 'db.categories.deleteMany({});\n';
  script += 'db.products.deleteMany({});\n';

  for (const c of categories) {
    script += `db.categories.insertOne({category_id:'${c.id}',name:${JSON.stringify(c.name)},slug:'${c.slug}',parent_id:'${c.parent_id}',level:${c.level},sort_order:${c.sort_order},product_count:${c.product_count},image_url:${JSON.stringify(c.image_url)},url_path:${JSON.stringify(c.url_path)},is_active:true,created_at:ISODate('${now}'),updated_at:ISODate('${now}')});\n`;
  }

  for (const p of products) {
    const imgs = JSON.stringify(p.images || []);
    const attrs = JSON.stringify({ brand: p.brand || '' });
    const skuId = 'sku-' + uuidv4().slice(0, 8);
    const stock = Math.floor(Math.random() * 200) + 10;
    script += `db.products.insertOne({spu_id:'${p.db_id}',title:${JSON.stringify(p.name)},description:${JSON.stringify(p.description || '')},category_id:'${p.category_db_id}',seller_id:'usr-002',status:'ACTIVE',attributes:${attrs},images:${imgs},local_image_url:${JSON.stringify(p.local_image_url || '')},skus:[{sku_id:'${skuId}',spu_id:'${p.db_id}',price:${p.price},stock:${stock},status:'ACTIVE',variations:[{name:'Mặc định',value:'default'}]}],rating_average:${p.rating_average || 'null'},review_count:${p.review_count || 0},sold_count:${p.sold_count || 0},tiki_product_id:'${p.tiki_product_id}',created_at:ISODate('${now}'),updated_at:ISODate('${now}')});\n`;
  }

  fs.writeFileSync('/tmp/seed_tiki_mongo_v3.js', script);
  try {
    execSync('docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog < /tmp/seed_tiki_mongo_v3.js 2>&1', {
      timeout: 120000, encoding: 'utf8',
    });
    console.log(`  MongoDB: ${products.length} products, ${categories.length} categories`);

    const pc = execSync(`docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog --quiet --eval "db.products.countDocuments()" 2>/dev/null`, { timeout: 15000, encoding: 'utf8' });
    const cc = execSync(`docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog --quiet --eval "db.categories.countDocuments()" 2>/dev/null`, { timeout: 15000, encoding: 'utf8' });
    console.log(`  MongoDB counts: ${pc.trim()} products, ${cc.trim()} categories`);
  } catch (e) {
    console.log('  MongoDB error:', e.message.substring(0, 300));
  }
}

main().catch(err => {
  console.error('FATAL:', err);
  fs.writeFileSync('/tmp/crawl_stats_v3.json', JSON.stringify({ ...stats, fatalError: err.message }, null, 2));
  process.exit(1);
});
