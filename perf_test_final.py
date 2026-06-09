"""
TikiClone Performance Test - Optimized for high throughput
Target: 1000 req/s, <20ms p99 latency
Uses internal HTTP (port 80) to bypass TLS overhead.
"""
import asyncio
import aiohttp
import time
import json
import statistics
import sys
from collections import defaultdict

# Use internal HTTP endpoint to bypass TLS overhead
TARGET_HOST = "http://10.10.10.133"  # Traefik on port 80, no TLS

# Weighted endpoints reflecting real traffic patterns
ENDPOINTS = [
    ("/health", 0.05),
    ("/api/v1/products", 0.30),
    ("/api/v1/categories", 0.20),
    ("/api/v1/products?page=1&limit=20", 0.15),
    ("/api/v1/products?page=2&limit=20", 0.10),
    ("/api/v1/products?sort=price_asc&page=1&limit=20", 0.10),
    ("/api/v1/products?sort=price_desc&page=1&limit=20", 0.10),
]

def pick_endpoint():
    import random
    r = random.random()
    cumulative = 0
    for endpoint, weight in ENDPOINTS:
        cumulative += weight
        if r <= cumulative:
            return endpoint
    return ENDPOINTS[-1][0]

async def worker(session, duration, worker_id, warmup=False):
    end_time = time.time() + duration
    results = []
    while time.time() < end_time:
        endpoint = pick_endpoint()
        url = f"{TARGET_HOST}{endpoint}"
        start = time.time()
        try:
            async with session.get(url, timeout=aiohttp.ClientTimeout(total=5)) as resp:
                await resp.read()
                latency = time.time() - start
                results.append({"status": resp.status, "latency": latency, "endpoint": endpoint, "error": None})
        except Exception as e:
            latency = time.time() - start
            results.append({"status": 0, "latency": latency, "endpoint": endpoint, "error": type(e).__name__})
    return results

async def warmup(session, duration=5):
    """Warm up connections and caches"""
    print("  Warming up ({}s)...".format(duration))
    tasks = [worker(session, duration, i, warmup=True) for i in range(50)]
    await asyncio.gather(*tasks)
    print("  Warmup complete.")

