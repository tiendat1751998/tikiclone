package tests

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/edgecache"
)

func TestCacheSetAndGet(t *testing.T) {
	c := edgecache.NewCache()

	err := c.Set("key1", "value1", 300)
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	entry, err := c.Get("key1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if entry == nil {
		t.Fatal("expected cache hit")
	}
	if entry.Value != "value1" {
		t.Errorf("expected value1, got %s", entry.Value)
	}
}

func TestCacheMiss(t *testing.T) {
	c := edgecache.NewCache()

	entry, err := c.Get("nonexistent")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if entry != nil {
		t.Fatal("expected nil for miss")
	}
}

func TestCacheTTL(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("ttl-key", "value", 1)

	entry, _ := c.Get("ttl-key")
	if entry == nil {
		t.Fatal("expected cache hit before expiry")
	}

	time.Sleep(1100 * time.Millisecond)

	entry, _ = c.Get("ttl-key")
	if entry != nil {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCacheDelete(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("delete-key", "value", 300)
	c.Delete("delete-key")

	entry, _ := c.Get("delete-key")
	if entry != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestCachePurgeByPattern(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("user:1", "alice", 300)
	c.Set("user:2", "bob", 300)
	c.Set("product:1", "item", 300)

	count, err := c.PurgeByPattern("user:*")
	if err != nil {
		t.Fatalf("purge failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 purged, got %d", count)
	}

	entry, _ := c.Get("user:1")
	if entry != nil {
		t.Error("user:1 should be purged")
	}
	entry, _ = c.Get("product:1")
	if entry == nil {
		t.Error("product:1 should still exist")
	}
}

func TestCacheStatsHitMiss(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("hit-key", "value", 300)
	c.Get("hit-key")
	c.Get("hit-key")
	c.Get("miss-key")

	stats := c.GetStats()
	if stats.HitCount != 2 {
		t.Errorf("expected 2 hits, got %d", stats.HitCount)
	}
	if stats.MissCount != 1 {
		t.Errorf("expected 1 miss, got %d", stats.MissCount)
	}
	if stats.Ratio != 2.0/3.0 {
		t.Errorf("expected ratio 0.666..., got %f", stats.Ratio)
	}
	if stats.Entries != 1 {
		t.Errorf("expected 1 entry, got %d", stats.Entries)
	}
}

func TestCacheHitCount(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("pop-key", "value", 300)

	for i := 0; i < 5; i++ {
		c.Get("pop-key")
	}

	entry, _ := c.Get("pop-key")
	if entry.HitCount != 6 {
		t.Errorf("expected hit count 6, got %d", entry.HitCount)
	}
}

func TestCacheSetEmptyKey(t *testing.T) {
	c := edgecache.NewCache()

	err := c.Set("", "value", 300)
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestCacheStatsTotalRequests(t *testing.T) {
	c := edgecache.NewCache()

	c.Set("a", "1", 300)
	c.Get("a")
	c.Get("a")
	c.Get("b")
	c.Get("c")

	stats := c.GetStats()
	total := stats.HitCount + stats.MissCount
	if total != 4 {
		t.Errorf("expected 4 total requests, got %d", total)
	}
}

func TestCacheBuildKey(t *testing.T) {
	c := edgecache.NewCache()

	policy := &edgecache.CachePolicy{
		PathPattern: "/api/*",
		TTLSeconds:  300,
		VaryBy: edgecache.VaryBy{
			Headers:     []string{"Authorization"},
			QueryParams: []string{"locale"},
		},
	}

	key := c.BuildCacheKey("/api/users", map[string]string{"authorization": "bearer xyz"}, map[string]string{"locale": "en"}, policy)

	if key != "/api/users|h:Authorization=bearer xyz|q:locale=en" {
		t.Errorf("unexpected cache key: %s", key)
	}
}

func TestCacheBuildKeyNoPolicy(t *testing.T) {
	c := edgecache.NewCache()

	key := c.BuildCacheKey("/api/users", nil, nil, nil)
	if key != "/api/users" {
		t.Errorf("expected just path, got %s", key)
	}
}

func TestCacheAddPolicy(t *testing.T) {
	c := edgecache.NewCache()

	policy := &edgecache.CachePolicy{
		PathPattern: "/api/*",
		TTLSeconds:  60,
	}

	err := c.AddPolicy(policy)
	if err != nil {
		t.Fatalf("add policy failed: %v", err)
	}

	p, err := c.GetPolicy("/api/*")
	if err != nil {
		t.Fatalf("get policy failed: %v", err)
	}
	if p == nil {
		t.Fatal("expected policy")
	}
	if p.TTLSeconds != 60 {
		t.Errorf("expected 60s, got %d", p.TTLSeconds)
	}
}
