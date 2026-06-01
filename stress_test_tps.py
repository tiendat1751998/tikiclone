"""
TikiClone High-TPS Stress Test
Tests: 100, 1000, 2000, 5000, 10000 TPS
"""
import concurrent.futures
import json
import random
import ssl
import statistics
import time
import urllib.request
from collections import defaultdict
import sys

# ==== Multi-level Configuration ====
TTPS_CONFIGS = [
    {"users": 20, "ramp": 5, "hold": 30},
    {"users": 100, "ramp": 10, "hold": 60},
    {"users": 200, "ramp": 15, "hold": 60},
    {"users": 500, "ramp": 30, "hold": 90},
    {"users": 1000, "ramp": 30, "hold": 90},
]

TARGET_HOST = "https://localhost:8443"

SSL_CTX = ssl.create_default_context()
SSL_CTX.check_hostname = False
SSL_CTX.verify_mode = ssl.CERT_NONE

PRODUCT_IDS = [f"9ebcd126-44ce-4066-928c-e71d4acb4811", f"d4edad4a-9e62-46e5-b26f-03c595a537db"][:10]
CATEGORY_SLUGS = ["dien-thoai", "laptop", "thoi-trang-nu", "giay-dep-nam"][:4]

def make_request(path):
    url = f"{TARGET_HOST}{path}"
    start = time.time()
    try:
        req = urllib.request.Request(url)
        resp = urllib.request.urlopen(req, timeout=5, context=SSL_CTX)
        return {"status": resp.status, "latency": time.time() - start, "error": None}
    except urllib.error.HTTPError as e:
        return {"status": e.code, "latency": time.time() - start, "error": None}
    except Exception as e:
        return {"status": 0, "latency": time.time() - start, "error": type(e).__name__}

def user_session(user_id, test_start, duration):
    results = []
    while time.time() - test_start < duration:
        task = random.choices(
            ["/" , "/api/v1/products", f"/api/v1/products/{random.choice(PRODUCT_IDS)}",
             f"/api/v1/products?category_slug={random.choice(CATEGORY_SLUGS)}"],
            weights=[40, 40, 15, 5]
        )[0]
        results.append({"name": task, **make_request(task)})
        if random.random() < 0.1:  # 10% think time
            time.sleep(random.uniform(0.05, 0.2))
    return results

def run_test(config):
    users, ramp, hold = config["users"], config["ramp"], config["hold"]
    total_dur = ramp + hold
    print(f"\n{'='*60}")
    print(f"  TPS TEST - {users} users, {hold}s hold")
    print(f"{'='*60}")
    
    test_start = time.time()
    all_results = []
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=users) as executor:
        futures = []
        for i in range(users):
            if i > 0 and ramp > 0:
                time.sleep(ramp / users)
            futures.append(executor.submit(user_session, i, test_start, total_dur))
        
        for f in concurrent.futures.as_completed(futures, timeout=total_dur + 30):
            all_results.extend(f.result())
    
    dur = time.time() - test_start
    success = len([r for r in all_results if 200 <= r["status"] < 400])
    latencies = [r["latency"] for r in all_results]
    
    print(f"  Requests: {len(all_results)}")
    print(f"  Throughput: {len(all_results)/dur:.1f} req/s")
    print(f"  Success: {success/len(all_results)*100:.1f}%")
    print(f"  p95: {sorted(latencies)[int(len(latencies)*0.95)]*1000:.1f}ms" if latencies else "N/A")
    print(f"  p99: {sorted(latencies)[int(len(latencies)*0.99)]*1000:.1f}ms" if latencies else "N/A")
    
    return {
        "users": users, "requests": len(all_results),
        "throughput": len(all_results)/dur, "success_rate": success/len(all_results) if all_results else 0,
        "p95_ms": sorted(latencies)[int(len(latencies)*0.95)]*1000 if latencies else 0,
        "p99_ms": sorted(latencies)[int(len(latencies)*0.99)]*1000 if latencies else 0
    }

if __name__ == "__main__":
    level = int(sys.argv[1]) if len(sys.argv) > 1 else 0
    results = [run_test(TTPS_CONFIGS[level])]
    with open("/home/datdt/tikiclone/stress_results.json", "w") as f:
        json.dump(results, f, indent=2)