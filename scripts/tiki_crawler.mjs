/**
 * Tiki.vn Mass Crawler v4
 * Uses Playwright to call Tiki API from within browser context (bypasses anti-bot)
 * Targets 50,000 products with local images
 *
 * Usage: node scripts/tiki_crawler.mjs [--resume] [--target=50000] [--skip-images]
 */

import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';
import https from 'https';
import http from 'http';
import { URL } from 'url';

// ===== Configuration =====
const TARGET_PRODUCTS = parseInt(process.argv.find(a => a.startsWith('--target='))?.split('=')[1] || '50000', 10);
const SKIP_IMAGES = process.argv.includes('--skip-images');
const RESUME = process.argv.includes('--resume');
const PRODUCTS_PER_PAGE = 40;
const MAX_PAGES_PER_CATEGORY = Math.ceil(TARGET_PRODUCTS / 40) + 50; // generous ceiling

const IMAGE_DIR = '/home/datdt/tikiclone/apps/web/public/images/products';
const CHECKPOINT_FILE = '/tmp/tiki_crawler_checkpoint.json';
const OUTPUT_FILE = '/tmp/tiki_crawler_products.json';

// 27 root categories (deduplicated)
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
  { id: '8992', slug: 'may-in', name: 'Máy In' },
];

const API_URL = 'https://tiki.vn/api/personalish/v1/blocks/listings';

// ===== State =====
let allProducts = [];
let seenIds = new Set();
let stats = {
  startTime: Date.now(),
  categoriesDone: 0,
  categoriesTotal: CATEGORIES.length,
  pagesFetched: 0,
  productsFound: 0,
  imagesDownloaded: 0,
  imagesFailed: 0,
  apiErrors: 0,
};

// ===== Helpers =====
function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }
function rand(min, max) { return min + Math.random() * (max - min); }

function extractPrice(v) {
  if (v === null || v === undefined) return 0;
  if (typeof v === 'number') return Math.round(v);
  return parseInt(String(v).replace(/[^0-9]/g, '')) || 0;
}

