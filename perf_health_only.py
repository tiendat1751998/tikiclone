"""
Health-only test to measure max gateway throughput
"""
import asyncio
import aiohttp
import ssl
import time
import json
import statistics

TARGET = "https://10.10.10.133/health"

async def worker(session, duration, worker_id):
    end_time = time.time() + duration
    results = []
    while time.time() < end_time:
        start = time.time()
        try:
            async with session.get(TARGET, timeout=aiohttp.ClientTimeout(total=2)) as resp:
                await resp.read()
                results.append({"status": resp.status, "latency": time.time() - start, "error": None})
        except Exception as e:
            results.append({"status": 0, "latency": time.time() - start, "error": type(e).__name__})
    return results

async def run_test(workers, duration):
    ssl_ctx = ssl.create_default_context()
    ssl_ctx.check_hostname = False
    ssl_ctx.verify_mode = ssl.CERT_NONE
    connector = aiohttp.TCPConnector(limit=0, limit_per_host=0, keepalive_timeout=60, ssl=ssl_ctx)
    
    async with aiohttp.ClientSession(connector=connector) as session:
        test_start = time.time()
        tasks = [worker(session, duration, i) for i in range(workers)]
        all_results = await asyncio.gather(*tasks)
        test_duration = time.time() - test_start

    results = []
    for wr in all_results:
        results.extend(wr)

    total = len(results)
    errors = sum(1 for r in results if r["error"] or r["status"] >= 400)
    latencies = sorted([r["latency"] for r in results])
    throughput = total / test_duration
    sr = (total - errors) / total * 100

    print(f"  Workers: {workers} | TPS: {throughput:.0f} | SR: {sr:.1f}%")
    print(f"  p50: {latencies[len(latencies)//2]*1000:.1f}ms | p95: {latencies[int(len(latencies)*0.95)]*1000:.1f}ms | p99: {latencies[int(len(latencies)*0.99)]*1000:.1f}ms")
    return throughput

async def main():
    print("=== Health endpoint max throughput test ===")
    for workers in [100, 200, 500, 1000, 2000]:
        tps = await run_test(workers, 10)
        if tps < 500:
            break

asyncio.run(main())
