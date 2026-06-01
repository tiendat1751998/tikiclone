#!/usr/bin/env python3
"""
Generate colored placeholder images for products that don't have real images.
This serves as a fallback when crawling real images fails.
"""
import mysql.connector
import os
import struct
import zlib

DB_CONFIG = {
    'host': os.environ.get('MYSQL_HOST', 'localhost'),
    'port': int(os.environ.get('MYSQL_PORT', '3306')),
    'user': os.environ.get('MYSQL_USER', 'tiki'),
    'password': os.environ.get('MYSQL_PASSWORD', 'tiki_dev'),
    'database': os.environ.get('MYSQL_DATABASE', 'tiki_platform'),
}

IMAGE_DIR = os.path.join(os.path.dirname(__file__), '..', 'apps', 'web', 'public', 'images', 'products')
os.makedirs(IMAGE_DIR, exist_ok=True)

COLORS = [
    (52, 152, 219), (231, 76, 60), (46, 204, 113), (155, 89, 182),
    (241, 196, 15), (230, 126, 34), (26, 188, 156), (149, 165, 166),
    (52, 73, 94), (243, 156, 18), (142, 68, 173), (22, 160, 133),
]


def create_png(width, height, r, g, b, text=""):
    """Create a minimal PNG with a solid color background."""
    def make_chunk(chunk_type, data):
        c = chunk_type + data
        return struct.pack('>I', len(data)) + c + struct.pack('>I', zlib.crc32(c) & 0xffffffff)

    header = b'\x89PNG\r\n\x1a\n'
    ihdr = make_chunk(b'IHDR', struct.pack('>IIBBBBB', width, height, 8, 2, 0, 0, 0))

    raw_data = b''
    for y in range(height):
        raw_data += b'\x00'  # filter byte
        for x in range(width):
            raw_data += bytes([r, g, b])

    idat = make_chunk(b'IDAT', zlib.compress(raw_data))
    iend = make_chunk(b'IEND', b'')

    return header + ihdr + idat + iend


def main():
    conn = mysql.connector.connect(**DB_CONFIG)
    cursor = conn.cursor()

    cursor.execute("""
        SELECT DISTINCT p.id, p.name
        FROM products p
        JOIN product_media pm ON pm.product_id = p.id
        WHERE pm.url = '/images/products/default-product.png'
           OR pm.url LIKE '/images/products/default-product%'
        LIMIT 50
    """)

    products = cursor.fetchall()
    print(f"Generating placeholders for {len(products)} products...")

    updated = 0
    for prod_id, name in products:
        filename = f"{prod_id}.jpg"
        filepath = os.path.join(IMAGE_DIR, filename)
        public_path = f"/images/products/{filename}"

        # Only generate if file doesn't exist
        if not os.path.exists(filepath):
            color = COLORS[hash(prod_id) % len(COLORS)]
            png_data = create_png(400, 400, *color)
            # Write as .png first then we'll link .jpg
            png_path = filepath.replace('.jpg', '.png')
            with open(png_path, 'wb') as f:
                f.write(png_data)
            # Copy as jpg too
            with open(filepath, 'wb') as f:
                f.write(png_data)

        cursor.execute(
            "UPDATE product_media SET url = %s, thumbnail_url = %s WHERE product_id = %s AND url = '/images/products/default-product.png'",
            (public_path, public_path, prod_id)
        )
        updated += cursor.rowcount

    conn.commit()
    cursor.close()
    conn.close()
    print(f"Updated {updated} media entries with placeholder images.")


if __name__ == '__main__':
    main()
