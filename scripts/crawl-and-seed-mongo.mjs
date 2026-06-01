import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import { execSync } from 'child_process';

const BASE_URL = 'https://hoanghamobile.com';
const DEFAULT_SELLER_ID = 'usr-002';

const CATEGORIES = [
  { slug: 'dien-thoai', name: 'Điện thoại', path: '/dien-thoai' },
  { slug: 'laptop', name: 'Laptop', path: '/laptop' },
  { slug: 'tablet', name: 'Tablet', path: '/tablet' },
  { slug: 'phu-kien', name: 'Phụ kiện', path: '/phu-kien' },
  { slug: 'am-thanh', name: 'Âm thanh', path: '/am-thanh' },
  { slug: 'dong-ho', name: 'Đồng hồ thông minh', path: '/dong-ho' },
  { slug: 'man-hinh', name: 'Màn hình', path: '/man-hinh' },
  { slug: 'linh-kien', name: 'Máy tính linh kiện', path: '/laptop-linh-kien-may-tinh' },
  { slug: 'tivi', name: 'Tivi - Điện tử', path: '/tivi-do-dien-tu' },
];

function extractPrice(text) {
  if (!text) return null;
  const cleaned = text.replace(/\./g, '').replace(/[^\d]/g, '');
  return cleaned ? parseInt(cleaned, 10) : null;
}

function slugify(text) {
  return text.toLowerCase()
    .normalize('NFD').replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
}

