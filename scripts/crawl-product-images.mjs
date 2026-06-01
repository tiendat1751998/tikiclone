import { chromium } from 'playwright';
import mysql from 'mysql2/promise';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const DB_CONFIG = {
  host: process.env.MYSQL_HOST || 'localhost',
  port: parseInt(process.env.MYSQL_PORT || '3306', 10),
  user: process.env.MYSQL_USER || 'tiki',
  password: process.env.MYSQL_PASSWORD || 'tiki_dev',
  database: process.env.MYSQL_DATABASE || 'tiki_platform',
};

const IMAGE_DIR = path.resolve(__dirname, '..', 'apps', 'web', 'public', 'images', 'products');
const TIKI_SEARCH_URL = 'https://tiki.vn/search?q=';
const MAX_PRODUCTS = 20; // limit per run to avoid getting blocked
const DELAY_MS = 2000;

if (!fs.existsSync(IMAGE_DIR)) {
  fs.mkdirSync(IMAGE_DIR, { recursive: true });
}

function slugify(name) {
  return name
    .toLowerCase()
    .replace(/[àáảãạăắằẵặâấầẩẫậ]/g, 'a')
    .replace(/[đ]/g, 'd')
    .replace(/[èéẻẽẹêếềểễệ]/g, 'e')
    .replace(/[ìíỉĩị]/g, 'i')
    .replace(/[òóỏõọôốồổỗộơớờởỡợ]/g, 'o')
    .replace(/[ùúủũụưứừửữự]/g, 'u')
    .replace(/[ýỳỷỹỵ]/g, 'y')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .substring(0, 80);
}

async function downloadImage(url, filePath) {
  try {
    const response = await fetch(url);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const buffer = Buffer.from(await response.arrayBuffer());
    fs.writeFileSync(filePath, buffer);
    return true;
  } catch (err) {
    console.error(`    Failed to download ${url}: ${err.message}`);
    return false;
  }
}

async function getFirstImageForProduct(page, productName) {
  const searchUrl = `${TIKI_SEARCH_URL}${encodeURIComponent(productName)}`;
  console.log(`  Searching: ${searchUrl}`);

  try {
    await page.goto(searchUrl, { waitUntil: 'domcontentloaded', timeout: 30000 });
    await page.waitForTimeout(3000);

    // Try to get the first product image from search results
    const imageUrl = await page.evaluate(() => {
      // Tiki search page selectors
      const selectors = [
        'a[data-view-id*="product"] img',
        '.product-item img',
        '[class*="product"] img',
        'div[class*="search"] img[src*="tikicdn"]',
        'img[src*="tikicdn"][loading="lazy"]',
      ];

      for (const sel of selectors) {
        const img = document.querySelector(sel);
        if (img) {
          const src = img.getAttribute('src') || img.getAttribute('data-src') || '';
          // Get the original size (largest) by removing cache size
          if (src) return src.replace(/\/cache\/\d+x\d+/, '');
        }
      }

      // Fallback: get any img with tikicdn in src
      const allImgs = document.querySelectorAll('img');
      for (const img of allImgs) {
        const src = img.src || '';
        if (src.includes('tikicdn') && src.includes('/ts/')) {
          return src.replace(/\/cache\/\d+x\d+/, '');
        }
      }
      return '';
    });

    if (imageUrl) {
      console.log(`  Found image: ${imageUrl.substring(0, 100)}`);
    } else {
      console.log(`  No image found`);
    }
    return imageUrl;
  } catch (err) {
    console.error(`  Search error: ${err.message}`);
    return '';
  }
}

async function main() {
  console.log('=== Tiki Product Image Crawler ===\n');

  const conn = await mysql.createConnection(DB_CONFIG);

  try {
    // Get products with placeholder images (limited set)
    const [rows] = await conn.execute(`
      SELECT DISTINCT p.id, p.name
      FROM products p
      JOIN product_media pm ON pm.product_id = p.id
      WHERE pm.url = '/images/products/default-product.png'
        AND p.status = 'active'
      LIMIT ?
    `, [MAX_PRODUCTS]);

    console.log(`Found ${rows.length} products to process\n`);

    if (rows.length === 0) {
      console.log('No products need image updates.');
      return;
    }

    const browser = await chromium.launch({ headless: true, args: ['--no-sandbox'] });
    const page = await browser.newPage({
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36',
    });

    let updated = 0;
    let failed = 0;

    for (const product of rows) {
      console.log(`\n[${updated + failed + 1}/${rows.length}] ${product.id}: ${product.name}`);

      const imageUrl = await getFirstImageForProduct(page, product.name);
      if (!imageUrl) {
        failed++;
        continue;
      }

      // Download image
      const ext = path.extname(new URL(imageUrl).pathname) || '.jpg';
      const filename = `${product.id}${ext}`;
      const filePath = path.join(IMAGE_DIR, filename);
      const publicPath = `/images/products/${filename}`;

      const downloaded = await downloadImage(imageUrl, filePath);
      if (!downloaded) {
        failed++;
        continue;
      }

      console.log(`  Saved: ${publicPath}`);

      // Update database
      await conn.execute(
        'UPDATE product_media SET url = ?, thumbnail_url = ? WHERE product_id = ?',
        [publicPath, publicPath, product.id]
      );

      updated++;
      console.log(`  DB updated`);

      if (updated + failed < rows.length) {
        console.log(`  Waiting ${DELAY_MS}ms...`);
        await new Promise(r => setTimeout(r, DELAY_MS));
      }
    }

    console.log(`\nDone! Updated: ${updated}, Failed: ${failed}`);
  } finally {
    await conn.end();
  }
}

main().catch(err => { console.error(err); process.exit(1); });
