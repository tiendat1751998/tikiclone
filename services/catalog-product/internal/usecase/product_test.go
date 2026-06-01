package usecase

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) Create(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *mockProductRepo) GetByID(ctx context.Context, spuID string) (*domain.Product, error) {
	args := m.Called(ctx, spuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductRepo) List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*domain.ProductList), args.Error(1)
}

func (m *mockProductRepo) Update(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *mockProductRepo) Delete(ctx context.Context, spuID string) error {
	args := m.Called(ctx, spuID)
	return args.Error(0)
}

func (m *mockProductRepo) GetSKU(ctx context.Context, skuID string) (*domain.SKU, error) {
	args := m.Called(ctx, skuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SKU), args.Error(1)
}

func (m *mockProductRepo) BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]*domain.SKU, error) {
	args := m.Called(ctx, skuIDs)
	return args.Get(0).(map[string]*domain.SKU), args.Error(1)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, spuID string) (*domain.Product, error) {
	args := m.Called(ctx, spuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *mockCache) Delete(ctx context.Context, spuID string) error {
	args := m.Called(ctx, spuID)
	return args.Error(0)
}

func (m *mockCache) GetOrFetch(ctx context.Context, spuID string, fetchFn func() (*domain.Product, error)) (*domain.Product, error) {
	args := m.Called(ctx, spuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockCache) GetList(ctx context.Context, cacheKey string) (*domain.ProductList, error) {
	args := m.Called(ctx, cacheKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductList), args.Error(1)
}

func (m *mockCache) SetList(ctx context.Context, cacheKey string, list *domain.ProductList) error {
	args := m.Called(ctx, cacheKey, list)
	return args.Error(0)
}

func TestProductUseCase_Create(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)
	uc := NewProductUseCase(repo, cache, nil)

	t.Run("success", func(t *testing.T) {
		product := &domain.Product{
			Title:      "Test Product",
			CategoryID: "cat-1",
			SellerID:   "seller-1",
			SKUs: []domain.SKU{
				{Price: 100.0, Stock: 10},
			},
		}

		repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Product")).Return(nil).Once()

		result, err := uc.Create(context.Background(), product)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.SPUID)
		repo.AssertExpectations(t)
	})

	t.Run("validation error - empty title", func(t *testing.T) {
		product := &domain.Product{
			CategoryID: "cat-1",
			SellerID:   "seller-1",
			SKUs: []domain.SKU{
				{Price: 100.0, Stock: 10},
			},
		}

		result, err := uc.Create(context.Background(), product)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("validation error - no skus", func(t *testing.T) {
		product := &domain.Product{
			Title:      "Test",
			CategoryID: "cat-1",
			SellerID:   "seller-1",
		}

		result, err := uc.Create(context.Background(), product)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("validation error - invalid price", func(t *testing.T) {
		product := &domain.Product{
			Title:      "Test",
			CategoryID: "cat-1",
			SellerID:   "seller-1",
			SKUs: []domain.SKU{
				{Price: 0, Stock: 10},
			},
		}

		result, err := uc.Create(context.Background(), product)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestProductUseCase_GetByID(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)
	uc := NewProductUseCase(repo, cache, nil)

	t.Run("found in cache", func(t *testing.T) {
		expected := &domain.Product{SPUID: "spu-1", Title: "Test"}
		cache.On("GetOrFetch", mock.Anything, "spu-1").Return(expected, nil).Once()

		result, err := uc.GetByID(context.Background(), "spu-1")
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		cache.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		cache.On("GetOrFetch", mock.Anything, "spu-404").Return(nil, nil).Once()

		result, err := uc.GetByID(context.Background(), "spu-404")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestProductUseCase_List(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)
	uc := NewProductUseCase(repo, cache, nil)

	t.Run("success", func(t *testing.T) {
		expectedList := &domain.ProductList{
			Products: []domain.Product{{SPUID: "spu-1"}},
			Total:    1,
			Page:     1,
			Size:     20,
		}

		repo.On("List", mock.Anything, mock.AnythingOfType("domain.ProductFilter")).Return(expectedList, nil).Once()

		result, err := uc.List(context.Background(), domain.ProductFilter{Page: 1, Size: 20})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Products, 1)
	})

	t.Run("default pagination", func(t *testing.T) {
		expectedList := &domain.ProductList{
			Products: []domain.Product{},
			Total:    0,
			Page:     1,
			Size:     20,
		}

		repo.On("List", mock.Anything, mock.Anything).Return(expectedList, nil).Once()

		result, err := uc.List(context.Background(), domain.ProductFilter{})
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.Size)
	})
}

func TestProductUseCase_Delete(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)
	uc := NewProductUseCase(repo, cache, nil)

	t.Run("success", func(t *testing.T) {
		product := &domain.Product{SPUID: "spu-1", Title: "Test"}
		repo.On("GetByID", mock.Anything, "spu-1").Return(product, nil).Once()
		repo.On("Delete", mock.Anything, "spu-1").Return(nil).Once()
		cache.On("Delete", mock.Anything, "spu-1").Return(nil).Once()

		err := uc.Delete(context.Background(), "spu-1")
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, "spu-404").Return(nil, nil).Once()

		err := uc.Delete(context.Background(), "spu-404")
		assert.Error(t, err)
	})
}
