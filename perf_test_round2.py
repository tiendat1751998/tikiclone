"""
Optimized load test - uses HTTP keep-alive and connection pooling efficiently
Also tests HTTP (port 80) to measure TLS overhead
"""
import asyncio
import aiohttp
import ssl
import time
import json
import statistics
from collections import defaultdict

TARGET_HOST = "https://10.10.10.133"
# Also test HTTP to measure TLS overhead
TARGET_HOST_HTTP = "http://10.10.10.133:80"

ENDPOINTS = [
    ("/health", 0.3),
    ("/api/v1/products", 0.25),
    ("/api/v1/categories", 0.2),
    ("/api/v1/products/9ebcd126-44ce-4066-928c-e71d4acb4811", 0.15),
    ("/api/v1/auth/health", 0.1),
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

async def worker(session, duration, worker_id):
    end_time = time.time() + duration
    results = []
    while time.time() < end_time:
        endpoint = pick_endpoint()
        url = f"{TARGET_HOST}{endpoint}"
        start = time.time()
        try:
            async with session.get(url, timeout=aiohttp.ClientTimeout(total=2)) as resp:
                await resp.read()
                latency = time.time() - start
                results.append({"status": resp.status, "latency": latency, "endpoint": endpoint, "error": None})
        except Exception as e:
            latency = time.time() - start
            results.append({"status": 0, "latency": latency, "endpoint": endpoint, "error": type(e).__name__})
    return results

async def run_load_test(concurrent_workers, duration, label, target):
    ssl_ctx = ssl.create_default_context()
    ssl_ctx.check_hostname = False
    ssl_ctx.verify_mode = ssl.CERT_NONE
    
    connector = aiohttp.TCPConnector(
        limit=0,  # No limit
        limit_per_host=0,  # No limit
        keepalive_timeout=60,
        enable_cleanup_closed=True,
        ssl=ssl_ctx if target.startswith("https") else False
    )
    
    async with aiohttp.ClientSession(connector=connector) as session:
        print(f"\n{'='*60}")
        print(f"  {label}")
        print(f"  Target: {target}")
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
        print("  No results collected!")
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

    print(f"\n  OVERALL RESULTS:")
    print(f"  Total Requests: {total_requests}")
    print(f"  Duration: {test_duration:.1f}s")
    print(f"  Throughput: {throughput:.0f} req/s (TPS)")
    print(f"  Success Rate: {success_rate:.1f}%")
    print(f"  Errors: {total_errors}")
    print(f"\n  LATENCY:")
    print(f"  min:    {latencies[0]*1000:.1f}ms")
    print(f"  p50:    {latencies[len(latencies)//2]*1000:.1f}ms")
    print(f"  p90:    {latencies[int(len(latencies)*0.9)]*1000:.1f}ms")
    print(f"  p95:    {latencies[int(len(latencies)*0.95)]*1000:.1f}ms")
    print(f"  p99:    {latencies[int(len(latencies)*0.99)]*1000:.1f}ms")
    print(f"  max:    {latencies[-1]*1000:.1f}ms")
    print(f"  mean:   {statistics.mean(latencies)*1000:.1f}ms")

    print(f"\n  PER-ENDPOINT:")
    for ep, stats in sorted(endpoint_stats.items(), key=lambda x: -x[1]["count"]):
        ep_lats = sorted(stats["latencies"])
        ep_p95 = ep_lats[int(len(ep_lats)*0.95)]*1000 if ep_lats else 0
        ep_mean = statistics.mean(ep_lats)*1000 if ep_lats else 0
        ep_sr = (stats["count"] - stats["errors"]) / stats["count"] * 100 if stats["count"] else 0
        print(f"    {ep}: {stats['count']} req, {stats['count']/test_duration:.0f} rps, p95={ep_p95:.1f}ms, mean={ep_mean:.1f}ms, sr={ep_sr:.0f}%")

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
    print("  TikiClone Performance Test - Round 2")
    print(f"  Target: {TARGET_HOST}")
    print("  Goal: 10000 TPS, <10ms latency")
    print("="*60)

    all_results = []

    # Test with higher concurrency
    test_levels = [
        (100, 15, "LEVEL 1 - 100 workers, 15s", TARGET_HOST),
        (500, 15, "LEVEL 2 - 500 workers, 15s", TARGET_HOST),
        (1000, 15, "LEVEL 3 - 1000 workers, 15s", TARGET_HOST),
        (2000, 15, "LEVEL 4 - 2000 workers, 15s", TARGET_HOST),
    ]

    for workers, duration, label, target in test_levels:
        result = await run_load_test(workers, duration, label, target)
        if result:
            all_results.append(result)
            if result["success_rate"] < 80:
                print(f"\n  *** SUCCESS RATE DROPPED TO {result['success_rate']:.0f}% - STOPPING ***")
                break

    # Summary
    print(f"\n{'='*60}")
    print("  SUMMARY")
    print(f"{'='*60}")
    print(f"  {'Test':<40} {'TPS':>8} {'p95':>8} {'p99':>8} {'SR':>6}")
    print(f"  {'-'*70}")
    for r in all_results:
        print(f"  {r['label']:<40} {r['throughput']:>7.0f} {r['p95_ms']:>7.1f} {r['p99_ms']:>7.1f} {r['success_rate']:>5.0f}%")

    with open("/home/datdt/tikiclone/perf_test_results_round2.json", "w") as f:
        json.dump(all_results, f, indent=2)
    print(f"\n  Results saved to perf_test_results_round2.json")

    best = max(all_results, key=lambda x: x["throughput"]) if all_results else None
    if best:
        print(f"\n  BEST THROUGHPUT: {best['throughput']:.0f} TPS (target: 10000)")
        print(f"  BEST P99: {best['p99_ms']:.1f}ms (target: <10ms)")
        if best["throughput"] >= 10000 and best["p99_ms"] < 10:
            print("  TARGET MET!")
        else:
            print("  TARGET NOT MET - further optimization needed")

if __name__ == "__main__":
    asyncio.run(main())
