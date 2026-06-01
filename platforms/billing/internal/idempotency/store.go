package idempotency

import (
	"context"
	"fmt"
	"time"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/billing/internal/infrastructure/redis"
	"go.uber.org/zap"
)

type Store struct {
	redis *redis.Store
}

func NewStore(redis *redis.Store) *Store {
	return &Store{redis: redis}
}

func (s *Store) CheckAndLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return true, nil
	}
	processed, err := s.redis.IsProcessed(ctx, key)
	if err != nil {
		return false, fmt.Errorf("idempotency check: %w", err)
	}
	if processed {
		return false, nil
	}
	locked, err := s.redis.AcquireIdempotencyLock(ctx, key, ttl)
	if err != nil {
		return false, fmt.Errorf("idempotency lock: %w", err)
	}
	return locked, nil
}

func (s *Store) MarkDone(ctx context.Context, key string) {
	if key == "" {
		return
	}
	if err := s.redis.MarkProcessed(ctx, key, 24*time.Hour); err != nil {
		observability.LogWithTrace(ctx).Error("idempotency mark failed", zap.Error(err))
	}
}
