#!/usr/bin/env python3
"""
Tiki.vn Product Data Extractor & Database Inserter
Two modes:
  1. extract: Uses browser cookies to fetch Tiki pages via requests
  2. insert: Reads JSON product data and inserts into MySQL

Since Tiki.vn blocks direct HTTP requests (403), we use a hybrid approach:
- Extract product data via browser automation (JavaScript evaluation)
- Save to JSON file
- Insert from JSON file into database
"""

import json
import uuid
import re
import logging
import argparse
from datetime import datetime

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger("tiki_processor")


# ============================================================
# Data Processing
# ============================================================

def generate_uuid():
    return str(uuid.uuid4())


def parse_price_vnd(text):
    """Parse Vietnamese price text to integer VND."""
    if not text:
        return None
    cleaned = re.sub(r"[₫đ\s]", "", str(text).strip())
    cleaned = cleaned.replace(".", "").replace(",", "")
    try:
        return int(cleaned)
    except ValueError:
        return None


def parse_sold_count(text):
    """Parse 'Đã bán 263' or 'Đã bán 1.3k' to integer."""
    if not text:
        return None, None
    raw = str(text).strip()
    match = re.search(r"Đã bán\s+(.+)", raw)
    if not match:
        return None, raw
    count_str = match.group(1).strip()
    if "k" in count_str.lower():
        try:
            num = float(count_str.lower().replace("k", "")), raw
            return int(num[0] * 1000), raw
        except ValueError:
            return None, raw
    count_str_clean = count_str.replace(".", "").replace(",", "")
    try:
        return int(count_str_clean), raw
    except ValueError:
        return None, raw


def process_raw_products(raw_products, category_mapping=None):
    """Process raw product data extracted from browser into DB-ready format."""
    processed = []

    for raw in raw_products:
        try:
            product = {
                "id": generate_uuid(),
                "tiki_product_id": str(raw.get("tiki_product_id", raw.get("id", ""))),
                "category_id": raw.get("category_id"),
                "category_name": raw.get("category_name", ""),
                "name": raw.get("name", raw.get("title", ""))[:500],
                "url": raw.get("url", ""),
                "image_url": raw.get("image_url", ""),
                "thumbnail_url": raw.get("thumbnail_url", raw.get("image_url", "")),
                "brand": raw.get("brand"),
                "price": raw.get("price", 0) or 0,
                "original_price": raw.get("original_price"),
                "discount_percent": raw.get("discount_percent"),
                "rating_average": raw.get("rating_average"),
                "rating_count": raw.get("rating_count"),
                "review_count": raw.get("review_count"),
                "sold_count": raw.get("sold_count"),
                "quantity_sold_text": raw.get("quantity_sold_text"),
                "seller_name": raw.get("seller_name"),
                "seller_avatar_url": raw.get("seller_avatar_url"),
                "is_tiki_trading": raw.get("is_tiki_trading", False),
                "is_official": raw.get("is_official", False),
                "is_sponsored": raw.get("is_sponsored", False),
                "badge_text": raw.get("badge_text"),
                "shipping_info": raw.get("shipping_info"),
                "freeship": raw.get("freeship", False),
                "installment": raw.get("installment", False),
                "status": raw.get("status", "active"),
                "crawl_page_num": raw.get("crawl_page_num"),
            }
            if product["name"] and product["tiki_product_id"]:
                processed.append(product)
        except Exception as e:
            logger.debug("Error processing product: %s", e)

    return processed


# ============================================================
# Database Operations
# ============================================================

def get_db(args):
    """Create database connection."""
    try:
        import pymysql
        conn = pymysql.connect(
            host=args.host, port=args.port, user=args.user,
            password=args.password, database=args.database,
            charset="utf8mb4", cursorclass=pymysql.cursors.DictCursor,
            autocommit=False,
        )
    except ImportError:
        import mysql.connector
        conn = mysql.connector.connect(
            host=args.host, port=args.port, user=args.user,
            password=args.password, database=args.database,
            charset="utf8mb4", autocommit=False,
        )
    return conn


