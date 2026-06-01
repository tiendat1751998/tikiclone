package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/kafka"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, spuID string) (*domain.Product, error)
	List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductList, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, spuID string) error
	GetSKU(ctx context.Context, skuID string) (*domain.SKU, error)
	BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]*domain.SKU, error)
}

type ProductCache interface {
	Get(ctx context.Context, spuID string) (*domain.Product, error)
	Set(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, spuID string) error
	GetOrFetch(ctx context.Context, spuID string, fetchFn func() (*domain.Product, error)) (*domain.Product, error)
	GetList(ctx context.Context, cacheKey string) (*domain.ProductList, error)
	SetList(ctx context.Context, cacheKey string, list *domain.ProductList) error
}

type ProductUseCase struct {
	repo    ProductRepository
	cache   ProductCache
	producer *kafka.Producer
}

func NewProductUseCase(repo ProductRepository, cache ProductCache, producer *kafka.Producer) *ProductUseCase {
	return &ProductUseCase{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

func (uc *ProductUseCase) Create(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.create")
	defer span.End()

	if product.Title == "" || product.CategoryID == "" || product.SellerID == "" {
		return nil, errors.NewValidation("title, category_id, and seller_id are required")
	}

	if len(product.SKUs) == 0 {
		return nil, errors.NewValidation("at least one SKU is required")
	}

	for i, sku := range product.SKUs {
		if sku.Price <= 0 {
			return nil, errors.NewValidation("SKU price must be greater than 0").
				WithDetail("skus", "price must be > 0")
		}
		if sku.Stock < 0 {
			product.SKUs[i].Stock = 0
		}
	}

	if err := uc.repo.Create(ctx, product); err != nil {
		observability.BusinessErrorsTotal.WithLabelValues("catalog-product", "CREATE_PRODUCT_FAILED").Inc()
		return nil, errors.NewInternalError(err)
	}

	if err := uc.publishProductEvent(ctx, "product.created", product); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to publish product created event",
			zap.String("spu_id", product.SPUID),
			zap.Error(err),
		)
	}

	return product, nil
}

func (uc *ProductUseCase) GetByID(ctx context.Context, spuID string) (*domain.Product, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.get_by_id")
	defer span.End()

	product, err := uc.cache.GetOrFetch(ctx, spuID, func() (*domain.Product, error) {
		return uc.repo.GetByID(ctx, spuID)
	})
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if product == nil {
		return nil, domain.ErrProductNotFound
	}

	return product, nil
}

func (uc *ProductUseCase) List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductList, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.list")
	defer span.End()

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 || filter.Size > 100 {
		filter.Size = 20
	}

	cacheKey := fmt.Sprintf("products:list:page=%d:size=%d:cat=%s:sort=%s:order=%s:search=%s",
		filter.Page, filter.Size, filter.CategoryID, filter.SortBy, filter.SortOrder, filter.Search)

	list, err := uc.cache.GetList(ctx, cacheKey)
	if err != nil {
		observability.LogWithTrace(ctx).Error("cache list get failed", zap.Error(err))
	}
	if list != nil {
		return list, nil
	}

	products, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if err := uc.cache.SetList(ctx, cacheKey, products); err != nil {
		observability.LogWithTrace(ctx).Error("cache list set failed", zap.Error(err))
	}

	return products, nil
}

func (uc *ProductUseCase) Update(ctx context.Context, product *domain.Product) error {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.update")
	defer span.End()

	existing, err := uc.repo.GetByID(ctx, product.SPUID)
	if err != nil {
		return errors.NewInternalError(err)
	}
	if existing == nil {
		return domain.ErrProductNotFound
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return errors.NewInternalError(err)
	}

	if err := uc.cache.Delete(ctx, product.SPUID); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to invalidate cache",
			zap.String("spu_id", product.SPUID),
			zap.Error(err),
		)
	}

	if err := uc.publishProductEvent(ctx, "product.updated", product); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to publish product updated event",
			zap.String("spu_id", product.SPUID),
			zap.Error(err),
		)
	}

	return nil
}

func (uc *ProductUseCase) Delete(ctx context.Context, spuID string) error {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.delete")
	defer span.End()

	existing, err := uc.repo.GetByID(ctx, spuID)
	if err != nil {
		return errors.NewInternalError(err)
	}
	if existing == nil {
		return domain.ErrProductNotFound
	}

	if err := uc.repo.Delete(ctx, spuID); err != nil {
		return errors.NewInternalError(err)
	}

	if err := uc.cache.Delete(ctx, spuID); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to invalidate cache",
			zap.String("spu_id", spuID),
			zap.Error(err),
		)
	}

	if err := uc.publishProductEvent(ctx, "product.deleted", existing); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to publish product deleted event",
			zap.String("spu_id", spuID),
			zap.Error(err),
		)
	}

	return nil
}

func (uc *ProductUseCase) GetSKU(ctx context.Context, skuID string) (*domain.SKU, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.get_sku")
	defer span.End()

	sku, err := uc.repo.GetSKU(ctx, skuID)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if sku == nil {
		return nil, domain.ErrSKUNotFound
	}

	return sku, nil
}

func (uc *ProductUseCase) BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]*domain.SKU, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.product.batch_get_skus")
	defer span.End()

	skus, err := uc.repo.BatchGetSKUs(ctx, skuIDs)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	return skus, nil
}

func (uc *ProductUseCase) publishProductEvent(ctx context.Context, eventType string, product *domain.Product) error {
	payload, err := json.Marshal(map[string]interface{}{
		"event_type": eventType,
		"spu_id":     product.SPUID,
		"product":    product,
	})
	if err != nil {
		return err
	}

	return uc.producer.Publish(ctx, kafka.Message{
		Key:   product.SPUID,
		Topic: "search.sync",
		Value: payload,
	})
}
