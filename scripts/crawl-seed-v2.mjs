import { chromium } from 'playwright';
import { v4 as uuidv4 } from 'uuid';
import { execSync } from 'child_process';
import fs from 'fs';

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
  // Handle "34,790,000 ₫" format (comma as thousands separator)
  let cleaned = text.replace(/,/g, '').replace(/\./g, '');
  // Also handle "34.790.000₫" format (dot as thousands separator)
  cleaned = cleaned.replace(/[^\d]/g, '');
  const num = parseInt(cleaned, 10);
  return (num && num > 0) ? num : null;
}

function matchCategory(productUrl, productName) {
  for (const cat of CATEGORIES) {
    if (productUrl.includes('/' + cat.slug + '/') || productUrl.startsWith(BASE_URL + cat.path + '/') || productUrl === BASE_URL + cat.path) {
      return cat;
    }
  }
  if (productUrl.includes('/may-tinh-bang/')) return CATEGORIES.find(c => c.slug === 'tablet') || CATEGORIES[0];
  if (productUrl.includes('/do-dien-tu/')) return CATEGORIES.find(c => c.slug === 'tivi') || CATEGORIES[0];
  // Infer from name
  const name = (productName || '').toLowerCase();
  if (name.includes('điện thoại') || name.includes('iphone') || name.includes('samsung') || name.includes('xiaomi')) return CATEGORIES[0];
  if (name.includes('laptop') || name.includes('macbook') || name.includes('dell') || name.includes('asus') || name.includes('lenovo') || name.includes('hp') || name.includes('acer')) return CATEGORIES[1];
  if (name.includes('tablet') || name.includes('ipad')) return CATEGORIES[2];
  if (name.includes('tai nghe') || name.includes('loa') || name.includes('soundbar')) return CATEGORIES[4];
  if (name.includes('đồng hồ') || name.includes('watch')) return CATEGORIES[5];
  if (name.includes('màn hình') || name.includes('monitor')) return CATEGORIES[6];
  return CATEGORIES[1]; // Default to laptop
}

