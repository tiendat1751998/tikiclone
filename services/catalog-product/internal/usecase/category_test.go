package usecase

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCategoryRepo struct {
	mock.Mock
}

func (m *mockCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) List(ctx context.Context, parentID string, level int32) ([]domain.Category, error) {
	args := m.Called(ctx, parentID, level)
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *mockCategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *mockCategoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

type mockCategoryCache struct {
	mock.Mock
}

func (m *mockCategoryCache) GetAll(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

func (m *mockCategoryCache) SetAll(ctx context.Context, categories []domain.Category) error {
	args := m.Called(ctx, categories)
	return args.Error(0)
}

func (m *mockCategoryCache) DeleteAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestCategoryUseCase_Create(t *testing.T) {
	repo := new(mockCategoryRepo)
	cache := new(mockCategoryCache)
	uc := NewCategoryUseCase(repo, cache, nil)

	t.Run("success", func(t *testing.T) {
		category := &domain.Category{
			Name:  "Electronics",
			Level: 1,
		}

		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil).Once()

		result, err := uc.Create(context.Background(), category)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.CategoryID)
	})

	t.Run("validation error", func(t *testing.T) {
		category := &domain.Category{}

		result, err := uc.Create(context.Background(), category)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryUseCase_GetByID(t *testing.T) {
	repo := new(mockCategoryRepo)
	cache := new(mockCategoryCache)
	uc := NewCategoryUseCase(repo, cache, nil)

	t.Run("found", func(t *testing.T) {
		expected := &domain.Category{CategoryID: "cat-1", Name: "Electronics"}
		repo.On("GetByID", mock.Anything, "cat-1").Return(expected, nil).Once()

		result, err := uc.GetByID(context.Background(), "cat-1")
		assert.NoError(t, err)
		assert.Equal(t, "Electronics", result.Name)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, "cat-404").Return(nil, nil).Once()

		result, err := uc.GetByID(context.Background(), "cat-404")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCategoryUseCase_List(t *testing.T) {
	repo := new(mockCategoryRepo)
	cache := new(mockCategoryCache)
	uc := NewCategoryUseCase(repo, cache, nil)

	t.Run("list root categories", func(t *testing.T) {
		categories := []domain.Category{
			{CategoryID: "cat-1", Name: "Electronics"},
			{CategoryID: "cat-2", Name: "Fashion"},
		}

		repo.On("List", mock.Anything, "", int32(0)).Return(categories, nil).Once()

		result, err := uc.List(context.Background(), "", 0)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})
}
