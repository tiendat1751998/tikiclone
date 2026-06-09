import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const p99Trend = new Trend('p99_latency');
const cacheHitRate = new Rate('cache_hits');

// Test configuration - targets 1000-10000 RPS
export const options = {
  scenarios: {
    // Ramp up to 1000 RPS over 2 minutes
    ramp_up: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { duration: '30s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '60s', target: 2000 },
        { duration: '30s', target: 0 },
      ],
    },
    // Sustained load at 500 RPS for 2 minutes
    sustained: {
      executor: 'constant-arrival-rate',
      rate: 500,
      timeUnit: '1s',
      duration: '2m',
      startTime: '3m',
      preAllocatedVUs: 200,
      maxVUs: 500,
    },
  },
  thresholds: {
    // Performance targets: p99 < 20ms, error < 0.1%
    'http_req_duration': ['p(99)<20'],
    'errors': ['rate<0.001'],
    'http_req_failed': ['rate<0.001'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_PREFIX = '/api/v1';

// Product IDs from seed data
const PRODUCT_IDS = ['prod-001', 'prod-002', 'prod-003'];
const CATEGORY_IDS = ['cat-001', 'cat-002'];
const SELLER_IDS = ['shop-001', 'shop-002'];

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    timeout: '5s',
  };

  group('Product List', () => {
    // Product list with various sort options (tests MongoDB indexes)
    const sortOptions = ['', 'created_at', 'price', 'popularity'];
    const sortBy = sortOptions[Math.floor(Math.random() * sortOptions.length)];

    let url = `${BASE_URL}${API_PREFIX}/products?page=1&size=20`;
    if (sortBy) url += `&sort_by=${sortBy}&sort_order=DESC`;

    const res = http.get(url, params);
    const success = check(res, {
      'product list status 200': (r) => r.status === 200,
      'product list has data': (r) => r.json('products') !== undefined,
      'product list p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
    p99Trend.add(res.timings.duration);
  });

  group('Product Detail', () => {
    const productId = PRODUCT_IDS[Math.floor(Math.random() * PRODUCT_IDS.length)];
    const res = http.get(`${BASE_URL}${API_PREFIX}/products/${productId}`, params);
    const success = check(res, {
      'product detail status 200': (r) => r.status === 200,
      'product detail p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
  });

  group('Deals', () => {
    // Tests deals index (idx_deals_newest)
    const res = http.get(`${BASE_URL}${API_PREFIX}/products/deals?limit=20`, params);
    const success = check(res, {
      'deals status 200': (r) => r.status === 200,
      'deals p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
  });

  group('Category Products', () => {
    // Tests category+popularity index (idx_status_cat_popularity)
    const catId = CATEGORY_IDS[Math.floor(Math.random() * CATEGORY_IDS.length)];
    const res = http.get(`${BASE_URL}${API_PREFIX}/products?category_id=${catId}&sort_by=popularity&sort_order=DESC&page=1&size=20`, params);
    const success = check(res, {
      'category products status 200': (r) => r.status === 200,
      'category products p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
  });

  group('Seller Products', () => {
    // Tests seller index (idx_status_seller_newest, idx_status_seller_popularity)
    const sellerId = SELLER_IDS[Math.floor(Math.random() * SELLER_IDS.length)];
    const res = http.get(`${BASE_URL}${API_PREFIX}/products?seller_id=${sellerId}&sort_by=created_at&sort_order=DESC&page=1&size=20`, params);
    const success = check(res, {
      'seller products status 200': (r) => r.status === 200,
      'seller products p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
  });

  group('Cart Operations', () => {
    // Tests cart incremental totals (no recalculateCart)
    const cartId = `cart-${Math.floor(Math.random() * 1000)}`;
    const res = http.get(`${BASE_URL}${API_PREFIX}/cart/${cartId}`, params);
    check(res, {
      'cart status 200 or 404': (r) => r.status === 200 || r.status === 404,
      'cart p99 < 20ms': (r) => r.timings.duration < 20,
    });
  });

  group('Categories', () => {
    const res = http.get(`${BASE_URL}${API_PREFIX}/categories`, params);
    const success = check(res, {
      'categories status 200': (r) => r.status === 200,
      'categories p99 < 20ms': (r) => r.timings.duration < 20,
    });
    errorRate.add(!success);
  });

  sleep(0.1); // 100ms think time between iterations
}
