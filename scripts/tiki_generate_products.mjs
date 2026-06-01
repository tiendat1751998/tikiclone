/**
 * Generate supplementary products from existing crawled data
 * Creates variations to reach 20,000+ total products
 * Real Tiki data + realistic variations for test purposes
 */

import { v4 as uuidv4 } from 'uuid';
import fs from 'fs';
import { execSync } from 'child_process';
import https from 'https';
import http from 'http';
import path from 'path';
import { URL } from 'url';

const IMAGE_DIR = '/home/datdt/tikiclone/public/images/products';
const PRODUCTS_FILE = '/tmp/crawled_products_v3.json';
const CATEGORIES_FILE = '/tmp/crawled_categories_v3.json';

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }
function sanitize(s) { return (s || '').replace(/'/g, "''").trim(); }

function localImgDest(imageUrl, productId) {
  try {
    const ext = (path.extname(new URL(imageUrl).pathname).split('?')[0] || '.jpg').toLowerCase();
    return path.join(IMAGE_DIR, `${productId}${ext}`);
  } catch { return path.join(IMAGE_DIR, `${productId}.jpg`); }
}
function localImgUrl(imageUrl, productId) {
  return `/images/products/${path.basename(localImgDest(imageUrl, productId))}`;
}

async function downloadImage(imageUrl, destPath, retries = 1) {
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
          timeout: 10000,
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

function generateVariations(existingProducts, targetCount) {
  const variations = [];
  const suffixes = [
    ' - Bản 2025', ' - Bản nâng cấp', ' - Phiên bản Pro',
    ' - Phiên bản Lite', ' - Bản giới hạn', ' - Bản Premium',
    ' - New', ' - Hot', ' - Plus', ' - Max',
  ];
  const colors = [
    'Đen', 'Trắng', 'Xanh', 'Đỏ', 'Vàng', 'Hồng', 'Tím',
    'Xám', 'Bạc', 'Xanh Navy', ' Đồng', 'Xanh Olive',
  ];
  const specs = [
    '64GB', '128GB', '256GB', '512GB', '1TB',
    '4GB/64GB', '6GB/128GB', '8GB/256GB', '12GB/512GB',
  ];

  let idx = 0;
  for (const p of existingProducts) {
    if (variations.length >= targetCount) break;
    
    // Create 1 variation per product (randomly selected)
    if (idx % 3 !== 0) { idx++; continue; }
    idx++;
    
    const suffix = suffixes[Math.floor(Math.random() * suffixes.length)];
    const color = colors[Math.floor(Math.random() * colors.length)];
    const spec = specs[Math.floor(Math.random() * specs.length)];
    
    const priceVariation = p.price > 0 ? Math.round(p.price * (0.7 + Math.random() * 0.8)) : Math.round(100000 + Math.random() * 5000000);
    
    const newPid = 'gen-' + uuidv4().slice(0, 12);
    
    // Reuse existing product's image
    const existingImage = p.images && p.images.length > 0 ? p.images[0] : '';
    const isLocal = existingImage.startsWith('/images/');
    
    variations.push({
      tiki_product_id: newPid,
      name: `${p.name}${suffix} ${color} ${spec}`.substring(0, 500),
      description: (p.description || '') + ` Phiên bản ${color}, ${spec}.`.substring(0, 2000),
      price: priceVariation,
      original_price: Math.round(priceVariation * (1.1 + Math.random() * 0.3)),
      discount_percent: Math.floor(Math.random() * 30) + 5,
      brand: p.brand || '',
      images: isLocal ? [existingImage] : (p.images || []),
      rating_average: p.rating_average ? Math.min(5, Math.max(1, p.rating_average + (Math.random() - 0.5))) : (3 + Math.random() * 2),
      review_count: Math.floor(Math.random() * 500),
      sold_count: Math.floor(Math.random() * 2000),
      seller_name: p.seller_name || '',
      category_id: p.category_id,
      category_name: p.category_name,
      category_slug: p.category_slug,
      url: p.url,
      db_id: uuidv4(),
      category_db_id: p.category_db_id,
      local_image_url: isLocal ? existingImage : '',
      image_url: isLocal ? existingImage : (p.images?.[0] || ''),
    });
  }

  return variations;
}

async function main() {
  console.log('=== Generating Supplementary Products ===');
  
  let existingProducts = [];
  let existingCategories = [];
  try { existingProducts = JSON.parse(fs.readFileSync(PRODUCTS_FILE, 'utf8')); } catch (e) {}
  try { existingCategories = JSON.parse(fs.readFileSync(CATEGORIES_FILE, 'utf8')); } catch (e) {}
  
  console.log(`Existing: ${existingProducts.length} products`);
  
  const targetNew = 22000 - existingProducts.length;
  console.log(`Generating ${targetNew} additional products...`);
  
  const newProducts = generateVariations(existingProducts, targetNew);
  
  console.log(`Generated ${newProducts.length} products`);
  
  // Download images for products that don't have local images
  console.log('Downloading images for new products...');
  let dlOk = 0, dlFail = 0;
  for (let i = 0; i < newProducts.length; i++) {
    const p = newProducts[i];
    if (p.local_image_url && p.local_image_url.startsWith('/images/')) {
      continue; // Already local
    }
    if (!p.images || p.images.length === 0) continue;
    
    const imgUrl = p.images[0];
    if (imgUrl.startsWith('/images/')) {
      p.local_image_url = imgUrl;
      p.image_url = imgUrl;
      continue;
    }
    
    // Need to download
    const newTikiId = 'genimg-' + uuidv4().slice(0, 10);
    const dest = localImgDest(imgUrl, newTikiId);
    const ok = await downloadImage(imgUrl, dest);
    if (ok) {
      p.local_image_url = localImgUrl(imgUrl, newTikiId);
      p.image_url = p.local_image_url;
      dlOk++;
    } else {
      p.local_image_url = '';
      p.image_url = imgUrl;
      dlFail++;
    }
    if ((i + 1) % 500 === 0) console.log(`  ${i + 1}/${newImages.length}`);
    if (i % 100 === 0) sleep(100);
  }
  console.log(`Images: ${dlOk} ok, ${dlFail} fail`);
  
  // Merge into existing
  const allProducts = [...existingProducts, ...newProducts];
  
  // Save to MySQL
  console.log('=== Saving to MySQL ===');
  await saveToMySQL(newProducts);
  
  // Save to MongoDB
  console.log('=== Saving to MongoDB ===');
  await saveToMongoDB(newProducts);
  
  // Save files
  fs.writeFileSync(PRODUCTS_FILE, JSON.stringify(allProducts));
  fs.writeFileSync(CATEGORIES_FILE, JSON.stringify(existingCategories));
  
  console.log(`\n========== DONE ==========`);
  console.log(`New products: ${newProducts.length}`);
  console.log(`Total products: ${allProducts.length}`);
  console.log(`Total categories: ${existingCategories.length}`);
  
  // Verify
  try {
    const mc = execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -N -e "SELECT COUNT(*) FROM tiki_products" 2>/dev/null`, { timeout: 10000, encoding: 'utf8' });
    const moc = execSync(`docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog --quiet --eval "db.products.countDocuments()" 2>/dev/null`, { timeout: 15000, encoding: 'utf8' });
    console.log(`MySQL: ${mc.trim()} products`);
    console.log(`MongoDB: ${moc.trim()} products`);
  } catch (e) { /* */ }
}

async function saveToMySQL(newProducts) {
  const batchSize = 100;
  let inserted = 0;
  for (let i = 0; i < newProducts.length; i += batchSize) {
    const batch = newProducts.slice(i, i + batchSize);
    const values = batch.map(p =>
      `('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average ? p.rating_average.toFixed(2) : 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`
    ).join(',');
    const sql = `INSERT IGNORE INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ${values}`;
    try {
      execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${sql.replace(/"/g, '\\"')}" 2>/dev/null`, { timeout: 60000, encoding: 'utf8' });
      inserted += batch.length;
    } catch (e) {
      for (const p of batch) {
        try {
          const s = `INSERT IGNORE INTO tiki_products (id,tiki_product_id,category_id,category_name,category_slug,name,description,brand,price,original_price,discount_percent,rating_average,review_count,sold_count,seller_name,image_url,local_image_url,images,url,status) VALUES ('${p.db_id}','${p.tiki_product_id}','${p.category_db_id}','${sanitize(p.category_name)}','${sanitize(p.category_slug)}','${sanitize(p.name)}','${sanitize(p.description)}','${sanitize(p.brand)}',${p.price},${p.original_price || 'NULL'},${p.discount_percent || 'NULL'},${p.rating_average ? p.rating_average.toFixed(2) : 'NULL'},${p.review_count || 'NULL'},${p.sold_count || 'NULL'},'${sanitize(p.seller_name)}','${sanitize(p.image_url || '')}','${sanitize(p.local_image_url || '')}','${JSON.stringify(p.images || []).replace(/'/g, "''")}','${sanitize(p.url)}','active')`;
          execSync(`docker compose -p tikiclone exec -T mysql-primary mysql -utiki -ptiki_dev tiki_platform -e "${s.replace(/"/g, '\\"')}" 2>/dev/null`, { timeout: 10000, encoding: 'utf8' });
          inserted++;
        } catch (e2) { /* skip */ }
      }
    }
    if ((i + batchSize) % 1000 === 0 || i + batchSize >= newProducts.length) {
      console.log(`  ${inserted}/${newProducts.length}`);
    }
  }
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
      script += `db.products.insertOne({spu_id:'${p.db_id}',title:${JSON.stringify(p.name)},description:${JSON.stringify(p.description || '')},category_id:'${p.category_db_id}',seller_id:'usr-002',status:'ACTIVE',attributes:${attrs},images:${imgs},local_image_url:${JSON.stringify(p.local_image_url || '')},skus:[{sku_id:'${skuId}',spu_id:'${p.db_id}',price:${p.price},stock:${Math.floor(Math.random()*200)+10},status:'ACTIVE',variations:[{name:'Mặc định',value:'default'}]}],rating_average:${p.rating_average ? p.rating_average.toFixed(2) : 'null'},review_count:${p.review_count || 0},sold_count:${p.sold_count || 0},tiki_product_id:'${p.tiki_product_id}',created_at:ISODate('${now}'),updated_at:ISODate('${now}')});\n`;
    }
    fs.writeFileSync(`/tmp/seed_mongo_gen_chunk_${Math.floor(i/chunkSize)}.js`, script);
    try {
      execSync(`docker compose -p tikiclone exec -T mongodb mongosh tiki_catalog < /tmp/seed_mongo_gen_chunk_${Math.floor(i/chunkSize)}.js 2>/dev/null`, { timeout: 120000, encoding: 'utf8' });
      console.log(`  Chunk ${Math.floor(i/chunkSize)} OK`);
    } catch (e) { console.log(`  Chunk ${Math.floor(i/chunkSize)} error`); }
  }
}

main().catch(err => {
  console.error('FATAL:', err);
  process.exit(1);
});
