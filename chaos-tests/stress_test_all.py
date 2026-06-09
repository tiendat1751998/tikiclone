#!/usr/bin/env python3
"""
TikiClone Comprehensive Stress Test
Tests all API endpoints: products, categories, cart, orders, payment
Identifies bottlenecks for 1000 RPS / p99.9 < 10ms target
"""

import asyncio
import aiohttp
import ssl
import time
import json
import sys
import statistics
from datetime import datetime
from collections import defaultdict
from pathlib import Path

RESULTS_DIR = Path("/home/datdt/tikiclone/chaos-tests/results")
RESULTS_DIR.mkdir(exist_ok=True)

# Entry points to test
HTTP_BASE = "http://localhost:8888"  # HTTP bypass (no TLS overhead)
HTTPS_BASE = "https://localhost:8443"  # HTTPS via Traefik

# API endpoints with weights (frequency of access)
ENDPOINTS = [
    # Product APIs (highest traffic)
    ("GET", "/api/v1/products", 30),
    ("GET", "/api/v1/products?page=1&size=20", 20),
    ("GET", "/api/v1/categories", 15),
    ("GET", "/api/v1/categories/tree", 10),
    # Search
    ("GET", "/api/v1/products/search?q=dien+thoai", 10),
    ("GET", "/api/v1/products/search?q=laptop", 5),
    # Product detail
    ("GET", "/api/v1/products/spu-dien-thoai-001", 5),
    # Cart (read-only for stress test)
    ("GET", "/api/v1/cart", 3),
    # Orders (read-only)
    ("GET", "/api/v1/orders", 2),
]

# Test levels: (concurrent_users, duration_seconds, target_rps_name)
TEST_LEVELS = [
    (5, 30, "50_rps"),
    (10, 30, "100_rps"),
    (25, 30, "250_rps"),
    (50, 30, "500_rps"),
    (100, 60, "1000_rps"),
]

import random

async def worker(session, endpoints, duration, results_queue, worker_id):
    """Send requests for the specified duration."""
    end_time = time.time() + duration
    while time.time() < end_time:
        # Pick endpoint by weight
        total_w = sum(w for _, _, w in endpoints)
        pick = random.uniform(0, total_w)
        cum = 0
        method, path = endpoints[0][0], endpoints[0][1]
        for m, p, w in endpoints:
            cum += w
            if pick <= cum:
                method, path = m, p
                break

        url = HTTP_BASE + path
        t0 = time.time()
        try:
            async with session.request(method, url, ssl=False, timeout=aiohttp.ClientTimeout(total=5)) as resp:
                await resp.read()
                status = resp.status
        except asyncio.TimeoutError:
            status = 0
        except Exception as e:
            status = -1
        rt = (time.time() - t0) * 1000  # ms
        await results_queue.put((path, status, rt, worker_id))

async def run_test(concurrent_users, duration, label):
    """Run a single stress test level."""
    print(f"\n{'='*60}")
    print(f"TEST: {label} | Users: {concurrent_users} | Duration: {duration}s")
    print(f"{'='*60}")

    connector = aiohttp.TCPConnector(limit=concurrent_users * 2)
    queue = asyncio.Queue()
    all_results = []

    t_start = time.time()

    async with aiohttp.ClientSession(connector=connector) as session:
        workers = [
            asyncio.create_task(worker(session, ENDPOINTS, duration, queue, i))
            for i in range(concurrent_users)
        ]

        collector_task = asyncio.create_task(collect_results(queue, all_results))

        await asyncio.gather(*workers, return_exceptions=True)
        await queue.put(None)
        await collector_task

    elapsed = time.time() - t_start
    return analyze_results(label, all_results, elapsed)

async def collect_results(queue, all_results):
    while True:
        item = await queue.get()
        if item is None:
            break
        all_results.append(item)

