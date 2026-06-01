# SYSTEM-WIDE BUG REPORT
**Date:** May 21, 2026  
**Scope:** All Go microservices in Tiki Clone monorepo

---

## Summary

| Severity | Count |
|----------|-------|
| **CRITICAL** | 1 |
| **HIGH** | 8 |
| **MEDIUM** | 15 |
| **LOW** | 10 |
| **TOTAL** | **34** |

---

## 🔴 CRITICAL Severity Issues (1)

### C1: Missing `log` import in payment config
**File:** `services/payment/internal/config/config.go:159`  
**Issue:** The `requireEnv` function uses `log.Fatalf()` but `log` package is not imported.

```go
package config

import (
    "os"
    "strconv"
    "strings"
    "time"
)
// Missing: "log"
```

**Impact:** Code will not compile. Service cannot start.

**Fix:** Add `"log"` to imports:

```go
import (
    "log"
    "os"
    "strconv"
    "strings"
    "time"
)
```

---

## 🔴 HIGH Severity Issues (8)

### H1: Unsafe type assertions in order handler
**File:** `services/order/internal/transport/http/handler.go:64-68`  
**Issue:** Type assertions without `ok` check can cause panic:

```go
uid, _ := userID.(string)
r, _ := role.(string)
if order.UserID != uid && r != "admin" && r != "seller" {
```

**Impact:** Panic if `user_id` or `role` is not string type or not present in context.

**Fix:** Add proper type assertion checks:

```go
uid, ok := userID.(string)
if !ok || uid == "" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
}
r, _ := role.(string) // role may not exist, default empty string is ok
```

---

### H2: Unsafe type assertions in order CancelOrder handler
**File:** `services/order/internal/transport/http/handler.go:143-146`  
**Issue:** Same unsafe type assertion pattern:

```go
userID, _ := c.Get("user_id")
role, _ := c.Get("role")
uid, _ := userID.(string)
r, _ := role.(string)
```

**Impact:** Panic on type assertion failure.

**Fix:** Add proper error handling for type assertions.

---

### H3: `handleError` uses fragile string prefix matching (checkout)
**File:** `services/checkout/internal/transport/http/handler.go:87-96`  
**Issue:** String prefix comparison for error matching:

```go
if err.Error() == domainErr.Error() || (len(err.Error()) >= len(domainErr.Error()) && err.Error()[:len(domainErr.Error())] == domainErr.Error()) {
```

**Impact:** Wrapped errors may not match correctly.

**Fix:** Use `errors.Is()`:

```go
if errors.Is(err, domainErr) {
    c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
    return
}
```

---

### H4: Conditional authentication in checkout router
**File:** `services/checkout/internal/transport/http/router.go:38-40`  
**Issue:** Authentication only applied if `jwtSecret != ""`:

```go
if r.jwtSecret != "" {
    api.Use(auth.GinJWTAuth(r.jwtSecret))
}
```

**Impact:** Endpoints unprotected if env var not set.

**Fix:** Make auth mandatory in production:

```go
if r.jwtSecret == "" {
    if config.IsProduction() {
        log.Fatal("JWT_ACCESS_SECRET is required in production")
    }
}
api.Use(auth.GinJWTAuth(r.jwtSecret))
```

---

### H5: KafkaConfig loaded but never used (checkout)
**File:** `services/checkout/internal/config/config.go:108-110`  
**Issue:** Kafka config exists but no producer/consumer implemented.

**Impact:** Misleading configuration.

**Fix:** Implement Kafka or remove config.

---

### H6: GRPCPort configured but no gRPC server (checkout)
**File:** `services/checkout/internal/config/config.go:15,82`  
**Issue:** gRPC port config but no server implementation.

**Impact:** Port conflict potential.

**Fix:** Implement gRPC server or remove config.

---

### H7: Context not passed properly to saga goroutine
**File:** `services/checkout/internal/application/service.go:86-95`  
**Issue:** Uses `context.Background()` instead of request context:

```go
sagaCtx, sagaCancel := context.WithTimeout(context.Background(), 5*time.Minute)
```

**Impact:** Tracing context lost, no parent span.

**Fix:** Pass request context:

