package application

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SKUCache defines the interface for SKU caching.
type SKUCache interface {
	Get(ctx context.Context, skuID string) (*domain.SKU, error)
	Set(ctx context.Context, skuID string, sku *domain.SKU, ttl time.Duration) error
	Delete(ctx context.Context, skuID string) error
	GetOrFetch(ctx context.Context, skuID string, ttl time.Duration, fetchFn func() (*domain.SKU, error)) (*domain.SKU, error)
}

// SKUService handles SKU use cases.
type SKUService struct {
	repo      ProductRepository
	cache     SKUCache
	publisher EventPublisher
	logger    *zap.Logger
	tracer    trace.Tracer
}

// NewSKUService creates a new SKUService.
func NewSKUService(
	repo ProductRepository,
	cache SKUCache,
	publisher EventPublisher,
	logger *zap.Logger,
) *SKUService {
	return &SKUService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		logger:    logger,
		tracer:    otel.Tracer("sku-service"),
	}
}

// GetSKU retrieves a single SKU by its SKU ID (cache-through).
func (s *SKUService) GetSKU(ctx context.Context, skuID string) (*SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.get")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if skuID == "" {
		return nil, errors.NewValidation("sku_id is required")
	}

	span.SetAttributes(attribute.String("sku.id", skuID))

	sku, err := s.cache.GetOrFetch(ctx, "sku:"+skuID, 5*time.Minute, func() (*domain.SKU, error) {
		return s.repo.GetSKU(ctx, skuID)
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get SKU")
		log.Error("failed to get SKU",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
		return nil, errors.NewInternalError(err)
	}

	if sku == nil {
		span.SetStatus(codes.Error, "SKU not found")
		return nil, errors.NewNotFound(fmt.Sprintf("SKU %s not found", skuID))
	}

	return ToSKUResponse(sku), nil
}

// BatchGetSKUs retrieves multiple SKUs by their IDs.
func (s *SKUService) BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.batch_get")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if len(skuIDs) == 0 {
		return nil, nil
	}

	span.SetAttributes(attribute.Int("sku.requested_count", len(skuIDs)))

	skuMap, err := s.repo.BatchGetSKUs(ctx, skuIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to batch get SKUs")
		observability.BusinessErrorsTotal.WithLabelValues("sku", "BATCH_GET_FAILED").Inc()
		log.Error("failed to batch get SKUs",
			zap.Error(err),
			zap.Int("count", len(skuIDs)),
		)
		return nil, errors.NewInternalError(err)
	}

	result := make(map[string]SKUResponse, len(skuMap))
	for id, sku := range skuMap {
		resp := ToSKUResponse(sku)
		result[id] = *resp

		// Populate cache for each SKU
		if err := s.cache.Set(ctx, "sku:"+id, sku, 5*time.Minute); err != nil {
			log.Warn("failed to cache SKU",
				zap.Error(err),
				zap.String("sku_id", id),
			)
		}
	}

	span.SetAttributes(attribute.Int("sku.returned_count", len(result)))

	return result, nil
}

