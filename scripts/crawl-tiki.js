#!/usr/bin/env node
/**
 * Tiki.vn Product Crawler
 * 
 * Crawls product data from tiki.vn and inserts into MySQL database.
 * Uses the browser automation approach since Tiki blocks direct API calls.
 * 
 * Usage: node crawl-tiki.js [--categories] [--products] [--all] [--max-pages N]
 */

const { execSync } = require('child_process');
const mysql = require('mysql2/promise');

// ─── Config ───────────────────────────────────────────────────────────
const DB_CONFIG = {
  host: 'localhost',
  port: 3306,
  user: 'tiki',
  password: 'tiki_dev',
  database: 'tiki_platform',
  charset: 'utf8mb4',
};

const TIKI_BASE = 'https://tiki.vn';
const USER_AGENT = 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';

// ─── Category mapping: Tiki category slug → our category ─────────────
const CATEGORY_MAP = {
  'dien-thoai-may-tinh-bang/c1789': { name: 'Điện Thoại & Máy Tính Bảng', slug: 'dien-thoai-may-tinh-bang' },
  'laptop-may-vi-tinh-linh-kien/c1846': { name: 'Laptop & Máy Tính', slug: 'laptop-may-vi-tinh' },
  'dien-gia-dung/c1882': { name: 'Điện Gia Dụng', slug: 'dien-gia-dung' },
  'me-be/c2549': { name: 'Mẹ & Bé', slug: 'me-be' },
  'khoe-dep/c1520': { name: 'Khỏe Đẹp', slug: 'khoe-dep' },
  'nha-cua-doi-song/c1883': { name: 'Nhà Cửa & Đời Sống', slug: 'nha-cua-doi-song' },
  'sach/c8320': { name: 'Sách', slug: 'sach' },
  'the-thao/c1975': { name: 'Thể Thao', slug: 'the-thao' },
  'thoi-trang-nu/c931': { name: 'Thời Trang Nữ', slug: 'thoi-trang-nu' },
  'thoi-trang-nam/c914': { name: 'Thời Trang Nam', slug: 'thoi-trang-nam' },
  'giay-dep-nu/c1706': { name: 'Giày Dép Nữ', slug: 'giay-dep-nu' },
  'giay-dep-nam/c1686': { name: 'Giày Dép Nam', slug: 'giay-dep-nam' },
  'tui-vi-nu/c976': { name: 'Túi Ví Nữ', slug: 'tui-vi-nu' },
  'dong-ho-trang-suc/c2843': { name: 'Đồng Hồ & Trang Sức', slug: 'dong-ho-trang-suc' },
  'may-anh/c1801': { name: 'Máy Ảnh', slug: 'may-anh' },
  'phu-kien-so/c2527': { name: 'Phụ Kiện Số', slug: 'phu-kien-so' },
  'oto-xe-may-xe-dap/c8594': { name: 'Ô Tô & Xe Máy', slug: 'oto-xe-may' },
  'bach-hoa-online/c4384': { name: 'Bách Hóa Online', slug: 'bach-hoa-online' },
  'do-choi/c2549': { name: 'Đồ Chơi', slug: 'do-choi' },
  'ngon/c4479': { name: 'Đồ Ăn & Đồ Uống', slug: 'ngon' },
};

// ─── Helpers ──────────────────────────────────────────────────────────
function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

function generateId() {
  return 'tki-' + Date.now().toString(36) + '-' + Math.random().toString(36).substring(2, 9);
}