async function crawlPage(browser, url) {
  const ctx = await browser.newContext({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36' });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);
  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    const products = await page.evaluate(() => {
      const items = [];
      const links = document.querySelectorAll('a[href*="/dien-thoai/"],a[href*="/laptop/"],a[href*="/tablet/"],a[href*="/phu-kien/"],a[href*="/am-thanh/"],a[href*="/dong-ho/"],a[href*="/man-hinh/"],a[href*="/tivi-"],a[href*="/linh-kien-"]');
      const seen = new Set();
      links.forEach(a => {
        const href = a.href;
        if (seen.has(href)) return;
        seen.add(href);
        const card = a.closest('[class*="product"],[class*="item"],[class*="pro-"],li,div.col') || a.parentElement;
        if (!card) return;
        const img = card.querySelector('img');
        const nameEl = card.querySelector('[class*="title"],[class*="name"],h2,h3,a[href*="/"]:not([class*="btn"])');
        const priceEl = card.querySelector('[class*="price"],[class*="gia-"],[class*="sale"]');
        const oldEl = card.querySelector('[class*="old"],[class*="original"],del');
        const name = (img?.alt || nameEl?.textContent || '').trim();
        if (!name || name.length <= 3 || name.length >= 200) return;
        const imgSrc = img?.getAttribute('data-src') || img?.getAttribute('src') || '';
        items.push({
          name,
          url: href,
          price: (priceEl?.textContent || '').trim(),
          oldPrice: (oldEl?.textContent || '').trim(),
          imageUrl: imgSrc.startsWith('//') ? 'https:' + imgSrc : imgSrc,
        });
      });
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
  const ctx = await browser.newContext({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36' });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);
  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    return await page.evaluate(() => {
      const imgs = [];
      document.querySelectorAll('img[class*="gallery"],img[class*="product"],img[class*="thumb"],[class*="gallery"] img,.slider img,.swiper img,[class*="product-image"] img').forEach(img => {
        const src = img.getAttribute('data-src') || img.getAttribute('data-lazy') || img.getAttribute('src') || '';
        const final = src.startsWith('//') ? 'https:' + src : src;
        if (final && !imgs.includes(final) && !final.includes('icon') && !final.includes('logo')) imgs.push(final);
      });
      const descEl = document.querySelector('.product-description,[class*="description"],.product-desc,.product-info,.tab-content');
      const desc = descEl ? descEl.textContent.trim().substring(0, 3000) : '';
      const brandEl = document.querySelector('[class*="brand"],.brand,[class*="thuong-hieu"]');
      const brand = brandEl ? brandEl.textContent.trim() : '';
      return { images: imgs.slice(0, 10), description: desc, brand };
    });
  } catch (err) {
    return { images: [], description: '', brand: '' };
  } finally {
    await ctx.close();
  }
}

function matchCategory(productUrl) {
  for (const cat of CATEGORIES) {
    if (productUrl.includes('/' + cat.slug + '/') || productUrl.startsWith(BASE_URL + cat.path + '/') || productUrl === BASE_URL + cat.path) {
      return cat;
    }
  }
  if (productUrl.includes('/may-tinh-bang/')) return CATEGORIES.find(c => c.slug === 'tablet') || CATEGORIES[0];
  if (productUrl.includes('/do-dien-tu/')) return CATEGORIES.find(c => c.slug === 'tivi') || CATEGORIES[0];
  return CATEGORIES[0];
}

async function main() {
  console.log('Starting hoanghamobile.com crawler...\n');

  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });

  try {
    // Step 1: Crawl all categories
    const allProducts = [];
    const seen = new Set();

    for (const cat of CATEGORIES) {
      const url = BASE_URL + cat.path;
      console.log(`Crawling ${cat.name}...`);
      const items = await crawlPage(browser, url);
      console.log(`  Found ${items.length} products`);
      for (const p of items) {
        if (!seen.has(p.url)) {
          seen.add(p.url);
          allProducts.push({ ...p, category: cat });
        }
      }
    }

    console.log(`\nTotal unique products: ${allProducts.length}`);

    // Step 2: Get details (limit to 50 for speed)
    const detailed = [];
    for (let i = 0; i < Math.min(allProducts.length, 50); i++) {
      const p = allProducts[i];
      console.log(`[${i+1}/${Math.min(50, allProducts.length)}] ${p.name.substring(0, 60)}...`);
      const detail = await crawlDetail(browser, p.url);
      const images = detail.images.length > 0 ? detail.images : (p.imageUrl ? [p.imageUrl] : []);
      const basePrice = extractPrice(p.oldPrice) || extractPrice(p.price) || 0;
      const salePrice = extractPrice(p.price) || null;

      detailed.push({
        title: p.name.substring(0, 500),
        description: detail.description.substring(0, 3000),
        category_slug: p.category.slug,
        category_name: p.category.name,
        brand: detail.brand,
        images,
        basePrice,
        salePrice,
        productUrl: p.url,
      });
    }

    // Step 3: Build MongoDB insert script
    const now = new Date().toISOString();
    let mongoScript = `
use tiki_catalog;

// Clear existing data
db.products.deleteMany({});
db.categories.deleteMany({});

// Insert categories
`;

    // Category id mapping
    const catIdMap = {};
    const catSlugToId = {};

    // Create store category
    const storeCatId = uuidv4();
    catIdMap['hoanghamobile'] = storeCatId;
    mongoScript += `db.categories.insertOne({
  category_id: '${storeCatId}',
  name: 'Hoang Ha Mobile',
  slug: 'hoang-ha-mobile',
  parent_id: '',
  level: 1,
  sort_order: 1,
  children: []
});\n`;

    // Sub-categories
    const catSlugs = [...new Set(detailed.map(p => p.category_slug))];
    for (const slug of catSlugs) {
      const cat = CATEGORIES.find(c => c.slug === slug);
      if (!cat) continue;
      const catId = uuidv4();
      catIdMap[slug] = catId;
      catSlugToId[slug] = catId;
      mongoScript += `db.categories.insertOne({
  category_id: '${catId}',
  name: '${cat.name.replace(/'/g, "\\'")}',
  slug: '${slug}',
  parent_id: '${storeCatId}',
  level: 2,
  sort_order: ${CATEGORIES.indexOf(cat) + 1},
  children: []
});\n`;
    }

    mongoScript += `\n// Insert products\n`;

    for (const p of detailed) {
      const productId = 'spu-' + uuidv4().slice(0, 8);
      const skuId = 'sku-' + uuidv4().slice(0, 8);
      const catId = catIdMap[p.category_slug] || storeCatId;
      const images = JSON.stringify(p.images.filter(u => u.startsWith('http')));
      const attrs = JSON.stringify({ brand: p.brand || '', source: p.productUrl });
      const variationName = 'Mặc định';

      mongoScript += `db.products.insertOne({
  spu_id: '${productId}',
  title: ${JSON.stringify(p.title)},
  description: ${JSON.stringify(p.description)},
  category_id: '${catId}',
  seller_id: '${DEFAULT_SELLER_ID}',
  status: 'active',
  attributes: ${attrs},
  images: ${images},
  skus: [{
    sku_id: '${skuId}',
    spu_id: '${productId}',
    price: ${p.salePrice || p.basePrice},
    compare_price: ${p.basePrice},
    stock: 100,
    status: 'active',
    variations: [{ name: ${JSON.stringify(variationName)}, value: 'default' }]
  }],
  created_at: ISODate('${now}'),
  updated_at: ISODate('${now}')
});\n`;
    }

    // Write the MongoDB script
    const fs = await import('fs');
    fs.writeFileSync('/tmp/seed-mongo.js', mongoScript);
    console.log(`\nMongoDB script written to /tmp/seed-mongo.js (${(mongoScript.length / 1024).toFixed(1)} KB)`);

    // Step 4: Execute against MongoDB
    console.log('Executing MongoDB seed...\n');
    try {
      const result = execSync('docker compose exec mongodb mongosh tiki_catalog /tmp/seed-mongo.js', {
        cwd: process.env.HOME || '/home/datdt/tikiclone',
        encoding: 'utf8',
        timeout: 120000,
      });
      console.log(result.split('\n').filter(l => l.trim() && !l.includes('time=') && !l.includes('WARNING')).slice(0, 10).join('\n'));
    } catch (err) {
      console.error('Failed to execute MongoDB seed:', err.message);
      console.log('Script saved at /tmp/seed-mongo.js - run manually with:');
      console.log('  docker compose exec -T mongodb mongosh tiki_catalog < /tmp/seed-mongo.js');
    }

    console.log(`\nDone! Crawled ${detailed.length} products.`);

  } finally {
    await browser.close();
  }
}

main().catch(err => { console.error(err); process.exit(1); });
