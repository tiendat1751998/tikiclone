package tests

import (
	"testing"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/ratelimit"
)

func TestRateLimitCreateRule(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rule := &ratelimit.RateLimitRule{
		Key:           "user:123",
		MaxRequests:   10,
		WindowSeconds: 60,
	}
	err := rl.CreateRule(rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRateLimitCheckAllowed(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(&ratelimit.RateLimitRule{
		Key: "user:123", MaxRequests: 10, WindowSeconds: 60,
	})

	resp, err := rl.Check("user:123")
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

func TestRateLimitSlidingWindow(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(&ratelimit.RateLimitRule{
		Key: "ip:10.0.0.1", MaxRequests: 3, WindowSeconds: 60,
	})

	key := "ip:10.0.0.1"

	for i := 0; i < 3; i++ {
		resp, _ := rl.Check(key)
		if !resp.Allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
		rl.Record(key)
	}

	resp, _ := rl.Check(key)
	if resp.Allowed {
		t.Error("expected request to be rate limited")
	}
	if resp.Remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", resp.Remaining)
	}
}

func TestRateLimitNoRule(t *testing.T) {
	rl := ratelimit.NewRateLimiter(ratelimit.NewInMemoryRepository())

	resp, err := rl.Check("unknown:key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected request to be allowed when no rule matches")
	}
}

func TestRateLimitReset(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(&ratelimit.RateLimitRule{
		Key: "user:test", MaxRequests: 1, WindowSeconds: 60,
	})

	rl.Record("user:test")
	rl.Reset("user:test")

	remaining, _ := rl.GetRemaining("user:test")
	if remaining != 1 {
		t.Errorf("expected 1 remaining after reset, got %d", remaining)
	}
}

func TestRateLimitGetRemaining(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(&ratelimit.RateLimitRule{
		Key: "api:test", MaxRequests: 50, WindowSeconds: 60,
	})

	remaining, _ := rl.GetRemaining("api:test")
	if remaining != 50 {
		t.Errorf("expected 50 remaining, got %d", remaining)
	}
}

func TestRateLimitValidateRule(t *testing.T) {
	rl := ratelimit.NewRateLimiter(ratelimit.NewInMemoryRepository())

	err := rl.CreateRule(&ratelimit.RateLimitRule{Key: "", MaxRequests: 0, WindowSeconds: 0})
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestRateLimitAllow(t *testing.T) {
	repo := ratelimit.NewInMemoryRepository()
	rl := ratelimit.NewRateLimiter(repo)

	rl.CreateRule(&ratelimit.RateLimitRule{
		Key: "test:key", MaxRequests: 1, WindowSeconds: 60,
	})

	if !rl.Allow("test:key") {
		t.Error("first request should be allowed")
	}
	if rl.Allow("test:key") {
		t.Error("second request should be denied")
	}
}
