#!/usr/bin/env python3
"""
TikiClone Raw Load Test  (no external dependencies beyond aiohttp)
===================================================================
Pure asyncio load test that exercises Nginx → Gateway → Backend.
Runs all 5 TPS levels and reports bottlenecks.

Usage:
    python3 raw_load_test.py           # Run all levels
    python3 raw_load_test.py --level 100  # Single level
"""

import asyncio
import aiohttp
import ssl
import time
import json
import sys
import subprocess
import statistics
import re
from datetime import datetime
from pathlib import Path
from collections import defaultdict

# ═══════════════════════════════════════════════════════════════
# Config
# ═══════════════════════════════════════════════════════════════

RESULTS_DIR = Path(__file__).parent / "results"
RESULTS_DIR.mkdir(exist_ok=True)

HTTPS_BASE = "https://localhost:8443"
HTTP_BASE = "http://localhost:8080"
VERIFY_SSL = False

TPS_LEVELS = {
    10:    {"duration": 30,  "users": 1},
    100:   {"duration": 30,  "users": 10},
    1000:  {"duration": 60,  "users": 50},
    5000:  {"duration": 60,  "users": 100},
    10000: {"duration": 60,  "users": 200},
}

ENDPOINTS = [
    ("GET",  "/",                    10),  # Homepage (Next.js via nginx)
    ("GET",  "/api/v1/products",     25),  # Product list
    ("GET",  "/api/v1/categories",   15),  # Categories
    ("GET",  "/api/v1/products/spu-dien-thoai-001", 10),  # Product detail
    ("GET",  "/api/v1/cart",         5),   # Cart (will 401, that's fine)
    ("GET",  "/api/v1/orders",       3),   # Orders (will 401)
    ("GET",  "/api/v1/search?q=iPhone", 10),  # Search
    ("GET",  "/api/v1/inventory?product_id=spu-dien-thoai-001", 5),
    ("GET",  "/api/v1/checkout",     5),   # Checkout
    ("GET",  "/api/v1/payments",     2),   # Payments
]

MONITORED = [
    "tikiclone-web-tls-1", "tikiclone-gateway-1", "tikiclone-web-1",
    "tikiclone-product-1", "tikiclone-cart-1", "tikiclone-order-1",
    "tikiclone-payment-1", "tikiclone-inventory-1",
    "tikiclone-auth-1", "tikiclone-identity-auth-1",
    "tikiclone-mysql-primary-1", "tikiclone-redis-master-1",
    "tikiclone-mongodb-1", "tikiclone-kafka-1",
]

# ═══════════════════════════════════════════════════════════════
# Colors
# ═══════════════════════════════════════════════════════════════
class C:
    R="\033[91m"; G="\033[92m"; Y="\033[93m"; B="\033[94m"
    M="\033[95m"; CY="\033[96m"; BD="\033[1m"; RS="\033[0m"

def log(msg, lvl="INFO"):
    ts = datetime.now().strftime("%H:%M:%S")
    clr = {"INFO":C.B,"OK":C.G,"WARN":C.Y,"FAIL":C.R,"RES":C.CY,"HDR":C.M}.get(lvl,C.RS)
    print(f"{clr}[{ts}] {msg}{C.RS}")

# ═══════════════════════════════════════════════════════════════
# Container metrics
# ═══════════════════════════════════════════════════════════════

def get_container_stats():
    """Get docker stats for monitored containers."""
    names = " ".join(MONITORED)
    try:
        r = subprocess.run(
            f"docker stats --no-stream --format '{{.Name}}|{{.CPUPerc}}|{{.MemPerc}}|{{.PIDs}}' {names}",
            shell=True, capture_output=True, text=True, timeout=30
        )
        stats = {}
        for line in r.stdout.strip().split("\n"):
            if "|" not in line: continue
            parts = line.split("|")
            if len(parts) >= 4:
                try:
                    stats[parts[0].strip()] = {
                        "cpu": float(parts[1].strip().replace("%","") or 0),
                        "mem": float(parts[2].strip().replace("%","") or 0),
                        "pids": parts[3].strip(),
                    }
                except ValueError:
                    pass
        return stats
    except Exception:
        return {}

# ═══════════════════════════════════════════════════════════════
# Load generator
# ═══════════════════════════════════════════════════════════════

