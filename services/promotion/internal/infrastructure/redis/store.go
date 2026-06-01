package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/promotion/internal/config"
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

func (s *Store) IncrementVoucherUsage(ctx context.Context, voucherID string) error {
	key := fmt.Sprintf("voucher_usage:%s", voucherID)
	return s.rdb.Incr(ctx, key).Err()
}

func (s *Store) IncrementUserVoucherUsage(ctx context.Context, userID, voucherID string) error {
	key := fmt.Sprintf("user_voucher:%s:%s", userID, voucherID)
	return s.rdb.Incr(ctx, key).Err()
}

func (s *Store) GetUserVoucherCount(ctx context.Context, userID, voucherID string) (int64, error) {
	key := fmt.Sprintf("user_voucher:%s:%s", userID, voucherID)
	return s.rdb.Get(ctx, key).Int64()
}

func (s *Store) SetVoucherCache(ctx context.Context, code string, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("voucher:%s", code)
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *Store) GetVoucherCache(ctx context.Context, code string) ([]byte, error) {
	key := fmt.Sprintf("voucher:%s", code)
	return s.rdb.Get(ctx, key).Bytes()
}

func (s *Store) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.rdb.SetNX(ctx, "idempotency:"+key, "1", ttl).Result()
}

func (s *Store) AllowRateLimit(ctx context.Context, key string, maxPerMinute int) (bool, error) {
	now := time.Now().Unix() / 60
	rateKey := fmt.Sprintf("ratelimit:%s:%d", key, now)
	pipe := s.rdb.Pipeline()
	incr := pipe.Incr(ctx, rateKey)
	pipe.Expire(ctx, rateKey, 2*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return int(incr.Val()) <= maxPerMinute, nil
}
