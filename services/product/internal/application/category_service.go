package application

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/services/product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// CategoryRepository defines the interface for category persistence
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	GetTree(ctx context.Context) (*domain.CategoryTree, error)
	List(ctx context.Context, parentID string) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
}

// CategoryCache defines the interface for category caching
type CategoryCache interface {
	Get(ctx context.Context, key string) (*domain.Category, error)
	Set(ctx context.Context, key string, category *domain.Category, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteTree(ctx context.Context) error
}

// CategoryService handles category use cases
type CategoryService struct {
	repo      CategoryRepository
	cache     CategoryCache
	publisher EventPublisher
}

// NewCategoryService creates a new CategoryService
func NewCategoryService(repo CategoryRepository, cache CategoryCache, publisher EventPublisher) *CategoryService {
	return &CategoryService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CategoryResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.create")
	defer span.End()

	if req.Name == "" {
		return nil, errors.NewValidation("name is required")
	}
	if req.Slug == "" {
		return nil, errors.NewValidation("slug is required")
	}

	existing, _ := s.repo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.NewDuplicate("category with this slug already exists")
	}

	category := &domain.Category{
		CategoryID: generateID("CAT"),
		Name:       req.Name,
		Slug:       req.Slug,
		ParentID:   req.ParentID,
		Level:      0,
		SortOrder:  req.SortOrder,
		ImageURL:   req.ImageURL,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.ParentID != "" {
		parent, err := s.repo.GetByID(ctx, req.ParentID)
		if err != nil {
			return nil, errors.NewInternalError(err)
		}
		if parent == nil {
			return nil, errors.NewValidation("parent category not found")
		}
		category.Level = parent.Level + 1
	}

	if err := s.repo.Create(ctx, category); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create category")
		return nil, errors.NewInternalError(err)
	}

	s.cache.DeleteTree(ctx)

	event := domain.NewCategoryUpdatedEvent(category)
	if payload, err := event.Marshal(); err == nil {
		s.publisher.Publish(ctx, "product.events", category.CategoryID, payload)
	}

	span.SetAttributes(attribute.String("category_id", category.CategoryID))
	return ToCategoryResponse(category), nil
}

// GetCategoryByID gets a category by ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, categoryID string) (*CategoryResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.get_by_id")
	defer span.End()

	if categoryID == "" {
		return nil, errors.NewValidation("category_id is required")
	}

	category, err := s.cache.Get(ctx, "category:"+categoryID)
	if err == nil && category != nil {
		return ToCategoryResponse(category), nil
	}

	category, err = s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	s.cache.Set(ctx, "category:"+categoryID, category, 10*time.Minute)
	return ToCategoryResponse(category), nil
}

// GetCategoryTree returns the full category tree
func (s *CategoryService) GetCategoryTree(ctx context.Context) (*CategoryTreeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.get_tree")
	defer span.End()

	tree, err := s.repo.GetTree(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	return ToCategoryTreeResponse(tree), nil
}

// ListCategories lists categories
func (s *CategoryService) ListCategories(ctx context.Context, parentID string) (*CategoryTreeResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.list")
	defer span.End()

	categories, err := s.repo.List(ctx, parentID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	nodes := make([]CategoryTreeNode, len(categories))
	for i, cat := range categories {
		children, _ := s.repo.List(ctx, cat.CategoryID)
		childNodes := make([]CategoryTreeNode, len(children))
		for j, child := range children {
			childNodes[j] = CategoryTreeNode{
				CategoryResponse: *ToCategoryResponse(&child),
				Depth:            child.Level,
			}
		}
		nodes[i] = CategoryTreeNode{
			CategoryResponse: *ToCategoryResponse(&cat),
			Children:         childNodes,
			Depth:            cat.Level,
		}
	}

	return &CategoryTreeResponse{Categories: nodes}, nil
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, categoryID string, req UpdateCategoryRequest) (*CategoryResponse, error) {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.update")
	defer span.End()

	if categoryID == "" {
		return nil, errors.NewValidation("category_id is required")
	}

	existing, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	if existing == nil {
		return nil, domain.ErrCategoryNotFound
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Slug != "" {
		existing.Slug = req.Slug
	}
	if req.SortOrder != 0 {
		existing.SortOrder = req.SortOrder
	}
	if req.ImageURL != "" {
		existing.ImageURL = req.ImageURL
	}
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, existing); err != nil {
		span.RecordError(err)
		return nil, errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, "category:"+categoryID)
	s.cache.DeleteTree(ctx)

	event := domain.NewCategoryUpdatedEvent(existing)
	if payload, err := event.Marshal(); err == nil {
		s.publisher.Publish(ctx, "product.events", categoryID, payload)
	}

	span.SetAttributes(attribute.String("category_id", categoryID))
	return ToCategoryResponse(existing), nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	ctx, span := otel.Tracer("product-service").Start(ctx, "application.category.delete")
	defer span.End()

	if categoryID == "" {
		return errors.NewValidation("category_id is required")
	}

	existing, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return errors.NewInternalError(err)
	}
	if existing == nil {
		return domain.ErrCategoryNotFound
	}

	children, _ := s.repo.List(ctx, categoryID)
	if len(children) > 0 {
		return errors.NewValidation("cannot delete category with children")
	}

	if err := s.repo.Delete(ctx, categoryID); err != nil {
		span.RecordError(err)
		return errors.NewInternalError(err)
	}

	s.cache.Delete(ctx, "category:"+categoryID)
	s.cache.DeleteTree(ctx)

	event := domain.NewCategoryUpdatedEvent(existing)
	if payload, err := event.Marshal(); err == nil {
		s.publisher.Publish(ctx, "product.events", categoryID, payload)
	}

	span.SetAttributes(attribute.String("category_id", categoryID))
	return nil
}
