#!/usr/bin/env python3
"""Focused stress test on public (no-auth) endpoints."""
import urllib.request, json, time, concurrent.futures, statistics, sys

BASE = "http://tiki_gateway:8080"

# Only public endpoints (no auth required)
ENDPOINTS = [
    ("/api/v1/products", 30),
    ("/api/v1/products?page=1&size=20", 20),
    ("/api/v1/categories", 15),
    ("/api/v1/categories/tree", 10),
    ("/api/v1/products/search?q=dien+thoai", 10),
    ("/api/v1/products/search?q=laptop", 5),
    ("/api/v1/products/spu-dien-thoai-001", 5),
]

def hit(path):
    t0 = time.time()
    try:
        req = urllib.request.urlopen(BASE + path, timeout=5)
        status = req.status
    except Exception as e:
        status = -1
    rt = (time.time() - t0) * 1000
    return (path, status, rt)

def run_test(num_requests, concurrency, label):
    results = []
    t_start = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as ex:
        futures = [ex.submit(hit, ENDPOINTS[i % len(ENDPOINTS)][0]) for i in range(num_requests)]
        for f in concurrent.futures.as_completed(futures):
            results.append(f.result())
    elapsed = time.time() - t_start

    total = len(results)
    successes = sum(1 for _, s, _ in results if 200 <= s < 500)
    failures = total - successes
    rps = total / elapsed
    rts = sorted([rt_val for _, s, rt_val in results if 200 <= s < 500 and rt_val > 0])

    if not rts:
        print(f"  {label}: ALL FAILED")
        return None

    avg_rt = statistics.mean(rts)
    p50 = rts[len(rts)//2]
    p90 = rts[int(len(rts)*0.9)]
    p95 = rts[int(len(rts)*0.95)]
    p99 = rts[int(len(rts)*0.99)]
    p999 = rts[min(int(len(rts)*0.999), len(rts)-1)]
    max_rt = max(rts)

    print(f"  {label:>12}: {rps:>6.0f} RPS | avg={avg_rt:>6.1f}ms p50={p50:>5.1f}ms p90={p90:>5.1f}ms p95={p95:>5.1f}ms p99={p99:>5.1f}ms p99.9={p999:>6.1f}ms max={max_rt:>6.1f}ms | fail={failures}/{total}")

    return {"label": label, "rps": rps, "avg": avg_rt, "p99": p99, "p999": p999, "fail": failures, "total": total}

print(f"TikiClone Public Endpoint Stress Test")
print(f"Gateway: {BASE}")
print()

# Verify endpoints
print("Endpoint check:")
for path, _ in ENDPOINTS:
    try:
        req = urllib.request.urlopen(BASE + path, timeout=5)
        print(f"  {path[:50]:50s} OK ({req.status})")
    except Exception as e:
        print(f"  {path[:50]:50s} FAIL ({e})")

print()
print(f"{'Level':>12} {'RPS':>7} {'Avg':>7} {'p50':>6} {'p90':>6} {'p95':>6} {'p99':>6} {'p99.9':>7} {'Max':>7} {'Fail':>6}")
print("-" * 90)

for concurrency, num_requests, label in [
    (5, 100, "warmup"),
    (10, 200, "100_rps"),
    (20, 400, "200_rps"),
    (50, 1000, "500_rps"),
    (100, 2000, "1000_rps"),
]:
    r = run_test(num_requests, concurrency, label)
    if r is None:
        break
    time.sleep(1)
