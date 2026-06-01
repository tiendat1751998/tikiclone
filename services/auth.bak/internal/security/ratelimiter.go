package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
)

type RateLimiter struct {
	rdb *redis.Client
	cfg config.RateLimitConfig
}

func NewRateLimiter(rdb *redis.Client, cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{rdb: rdb, cfg: cfg}
}

func (l *RateLimiter) CheckLogin(ctx context.Context, email, ip string) error {
	if l.rdb == nil {
		return nil
	}

	attempts, err := l.getAttempts(ctx, fmt.Sprintf("login:%s", email), l.cfg.LoginWindow)
	if err == nil && attempts >= l.cfg.LoginMaxAttempts {
		return domain.ErrRateLimited
	}

	ipAttempts, err := l.getAttempts(ctx, fmt.Sprintf("login_ip:%s", ip), l.cfg.LoginWindow)
	if err == nil && ipAttempts >= l.cfg.LoginMaxAttempts*2 {
		return domain.ErrRateLimited
	}

	if err := l.increment(ctx, fmt.Sprintf("login:%s", email), l.cfg.LoginWindow); err != nil {
		return nil
	}
	if err := l.increment(ctx, fmt.Sprintf("login_ip:%s", ip), l.cfg.LoginWindow); err != nil {
		return nil
	}

	return nil
}

func (l *RateLimiter) CheckRegister(ctx context.Context, ip string) error {
	if l.rdb == nil {
		return nil
	}

	key := fmt.Sprintf("register_ip:%s", ip)
	attempts, err := l.getAttempts(ctx, key, l.cfg.RegisterWindow)
	if err == nil && attempts >= l.cfg.RegisterMaxPerIP {
		return domain.ErrRateLimited
	}

	return l.increment(ctx, key, l.cfg.RegisterWindow)
}

func (l *RateLimiter) CheckPasswordReset(ctx context.Context, email string) error {
	if l.rdb == nil {
		return nil
	}

	key := fmt.Sprintf("password_reset:%s", email)
	attempts, err := l.getAttempts(ctx, key, l.cfg.PasswordResetWindow)
	if err == nil && attempts >= l.cfg.PasswordResetMax {
		return domain.ErrRateLimited
	}

	return l.increment(ctx, key, l.cfg.PasswordResetWindow)
}

func (l *RateLimiter) getAttempts(ctx context.Context, key string, window time.Duration) (int, error) {
	val, err := l.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (l *RateLimiter) increment(ctx context.Context, key string, window time.Duration) error {
	pipe := l.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	return err
}

func (l *RateLimiter) ResetLoginAttempts(ctx context.Context, email string) error {
	if l.rdb == nil {
		return nil
	}
	return l.rdb.Del(ctx, fmt.Sprintf("login:%s", email)).Err()
}
