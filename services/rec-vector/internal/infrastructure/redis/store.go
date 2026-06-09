package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/rec-vector/internal/config"
)

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{client: client}
}

func (s *Store) CacheRecommendations(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "rec:"+key, data, ttl).Err()
}

func (s *Store) GetCachedRecommendations(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, "rec:"+key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *Store) InvalidateCache(ctx context.Context, key string) error {
	return s.client.Del(ctx, "rec:"+key).Err()
}
