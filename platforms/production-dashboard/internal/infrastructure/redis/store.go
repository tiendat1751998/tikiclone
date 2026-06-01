package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/config"
)

type Store struct {
	rdb *redis.Client
	cfg config.RedisConfig
}

func NewStore(rdb *redis.Client, cfg config.RedisConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

func (s *Store) isAvailable() bool {
	return s.rdb != nil
}

func (s *Store) Ping(ctx context.Context) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Ping(ctx).Err()
}

func (s *Store) Close() error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Close()
}

func (s *Store) GetDashboardSummary(ctx context.Context) ([]byte, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	data, err := s.rdb.Get(ctx, "dashboard:summary").Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

func (s *Store) SetDashboardSummary(ctx context.Context, data []byte, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Set(ctx, "dashboard:summary", data, ttl).Err()
}

func (s *Store) InvalidateSummary(ctx context.Context) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Del(ctx, "dashboard:summary").Err()
}

func (s *Store) CacheIncident(ctx context.Context, incidentID string, data json.RawMessage, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Set(ctx, fmt.Sprintf("incident:%s", incidentID), data, ttl).Err()
}

func (s *Store) GetCachedIncident(ctx context.Context, incidentID string) ([]byte, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	data, err := s.rdb.Get(ctx, fmt.Sprintf("incident:%s", incidentID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

func (s *Store) CacheDeployment(ctx context.Context, deploymentID string, data json.RawMessage, ttl time.Duration) error {
	if !s.isAvailable() {
		return nil
	}
	return s.rdb.Set(ctx, fmt.Sprintf("deployment:%s", deploymentID), data, ttl).Err()
}

func (s *Store) GetCachedDeployment(ctx context.Context, deploymentID string) ([]byte, error) {
	if !s.isAvailable() {
		return nil, nil
	}
	data, err := s.rdb.Get(ctx, fmt.Sprintf("deployment:%s", deploymentID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}