async def run_load_test(concurrent_workers, duration, label, do_warmup=True):
    connector = aiohttp.TCPConnector(
        limit=concurrent_workers * 2,  # Allow some queueing
        limit_per_host=concurrent_workers,
        keepalive_timeout=60,
        enable_cleanup_closed=True,
    )
    
    async with aiohttp.ClientSession(connector=connector) as session:
        if do_warmup:
            await warmup(session, duration=5)
        
        print(f"\n{'='*60}")
        print(f"  {label}")
        print(f"  Workers: {concurrent_workers} | Duration: {duration}s")
        print(f"{'='*60}")
        
        test_start = time.time()
        tasks = [worker(session, duration, i) for i in range(concurrent_workers)]
        all_worker_results = await asyncio.gather(*tasks)
        test_duration = time.time() - test_start

    results = []
    for wr in all_worker_results:
        results.extend(wr)

    total_requests = len(results)
    if total_requests == 0:
        print("  No results!")
        return None

    total_errors = sum(1 for r in results if r["error"] or r["status"] >= 400)
    latencies = sorted([r["latency"] for r in results])
    throughput = total_requests / test_duration
    success_rate = (total_requests - total_errors) / total_requests * 100

    endpoint_stats = defaultdict(lambda: {"count": 0, "latencies": [], "errors": 0})
    for r in results:
        ep = r["endpoint"]
        endpoint_stats[ep]["count"] += 1
        endpoint_stats[ep]["latencies"].append(r["latency"])
        if r["error"] or r["status"] >= 400:
            endpoint_stats[ep]["errors"] += 1

    print(f"\n  OVERALL:")
    print(f"  Requests: {total_requests} | Duration: {test_duration:.1f}s")
    print(f"  Throughput: {throughput:.0f} req/s | Success: {success_rate:.1f}%")
    print(f"  LATENCY: p50={latencies[len(latencies)//2]*1000:.1f}ms "
          f"p90={latencies[int(len(latencies)*0.9)]*1000:.1f}ms "
          f"p95={latencies[int(len(latencies)*0.95)]*1000:.1f}ms "
          f"p99={latencies[int(len(latencies)*0.99)]*1000:.1f}ms")
    print(f"  min={latencies[0]*1000:.1f}ms max={latencies[-1]*1000:.1f}ms mean={statistics.mean(latencies)*1000:.1f}ms")

    for ep, stats in sorted(endpoint_stats.items(), key=lambda x: -x[1]["count"]):
        ep_lats = sorted(stats["latencies"])
        p50 = ep_lats[len(ep_lats)//2]*1000 if ep_lats else 0
        p95 = ep_lats[int(len(ep_lats)*0.95)]*1000 if ep_lats else 0
        p99 = ep_lats[int(len(ep_lats)*0.99)]*1000 if ep_lats else 0
        ep_sr = (stats["count"] - stats["errors"]) / stats["count"] * 100 if stats["count"] else 0
        print(f"  {ep:<50} {stats['count']/test_duration:>6.0f} rps  "
              f"p50={p50:.1f} p95={p95:.1f} p99={p99:.1f} sr={ep_sr:.0f}%")

    return {
        "label": label, "workers": concurrent_workers,
        "total_requests": total_requests, "throughput": throughput,
        "success_rate": success_rate, "errors": total_errors,
        "p50_ms": latencies[len(latencies)//2]*1000,
        "p90_ms": latencies[int(len(latencies)*0.9)]*1000,
        "p95_ms": latencies[int(len(latencies)*0.95)]*1000,
        "p99_ms": latencies[int(len(latencies)*0.99)]*1000,
        "mean_ms": statistics.mean(latencies)*1000,
    }

async def main():
    print("="*60)
    print("  TikiClone Performance Test")
    print(f"  Target: {TARGET_HOST}")
    print("  Goal: 1000 req/s, <20ms p99 latency")
    print("="*60)

    all_results = []
    
    # Progressive load levels with warmup on first run
    test_levels = [
        (50, 15, "50 workers"),
        (100, 15, "100 workers"),
        (200, 20, "200 workers"),
        (300, 20, "300 workers"),
        (500, 20, "500 workers"),
        (800, 20, "800 workers"),
        (1000, 20, "1000 workers"),
    ]

    first = True
    for workers, duration, label in test_levels:
        result = await run_load_test(workers, duration, label, do_warmup=first)
        first = False
        if result:
            all_results.append(result)
            if result["success_rate"] < 90:
                print(f"\n  *** SR dropped to {result['success_rate']:.0f}% - stopping ***")
                break
            # Check if we've hit the target
            if result["throughput"] >= 1000 and result["p99_ms"] < 20:
                print(f"\n  *** TARGET MET: {result['throughput']:.0f} req/s, p99={result['p99_ms']:.1f}ms ***")
                # Push harder to find the ceiling
                if workers < 1000:
                    continue

    print(f"\n{'='*60}")
    print("  FINAL SUMMARY")
    print(f"{'='*60}")
    print(f"  {'Test':<20} {'TPS':>8} {'p50':>8} {'p95':>8} {'p99':>8} {'SR':>6}")
    print(f"  {'-'*60}")
    for r in all_results:
        print(f"  {r['label']:<20} {r['throughput']:>7.0f} {r['p50_ms']:>7.1f} {r['p95_ms']:>7.1f} {r['p99_ms']:>7.1f} {r['success_rate']:>5.0f}%")

    output_path = "/home/datdt/tikiclone/perf_test_final.json"
    with open(output_path, "w") as f:
        json.dump(all_results, f, indent=2)
    print(f"\n  Results saved to {output_path}")

    best = max(all_results, key=lambda x: x["throughput"]) if all_results else None
    if best:
        print(f"\n  BEST: {best['throughput']:.0f} req/s, p99={best['p99_ms']:.1f}ms, SR={best['success_rate']:.0f}%")
        if best["throughput"] >= 1000 and best["p99_ms"] < 20:
            print("  ✓ TARGET MET: 1000 req/s with <20ms p99 latency!")
        elif best["throughput"] >= 1000:
            print(f"  - TPS target met, but latency p99={best['p99_ms']:.1f}ms (need <20ms)")
        else:
            print(f"  - Still need {1000-best['throughput']:.0f} more req/s")

if __name__ == "__main__":
    asyncio.run(main())
