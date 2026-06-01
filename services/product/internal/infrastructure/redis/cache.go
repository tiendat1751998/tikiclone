package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/product/internal/domain"
)

var productPool = sync.Pool{
	New: func() interface{} {
		return &domain.Product{}
	},
}

func getProductFromPool() *domain.Product {
	return productPool.Get().(*domain.Product)
}

func putProductToPool(p *domain.Product) {
	*p = domain.Product{}
	productPool.Put(p)
}

// Cache implements ProductCache and CategoryCache using Redis
type Cache struct {
	client *redis.Client
	prefix string
	sem    chan struct{}
	// [SECURITY] In-memory LRU cache with max size to prevent unbounded growth
	localCache   map[string]*cacheEntry
	localCacheMu sync.RWMutex
	maxLocalSize int
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

const maxConcurrentCacheWrites = 100
const maxLocalCacheSize = 10000

// NewCache creates a new Redis cache
func NewCache(client *redis.Client) *Cache {
	c := &Cache{
		client:       client,
		prefix:       "product:",
		sem:          make(chan struct{}, maxConcurrentCacheWrites),
		localCache:   make(map[string]*cacheEntry),
		maxLocalSize: maxLocalCacheSize,
	}
	// Start cleanup goroutine for expired local cache entries
	go c.cleanupLoop()
	return c
}

// cleanupLoop periodically removes expired entries from local cache
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanupExpired()
	}
}

func (c *Cache) cleanupExpired() {
	c.localCacheMu.Lock()
	defer c.localCacheMu.Unlock()
	now := time.Now()
	for key, entry := range c.localCache {
		if now.After(entry.expiresAt) {
			delete(c.localCache, key)
		}
	}
}

// evictOldest removes oldest entries when cache exceeds max size
func (c *Cache) evictOldest() {
	c.localCacheMu.Lock()
	defer c.localCacheMu.Unlock()
	if len(c.localCache) <= c.maxLocalSize {
		return
	}
	// Evict 20% of entries when over limit
	evictCount := c.maxLocalSize / 5
	count := 0
	for key := range c.localCache {
		delete(c.localCache, key)
		count++
		if count >= evictCount {
			break
		}
	}
}

// Get retrieves a product from cache
func (c *Cache) Get(ctx context.Context, key string) (*domain.Product, error) {
	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var product domain.Product
	if err := sonic.Unmarshal(data, &product); err != nil {
		return nil, fmt.Errorf("unmarshal product: %w", err)
	}
	return &product, nil
}

// Set stores a product in cache
func (c *Cache) Set(ctx context.Context, key string, product *domain.Product, ttl time.Duration) error {
	data, err := sonic.Marshal(product)
	if err != nil {
		return fmt.Errorf("marshal product: %w", err)
	}
	return c.client.Set(ctx, c.prefix+key, data, ttl).Err()
}

// Delete removes a product from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.prefix+key).Err()
}

// GetOrFetch gets from cache or fetches and caches
func (c *Cache) GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.Product, error)) (*domain.Product, error) {
	// Try cache first
	product, err := c.Get(ctx, key)
	if err != nil {
		// Log but continue to fetch
	}
	if product != nil {
		return product, nil
	}

	// Fetch from source
	product, err = fetchFn()
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	// Cache the result (bounded concurrency, propagate caller context)
	select {
	case c.sem <- struct{}{}:
		go func() {
			defer func() { <-c.sem }()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			c.Set(ctx, key, product, ttl)
		}()
	default:
	}

	return product, nil
}

// GetOrFetchCategory gets a category from cache or fetches it
func (c *Cache) GetOrFetchCategory(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.Category, error)) (*domain.Category, error) {
	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}
	if err == nil {
		var category domain.Category
		if err := sonic.Unmarshal(data, &category); err == nil {
			return &category, nil
		}
	}

	category, err := fetchFn()
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, nil
	}

	select {
	case c.sem <- struct{}{}:
		go func() {
			defer func() { <-c.sem }()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			if data, err := sonic.Marshal(category); err == nil {
				c.client.Set(ctx, c.prefix+key, data, ttl)
			}
		}()
	default:
	}

	return category, nil
}

