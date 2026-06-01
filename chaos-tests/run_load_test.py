#!/usr/bin/env python3
"""
TikiClone Load Test Orchestrator
=================================
Runs locust at 5 TPS levels (10, 100, 1000, 5000, 10000) and produces
a comprehensive bottleneck analysis report.

Usage:
    python3 run_load_test.py                  # Run all 5 levels
    python3 run_load_test.py --level 100      # Run single level
    python3 run_load_test.py --level all      # Run all (default)
"""

import subprocess
import json
import time
import sys
import os
import signal
import re
from datetime import datetime
from pathlib import Path

# ═══════════════════════════════════════════════════════════════
# Configuration
# ═══════════════════════════════════════════════════════════════

CHAOS_DIR = Path(__file__).parent
RESULTS_DIR = CHAOS_DIR / "results"
RESULTS_DIR.mkdir(exist_ok=True)

# TPS levels: (target_tps, users, spawn_rate, duration_seconds, user_classes)
TPS_LEVELS = {
    10:    {"users": 1,     "spawn_rate": 1,     "duration": "30s",  "classes": "TikiBrowseUser"},
    100:   {"users": 10,    "spawn_rate": 5,     "duration": "30s",  "classes": "TikiBrowseUser"},
    1000:  {"users": 100,   "spawn_rate": 50,    "duration": "60s",  "classes": "TikiBrowseUser"},
    5000:  {"users": 500,   "spawn_rate": 200,   "duration": "60s",  "classes": "TikiStressUser"},
    10000: {"users": 1000,  "spawn_rate": 500,   "duration": "60s",  "classes": "TikiStressUser"},
}

# Services to monitor
MONITORED_SERVICES = [
    "tikiclone-web-tls-1",      # Nginx entry point
    "tikiclone-gateway-1",        # API gateway (primary)
    "tikiclone-gateway-2-1",      # API gateway (secondary)
    "tikiclone-web-1",            # Next.js frontend
    "tikiclone-product-1",        # Product service (heavily used)
    "tikiclone-cart-1",           # Cart service
    "tikiclone-order-1",          # Order service
    "tikiclone-payment-1",        # Payment service
    "tikiclone-inventory-1",      # Inventory service
    "tikiclone-auth-1",           # Auth service
    "tikiclone-identity-auth-1",  # Identity auth
    "tikiclone-mysql-primary-1",  # MySQL
    "tikiclone-redis-master-1",   # Redis
    "tikiclone-mongodb-1",        # MongoDB
    "tikiclone-kafka-1",          # Kafka
    "ollama",                     # Ollama
]

# Colors
class C:
    R = "\033[91m"; G = "\033[92m"; Y = "\033[93m"
    B = "\033[94m"; M = "\033[95m"; CY = "\033[96m"
    BD = "\033[1m"; RS = "\033[0m"

def log(msg, level="INFO"):
    ts = datetime.now().strftime("%H:%M:%S")
    color = {"INFO":C.B, "OK":C.G, "WARN":C.Y, "FAIL":C.R, "RES":C.CY}.get(level, C.RS)
    print(f"{color}[{ts}] {msg}{C.RS}")

def run_cmd(cmd, timeout=120):
    try:
        r = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=timeout)
        return r.stdout.strip(), r.returncode
    except subprocess.TimeoutExpired:
        return "TIMEOUT", -1
    except Exception as e:
        return str(e), -1

# ═══════════════════════════════════════════════════════════════
# Container Metrics Collector
# ═══════════════════════════════════════════════════════════════

class ContainerMetrics:
    """Collects Docker container performance metrics."""

    @staticmethod
    def get_all_stats():
        """Get CPU, memory, network I/O for all monitored containers."""
        names = " ".join(MONITORED_SERVICES)
        stdout, _ = run_cmd(
            f"docker stats --no-stream --format "
            "'{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}|{{.PIDs}}' "
            f"{names}", timeout=30
        )
        stats = {}
        for line in stdout.split("\n"):
            if "|" not in line:
                continue
            parts = line.split("|")
            if len(parts) >= 7:
                name = parts[0].strip()
                stats[name] = {
                    "cpu_pct": ContainerMetrics._parse_pct(parts[1]),
                    "mem_usage": parts[2].strip(),
                    "mem_pct": ContainerMetrics._parse_pct(parts[3]),
                    "net_io": parts[4].strip(),
                    "block_io": parts[5].strip(),
                    "pids": parts[6].strip(),
                }
        return stats

    @staticmethod
    def get_container_logs(service, tail=20):
        """Get recent logs from a container."""
        stdout, _ = run_cmd(f"docker logs {service} --tail {tail} 2>&1", timeout=10)
        return stdout

    @staticmethod
    def get_docker_compose_resources():
        """Get resource limits from docker-compose."""
        stdout, _ = run_cmd(
            "docker inspect -f '{{.Name}} | {{.HostConfig.CpuQuota}} | {{.HostConfig.Memory}} | {{.HostConfig.NanoCpus}}' "
            + " ".join(MONITORED_SERVICES), timeout=15
        )
        return stdout

    @staticmethod
    def find_slow_logs(service, pattern="slow|SLOW|timeout|TIMEOUT|error|ERROR", tail=50):
        """Find slow/error log entries."""
        stdout, _ = run_cmd(
            f"docker logs {service} --tail {tail} 2>&1 | grep -iE '{pattern}' | tail -10",
            timeout=10
        )
        return stdout

    @staticmethod
    def _parse_pct(s):
        try:
            return float(s.strip().replace("%", ""))
        except (ValueError, AttributeError):
            return 0.0


