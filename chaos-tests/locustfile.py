"""
TikiClone Load Test Suite
==========================
Tests through Nginx (port 8443) → Gateway → Backend services.
Run at 10, 100, 1000, 5000, 10000 TPS to find bottlenecks.

Usage:
    locust -f locustfile.py --headless -u 10 -r 1 --run-time 30s --csv=results/10tps
    locust -f locustfile.py --headless -u 100 -r 10 --run-time 30s --csv=results/100tps
    locust -f locustfile.py --headless -u 1000 -r 50 --run-time 60s --csv=results/1000tps
    locust -f locustfile.py --headless -u 5000 -r 200 --run-time 60s --csv=results/5000tps
    locust -f locustfile.py --headless -u 10000 -r 500 --run-time 60s --csv=results/10000tps

Or run the orchestrator:
    python3 run_load_test.py
"""

import json
import time
import random
import subprocess
import sys
from datetime import datetime
from pathlib import Path

from locust import HttpUser, task, between, events, constant_throughput
from locust.runners import MasterRunner

# ═══════════════════════════════════════════════════════════════
# Configuration
# ═══════════════════════════════════════════════════════════════

# Entry points
HTTPS_BASE = "https://localhost:8443"   # Nginx → Gateway + Next.js
HTTP_GATEWAY = "http://localhost:8080"   # Direct gateway

# Test data extracted from actual API responses
SAMPLE_PRODUCTS = [
    "spu-dien-thoai-001", "spu-dien-thoai-002", "spu-dien-thoai-003",
    "spu-dien-thoai-004", "spu-laptop-001", "spu-laptop-002",
    "spu-laptop-003", "spu-thoi-trang-001", "spu-thoi-trang-002",
    "spu-nha-bep-001", "spu-nha-bep-002", "spu-den-001",
    "spu-nha-tam-001", "spu-trang-tri-001", "spu-noi-that-001",
    "spu-do-choi-me-be-001",
]

SAMPLE_CATEGORIES = [
    "dien-thoai", "laptop", "thoi-trang", "1951", "1973", "2150",
    "1966", "2015", "1883", "do-choi-me-be",
]

SEARCH_QUERIES = [
    "iPhone", "Samsung", "laptop", "MacBook", "Dell", "Lenovo",
    "áo thun", "túi xách", "đèn", "bàn", "nồi", "sen vòi",
    "xiaomi", "oppo", "túi", "kệ", "bộ nồi",
]


# ═══════════════════════════════════════════════════════════════
# User Behaviors
# ═══════════════════════════════════════════════════════════════

class TikiBrowseUser(HttpUser):
    """
    Simulates a browsing user: homepage → categories → products → search.
    This is the most common user pattern.
    """
    # Wait between tasks (think time)
    wait_time = between(0.1, 0.5)

    def on_start(self):
        """Called when a user starts."""
        self.client.verify = False  # Self-signed cert

    @task(10)
    def view_homepage(self):
        """Load the Next.js homepage."""
        with self.client.get("/", name="01_homepage", catch_response=True) as resp:
            if resp.status_code in (200, 301, 302):
                resp.success()
            else:
                resp.failure(f"Homepage: {resp.status_code}")

    @task(20)
    def view_products_list(self):
        """Browse product listing page (through gateway)."""
        page = random.randint(1, 5)
        category = random.choice(SAMPLE_CATEGORIES)
        self.client.get(
            f"/api/v1/products?page={page}&category={category}",
            name="02_products_list",
        )

    @task(15)
    def view_product_detail(self):
        """View a specific product detail."""
        product_id = random.choice(SAMPLE_PRODUCTS)
        self.client.get(
            f"/api/v1/products/{product_id}",
            name="03_product_detail",
        )

    @task(15)
    def view_categories(self):
        """Load category tree."""
        self.client.get(
            "/api/v1/categories",
            name="04_categories",
        )

    @task(10)
    def search_products(self):
        """Search for products."""
        query = random.choice(SEARCH_QUERIES)
        self.client.get(
            f"/api/v1/search?q={query}",
            name="05_search",
        )

    @task(5)
    def view_recommendations(self):
        """Load recommendations."""
        self.client.get(
            "/api/v1/recommendations",
            name="06_recommendations",
        )


