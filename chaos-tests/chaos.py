#!/usr/bin/env python3
"""
TikiClone Chaos Engineering Test Suite
=======================================
Tests resilience of the microservice platform by injecting failures.

Usage:
    python3 chaos.py --scenario all          # Run all scenarios
    python3 chaos.py --scenario kill         # Container kill/restart
    python3 chaos.py --scenario latency      # Network latency injection
    python3 chaos.py --scenario partition    # Network partition
    python3 chaos.py --scenario loss         # Packet loss
    python3 chaos.py --scenario cascade      # Dependency cascade failure
    python3 chaos.py --scenario stress       # CPU stress on containers
    python3 chaos.py --monitor               # Continuous health monitor only

Requirements: docker CLI, tc (traffic control), Python 3.7+
"""

import subprocess
import json
import time
import random
import sys
import signal
import threading
import argparse
from datetime import datetime
from pathlib import Path
from typing import Optional

# ═══════════════════════════════════════════════════════════════
# Configuration
# ═══════════════════════════════════════════════════════════════

COMPOSE_PROJECT = "tikiclone"
CHAOS_LOG_DIR = Path(__file__).parent / "results"
CHAOS_LOG_DIR.mkdir(exist_ok=True)

# Service categories
INFRASTRUCTURE = ["mysql-primary", "redis-master", "mongodb", "kafka", "zookeeper"]
GATEWAY = ["gateway"]
CORE_SERVICES = [
    "identity-auth", "auth",
    "cart", "order", "payment", "checkout",
    "catalog-product", "product", "product-catalog",
    "inventory", "shipment", "promotion",
]
FRONTEND = ["web", "web-tls"]
OBSERVABILITY = ["prometheus", "grafana", "jaeger", "otel-collector"]
AI = ["ollama"]

# All service names as they appear in docker compose
ALL_SERVICES = INFRASTRUCTURE + GATEWAY + CORE_SERVICES + FRONTEND + OBSERVABILITY + AI

# Health check endpoints
HEALTH_ENDPOINTS = {
    "gateway": "http://localhost:8080/health",
    "identity-auth": "http://localhost:8081/api/auth/health",
    "cart": "http://localhost:8082/health",
    "payment": "http://localhost:8083/health",
    "order": "http://localhost:8084/health",
    "checkout": "http://localhost:8085/health",
    "inventory": "http://localhost:8086/health",
    "auth": "http://localhost:8087/health",
    "catalog-product": "http://localhost:8088/health",
    "product": "http://localhost:8089/health",
    "product-catalog": "http://localhost:8090/health",
    "promotion": "http://localhost:8091/health",
    "shipment": "http://localhost:8092/health",
    "web": "http://localhost:3000",
    "ollama": "http://localhost:11434/api/tags",
}

