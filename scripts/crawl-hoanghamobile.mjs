import { chromium } from 'playwright';
import mysql from 'mysql2/promise';
import { v4 as uuidv4 } from 'uuid';

const DB_CONFIG = {
  host: process.env.DB_HOST || 'localhost',
  port: parseInt(process.env.DB_PORT || '3306'),
  user: process.env.DB_USER || 'tiki',
  password: process.env.DB_PASSWORD || 'tiki_dev',
  database: process.env.DB_NAME || 'tiki_platform',
};

const BASE_URL = 'https://hoanghamobile.com';
const DEFAULT_SHOP_ID = 'usr-002';

const CATEGORIES = [
  { name: 'Điện thoại', slug: 'dien-thoai', path: '/dien-thoai' },
  { name: 'Laptop', slug: 'laptop', path: '/laptop' },
  { name: 'Tablet', slug: 'tablet', path: '/tablet' },
  { name: 'Phụ kiện', slug: 'phu-kien', path: '/phu-kien' },
  { name: 'Âm thanh', slug: 'am-thanh', path: '/am-thanh' },
  { name: 'Đồng hồ thông minh', slug: 'dong-ho', path: '/dong-ho' },
  { name: 'Màn hình', slug: 'man-hinh', path: '/man-hinh' },
  { name: 'Máy tính linh kiện', slug: 'linh-kien', path: '/laptop-linh-kien-may-tinh' },
  { name: 'Tivi - Điện tử', slug: 'tivi', path: '/tivi-do-dien-tu' },
];

const CATEGORY_SLUGS = new Set(CATEGORIES.map(c => c.slug));

function extractPrice(text) {
  if (!text) return null;
  const cleaned = text.replace(/\./g, '').replace(/[^\d]/g, '');
  return cleaned ? parseInt(cleaned, 10) : null;
}

function matchCategory(url) {
  for (const cat of CATEGORIES) {
    if (url.includes('/' + cat.slug + '/') || url.includes('/' + cat.slug) && !/[a-z]/.test(url.split('/' + cat.slug).pop()?.[0] || '')) {
      // Try exact match
    }
    if (url.includes('/' + cat.slug + '/') || url === BASE_URL + cat.path || url.startsWith(BASE_URL + cat.path + '/')) {
      return cat;
    }
  }
  return CATEGORIES[0];
}

async function upsertCategory(conn, cat) {
  const [existing] = await conn.execute('SELECT id FROM categories WHERE slug = ? LIMIT 1', [cat.slug]);
  if (existing.length > 0) return existing[0].id;
  const id = uuidv4();
  await conn.execute(
    `INSERT INTO categories (id, name, slug, level, sort_order, is_active, created_at, updated_at)
     VALUES (?, ?, ?, 1, 0, 1, NOW(), NOW())`,
    [id, cat.name, cat.slug]
  );
  return id;
}

async function crawlPage(browser, url) {
  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36'
  });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);

  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);

    const products = await page.evaluate(() => {
      const seen = new Set();
      const items = [];

      const selectors = [
        'a[href*="/dien-thoai/"]', 'a[href*="/laptop/"]', 'a[href*="/tablet/"]',
        'a[href*="/phu-kien/"]', 'a[href*="/am-thanh/"]', 'a[href*="/dong-ho/"]',
        'a[href*="/man-hinh/"]', 'a[href*="/tivi-"]', 'a[href*="/linh-kien-"]',
        'a[href*="/may-tinh-bang/"]', 'a[href*="/do-dien-tu/"]',
      ];

      const productLinks = new Set();
      for (const sel of selectors) {
        document.querySelectorAll(sel).forEach(a => {
          const h = a.href;
          if (h && !seen.has(h)) {
            seen.add(h);
            const card = a.closest('[class*="product"],[class*="item"],[class*="pro-"],li,div.col') || a.parentElement;
            if (card) {
              const img = card.querySelector('img');
              const nameEl = card.querySelector('[class*="title"],[class*="name"],h2,h3,a[href*="/"]:not([class*="btn"])');
              const priceEl = card.querySelector('[class*="price"],[class*="gia-"],[class*="sale"]');
              const oldEl = card.querySelector('[class*="old"],[class*="original"],del');

              const name = (img?.alt || nameEl?.textContent || '').trim();
              if (name && name.length > 3 && name.length < 200) {
                const imgSrc = img?.getAttribute('data-src') || img?.getAttribute('src') || '';
                items.push({
                  name,
                  url: h,
                  price: priceEl?.textContent?.trim() || '',
                  oldPrice: oldEl?.textContent?.trim() || '',
                  imageUrl: imgSrc.startsWith('//') ? 'https:' + imgSrc : imgSrc,
                });
              }
            }
          }
        });
      }

      return items;
    });

    return products;
  } catch (err) {
    console.log(`  Error: ${err.message}`);
    return [];
  } finally {
    await ctx.close();
  }
}

