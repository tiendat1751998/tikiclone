#!/usr/bin/env python3
"""API-only cache load test - skip Next.js homepage"""

import asyncio
import aiohttp
import ssl
import time
import statistics
from datetime import datetime

HTTPS_BASE = "https://localhost:8443"
VERIFY_SSL = False

# API endpoints only (no homepage)
ENDPOINTS = [
    ("GET", "/api/v1/products", 40),
    ("GET", "/api/v1/categories", 20),
    ("GET", "/api/v1/products/spu-dien-thoai-001", 10),
    ("GET", "/api/v1/search?q=iPhone", 15),
    ("GET", "/api/v1/inventory?product_id=spu-dien-thoai-001", 8),
    ("GET", "/api/v1/cart", 5),
    ("GET", "/api/v1/orders", 2),
]

async def test_api_only(tps_target=1000, num_users=50, duration=30):
    import random
    
    ssl_ctx = ssl.create_default_context()
    ssl_ctx.check_hostname = False
    ssl_ctx.verify_mode = ssl.CERT_NONE
    connector = aiohttp.TCPConnector(limit=num_users * 2, ssl=ssl_ctx)
    
    results = []
    end_time = time.time() + duration
    
    async def worker(session):
        nonlocal results
        while time.time() < end_time:
            total_w = sum(w for _, _, w in ENDPOINTS)
            pick = random.uniform(0, total_w)
            cum = 0
            for m, p, w in ENDPOINTS:
                cum += w
                if pick <= cum:
                    path = p
                    break
            
            url = HTTPS_BASE + path
            t0 = time.time()
            try:
                async with session.request("GET", url, ssl=VERIFY_SSL, timeout=aiohttp.ClientTimeout(total=10)) as resp:
                    await resp.read()
                    status = resp.status
            except Exception:
                status = -1
            rt = (time.time() - t0) * 1000
            results.append((path, status, rt))
    
    async with aiohttp.ClientSession(connector=connector) as session:
        workers = [asyncio.create_task(worker(session)) for _ in range(num_users)]
        await asyncio.gather(*workers)
    
    rts = sorted([rt for _, s, rt in results if 200 <= s < 500 and rt > 0])
    if rts:
        print(f"\nAPI-only test: {len(results)} requests in {duration}s")
        print(f"Avg RT: {statistics.mean(rts):.0f}ms | p50: {rts[len(rts)//2]:.0f}ms | p90: {rts[int(len(rts)*0.9)]:.0f}ms")
        print(f"p95: {rts[int(len(rts)*0.95)]:.0f}ms | p99: {rts[int(len(rts)*0.99)]:.0f}ms | max: {max(rts):.0f}ms")
        print(f"RPS: {len(results)/duration:.1f}")

if __name__ == "__main__":
    asyncio.run(test_api_only())