# ═══════════════════════════════════════════════════════════════
# Load Test Runner
# ═══════════════════════════════════════════════════════════════

class LoadTestRunner:
    def __init__(self):
        self.results = {}
        self.metrics_snapshots = {}
        self.metrics = ContainerMetrics()

    def run_level(self, tps_target):
        """Run a single load test level."""
        config = TPS_LEVELS[tps_target]
        ts_tag = datetime.now().strftime("%Y%m%d_%H%M%S")
        csv_prefix = str(RESULTS_DIR / f"tps{tps_target}_{ts_tag}")
        stats_tag = f"tps{tps_target}"

        log(f"{'═'*60}", "RES")
        log(f"LOAD TEST: {tps_target} TPS target", "RES")
        log(f"  Users: {config['users']} | Spawn: {config['spawn_rate']} | Duration: {config['duration']}", "INFO")
        log(f"  User class: {config['classes']}", "INFO")
        log(f"{'═'*60}", "RES")

        # Pre-test metrics
        log("Collecting pre-test baseline...", "INFO")
        self.metrics_snapshots[f"{stats_tag}_before"] = self.metrics.get_all_stats()
        time.sleep(2)

        # Run locust
        cmd = (
            f"cd {CHAOS_DIR} && locust -f locustfile.py "
            f"--headless "
            f"--users {config['users']} "
            f"--spawn-rate {config['spawn_rate']} "
            f"--run-time {config['duration']} "
            f"--csv {csv_prefix} "
            f"--html {csv_prefix}.html "
            f"--only-summary "
            f"--class-picker "
            f"--tags {config['classes'].replace('Tiki', '')} "
            f"2>&1"
        )

        # Simpler: just specify the classes directly
        user_classes = config['classes']
        cmd = (
            f"cd {CHAOS_DIR} && locust -f locustfile.py "
            f"--headless "
            f"--users {config['users']} "
            f"--spawn-rate {config['spawn_rate']} "
            f"--run-time {config['duration']} "
            f"--csv {csv_prefix} "
            f"--html {csv_prefix}.html "
            f"--only-summary "
            f"2>&1"
        )

        log(f"Running: locust -u {config['users']} -r {config['spawn_rate']} --run-time {config['duration']}", "RES")
        start = time.time()

        process = subprocess.Popen(
            cmd, shell=True,
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            text=True, bufsize=1,
        )

        # Collect metrics during test
        mid_test_stats = None
        output_lines = []

        while process.poll() is None:
            line = process.stdout.readline() if process.stdout else None
            if line:
                output_lines.append(line.strip())
                print(f"  {line.strip()}")

                # Capture mid-test stats at halfway
                elapsed = time.time() - start
                if mid_test_stats is None and elapsed > self._duration_seconds(config['duration']) / 2:
                    log("Capturing mid-test container metrics...", "INFO")
                    mid_test_stats = self.metrics.get_all_stats()

        # Collect post-test metrics
        time.sleep(3)  # Let things settle slightly
        self.metrics_snapshots[f"{stats_tag}_after"] = self.metrics.get_all_stats()
        if mid_test_stats:
            self.metrics_snapshots[f"{stats_tag}_mid"] = mid_test_stats

        elapsed = time.time() - start
        log(f"Test completed in {elapsed:.1f}s", "OK")

        # Parse locust CSV results
        result = self._parse_results(csv_prefix, tps_target)
        self.results[tps_target] = result

        # Also capture error logs from containers
        self._capture_error_logs(stats_tag)

        return result

    def _duration_seconds(self, dur_str):
        """Parse duration string like '30s', '1m'."""
        match = re.match(r'(\d+)([smh])', dur_str)
        if match:
            val, unit = int(match.group(1)), match.group(2)
            return val * {'s': 1, 'm': 60, 'h': 3600}[unit]
        return 30

    def _parse_results(self, csv_prefix, tps_target):
        """Parse locust CSV output."""
        stats_file = Path(f"{csv_prefix}_stats.csv")
        if not stats_file.exists():
            log(f"CSV not found: {stats_file}", "WARN")
            return {"tps_target": tps_target, "error": "no results"}

        lines = stats_file.read_text().strip().split("\n")
        if len(lines) < 2:
            return {"tps_target": tps_target, "error": "empty results"}

        # Parse header and find Aggregate row
        header = lines[0].split(",")
        aggregate = None
        endpoints = []
        for line in lines[1:]:
            parts = line.split(",")
            if len(parts) >= 10:
                row = dict(zip(header, parts))
                name = row.get("Name", "")
                if name == "Aggregated" or name == "Total":
                    aggregate = row
                elif name and name != "Aggregated":
                    endpoints.append(row)

        result = {
            "tps_target": tps_target,
            "aggregate": aggregate,
            "endpoints": endpoints,
            "stats_file": str(stats_file),
        }

        # Print summary
        if aggregate:
            log(f"Results for {tps_target} TPS:", "RES")
            log(f"  Total requests: {aggregate.get('Request Count', 'N/A')}", "INFO")
            log(f"  Failures: {aggregate.get('Failure Count', 'N/A')}", "INFO")

            # Parse avg response time
            try:
                p50 = float(aggregate.get("50%", 0))
                p90 = float(aggregate.get("90%", 0))
                p95 = float(aggregate.get("95%", 0))
                p99 = float(aggregate.get("99%", 0))
                avg = float(aggregate.get("Average Response Time", 0))
                rps = float(aggregate.get("Requests/s", 0))

                log(f"  Actual RPS: {rps:.1f}", "RES")
                log(f"  Avg RT: {avg:.0f}ms | p50: {p50:.0f}ms | p90: {p90:.0f}ms | p95: {p95:.0f}ms | p99: {p99:.0f}ms", "RES")

                result.update({
                    "rps": rps,
                    "avg_ms": avg,
                    "p50_ms": p50,
                    "p90_ms": p90,
                    "p95_ms": p95,
                    "p99_ms": p99,
                    "failures": aggregate.get("Failure Count", "0"),
                })
            except (ValueError, TypeError) as e:
                log(f"Parse error: {e}", "WARN")

        # Print per-endpoint breakdown for slow endpoints
        if endpoints:
            log(f"  Per-endpoint breakdown:", "RES")
            for ep in sorted(endpoints, key=lambda x: float(x.get("Average Response Time", 0) or 0), reverse=True)[:10]:
                name = ep.get("Name", "?")[:40]
                try:
                    ep_avg = float(ep.get("Average Response Time", 0) or 0)
                    ep_p99 = float(ep.get("99%", 0) or 0)
                    ep_count = ep.get("Request Count", "0")
                    ep_fail = ep.get("Failure Count", "0")
                    color = "OK" if ep_avg < 500 else ("WARN" if ep_avg < 2000 else "FAIL")
                    log(f"    {name:40s} avg={ep_avg:6.0f}ms p99={ep_p99:6.0f}ms n={ep_count} f={ep_fail}", color)
                except (ValueError, TypeError):
                    pass

        return result

    def _capture_error_logs(self, stats_tag):
        """Capture error logs from all monitored services."""
        error_data = {}
        for service in MONITORED_SERVICES:
            errors = self.metrics.find_slow_logs(service, tail=100)
            if errors:
                error_data[service] = errors
        if error_data:
            self.metrics_snapshots[f"{stats_tag}_errors"] = error_data

    def generate_bottleneck_report(self):
        """Generate comprehensive bottleneck analysis."""
        log(f"\n{'═'*60}", "RES")
        log("BOTTLENECK ANALYSIS REPORT", "RES")
        log(f"{'═'*60}", "RES")

        report_lines = [
            f"{'═'*70}",
            "TIKI-CLONE LOAD TEST / BOTTLENECK ANALYSIS REPORT",
            f"Generated: {datetime.now().isoformat()}",
            f"{'═'*70}",
            "",
        ]

        # 1. RPS scaling table
        report_lines.append("── RPS SCALING SUMMARY " + "─"*45)
        report_lines.append(f"{'Target TPS':>12s} {'Actual RPS':>12s} {'Avg RT':>10s} {'p95 RT':>10s} {'p99 RT':>10s} {'Failures':>10s} {'Status':>10s}")
        baseline_avg = None
        bottleneck_found = None

        for tps in sorted(self.results.keys()):
            r = self.results[tps]
            if "error" in r:
                report_lines.append(f"{tps:>12d} {'N/A':>12s}")
                continue

            avg = r.get("avg_ms", 0)
            p95 = r.get("p95_ms", 0)
            p99 = r.get("p99_ms", 0)
            rps = r.get("rps", 0)
            fails = r.get("failures", "0")

            if baseline_avg is None:
                baseline_avg = avg
            status = "✓ OK" if avg < 500 else ("△ WARN" if avg < 2000 else "✗ SLOW")
            if avg >= 2000 and bottleneck_found is None:
                bottleneck_found = tps
                status = "✗ BOTTLENECK"

            report_lines.append(
                f"{tps:>12d} {rps:>12.1f} {avg:>9.0f}ms {p95:>9.0f}ms {p99:>9.0f}ms {str(fails):>10s} {status:>10s}"
            )

        report_lines.append("")

        # 2. Container metrics comparison
        report_lines.append("── CONTAINER RESOURCE USAGE " + "─"*42)

        for tps in sorted(self.results.keys()):
            tag = f"tps{tps}"
            before = self.metrics_snapshots.get(f"{tag}_before", {})
            after = self.metrics_snapshots.get(f"{tag}_after", {})

            report_lines.append(f"\n  At {tps} TPS:")
            report_lines.append(f"  {'Service':30s} {'CPU before':>10s} {'CPU after':>10s} {'CPU Δ':>8s} {'Mem after':>12s}")
            for svc in MONITORED_SERVICES:
                b = before.get(svc, {})
                a = after.get(svc, {})
                cpu_before = b.get("cpu_pct", 0)
                cpu_after = a.get("cpu_pct", 0)
                cpu_delta = cpu_after - cpu_before
                mem_pct = a.get("mem_pct", 0)
                cpu_usage = a.get("cpu_pct", 0)

                # Flag hot services
                hot = " ← HOT" if cpu_usage > 80 or cpu_delta > 30 else ""
                report_lines.append(
                    f"  {svc:30s} {cpu_before:>9.1f}% {cpu_after:>9.1f}% {cpu_delta:>+7.1f}% {mem_pct:>10.1f}%{hot}"
                )

        report_lines.append("")

        # 3. Error log summary
        report_lines.append("── ERROR LOG SUMMARY " + "─"*48)
        has_errors = False
        for tps in sorted(self.results.keys()):
            tag = f"tps{tps}_errors"
            errors = self.metrics_snapshots.get(tag, {})
            if errors:
                has_errors = True
                report_lines.append(f"\n  At {tps} TPS:")
                for svc, errs in errors.items():
                    report_lines.append(f"    {svc}:")
                    for line in errs.strip().split("\n")[:5]:
                        report_lines.append(f"      {line[:120]}")

        if not has_errors:
            report_lines.append("  No errors detected in container logs.")

        report_lines.append("")

        # 4. Bottleneck diagnosis
        report_lines.append("── BOTTLENECK DIAGNOSIS " + "─"*44)
        if bottleneck_found:
            report_lines.append(f"  ✗ System becomes bottlenecked at ~{bottleneck_found} TPS")

            # Analyze which service is the bottleneck
            tag = f"tps{bottleneck_found}_after"
            stats = self.metrics_snapshots.get(tag, {})
            hot_services = []
            for svc, data in stats.items():
                cpu = data.get("cpu_pct", 0)
                mem = data.get("mem_pct", 0)
                if cpu > 50 or mem > 80:
                    hot_services.append((svc, cpu, mem))

            if hot_services:
                report_lines.append(f"\n  Hottest services at bottleneck ({bottleneck_found} TPS):")
                for svc, cpu, mem in sorted(hot_services, key=lambda x: x[1], reverse=True):
                    report_lines.append(f"    {svc:35s} CPU: {cpu:.1f}%  Mem: {mem:.1f}%")

                # Identify the primary bottleneck
                primary = hot_services[0]
                report_lines.append(f"\n  ► PRIMARY BOTTLENECK: {primary[0]}")
                if primary[1] > 80:
                    report_lines.append(f"    CPU saturated at {primary[1]:.1f}%")
                if primary[2] > 80:
                    report_lines.append(f"    Memory saturated at {primary[2]:.1f}%")

                # Check for specific patterns
                web_tls_cpu = stats.get("tikiclone-web-tls-1", {}).get("cpu_pct", 0)
                gateway_cpu = stats.get("tikiclone-gateway-1", {}).get("cpu_pct", 0)
                web_cpu = stats.get("tikiclone-web-1", {}).get("cpu_pct", 0)
                mysql_cpu = stats.get("tikiclone-mysql-primary-1", {}).get("cpu_pct", 0)
                redis_cpu = stats.get("tikiclone-redis-master-1", {}).get("cpu_pct", 0)

                if mysql_cpu > gateway_cpu and mysql_cpu > web_cpu:
                    report_lines.append(f"\n    → Database (MySQL) is the limiting factor ({mysql_cpu:.1f}% CPU)")
                    report_lines.append(f"      Recommendation: Add DB read replicas, optimize queries, add connection pooling")
                elif gateway_cpu > 70:
                    report_lines.append(f"\n    → API Gateway is the limiting factor ({gateway_cpu:.1f}% CPU)")
                    report_lines.append(f"      Recommendation: Scale gateway horizontally, add rate limiting")
                elif web_tls_cpu > 70:
                    report_lines.append(f"\n    → Nginx/TLS termination is the limiting factor ({web_tls_cpu:.1f}% CPU)")
                    report_lines.append(f"      Recommendation: Add nginx replicas, use HTTP/2, offload TLS")
                elif web_cpu > 70:
                    report_lines.append(f"\n    → Next.js frontend is the limiting factor ({web_cpu:.1f}% CPU)")
                    report_lines.append(f"      Recommendation: Enable SSR caching, add CDN, scale web tier")
                elif redis_cpu > 50:
                    report_lines.append(f"\n    → Redis cache is under pressure ({redis_cpu:.1f}% CPU)")
                    report_lines.append(f"      Recommendation: Add Redis cluster, optimize cache TTLs")
        else:
            report_lines.append("  ✓ No bottleneck detected within tested range")

        report_lines.append("")

        # 5. Recommendations
        report_lines.append("── RECOMMENDATIONS " + "─"*49)
        report_lines.append("  1. Scale the primary bottleneck service horizontally")
        report_lines.append("  2. Add connection pooling to database-facing services")
        report_lines.append("  3. Enable response caching at gateway level")
        report_lines.append("  4. Add Redis caching for frequently accessed data")
        report_lines.append("  5. Consider CDN for static assets (images, JS, CSS)")
        report_lines.append("  6. Profile slow endpoints with pprof/trace")
        report_lines.append(f"\n{'═'*70}")

        report = "\n".join(report_lines)

        # Save report
        report_path = RESULTS_DIR / f"bottleneck_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"
        report_path.write_text(report)
        log(f"Report saved to {report_path}", "OK")

        # Print report
        print(report)

        return report


