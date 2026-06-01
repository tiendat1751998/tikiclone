package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product-catalog/internal/domain"
	"github.com/tikiclone/tiki/services/product-catalog/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/product-catalog/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type CatalogService struct {
	productRepo  domain.ProductRepository
	skuRepo      domain.SKURepository
	categoryRepo domain.CategoryRepository
	attrRepo     domain.AttributeRepository
	mediaRepo    domain.ProductMediaRepository
	redis        *redis.Store
	productTTL   time.Duration
	categoryTTL  time.Duration
	publisher    EventPublisher
}

type EventPublisher interface {
	Publish(ctx context.Context, event *domain.CatalogEvent) error
}

func NewCatalogService(pr domain.ProductRepository, sr domain.SKURepository, cr domain.CategoryRepository, ar domain.AttributeRepository, mr domain.ProductMediaRepository, rs *redis.Store, pt, ct time.Duration, pub EventPublisher) *CatalogService {
	return &CatalogService{productRepo: pr, skuRepo: sr, categoryRepo: cr, attrRepo: ar, mediaRepo: mr, redis: rs, productTTL: pt, categoryTTL: ct, publisher: pub}
}

func (s *CatalogService) CreateProduct(ctx context.Context, shopID, name, description, categoryID, currency, idempotencyKey string) (*domain.Product, error) {
	ctx, span := otel.Tracer("tiki-catalog").Start(ctx, "catalog.create_product")
	defer span.End()

	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if categoryID == "" {
		return nil, fmt.Errorf("category_id is required")
	}

	if idempotencyKey != "" {
		exists, err := s.redis.CheckIdempotency(ctx, "create_product:"+idempotencyKey, 24*time.Hour)
		if err != nil {
			observability.LogWithTrace(ctx).Warn("idempotency check failed", zap.Error(err))
		} else if !exists {
			metrics.IdempotentRequests.Inc()
			return nil, domain.ErrDuplicateRequest
		}
	}
	product := domain.NewProduct(shopID, name, description, categoryID, currency)
	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	s.redis.DeleteProduct(ctx, product.ID)
	metrics.ProductsCreated.Inc()
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, &domain.CatalogEvent{EventType: domain.EventProductCreated, AggregateType: "product", AggregateID: product.ID, Payload: product, CreatedAt: time.Now()}); pubErr != nil {
			observability.LogWithTrace(ctx).Error("failed to publish product created event",
				zap.String("product_id", product.ID), zap.Error(pubErr))
		}
	}
	return product, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	ctx, span := otel.Tracer("tiki-catalog").Start(ctx, "catalog.get_product")
	defer span.End()

	if data, err := s.redis.GetProduct(ctx, id); err == nil && len(data) > 0 {
		metrics.CacheHitsTotal.WithLabelValues("catalog", "redis").Inc()
		var cached domain.Product
		if err := json.Unmarshal(data, &cached); err == nil {
			return &cached, nil
		}
		observability.LogWithTrace(ctx).Warn("failed to unmarshal cached product",
			zap.String("product_id", id), zap.Error(err))
	}
	metrics.CacheMissesTotal.WithLabelValues("catalog", "redis").Inc()
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrProductNotFound
	}
	data, _ := json.Marshal(p)
	s.redis.SetProduct(ctx, id, data, s.productTTL)
	return p, nil
}

func (s *CatalogService) UpdateProduct(ctx context.Context, id, name, description, categoryID string) error {
	if id == "" {
		return fmt.Errorf("product id is required")
	}
	if name == "" && description == "" && categoryID == "" {
		return fmt.Errorf("at least one field (name, description, category_id) must be provided for update")
	}

	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrProductNotFound
	}
	p.Update(name, description)
	if categoryID != "" {
		p.CategoryID = categoryID
	}
	if err := s.productRepo.Update(ctx, p); err != nil {
		return err
	}
	s.redis.DeleteProduct(ctx, id)
	metrics.ProductsUpdated.Inc()
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, &domain.CatalogEvent{EventType: domain.EventProductUpdated, AggregateType: "product", AggregateID: id, CreatedAt: time.Now()}); pubErr != nil {
			observability.LogWithTrace(ctx).Error("failed to publish product updated event",
				zap.String("product_id", id), zap.Error(pubErr))
		}
	}
	return nil
}

