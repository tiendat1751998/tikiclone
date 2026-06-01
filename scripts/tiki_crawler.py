#!/usr/bin/env python3
"""
Tiki.vn Product Crawler
Scrapes product data from Tiki.vn category pages and stores in MySQL database.
"""

import re
import time
import uuid
import random
import logging
import argparse
from datetime import datetime
from urllib.parse import urljoin, urlparse, parse_qs

import requests
from bs4 import BeautifulSoup

# ============================================================
# Configuration
# ============================================================

TIKI_BASE_URL = "https://tiki.vn"

# Category URLs to crawl (slug, category_id, name)
DEFAULT_CATEGORIES = [
    ("dien-thoai-may-tinh-bang", "1789", "Điện Thoại - Máy Tính Bảng"),
    ("laptop-may-vi-tinh-linh-kien", "1846", "Laptop - Máy Vi Tính - Linh Kiện"),
    ("dien-gia-dung", "1882", "Điện Gia Dụng"),
    ("nha-cua-doi-song", "1883", "Nhà Cửa - Đời Sống"),
    ("lam-dep-suc-khoe", "1520", "Làm Đẹp - Sức Khỏe"),
    ("me-be", "2549", "Mẹ & Bé"),
    ("do-choi-me-be", "2549", "Đồ Chơi - Mẹ & Bé"),
    ("thoi-trang-nu", "915", "Thời Trang Nữ"),
    ("thoi-trang-nam", "931", "Thời Trang Nam"),
    ("giay-dep-nam", "1686", "Giày - Dép Nam"),
    ("giay-dep-nu", "1703", "Giày - Dép Nữ"),
    ("dien-tu-dien-lanh", "4221", "Điện Tử - Điện Lạnh"),
    ("nha-sach-tiki", "8322", "Nhà Sách Tiki"),
    ("the-thao-da-ngoai", "1975", "Thể Thao - Dã Ngoại"),
    ("o-to-xe-may-xe-dap", "8594", "Ô Tô - Xe Máy - Xe Đạp"),
    ("bach-hoa-online", "4384", "Bách Hóa Online"),
    ("dong-ho-va-trang-suc", "27497", "Đồng Hồ và Trang Sức"),
    ("balo-va-vali", "6000", "Balo và Vali"),
    ("phu-kien-thoi-trang", "27498", "Phụ Kiện Thời Trang"),
    ("tui-thoi-trang-nu", "914", "Túi Thời Trang Nữ"),
]

# User agents to rotate
USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
]

# Logging setup
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger("tiki_crawler")


# ============================================================
# Database helper
# ============================================================

class Database:
    """MySQL database helper using mysql-connector-python or pymysql."""

    def __init__(self, host="mysql-primary", port=3306, user="tiki",
                 password="tiki_dev", database="tiki_platform"):
        self.host = host
        self.port = port
        self.user = user
        self.password = password
        self.database = database
        self.conn = None

    def connect(self):
        """Connect to MySQL database."""
        try:
            import pymysql
            self.conn = pymysql.connect(
                host=self.host, port=self.port, user=self.user,
                password=self.password, database=self.database,
                charset="utf8mb4", cursorclass=pymysql.cursors.DictCursor,
                autocommit=False,
            )
            logger.info("Connected to MySQL database %s", self.database)
        except ImportError:
            import mysql.connector
            self.conn = mysql.connector.connect(
                host=self.host, port=self.port, user=self.user,
                password=self.password, database=self.database,
                charset="utf8mb4", autocommit=False,
            )
            logger.info("Connected to MySQL database %s (mysql-connector)", self.database)

    def execute(self, sql, params=None):
        """Execute a single SQL statement."""
        cursor = self.conn.cursor()
        try:
            cursor.execute(sql, params)
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise
        finally:
            cursor.close()

    def executemany(self, sql, params_list):
        """Execute a SQL statement with multiple parameter sets."""
        cursor = self.conn.cursor()
        try:
            cursor.executemany(sql, params_list)
            self.conn.commit()
        except Exception:
            self.conn.rollback()
            raise
        finally:
            cursor.close()

    def fetchone(self, sql, params=None):
        """Fetch a single row."""
        cursor = self.conn.cursor()
        try:
            cursor.execute(sql, params)
            return cursor.fetchone()
        finally:
            cursor.close()

    def fetchall(self, sql, params=None):
        """Fetch all rows."""
        cursor = self.conn.cursor()
        try:
            cursor.execute(sql, params)
            return cursor.fetchall()
        finally:
            cursor.close()

    def close(self):
        if self.conn:
            self.conn.close()


