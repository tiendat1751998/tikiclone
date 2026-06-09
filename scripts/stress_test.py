#!/usr/bin/env python3
"""
TikiClone Stress Test
Tests the main API endpoints through the load balancer.
"""
import http.client
import json
import os
import statistics
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.parse import urlparse

# Configuration
BASE_URL = os.getenv("BASE_URL", "http://localhost")
HOST = "localhost"
PORT = 80

# Test endpoints
ENDPOINTS = {
    "health": "/health",
    "products_list": "/api/v1/products?page=1&limit=20",
    "category": "/api/v1/categories",
    "search": "/api/v1/search?q=laptop&page=1&limit=20",
}

# Test parameters
CONCURRENT_CONNECTIONS = [10, 50, 100, 200]
REQUESTS_PER_TEST = 500
TIMEOUT = 30


def make_request(endpoint, method="GET", body=None, headers=None):
    """Make a single HTTP request and return timing and status."""
    start = time.time()
    try:
        conn = http.client.HTTPConnection(HOST, PORT, timeout=TIMEOUT)
        if headers:
            conn.request(method, endpoint, body=body, headers=headers)
        else:
            conn.request(method, endpoint)
        resp = conn.getresponse()
        resp.read()  # consume body
        elapsed = time.time() - start
        status = resp.status
        conn.close()
        return {"status": status, "latency": elapsed, "error": None}
    except Exception as e:
        elapsed = time.time() - start
        return {"status": 0, "latency": elapsed, "error": str(e)}


def run_load_test(endpoint_name, path, concurrency, total_requests):
    """Run a load test with given concurrency."""
    results = []
    errors = []
    start_time = time.time()

    def worker():
        result = make_request(path)
        results.append(result)
        if result["error"] or result["status"] >= 500:
            errors.append(result)

    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(worker) for _ in range(total_requests)]
        for f in as_completed(futures):
            pass

    total_time = time.time() - start_time
    latencies = [r["latency"] for r in results if not r["error"]]
    statuses = [r["status"] for r in results]

    return {
        "endpoint": endpoint_name,
        "concurrency": concurrency,
        "total_requests": total_requests,
        "successful": len([s for s in statuses if 200 <= s < 300]),
        "failed": len([s for s in statuses if s >= 400 or s == 0]),
        "error_count": len(errors),
        "total_time": total_time,
        "rps": total_requests / total_time if total_time > 0 else 0,
        "latency_avg": statistics.mean(latencies) if latencies else 0,
        "latency_p50": statistics.median(latencies) if latencies else 0,
        "latency_p95": sorted(latencies)[int(len(latencies) * 0.95)] if latencies else 0,
        "latency_p99": sorted(latencies)[int(len(latencies) * 0.99)] if latencies else 0,
        "latency_min": min(latencies) if latencies else 0,
        "latency_max": max(latencies) if latencies else 0,
        "status_codes": {s: statuses.count(s) for s in set(statuses)},
    }


def print_result(r):
    """Pretty print test results."""
    print(f"\n{'='*70}")
    print(f"Endpoint: {r['endpoint']} | Concurrency: {r['concurrency']} | Requests: {r['total_requests']}")
    print(f"{'='*70}")
    print(f"  Total time:     {r['total_time']:.2f}s")
    print(f"  Throughput:     {r['rps']:.0f} req/s")
    print(f"  Successful:     {r['successful']}/{r['total_requests']}")
    print(f"  Failed:         {r['failed']}")
    print(f"  Errors:         {r['error_count']}")
    print(f"  Latency avg:    {r['latency_avg']*1000:.1f}ms")
    print(f"  Latency p50:    {r['latency_p50']*1000:.1f}ms")
    print(f"  Latency p95:    {r['latency_p95']*1000:.1f}ms")
    print(f"  Latency p99:    {r['latency_p99']*1000:.1f}ms")
    print(f"  Latency min:    {r['latency_min']*1000:.1f}ms")
    print(f"  Latency max:    {r['latency_max']*1000:.1f}ms")
    print(f"  Status codes:   {r['status_codes']}")


def main():
    print("="*70)
    print("TIKI CLONE - STRESS TEST")
    print(f"Target: http://{HOST}:{PORT}")
    print(f"Time: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print("="*70)

    # First: verify endpoints are reachable
    print("\n--- Endpoint Health Check ---")
    all_ok = True
    for name, path in ENDPOINTS.items():
        result = make_request(path)
        status = "OK" if result["status"] == 200 else "FAIL"
        if result["status"] != 200:
            all_ok = False
        print(f"  {name:20s} -> HTTP {result['status']:3d} ({result['latency']*1000:.0f}ms) [{status}]")
        if result["error"]:
            print(f"    Error: {result['error']}")

    if not all_ok:
        print("\nWARNING: Some endpoints are not responding. Stress test may fail.")

    # Phase 1: Single endpoint ramp-up
    print("\n\n--- Phase 1: Products API Ramp-Up ---")
    for conc in CONCURRENT_CONNECTIONS:
        r = run_load_test("products_list", ENDPOINTS["products_list"], conc, REQUESTS_PER_TEST)
        print_result(r)

    # Phase 2: All endpoints at moderate concurrency
    print("\n\n--- Phase 2: All Endpoints at 50 concurrent ---")
    for name, path in ENDPOINTS.items():
        r = run_load_test(name, path, 50, 200)
        print_result(r)

    # Phase 3: Sustained load
    print("\n\n--- Phase 3: Sustained Load (100 concurrent, 2000 requests) ---")
    r = run_load_test("products_sustained", ENDPOINTS["products_list"], 100, 2000)
    print_result(r)

    # Phase 4: Burst test
    print("\n\n--- Phase 4: Burst Test (500 concurrent, 2500 requests) ---")
    r = run_load_test("products_burst", ENDPOINTS["products_list"], 500, 2500)
    print_result(r)

    print("\n" + "="*70)
    print("STRESS TEST COMPLETE")
    print("="*70)


if __name__ == "__main__":
    main()
