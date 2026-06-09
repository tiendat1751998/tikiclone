#!/usr/bin/env python3
"""
TikiClone Stress Test - Final Version
Tests actual available endpoints through Traefik LB.
"""
import http.client
import json
import os
import statistics
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

HOST = "localhost"
PORT = 80
TIMEOUT = 30

# Discover working endpoints first
ENDPOINTS_TO_TEST = [
    "/health",
    "/metrics",
    "/api/v1/products?page=1&limit=20",
    "/api/v1/products?page=1&limit=20&sort=price",
    "/api/v1/categories",
    "/api/v1/search?q=laptop&page=1&limit=20",
]

CONCURRENT_LEVELS = [10, 50, 100, 200, 500]
REQUESTS_PER_LEVEL = 1000


def make_request(path):
    start = time.time()
    try:
        conn = http.client.HTTPConnection(HOST, PORT, timeout=TIMEOUT)
        conn.request("GET", path, headers={"Host": "localhost", "Accept": "application/json"})
        resp = conn.getresponse()
        body = resp.read()
        elapsed = time.time() - start
        status = resp.status
        conn.close()
        return {"status": status, "latency": elapsed, "error": None, "size": len(body)}
    except Exception as e:
        elapsed = time.time() - start
        return {"status": 0, "latency": elapsed, "error": str(e), "size": 0}


def run_test(name, path, concurrency, total):
    results = []
    start = time.time()

    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(make_request, path) for _ in range(total)]
        for f in as_completed(futures):
            results.append(f.result())

    elapsed = time.time() - start
    latencies = [r["latency"] for r in results if not r["error"]]
    statuses = [r["status"] for r in results]

    return {
        "name": name,
        "path": path,
        "concurrency": concurrency,
        "total": total,
        "time": elapsed,
        "rps": total / elapsed if elapsed > 0 else 0,
        "success": len([s for s in statuses if 200 <= s < 300]),
        "failed": len([s for s in statuses if s >= 400 or s == 0]),
        "errors": len([r for r in results if r["error"]]),
        "avg_ms": statistics.mean(latencies) * 1000 if latencies else 0,
        "p50_ms": statistics.median(latencies) * 1000 if latencies else 0,
        "p95_ms": sorted(latencies)[int(len(latencies) * 0.95)] * 1000 if latencies else 0,
        "p99_ms": sorted(latencies)[int(len(latencies) * 0.99)] * 1000 if latencies else 0,
        "statuses": {s: statuses.count(s) for s in set(statuses)},
    }


def fmt(r):
    print(f"\n{'='*70}")
    print(f"[{r['name']}] concurrency={r['concurrency']} requests={r['total']}")
    print(f"  Path:         {r['path']}")
    print(f"  Time:         {r['time']:.2f}s")
    print(f"  Throughput:   {r['rps']:.0f} req/s")
    print(f"  Success:      {r['success']}/{r['total']}")
    print(f"  Failed:       {r['failed']}")
    print(f"  Errors:       {r['errors']}")
    print(f"  Latency avg:  {r['avg_ms']:.1f}ms")
    print(f"  Latency p50:  {r['p50_ms']:.1f}ms")
    print(f"  Latency p95:  {r['p95_ms']:.1f}ms")
    print(f"  Latency p99:  {r['p99_ms']:.1f}ms")
    print(f"  Status codes: {r['statuses']}")


print("=" * 70)
print("TIKI CLONE - STRESS TEST")
print(f"Target: http://{HOST}:{PORT}")
print(f"Time: {time.strftime('%Y-%m-%d %H:%M:%S')}")
print("=" * 70)

# Phase 1: Endpoint discovery
print("\n--- Phase 1: Endpoint Discovery ---")
working = []
for path in ENDPOINTS_TO_TEST:
    r = make_request(path)
    status = "OK" if r["status"] == 200 else ("FAIL" if r["status"] >= 400 else "ERROR")
    print(f"  {path:50s} -> HTTP {r['status']:3d} ({r['latency']*1000:.0f}ms) [{status}]")
    if r["status"] == 200:
        working.append(path)

if not working:
    print("\nWARNING: No working endpoints found. Testing with /health anyway.")
    working = ["/health"]

# Phase 2: Ramp-up on working endpoints
print("\n--- Phase 2: Ramp-Up Test ---")
for path in working[:2]:
    for conc in CONCURRENT_LEVELS:
        r = run_test(f"ramp_{conc}", path, conc, REQUESTS_PER_LEVEL)
        fmt(r)

# Phase 3: Sustained load on all working endpoints
print("\n--- Phase 3: Sustained Load (100 concurrent, 5000 requests) ---")
for path in working:
    r = run_test("sustained", path, 100, 5000)
    fmt(r)

# Phase 4: Burst test
print("\n--- Phase 4: Burst Test (500 concurrent, 10000 requests) ---")
if working:
    r = run_test("burst", working[0], 500, 10000)
    fmt(r)

# Phase 5: Mixed endpoint test
print("\n--- Phase 5: Mixed Endpoint Test (200 concurrent, 5000 requests) ---")
if len(working) > 1:
    mixed_results = []
    start = time.time()
    with ThreadPoolExecutor(max_workers=200) as pool:
        futures = []
        for i in range(5000):
            path = working[i % len(working)]
            futures.append(pool.submit(make_request, path))
        for f in as_completed(futures):
            mixed_results.append(f.result())
    elapsed = time.time() - start
    latencies = [r["latency"] for r in mixed_results if not r["error"]]
    statuses = [r["status"] for r in mixed_results]
    r = {
        "name": "mixed",
        "path": "multiple",
        "concurrency": 200,
        "total": 5000,
        "time": elapsed,
        "rps": 5000 / elapsed if elapsed > 0 else 0,
        "success": len([s for s in statuses if 200 <= s < 300]),
        "failed": len([s for s in statuses if s >= 400 or s == 0]),
        "errors": len([r for r in mixed_results if r["error"]]),
        "avg_ms": statistics.mean(latencies) * 1000 if latencies else 0,
        "p50_ms": statistics.median(latencies) * 1000 if latencies else 0,
        "p95_ms": sorted(latencies)[int(len(latencies) * 0.95)] * 1000 if latencies else 0,
        "p99_ms": sorted(latencies)[int(len(latencies) * 0.99)] * 1000 if latencies else 0,
        "statuses": {s: statuses.count(s) for s in set(statuses)},
    }
    fmt(r)

print("\n" + "=" * 70)
print("STRESS TEST COMPLETE")
print("=" * 70)
