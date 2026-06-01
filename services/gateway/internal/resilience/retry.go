package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	JitterFactor    float64
	RetryableErrors func(error) bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		JitterFactor:    0.1,
		RetryableErrors: DefaultRetryableCheck,
	}
}

func DefaultRetryableCheck(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	retryablePrefixes := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"temporary",
		"too many",
		"service unavailable",
		"circuit breaker",
		"upstream service error",
		"i/o timeout",
		"eof",
		"dial tcp",
	}
	for _, prefix := range retryablePrefixes {
		if strings.Contains(errStr, prefix) {
			return true
		}
	}
	return false
}

func DoWithRetry(ctx context.Context, config RetryConfig, name string, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("%s cancelled after %d attempts: %w", name, attempt, ctx.Err())
			case <-time.After(calculateBackoff(attempt, config)):
			}
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		if !config.RetryableErrors(err) {
			return fmt.Errorf("%s non-retryable error: %w", name, err)
		}
	}

	return fmt.Errorf("%s exhausted %d attempts, last error: %w", name, config.MaxAttempts, lastErr)
}

func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.InitialInterval) * math.Pow(config.Multiplier, float64(attempt-1))
	if delay > float64(config.MaxInterval) {
		delay = float64(config.MaxInterval)
	}

	jitter := delay * config.JitterFactor * (rand.Float64()*2 - 1)
	return time.Duration(delay + jitter)
}