// GetOrFetchTree gets the category tree from cache or fetches it
func (c *Cache) GetOrFetchTree(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.CategoryTree, error)) (*domain.CategoryTree, error) {
	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}
	if err == nil {
		var tree domain.CategoryTree
		if err := sonic.Unmarshal(data, &tree); err == nil {
			return &tree, nil
		}
	}

	tree, err := fetchFn()
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return nil, nil
	}

	select {
	case c.sem <- struct{}{}:
		go func() {
			defer func() { <-c.sem }()
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			if data, err := sonic.Marshal(tree); err == nil {
				c.client.Set(ctx, c.prefix+key, data, ttl)
			}
		}()
	default:
	}

	return tree, nil
}

// DeleteTree invalidates the category tree cache
func (c *Cache) DeleteTree(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, c.prefix+"category:tree*", 100).Iterator()
	batchSize := 100
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return iter.Err()
}

// AsCategoryCache returns a CategoryCache implementation
func (c *Cache) AsCategoryCache() CategoryCache {
	return &CategoryRedisCache{Cache: c}
}

// CategoryRedisCache wraps Cache to implement CategoryCache
type CategoryRedisCache struct {
	*Cache
}

func (c *CategoryRedisCache) Get(ctx context.Context, key string) (*domain.Category, error) {
	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var category domain.Category
	if err := sonic.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("unmarshal category: %w", err)
	}
	return &category, nil
}

func (c *CategoryRedisCache) Set(ctx context.Context, key string, category *domain.Category, ttl time.Duration) error {
	data, err := sonic.Marshal(category)
	if err != nil {
		return fmt.Errorf("marshal category: %w", err)
	}
	return c.client.Set(ctx, c.prefix+key, data, ttl).Err()
}

// AsAttributeCache returns an AttributeCache implementation
func (c *Cache) AsAttributeCache() AttributeCache {
	return &AttributeRedisCache{Cache: c}
}

// AttributeRedisCache wraps Cache to implement AttributeCache
type AttributeRedisCache struct {
	*Cache
}

func (c *AttributeRedisCache) Get(ctx context.Context, key string) (*domain.Attribute, error) {
	data, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var attr domain.Attribute
	if err := sonic.Unmarshal(data, &attr); err != nil {
		return nil, fmt.Errorf("unmarshal attribute: %w", err)
	}
	return &attr, nil
}

func (c *AttributeRedisCache) Set(ctx context.Context, key string, attr *domain.Attribute, ttl time.Duration) error {
	data, err := sonic.Marshal(attr)
	if err != nil {
		return fmt.Errorf("marshal attribute: %w", err)
	}
	return c.client.Set(ctx, c.prefix+key, data, ttl).Err()
}

func (c *AttributeRedisCache) InvalidateByCategory(ctx context.Context, categoryID string) error {
	iter := c.client.Scan(ctx, 0, c.prefix+"attr:category:"+categoryID+":*", 100).Iterator()
	batchSize := 100
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return iter.Err()
}

// Ensure interfaces are satisfied
var _ ProductCache = (*Cache)(nil)
var _ CategoryCache = (*CategoryRedisCache)(nil)
var _ AttributeCache = (*AttributeRedisCache)(nil)

// ProductCache interface
type ProductCache interface {
	Get(ctx context.Context, key string) (*domain.Product, error)
	Set(ctx context.Context, key string, product *domain.Product, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.Product, error)) (*domain.Product, error)
}

// CategoryCache interface
type CategoryCache interface {
	Get(ctx context.Context, key string) (*domain.Category, error)
	Set(ctx context.Context, key string, category *domain.Category, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteTree(ctx context.Context) error
}

// AttributeCache interface
type AttributeCache interface {
	Get(ctx context.Context, key string) (*domain.Attribute, error)
	Set(ctx context.Context, key string, attr *domain.Attribute, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	InvalidateByCategory(ctx context.Context, categoryID string) error
}
