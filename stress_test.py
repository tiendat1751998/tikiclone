import concurrent.futures, json, random, ssl, time, urllib.parse, urllib.request

TARGET_HOST = "https://localhost:8443"
NUM_WORKERS = 300
RAMP_UP_SEC = 5
HOLD_TIME_SEC = 60

PRODUCT_IDS = ["9ebcd126-44ce-4066-928c-e71d4acb4811"]
SORT_OPTIONS = [("price", "ASC"), ("price", "DESC"), ("newest", "")]

SSL_CTX = ssl.create_default_context(); SSL_CTX.check_hostname = False; SSL_CTX.verify_mode = ssl.CERT_NONE
all_results = []

def make_request(name, path):
    url = f"{TARGET_HOST}{path}"; start = time.time()
    try:
        resp = urllib.request.urlopen(urllib.request.Request(url), timeout=5, context=SSL_CTX)
        return {"status": resp.status, "latency": time.time() - start, "error": None}
    except urllib.error.HTTPError as e:
        return {"status": e.code, "latency": time.time() - start, "error": None}
    except Exception as ex:
        return {"status": 0, "latency": time.time() - start, "error": type(ex).__name__}

def user_session(start_time):
    local = []
    while time.time() - start_time < RAMP_UP_SEC + HOLD_TIME_SEC:
        r = random.randint(1, 100)
        if r <= 40:
            local.append(make_request("Products_List", f"/api/v1/products?page=1&size=20"))
        elif r <= 60:
            sb, so = random.choice(SORT_OPTIONS); p = f"?page=1&size=20&sort_by={sb}"
            if so: p += f"&sort_order={so}"
            local.append(make_request("Products_Sorted", f"/api/v1/products{p}"))
        else:
            local.append(make_request("Categories_List", "/api/v1/categories"))
        time.sleep(random.uniform(0.05, 0.2))
    return local

if __name__ == "__main__":
    test_start = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=NUM_WORKERS) as ex:
        futures = [ex.submit(user_session, test_start) for _ in range(NUM_WORKERS)]
        for f in concurrent.futures.as_completed(futures): all_results.extend(f.result())
    total = len(all_results); succ = [r for r in all_results if 200 <= r["status"] < 400]
    lat = [r["latency"] for r in all_results]
    print(f"Throughput: {total/65:.1f} req/s | Success: {100*len(succ)/total:.1f}% | p95: {sorted(lat)[int(len(lat)*0.95)]*1000:.1f}ms")
