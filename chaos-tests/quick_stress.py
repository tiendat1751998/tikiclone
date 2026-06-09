#!/usr/bin/env python3
"""Quick stress test for TikiClone APIs via Swarm overlay network."""
import urllib.request, json, time, concurrent.futures, statistics, sys

BASE = "http://tiki_gateway:8080"

ENDPOINTS = [
    ("/api/v1/products", 25),
    ("/api/v1/products?page=1&size=20", 20),
    ("/api/v1/categories", 15),
    ("/api/v1/categories/tree", 10),
    ("/api/v1/products/search?q=dien+thoai", 10),
    ("/api/v1/products/search?q=laptop", 5),
    ("/api/v1/products/spu-dien-thoai-001", 5),
    ("/api/v1/cart", 3),
    ("/api/v1/orders", 2),
    ("/api/v1/inventory?product_id=spu-dien-thoai-001", 3),
    ("/api/v1/checkout", 2),
]

def hit(path):
    t0 = time.time()
    try:
        req = urllib.request.urlopen(BASE + path, timeout=5)
        status = req.status
        body = req.read()[:200]
    except Exception as e:
        status = -1
        body = str(e).encode()
    rt = (time.time() - t0) * 1000
    return (path, status, rt, body)

def run_test(num_requests, concurrency, label):
    print(f"\n{'='*70}")
    print(f"TEST: {label} | {num_requests} requests | {concurrency} concurrent")
    print(f"{'='*70}")

    results = []
    t_start = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as ex:
        futures = [ex.submit(hit, ENDPOINTS[i % len(ENDPOINTS)][0]) for i in range(num_requests)]
        for f in concurrent.futures.as_completed(futures):
            results.append(f.result())
    elapsed = time.time() - t_start

    total = len(results)
    successes = sum(1 for _, s, _, _ in results if 200 <= s < 500)
    failures = total - successes
    rps = total / elapsed

    rts = sorted([rt_val for _, s, rt_val, _ in results if 200 <= s < 500 and rt_val > 0])
    avg_rt = statistics.mean(rts) if rts else 0
    p50 = rts[len(rts)//2] if rts else 0
    p90 = rts[int(len(rts)*0.9)] if rts else 0
    p95 = rts[int(len(rts)*0.95)] if rts else 0
    p99 = rts[int(len(rts)*0.99)] if rts else 0
    p999 = rts[min(int(len(rts)*0.999), len(rts)-1)] if rts else 0
    max_rt = max(rts) if rts else 0

    print(f"  {total} requests in {elapsed:.1f}s = {rps:.0f} RPS")
    print(f"  Success: {successes}/{total} ({successes/total*100:.1f}%)  Failures: {failures}")
    print(f"  Latency: avg={avg_rt:.1f}ms p50={p50:.1f}ms p90={p90:.1f}ms p95={p95:.1f}ms p99={p99:.1f}ms p99.9={p999:.1f}ms max={max_rt:.1f}ms")

    # Status breakdown
    status_counts = {}
    for _, s, _, _ in results:
        status_counts[s] = status_counts.get(s, 0) + 1
    print(f"  Status: {dict(sorted(status_counts.items()))}")

    # Per-endpoint
    eps = {}
    for path, status, rt, _ in results:
        if path not in eps: eps[path] = {"ok": 0, "fail": 0, "rts": []}
        if 200 <= status < 500:
            eps[path]["ok"] += 1
            eps[path]["rts"].append(rt)
        else:
            eps[path]["fail"] += 1

    print(f"\n  Per-endpoint (sorted by p99):")
    for path, d in sorted(eps.items(), key=lambda x: x[1]["rts"][-1] if x[1]["rts"] else 0, reverse=True)[:10]:
        if d["rts"]:
            ep_p99 = d["rts"][int(len(d["rts"])*0.99)]
            ep_avg = statistics.mean(d["rts"])
            flag = " ⚠️" if ep_p99 > 50 else ""
            print(f"    {path[:50]:50s} avg={ep_avg:6.1f}ms p99={ep_p99:6.1f}ms ok={d['ok']:>4} fail={d['fail']}{flag}")
        else:
            print(f"    {path[:50]:50s} ALL FAILED fail={d['fail']}")

    return {
        "label": label, "rps": rps, "total": total, "successes": successes, "failures": failures,
        "avg_ms": avg_rt, "p99_ms": p99, "p999_ms": p999, "max_ms": max_rt,
    }

if __name__ == "__main__":
    print(f"TikiClone Stress Test via Swarm overlay")
    print(f"Gateway: {BASE}")

    # Check gateway health
    try:
        req = urllib.request.urlopen(f"{BASE}/health", timeout=5)
        print(f"Gateway health: {req.status}")
    except Exception as e:
        print(f"Gateway unreachable: {e}")
        sys.exit(1)

    all_results = []
    for concurrency, num_requests, label in [
        (5, 50, "warmup"),
        (10, 100, "100_rps"),
        (20, 200, "200_rps"),
        (50, 500, "500_rps"),
    ]:
        r = run_test(num_requests, concurrency, label)
        all_results.append(r)
        time.sleep(2)

    # Summary
    print(f"\n\n{'='*70}")
    print(f"SUMMARY")
    print(f"{'='*70}")
    print(f"{'Level':>12} {'RPS':>7} {'Avg':>7} {'p99':>7} {'p99.9':>7} {'Max':>7} {'Fail%':>7}")
    for r in all_results:
        fail_pct = r["failures"] / r["total"] * 100 if r["total"] else 0
        target_met = "✅" if r["p999_ms"] < 10 and fail_pct < 1 else ("⚠️" if r["p999_ms"] < 100 and fail_pct < 5 else "❌")
        print(f"{r['label']:>12} {r['rps']:>7.0f} {r['avg_ms']:>6.1f}ms {r['p99_ms']:>6.1f}ms {r['p999_ms']:>6.1f}ms {r['max_ms']:>6.1f}ms {fail_pct:>6.1f}% {target_met}")