function generateSlug(name) {
  return name
    .toLowerCase()
    .normalize('NFD').replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function parsePrice(text) {
  if (!text) return 0;
  const cleaned = text.replace(/[^\d]/g, '');
  return parseInt(cleaned, 10) || 0;
}

function extractProductIdFromUrl(url) {
  const match = url.match(/-p(\d+)\.html/);
  return match ? match[1] : null;
}

// ─── Fetch page HTML via curl (Tiki blocks Node.js http) ──────────────
function fetchPage(url) {
  try {
    const html = execSync(
      `curl -sL '${url.replace(/'/g, "'\\''")}' -H 'User-Agent: ${USER_AGENT}' -H 'Accept: text/html,application/xhtml+xml' --max-time 15`,
      { maxBuffer: 10 * 1024 * 1024, encoding: 'utf8' }
    );
    return html;
  } catch (e) {
    console.error(`  [ERROR] Failed to fetch: ${url} - ${e.message}`);
    return null;
  }
}

// ─── Extract JSON-LD product data from HTML ───────────────────────────
function extractJsonLd(html) {
  const results = [];
  const regex = /<script[^>]*type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/gi;
  let match;
  while ((match = regex.exec(html)) !== null) {
    try {
      const data = JSON.parse(match[1].trim());
      if (data['@graph']) {
        for (const item of data['@graph']) {
          if (item['@type'] === 'Product') results.push(item);
        }
      } else if (data['@type'] === 'Product') {
        results.push(data);
      }
    } catch (e) { /* skip malformed */ }
  }
  return results;
}

// ─── Extract product links from category page HTML ────────────────────
function extractProductLinks(html) {
  const links = new Set();
  // Match product URLs like /some-product-name-p12345678.html
  const regex = /href="(https:\/\/tiki\.vn\/[^"]+-p\d+\.html[^"]*)"/gi;
  let match;
  while ((match = regex.exec(html)) !== null) {
    const url = match[1].split('?')[0]; // strip query params
    links.add(url);
  }
  return [...links];
}

// ─── Extract category links from main page ────────────────────────────
function extractCategoryLinks(html) {
  const links = new Map();
  const regex = /href="(https:\/\/tiki\.vn\/([a-z0-9-]+\/c\d+))[^"]*"/gi;
  let match;
  while ((match = regex.exec(html)) !== null) {
    const url = match[1];
    const path = match[2];
    if (!links.has(path)) {
      links.set(path, url);
    }
  }
  return links;
}

// ─── Parse product detail page ────────────────────────────────────────
function parseProductPage(html, productUrl) {
  const jsonLdList = extractJsonLd(html);
  if (jsonLdList.length === 0) return null;

  const data = jsonLdList[0];
  const productId = extractProductIdFromUrl(productUrl);
  if (!productId) return null;

  // Extract price
  let price = 0;
  let salePrice = null;
  if (data.offers) {
    price = Math.round((data.offers.priceSpecification?.price || data.offers.price || 0));
    const originalPrice = parsePrice(data.offers.priceSpecification?.price);
    const currentPrice = parsePrice(String(data.offers.price));
    if (originalPrice > currentPrice) {
      salePrice = currentPrice;
      price = originalPrice;
    }
  }

  // Extract images
  const images = [];
  if (data.image) {
    if (Array.isArray(data.image)) {
      for (const img of data.image) {
        const url = typeof img === 'string' ? img : (img.url || img.contentUrl);
        if (url) images.push({ url, thumbnail: url.replace('/cache/750x750/', '/cache/200x200/') });
      }
    } else {
      const url = typeof data.image === 'string' ? data.image : (data.image.url || data.image.contentUrl);
      if (url) images.push({ url, thumbnail: url.replace('/cache/750x750/', '/cache/200x200/') });
    }
  }

  // Extract rating
  let rating = null;
  if (data.aggregateRating) {
    rating = {
      average: parseFloat(data.aggregateRating.ratingValue) || 0,
      count: parseInt(data.aggregateRating.reviewCount) || 0,
    };
  }

  // Extract breadcrumbs for category
  let categoryName = '';
  let categorySlug = '';
  if (data.mainEntityOfPage) {
    // Try to extract from breadcrumbs in HTML
  }

  // Extract attributes from additionalProperty
  const attributes = [];
  if (data.additionalProperty) {
    for (const prop of data.additionalProperty) {
      attributes.push({ name: prop.name, value: prop.value });
    }
  }

  // Extract sold count from HTML
  let soldCount = 0;
  const soldMatch = html.match(/Đã bán[\s]*(\d+)/);
  if (soldMatch) soldCount = parseInt(soldMatch[1], 10);

  // Extract description
  let description = data.description || '';
  // Clean HTML tags from description
  description = description.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();

  return {
    tikiId: productId,
    name: data.name || '',
    description,
    brand: (data.brand?.name || data.manufacturer?.name || ''),
    price,
    salePrice,
    images,
    rating,
    attributes,
    soldCount,
    url: productUrl,
    color: data.color || '',
    material: data.material || '',
    sku: data.sku || data.mpn || '',
  };
}

// ─── Database operations ──────────────────────────────────────────────
class Database {
  constructor() {
    this.conn = null;
  }

  async connect() {
    this.conn = await mysql.createConnection(DB_CONFIG);
    console.log('[DB] Connected to MySQL');
  }

  async close() {
    if (this.conn) await this.conn.end();
  }

  async upsertCategory(name, slug, parentId = null, level = 1, imageUrl = null) {
    const id = 'cat-' + slug.replace(/[^a-z0-9]/g, '-').substring(0, 30);
    await this.conn.execute(
      `INSERT INTO categories (id, parent_id, name, slug, level, image_url, is_active)
       VALUES (?, ?, ?, ?, ?, ?, 1)
       ON DUPLICATE KEY UPDATE name = VALUES(name), image_url = VALUES(image_url), updated_at = NOW()`,
      [id, parentId, name, slug, level, imageUrl]
    );
    return id;
  }