# ============================================================
# Tiki Crawler
# ============================================================

class TikiCrawler:
    """Crawls product data from Tiki.vn category listing pages."""

    def __init__(self, db: Database, max_pages_per_category=5, delay_range=(1.5, 3.5)):
        self.db = db
        self.max_pages = max_pages_per_category
        self.delay_range = delay_range
        self.session = requests.Session()
        self.session.headers.update({
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
            "Accept-Language": "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7",
            "Accept-Encoding": "gzip, deflate, br",
            "Connection": "keep-alive",
            "Upgrade-Insecure-Requests": "1",
            "Sec-Fetch-Dest": "document",
            "Sec-Fetch-Mode": "navigate",
            "Sec-Fetch-Site": "none",
            "Sec-Fetch-User": "?1",
            "Cache-Control": "max-age=0",
        })

    def _rotate_user_agent(self):
        ua = random.choice(USER_AGENTS)
        self.session.headers["User-Agent"] = ua

    def _sleep(self):
        delay = random.uniform(*self.delay_range)
        time.sleep(delay)

    def _get_page(self, url: str) -> str | None:
        """Fetch a page and return HTML content."""
        self._rotate_user_agent()
        try:
            resp = self.session.get(url, timeout=30)
            if resp.status_code == 200:
                return resp.text
            elif resp.status_code == 404:
                logger.warning("Page not found: %s", url)
                return None
            elif resp.status_code == 429:
                logger.warning("Rate limited! Sleeping 30s...")
                time.sleep(30)
                return self._get_page(url)
            else:
                logger.warning("HTTP %d for %s", resp.status_code, url)
                return None
        except requests.RequestException as e:
            logger.error("Request failed for %s: %s", url, e)
            return None

    def _parse_price(self, text: str) -> int | None:
        """Parse Vietnamese price text like '11.990.000 ₫' to integer."""
        if not text:
            return None
        # Remove currency symbol and whitespace
        cleaned = re.sub(r"[₫\s]", "", text.strip())
        # Remove thousand separators (dots in Vietnamese format)
        cleaned = cleaned.replace(".", "")
        try:
            return int(cleaned)
        except ValueError:
            return None

    def _parse_sold_count(self, text: str) -> tuple[int | None, str | None]:
        """Parse 'Đã bán 263' or 'Đã bán 1.3k' to integer count."""
        if not text:
            return None, None
        raw = text.strip()
        match = re.search(r"Đã bán\s+(.+)", raw)
        if not match:
            return None, raw
        count_str = match.group(1).strip()
        # Handle '1.3k' format
        if "k" in count_str.lower():
            try:
                num = float(count_str.lower().replace("k", ""))
                return int(num * 1000), raw
            except ValueError:
                return None, raw
        # Handle plain number
        count_str = count_str.replace(".", "").replace(",", "")
        try:
            return int(count_str), raw
        except ValueError:
            return None, raw

    def _parse_discount(self, text: str) -> int | None:
        """Parse '-29%' to integer 29."""
        if not text:
            return None
        match = re.search(r"-(\d+)%", text.strip())
        if match:
            return int(match.group(1))
        return None

    def _extract_product_id_from_url(self, url: str) -> str | None:
        """Extract Tiki product ID from URL like ...-p278600678.html"""
        match = re.search(r"-p(\d+)\.html", url)
        if match:
            return match.group(1)
        return None

    def _count_star_rating(self, element) -> float | None:
        """Count filled stars from SVG star icons in rating element."""
        if not element:
            return None
        # Each star is an SVG with class "star-icon"
        stars = element.find_all("svg", class_="star-icon")
        return float(len(stars)) if stars else None

    def crawl_category_page(self, html: str, category_name: str, page_num: int) -> list[dict]:
        """Parse product cards from a category listing page HTML."""
        soup = BeautifulSoup(html, "html.parser")
        products = []

        # Tiki product cards are in divs with background-color: rgb(245, 245, 250)
        # Each card is an <a> tag with class starting with "sc-4eae099-0"
        product_links = soup.find_all("a", class_=re.compile(r"sc-4eae099-0"))

        for link in product_links:
            try:
                product = self._parse_product_card(link, category_name, page_num)
                if product and product.get("name"):
                    products.append(product)
            except Exception as e:
                logger.debug("Error parsing product card: %s", e)
                continue

        logger.info("Found %d products on page %d of %s", len(products), page_num, category_name)
        return products

    def _parse_product_card(self, link_element, category_name: str, page_num: int) -> dict | None:
        """Parse a single product card <a> element."""
        # Product URL
        href = link_element.get("href", "")
        if not href:
            return None
        product_url = urljoin(TIKI_BASE_URL, href)
        tiki_product_id = self._extract_product_id_from_url(product_url)
        if not tiki_product_id:
            return None

        # Image
        img = link_element.find("img", class_=re.compile(r"sc-4eae099-1"))
        image_url = img.get("src", "") if img else ""
        thumbnail_url = img.get("srcset", "").split(" ")[0] if img else image_url

        # Product name
        name_el = link_element.find("div", class_=re.compile(r"sc-4eae099-7"))
        name = name_el.get_text(strip=True) if name_el else ""
        if not name:
            return None

        # Seller info
        seller_el = link_element.find("div", class_=re.compile(r"sc-4eae099-6"))
        seller_name = seller_el.get_text(strip=True) if seller_el else None
        seller_avatar_el = link_element.find("img", class_=re.compile(r"sc-4eae099-5"))
        seller_avatar_url = seller_avatar_el.get("src", "") if seller_avatar_el else None
        is_tiki_trading = seller_name == "Tiki Trading" if seller_name else False

        # Sponsored badge
        sponsored_el = link_element.find("div", class_=re.compile(r"sc-4eae099-2"))
        is_sponsored = sponsored_el is not None and "Tài trợ" in sponsored_el.get_text()

        # Rating
        rating_el = link_element.find("div", class_=re.compile(r"sc-f5f45ccf-0"))
        rating_average = None
        if rating_el:
            stars_el = rating_el.find("p", class_=re.compile(r"sc-eed848ce-0"))
            if stars_el:
                rating_average = self._count_star_rating(stars_el)

        # Sold count
        sold_count = None
        sold_text = None
        if rating_el:
            sold_divs = rating_el.find_all("div")
            for div in sold_divs:
                text = div.get_text(strip=True)
                if "Đã bán" in text:
                    sold_count, sold_text = self._parse_sold_count(text)
                    break

        # Price
        price_el = link_element.find("div", class_=re.compile(r"sc-4eae099-9"))
        price = self._parse_price(price_el.get_text()) if price_el else 0

        # Discount
        discount_el = link_element.find("div", class_=re.compile(r"sc-4eae099-10"))
        discount_percent = self._parse_discount(discount_el.get_text()) if discount_el else None

        # Calculate original price from discount
        original_price = None
        if discount_percent and price and discount_percent > 0:
            original_price = int(price / (1 - discount_percent / 100))

        # "Xem thêm" link (more variants)
        xem_them = link_element.find("div", class_=re.compile(r"sc-aeef9a0f-5"))

        return {
            "id": str(uuid.uuid4()),
            "tiki_product_id": tiki_product_id,
            "category_name": category_name,
            "name": name,
            "url": product_url,
            "image_url": image_url,
            "thumbnail_url": thumbnail_url or image_url,
            "brand": None,  # Not always available on listing page
            "price": price or 0,
            "original_price": original_price,
            "discount_percent": discount_percent,
            "rating_average": rating_average,
            "rating_count": None,
            "review_count": None,
            "sold_count": sold_count,
            "quantity_sold_text": sold_text,
            "seller_name": seller_name,
            "seller_avatar_url": seller_avatar_url,
            "is_tiki_trading": is_tiki_trading,
            "is_official": is_tiki_trading,
            "is_sponsored": is_sponsored,
            "badge_text": None,
            "shipping_info": None,
            "freeship": False,
            "installment": False,
            "status": "active",
            "crawl_page_num": page_num,
        }

    def crawl_category(self, slug: str, tiki_category_id: str, name: str) -> list[dict]:
        """Crawl all pages of a category."""
        all_products = []
        crawl_job_id = str(uuid.uuid4())

        # Record crawl job
        category_url = f"{TIKI_BASE_URL}/{slug}/c{tiki_category_id}"
        self.db.execute(
            "INSERT INTO tiki_crawl_jobs (id, category_url, category_name, status, started_at) VALUES (%s, %s, %s, 'running', NOW())",
            (crawl_job_id, category_url, name),
        )

        # Store category
        cat_db_id = str(uuid.uuid4())
        self.db.execute(
            """INSERT INTO tiki_categories (id, tiki_category_id, name, slug, url_path)
               VALUES (%s, %s, %s, %s, %s)
               ON DUPLICATE KEY UPDATE name=VALUES(name), updated_at=NOW()""",
            (cat_db_id, tiki_category_id, name, slug, f"/{slug}/c{tiki_category_id}"),
        )

        for page in range(1, self.max_pages + 1):
            url = f"{TIKI_BASE_URL}/{slug}/c{tiki_category_id}?page={page}"
            logger.info("Crawling: %s", url)

            html = self._get_page(url)
            if not html:
                logger.warning("Failed to fetch page %d of %s", page, name)
                break

            products = self.crawl_category_page(html, name, page)
            if not products:
                logger.info("No more products found on page %d of %s", page, name)
                break

            # Update category_id in products
            for p in products:
                p["category_id"] = cat_db_id

            all_products.extend(products)
            self._sleep()

        # Update crawl job
        self.db.execute(
            """UPDATE tiki_crawl_jobs SET status='completed', products_found=%s,
               products_stored=%s, pages_crawled=%s, completed_at=NOW() WHERE id=%s""",
            (len(all_products), len(all_products), self.max_pages, crawl_job_id),
        )

        logger.info("Category %s: crawled %d products total", name, len(all_products))
        return all_products

    def save_products(self, products: list[dict]):
        """Save crawled products to database."""
        if not products:
            return

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

        # Batch insert
        batch_size = 50
        for i in range(0, len(products), batch_size):
            batch = products[i:i + batch_size]
            try:
                self.db.executemany(sql, batch)
                logger.info("Saved batch %d-%d of %d products", i + 1, min(i + batch_size, len(products)), len(products))
            except Exception as e:
                logger.error("Error saving batch %d-%d: %s", i, i + batch_size, e)
                # Try one by one
                for p in batch:
                    try:
                        self.db.execute(sql, p)
                    except Exception as e2:
                        logger.error("Error saving product %s: %s", p.get("tiki_product_id"), e2)

    def crawl_all(self, categories=None):
        """Crawl all configured categories."""
        if categories is None:
            categories = DEFAULT_CATEGORIES

        total_products = 0
        for slug, cat_id, name in categories:
            logger.info("=" * 60)
            logger.info("Starting crawl: %s (ID: %s)", name, cat_id)
            logger.info("=" * 60)

            try:
                products = self.crawl_category(slug, cat_id, name)
                self.save_products(products)
                total_products += len(products)
            except Exception as e:
                logger.error("Error crawling category %s: %s", name, e)

            self._sleep()

        logger.info("=" * 60)
        logger.info("CRAWL COMPLETE: %d total products from %d categories", total_products, len(categories))
        return total_products