# Colors for terminal output
class Colors:
    RED = "\033[91m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    BLUE = "\033[94m"
    MAGENTA = "\033[95m"
    CYAN = "\033[96m"
    BOLD = "\033[1m"
    RESET = "\033[0m"

def log(msg: str, level: str = "INFO"):
    ts = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    color = {
        "INFO": Colors.BLUE,
        "OK": Colors.GREEN,
        "WARN": Colors.YELLOW,
        "FAIL": Colors.RED,
        "CHAOS": Colors.MAGENTA,
    }.get(level, Colors.RESET)
    print(f"{color}[{ts}] [{level:5s}] {msg}{Colors.RESET}")

def run(cmd: str, timeout: int = 30, check: bool = False) -> subprocess.CompletedProcess:
    """Run a shell command."""
    return subprocess.run(
        cmd, shell=True, capture_output=True, text=True,
        timeout=timeout, check=check,
    )

def docker_compose(cmd: str, timeout: int = 60) -> subprocess.CompletedProcess:
    """Run a docker compose command."""
    return run(f"cd /home/datdt/tikiclone && docker compose -p {COMPOSE_PROJECT} {cmd}", timeout=timeout)

def get_container_name(service: str) -> Optional[str]:
    """Get the actual container name for a service."""
    result = docker_compose(f"ps --format json {service}")
    if result.returncode == 0 and result.stdout.strip():
        try:
            data = json.loads(result.stdout.strip().split("\n")[0])
            return data.get("Name", f"tikiclone-{service}-1")
        except (json.JSONDecodeError, IndexError):
            pass
    return f"tikiclone-{service}-1"

def is_container_running(service: str) -> bool:
    name = get_container_name(service)
    result = run(f"docker inspect {name} --format '{{{{.State.Status}}}}'", timeout=10)
    return result.stdout.strip() == "run"

# ═══════════════════════════════════════════════════════════════
# Health Monitor
# ═══════════════════════════════════════════════════════════════

class HealthMonitor:
    """Continuously monitors service health."""

    def __init__(self):
        self.results: list[dict] = []
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None

    def start(self):
        self._stop.clear()
        self._thread = threading.Thread(target=self._monitor_loop, daemon=True)
        self._thread.start()
        log("Health monitor started", "OK")

    def stop(self):
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=5)
        log("Health monitor stopped", "OK")

    def _monitor_loop(self):
        while not self._stop.is_set():
            snapshot = {}
            for service, url in HEALTH_ENDPOINTS.items():
                try:
                    r = run(f"curl -sf -o /dev/null -w '%{{http_code}}' --max-time 3 {url}", timeout=5)
                    snapshot[service] = {
                        "status": "up" if r.stdout.strip() in ("200", "201", "204") else "degraded",
                        "http_code": r.stdout.strip(),
                    }
                except Exception:
                    snapshot[service] = {"status": "down", "http_code": "000"}

                # Also check container state
                name = get_container_name(service)
                cr = run(f"docker inspect {name} --format '{{{{.State.Status}}}}'", timeout=5)
                snapshot[service]["container"] = cr.stdout.strip()

            self.results.append({"time": time.time(), "services": snapshot})
            self._stop.wait(timeout=2)

    def get_summary(self) -> dict:
        if not self.results:
            return {}
        latest = self.results[-1]
        summary = {}
        for svc, data in latest["services"].items():
            summary[svc] = data
        return summary

    def get_downtime_report(self) -> dict:
        """Calculate total downtime per service during monitoring."""
        if not self.results:
            return {}
        downtime = {}
        for service in HEALTH_ENDPOINTS:
            down_count = sum(
                1 for r in self.results
                if r["services"].get(service, {}).get("status") == "down"
            )
            total = len(self.results)
            downtime[service] = {
                "down_checks": down_count,
                "total_checks": total,
                "availability_pct": round((1 - down_count / total) * 100, 2) if total else 100,
            }
        return downtime

# ═══════════════════════════════════════════════════════════════
# Chaos Scenarios
# ═══════════════════════════════════════════════════════════════

