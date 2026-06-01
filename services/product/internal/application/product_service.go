package application

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var idCounter atomic.Uint64

func generateID(prefix string) string {
	return prefix + "-" + strconv.FormatUint(idCounter.Add(1), 36) + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// ProductRepository defines the interface for product persistence
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySPU(ctx context.Context, spuID string) (*domain.Product, error)
	List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductList, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, filter domain.ProductFilter) (*domain.ProductList, error)
	GetSKU(ctx context.Context, skuID string) (*domain.SKU, error)
	BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]*domain.SKU, error)
	CreateSKU(ctx context.Context, sku *domain.SKU) error
	UpdateSKU(ctx context.Context, sku *domain.SKU) error
	ListSKUsByProduct(ctx context.Context, spuID string) ([]domain.SKU, error)
}

// ProductCache defines the interface for product caching
type ProductCache interface {
	Get(ctx context.Context, key string) (*domain.Product, error)
	Set(ctx context.Context, key string, product *domain.Product, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.Product, error)) (*domain.Product, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

// ProductService handles product use cases
type ProductService struct {
	repo      ProductRepository
	cache     ProductCache
	publisher EventPublisher
}

// NewProductService creates a new ProductService
func NewProductService(repo ProductRepository, cache ProductCache, publisher EventPublisher) *ProductService {
	return &ProductService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

// CreateProduct creates a new product with idempotency support
func (s *ProductService) CreateProduct(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.create")
	defer span.End()

	if req.Title == "" {
		return nil, errors.NewValidation("title is required")
	}
	if req.CategoryID == "" {
		return nil, errors.NewValidation("category_id is required")
	}
	if req.SellerID == "" {
		return nil, errors.NewValidation("seller_id is required")
	}
	if len(req.SKUs) == 0 {
		return nil, errors.NewValidation("at least one SKU is required")
	}

	// [SECURITY] Idempotency check — prevent duplicate creation on retries
	if req.IdempotencyKey != "" {
		exists, err := s.isDuplicateRequest(ctx, req.IdempotencyKey)
		if err != nil {
			observability.LogWithTrace(ctx).Warn("idempotency check failed", zap.Error(err))
		} else if exists {
			return nil, errors.NewConflict("duplicate request: product already created with this idempotency key")
		}
	}

	skus := make([]domain.SKU, len(req.SKUs))
	for i, skuReq := range req.SKUs {
		if skuReq.Price <= 0 {
			return nil, errors.NewValidation(fmt.Sprintf("SKU[%d] price must be greater than 0", i))
		}
		skus[i] = domain.SKU{
			SKUID:     generateID("SKU"),
			Price:     skuReq.Price,
			SalePrice: skuReq.SalePrice,
			Stock:     skuReq.Stock,
			Status:    domain.SKUStatusActive,
		}
	}

	product := &domain.Product{
		SPUID:       generateID("SPU"),
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		SellerID:    req.SellerID,
		Status:      domain.ProductStatusDraft,
		SKUs:        skus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if len(req.Images) > 0 {
		product.Images = make([]domain.ProductImage, len(req.Images))
		for i, img := range req.Images {
			product.Images[i] = domain.ProductImage{
				URL:       img.URL,
				AltText:   img.AltText,
				SortOrder: img.SortOrder,
				IsPrimary: img.IsPrimary,
			}
		}
	}

	if err := s.repo.Create(ctx, product); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create product")
		observability.BusinessErrorsTotal.WithLabelValues("product", "CREATE_FAILED").Inc()
		return nil, errors.NewInternalError(err)
	}

	// Store idempotency key after successful creation
	if req.IdempotencyKey != "" {
		s.storeIdempotencyKey(ctx, req.IdempotencyKey, product.SPUID, 24*time.Hour)
	}

	event := domain.NewProductCreatedEvent(product)
	if payload, err := event.Marshal(); err == nil {
		if pubErr := s.publisher.Publish(ctx, "product.events", product.SPUID, payload); pubErr != nil {
			observability.LogWithTrace(ctx).Warn("failed to publish product created event",
				zap.String("spu_id", product.SPUID), zap.Error(pubErr))
		}
	}

	span.SetAttributes(attribute.String("spu_id", product.SPUID))
	observability.LogWithTrace(ctx).Info("product created",
		zap.String("spu_id", product.SPUID), zap.String("seller_id", product.SellerID))

	return ToProductResponse(product), nil
}

// isDuplicateRequest checks if an idempotency key was already processed
func (s *ProductService) isDuplicateRequest(ctx context.Context, key string) (bool, error) {
	// Use cache (Redis) for fast idempotency check
	cached, err := s.cache.Get(ctx, "idempotency:"+key)
	if err != nil {
		return false, err
	}
	return cached != nil, nil
}

// storeIdempotencyKey stores the idempotency key after successful creation
func (s *ProductService) storeIdempotencyKey(ctx context.Context, key, spuID string, ttl time.Duration) {
	// Store a minimal marker in cache
	if err := s.cache.Set(ctx, "idempotency:"+key, &domain.Product{SPUID: spuID}, ttl); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to store idempotency key",
			zap.String("key", key), zap.Error(err))
	}
}