# ============================================================
# Main
# ============================================================

def main():
    parser = argparse.ArgumentParser(description="Tiki.vn Product Crawler")
    parser.add_argument("--host", default="mysql-primary", help="MySQL host")
    parser.add_argument("--port", type=int, default=3306, help="MySQL port")
    parser.add_argument("--user", default="tiki", help="MySQL user")
    parser.add_argument("--password", default="tiki_dev", help="MySQL password")
    parser.add_argument("--database", default="tiki_platform", help="MySQL database")
    parser.add_argument("--max-pages", type=int, default=5, help="Max pages per category")
    parser.add_argument("--delay-min", type=float, default=1.5, help="Min delay between requests")
    parser.add_argument("--delay-max", type=float, default=3.5, help="Max delay between requests")
    parser.add_argument("--categories", nargs="*", help="Specific category slugs to crawl (default: all)")
    args = parser.parse_args()

    # Connect to database
    db = Database(
        host=args.host, port=args.port, user=args.user,
        password=args.password, database=args.database,
    )
    db.connect()

    try:
        # Initialize crawler
        crawler = TikiCrawler(
            db=db,
            max_pages_per_category=args.max_pages,
            delay_range=(args.delay_min, args.delay_max),
        )

        # Filter categories if specified
        categories = DEFAULT_CATEGORIES
        if args.categories:
            categories = [c for c in DEFAULT_CATEGORIES if c[0] in args.categories]

        # Run crawl
        total = crawler.crawl_all(categories)
        logger.info("Done! Total products crawled: %d", total)
    finally:
        db.close()


if __name__ == "__main__":
    main()