function sanitize(s) { return (s || '').replace(/['"]/g, '').trim().substring(0, 500); }

function cleanHtml(s) {
  return (s || '').replace(/<[^>]+>/g, '').replace(/['"]/g, '').trim().substring(0, 2000);
}

function getImageExtension(url) {
  try {
    const p = new URL(url).pathname.split('?')[0];
    const ext = path.extname(p).toLowerCase();
    if (['.jpg', '.jpeg', '.png', '.webp', '.gif'].includes(ext)) return ext;
  } catch {}
  return '.jpg';
}

function localFilename(tikiProductId, imageUrl) {
  return `spu-${tikiProductId}${getImageExtension(imageUrl)}`;
}

function localImagePath(tikiProductId, imageUrl) {
  return path.join(IMAGE_DIR, localFilename(tikiProductId, imageUrl));
}

function localImageUrl(tikiProductId, imageUrl) {
  return `/images/products/${localFilename(tikiProductId, imageUrl)}`;
}

function checkpointKey(cat, page) {
  return `${cat.id}:${page}`;
}

// ===== Checkpoint Manager =====
function loadCheckpoint() {
  if (!RESUME) return null;
  try {
    const data = JSON.parse(fs.readFileSync(CHECKPOINT_FILE, 'utf8'));
    if (data.allProducts) {
      allProducts = data.allProducts;
      seenIds = new Set(data.allProducts.map(p => p.tiki_product_id));
      stats = data.stats || { startTime: Date.now(), categoriesDone: 0, categoriesTotal: CATEGORIES.length, pagesFetched: 0, productsFound: allProducts.length, imagesDownloaded: 0, imagesFailed: 0, apiErrors: 0 };
      console.log(`Resumed: ${allProducts.length} products, ${data.doneKeys?.length || 0} pages done`);
    }
    return data.doneKeys || [];
  } catch {
    return [];
  }
}

function saveCheckpoint(doneKeys) {
  const data = {
    doneKeys,
    allProducts,
    stats,
    updatedAt: new Date().toISOString(),
  };
  fs.writeFileSync(CHECKPOINT_FILE, JSON.stringify(data));
  fs.writeFileSync(OUTPUT_FILE, JSON.stringify(allProducts));
}

// ===== Image Downloader =====
class ImageDownloader {
  constructor(concurrency = 10) {
    this.queue = [];
    this.active = 0;
    this.concurrency = concurrency;
    this.resolvePromise = null;
  }

  enqueue(imageUrl, destPath) {
    return new Promise((resolve) => {
      this.queue.push({ imageUrl, destPath, resolve });
      this.processNext();
    });
  }

  async processNext() {
    if (this.active >= this.concurrency || this.queue.length === 0) return;
    this.active++;
    const item = this.queue.shift();
    try {
      const ok = await this.downloadOne(item.imageUrl, item.destPath);
      item.resolve(ok);
    } catch {
      item.resolve(false);
    }
    this.active--;
    this.processNext();
  }

  async downloadOne(imageUrl, destPath, retries = 2) {
    if (fs.existsSync(destPath)) return true;
    for (let i = 0; i <= retries; i++) {
      try {
        await new Promise((resolve, reject) => {
          const mod = imageUrl.startsWith('https') ? https : http;
          const req = mod.get(imageUrl, {
            headers: {
              'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
              'Referer': 'https://tiki.vn/',
            },
            timeout: 15000,
          }, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
              this.downloadOne(res.headers.location, destPath, 0).then(resolve).catch(reject);
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
        if (i < retries) await sleep(500 + Math.random() * 1000);
      }
    }
    return false;
  }

  async waitForIdle() {
    while (this.active > 0 || this.queue.length > 0) {
      await sleep(100);
    }
  }
}

// ===== Transform Tiki product to our format =====
function transformProduct(item, category) {
  const tikiId = String(item.id || '');
  if (!tikiId) return null;

  // Extract primary image URL
  let primaryImage = '';
  const allImages = [];

  if (item.thumbnail_url) {
    primaryImage = item.thumbnail_url;
    allImages.push(item.thumbnail_url);
  }
  if (item.images && Array.isArray(item.images)) {
    for (const img of item.images) {
      const u = typeof img === 'string' ? img : (img?.url || img?.base_url || '');
      if (u && u.startsWith('http') && !allImages.includes(u)) allImages.push(u);
    }
  }

  const price = extractPrice(item.price || item.final_price);
  const origPrice = extractPrice(item.original_price || item.list_price);
  const hasDiscount = origPrice > price && origPrice > 0;

  return {
    tiki_product_id: tikiId,
    category_id: category.id,
    category_name: category.name || '',
    category_slug: category.slug || '',
    name: (item.name || '').substring(0, 500),
    description: cleanHtml(item.short_description || item.description || ''),
    brand: (item.brand_name || item.brand?.name || '').substring(0, 255),
    price,
    original_price: hasDiscount ? origPrice : null,
    discount_percent: hasDiscount ? Math.round((1 - price / origPrice) * 100) : null,
    compare_price: hasDiscount ? origPrice : null,
    images: allImages.filter(u => u.startsWith('http')).slice(0, 8),
    thumbnail_url: primaryImage,
    rating_average: item.rating_average || null,
    review_count: item.review_count || null,
    sold_count: (item.quantity_sold?.value) ?? item.order_count ?? null,
    url_path: item.url_path || '',
    is_deal: hasDiscount && item.is_deal !== false,
  };
}

// ===== Fetch products via page.evaluate(fetch) =====
// Uses browser's native fetch() — more reliable in background mode

async function fetchPage(page, ctx, category, pageNum) {
  const url = `${API_URL}?limit=${PRODUCTS_PER_PAGE}&category=${category.id}&page=${pageNum}`;

  try {
    const result = await page.evaluate(async (apiUrl) => {
      const res = await fetch(apiUrl, {
        headers: { 'Accept': 'application/json', 'Referer': 'https://tiki.vn/' },
      });
      if (!res.ok) {
        const text = await res.text().catch(() => '');
        return { error: `HTTP ${res.status} ${text.substring(0, 100)}`, data: [] };
      }
      const data = await res.json();
      return { error: null, data: data.data || [], paging: data.paging || {} };
    }, url);
    return result;
  } catch (e) {
    return { error: `EVALUATE: ${e.message.substring(0, 150)}`, data: [] };
  }
}

// ===== Main Crawl Function =====
async function crawlProducts() {
  console.log('========================================');
  console.log(`  TIKI.VN MASS CRAWLER v4`);
  console.log(`  Target: ${TARGET_PRODUCTS.toLocaleString()} products`);
  console.log(`  Categories: ${CATEGORIES.length}`);
  console.log(`  Resume: ${RESUME}`);
  console.log(`  Skip images: ${SKIP_IMAGES}`);
  console.log(`  Image dir: ${IMAGE_DIR}`);
  console.log('========================================\n');

  fs.mkdirSync(IMAGE_DIR, { recursive: true });

  // Load checkpoint
  const doneKeys = loadCheckpoint();
  const doneSet = new Set(doneKeys);

  if (!RESUME || allProducts.length === 0) {
    allProducts = [];
    seenIds = new Set();
    stats = { startTime: Date.now(), categoriesDone: 0, categoriesTotal: CATEGORIES.length, pagesFetched: 0, productsFound: 0, imagesDownloaded: 0, imagesFailed: 0, apiErrors: 0 };
  }

  // Launch browser
  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage'],
  });

  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
    viewport: { width: 1280, height: 900 },
  });

  let page = await ctx.newPage();

  // Establish session by visiting homepage
  try {
    await page.goto('https://tiki.vn/', { waitUntil: 'domcontentloaded', timeout: 30000 });
    await sleep(2000);
    console.log('Session established on tiki.vn\n');
  } catch (e) {
    console.warn('Homepage load warning (non-fatal):', e.message.substring(0, 80));
    // Still try to proceed
  }

  try {
    for (let ci = 0; ci < CATEGORIES.length; ci++) {
      const cat = CATEGORIES[ci];
      const prevTotal = allProducts.length;

      console.log(`[${ci + 1}/${CATEGORIES.length}] ${cat.name} (${cat.slug}) - ${allProducts.length.toLocaleString()} so far`);

      let pagesWithoutNew = 0;
      let consecutiveErrors = 0;

      for (let pg = 1; pg <= MAX_PAGES_PER_CATEGORY; pg++) {
        const key = checkpointKey(cat, pg);

        // Skip if already done
        if (doneSet.has(key)) {
          continue;
        }

        // Check if we hit target
        if (allProducts.length >= TARGET_PRODUCTS) {
          console.log(`  Reached ${TARGET_PRODUCTS.toLocaleString()} products, stopping crawl`);
          break;
        }

        const beforeCount = allProducts.length;

        try {
          const result = await fetchPage(page, ctx, cat, pg);

          if (result.error) {
            consecutiveErrors++;
            stats.apiErrors++;
            if (consecutiveErrors === 1) {
              process.stdout.write(`  Error on page ${pg} (cat ${cat.slug}): ${result.error}\n`);
            }
            if (consecutiveErrors >= 3) {
              console.log(`    3 consecutive API errors at page ${pg}, moving to next category`);
              break;
            }
            await sleep(rand(1000, 2000));
            continue;
          }

          consecutiveErrors = 0;
          const items = result.data || [];

          if (items.length === 0) {
            pagesWithoutNew++;
            if (pagesWithoutNew >= 2) break;
            continue;
          }

          pagesWithoutNew = 0;

          // Process items
          for (const item of items) {
            const transformed = transformProduct(item, cat);
            if (!transformed) continue;
            if (seenIds.has(transformed.tiki_product_id)) continue;

            seenIds.add(transformed.tiki_product_id);
            allProducts.push(transformed);
          }

          stats.pagesFetched++;

          // Mark as done
          doneSet.add(key);

          // Checkpoint every 10 pages
          if (allProducts.length % (PRODUCTS_PER_PAGE * 5) < PRODUCTS_PER_PAGE && allProducts.length > prevTotal) {
            saveCheckpoint([...doneSet]);
            const elapsed = ((Date.now() - stats.startTime) / 1000).toFixed(0);
            const rate = (allProducts.length / Math.max(1, elapsed)).toFixed(1);
            process.stdout.write(`  Page ${pg}: ${(allProducts.length - prevTotal).toLocaleString()} new, ${allProducts.length.toLocaleString()} total (${rate}/s)\n`);
          }

          // Small delay between pages
          await sleep(rand(200, 500));

        } catch (e) {
          consecutiveErrors++;
          stats.apiErrors++;
          await sleep(rand(1000, 3000));
          if (consecutiveErrors >= 5) {
            console.log(`  Too many errors, moving to next category`);
            break;
          }
        }

        if (allProducts.length >= TARGET_PRODUCTS) break;
      }

      stats.categoriesDone++;

      const newInCategory = allProducts.length - prevTotal;
      console.log(`  → ${newInCategory.toLocaleString()} new from this category (total: ${allProducts.length.toLocaleString()})`);

      // Save checkpoint after each category
      saveCheckpoint([...doneSet]);

      // Delay between categories
      await sleep(rand(800, 1500));

      if (allProducts.length >= TARGET_PRODUCTS) break;
    }

    console.log(`\n=== CRAWL COMPLETE: ${allProducts.length.toLocaleString()} products ===`);
    console.log(`Pages: ${stats.pagesFetched}, API errors: ${stats.apiErrors}`);
    saveCheckpoint([...doneSet]);

  } finally {
    await browser.close();
  }

  return allProducts;
}

// ===== Download Images =====
async function downloadAllImages() {
  if (SKIP_IMAGES) {
    console.log('\n=== Skipping image download (--skip-images) ===');
    return;
  }

  console.log('\n=== Downloading Images ===');
  const downloader = new ImageDownloader(10);

  const validProducts = allProducts.filter(p => p.thumbnail_url);
  console.log(`Products with images: ${validProducts.length}/${allProducts.length}`);

  const promises = [];
  for (const p of validProducts) {
    const dest = localImagePath(p.tiki_product_id, p.thumbnail_url);
    const url = p.thumbnail_url;
    const promise = downloader.enqueue(url, dest).then(ok => {
      if (ok) {
        p.local_image_url = localImageUrl(p.tiki_product_id, p.thumbnail_url);
        p.image_url = p.local_image_url;
        stats.imagesDownloaded++;
      } else {
        p.local_image_url = '';
        p.image_url = p.thumbnail_url;
        stats.imagesFailed++;
      }
      if ((stats.imagesDownloaded + stats.imagesFailed) % 500 === 0) {
        console.log(`  Images: ${stats.imagesDownloaded} ok, ${stats.imagesFailed} fail (${((stats.imagesDownloaded+stats.imagesFailed)/validProducts.length*100).toFixed(1)}%)`);
      }
    });
    promises.push(promise);

    // Throttle if we already have enough pending
    if (promises.length >= 500) {
      await Promise.race(promises);
      promises.splice(0, promises.findIndex(p => p !== undefined));
    }
  }

  // Wait for remaining downloads
  await Promise.all(promises);
  await downloader.waitForIdle();

  console.log(`\nImages: ${stats.imagesDownloaded} downloaded, ${stats.imagesFailed} failed`);

  // Save updated product data
  const [{ data }] = await Promise.all([
    fs.promises.writeFile(OUTPUT_FILE, JSON.stringify(allProducts)),
    fs.promises.writeFile(CHECKPOINT_FILE, JSON.stringify({
      doneKeys: [...loadCheckpoint()],
      allProducts,
      stats,
      updatedAt: new Date().toISOString(),
    })),
  ]);
}

// ===== Generate MongoDB Seed Script =====
async function generateMongoSeed() {
  console.log('\n=== Generating MongoDB Seed Script ===');

  const now = new Date().toISOString().replace('Z', '+0000');
  const sellerId = 'usr-002';
  const categoryMap = {};
  for (const c of CATEGORIES) {
    categoryMap[c.id] = c;
  }

  let script = 'const db = db.getSiblingDB(\'tiki_catalog\');\n\n';

  // Categories - merge existing + new from Tiki
  script += `// ===== Categories =====\n`;
  script += `db.categories.drop();\n\n`;
  script += `const categories = [\n`;
  const catEntries = CATEGORIES.map((c, i) => {
    const json = JSON.stringify({
      category_id: c.id,
      name: c.name,
      slug: c.slug,
      parent_id: '',
      level: 0,
      sort_order: i + 1,
    });
    return `  ${json}`;
  });
  script += catEntries.join(',\n');
  script += `\n];\n`;
  script += `db.categories.insertMany(categories);\n`;
  script += `print('Inserted ' + categories.length + ' categories');\n\n`;

  // Products
  script += `// ===== Products (${allProducts.length.toLocaleString()}) =====\n`;
  script += `db.products.drop();\n\n`;
  script += `const products = [\n`;

  let productCount = 0;
  for (const p of allProducts) {
    productCount++;
    const spuId = `spu-${p.tiki_product_id}`;
    const skuId = `sku-${p.tiki_product_id}`;

    const images = p.local_image_url ? [p.local_image_url] : (p.thumbnail_url ? [p.thumbnail_url] : []);
    const stock = Math.floor(Math.random() * 200) + 10;
    const rAvg = p.rating_average ?? 'null';

    let skuStr = `{sku_id:${JSON.stringify(skuId)},spu_id:${JSON.stringify(spuId)},price:${p.price},stock:${stock},status:'ACTIVE',variations:[{name:'Mặc định',value:'default'}],image:${JSON.stringify(p.local_image_url || p.thumbnail_url || '')}`;
    if (p.compare_price) skuStr += `,compare_price:${p.compare_price}`;
    skuStr += '}';

    script += `  {spu_id:${JSON.stringify(spuId)},title:${JSON.stringify(p.name)},description:${JSON.stringify(p.description || '')},category_id:${JSON.stringify(p.category_id)},seller_id:${JSON.stringify(sellerId)},status:'ACTIVE',attributes:${JSON.stringify({ brand: p.brand || '' })},images:${JSON.stringify(images)},local_image_url:${JSON.stringify(p.local_image_url || '')},skus:[${skuStr}],rating_average:${rAvg},review_count:${p.review_count || 0},sold_count:${p.sold_count || 0},is_deal:${p.is_deal || false},tiki_product_id:${JSON.stringify(p.tiki_product_id)},created_at:new Date(),updated_at:new Date()},\n`;
  }

  script += `];\n\n`;
  script += `db.products.insertMany(products);\n`;
  script += `print('Inserted ' + products.length + ' products');\n\n`;

  // Indexes
  script += `// ===== Indexes =====\n`;
  script += `db.products.createIndex({ spu_id: 1 }, { unique: true });\n`;
  script += `db.products.createIndex({ category_id: 1 });\n`;
  script += `db.products.createIndex({ status: 1 });\n`;
  script += `db.products.createIndex({ title: "text" });\n`;
  script += `db.products.createIndex({ "skus.price": 1 });\n`;
  script += `db.products.createIndex({ created_at: -1 });\n`;
  script += `db.categories.createIndex({ category_id: 1 }, { unique: true });\n`;
  script += `db.categories.createIndex({ slug: 1 });\n`;
  script += `db.categories.createIndex({ level: 1 });\n`;
  script += `print('Indexes created');\n`;

  const seedFile = '/tmp/seed_tiki_crawled.js';
  fs.writeFileSync(seedFile, script);
  console.log(`Seed script written to ${seedFile} (${(script.length / 1024 / 1024).toFixed(1)} MB)`);

  return seedFile;
}

// ===== Apply to MongoDB =====
async function applyToMongoDB(seedFile) {
  console.log('\n=== Applying to MongoDB ===');
  try {
    const { execSync } = await import('child_process');
    const result = execSync(
      `docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog < ${seedFile} 2>&1`,
      { timeout: 300000, encoding: 'utf8', maxBuffer: 1024 * 1024 * 50 }
    );

    // Extract insert counts
    const prodMatch = result.match(/Inserted (\d+) products/);
    const catMatch = result.match(/Inserted (\d+) categories/);
    console.log(`  MongoDB: ${prodMatch?.[1] || '?'} products, ${catMatch?.[1] || '?'} categories`);

    // Verify
    const count = execSync(
      `docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog --quiet --eval "db.products.countDocuments()" 2>/dev/null`,
      { timeout: 15000, encoding: 'utf8' }
    );
    console.log(`  Verified: ${count.trim()} products in MongoDB`);

    return true;
  } catch (e) {
    console.error('  MongoDB apply error:', e.message.substring(0, 300));
    return false;
  }
}

// ===== Summary =====
function printSummary() {
  const elapsed = ((Date.now() - stats.startTime) / 1000);
  const hours = Math.floor(elapsed / 3600);
  const mins = Math.floor((elapsed % 3600) / 60);
  const rate = (allProducts.length / Math.max(1, elapsed)).toFixed(1);

  console.log('\n========================================');
  console.log('  FINAL SUMMARY');
  console.log('========================================');
  console.log(`  Time: ${hours}h ${mins}m`);
  console.log(`  Products: ${allProducts.length.toLocaleString()}`);
  console.log(`  Categories: ${CATEGORIES.length}`);
  console.log(`  Pages fetched: ${stats.pagesFetched}`);
  console.log(`  API errors: ${stats.apiErrors}`);
  console.log(`  Rate: ${rate} products/s`);
  console.log(`  Images: ${stats.imagesDownloaded} downloaded, ${stats.imagesFailed} failed`);
  console.log(`  Image dir: ${IMAGE_DIR}`);
  console.log(`  Output: ${OUTPUT_FILE}`);

  const hasDeal = allProducts.filter(p => p.is_deal).length;
  const hasCompare = allProducts.filter(p => p.compare_price).length;
  console.log(`\n  Deals: ${hasDeal.toLocaleString()}`);
  console.log(`  With compare_price: ${hasCompare.toLocaleString()}`);
  console.log('========================================\n');

  const statsFile = '/tmp/tiki_crawler_stats.json';
  fs.writeFileSync(statsFile, JSON.stringify({ ...stats, endTime: Date.now(), elapsedSeconds: elapsed }, null, 2));
  console.log(`Stats saved to ${statsFile}`);
}

// ===== Main =====
async function main() {
  try {
    await crawlProducts();
    if (allProducts.length === 0) {
      console.log('No products found, exiting.');
      return;
    }
    await downloadAllImages();
    const seedFile = await generateMongoSeed();
    await applyToMongoDB(seedFile);
    printSummary();
  } catch (err) {
    console.error('\nFATAL ERROR:', err.message);
    console.error(err.stack?.substring(0, 500));
    // Save state on crash
    if (allProducts.length > 0) {
      fs.writeFileSync(OUTPUT_FILE, JSON.stringify(allProducts));
    }
    process.exit(1);
  }
}

main();
