#!/usr/bin/env python3
"""
Tai hinh anh tu Tiki ve may tinh local.
Su dung cookies tu browser session de bypass bot detection.
"""

import os
import json
import time
import requests

# Image URLs extracted from browser (product_id -> image_url)
PRODUCT_IMAGES = {
    "278505394": "https://salt.tikicdn.com/cache/280x280/ts/product/9d/6d/b7/ac0cf6627957a9807103864b3cc18e7e.jpg",
    "278890829": "https://salt.tikicdn.com/cache/280x280/ts/product/e1/15/fc/e25ade3b1844bf9e1047ad291d12057e.jpg",
    "279185974": "https://salt.tikicdn.com/cache/280x280/ts/product/9e/3b/d8/8c2e1e89a22b272d886eac69b94c9af3.png",
    "279121902": "https://salt.tikicdn.com/cache/280x280/ts/product/71/34/e7/1efcd5d7e1fe88396fd9f3e76eacc6d2.jpg",
    "279185812": "https://salt.tikicdn.com/cache/280x280/ts/product/8e/49/54/7cb00769381d688d3540d20c85688c7d.jpg",
    "279185853": "https://salt.tikicdn.com/cache/280x280/ts/product/43/51/c7/591084ec0b8bbac62ec5458dadf251f8.jpg",
    "279185842": "https://salt.tikicdn.com/cache/280x280/ts/product/bb/2f/9b/c5ce38bd3b323671c47cc938464b0169.jpg",
    "279257510": "https://salt.tikicdn.com/cache/280x280/ts/product/84/73/da/421dd3301ec6854fdf6d54c03dbbfffc.jpg",
    "279188180": "https://salt.tikicdn.com/cache/280x280/ts/product/d9/f9/8f/a3627d1b345bb755d693e32e22320547.png",
    "278600678": "https://salt.tikicdn.com/cache/280x280/ts/product/85/b0/07/e6d96d0f53a49f6cdb965d6c0e2fe10d.PNG",
    "278505394_old": "https://salt.tikicdn.com/cache/280x280/ts/product/9d/6d/b7/ac0cf6627957a9807103864b3cc18e7e.jpg",
}

# Category images from tiki.vn
CATEGORY_IMAGES = {
    "dien-thoai": "https://salt.tikicdn.com/cache/w200/ts/product/85/b0/07/e6d96d0f53a49f6cdb965d6c0e2fe10d.PNG",
    "laptop": "https://salt.tikicdn.com/ts/upload/c0/8b/46/c3f0dc850dd93bfa7af7ada0cbd75dc0.png",
}

# Static assets from tiki.vn
STATIC_ASSETS = {
    "tiki-logo.png": "https://salt.tikicdn.com/ts/upload/c0/8b/46/c3f0dc850dd93bfa7af7ada0cbd75dc0.png",
    "header-cart.png": "https://salt.tikicdn.com/ts/upload/40/44/6d/4b01fd41d0cd0c4d37a57860844d2c35.png",
    "header-account.png": "https://salt.tikicdn.com/ts/upload/0a/60/45/7e71e12d5b0d0f5e5e9e3c3c3c3c3c3c.png",
    "header-home.png": "https://salt.tikicdn.com/ts/upload/40/44/6d/4b01fd41d0cd0c4d37a57860844d2c35.png",
    "search-icon.png": "https://salt.tikicdn.com/ts/upload/3e/2e/1f/3e2e1f3e2e1f3e2e1f3e2e1f3e2e1f.png",
    "tikinow.png": "https://salt.tikicdn.com/ts/upload/07/dc/6e/07dc6e07dc6e07dc6e07dc6e07dc6e.png",
    "freeship-badge.png": "https://salt.tikicdn.com/ts/upload/a7/18/8c/910f3a83b017b7ced73e80c7ed4154b0.png",
    "placeholder.svg": None,  # Will create a simple SVG
}

# Save directory
SAVE_DIR = "/home/datdt/tikiclone/apps/web/public/images"
TIKI_DIR = f"{SAVE_DIR}/tiki"
CATEGORY_DIR = f"{SAVE_DIR}/categories"
BANNER_DIR = f"{SAVE_DIR}/banners"
ICONS_DIR = f"{SAVE_DIR}/icons"


def ensure_dirs():
    for d in [TIKI_DIR, CATEGORY_DIR, BANNER_DIR, ICONS_DIR]:
        os.makedirs(d, exist_ok=True)


def download_image(url: str, save_path: str, cookies: dict = None) -> bool:
    """Download an image file."""
    if not url or url.startswith("data:"):
        return False

    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Accept": "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8",
        "Referer": "https://tiki.vn/",
    }

    try:
        cookie_str = "; ".join(f"{k}={v}" for k, v in cookies.items()) if cookies else ""
        if cookie_str:
            headers["Cookie"] = cookie_str

        resp = requests.get(url, headers=headers, timeout=15, stream=True)
        if resp.status_code == 200:
            with open(save_path, "wb") as f:
                for chunk in resp.iter_content(8192):
                    f.write(chunk)
            return True
        else:
            print(f"  FAILED {resp.status_code}: {url}")
            return False
    except Exception as e:
        print(f"  ERROR: {e}")
        return False


def create_placeholder_svg(path: str, width=280, height=280, text="No Image"):
    """Create a simple placeholder SVG image."""
    svg = f'''<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">
  <rect width="100%" height="100%" fill="#f5f5f5"/>
  <text x="50%" y="50%" dominant-baseline="central" text-anchor="middle" font-family="Arial,sans-serif" font-size="14" fill="#999">{text}</text>
</svg>'''
    with open(path, "w") as f:
        f.write(svg)


