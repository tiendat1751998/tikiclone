package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/payment/internal/config"
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
	return s.client.Get(ctx, fmt.Sprintf("payment:idempotency:%s", key)).Result()
}

func (s *Store) StoreIdempotencyKey(ctx context.Context, key, paymentID string, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.Set(ctx, fmt.Sprintf("payment:idempotency:%s", key), paymentID, ttl).Err()
}

func (s *Store) CheckWebhookReplay(ctx context.Context, key string) (bool, error) {
	if s.client == nil { return false, nil }
	exists, err := s.client.Exists(ctx, fmt.Sprintf("webhook:%s", key)).Result()
	return exists > 0, err
}

func (s *Store) MarkWebhookProcessed(ctx context.Context, key string, ttl time.Duration) error {
	if s.client == nil { return nil }
	return s.client.Set(ctx, fmt.Sprintf("webhook:%s", key), "1", ttl).Err()
}

func (s *Store) AcquirePaymentLock(ctx context.Context, orderID string, ttl time.Duration) (token string, ok bool, err error) {
	if s.client == nil { return "", true, nil }
	token = generateLockToken()
	ok, err = s.client.SetNX(ctx, fmt.Sprintf("lock:payment:%s", orderID), token, ttl).Result()
	return token, ok, err
}

func (s *Store) ReleasePaymentLock(ctx context.Context, orderID, token string) error {
	if s.client == nil || token == "" { return nil }
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return s.client.Eval(ctx, script, []string{fmt.Sprintf("lock:payment:%s", orderID)}, token).Err()
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

func generateLockToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
