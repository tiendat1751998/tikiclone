package unit

import (
	"context"
	"math"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/ratelimit"
)

func TestRateLimitRuleCreate(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rule := &ratelimit.RateLimitRule{
		KeyPattern:    "api:*",
		MaxRequests:   100,
		WindowSeconds: 60,
		Strategy:      ratelimit.StrategyAPI,
	}
	err := rl.CreateRule(context.Background(), rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRateLimitCheckAllowed(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(context.Background(), &ratelimit.RateLimitRule{
		KeyPattern:    "user:*",
		MaxRequests:   10,
		WindowSeconds: 60,
		Strategy:      ratelimit.StrategyUser,
	})

	resp, err := rl.Check(context.Background(), "user:123", string(ratelimit.StrategyUser))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected request to be allowed")
	}
	if resp.Remaining != 10 {
		t.Errorf("expected 10 remaining, got %d", resp.Remaining)
	}
}

func TestRateLimitRecordAndCheck(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(context.Background(), &ratelimit.RateLimitRule{
		KeyPattern:    "ip:*",
		MaxRequests:   3,
		WindowSeconds: 60,
		Strategy:      ratelimit.StrategyIP,
	})

	key := "ip:10.0.0.1"
	strategy := string(ratelimit.StrategyIP)

	for i := 0; i < 3; i++ {
		resp, _ := rl.Check(context.Background(), key, strategy)
		if !resp.Allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
		rl.Record(context.Background(), key, strategy)
	}

	resp, _ := rl.Check(context.Background(), key, strategy)
	if resp.Allowed {
		t.Error("expected request to be rate limited")
	}
	if resp.Remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", resp.Remaining)
	}
}

func TestRateLimitGetRemaining(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(context.Background(), &ratelimit.RateLimitRule{
		KeyPattern:    "api:*",
		MaxRequests:   50,
		WindowSeconds: 60,
		Strategy:      ratelimit.StrategyAPI,
	})

	remaining, _ := rl.GetRemaining(context.Background(), "api:test", string(ratelimit.StrategyAPI))
	if remaining != 50 {
		t.Errorf("expected 50 remaining, got %d", remaining)
	}
}

func TestRateLimitNoRule(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	resp, err := rl.Check(context.Background(), "unknown:key", string(ratelimit.StrategyAPI))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected request to be allowed when no rule matches")
	}
	if resp.Remaining != math.MaxInt32 {
		t.Errorf("expected MaxInt32 remaining, got %d", resp.Remaining)
	}
}

func TestRateLimitReset(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(context.Background(), &ratelimit.RateLimitRule{
		KeyPattern:    "api:test",
		MaxRequests:   1,
		WindowSeconds: 60,
		Strategy:      ratelimit.StrategyAPI,
	})

	rl.Record(context.Background(), "api:test", string(ratelimit.StrategyAPI))
	rl.Reset(context.Background(), "api:test")

	remaining, _ := rl.GetRemaining(context.Background(), "api:test", string(ratelimit.StrategyAPI))
	if remaining != 1 {
		t.Errorf("expected 1 remaining after reset, got %d", remaining)
	}
}

func TestRateLimitValidateRule(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	err := rl.CreateRule(context.Background(), &ratelimit.RateLimitRule{KeyPattern: "", MaxRequests: 0, WindowSeconds: 0})
	if err == nil {
		t.Error("expected validation error")
	}
}
