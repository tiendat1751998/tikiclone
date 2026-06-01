package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type ProductCache struct {
	client      *redis.Client
	sf          singleflight.Group
	defaultTTL  time.Duration
	listTTL     time.Duration
}

func NewProductCache(client *redis.Client) *ProductCache {
	if client == nil {
		return nil
	}
	return &ProductCache{
		client:     client,
		defaultTTL: 1 * time.Minute,
		listTTL:    1 * time.Minute,
	}
}

func (c *ProductCache) Get(ctx context.Context, spuID string) (*domain.Product, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}

	ctx, span := otel.Tracer("catalog-product").Start(ctx, "cache.product.get")
	defer span.End()

	cacheKey := "product:" + spuID

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		observability.CacheMissesTotal.WithLabelValues("catalog-product", "redis").Inc()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	observability.CacheHitsTotal.WithLabelValues("catalog-product", "redis").Inc()

	var product domain.Product
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (c *ProductCache) Set(ctx context.Context, product *domain.Product) error {
	if c == nil || c.client == nil {
		return nil
	}

	cacheKey := "product:" + product.SPUID

	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, cacheKey, data, c.defaultTTL).Err()
}

func (c *ProductCache) Delete(ctx context.Context, spuID string) error {
	if c == nil || c.client == nil {
		return nil
	}

	return c.client.Del(ctx, "product:"+spuID).Err()
}

func (c *ProductCache) GetOrFetch(ctx context.Context, spuID string, fetchFn func() (*domain.Product, error)) (*domain.Product, error) {
	product, err := c.Get(ctx, spuID)
	if err != nil {
		observability.LogWithTrace(ctx).Error("cache get failed", zap.Error(err))
	}
	if product != nil {
		return product, nil
	}

	result, err, _ := c.sf.Do(spuID, func() (interface{}, error) {
		product, err := fetchFn()
		if err != nil {
			return nil, err
		}
		if product != nil {
			if err := c.Set(ctx, product); err != nil {
				observability.LogWithTrace(ctx).Error("cache set failed", zap.Error(err))
			}
		}
		return product, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*domain.Product), nil
}

func (c *ProductCache) GetList(ctx context.Context, cacheKey string) (*domain.ProductList, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}

	ctx, span := otel.Tracer("catalog-product").Start(ctx, "cache.product.list.get")
	defer span.End()

	val, err := c.client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		observability.CacheMissesTotal.WithLabelValues("catalog-product", "redis").Inc()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	observability.CacheHitsTotal.WithLabelValues("catalog-product", "redis").Inc()

	var productList domain.ProductList
	if err := json.Unmarshal([]byte(val), &productList); err != nil {
		return nil, err
	}

	return &productList, nil
}

func (c *ProductCache) SetList(ctx context.Context, cacheKey string, list *domain.ProductList) error {
	if c == nil || c.client == nil {
		return nil
	}

	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, cacheKey, data, c.listTTL).Err()
}

type CategoryCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

func NewCategoryCache(client *redis.Client) *CategoryCache {
	if client == nil {
		return nil
	}
	return &CategoryCache{
		client:     client,
		defaultTTL: 10 * time.Minute,
	}
}

func (c *CategoryCache) GetAll(ctx context.Context) ([]domain.Category, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}

	ctx, span := otel.Tracer("catalog-product").Start(ctx, "cache.category.get_all")
	defer span.End()

	val, err := c.client.Get(ctx, "categories:all").Result()
	if err == redis.Nil {
		observability.CacheMissesTotal.WithLabelValues("catalog-product", "redis").Inc()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	observability.CacheHitsTotal.WithLabelValues("catalog-product", "redis").Inc()

	var categories []domain.Category
	if err := json.Unmarshal([]byte(val), &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (c *CategoryCache) SetAll(ctx context.Context, categories []domain.Category) error {
	if c == nil || c.client == nil {
		return nil
	}

	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, "categories:all", data, c.defaultTTL).Err()
}

func (c *CategoryCache) DeleteAll(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}

	return c.client.Del(ctx, "categories:all").Err()
}