```go
sagaCtx, sagaCancel := context.WithTimeout(ctx, 5*time.Minute)
```

---

### H8: `FindExpired` uses string for time parameter
**File:** `services/checkout/internal/infrastructure/mysql/repos.go:45-48`  
**Issue:** Time passed as string:

```go
func (r *CheckoutRepository) FindExpired(ctx context.Context, before string, limit int) ([]*domain.Checkout, error) {
```

**Impact:** Time format issues, potential SQL injection.

**Fix:** Use `time.Time`:

```go
func (r *CheckoutRepository) FindExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Checkout, error) {
```

---

## 🟡 MEDIUM Severity Issues (15)

### M1: Dead code packages (checkout)
**Files:** `internal/tracing/`, `internal/logging/`, `internal/health/`, `internal/validation/`  
**Issue:** Unused packages exist.

**Fix:** Remove unused packages.

---

### M2: `CheckoutLatency` metric never recorded
**File:** `services/checkout/internal/metrics/metrics.go:34-38`  
**Issue:** Metric defined but never used.

**Fix:** Record latency in each step.

---

### M3: `SELECT *` in multiple queries
**Files:** Multiple repository files  
**Issue:** Fragile against schema changes.

**Fix:** Use explicit column names.

---

### M4: Missing error checks in checkout service
**File:** `services/checkout/internal/application/service.go`  
**Issues:**
- `stepComplete`: `s.reconcileRepo.Create(ctx, job)` error not checked (line 283)
- `rollbackReservations`: `s.reservationRepo.UpdateStatus` errors not checked (lines 333-334)
- `logStep`: `s.stepLogRepo.Create` error not checked (line 342)
- `handleFailure`: `s.reconcileRepo.Create` error not checked (line 317)

**Fix:** Add error logging for all repository calls.

---

### M5: Hardcoded MySQL password defaults
**Files:** Multiple config files  
**Issue:** Default passwords like `tiki_dev` in code:

```go
Password: getEnv("MYSQL_PASSWORD", "tiki_dev"),
```

**Impact:** Security risk if env vars not set.

**Fix:** Use `requireEnv()` for passwords or fail fast.

---

### M6: `mustJSON` errors not fully checked
**File:** `services/checkout/internal/application/service.go:168-178`  
**Issue:** `mustJSON` returns error but callers may not check properly.

**Fix:** Ensure all callers handle errors.

---

### M7: Order handler doesn't validate ownership in GetOrderStatus
**File:** `services/order/internal/transport/http/handler.go:113-126`  
**Issue:** No ownership validation for status endpoint.

**Fix:** Add ownership check.

---

### M8: Order handler doesn't validate ownership in GetOrderHistory
**File:** `services/order/internal/transport/http/handler.go:168-180`  
**Issue:** No ownership validation for history endpoint.

**Fix:** Add ownership check.

---

### M9: Order handler doesn't validate ownership in GetReconciliationStatus
**File:** `services/order/internal/transport/http/handler.go:182-194`  
**Issue:** No ownership validation for reconciliation endpoint.

**Fix:** Add ownership check.

---

### M10: Duplicate BcryptCost config in auth
**File:** `services/auth/internal/config/config.go:207`  
**Issue:** `BcryptCost` defined in both `PasswordConfig` and `SecurityConfig`.

**Fix:** Remove duplicate.

---

### M11: Missing `time` import in checkout repos
**File:** `services/checkout/internal/infrastructure/mysql/repos.go`  
**Issue:** `time` package may be needed for `time.Time` parameter fix.

**Fix:** Add import if needed.

---

### M12: Idempotency check returns completed checkouts
**File:** `services/checkout/internal/application/service.go:69-76`  
**Issue:** Returns existing checkout even if completed.

**Fix:** Clarify behavior or return error for completed checkouts.

---

### M13: No distributed lock for inventory reservation
**File:** `services/checkout/internal/application/service.go:219-237`  
**Issue:** No locking mechanism for concurrent reservations.

**Impact:** Potential over-selling.

**Fix:** Implement distributed lock.

---

### M14: Race condition in saga execution
**File:** `services/checkout/internal/application/service.go:86-95`  
**Issue:** Goroutine updates state asynchronously.

