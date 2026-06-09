#!/usr/bin/env python3
"""Quick stress test from within the swarm network."""
import http.client
import time
import statistics
from concurrent.futures import ThreadPoolExecutor, as_completed

TARGET = "tiki_gateway"
PORT = 8080
PATH = "/health"

def make_request():
    try:
        conn = http.client.HTTPConnection(TARGET, PORT, timeout=30)
        conn.request("GET", PATH)
        resp = conn.getresponse()
        body = resp.read()
        conn.close()
        return resp.status, len(body), None
    except Exception as e:
        return 0, 0, str(e)

for concurrency in [1, 10, 50, 100]:
    results = []
    start = time.time()
    with ThreadPoolExecutor(max_workers=concurrency) as pool:
        futures = [pool.submit(make_request) for _ in range(100)]
        for f in as_completed(futures):
            results.append(f.result())
    elapsed = time.time() - start
    success = len([r for r in results if r[0] == 200])
    errors = [r[2] for r in results if r[2]]
    print(f"concurrency={concurrency:3d} | success={success:3d}/100 | time={elapsed:.1f}s | rps={int(100/elapsed)} | errors={len(errors)}")
    if errors:
        print(f"  Sample error: {errors[0]}")
