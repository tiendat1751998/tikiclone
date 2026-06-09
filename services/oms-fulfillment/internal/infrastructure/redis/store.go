package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/config"
)

type Store struct {
	client *redis.Client
	cfg    config.RedisConfig
}

func NewStore(client *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{client: client, cfg: cfg}
}

func (s *Store) CacheFulfillment(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "fulfillment:"+key, data, ttl).Err()
}

func (s *Store) GetCachedFulfillment(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, "fulfillment:"+key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *Store) InvalidateCache(ctx context.Context, key string) error {
	return s.client.Del(ctx, "fulfillment:"+key).Err()
}

func (s *Store) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
}

func (s *Store) ReleaseLock(ctx context.Context, key string) error {
	return s.client.Del(ctx, "lock:"+key).Err()
}