async function crawlCategoryProducts(browser, cat) {
  const ctx = await browser.newContext({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36' });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);
  const url = BASE_URL + cat.path;
  const products = [];
  try {
    console.log(`  Navigating to ${url}...`);
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(3000);
    const items = await page.evaluate(() => {
      const result = [];
      // Main product grid
      const grid = document.querySelector('.v5-grid-items, [class*="grid"]');
      if (!grid) return result;
      const links = grid.querySelectorAll('a[href*="/' + window.location.pathname.split('/')[1] + '/"]');
      const seen = new Set();
      links.forEach(a => {
        const href = a.href;
        if (seen.has(href)) return;
        seen.add(href);
        const card = a.closest('[class*="price-tags-home"]') ? a : a.parentElement;
        const priceEl = card.querySelector('.price-tags-home, [class*="price"], [class*="gia"]');
        const priceText = priceEl ? priceEl.textContent.trim() : '';
        const img = card.querySelector('img');
        const imgSrc = (img?.getAttribute('data-src') || img?.getAttribute('src') || '').startsWith('//')
          ? 'https:' + (img?.getAttribute('data-src') || img?.getAttribute('src') || '')
          : (img?.getAttribute('data-src') || img?.getAttribute('src') || '');
        const name = (a.textContent || img?.alt || '').trim();
        if (name && name.length > 3 && name.length < 200 && !href.includes('/laptop/van-phong') && !href.includes('/laptop/gaming')) {
          result.push({ name: name.substring(0, 200), url: href, price: priceText, imageUrl: imgSrc });
        }
      });
      return result;
    });
    products.push(...items.map(p => ({ ...p, category: cat })));
  } catch (err) {
    console.log(`  Error crawling ${cat.name}: ${err.message}`);
  } finally {
    await ctx.close();
  }
  return products;
}

async function crawlProductDetail(browser, url) {
  const ctx = await browser.newContext({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36' });
  const page = await ctx.newPage();
  page.setDefaultTimeout(30000);
  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    return await page.evaluate(() => {
      const imgs = [];
      // Main product gallery images
      document.querySelectorAll('.swiper-slide img, [class*="gallery"] img, [class*="product-image"] img, .product-thumb img, .image-gallery img, img[class*="lazy"]').forEach(img => {
        const src = img.getAttribute('data-src') || img.getAttribute('src') || '';
        const final = src.startsWith('//') ? 'https:' + src : src;
        if (final && !imgs.includes(final) && !final.includes('icon') && !final.includes('logo') && final.includes('hoanghamobile')) imgs.push(final);
      });
      // Description
      const descEl = document.querySelector('.product-description, [class*="description"], .product-desc, .tab-content, [class*="product-info"]');
      const desc = descEl ? descEl.textContent.trim().substring(0, 3000) : '';
      // Brand
      const brandEl = document.querySelector('[class*="brand"], .brand');
      const brand = brandEl ? brandEl.textContent.trim() : '';
      // Also try to get price from detail page
      const priceEl = document.querySelector('.product-detail .price, [class*="product-price"], .current-price, .price-box .price');
      const detailPrice = priceEl ? priceEl.textContent.trim() : '';
      return { images: imgs.slice(0, 8), description: desc, brand, detailPrice };
    });
  } catch (err) {
    return { images: [], description: '', brand: '', detailPrice: '' };
  } finally {
    await ctx.close();
  }
}

async function main() {
  console.log('=== Hoang Ha Mobile Crawler v2 ===\n');
  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
  const allProducts = [];
  const seen = new Set();

  try {
    // Step 1: Crawl each category page
    for (const cat of CATEGORIES.slice(0, 3)) {
      console.log(`\n--- ${cat.name} ---`);
      const items = await crawlCategoryProducts(browser, cat);
      console.log(`  Found ${items.length} products`);
      for (const p of items) {
        const key = p.url.split('?')[0];
        if (!seen.has(key)) {
          seen.add(key);
          allProducts.push(p);
        }
      }
    }

    console.log(`\nTotal unique products: ${allProducts.length}`);

    // Step 2: Get details for each product
    const detailed = [];
    for (let i = 0; i < Math.min(allProducts.length, 60); i++) {
      const p = allProducts[i];
      const name = p.name.substring(0, 60);
      process.stdout.write(`[${i+1}/${Math.min(allProducts.length, 60)}] ${' '.repeat(60)}`);
      process.stdout.write(`\r[${i+1}/${Math.min(allProducts.length, 60)}] ${name}`);
      
      const detail = await crawlProductDetail(browser, p.url);
      const images = detail.images.length > 0 ? detail.images : (p.imageUrl ? [p.imageUrl] : []);
      
      // Price: try detail page price first, then listing price
      const price = extractPrice(detail.detailPrice) || extractPrice(p.price) || 0;
      const cat = matchCategory(p.url, p.name);

      detailed.push({
        title: p.name.substring(0, 500),
        description: detail.description.substring(0, 3000),
        category_slug: cat.slug,
        category_name: cat.name,
        brand: detail.brand || '',
        images,
        price,
        productUrl: p.url,
      });
    }

    console.log(`\n\nCrawled ${detailed.length} products with details.`);

    // Step 3: Build MongoDB seed script
    const now = new Date().toISOString();
    const storeCatId = uuidv4();
    const catMap = {};

    let script = `
use tiki_catalog;
db.products.deleteMany({});
db.categories.deleteMany({});

// Categories
db.categories.insertOne({ category_id: '${storeCatId}', name: 'Hoang Ha Mobile', slug: 'hoang-ha-mobile', parent_id: '', level: 1, sort_order: 1, children: [] });
`;

    const catSlugs = [...new Set(detailed.map(p => p.category_slug))];
    let order = 1;
    for (const slug of catSlugs) {
      const cat = CATEGORIES.find(c => c.slug === slug) || { name: slug, slug };
      const catId = uuidv4();
      catMap[slug] = catId;
      script += `db.categories.insertOne({ category_id: '${catId}', name: ${JSON.stringify(cat.name)}, slug: '${slug}', parent_id: '${storeCatId}', level: 2, sort_order: ${order++}, children: [] });\n`;
    }

    script += `\n// Products\n`;
    for (const p of detailed) {
      const productId = 'spu-' + uuidv4().slice(0, 8);
      const skuId = 'sku-' + uuidv4().slice(0, 8);
      const catId = catMap[p.category_slug] || storeCatId;
      const images = JSON.stringify(p.images.filter(u => u.startsWith('http')));
      const attrs = JSON.stringify({ brand: p.brand || '' });
      const variationName = 'Mặc định';
      const price = p.price;

      script += `db.products.insertOne({
  spu_id: '${productId}',
  title: ${JSON.stringify(p.title)},
  description: ${JSON.stringify(p.description)},
  category_id: '${catId}',
  seller_id: '${DEFAULT_SELLER_ID}',
  status: 'ACTIVE',
  attributes: ${attrs},
  images: ${images},
  skus: [{
    sku_id: '${skuId}',
    spu_id: '${productId}',
    price: ${price},
    stock: 100,
    status: 'ACTIVE',
    variations: [{ name: ${JSON.stringify(variationName)}, value: 'default' }]
  }],
  created_at: ISODate('${now}'),
  updated_at: ISODate('${now}')
});\n`;
    }

    // Also add some products for remaining categories with placeholder data
    const remainingCats = CATEGORIES.slice(3);
    for (const cat of remainingCats) {
      if (catMap[cat.slug]) continue;
      const catId = uuidv4();
      catMap[cat.slug] = catId;
      script += `db.categories.insertOne({ category_id: '${catId}', name: ${JSON.stringify(cat.name)}, slug: '${cat.slug}', parent_id: '${storeCatId}', level: 2, sort_order: ${order++}, children: [] });\n`;
    }

    // Write file
    fs.writeFileSync('/tmp/seed-mongo-v2.js', script);
    console.log(`\nSeed script written to /tmp/seed-mongo-v2.js (${(script.length/1024).toFixed(1)} KB, ${detailed.length} products)`);

    // Execute
    console.log('Executing MongoDB seed...');
    try {
      const result = execSync('docker compose exec -T mongodb mongosh tiki_catalog < /tmp/seed-mongo-v2.js', {
        cwd: process.env.HOME || '/home/datdt',
        encoding: 'utf8',
        timeout: 60000,
        shell: '/bin/bash',
      });
      console.log('Seed completed successfully!');
    } catch (err) {
      console.error('MongoDB seed error:', err.message.substring(0, 200));
      console.log('Run manually: docker compose exec -T mongodb mongosh tiki_catalog < /tmp/seed-mongo-v2.js');
    }

    console.log(`\nDone!`);

  } finally {
    await browser.close();
  }
}

main().catch(err => { console.error(err); process.exit(1); });