// UpdateSKUPrice updates the price and/or sale price of an SKU.
func (s *SKUService) UpdateSKUPrice(ctx context.Context, skuID string, price float64, salePrice float64) (*SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.update_price")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if skuID == "" {
		return nil, errors.NewValidation("sku_id is required")
	}
	if price < 0 {
		return nil, errors.NewValidation("price must be non-negative")
	}
	if salePrice < 0 {
		return nil, errors.NewValidation("sale_price must be non-negative")
	}
	if salePrice > 0 && salePrice > price {
		return nil, errors.NewValidation("sale_price cannot exceed regular price")
	}

	span.SetAttributes(
		attribute.String("sku.id", skuID),
		attribute.Float64("sku.price", price),
		attribute.Float64("sku.sale_price", salePrice),
	)

	// Fetch existing SKU
	existing, err := s.repo.GetSKU(ctx, skuID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch SKU for price update")
		log.Error("failed to fetch SKU for price update",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
		return nil, errors.NewInternalError(err)
	}
	if existing == nil {
		span.SetStatus(codes.Error, "SKU not found")
		return nil, errors.NewNotFound(fmt.Sprintf("SKU %s not found", skuID))
	}

	// Apply updates
	existing.Price = price
	existing.SalePrice = salePrice
	existing.UpdatedAt = time.Now().UTC()

	// Auto-update status based on stock
	if existing.Stock == 0 {
		existing.Status = domain.SKUStatusOutOfStock
	} else if existing.Status == domain.SKUStatusOutOfStock && existing.Stock > 0 {
		existing.Status = domain.SKUStatusActive
	}

	// Persist
	if err := s.repo.UpdateSKU(ctx, existing); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update SKU price")
		observability.BusinessErrorsTotal.WithLabelValues("sku", "PRICE_UPDATE_FAILED").Inc()
		log.Error("failed to update SKU price",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
		if domain.IsDomainError(err, "CONCURRENT_MODIFICATION") {
			return nil, errors.NewConflict("sku was modified by another request")
		}
		return nil, errors.NewInternalError(err)
	}

	// Invalidate cache
	if err := s.cache.Delete(ctx, "sku:"+skuID); err != nil {
		log.Warn("failed to invalidate SKU cache",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
	}

	// Publish event
	event := domain.NewSKUUpdatedEvent(existing.SPUID, existing.SKUID, price, existing.Stock, string(existing.Status))
	event.SalePrice = salePrice
	if payload, err := sonic.Marshal(event); err == nil {
		if pubErr := s.publisher.Publish(ctx, "sku.events", skuID, payload); pubErr != nil {
			log.Warn("failed to publish SKUUpdatedEvent",
				zap.Error(pubErr),
				zap.String("sku_id", skuID),
			)
		}
	}

	log.Info("SKU price updated",
		zap.String("sku_id", skuID),
		zap.Float64("price", price),
		zap.Float64("sale_price", salePrice),
	)

	return ToSKUResponse(existing), nil
}

// UpdateSKUStock updates the stock quantity of an SKU.
func (s *SKUService) UpdateSKUStock(ctx context.Context, skuID string, stock int32) (*SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.update_stock")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if skuID == "" {
		return nil, errors.NewValidation("sku_id is required")
	}
	if stock < 0 {
		return nil, errors.NewValidation("stock must be non-negative")
	}

	span.SetAttributes(
		attribute.String("sku.id", skuID),
		attribute.Int("sku.stock", int(stock)),
	)

	// Fetch existing SKU
	existing, err := s.repo.GetSKU(ctx, skuID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch SKU for stock update")
		log.Error("failed to fetch SKU for stock update",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
		return nil, errors.NewInternalError(err)
	}
	if existing == nil {
		span.SetStatus(codes.Error, "SKU not found")
		return nil, errors.NewNotFound(fmt.Sprintf("SKU %s not found", skuID))
	}

	// Apply updates
	existing.Stock = stock
	existing.UpdatedAt = time.Now().UTC()

	// Auto-update status based on stock
	if stock == 0 {
		existing.Status = domain.SKUStatusOutOfStock
	} else if existing.Status == domain.SKUStatusOutOfStock {
		existing.Status = domain.SKUStatusActive
	}

	// Persist
	if err := s.repo.UpdateSKU(ctx, existing); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update SKU stock")
		observability.BusinessErrorsTotal.WithLabelValues("sku", "STOCK_UPDATE_FAILED").Inc()
		log.Error("failed to update SKU stock",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
		if domain.IsDomainError(err, "CONCURRENT_MODIFICATION") {
			return nil, errors.NewConflict("sku was modified by another request")
		}
		return nil, errors.NewInternalError(err)
	}

	// Invalidate cache
	if err := s.cache.Delete(ctx, "sku:"+skuID); err != nil {
		log.Warn("failed to invalidate SKU cache",
			zap.Error(err),
			zap.String("sku_id", skuID),
		)
	}

	// Publish event
	event := domain.NewSKUUpdatedEvent(existing.SPUID, existing.SKUID, existing.Price, stock, string(existing.Status))
	if payload, err := sonic.Marshal(event); err == nil {
		if pubErr := s.publisher.Publish(ctx, "sku.events", skuID, payload); pubErr != nil {
			log.Warn("failed to publish SKUUpdatedEvent",
				zap.Error(pubErr),
				zap.String("sku_id", skuID),
			)
		}
	}

	log.Info("SKU stock updated",
		zap.String("sku_id", skuID),
		zap.Int32("stock", stock),
		zap.String("status", string(existing.Status)),
	)

	return ToSKUResponse(existing), nil
}

// CreateSKU creates a new SKU for a product.
func (s *SKUService) CreateSKU(ctx context.Context, spuID string, req CreateSKURequest) (*SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.create")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if spuID == "" {
		return nil, errors.NewValidation("spu_id is required")
	}
	if req.Price < 0 {
		return nil, errors.NewValidation("price must be non-negative")
	}
	if req.Stock < 0 {
		return nil, errors.NewValidation("stock must be non-negative")
	}

	span.SetAttributes(
		attribute.String("spu.id", spuID),
		attribute.Float64("sku.price", req.Price),
	)

	// Verify product exists
	product, err := s.repo.GetBySPU(ctx, spuID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch parent product")
		log.Error("failed to fetch parent product for SKU creation",
			zap.Error(err),
			zap.String("spu_id", spuID),
		)
		return nil, errors.NewInternalError(err)
	}
	if product == nil {
		span.SetStatus(codes.Error, "parent product not found")
		return nil, errors.NewNotFound(fmt.Sprintf("product %s not found", spuID))
	}

	// Build SKU
	now := time.Now().UTC()
	skuStatus := domain.SKUStatusActive
	if req.Stock == 0 {
		skuStatus = domain.SKUStatusOutOfStock
	}

	sku := &domain.SKU{
		SPUID:     spuID,
		SKUID:     generateID("SKU"),
		Price:     req.Price,
		SalePrice: req.SalePrice,
		Stock:     req.Stock,
		Weight:    req.Weight,
		Length:    req.Length,
		Width:     req.Width,
		Height:    req.Height,
		Status:    skuStatus,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Persist
	if err := s.repo.CreateSKU(ctx, sku); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create SKU")
		observability.BusinessErrorsTotal.WithLabelValues("sku", "CREATE_FAILED").Inc()
		log.Error("failed to create SKU",
			zap.Error(err),
			zap.String("spu_id", spuID),
		)
		return nil, errors.NewInternalError(err)
	}

	span.SetAttributes(attribute.String("sku.id", sku.SKUID))

	// Publish event
	event := domain.NewSKUUpdatedEvent(sku.SPUID, sku.SKUID, sku.Price, sku.Stock, string(sku.Status))
	event.SalePrice = sku.SalePrice
	if payload, err := sonic.Marshal(event); err == nil {
		if pubErr := s.publisher.Publish(ctx, "sku.events", sku.SKUID, payload); pubErr != nil {
			log.Warn("failed to publish SKUUpdatedEvent",
				zap.Error(pubErr),
				zap.String("sku_id", sku.SKUID),
			)
		}
	}

	log.Info("SKU created",
		zap.String("sku_id", sku.SKUID),
		zap.String("spu_id", spuID),
		zap.Float64("price", sku.Price),
	)

	return ToSKUResponse(sku), nil
}

// ListSKUsByProduct returns all SKUs belonging to a product.
func (s *SKUService) ListSKUsByProduct(ctx context.Context, spuID string) ([]SKUResponse, error) {
	ctx, span := s.tracer.Start(ctx, "application.sku.list_by_product")
	defer span.End()

	log := observability.LogWithTrace(ctx)

	if spuID == "" {
		return nil, errors.NewValidation("spu_id is required")
	}

	span.SetAttributes(attribute.String("spu.id", spuID))

	skus, err := s.repo.ListSKUsByProduct(ctx, spuID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list SKUs by product")
		observability.BusinessErrorsTotal.WithLabelValues("sku", "LIST_BY_PRODUCT_FAILED").Inc()
		log.Error("failed to list SKUs by product",
			zap.Error(err),
			zap.String("spu_id", spuID),
		)
		return nil, errors.NewInternalError(err)
	}

	responses := make([]SKUResponse, 0, len(skus))
	for i := range skus {
		responses = append(responses, *ToSKUResponse(&skus[i]))
	}

	span.SetAttributes(attribute.Int("sku.count", len(responses)))

	return responses, nil
}
