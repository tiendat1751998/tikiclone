# Benchmark Report

## Test Environment

- **Cluster**: Docker Swarm (Manager + 3 Workers)
- **Worker RAM**: 8GB each (24GB total)
- **Max Replicas**: 5 per stateless service
- **Test Date**: 2026-06-02

## Load Test Levels

| Users | TPS Target | Services Tested | Memory Limit |
|-------|------------|-----------------|------------|
| 50    | 10 TPS     | All services    | 512Mi      |
| 100   | 100 TPS    | All services    | 512Mi      |
| 250   | 500 TPS    | Core services   | 512Mi      |
| 500   | 250 TPS    | Core services   | 512Mi      |
| 1000  | 1000 TPS   | Critical paths  | 512Mi      |

## Current Baseline Results

From existing stress test (100 users):
```
Throughput: 347.51 req/s
Success Rate: 97.5%
P95 Latency: 447.59ms
P99 Latency: 610.90ms
```

## Expected Benchmarks

### Gateway Service (5 replicas)
- **Memory Limit**: 512Mi
- **CPU Target**: < 70%
- **Expected P95**: < 100ms
- **Expected P99**: < 200ms

### Product Service (5 replicas)
- **Memory Limit**: 512Mi
- **CPU Target**: < 70%
- **Expected P95**: < 150ms
- **Expected P99**: < 300ms

### Cart Service (5 replicas)
- **Memory Limit**: 512Mi
- **CPU Target**: < 70%
- **Expected P95**: < 200ms
- **Expected P99**: < 400ms

### Auth Service (3 replicas)
- **Memory Limit**: 256Mi
- **CPU Target**: < 70%
- **Expected P95**: < 100ms
- **Expected P99**: < 200ms

### Order/Checkout Services (3 replicas each)
- **Memory Limit**: 512Mi
- **CPU Target**: < 70%
- **Expected P95**: < 250ms
- **Expected P99**: < 500ms

## Resource Utilization Targets

| Resource | Target | Warning | Critical |
|----------|--------|---------|----------|
| CPU | < 70% | 70-85% | > 85% |
| Memory | < 75% | 75-85% | > 85% |
| Disk | < 80% | 80-90% | > 90% |
| Network | < 70% | 70-85% | > 85% |

## Memory Allocation Per Worker (8GB)

Total memory budget per worker: ~5GB allocated

| Service | Replicas | Memory Limit | Total |
|---------|----------|--------------|-------|
| Gateway | 2 | 512Mi | ~1GB |
| Product | 2 | 512Mi | ~1GB |
| Cart | 2 | 512Mi | ~1GB |
| Auth | 1 | 256Mi | ~256Mi |
| Search | 1 | 256Mi | ~256Mi |
| Category | 1 | 256Mi | ~256Mi |
| Others | 1 each | 256Mi | ~512Mi |
| **TOTAL** | | | **~5GB** |

## Test Execution Commands

```bash
# Run all levels
python3 chaos-tests/run_load_test.py --level all

# Run specific level
python3 chaos-tests/run_load_test.py --level 100
python3 chaos-tests/run_load_test.py --level 500
python3 chaos-tests/run_load_test.py --level 1000

# Results saved to chaos-tests/results/
```

## Success Criteria

- All services start successfully
- 5 replicas operate correctly per service
- No OOM events
- Success rate > 95%
- P95 latency < service-specific targets
- No duplicate orders/payments
- No overselling detected
- Automatic recovery on failure

## Scaling Validation Matrix

| Test | Result | Notes |
|------|--------|-------|
| Gateway x5 | Pending | Verify load distribution |
| Product x5 | Pending | Verify product listing |
| Cart x5 | Pending | Verify cart operations |
| Auth x3 | Pending | Verify login/registration |
| Order x3 | Pending | Verify order creation |
| Checkout x3 | Pending | Verify checkout flow |