// GetProduct gets a product by SPU ID (cache-through)
func (s *ProductService) GetProduct(ctx context.Context, spuID string) (*ProductResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.get")
	defer span.End()

	if spuID == "" {
		return nil, errors.NewValidation("spu_id is required")
	}

	product, err := s.cache.GetOrFetch(ctx, "product:"+spuID, 5*time.Minute, func() (*domain.Product, error) {
		return s.repo.GetBySPU(ctx, spuID)
	})
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}
	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return ToProductResponse(product), nil
}

// ListProducts lists products with filtering and pagination
func (s *ProductService) ListProducts(ctx context.Context, filter domain.ProductFilter) (*ProductListResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.list")
	defer span.End()

	filter.Normalize()

	productList, err := s.repo.List(ctx, filter)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	return ToProductListResponse(productList), nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, spuID string, req UpdateProductRequest) (*ProductResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.update")
	defer span.End()

	if spuID == "" {
		return nil, errors.NewValidation("spu_id is required")
	}

	existing, err := s.repo.GetBySPU(ctx, spuID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}
	if existing == nil {
		return nil, domain.ErrProductNotFound
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.CategoryID != "" {
		existing.CategoryID = req.CategoryID
	}
	if req.BrandID != "" {
		existing.BrandID = req.BrandID
	}
	if req.Status != "" {
		existing.Status = domain.ProductStatus(req.Status)
	}
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, existing); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update product")
		return nil, errors.NewInternalError(err)
	}

	if err := s.cache.Delete(ctx, "product:"+spuID); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to invalidate cache",
			zap.String("spu_id", spuID), zap.Error(err))
	}

	event := domain.NewProductUpdatedEvent(existing, nil)
	if payload, err := event.Marshal(); err == nil {
		if pubErr := s.publisher.Publish(ctx, "product.events", spuID, payload); pubErr != nil {
			observability.LogWithTrace(ctx).Warn("failed to publish product updated event",
				zap.String("spu_id", spuID), zap.Error(pubErr))
		}
	}

	span.SetAttributes(attribute.String("spu_id", spuID))
	return ToProductResponse(existing), nil
}

// DeleteProduct soft-deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, spuID string) error {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.delete")
	defer span.End()

	if spuID == "" {
		return errors.NewValidation("spu_id is required")
	}

	existing, err := s.repo.GetBySPU(ctx, spuID)
	if err != nil {
		return errors.NewInternalError(err)
	}
	if existing == nil {
		return domain.ErrProductNotFound
	}

	if err := s.repo.Delete(ctx, spuID); err != nil {
		span.RecordError(err)
		return errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, "product:"+spuID)

	event := domain.NewProductDeletedEvent(existing)
	payload, err := event.Marshal()
	if err != nil {
		observability.LogWithTrace(ctx).Error("failed to marshal delete event", zap.Error(err))
	} else if s.publisher != nil {
		if err := s.publisher.Publish(ctx, "product.events", spuID, payload); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish delete event", zap.Error(err))
		}
	}

	span.SetAttributes(attribute.String("spu_id", spuID))
	return nil
}

// SearchProducts searches products
func (s *ProductService) SearchProducts(ctx context.Context, query string, filter domain.ProductFilter) (*ProductListResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.search")
	defer span.End()

	filter.Normalize()

	productList, err := s.repo.Search(ctx, query, filter)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	return ToProductListResponse(productList), nil
}

// BatchGetSKUs gets multiple SKUs by ID
func (s *ProductService) BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]SKUResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.product.batch_get_skus")
	defer span.End()

	if len(skuIDs) == 0 {
		return nil, nil
	}

	skuMap, err := s.repo.BatchGetSKUs(ctx, skuIDs)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	result := make(map[string]SKUResponse, len(skuMap))
	for id, sku := range skuMap {
		resp := ToSKUResponse(sku)
		result[id] = *resp
	}

	return result, nil
}