class TikiAuthenticatedUser(HttpUser):
    """
    Simulates an authenticated user performing cart/order operations.
    Requires valid auth tokens.
    """
    wait_time = between(0.2, 1.0)

    def on_start(self):
        self.client.verify = False
        # Get auth token
        resp = self.client.post("/api/v1/auth/login", json={
            "email": "test@tiki.vn",
            "password": "test123456",
        }, name="auth_login")
        if resp.status_code == 200:
            try:
                token = resp.json().get("access_token", "")
                self.headers = {"Authorization": f"Bearer {token}"}
            except Exception:
                self.headers = {}
        else:
            self.headers = {}

    @task(10)
    def view_cart(self):
        """View cart contents."""
        self.client.get("/api/v1/cart", name="10_view_cart", headers=self.headers)

    @task(5)
    def add_to_cart(self):
        """Add item to cart."""
        product_id = random.choice(SAMPLE_PRODUCTS)
        self.client.post("/api/v1/cart/items", json={
            "product_id": product_id,
            "sku_id": f"sku-{product_id.split('-')[-1]}",
            "quantity": random.randint(1, 3),
        }, name="11_add_cart", headers=self.headers)

    @task(3)
    def checkout(self):
        """Initiate checkout."""
        self.client.post("/api/v1/checkout", json={
            "items": [{"product_id": random.choice(SAMPLE_PRODUCTS), "quantity": 1}],
            "shipping_address": {"city": "HCMC", "district": "District 1"},
        }, name="12_checkout", headers=self.headers)

    @task(5)
    def view_orders(self):
        """View order history."""
        self.client.get("/api/v1/orders", name="13_orders", headers=self.headers)

    @task(2)
    def view_inventory(self):
        """Check inventory for a product."""
        product_id = random.choice(SAMPLE_PRODUCTS)
        self.client.get(
            f"/api/v1/inventory?product_id={product_id}",
            name="14_inventory", headers=self.headers,
        )


class TikiStressUser(HttpUser):
    """
    High-intensity user for stress testing. Minimal wait times.
    Used for 5000+ TPS tests.
    """
    wait_time = constant_throughput(1)  # 1 request per second per user

    def on_start(self):
        self.client.verify = False

    @task(5)
    def get_products(self):
        """Fast product list fetch."""
        self.client.get(
            f"/api/v1/products?page={random.randint(1, 3)}",
            name="stress_products",
        )

    @task(3)
    def get_categories(self):
        """Fast category fetch."""
        self.client.get(
            "/api/v1/categories",
            name="stress_categories",
        )

    @task(5)
    def get_product_detail(self):
        """Fast product detail fetch."""
        self.client.get(
            f"/api/v1/products/{random.choice(SAMPLE_PRODUCTS)}",
            name="stress_product_detail",
        )

    @task(2)
    def get_homepage(self):
        """Load homepage."""
        self.client.get("/", name="stress_homepage")


# ═══════════════════════════════════════════════════════════════
# Custom Metrics & Hooks
# ═══════════════════════════════════════════════════════════════

# Track per-service latency from container stats
_container_stats = {}

@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Capture baseline container stats before test."""
    print(f"\n{'='*60}")
    print(f"Load test starting at {datetime.now().isoformat()}")
    print(f"{'='*60}\n")
    _capture_container_stats("baseline")

@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Capture post-test container stats."""
    _capture_container_stats("post_test")

def _capture_container_stats(label):
    """Capture Docker container CPU/memory stats."""
    try:
        result = subprocess.run(
            "docker stats --no-stream --format '{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}' "
            "tikiclone-gateway-1 tikiclone-web-1 tikiclone-web-tls-1 "
            "tikiclone-product-1 tikiclone-cart-1 tikiclone-order-1 "
            "tikiclone-payment-1 tikiclone-inventory-1 "
            "tikiclone-mysql-primary-1 tikiclone-redis-master-1 "
            "tikiclone-mongodb-1 tikiclone-kafka-1 2>/dev/null",
            shell=True, capture_output=True, text=True, timeout=30
        )
        _container_stats[label] = result.stdout.strip()
        print(f"\n[{label}] Container stats:")
        print(result.stdout)
    except Exception as e:
        print(f"Could not capture stats: {e}")

@events.request.add_listener
def on_request(request_type, name, response_time, response_length,
               response, context, exception, **kwargs):
    """Log slow requests for analysis."""
    if response_time > 5000:  # > 5 seconds
        print(f"SLOW: {name} took {response_time:.0f}ms")
    if exception:
        print(f"ERROR: {name} - {exception}")