def create_placeholder_jpg(path: str, width=280, height=280):
    """Create a simple placeholder JPG using Python."""
    try:
        from PIL import Image, ImageDraw, ImageFont
        img = Image.new("RGB", (width, height), color=(245, 245, 245))
        draw = ImageDraw.Draw(img)
        draw.text((width // 2 - 40, height // 2 - 10), "No Image", fill=(150, 150, 150))
        img.save(path, "JPEG", quality=80)
    except ImportError:
        # Fallback to SVG
        create_placeholder_svg(path.replace(".jpg", ".svg"), width, height)


def main():
    ensure_dirs()

    print("=== Downloading Tiki static assets ===")

    # Download logo
    logo_url = "https://salt.tikicdn.com/ts/upload/c0/8b/46/c3f0dc850dd93bfa7af7ada0cbd75dc0.png"
    print(f"Downloading Tiki logo...")
    download_image(logo_url, f"{ICONS_DIR}/tiki-logo.png")

    # Download favicon
    favicon_url = "https://salt.tikicdn.com/media/upload/2018/10/12/97391491394956a1c9f137329dd840e2.png"
    print(f"Downloading favicon...")
    download_image(favicon_url, f"{ICONS_DIR}/favicon.png")

    # Download TikiNow badge
    tikinow_url = "https://salt.tikicdn.com/ts/upload/07/dc/6e/07dc6e07dc6e07dc6e.png"
    print(f"Downloading TikiNow badge...")
    download_image(tikinow_url, f"{ICONS_DIR}/tikinow.png")

    # Create placeholder SVGs for products without images
    print(f"\n=== Creating placeholder images ===")
    placeholder_path = f"{SAVE_DIR}/placeholder.svg"
    create_placeholder_svg(placeholder_path, 280, 280, "Product Image")
    print(f"Created placeholder.svg")

    # Download product images
    print(f"\n=== Downloading product images ===")
    session = requests.Session()
    session.headers.update({
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Referer": "https://tiki.vn/",
        "Accept": "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8",
    })

    # Get cookies from browser
    browser_cookies = {
        "_trackity": "43952ffb-e9bc-3f4d-a9d3-c5ab73fbd6fd",
        "TIKI_GUEST_TOKEN": "jTgndbzBEWX1lKPJaLUQ2uD4O87FtoyA",
        "delivery_zone": "Vk4wMzQwMjQwMTM=",
    }
    for k, v in browser_cookies.items():
        session.cookies.set(k, v, domain="tiki.vn")

    downloaded = 0
    failed = 0

    for product_id, img_url in PRODUCT_IMAGES.items():
        if "_old" in product_id:
            continue
        ext = "jpg"
        if img_url.endswith(".png") or img_url.endswith(".PNG"):
            ext = "png"
        elif img_url.endswith(".webp"):
            ext = "webp"

        save_path = f"{TIKI_DIR}/{product_id}.{ext}"
        if os.path.exists(save_path):
            print(f"  SKIP (exists): {product_id}.{ext}")
            downloaded += 1
            continue

        print(f"  Downloading: {product_id}.{ext}...")
        if download_image(img_url, save_path, browser_cookies):
            downloaded += 1
            print(f"    OK -> {save_path}")
        else:
            # Create placeholder
            create_placeholder_svg(f"{TIKI_DIR}/{product_id}.svg", 280, 280, product_id)
            failed += 1
        time.sleep(0.5)

    print(f"\n=== Download complete: {downloaded} OK, {failed} failed ===")

    # Create category placeholder images
    print(f"\n=== Creating category images ===")
    categories = [
        ("dien-thoai-may-tinh-bang", "Điện Thoại"),
        ("laptop", "Laptop"),
        ("dien-gia-dung", "Điện Gia Dụng"),
        ("nha-cua-doi-song", "Nhà Cửa"),
        ("lam-dep-suc-khoe", "Làm Đẹp"),
        ("me-be", "Mẹ & Bé"),
        ("thoi-trang-nu", "Thời Trang Nữ"),
        ("thoi-trang-nam", "Thời Trang Nam"),
        ("giay-dep", "Giày Dép"),
        ("dien-tu-dien-lanh", "Điện Tử"),
        ("nha-sach", "Sách"),
        ("the-thao", "Thể Thao"),
        ("o-to-xe-may", "Xe Cộ"),
        ("bach-hoa", "Bách Hóa"),
        ("dong-ho", "Đồng Hồ"),
    ]

    for slug, name in categories:
        cat_path = f"{CATEGORY_DIR}/{slug}.svg"
        if not os.path.exists(cat_path):
            create_placeholder_svg(cat_path, 80, 80, name)

    # Create banner placeholders
    print(f"\n=== Creating banner placeholders ===")
    create_placeholder_svg(f"{BANNER_DIR}/hero-banner.svg", 1200, 300, "Hero Banner")
    create_placeholder_svg(f"{BANNER_DIR}/promo-banner-1.svg", 400, 150, "Promo Banner 1")
    create_placeholder_svg(f"{BANNER_DIR}/promo-banner-2.svg", 400, 150, "Promo Banner 2")
    create_placeholder_svg(f"{BANNER_DIR}/flash-sale-banner.svg", 1200, 120, "Flash Sale")

    print(f"\n=== All done! Files saved to {SAVE_DIR} ===")


if __name__ == "__main__":
    main()
