package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// AttributeRepository defines the interface for attribute persistence.
type AttributeRepository interface {
	Create(ctx context.Context, attr *domain.Attribute) error
	GetByID(ctx context.Context, id string) (*domain.Attribute, error)
	ListByCategory(ctx context.Context, categoryID string) ([]domain.Attribute, error)
	Update(ctx context.Context, attr *domain.Attribute) error
	Delete(ctx context.Context, id string) error
	CreateValue(ctx context.Context, value *domain.AttributeValue) error
	ListValues(ctx context.Context, attributeID string) ([]domain.AttributeValue, error)
}

// AttributeCache defines the interface for attribute caching.
type AttributeCache interface {
	Get(ctx context.Context, id string) (*domain.Attribute, error)
	Set(ctx context.Context, id string, attr *domain.Attribute, ttl time.Duration) error
	Delete(ctx context.Context, id string) error
	InvalidateByCategory(ctx context.Context, categoryID string) error
}

// AttributeService handles product attribute use cases.
type AttributeService struct {
	repo      AttributeRepository
	cache     AttributeCache
	publisher EventPublisher
}

// NewAttributeService creates a new AttributeService.
func NewAttributeService(repo AttributeRepository, cache AttributeCache, publisher EventPublisher) *AttributeService {
	return &AttributeService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

// CreateAttribute creates a new attribute definition for a category.
func (s *AttributeService) CreateAttribute(ctx context.Context, req CreateAttributeRequest) (*AttributeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.create")
	defer span.End()

	if req.Name == "" {
		return nil, errors.NewValidation("attribute name is required")
	}
	if req.Type == "" {
		return nil, errors.NewValidation("attribute type is required")
	}

	now := time.Now().UTC()
	attr := &domain.Attribute{
		ID:           generateID("ATTR"),
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		Type:         domain.AttributeType(req.Type),
		IsRequired:   req.IsRequired,
		IsFilterable: req.IsFilterable,
		IsSearchable: req.IsSearchable,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := attr.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "domain validation failed")
		return nil, errors.NewValidation(err.Error())
	}

	if err := s.repo.Create(ctx, attr); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create attribute")
		observability.BusinessErrorsTotal.WithLabelValues("attribute", "CREATE_FAILED").Inc()
		return nil, errors.NewInternalError(err)
	}

	span.SetAttributes(attribute.String("attribute.id", attr.ID))
	return ToAttributeResponse(attr), nil
}

// GetAttribute retrieves a single attribute by ID.
func (s *AttributeService) GetAttribute(ctx context.Context, id string) (*AttributeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.get")
	defer span.End()

	if id == "" {
		return nil, errors.NewValidation("attribute id is required")
	}

	span.SetAttributes(attribute.String("attribute.id", id))

	cached, err := s.cache.Get(ctx, id)
	if err == nil && cached != nil {
		observability.CacheHitsTotal.WithLabelValues("attribute", "single").Inc()
		return ToAttributeResponse(cached), nil
	}
	observability.CacheMissesTotal.WithLabelValues("attribute", "single").Inc()

	attr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}
	if attr == nil {
		return nil, errors.NewNotFound(fmt.Sprintf("attribute %s not found", id))
	}

	if err := s.cache.Set(ctx, id, attr, 30*time.Minute); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to cache attribute", zap.Error(err))
	}

	return ToAttributeResponse(attr), nil
}

// ListAttributesByCategory returns all attributes for a given category.
func (s *AttributeService) ListAttributesByCategory(ctx context.Context, categoryID string) ([]AttributeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.list_by_category")
	defer span.End()

	if categoryID == "" {
		return nil, errors.NewValidation("category_id is required")
	}

	span.SetAttributes(attribute.String("category.id", categoryID))

	attrs, err := s.repo.ListByCategory(ctx, categoryID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	responses := make([]AttributeResponse, 0, len(attrs))
	for i := range attrs {
		responses = append(responses, *ToAttributeResponse(&attrs[i]))
	}

	return responses, nil
}

// UpdateAttribute modifies an existing attribute definition.
func (s *AttributeService) UpdateAttribute(ctx context.Context, id string, req UpdateAttributeRequest) (*AttributeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.update")
	defer span.End()

	if id == "" {
		return nil, errors.NewValidation("attribute id is required")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}
	if existing == nil {
		return nil, errors.NewNotFound(fmt.Sprintf("attribute %s not found", id))
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Type != "" {
		existing.Type = domain.AttributeType(req.Type)
	}
	existing.IsRequired = req.IsRequired
	existing.IsFilterable = req.IsFilterable
	existing.IsSearchable = req.IsSearchable
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		span.RecordError(err)
		return nil, errors.NewValidation(err.Error())
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, id)
	if existing.CategoryID != "" {
		s.cache.InvalidateByCategory(ctx, existing.CategoryID)
	}

	return ToAttributeResponse(existing), nil
}

// DeleteAttribute removes an attribute definition.
func (s *AttributeService) DeleteAttribute(ctx context.Context, id string) error {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.delete")
	defer span.End()

	if id == "" {
		return errors.NewValidation("attribute id is required")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return errors.NewInternalError(err)
	}
	if existing == nil {
		return errors.NewNotFound(fmt.Sprintf("attribute %s not found", id))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		return errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, id)
	if existing.CategoryID != "" {
		s.cache.InvalidateByCategory(ctx, existing.CategoryID)
	}

	return nil
}

// CreateAttributeValue creates a new predefined value for an attribute.
func (s *AttributeService) CreateAttributeValue(ctx context.Context, req CreateAttributeValueRequest) (*AttributeValueResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.create_value")
	defer span.End()

	if req.AttributeID == "" {
		return nil, errors.NewValidation("attribute_id is required")
	}
	if req.Value == "" {
		return nil, errors.NewValidation("value is required")
	}

	attr, err := s.repo.GetByID(ctx, req.AttributeID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}
	if attr == nil {
		return nil, errors.NewNotFound(fmt.Sprintf("attribute %s not found", req.AttributeID))
	}

	value := &domain.AttributeValue{
		AttributeID:  req.AttributeID,
		Value:        req.Value,
		DisplayValue: req.DisplayValue,
		SortOrder:    req.SortOrder,
	}

	if err := s.repo.CreateValue(ctx, value); err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, req.AttributeID)

	return &AttributeValueResponse{
		ID:           value.ID,
		AttributeID:  value.AttributeID,
		Value:        value.Value,
		DisplayValue: value.DisplayValue,
		SortOrder:    value.SortOrder,
	}, nil
}

// ListAttributeValues returns all predefined values for an attribute.
func (s *AttributeService) ListAttributeValues(ctx context.Context, attributeID string) ([]AttributeValueResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.attribute.list_values")
	defer span.End()

	if attributeID == "" {
		return nil, errors.NewValidation("attribute_id is required")
	}

	values, err := s.repo.ListValues(ctx, attributeID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	responses := make([]AttributeValueResponse, 0, len(values))
	for i := range values {
		responses = append(responses, AttributeValueResponse{
			ID:           values[i].ID,
			AttributeID:  values[i].AttributeID,
			Value:        values[i].Value,
			DisplayValue: values[i].DisplayValue,
			SortOrder:    values[i].SortOrder,
		})
	}

	return responses, nil
}