func (s *CatalogService) ArchiveProduct(ctx context.Context, id string) error {
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrProductNotFound
	}
	if err := p.Archive(); err != nil {
		return err
	}
	if err := s.productRepo.Update(ctx, p); err != nil {
		return err
	}
	s.redis.DeleteProduct(ctx, id)
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, &domain.CatalogEvent{EventType: domain.EventProductArchived, AggregateType: "product", AggregateID: id, CreatedAt: time.Now()}); pubErr != nil {
			observability.LogWithTrace(ctx).Error("failed to publish product archived event",
				zap.String("product_id", id), zap.Error(pubErr))
		}
	}
	return nil
}

func (s *CatalogService) DeleteProduct(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("product id is required")
	}
	p, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrProductNotFound
	}
	if err := p.Archive(); err != nil {
		return err
	}
	if err := s.productRepo.Update(ctx, p); err != nil {
		return err
	}
	s.redis.DeleteProduct(ctx, id)
	metrics.ProductsDeleted.Inc()
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, &domain.CatalogEvent{EventType: domain.EventProductArchived, AggregateType: "product", AggregateID: id, CreatedAt: time.Now()}); pubErr != nil {
			observability.LogWithTrace(ctx).Error("failed to publish product deleted event",
				zap.String("product_id", id), zap.Error(pubErr))
		}
	}
	return nil
}

func (s *CatalogService) GetCategoryTree(ctx context.Context) ([]*domain.Category, error) {
	if data, err := s.redis.GetCategoryTree(ctx); err == nil && len(data) > 0 {
		var cats []*domain.Category
		if err := json.Unmarshal(data, &cats); err == nil {
			return cats, nil
		}
		observability.LogWithTrace(ctx).Warn("failed to unmarshal cached category tree", zap.Error(err))
	}
	cats, err := s.categoryRepo.GetTree(ctx)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(cats)
	s.redis.SetCategoryTree(ctx, data, s.categoryTTL)
	return cats, nil
}

func (s *CatalogService) CreateCategory(ctx context.Context, parentID, name, slug string, level, sortOrder int) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	if slug == "" {
		return nil, fmt.Errorf("category slug is required")
	}
	var parentPtr *string
	if parentID != "" {
		parentPtr = &parentID
	}
	c := domain.NewCategory(name, slug, "", parentPtr, level)
	c.SortOrder = sortOrder
	if err := s.categoryRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	s.redis.InvalidateCategory(ctx, c.ID)
	return c, nil
}

func (s *CatalogService) AddSKU(ctx context.Context, productID, skuCode, name, currency string, price int64) (*domain.SKU, error) {
	if productID == "" {
		return nil, fmt.Errorf("product_id is required")
	}
	if skuCode == "" {
		return nil, fmt.Errorf("sku_code is required")
	}
	sku := domain.NewSKU(productID, skuCode, name, currency, price)
	if err := s.skuRepo.Create(ctx, sku); err != nil {
		return nil, err
	}
	s.redis.DeleteProduct(ctx, productID)
	metrics.SKUsCreated.Inc()
	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, &domain.CatalogEvent{EventType: domain.EventSKUUpdated, AggregateType: "sku", AggregateID: sku.ID, CreatedAt: time.Now()}); pubErr != nil {
			observability.LogWithTrace(ctx).Error("failed to publish SKU updated event",
				zap.String("sku_id", sku.ID), zap.Error(pubErr))
		}
	}
	return sku, nil
}

func (s *CatalogService) ListProductsByShop(ctx context.Context, shopID string, offset, limit int) ([]*domain.Product, int64, error) {
	if shopID == "" {
		return nil, 0, fmt.Errorf("shop_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.productRepo.FindByShopID(ctx, shopID, offset, limit)
}