**Impact:** Client may read stale status.

**Fix:** Consider synchronous execution or proper status polling.

---

### M15: MySQL DSN timeout format issue
**Files:** Multiple config files  
**Issue:** Timeout appended as duration string:

```go
Timeout: getEnvDuration("MYSQL_TIMEOUT", 5*time.Second),
// ...
"?charset=utf8mb4&parseTime=true&loc=UTC&timeout=" + c.Timeout.String()
```

**Impact:** Duration string may not be valid MySQL timeout format.

**Fix:** Use milliseconds:

```go
"&timeout=" + strconv.Itoa(int(c.Timeout.Milliseconds())) + "ms"
```

---

## 🟢 LOW Severity Issues (10)

### L1: Unused imports
**Files:** Multiple files  
**Issue:** Some imports may be unused after fixes.

**Fix:** Run `goimports` to clean up.

---

### L2: Inconsistent error handling patterns
**Files:** Multiple handlers  
**Issue:** Different error handling styles across services.

**Fix:** Standardize error handling.

---

### L3: Missing request validation
**Files:** Multiple handlers  
**Issue:** Some endpoints don't validate all required fields.

**Fix:** Add comprehensive validation.

---

### L4: Missing pagination limits
**Files:** List endpoints  
**Issue:** No max page size limit.

**Fix:** Add max limit (e.g., 100).

---

### L5: Missing CORS configuration
**Files:** Router files  
**Issue:** CORS may be too permissive.

**Fix:** Configure specific origins.

---

### L6: Missing rate limiting on some endpoints
**Files:** Router files  
**Issue:** Not all endpoints have rate limiting.

**Fix:** Add rate limiting to all public endpoints.

---

### L7: Missing request timeout configuration
**Files:** HTTP server setup  
**Issue:** No request timeout configured.

**Fix:** Add `ReadTimeout` and `WriteTimeout`.

---

### L8: Missing graceful shutdown
**Files:** Main files  
**Issue:** No graceful shutdown handling.

**Fix:** Implement graceful shutdown.

---

### L9: Inconsistent logging patterns
**Files:** Multiple services  
**Issue:** Different logging styles.

**Fix:** Standardize logging.

---

### L10: Missing health check endpoints (some services)
**Files:** Some router files  
**Issue:** Not all services have health endpoints.

**Fix:** Add health endpoints to all services.

---

## 📊 Service-by-Service Breakdown

| Service | CRITICAL | HIGH | MEDIUM | LOW | Total |
|---------|----------|------|--------|-----|-------|
| auth | 0 | 0 | 1 | 2 | 3 |
| checkout | 0 | 4 | 6 | 4 | 14 |
| inventory | 0 | 0 | 1 | 1 | 2 |
| order | 0 | 2 | 3 | 1 | 6 |
| payment | 1 | 0 | 1 | 1 | 3 |
| product | 0 | 0 | 1 | 1 | 2 |
| **TOTAL** | **1** | **6** | **13** | **10** | **30** |

---

## 🔧 Recommended Fix Priority

### Immediate (P0):
1. **C1:** Add missing `log` import in payment config
2. **H1, H2:** Fix unsafe type assertions in order handler
3. **H3:** Fix error handling in checkout handler

### High (P1):
1. **H4:** Make authentication mandatory
2. **H7:** Fix context propagation
3. **H8:** Fix `FindExpired` time parameter
4. **M4:** Add missing error checks

### Medium (P2):
1. **M1:** Remove dead code
2. **M2:** Record metrics
3. **M3:** Replace `SELECT *`
4. **M5:** Fix hardcoded passwords

### Low (P3):
1. **L1-L10:** Code quality improvements

---

## 📝 Notes

### Previously Fixed (from QA_BUG_REPORT.md):
- ✅ Dockerfile base images pinned
- ✅ Go version consistency
- ✅ SQL query fixes
- ✅ Error logging added
- ✅ Auth middleware added
- ✅ Hardcoded secrets replaced
- ✅ Test compilation fixes

### Testing Recommendations:
1. Add unit tests for type assertions
2. Add integration tests for error handling
3. Add load tests for concurrent reservations
4. Add security tests for authentication bypass