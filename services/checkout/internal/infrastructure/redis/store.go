package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/checkout/internal/config"
)

type Store struct {
	rdb *redis.Client
	cfg config.RedisConfig
}

func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

func (s *Store) Ping(ctx context.Context) error { return s.rdb.Ping(ctx).Err() }
func (s *Store) Close() error                   { return s.rdb.Close() }

func (s *Store) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.rdb.SetNX(ctx, "idempotency:"+key, "1", ttl).Result()
}

func (s *Store) GetIdempotencyResult(ctx context.Context, key string) ([]byte, error) {
	return s.rdb.Get(ctx, "idempotency_result:"+key).Bytes()
}

func (s *Store) SetIdempotencyResult(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	return s.rdb.Set(ctx, "idempotency_result:"+key, result, ttl).Err()
}

func (s *Store) AcquireLock(ctx context.Context, resource string, ttl time.Duration) (string, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	key := "lock:" + resource
	ok, err := s.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil { return "", err }
	if !ok { return "", fmt.Errorf("lock already held") }
	return token, nil
}

func (s *Store) ReleaseLock(ctx context.Context, resource, token string) error {
	key := "lock:" + resource
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return s.rdb.Eval(ctx, script, []string{key}, token).Err()
}