# ═══════════════════════════════════════════════════════════════
# Main
# ═══════════════════════════════════════════════════════════════

def main():
    import argparse
    parser = argparse.ArgumentParser(description="TikiClone Load Test Orchestrator")
    parser.add_argument("--level", default="all",
                        help="TPS level to run (10, 100, 1000, 5000, 10000, or all)")
    parser.add_argument("--skip-cooldown", action="store_true",
                        help="Skip cooldown between tests")
    args = parser.parse_args()

    log(f"{C.BD}{'═'*60}")
    log(f"  TikiClone Load Test Orchestrator")
    log(f"{'═'*60}{C.RS}")
    log(f"Results directory: {RESULTS_DIR}")
    log(f"Entry point: HTTPS via Nginx (port 8443)")
    log(f"Levels: {list(TPS_LEVELS.keys())} TPS")

    runner = LoadTestRunner()

    # Determine which levels to run
    if args.level == "all":
        levels = list(TPS_LEVELS.keys())
    else:
        try:
            levels = [int(args.level)]
        except ValueError:
            log(f"Invalid level: {args.level}", "FAIL")
            sys.exit(1)

    # Run each level sequentially
    for i, tps in enumerate(levels):
        try:
            runner.run_level(tps)

            # Cooldown between tests (unless last)
            if i < len(levels) - 1 and not args.skip_cooldown:
                cooldown = 15
                log(f"Cooling down {cooldown}s before next level...", "INFO")
                time.sleep(cooldown)
        except KeyboardInterrupt:
            log("Load test interrupted!", "WARN")
            break
        except Exception as e:
            log(f"Error at {tps} TPS: {e}", "FAIL")
            import traceback
            traceback.print_exc()

    # Generate report
    if runner.results:
        runner.generate_bottleneck_report()
    else:
        log("No results to report", "FAIL")

if __name__ == "__main__":
    main()
