package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/inventory/internal/config"
)

type Store struct {
	client *redis.Client
	cfg    config.RedisConfig
}

func NewStore(client *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{client: client, cfg: cfg}
}

func (s *Store) IsAvailable() bool { return s.client != nil }

func (s *Store) CheckIdempotencyKey(ctx context.Context, key string) (string, error) {
	if s.client == nil { return "", nil }
	return s.client.Get(ctx, fmt.Sprintf("inventory:idempotency:%s", key)).Result()
}

func (s *Store) StoreIdempotencyKey(ctx context.Context, key, reservationID string, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.Set(ctx, fmt.Sprintf("inventory:idempotency:%s", key), reservationID, ttl).Err()
}

// AcquireStockLock acquires a distributed lock with a unique token.
// Returns (token, acquired, error). The token must be used when releasing the lock.
func (s *Store) AcquireStockLock(ctx context.Context, skuID string, ttl time.Duration) (string, bool, error) {
	if s.client == nil { return "", true, nil }

	token := generateLockToken()
	key := fmt.Sprintf("lock:inventory:%s", skuID)

	ok, err := s.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil { return "", false, err }
	return token, ok, nil
}

// ReleaseStockLock releases a distributed lock only if the token matches.
// [SECURITY] Uses Lua script for atomic check-and-delete to prevent lock theft.
func (s *Store) ReleaseStockLock(ctx context.Context, skuID, token string) error {
	if s.client == nil { return nil }

	key := fmt.Sprintf("lock:inventory:%s", skuID)

	// [SECURITY] Atomic Lua script: only delete if token matches our lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return s.client.Eval(ctx, script, []string{key}, token).Err()
}

// InvalidateStockCache deletes the cache key for a SKU.
// [SECURITY] Delete instead of update prevents stale data on cache write failure.
func (s *Store) InvalidateStockCache(ctx context.Context, skuID string) error {
	if s.client == nil { return nil }
	return s.client.Del(ctx, fmt.Sprintf("stock:%s", skuID)).Err()
}

func (s *Store) CacheStock(ctx context.Context, skuID string, quantity, reserved int, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.HMSet(ctx, fmt.Sprintf("stock:%s", skuID), "quantity", quantity, "reserved", reserved, "available", quantity-reserved).Err()
}

func (s *Store) GetCachedStock(ctx context.Context, skuID string) (int, int, error) {
	if s.client == nil { return 0, 0, fmt.Errorf("redis not available") }
	vals, err := s.client.HMGet(ctx, fmt.Sprintf("stock:%s", skuID), "quantity", "reserved").Result()
	if err != nil { return 0, 0, err }
	if len(vals) < 2 {
		return 0, 0, fmt.Errorf("unexpected number of values from HMGet: got %d, want 2", len(vals))
	}
	qtyVal, ok := vals[0].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected type for quantity: %T", vals[0])
	}
	resVal, ok := vals[1].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected type for reserved: %T", vals[1])
	}
	return int(qtyVal), int(resVal), nil
}

func (s *Store) IncrementCounter(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if s.client == nil { return 0, nil }
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil { return 0, err }
	return incr.Val(), nil
}

// generateLockToken creates a cryptographically secure random token for distributed locks.
func generateLockToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
