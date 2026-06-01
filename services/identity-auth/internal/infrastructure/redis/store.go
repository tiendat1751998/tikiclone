package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/identity-auth/internal/config"
)

type Store struct {
	client *redis.Client
	cfg    config.RedisConfig
}

func NewStore(cfg config.RedisConfig) *Store {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Store{
		client: client,
		cfg:    cfg,
	}
}

func (s *Store) Close() error {
	return s.client.Close()
}

func (s *Store) CheckLoginRate(ctx context.Context, email string, maxAttempts int, window time.Duration) (bool, error) {
	key := "ratelimit:login:" + email
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	var attempts int
	fmt.Sscanf(val, "%d", &attempts)
	return attempts < maxAttempts, nil
}

func (s *Store) RecordLoginAttempt(ctx context.Context, email string, window time.Duration) error {
	key := "ratelimit:login:" + email
	pipe := s.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) ResetLoginRate(ctx context.Context, email string) error {
	key := "ratelimit:login:" + email
	return s.client.Del(ctx, key).Err()
}

func (s *Store) CheckIPRate(ctx context.Context, ip string, maxAttempts int, window time.Duration) (bool, error) {
	key := "ratelimit:ip:" + ip
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	var attempts int
	fmt.Sscanf(val, "%d", &attempts)
	return attempts < maxAttempts, nil
}

func (s *Store) RecordIPAttempt(ctx context.Context, ip string, window time.Duration) error {
	key := "ratelimit:ip:" + ip
	pipe := s.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	return err
}