  async upsertProduct(product, categoryId, shopId) {
    const id = 'prod-' + product.tikiId;
    await this.conn.execute(
      `INSERT INTO products (id, shop_id, category_id, name, description, brand, status, currency, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, 'active', 'VND', NOW(), NOW())
       ON DUPLICATE KEY UPDATE
         name = VALUES(name), description = VALUES(description), brand = VALUES(brand),
         updated_at = NOW()`,
      [id, shopId, categoryId, product.name, product.description, product.brand]
    );
    return id;
  }

  async upsertSku(productDbId, product) {
    const id = 'sku-' + product.tikiId;
    const skuCode = product.sku || ('TKI-' + product.tikiId);
    await this.conn.execute(
      `INSERT INTO skus (id, product_id, sku_code, price, sale_price, stock, status, attributes, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, 'active', ?, NOW(), NOW())
       ON DUPLICATE KEY UPDATE
         price = VALUES(price), sale_price = VALUES(sale_price), stock = VALUES(stock),
         attributes = VALUES(attributes), updated_at = NOW()`,
      [id, productDbId, skuCode, product.price, product.salePrice, Math.max(10, product.soldCount * 2), JSON.stringify(product.attributes)]
    );
    return id;
  }

  async upsertMedia(productDbId, images) {
    // Delete old media for this product
    await this.conn.execute('DELETE FROM product_media WHERE product_id = ?', [productDbId]);

    for (let i = 0; i < images.length; i++) {
      const img = images[i];
      const id = 'img-' + productDbId.substring(5) + '-' + i;
      await this.conn.execute(
        `INSERT INTO product_media (id, product_id, media_type, url, thumbnail_url, alt_text, sort_order, is_primary, created_at)
         VALUES (?, ?, 'image', ?, ?, ?, ?, ?, NOW())`,
        [id, productDbId, img.url, img.thumbnail, '', i, i === 0 ? 1 : 0]
      );
    }
  }

  async getProductCount() {
    const [rows] = await this.conn.execute('SELECT COUNT(*) as cnt FROM products');
    return rows[0].cnt;
  }

  async getCategoryCount() {
    const [rows] = await this.conn.execute('SELECT COUNT(*) as cnt FROM categories');
    return rows[0].cnt;
  }
}

// ─── Main crawler ─────────────────────────────────────────────────────
class TikiCrawler {
  constructor(db, options = {}) {
    this.db = db;
    this.maxPages = options.maxPages || 3;
    this.delay = options.delay || 1500; // ms between requests
    this.shopId = 'shop-tiki-trading';
    this.stats = { categories: 0, products: 0, errors: 0, skipped: 0 };
  }

  async crawlCategories() {
    console.log('\n[CRAWL] Fetching main page for categories...');
    const html = fetchPage(TIKI_BASE);
    if (!html) return;

    const catLinks = extractCategoryLinks(html);
    console.log(`[CRAWL] Found ${catLinks.size} category links`);

    let order = 0;
    for (const [path, url] of catLinks) {
      const mapping = CATEGORY_MAP[path];
      if (!mapping) {
        // Still crawl unknown categories
        const name = path.split('/')[0].replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
        const slug = generateSlug(name);
        await this.db.upsertCategory(name, slug, null, 1);
        this.stats.categories++;
        console.log(`  [CAT] Inserted: ${name} (${slug})`);
      } else {
        await this.db.upsertCategory(mapping.name, mapping.slug, null, 1);
        this.stats.categories++;
        console.log(`  [CAT] Inserted: ${mapping.name} (${mapping.slug})`);
      }
      order++;
      await sleep(100);
    }
  }