def save_products_to_db(conn, products):
    """Insert products into tiki_products table."""
    if not products:
        return 0

    sql = """INSERT INTO tiki_products
        (id, tiki_product_id, category_id, category_name, name, url, image_url,
         thumbnail_url, brand, price, original_price, discount_percent,
         rating_average, rating_count, review_count, sold_count, quantity_sold_text,
         seller_name, seller_avatar_url, is_tiki_trading, is_official, is_sponsored,
         badge_text, shipping_info, freeship, installment, status, crawl_page_num)
        VALUES
        (%(id)s, %(tiki_product_id)s, %(category_id)s, %(category_name)s, %(name)s,
         %(url)s, %(image_url)s, %(thumbnail_url)s, %(brand)s, %(price)s,
         %(original_price)s, %(discount_percent)s, %(rating_average)s, %(rating_count)s,
         %(review_count)s, %(sold_count)s, %(quantity_sold_text)s, %(seller_name)s,
         %(seller_avatar_url)s, %(is_tiki_trading)s, %(is_official)s, %(is_sponsored)s,
         %(badge_text)s, %(shipping_info)s, %(freeship)s, %(installment)s, %(status)s,
         %(crawl_page_num)s)
        ON DUPLICATE KEY UPDATE
            name=VALUES(name), price=VALUES(price), original_price=VALUES(original_price),
            discount_percent=VALUES(discount_percent), sold_count=VALUES(sold_count),
            rating_average=VALUES(rating_average), status=VALUES(status),
            updated_at=NOW()"""

    cursor = conn.cursor()
    inserted = 0
    batch_size = 50

    for i in range(0, len(products), batch_size):
        batch = products[i:i + batch_size]
        try:
            cursor.executemany(sql, batch)
            conn.commit()
            inserted += len(batch)
            logger.info("Saved batch %d-%d", i + 1, min(i + batch_size, len(products)))
        except Exception as e:
            conn.rollback()
            logger.error("Batch error: %s", e)
            for p in batch:
                try:
                    cursor.execute(sql, p)
                    conn.commit()
                    inserted += 1
                except Exception as e2:
                    logger.error("Product %s error: %s", p.get("tiki_product_id"), e2)

    cursor.close()
    return inserted


def save_categories_to_db(conn, categories):
    """Insert categories into tiki_categories table."""
    cursor = conn.cursor()
    sql = """INSERT INTO tiki_categories (id, tiki_category_id, name, slug, url_path)
             VALUES (%s, %s, %s, %s, %s)
             ON DUPLICATE KEY UPDATE name=VALUES(name), updated_at=NOW()"""

    for cat in categories:
        try:
            cursor.execute(sql, (cat["id"], cat["tiki_category_id"], cat["name"], cat["slug"], cat["url_path"]))
        except Exception as e:
            logger.error("Category error: %s", e)

    conn.commit()
    cursor.close()


# ============================================================
# Main
# ============================================================

def main():
    parser = argparse.ArgumentParser(description="Tiki Product Data Processor")
    parser.add_argument("mode", choices=["insert", "stats", "export-sql"],
                        help="Mode: insert from JSON, show stats, or export SQL seed")
    parser.add_argument("--input", "-i", help="Input JSON file with product data")
    parser.add_argument("--output", "-o", help="Output file")
    parser.add_argument("--host", default="127.0.0.1", help="MySQL host")
    parser.add_argument("--port", type=int, default=3306, help="MySQL port")
    parser.add_argument("--user", default="tiki", help="MySQL user")
    parser.add_argument("--password", default="tiki_dev", help="MySQL password")
    parser.add_argument("--database", default="tiki_platform", help="MySQL database")
    args = parser.parse_args()

    conn = get_db(args)

    try:
        if args.mode == "insert":
            if not args.input:
                logger.error("--input JSON file required for insert mode")
                return

            with open(args.input, "r", encoding="utf-8") as f:
                data = json.load(f)

            raw_products = data if isinstance(data, list) else data.get("products", [])
            logger.info("Loaded %d products from %s", len(raw_products), args.input)

            products = process_raw_products(raw_products)
            inserted = save_products_to_db(conn, products)
            logger.info("Inserted/updated %d products in database", inserted)

        elif args.mode == "stats":
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) as total FROM tiki_products")
            total = cursor.fetchone()
            if isinstance(total, dict):
                total = total.get("total", 0)
            else:
                total = total[0]

            cursor.execute("SELECT COUNT(*) as total FROM tiki_categories")
            cats = cursor.fetchone()
            if isinstance(cats, dict):
                cats = cats.get("total", 0)
            else:
                cats = cats[0]

            cursor.execute("SELECT category_name, COUNT(*) as cnt FROM tiki_products GROUP BY category_name ORDER BY cnt DESC LIMIT 10")
            by_cat = cursor.fetchall()

            logger.info("=== Database Stats ===")
            logger.info("Total products: %d", total)
            logger.info("Total categories: %d", cats)
            logger.info("Top categories:")
            for row in by_cat:
                name = row["category_name"] if isinstance(row, dict) else row[0]
                cnt = row["cnt"] if isinstance(row, dict) else row[1]
                logger.info("  %s: %d products", name, cnt)
            cursor.close()

        elif args.mode == "export-sql":
            cursor = conn.cursor()
            cursor.execute("SELECT * FROM tiki_products LIMIT 100")
            rows = cursor.fetchall()
            logger.info("Exported %d products", len(rows))
            if args.output:
                with open(args.output, "w", encoding="utf-8") as f:
                    json.dump(rows, f, ensure_ascii=False, indent=2, default=str)
                logger.info("Saved to %s", args.output)
            cursor.close()

    finally:
        conn.close()


if __name__ == "__main__":
    main()
