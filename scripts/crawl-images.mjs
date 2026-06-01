#!/usr/bin/env node
/**
 * Crawl product images from salt.tikicdn.com and save locally.
 * Reads products from the local API, downloads media URLs to public/images/products/
 */
import fs from "fs";
import path from "path";
import { execSync } from "child_process";

const GATEWAY_URL = process.env.GATEWAY_URL || "http://localhost:8080";
const PUBLIC_DIR = path.resolve(process.cwd(), "apps/web/public");
const IMAGES_DIR = path.join(PUBLIC_DIR, "images", "products");
const CATEGORY_IMAGES_DIR = path.join(PUBLIC_DIR, "images", "categories");
const PLACEHOLDER_DIR = path.join(PUBLIC_DIR, "images");

const API_BASE = `${GATEWAY_URL}/api/v1`;

// Ensure directories exist
for (const dir of [IMAGES_DIR, CATEGORY_IMAGES_DIR, PLACEHOLDER_DIR]) {
  fs.mkdirSync(dir, { recursive: true });
}

// Create placeholder SVG if it doesn't exist
const placeholderPath = path.join(PLACEHOLDER_DIR, "placeholder.svg");
if (!fs.existsSync(placeholderPath)) {
  fs.writeFileSync(placeholderPath, `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="400" viewBox="0 0 400 400">
  <rect fill="#F5F5FA" width="400" height="400"/>
  <text fill="#808089" font-family="sans-serif" font-size="16" text-anchor="middle" x="200" y="200">No Image</text>
</svg>`);
  console.log("Created placeholder.svg");
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function downloadImage(url, filepath, retries = 3) {
  if (fs.existsSync(filepath) && fs.statSync(filepath).size > 1000) {
    return true; // already downloaded
  }
  for (let attempt = 1; attempt <= retries; attempt++) {
    try {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 15000);
      const res = await fetch(url, {
        signal: controller.signal,
        headers: {
          "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
          "Referer": "https://tiki.vn/",
        },
      });
      clearTimeout(timeout);
      if (!res.ok) {
        console.warn(`  HTTP ${res.status} for ${url.split("?").shift()?.slice(-30)}`);
        return false;
      }
      const buffer = Buffer.from(await res.arrayBuffer());
      fs.writeFileSync(filepath, buffer);
      const size = (buffer.length / 1024).toFixed(1);
      console.log(`  Downloaded (${size} KB): ${path.basename(filepath)}`);
      return true;
    } catch (err) {
      if (attempt < retries) {
        const wait = attempt * 2000;
        console.warn(`  Retry ${attempt}/${retries} in ${wait}ms: ${err.message?.slice(0, 60)}`);
        await sleep(wait);
      } else {
        console.warn(`  Failed after ${retries} retries: ${err.message?.slice(0, 80)}`);
        return false;
      }
    }
  }
  return false;
}

async function crawlProducts() {
  console.log("\n=== Crawling product images ===\n");
  let page = 1;
  let totalDownloaded = 0;
  let totalSkipped = 0;
  let totalFailed = 0;

  while (true) {
    console.log(`Fetching page ${page}...`);
    let products;
    try {
      const res = await fetch(`${API_BASE}/products?page=${page}&size=50`, {
        headers: { "User-Agent": "tikiclone-crawler/1.0" },
      });
      const data = await res.json();
      products = data.products || data.data?.products || [];
      if (!Array.isArray(products) || products.length === 0) break;
    } catch (err) {
      console.error(`Failed to fetch page ${page}:`, err.message);
      await sleep(3000);
      page++;
      continue;
    }

    for (const p of products) {
      const productId = p.id;
      const mediaItems = p.media || [];
      let downloaded = false;

      for (const m of mediaItems) {
        if (m.type !== "image") continue;
        const url = m.url || m.thumbnail_url;
        if (!url) continue;

        // Determine extension from URL
        const urlPath = new URL(url).pathname;
        const ext = path.extname(urlPath).split("?")[0] || ".jpg";
        const filename = `${productId}${ext}`;
        const filepath = path.join(IMAGES_DIR, filename);

        const ok = await downloadImage(url, filepath);
        if (ok) {
          downloaded = true;
          totalDownloaded++;
          break; // download only primary image
        }
      }

      if (!downloaded) {
        // Try without cache query params
        for (const m of mediaItems) {
          if (m.type !== "image") continue;
          const rawUrl = m.url?.split("?")[0] || m.thumbnail_url?.split("?")[0];
          if (!rawUrl) continue;
          const ext = path.extname(new URL(rawUrl).pathname) || ".jpg";
          const filename = `${productId}${ext}`;
          const filepath = path.join(IMAGES_DIR, filename);

          if (fs.existsSync(filepath) && fs.statSync(filepath).size > 1000) {
            totalSkipped++;
            downloaded = true;
            break;
          }

          const ok = await downloadImage(rawUrl, filepath);
          if (ok) {
            downloaded = true;
            totalDownloaded++;
            break;
          }
        }
      }

      if (!downloaded) {
        totalFailed++;
      }

      // Rate limiting - be nice to tiki CDN
      await sleep(500 + Math.random() * 500);
    }

    console.log(`  Page ${page} done. Total: ${totalDownloaded} OK, ${totalSkipped} cached, ${totalFailed} failed`);
    page++;
  }

  console.log(`\n=== Product images complete: ${totalDownloaded} downloaded, ${totalSkipped} cached, ${totalFailed} failed ===`);
}

async function crawlCategories() {
  console.log("\n=== Crawling category images ===\n");
  try {
    const res = await fetch(`${API_BASE}/categories/tree`, {
      headers: { "User-Agent": "tikiclone-crawler/1.0" },
    });
    const data = await res.json();
    const categories = data.data || data || [];
    const flat = [];
    function flatten(cats) {
      for (const c of cats) {
        flat.push(c);
        if (c.children) flatten(c.children);
      }
    }
    flatten(Array.isArray(categories) ? categories : [categories]);

    let downloaded = 0;
    for (const cat of flat) {
      if (!cat.image_url) continue;
      const filename = `${cat.slug || cat.id}.jpg`;
      const filepath = path.join(CATEGORY_IMAGES_DIR, filename);
      const ok = await downloadImage(cat.image_url, filepath);
      if (ok) downloaded++;
      await sleep(300);
    }
    console.log(`\n=== Category images complete: ${downloaded} downloaded ===`);
  } catch (err) {
    console.error("Failed to crawl category images:", err.message);
  }
}

async function main() {
  console.log("=== Tiki Image Crawler ===\n");
  await crawlProducts();
  await crawlCategories();
  console.log("\n=== All done ===");
}

main().catch(console.error);
