# Race Condition Analysis

## Critical Sections Analysis

### 1. Inventory Management

**Risk**: Overselling when multiple orders attempt to reserve the same stock

**Current State**: Uses Redis with atomic operations

**Validation Required**:
```go
// Check inventory service for race conditions
// File: services/inventory/internal/repository/stock.go

func (r *stockRepo) ReserveStock(ctx context.Context, productID string, quantity int) error {
    // Must use WATCH/MULTI/EXEC or Lua script for atomicity
    // Verify this is implemented correctly
}
```

**Test Cases**:
- Concurrent cart updates for same product
- High-frequency inventory decrements
- WebSocket inventory sync under load

### 2. Order Creation

**Risk**: Duplicate orders, duplicate payments

**Critical Operations**:
- Order number generation (must be unique)
- Payment processing (idempotent required)
- Order state transitions

**Validation**:
```
Order ID generation:
- UUID v4 or Snowflake IDs
- Database unique constraints
- Check for duplicates in result set
```

### 3. Cart Operations

**Risk**: Session inconsistency, cart data corruption

**Validation Required**:
- Concurrent cart modifications
- Redis transaction atomicity
- User session isolation

### 4. Payment Processing

**Risk**: Duplicate payments

**Validation Required**:
- Idempotency keys on payment endpoint
- Transaction rollback on failure
- Payment status check before charge

## Race Detection Commands

```bash
# Test concurrent order creation (500 users)
locust -f locustfile.py \
  --headless \
  --users 250 \
  --spawn-rate 50 \
  --run-time 60s \
  --tags OrderCreation

# Test inventory race conditions
locust -f locustfile.py \
  --headless \
  --users 100 \
  --spawn-rate 20 \
  --run-time 30s \
  --tags AddToCart

# Test payment idempotency
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 10 \
  --run-time 30s \
  --tags Checkout
```

## Validation Queries

### Check for Duplicate Orders

```sql
SELECT order_number, COUNT(*) as cnt 
FROM orders 
GROUP BY order_number 
HAVING cnt > 1;
```

### Check for Race Conditions in Inventory

```bash
# Monitor for negative stock values
docker exec mysql-primary mysql -u tiki -ptiki_dev tiki_platform \
  -e "SELECT product_id, quantity FROM inventory WHERE quantity < 0;"

# Check reservation table for orphans
docker exec mysql-primary mysql -u tiki -ptiki_dev tiki_inventory \
  -e "SELECT * FROM stock_reservations WHERE expires_at < NOW() AND status = 'reserved';"
```

## Concurrency Safeties

| Service | Lock Strategy | Atomic Operations | Notes |
|---------|--------------|-------------------|-------|
| Gateway | Redis lock | N/A | Rate limiting, circuit breaker |
| Auth | DB tx | UUID generation | Session tokens |
| Cart | Redis WATCH | Lua scripts | Cart merge, TTL cleanup |
| Order | DB tx | INSERT with unique | Order number uniqueness |
| Payment | Idempotent | External API call | Payment intent deduplication |
| Inventory | Redis lock | DECR with check | Stock reservation |

## Deadlock Scenarios

### Known Risk Areas

1. **MySQL Deadlocks** - Foreign key constraints on order->cart->product
2. **Redis Lock Contention** - High-frequency cart updates
3. **Kafka Partition Blocking** - Message ordering waits

### Prevention

- Connection timeouts: 30s
- Retry logic with exponential backoff
- Circuit breaker for external dependencies

## Metrics to Monitor

```
race_condition_checks{
  duplicate_orders_total,
  oversold_items_total,
  inventory_corruption_total,
  payment_duplicates_total,
  session_conflicts_total
}
```

## Recovery Procedures

If race condition detected:

1. **Immediate**: Rollback transaction, log incident
2. **Short-term**: Implement distributed lock
3. **Long-term**: Add idempotency key validation