def analyze_results(label, results, elapsed):
    if not results:
        print("  NO RESULTS!")
        return None

    total = len(results)
    successes = sum(1 for _, s, _, _ in results if 200 <= s < 500)
    failures = total - successes
    rps = total / elapsed if elapsed > 0 else 0

    rts_list = sorted([rt for _, s, _, _ in results if 200 <= s < 500 and rt > 0])

    def percentile(data, p):
        if not data:
            return 0
        idx = min(int(len(data) * p), len(data) - 1)
        return data[idx]

    avg_rt = statistics.mean(rts_list) if rts_list else 0
    p50 = percentile(rts_list, 0.50)
    p90 = percentile(rts_list, 0.90)
    p95 = percentile(rts_list, 0.95)
    p99 = percentile(rts_list, 0.99)
    p999 = percentile(rts_list, 0.999)
    max_rt = max(rts_list) if rts_list else 0

    # Status code breakdown
    status_counts = defaultdict(int)
    for _, s, _, _ in results:
        status_counts[s] += 1

    # Per-endpoint breakdown
    ep_stats = {}
    for path, status, rt, _ in results:
        if path not in ep_stats:
            ep_stats[path] = {"count": 0, "success": 0, "rts": [], "errors": 0}
        ep_stats[path]["count"] += 1
        if 200 <= status < 500:
            ep_stats[path]["success"] += 1
            ep_stats[path]["rts"].append(rt)
        else:
            ep_stats[path]["errors"] += 1

    print(f"\n  Results: {total} requests in {elapsed:.1f}s = {rps:.0f} RPS")
    print(f"  Success: {successes} ({successes/total*100:.1f}%)  Failures: {failures} ({failures/total*100:.1f}%)")
    print(f"  Latency: avg={avg_rt:.1f}ms p50={p50:.1f}ms p90={p90:.1f}ms p95={p95:.1f}ms p99={p99:.1f}ms p99.9={p999:.1f}ms max={max_rt:.1f}ms")

    print(f"\n  Status codes:")
    for code, count in sorted(status_counts.items(), key=lambda x: -x[1]):
        print(f"    {code:>4d}: {count:>6d} ({count/total*100:5.1f}%)")

    print(f"\n  Slowest endpoints (by p99):")
    ep_list = []
    for path, data in ep_stats.items():
        ep_rts = sorted(data["rts"])
        ep_p99 = percentile(ep_rts, 0.99) if ep_rts else 0
        ep_avg = statistics.mean(ep_rts) if ep_rts else 0
        ep_list.append((path, ep_avg, ep_p99, data["count"], data["success"], data["errors"]))

    for path, avg, p99, count, succ, errs in sorted(ep_list, key=lambda x: x[2], reverse=True)[:8]:
        path_short = path[:50]
        flag = " ⚠️ SLOW" if p99 > 100 else (" 🔴 VERY SLOW" if p99 > 500 else "")
        err_flag = f" {errs} errors" if errs > 0 else ""
        print(f"    {path_short:50s} avg={avg:6.1f}ms p99={p99:6.1f}ms n={count}{err_flag}{flag}")

    return {
        "label": label,
        "total": total, "successes": successes, "failures": failures,
        "rps": rps, "elapsed": elapsed,
        "avg_ms": avg_rt, "p50_ms": p50, "p90_ms": p90, "p95_ms": p95,
        "p99_ms": p99, "p999_ms": p999, "max_ms": max_rt,
        "status_counts": dict(status_counts),
        "ep_stats": ep_stats,
    }

async def main():
    import argparse
    parser = argparse.ArgumentParser(description="TikiClone Stress Test")
    parser.add_argument("--level", default="all", help="Test level or 'all'")
    parser.add_argument("--http-base", default=HTTP_BASE, help="Base URL")
    args = parser.parse_args()

    global HTTP_BASE
    HTTP_BASE = args.http_base

    print(f"TikiClone Stress Test")
    print(f"Target: {HTTP_BASE}")
    print(f"Time: {datetime.now().isoformat()}")

    all_results = []

    if args.level == "all":
        for users, duration, label in TEST_LEVELS:
            try:
                result = await run_test(users, duration, label)
                if result:
                    all_results.append(result)
            except KeyboardInterrupt:
                print("\nInterrupted!")
                break
            await asyncio.sleep(5)  # cooldown
    else:
        # Single level
        level_map = {lvl[2]: lvl for lvl in TEST_LEVELS}
        if args.level in level_map:
            users, duration, label = level_map[args.level]
            result = await run_test(users, duration, label)
            if result:
                all_results.append(result)
        else:
            print(f"Unknown level: {args.level}. Available: {[l[2] for l in TEST_LEVELS]}")

    # Summary
    print(f"\n\n{'='*60}")
    print(f"SUMMARY REPORT")
    print(f"{'='*60}")
    print(f"{'Level':>12} {'RPS':>7} {'Avg':>7} {'p90':>7} {'p95':>7} {'p99':>7} {'p99.9':>7} {'Max':>7} {'Fail%':>7}")
    for r in all_results:
        fail_pct = r["failures"] / r["total"] * 100 if r["total"] else 0
        status = "✅" if r["p999_ms"] < 10 and fail_pct < 1 else ("⚠️" if r["p999_ms"] < 100 and fail_pct < 5 else "❌")
        print(f"{r['label']:>12} {r['rps']:>7.0f} {r['avg_ms']:>6.1f}ms {r['p90_ms']:>6.1f}ms {r['p95_ms']:>6.1f}ms {r['p99_ms']:>6.1f}ms {r['p999_ms']:>6.1f}ms {r['max_ms']:>6.1f}ms {fail_pct:>6.1f}% {status}")

    # Identify bottleneck
    slow_endpoint = None
    for r in all_results:
        for path, data in r.get("ep_stats", {}).items():
            ep_rts = sorted(data["rts"])
            if ep_rts:
                p999 = ep_rts[min(int(len(ep_rts) * 0.999), len(ep_rts) - 1)]
                if p999 > 100:
                    slow_endpoint = (path, p999, data["count"])
                    break

    if slow_endpoint:
        print(f"\n🔴 BOTTLENECK: {slow_endpoint[0]} p99.9={slow_endpoint[1]:.0f}ms ({slow_endpoint[2]} requests)")
        print(f"   Fix: Add caching, optimize DB query, or add replicas")

    # Save results
    result_path = RESULTS_DIR / f"stress_test_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    with open(result_path, "w") as f:
        json.dump(all_results, f, indent=2, default=str)
    print(f"\nResults saved: {result_path}")

if __name__ == "__main__":
    asyncio.run(main())