async function crawlDetail(browser, url) {
  const ctx = await browser.newContext({
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36'
  });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);

  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);

    return await page.evaluate(() => {
      const imgs = [];
      document.querySelectorAll('img[class*="gallery"],img[class*="product"],img[class*="thumb"],[class*="gallery"] img,[class*="product-image"] img,.slider img,.swiper img').forEach(img => {
        const src = img.getAttribute('data-src') || img.getAttribute('data-lazy') || img.getAttribute('src') || '';
        const final = src.startsWith('//') ? 'https:' + src : src;
        if (final && !imgs.includes(final) && !final.includes('icon') && !final.includes('logo')) imgs.push(final);
      });

      const descEl = document.querySelector('.product-description,[class*="description"],.product-desc,.product-info,.tab-content');
      const description = descEl ? descEl.textContent.trim().substring(0, 3000) : '';

      const brandEl = document.querySelector('[class*="brand"],.brand,[class*="thuong-hieu"]');
      const brand = brandEl ? brandEl.textContent.trim() : '';

      return { images: imgs.slice(0, 10), description, brand };
    });
  } catch (err) {
    return { images: [], description: '', brand: '' };
  } finally {
    await ctx.close();
  }
}

async function main() {
  console.log('Connecting to MySQL...');
  const conn = await mysql.createConnection(DB_CONFIG);
  console.log('Connected!');

  const catIds = {};
  for (const cat of CATEGORIES) {
    catIds[cat.slug] = await upsertCategory(conn, cat);
  }
  console.log('Categories ready');

  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });

  try {
    const allProducts = [];
    const seen = new Set();

    for (const cat of CATEGORIES.slice(0, 6)) {
      const url = BASE_URL + cat.path;
      console.log(`\nCrawling ${cat.name}...`);
      const items = await crawlPage(browser, url);
      console.log(`  Found ${items.length} products`);
      for (const p of items) {
        if (!seen.has(p.url)) {
          seen.add(p.url);
          allProducts.push(p);
        }
      }
    }

    console.log(`\nTotal unique: ${allProducts.length}`);

    let inserted = 0;
    for (const p of allProducts.slice(0, 80)) {
      const cat = matchCategory(p.url);
      const categoryId = catIds[cat.slug];

      console.log(`[${inserted + 1}] ${p.name.substring(0, 60)}...`);

      try {
        const detail = await crawlDetail(browser, p.url);
        const images = detail.images.length > 0 ? detail.images : (p.imageUrl ? [p.imageUrl] : []);
        const basePrice = extractPrice(p.oldPrice) || extractPrice(p.price) || 0;
        const salePrice = extractPrice(p.price) || null;

        const productId = uuidv4();
        await conn.execute(
          `INSERT INTO products (id, shop_id, category_id, name, description, brand, status, currency, version, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?, ?, 'active', 'VND', 1, NOW(), NOW())`,
          [productId, DEFAULT_SHOP_ID, categoryId, p.name.substring(0, 500), detail.description, detail.brand]
        );

        const skuCode = `SKU-${productId.slice(0, 8).toUpperCase()}`;
        await conn.execute(
          `INSERT INTO skus (id, product_id, sku_code, price, sale_price, stock, status, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?, 100, 'active', NOW(), NOW())`,
          [uuidv4(), productId, skuCode, basePrice, salePrice]
        );

        for (let i = 0; i < images.length; i++) {
          await conn.execute(
            `INSERT INTO product_media (id, product_id, media_type, url, thumbnail_url, alt_text, sort_order, is_primary, created_at)
             VALUES (?, ?, 'image', ?, ?, ?, ?, ?, NOW())`,
            [uuidv4(), productId, images[i], images[i], p.name, i, i === 0 ? 1 : 0]
          );
        }

        inserted++;
      } catch (err) {
        console.log(`  Error: ${err.message}`);
      }
    }

    console.log(`\nDone! Inserted ${inserted} products.`);
  } finally {
    await browser.close();
    await conn.end();
  }
}

main().catch(err => { console.error(err); process.exit(1); });