  async crawlCategoryProducts(categorySlug, categoryId, maxPages = null) {
    const pages = maxPages || this.maxPages;
    console.log(`\n[CRAWL] Crawling products for category: ${categorySlug} (max ${pages} pages)`);

    const allProductUrls = [];

    for (let page = 1; page <= pages; page++) {
      const url = `${TIKI_BASE}/${categorySlug}/c${this.getCategoryId(categorySlug)}?page=${page}`;
      console.log(`  [PAGE] Fetching page ${page}: ${url}`);

      const html = fetchPage(url);
      if (!html) {
        this.stats.errors++;
        continue;
      }

      const links = extractProductLinks(html);
      console.log(`  [PAGE] Found ${links.length} product links on page ${page}`);

      if (links.length === 0) {
        console.log('  [PAGE] No more products, stopping.');
        break;
      }

      allProductUrls.push(...links);
      await sleep(this.delay);
    }

    // Deduplicate
    const uniqueUrls = [...new Set(allProductUrls)];
    console.log(`\n[CRAWL] Total unique products: ${uniqueUrls.length}`);

    // Now crawl each product detail page
    for (let i = 0; i < uniqueUrls.length; i++) {
      const productUrl = uniqueUrls[i];
      console.log(`  [PRODUCT] ${i + 1}/${uniqueUrls.length}: ${productUrl}`);

      try {
        const html = fetchPage(productUrl);
        if (!html) {
          this.stats.errors++;
          await sleep(this.delay);
          continue;
        }

        const product = parseProductPage(html, productUrl);
        if (!product || !product.name) {
          console.log(`    [SKIP] No product data found`);
          this.stats.skipped++;
          await sleep(this.delay);
          continue;
        }

        // Insert into DB
        const productDbId = await this.db.upsertProduct(product, categoryId, this.shopId);
        await this.db.upsertSku(productDbId, product);

        if (product.images.length > 0) {
          await this.db.upsertMedia(productDbId, product.images);
        }

        this.stats.products++;
        console.log(`    [OK] ${product.name.substring(0, 60)} - ${product.price.toLocaleString()}₫`);
      } catch (e) {
        console.error(`    [ERROR] ${e.message}`);
        this.stats.errors++;
      }

      await sleep(this.delay);
    }
  }

  getCategoryId(slug) {
    // Map slug to Tiki category ID
    const map = {
      'dien-thoai-may-tinh-bang': 1789,
      'laptop-may-vi-tinh': 1846,
      'dien-gia-dung': 1882,
      'me-be': 2549,
      'khoe-dep': 1520,
      'nha-cua-doi-song': 1883,
      'sach': 8320,
      'the-thao': 1975,
      'thoi-trang-nu': 931,
      'thoi-trang-nam': 914,
      'giay-dep-nu': 1706,
      'giay-dep-nam': 1686,
      'tui-vi-nu': 976,
      'dong-ho-trang-suc': 2843,
      'may-anh': 1801,
      'phu-kien-so': 2527,
      'oto-xe-may': 8594,
      'bach-hoa-online': 4384,
      'do-choi': 2549,
      'ngon': 4479,
    };
    return map[slug] || 1789;
  }

  async crawlAll() {
    // First crawl categories
    await this.crawlCategories();

    // Get all categories from DB
    const [rows] = await this.db.conn.execute('SELECT id, slug FROM categories WHERE is_active = 1');

    for (const cat of rows) {
      await this.crawlCategoryProducts(cat.slug, cat.id);
    }
  }

  printStats() {
    console.log('\n═══════════════════════════════════════');
    console.log('CRAWL STATS:');
    console.log(`  Categories: ${this.stats.categories}`);
    console.log(`  Products:   ${this.stats.products}`);
    console.log(`  Skipped:    ${this.stats.skipped}`);
    console.log(`  Errors:     ${this.stats.errors}`);
    console.log('═══════════════════════════════════════\n');
  }
}

// ─── CLI ───────────────────────────────────────────────────────────────
async function main() {
  const args = process.argv.slice(2);
  const options = {
    categories: args.includes('--categories'),
    products: args.includes('--products'),
    all: args.includes('--all'),
    maxPages: 3,
  };

  // Parse --max-pages
  const maxIdx = args.indexOf('--max-pages');
  if (maxIdx !== -1 && args[maxIdx + 1]) {
    options.maxPages = parseInt(args[maxIdx + 1], 10);
  }

  // Default to --all if no specific option
  if (!options.categories && !options.products && !options.all) {
    options.all = true;
  }

  const db = new Database();
  await db.connect();

  const crawler = new TikiCrawler(db, { maxPages: options.maxPages });

  try {
    if (options.all) {
      await crawler.crawlAll();
    } else if (options.categories) {
      await crawler.crawlCategories();
    } else if (options.products) {
      // Crawl products for a specific category
      const catSlug = args[args.indexOf('--products') + 1] || 'dien-thoai-may-tinh-bang';
      const [rows] = await db.conn.execute('SELECT id FROM categories WHERE slug = ?', [catSlug]);
      if (rows.length > 0) {
        await crawler.crawlCategoryProducts(catSlug, rows[0].id);
      } else {
        console.error(`Category not found: ${catSlug}`);
      }
    }

    crawler.printStats();

    const productCount = await db.getProductCount();
    const categoryCount = await db.getCategoryCount();
    console.log(`[DB] Total products in DB: ${productCount}`);
    console.log(`[DB] Total categories in DB: ${categoryCount}`);
  } catch (e) {
    console.error('[FATAL]', e);
  } finally {
    await db.close();
  }
}

main();
