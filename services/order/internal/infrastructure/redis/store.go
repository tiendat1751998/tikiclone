package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/services/order/internal/domain"
)

type Store struct {
	client *redis.Client
	cfg    config.RedisConfig
}

func NewStore(client *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{client: client, cfg: cfg}
}

func (s *Store) IsAvailable() bool {
	return s.client != nil
}

func (s *Store) Ping(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("redis not available")
	}
	return s.client.Ping(ctx).Err()
}

// Order cache
func (s *Store) CacheOrder(ctx context.Context, order *domain.Order, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}
	key := fmt.Sprintf("order:%s", order.ID)
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Store) GetCachedOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	if s.client == nil {
		return nil, fmt.Errorf("redis not available")
	}
	key := fmt.Sprintf("order:%s", orderID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *Store) InvalidateOrderCache(ctx context.Context, orderID string) error {
	if s.client == nil {
		return nil
	}
	key := fmt.Sprintf("order:%s", orderID)
	return s.client.Del(ctx, key).Err()
}

// Idempotency
func (s *Store) CheckIdempotencyKey(ctx context.Context, key string) (string, error) {
	if s.client == nil {
		return "", nil
	}
	redisKey := fmt.Sprintf("idempotency:%s", key)
	return s.client.Get(ctx, redisKey).Result()
}

func (s *Store) StoreIdempotencyKey(ctx context.Context, key, orderID string, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}
	redisKey := fmt.Sprintf("idempotency:%s", key)
	return s.client.Set(ctx, redisKey, orderID, ttl).Err()
}

// Distributed lock for state transitions
func (s *Store) AcquireTransitionLock(ctx context.Context, orderID string, ttl time.Duration) (bool, error) {
	if s.client == nil {
		return true, nil
	}
	key := fmt.Sprintf("lock:order:transition:%s", orderID)
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}

func (s *Store) ReleaseTransitionLock(ctx context.Context, orderID string) error {
	if s.client == nil {
		return nil
	}
	key := fmt.Sprintf("lock:order:transition:%s", orderID)
	return s.client.Del(ctx, key).Err()
}

// Workflow coordination
func (s *Store) StoreWorkflowState(ctx context.Context, workflowID string, state []byte, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}
	key := fmt.Sprintf("workflow:%s", workflowID)
	return s.client.Set(ctx, key, state, ttl).Err()
}

func (s *Store) GetWorkflowState(ctx context.Context, workflowID string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("redis not available")
	}
	key := fmt.Sprintf("workflow:%s", workflowID)
	return s.client.Get(ctx, key).Bytes()
}

// Rate limiting helper
func (s *Store) IncrementCounter(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if s.client == nil {
		return 0, nil
	}
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