class ChaosEngine:
    def __init__(self):
        self.monitor = HealthMonitor()
        self.scenario_results: list[dict] = []

    def _report(self, scenario: str, action: str, detail: str, passed: bool):
        status = "PASS" if passed else "FAIL"
        self.scenario_results.append({
            "scenario": scenario, "action": action,
            "detail": detail, "passed": passed, "time": time.time(),
        })
        log(f"[{scenario}] {action}: {detail}", status)

    # ── Scenario 1: Container Kill/Restart ────────────────────

    def scenario_kill(self, count: int = 3, grace_period: int = 15):
        """Kill random non-infrastructure containers and verify recovery."""
        log(f"{'═'*60}", "CHAOS")
        log("SCENARIO: Container Kill/Restart", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        targets = [s for s in CORE_SERVICES + GATEWAY + FRONTEND if is_container_running(s)]
        if len(targets) < count:
            count = len(targets)
        victims = random.sample(targets, count)

        self.monitor.start()
        time.sleep(3)  # baseline

        for victim in victims:
            container = get_container_name(victim)
            log(f"Killing container: {container} ({victim})", "CHAOS")
            run(f"docker kill {container}", timeout=10)
            self._report("kill", "container_kill", f"Killed {container}", True)

            # Wait for recovery
            recovered = False
            for i in range(grace_period):
                time.sleep(1)
                if is_container_running(victim):
                    log(f"{victim} recovered after {i+1}s", "OK")
                    recovered = True
                    break
                if i % 5 == 0:
                    log(f"Waiting for {victim}... ({i}s)", "WARN")

            recovery_time = i + 1
            self._report("kill", "container_recovery",
                         f"{victim} recovery in {recovery_time}s" if recovered else f"{victim} NOT recovered",
                         recovered)

        time.sleep(5)
        self.monitor.stop()

        # Verify all services healthy
        all_ok = all(is_container_running(s) for s in targets)
        self._report("kill", "final_check", "All targeted services running", all_ok)
        return all_ok

    # ── Scenario 2: Network Latency Injection ─────────────────

    def scenario_latency(self, target_service: str = "mysql-primary",
                         latency_ms: int = 200, duration: int = 30):
        """Inject network latency to a dependency and measure impact."""
        log(f"{'═'*60}", "CHAOS")
        log(f"SCENARIO: Network Latency ({latency_ms}ms → {target_service})", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        container = get_container_name(target_service)
        if not container:
            log(f"Container not found for {target_service}", "FAIL")
            return False

        # Pick a few gateway-side containers to inject latency into
        test_sources = ["tikiclone-gateway-1", "tikiclone-auth-1", "tikiclone-order-1"]
        test_sources = [c for c in test_sources if is_container_running(c.replace("tikiclone-", "").replace("-1", ""))]

        if not test_sources:
            test_sources = [get_container_name("gateway")]

        self.monitor.start()
        time.sleep(3)

        # Inject latency using tc
        for src in test_sources:
            log(f"Injecting {latency_ms}ms latency on {src} → {target_service}", "CHAOS")
            # Add network delay using tc
            inject_cmd = (
                f"docker exec {src} sh -c '"
                f"tc qdisc add dev eth0 root netem delay {latency_ms}ms 2>/dev/null || "
                f"tc qdisc replace dev eth0 root netem delay {latency_ms}ms'"
            )
            r = run(inject_cmd, timeout=10)
            self._report("latency", "inject",
                         f"tc delay {latency_ms}ms on {src}: {'OK' if r.returncode == 0 else r.stderr[:100]}",
                         r.returncode == 0)

            # Measure impact with a curl through gateway
            t0 = time.time()
            run("curl -sf --max-time 10 http://localhost:8080/health", timeout=15)
            elapsed = time.time() - t0
            self._report("latency", "measure",
                         f"Gateway health check: {elapsed:.2f}s (baseline ~0.1s)",
                         elapsed < 5.0)

        log(f"Waiting {duration}s with active latency...", "CHAOS")
        time.sleep(duration)

        # Remove latency
        for src in test_sources:
            log(f"Removing latency from {src}", "OK")
            run(f"docker exec {src} sh -c 'tc qdisc del dev eth0 root 2>/dev/null; tc qdisc del dev eth0 root 2>/dev/null'", timeout=10)
            self._report("latency", "remove", f"tc cleared on {src}", True)

        time.sleep(5)
        self.monitor.stop()
        return True

    # ── Scenario 3: Network Partition ──────────────────────────

    def scenario_partition(self, partition_service: str = "redis-master",
                           duration: int = 20):
        """Isolate a service from the network entirely."""
        log(f"{'═'*60}", "CHAOS")
        log(f"SCENARIO: Network Partition ({partition_service} isolated)", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        container = get_container_name(partition_service)
        if not container:
            log(f"Container not found for {partition_service}", "FAIL")
            return False

        # Get current gateway response
        r = run("curl -sf --max-time 5 http://localhost:8080/health", timeout=10)
        baseline_ok = r.returncode == 0
        log(f"Baseline gateway health: {'OK' if baseline_ok else 'FAIL'}", "INFO")

        self.monitor.start()
        time.sleep(3)

        # Disconnect the service from the network
        network_name = "tikiclone_default"
        log(f"Disconnecting {container} from {network_name}", "CHAOS")
        r = run(f"docker network disconnect {network_name} {container}", timeout=10)
        self._report("partition", "disconnect",
                     f"{container} from {network_name}", r.returncode == 0)

        # Observe impact
        time.sleep(3)
        r = run("curl -sf --max-time 5 http://localhost:8080/health", timeout=10)
        partition_impact = r.returncode != 0
        log(f"Gateway reachable during partition: {not partition_impact}",
            "WARN" if partition_impact else "OK")
        self._report("partition", "impact_check",
                     f"Gateway hit during partition: {partition_impact}", True)

        log(f"Waiting {duration}s with partition active...", "CHAOS")
        time.sleep(duration)

        # Reconnect
        log(f"Reconnecting {container} to {network_name}", "OK")
        r = run(f"docker network connect {network_name} {container}", timeout=10)
        self._report("partition", "reconnect",
                     f"{container} to {network_name}", r.returncode == 0)

        # Wait for recovery
        time.sleep(5)
        r = run("curl -sf --max-time 5 http://localhost:8080/health", timeout=10)
        recovered = r.returncode == 0
        self._report("partition", "recovery",
                     f"Gateway recovered: {recovered}", recovered)

        self.monitor.stop()
        return recovered

    # ── Scenario 4: Packet Loss ───────────────────────────────

    def scenario_loss(self, target_service: str = "gateway",
                      loss_pct: int = 30, duration: int = 20):
        """Inject packet loss to simulate unreliable network."""
        log(f"{'═'*60}", "CHAOS")
        log(f"SCENARIO: Packet Loss ({loss_pct}% on {target_service})", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        container = get_container_name(target_service)
        if not container:
            log("Container not found", "FAIL")
            return False

        self.monitor.start()
        time.sleep(3)

        # Inject packet loss
        log(f"Injecting {loss_pct}% packet loss on {container}", "CHAOS")
        inject_cmd = (
            f"docker exec {container} sh -c '"
            f"tc qdisc replace dev eth0 root netem loss {loss_pct}%'"
        )
        r = run(inject_cmd, timeout=10)
        self._report("loss", "inject",
                     f"{loss_pct}% loss on {container}", r.returncode == 0)

        # Send test requests
        successes = 0
        total = 20
        for i in range(total):
            r = run("curl -sf --max-time 3 http://localhost:8080/health", timeout=5)
            if r.returncode == 0:
                successes += 1
            time.sleep(0.5)

        success_rate = successes / total * 100
        log(f"Requests: {successes}/{total} succeeded ({success_rate:.0f}%)", "WARN")
        self._report("loss", "during_loss",
                     f"Success rate: {success_rate:.0f}% with {loss_pct}% packet loss", True)

        time.sleep(duration - 10)

        # Remove packet loss
        run(f"docker exec {container} sh -c 'tc qdisc del dev eth0 root 2>/dev/null'", timeout=10)
        self._report("loss", "remove", "tc cleared", True)

        time.sleep(3)
        self.monitor.stop()
        return True

    # ── Scenario 5: Dependency Cascade Failure ────────────────

    def scenario_cascade(self):
        """Kill infrastructure dependency and observe cascade."""
        log(f"{'═'*60}", "CHAOS")
        log("SCENARIO: Dependency Cascade Failure", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        # We'll stop redis (used by many services) briefly
        cascade_target = "redis-master"
        container = get_container_name(cascade_target)

        self.monitor.start()
        time.sleep(3)

        # Check which services use redis by looking at their env
        affected_services = []
        for svc in CORE_SERVICES:
            c = get_container_name(svc)
            if c:
                r = run(f"docker exec {c} env 2>/dev/null | grep -i redis", timeout=5)
                if r.returncode == 0 and r.stdout.strip():
                    affected_services.append(svc)

        log(f"Services using Redis: {', '.join(affected_services)}", "INFO")

        # Stop redis
        log(f"Stopping {cascade_target}...", "CHAOS")
        docker_compose(f"stop {cascade_target}", timeout=30)
        self._report("cascade", "stop", f"Stopped {cascade_target}", True)

        time.sleep(5)

        # Check impact on affected services
        for svc in affected_services[:5]:  # check first 5
            name = get_container_name(svc)
            status = run(f"docker inspect {name} --format '{{{{.State.Status}}}}'", timeout=5).stdout.strip()
            r = run(f"curl -sf --max-time 3 {HEALTH_ENDPOINTS.get(svc, '')} 2>/dev/null; echo $?", timeout=5)
            healthy = "0" in r.stdout
            self._report("cascade", f"impact_{svc}",
                         f"{svc} container={status} http_ok={healthy}", True)

        # Restart redis
        log(f"Restarting {cascade_target}...", "CHAOS")
        docker_compose(f"start {cascade_target}", timeout=30)
        self._report("cascade", "restart", f"Started {cascade_target}",
                     is_container_running(cascade_target))

        # Wait for redis to be healthy
        time.sleep(10)
        redis_healthy = run(
            "docker exec tikiclone-redis-master-1 redis-cli ping 2>/dev/null", timeout=5
        ).stdout.strip() == "PONG"
        self._report("cascade", "redis_health", f"Redis PING: {redis_healthy}", redis_healthy)

        # Verify services recovered
        time.sleep(5)
        for svc in affected_services[:5]:
            healthy = run(
                f"curl -sf --max-time 3 {HEALTH_ENDPOINTS.get(svc, '')} 2>/dev/null; echo $?",
                timeout=5
            )
            recovered = "0" in healthy.stdout
            self._report("cascade", f"recovery_{svc}", f"{svc} recovered: {recovered}", recovered)

        self.monitor.stop()
        return redis_healthy

    # ── Scenario 6: CPU Stress ────────────────────────────────

    def scenario_stress(self, target_service: str = "gateway",
                        cpu_load: int = 80, duration: int = 30):
        """Apply CPU stress to a container."""
        log(f"{'═'*60}", "CHAOS")
        log(f"SCENARIO: CPU Stress ({cpu_load}% on {target_service})", "CHAOS")
        log(f"{'═'*60}", "CHAOS")

        container = get_container_name(target_service)
        if not container:
            log("Container not found", "FAIL")
            return False

        self.monitor.start()
        time.sleep(3)

        # Use docker stats to get CPU count, then spawn stress processes
        log(f"Starting CPU stress on {container}", "CHAOS")
        # Spawn a background stress process inside the container using /dev/urandom
        stress_cmd = (
            f"docker exec -d {container} sh -c '"
            f"for i in $(seq 1 $(nproc)); do "
            f"  md5sum /dev/urandom & "
            f"done; wait'"
        )
        run(stress_cmd, timeout=10)
        self._report("stress", "start", f"CPU stress on {container}", True)

        time.sleep(duration)

        # Kill stress processes
        log(f"Killing stress processes in {container}", "OK")
        run(f"docker exec {container} sh -c 'pkill -f md5sum 2>/dev/null; killall md5sum 2>/dev/null'", timeout=10)
        self._report("stress", "stop", "Stress processes killed", True)

        time.sleep(3)

        # Verify service still responsive
        r = run("curl -sf --max-time 5 http://localhost:8080/health", timeout=10)
        recovered = r.returncode == 0
        self._report("stress", "recovery",
                     f"Gateway responsive after stress: {recovered}", recovered)

        self.monitor.stop()
        return recovered

    # ── Report Generation ─────────────────────────────────────

    def generate_report(self) -> str:
        """Generate a test report."""
        report_lines = [
            "",
            f"{'═'*60}",
            "CHAOS TEST REPORT",
            f"{'═'*60}",
            f"Time: {datetime.now().isoformat()}",
            f"Project: {COMPOSE_PROJECT}",
            "",
        ]

        passed = sum(1 for r in self.scenario_results if r["passed"])
        total = len(self.scenario_results)
        report_lines.append(f"Results: {passed}/{total} checks passed ({passed/total*100:.0f}%)" if total else "No results")

        # Group by scenario
        scenarios = {}
        for r in self.scenario_results:
            scenarios.setdefault(r["scenario"], []).append(r)

        for scenario, checks in scenarios.items():
            report_lines.append(f"\n── {scenario.upper()} ──")
            spaased = sum(1 for c in checks if c["passed"])
            report_lines.append(f"  {spaased}/{len(checks)} passed")
            for c in checks:
                icon = "✓" if c["passed"] else "✗"
                report_lines.append(f"  {icon} [{c['action']}] {c['detail']}")

        # Downtime report
        downtime = self.monitor.get_downtime_report()
        if downtime:
            report_lines.append(f"\n── AVAILABILITY DURING CHAOS ──")
            for svc, data in sorted(downtime.items(), key=lambda x: x[1]["availability_pct"]):
                pct = data["availability_pct"]
                pct_color = ""  # Can't use colors in file
                report_lines.append(
                    f"  {svc:25s} {pct:6.2f}%  "
                    f"(down {data['down_checks']}/{data['total_checks']} checks)"
                )

        report_lines.append(f"\n{'═'*60}")
        report = "\n".join(report_lines)

        # Save to file
        report_path = CHAOS_LOG_DIR / f"chaos_report_{int(time.time())}.txt"
        report_path.write_text(report)
        log(f"Report saved to {report_path}", "OK")

        return report


# ═══════════════════════════════════════════════════════════════
# Main
# ═══════════════════════════════════════════════════════════════

def main():
    parser = argparse.ArgumentParser(description="TikiClone Chaos Testing")
    parser.add_argument("--scenario", default="monitor",
                        choices=["all", "kill", "latency", "partition", "loss", "cascade", "stress", "monitor"],
                        help="Chaos scenario to run")
    parser.add_argument("--target", default=None,
                        help="Target service name (for applicable scenarios)")
    parser.add_argument("--duration", type=int, default=20,
                        help="Duration in seconds for network chaos")
    parser.add_argument("--intensity", type=int, default=None,
                        help="Intensity (latency ms, loss %, cpu %)")
    parser.add_argument("--count", type=int, default=3,
                        help="Number of targets for kill scenario")
    args = parser.parse_args()

    log(f"{Colors.BOLD}{'═'*60}")
    log(f"  TikiClone Chaos Engineering Test Suite")
    log(f"{'═'*60}{Colors.RESET}")
    log(f"Scenario: {args.scenario}")
    log(f"Time: {datetime.now().isoformat()}")
    print()

    engine = ChaosEngine()

    # Handle Ctrl+C gracefully
    def signal_handler(sig, frame):
        log("Chaos interrupted! Cleaning up...", "WARN")
        run("docker exec tikiclone-gateway-1 sh -c 'tc qdisc del dev eth0 root 2>/dev/null'", timeout=5)
        for svc in ALL_SERVICES:
            c = get_container_name(svc)
            run(f"docker exec {c} sh -c 'tc qdisc del dev eth0 root 2>/dev/null; pkill -f md5sum 2>/dev/null'", timeout=5)
        engine.monitor.stop()
        sys.exit(1)
    signal.signal(signal.SIGINT, signal_handler)

    if args.scenario == "monitor":
        engine.monitor.start()
        log("Monitoring all services. Press Ctrl+C to stop.", "INFO")
        try:
            while True:
                time.sleep(5)
                summary = engine.monitor.get_summary()
                up = sum(1 for s in summary.values() if s.get("status") == "up")
                total = len(summary)
                log(f"Services healthy: {up}/{total}")
        except KeyboardInterrupt:
            pass
        finally:
            engine.monitor.stop()

    elif args.scenario == "kill":
        engine.scenario_kill(count=args.count, grace_period=15)

    elif args.scenario == "latency":
        target = args.target or "mysql-primary"
        latency = args.intensity or 200
        engine.scenario_latency(target, latency_ms=latency, duration=args.duration)

    elif args.scenario == "partition":
        target = args.target or "redis-master"
        engine.scenario_partition(target, duration=args.duration)

    elif args.scenario == "loss":
        target = args.target or "gateway"
        loss = args.intensity or 30
        engine.scenario_loss(target, loss_pct=loss, duration=args.duration)

    elif args.scenario == "cascade":
        engine.scenario_cascade()

    elif args.scenario == "stress":
        target = args.target or "gateway"
        cpu_pct = args.intensity or 80
        engine.scenario_stress(target, cpu_load=cpu_pct, duration=args.duration)

    elif args.scenario == "all":
        log("Running ALL chaos scenarios sequentially...", "CHAOS")
        log("=" * 60, "CHAOS")

        engine.monitor.start()
        time.sleep(3)

        # 1. Container kill
        engine.scenario_kill(count=2, grace_period=15)
        time.sleep(10)

        # 2. Latency
        engine.scenario_latency("mysql-primary", latency_ms=200, duration=20)
        time.sleep(5)

        # 3. Packet loss
        engine.scenario_loss("gateway", loss_pct=20, duration=15)
        time.sleep(5)

        # 4. Cascade
        engine.scenario_cascade()
        time.sleep(10)

        # 5. Stress
        engine.scenario_stress("gateway", cpu_load=80, duration=20)

        engine.monitor.stop()

    # Print report
    report = engine.generate_report()
    print(report)

    # Final status
    passed = sum(1 for r in engine.scenario_results if r["passed"])
    total = len(engine.scenario_results)
    if total > 0:
        if passed == total:
            log(f"ALL CHECKS PASSED: {passed}/{total}", "OK")
        else:
            log(f"SOME CHECKS FAILED: {passed}/{total}", "FAIL")

if __name__ == "__main__":
    main()
