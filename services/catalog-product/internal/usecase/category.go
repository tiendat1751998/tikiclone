package usecase

import (
	"context"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/kafka"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	List(ctx context.Context, parentID string, level int32) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
}

type CategoryCache interface {
	GetAll(ctx context.Context) ([]domain.Category, error)
	SetAll(ctx context.Context, categories []domain.Category) error
	DeleteAll(ctx context.Context) error
}

type CategoryUseCase struct {
	repo     CategoryRepository
	cache    CategoryCache
	producer *kafka.Producer
}

func NewCategoryUseCase(repo CategoryRepository, cache CategoryCache, producer *kafka.Producer) *CategoryUseCase {
	return &CategoryUseCase{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

func (uc *CategoryUseCase) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.category.create")
	defer span.End()

	if category.Name == "" {
		return nil, errors.NewValidation("category name is required")
	}

	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, errors.NewInternalError(err)
	}

	if uc.producer != nil {
		event := domain.NewCategoryCreatedEvent(category)
		if payload, err := event.Marshal(); err == nil {

			uc.producer.Publish(ctx, kafka.Message{Key: category.CategoryID, Value: payload, Topic: "catalog.events"})

		}
	}

	return category, nil
}

func (uc *CategoryUseCase) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.category.get_by_id")
	defer span.End()

	category, err := uc.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

// GetBySlug retrieves a category by its slug (e.g., "dien-thoai" -> category)
func (uc *CategoryUseCase) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.category.get_by_slug")
	defer span.End()

	category, err := uc.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	if category == nil {
		return nil, domain.ErrCategoryNotFound
	}

	return category, nil
}

func (uc *CategoryUseCase) List(ctx context.Context, parentID string, level int32) ([]domain.Category, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "usecase.category.list")
	defer span.End()

	// Only cache root-level lists (parent="" and level<=0)
	if parentID == "" && level <= 0 && uc.cache != nil {
		categories, err := uc.cache.GetAll(ctx)
		if err != nil {
			observability.LogWithTrace(ctx).Error("cache get failed", zap.Error(err))
		}
		if categories != nil {
			return categories, nil
		}
	}

	categories, err := uc.repo.List(ctx, parentID, level)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// Cache root-level lists
	if parentID == "" && level <= 0 && uc.cache != nil {
		if err := uc.cache.SetAll(ctx, categories); err != nil {
			observability.LogWithTrace(ctx).Error("cache set failed", zap.Error(err))
		}
	}

	return categories, nil
}