class LoadGenerator:
    def __init__(self):
        self.all_results = {}       # tps -> list of (endpoint, status, rt_ms)
        self.metrics_snapshots = {}

    async def _worker(self, session, endpoints, duration, results_queue):
        """Each worker picks random endpoints and fires requests."""
        end_time = time.time() + duration
        while time.time() < end_time:
            # Pick endpoint by weight
            total_w = sum(w for _, _, w in endpoints)
            pick = random.uniform(0, total_w)
            cum = 0
            method, path, _ = endpoints[0]
            for m, p, w in endpoints:
                cum += w
                if pick <= cum:
                    method, path = m, p
                    break

            url = HTTPS_BASE + path
            t0 = time.time()
            try:
                async with session.request(method, url, ssl=VERIFY_SSL, timeout=aiohttp.ClientTimeout(total=10)) as resp:
                    await resp.read()
                    status = resp.status
            except asyncio.TimeoutError:
                status = 0
            except Exception:
                status = -1
            rt = (time.time() - t0) * 1000  # ms
            await results_queue.put((path, status, rt))

    async def run_level(self, tps_target, num_users, duration):
        """Run concurrent users for a duration."""
        import random

        log(f"\n{'═'*60}", "HDR")
        log(f"LOAD TEST: Target ~{tps_target} TPS | Users: {num_users} | Duration: {duration}s", "HDR")
        log(f"{'═'*60}", "HDR")

        # Pre-test metrics
        log("Collecting pre-test container stats...", "INFO")
        pre_stats = get_container_stats()
        self.metrics_snapshots[f"tps{tps_target}_pre"] = pre_stats

        # Setup SSL context
        ssl_ctx = ssl.create_default_context()
        ssl_ctx.check_hostname = False
        ssl_ctx.verify_mode = ssl.CERT_NONE

        connector = aiohttp.TCPConnector(limit=num_users * 2, ssl=ssl_ctx)

        queue = asyncio.Queue()
        all_results = []

        t_start = time.time()

        async with aiohttp.ClientSession(connector=connector) as session:
            # Start workers
            workers = [
                asyncio.create_task(self._worker(session, ENDPOINTS, duration, queue))
                for _ in range(num_users)
            ]

            # Collect results while running
            collector_task = asyncio.create_task(self._collect_results(queue, all_results))

            # Mid-test snapshot
            await asyncio.sleep(duration / 2)
            mid_stats = get_container_stats()
            self.metrics_snapshots[f"tps{tps_target}_mid"] = mid_stats

            # Wait for all workers
            await asyncio.gather(*workers, return_exceptions=True)

            # Signal collector to stop
            await queue.put(None)
            await collector_task

        elapsed = time.time() - t_start

        # Post-test metrics
        await asyncio.sleep(2)
        post_stats = get_container_stats()
        self.metrics_snapshots[f"tps{tps_target}_post"] = post_stats

        # Analyse results
        return self._analyse_results(tps_target, all_results, pre_stats, post_stats, elapsed)

    async def _collect_results(self, queue, all_results):
        while True:
            item = await queue.get()
            if item is None:
                break
            all_results.append(item)

    def _analyse_results(self, tps_target, results, pre_stats, post_stats, elapsed):
        """Analyse and report results for one level."""
        if not results:
            log("No results collected!", "FAIL")
            return None

        total = len(results)
        successes = sum(1 for _, s, _ in results if 200 <= s < 500)
        failures = total - successes
        rps = total / elapsed if elapsed > 0 else 0

        rts_list = sorted([rt for _, s, rt in results if 200 <= s < 500 and rt > 0])

        avg_rt = statistics.mean(rts_list) if rts_list else 0
        p50 = rts_list[len(rts_list)//2] if rts_list else 0
        p90 = rts_list[int(len(rts_list)*0.9)] if rts_list else 0
        p95 = rts_list[int(len(rts_list)*0.95)] if rts_list else 0
        p99 = rts_list[int(len(rts_list)*0.99)] if rts_list else 0
        max_rt = max(rts_list) if rts_list else 0

        # Status code breakdown
        status_counts = defaultdict(int)
        for _, s, _ in results:
            status_counts[s] += 1

        # Per-endpoint breakdown
        ep_stats = {}
        for path, status, rt in results:
            if path not in ep_stats:
                ep_stats[path] = {"count": 0, "success": 0, "rts": []}
            ep_stats[path]["count"] += 1
            if 200 <= status < 500:
                ep_stats[path]["success"] += 1
                ep_stats[path]["rts"].append(rt)

        log(f"\n{'─'*60}", "RES")
        log(f"RESULTS: ~{tps_target} TPS target ({elapsed:.1f}s elapsed)", "RES")
        log(f"{'─'*60}", "RES")
        log(f"  Total requests:  {total:>8d}  |  Actual RPS: {rps:.1f}", "RES")
        log(f"  Success (2xx):   {successes:>8d}  |  Failures:   {failures} ({failures/total*100:.1f}%)" if total else "", "RES")
        log(f"  Avg RT: {avg_rt:.0f}ms  |  p50: {p50:.0f}ms  |  p90: {p90:.0f}ms  |  p95: {p95:.0f}ms  |  p99: {p99:.0f}ms  |  max: {max_rt:.0f}ms", "RES")

        log(f"\n  Status codes:", "INFO")
        for code, count in sorted(status_counts.items(), key=lambda x: -x[1]):
            pct = count / total * 100
            log(f"    {code:>4d}: {count:>6d} ({pct:5.1f}%)", "INFO")

        log(f"\n  Per-endpoint breakdown:", "RES")
        for path, data in sorted(ep_stats.items(), key=lambda x: statistics.mean(x[1]["rts"]) if x[1]["rts"] else 0, reverse=True):
            ep_rts = sorted(data["rts"])
            ep_avg = statistics.mean(ep_rts) if ep_rts else 0
            ep_p99 = ep_rts[int(len(ep_rts)*0.99)] if ep_rts else 0
            ep_succ = data["success"]
            ep_count = data["count"]
            clr = "OK" if ep_avg < 500 else ("WARN" if ep_avg < 2000 else "FAIL")
            path_short = path[:45]
            log(f"    {path_short:45s} avg={ep_avg:6.0f}ms p99={ep_p99:6.0f}ms n={ep_count} ok={ep_succ}", clr)

        # Container resource delta
        log(f"\n  Container resources (pre → post):", "RES")
        hot_services = []
        for svc in MONITORED:
            pre = pre_stats.get(svc, {}).get("cpu", 0)
            post = post_stats.get(svc, {}).get("cpu", 0)
            pre_mem = pre_stats.get(svc, {}).get("mem", 0)
            post_mem = post_stats.get(svc, {}).get("mem", 0)
            delta = post - pre
            hot = " ← HOT" if post > 50 or delta > 20 else ""
            if post > 30 or delta > 15:
                hot_services.append((svc, post, post_mem, delta))
            clr = "INFO"
            if post > 80: clr = "FAIL"
            elif post > 50: clr = "WARN"
            log(f"    {svc:35s} CPU: {pre:5.1f}% → {post:5.1f}% (Δ{delta:+.1f}%)  Mem: {post_mem:.1f}%{hot}", clr)

        result = {
            "tps_target": tps_target,
            "total": total, "successes": successes, "failures": failures,
            "rps": rps, "elapsed": elapsed,
            "avg_ms": avg_rt, "p50_ms": p50, "p90_ms": p90, "p95_ms": p95, "p99_ms": p99,
            "max_ms": max_rt,
            "status_counts": dict(status_counts),
            "hot_services": hot_services,
        }
        self.all_results[tps_target] = result
        return result

    def generate_report(self):
        """Final comprehensive bottleneck report."""
        log(f"\n\n{C.BD}{'═'*70}")
        log("  TIKI-CLONE BOTTLENECK ANALYSIS REPORT")
        log(f"{'═'*70}{C.RS}")

        lines = [
            f"{'═'*70}",
            "TIKI-CLONE LOAD TEST / BOTTLENECK ANALYSIS REPORT",
            f"Generated: {datetime.now().isoformat()}",
            f"Entry point: {HTTPS_BASE} (Nginx TLS) → Gateway → Backend",
            f"{'═'*70}",
            "",
            "── RPS SCALING SUMMARY ───────────────────────────────────",
            f"{'Target':>8} {'Actual':>8} {'Avg RT':>8} {'p90 RT':>8} {'p95 RT':>8} {'p99 RT':>8} {'Max RT':>8} {'Fail%':>7}  Status",
        ]

        bottleneck_tps = None
        prev_avg = None

        for tps in sorted(self.all_results.keys()):
            r = self.all_results.get(tps)
            if not r:
                lines.append(f"{tps:>8} {'N/A':>8}")
                continue

            fail_pct = r["failures"] / r["total"] * 100 if r["total"] else 0
            status = "✓"
            if r["avg_ms"] > 2000 or fail_pct > 10:
                status = "✗ BOTTLENECK"
                if bottleneck_tps is None:
                    bottleneck_tps = tps
            elif r["avg_ms"] > 1000 or fail_pct > 5:
                status = "△ WARN"
            elif r["avg_ms"] > 500:
                status = "~ OK"

            lines.append(
                f"{tps:>8} {r['rps']:>8.1f} {r['avg_ms']:>7.0f}ms {r['p90_ms']:>7.0f}ms "
                f"{r['p95_ms']:>7.0f}ms {r['p99_ms']:>7.0f}ms {r['max_ms']:>7.0f}ms {fail_pct:>6.1f}%  {status}"
            )

        lines.append("")
        lines.append("── CONTAINER RESOURCE USAGE AT EACH LEVEL ───────────────")

        for tps in sorted(self.all_results.keys()):
            post = self.metrics_snapshots.get(f"tps{tps}_post", {})
            lines.append(f"\n  At {tps} TPS:")
            lines.append(f"  {'Service':35s} {'CPU%':>7s} {'Mem%':>7s} {'PIDs':>6s}")
            for svc in MONITORED:
                data = post.get(svc, {})
                cpu = data.get("cpu", 0)
                mem = data.get("mem", 0)
                hot = " ←" if cpu > 50 or mem > 80 else ""
                lines.append(f"  {svc:35s} {cpu:>6.1f}% {mem:>6.1f}% {data.get('pids',''):>6s}{hot}")

        # Bottleneck diagnosis
        lines.append("")
        lines.append("── BOTTLENECK DIAGNOSIS ────────────────────────────────")

        if bottleneck_tps:
            r = self.all_results[bottleneck_tps]
            post = self.metrics_snapshots.get(f"tps{bottleneck_tps}_post", {})

            lines.append(f"  System becomes bottlenecked at ~{bottleneck_tps} TPS")
            lines.append(f"  At {tps} TPS: avg={r['avg_ms']:.0f}ms, p99={r['p99_ms']:.0f}ms, fail={r['failures']/r['total']*100:.1f}%")
            lines.append("")

            # Sort services by CPU
            svc_cpu = [(svc, d.get("cpu", 0), d.get("mem", 0))
                       for svc, d in post.items()]
            svc_cpu.sort(key=lambda x: x[1], reverse=True)

            lines.append("  Top resource consumers:")
            for svc, cpu, mem in svc_cpu[:8]:
                marker = " ← PRIMARY BOTTLENECK" if cpu == svc_cpu[0][1] and cpu > 50 else ""
                lines.append(f"    {svc:35s} CPU: {cpu:5.1f}%  Mem: {mem:5.1f}%{marker}")

            if svc_cpu and svc_cpu[0][1] > 50:
                primary = svc_cpu[0][0]
                primary_cpu = svc_cpu[0][1]
                lines.append(f"\n  ► PRIMARY BOTTLENECK: {primary} ({primary_cpu:.1f}% CPU)")

                # Pattern-based diagnosis
                gw_cpu = post.get("tikiclone-gateway-1", {}).get("cpu", 0)
                nginx_cpu = post.get("tikiclone-web-tls-1", {}).get("cpu", 0)
                mysql_cpu = post.get("tikiclone-mysql-primary-1", {}).get("cpu", 0)
                redis_cpu = post.get("tikiclone-redis-master-1", {}).get("cpu", 0)
                web_cpu = post.get("tikiclone-web-1", {}).get("cpu", 0)
                product_cpu = post.get("tikiclone-product-1", {}).get("cpu", 0)

                db_services = [
                    ("tikiclone-mysql-primary-1", mysql_cpu),
                    ("tikiclone-redis-master-1", redis_cpu),
                    ("tikiclone-mongodb-1", post.get("tikiclone-mongodb-1", {}).get("cpu", 0)),
                ]
                db_services.sort(key=lambda x: x[1], reverse=True)

                if mysql_cpu > 60 and mysql_cpu > product_cpu:
                    lines.append(f"\n    → MySQL is the limiting factor ({mysql_cpu:.1f}% CPU)")
                    lines.append(f"      - Add read replicas for product listing queries")
                    lines.append(f"      - Add connection pooling (max_connections in MySQL)")
                    lines.append(f"      - Add database-level query cache / slow query log analysis")
                elif nginx_cpu > 60:
                    lines.append(f"\n    → Nginx/TLS is the limiting factor ({nginx_cpu:.1f}% CPU)")
                    lines.append(f"      - Add nginx replicas behind a load balancer")
                    lines.append(f"      - Enable keep-alive connections to upstream")
                    lines.append(f"      - Consider TLS session resumption, OCSP stapling")
                elif gw_cpu > 60:
                    lines.append(f"\n    → API Gateway is the limiting factor ({gw_cpu:.1f}% CPU)")
                    lines.append(f"      - Horizontal scale: add gateway replicas")
                    lines.append(f"      - Add response caching at gateway layer")
                    lines.append(f"      - Review middleware chain (each request passes ~15 handlers)")
                elif web_cpu > 60:
                    lines.append(f"\n    → Next.js SSR is the limiting factor ({web_cpu:.1f}% CPU)")
                    lines.append(f"      - Enable incremental static regeneration (ISR)")
                    lines.append(f"      - Move to edge CDN for static pages")
                    lines.append(f"      - Profile SSR rendering time")
                elif product_cpu > 60:
                    lines.append(f"\n    → Product service is the limiting factor ({product_cpu:.1f}% CPU)")
                    lines.append(f"      - Add product service replicas")
                    lines.append(f"      - Cache product data in Redis")
                elif redis_cpu > 50:
                    lines.append(f"\n    → Redis is under pressure ({redis_cpu:.1f}% CPU)")
                    lines.append(f"      - Add Redis cluster/sentinel")
                    lines.append(f"      - Review TTLs and eviction policies")

                if db_services[0][1] > 40:
                    lines.append(f"\n    → Database tier is under pressure:")
                    for db_svc, db_cpu in db_services[:3]:
                        if db_cpu > 10:
                            lines.append(f"      {db_svc}: {db_cpu:.1f}% CPU")
        else:
            lines.append("  ✓ No clear bottleneck detected within tested TPS range")

        lines.append("")
        lines.append("── RECOMMENDATIONS ─────────────────────────────────────")
        lines.append("  1. Horizontal scale the primary bottleneck service")
        lines.append("  2. Add Redis caching for product listing (most hit endpoint)")
        lines.append("  3. Enable connection pooling to MySQL")
        lines.append("  4. Add nginx replicas if TLS termination is the limit")
        lines.append("  5. Profile slow queries: SET GLOBAL slow_query_log = 'ON'")
        lines.append("  6. Add CDN for static assets (product images, JS bundles)")
        lines.append("  7. Enable HTTP/2 on nginx for multiplexing")
        lines.append(f"\n{'══'*70}")

        report = "\n".join(lines)

        # Save
        report_path = RESULTS_DIR / f"bottleneck_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"
        report_path.write_text(report)
        print(report)
        log(f"\nReport saved: {report_path}", "OK")

        return report


# ═══════════════════════════════════════════════════════════════
# Main
# ═══════════════════════════════════════════════════════════════

async def main():
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("--level", default="all", help="TPS level or 'all'")
    parser.add_argument("--quick", action="store_true", help="Quick test (10 + 100 only)")
    args = parser.parse_args()

    log(f"{C.BD}{'═'*60}")
    log(f"  TikiClone Load Test Suite")
    log(f"  Entry: {HTTPS_BASE} (Nginx → Gateway → Backend)")
    log(f"{'═'*60}{C.RS}")

    gen = LoadGenerator()

    if args.level == "all":
        if args.quick:
            levels = [10, 100]
        else:
            levels = [10, 100, 1000, 5000, 10000]
    else:
        levels = [int(args.level)]

    for i, tps in enumerate(levels):
        cfg = TPS_LEVELS[tps]
        try:
            await gen.run_level(tps, cfg["users"], cfg["duration"])
        except KeyboardInterrupt:
            log("Interrupted!", "WARN")
            break

        # Cooldown
        if i < len(levels) - 1:
            log(f"\nCooling down 10s...", "INFO")
            await asyncio.sleep(10)

    gen.generate_report()

if __name__ == "__main__":
    import random
    asyncio.run(main())
