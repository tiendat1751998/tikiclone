package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tikiclone/tiki/services/product/internal/domain"
	app "github.com/tikiclone/tiki/services/product/internal/application"
)

// Type aliases for application types
type (
	CreateProductRequest = app.CreateProductRequest
	CreateSKURequest     = app.CreateSKURequest
	UpdateProductRequest = app.UpdateProductRequest
)

var NewProductService = app.NewProductService

// Mock ProductRepository
type mockProductRepo struct {
	products map[string]*domain.Product
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{products: make(map[string]*domain.Product)}
}

func (m *mockProductRepo) Create(ctx context.Context, product *domain.Product) error {
	m.products[product.SPUID] = product
	return nil
}

func (m *mockProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.ID == 0 && p.SPUID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProductRepo) GetBySPU(ctx context.Context, spuID string) (*domain.Product, error) {
	if p, ok := m.products[spuID]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockProductRepo) List(ctx context.Context, filter domain.ProductFilter) (*domain.ProductList, error) {
	var products []domain.Product
	for _, p := range m.products {
		products = append(products, *p)
	}
	return &domain.ProductList{Products: products, Total: int64(len(products)), Page: filter.Page, Size: filter.Size}, nil
}

func (m *mockProductRepo) Update(ctx context.Context, product *domain.Product) error {
	m.products[product.SPUID] = product
	return nil
}

func (m *mockProductRepo) Delete(ctx context.Context, id string) error {
	delete(m.products, id)
	return nil
}

func (m *mockProductRepo) Search(ctx context.Context, query string, filter domain.ProductFilter) (*domain.ProductList, error) {
	return m.List(ctx, filter)
}

func (m *mockProductRepo) GetSKU(ctx context.Context, skuID string) (*domain.SKU, error) {
	for _, p := range m.products {
		for _, sku := range p.SKUs {
			if sku.SKUID == skuID {
				return &sku, nil
			}
		}
	}
	return nil, nil
}

func (m *mockProductRepo) BatchGetSKUs(ctx context.Context, skuIDs []string) (map[string]*domain.SKU, error) {
	result := make(map[string]*domain.SKU)
	for _, p := range m.products {
		for _, sku := range p.SKUs {
			for _, id := range skuIDs {
				if sku.SKUID == id {
					result[id] = &sku
				}
			}
		}
	}
	return result, nil
}

func (m *mockProductRepo) CreateSKU(ctx context.Context, sku *domain.SKU) error {
	return nil
}

func (m *mockProductRepo) UpdateSKU(ctx context.Context, sku *domain.SKU) error {
	return nil
}

func (m *mockProductRepo) ListSKUsByProduct(ctx context.Context, spuID string) ([]domain.SKU, error) {
	if p, ok := m.products[spuID]; ok {
		return p.SKUs, nil
	}
	return nil, nil
}

// Mock ProductCache
type mockCache struct {
	data map[string]*domain.Product
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]*domain.Product)}
}

func (m *mockCache) Get(ctx context.Context, key string) (*domain.Product, error) {
	if p, ok := m.data[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockCache) Set(ctx context.Context, key string, product *domain.Product, ttl time.Duration) error {
	m.data[key] = product
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetchFn func() (*domain.Product, error)) (*domain.Product, error) {
	if p, ok := m.data[key]; ok {
		return p, nil
	}
	return fetchFn()
}

// Mock EventPublisher
type mockPublisher struct {
	events []mockEvent
}

type mockEvent struct {
	Topic string
	Key   string
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{}
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	m.events = append(m.events, mockEvent{Topic: topic, Key: key})
	return nil
}

func TestProductService_CreateProduct(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	tests := []struct {
		name    string
		req     CreateProductRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid product creation",
			req: CreateProductRequest{
				Title:      "Test Product",
				CategoryID: "CAT-001",
				SellerID:   "SELLER-001",
				SKUs: []CreateSKURequest{
					{Price: 99.99, Stock: 100},
				},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: CreateProductRequest{
				CategoryID: "CAT-001",
				SellerID:   "SELLER-001",
				SKUs: []CreateSKURequest{
					{Price: 99.99, Stock: 100},
				},
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing category",
			req: CreateProductRequest{
				Title:    "Test Product",
				SellerID: "SELLER-001",
				SKUs: []CreateSKURequest{
					{Price: 99.99, Stock: 100},
				},
			},
			wantErr: true,
			errMsg:  "category_id is required",
		},
		{
			name: "missing seller",
			req: CreateProductRequest{
				Title:      "Test Product",
				CategoryID: "CAT-001",
				SKUs: []CreateSKURequest{
					{Price: 99.99, Stock: 100},
				},
			},
			wantErr: true,
			errMsg:  "seller_id is required",
		},
		{
			name: "no SKUs",
			req: CreateProductRequest{
				Title:      "Test Product",
				CategoryID: "CAT-001",
				SellerID:   "SELLER-001",
				SKUs:       []CreateSKURequest{},
			},
			wantErr: true,
			errMsg:  "at least one SKU is required",
		},
		{
			name: "invalid SKU price",
			req: CreateProductRequest{
				Title:      "Test Product",
				CategoryID: "CAT-001",
				SellerID:   "SELLER-001",
				SKUs: []CreateSKURequest{
					{Price: 0, Stock: 100},
				},
			},
			wantErr: true,
			errMsg:  "price must be greater than 0",
		},
		{
			name: "negative SKU price",
			req: CreateProductRequest{
				Title:      "Test Product",
				CategoryID: "CAT-001",
				SellerID:   "SELLER-001",
				SKUs: []CreateSKURequest{
					{Price: -10, Stock: 100},
				},
			},
			wantErr: true,
			errMsg:  "price must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.CreateProduct(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("response should not be nil")
			}
			if resp.Title != tt.req.Title {
				t.Errorf("Title = %q, want %q", resp.Title, tt.req.Title)
			}
			if resp.SPUID == "" {
				t.Error("SPUID should not be empty")
			}
			if resp.Status != string(domain.ProductStatusDraft) {
				t.Errorf("Status = %q, want %q", resp.Status, domain.ProductStatusDraft)
			}
		})
	}
}

func TestProductService_GetProduct(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	// Create a product first
	createReq := CreateProductRequest{
		Title:      "Test Product",
		CategoryID: "CAT-001",
		SellerID:   "SELLER-001",
		SKUs:       []CreateSKURequest{{Price: 99.99, Stock: 100}},
	}
	created, err := service.CreateProduct(context.Background(), createReq)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Test: get existing product
	resp, err := service.GetProduct(context.Background(), created.SPUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
	if resp.SPUID != created.SPUID {
		t.Errorf("SPUID = %q, want %q", resp.SPUID, created.SPUID)
	}

	// Test: get non-existent product
	_, err = service.GetProduct(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Error("expected error for non-existent product")
	}

	// Test: empty SPUID
	_, err = service.GetProduct(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty SPUID")
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	// Create a product
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Title:      "To Delete",
		CategoryID: "CAT-001",
		SellerID:   "SELLER-001",
		SKUs:       []CreateSKURequest{{Price: 50, Stock: 10}},
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Delete it
	err = service.DeleteProduct(context.Background(), created.SPUID)
	if err != nil {
		t.Fatalf("failed to delete product: %v", err)
	}

	// Verify it's gone
	_, err = service.GetProduct(context.Background(), created.SPUID)
	if err == nil {
		t.Error("expected error after deletion")
	}

	// Test: delete non-existent
	err = service.DeleteProduct(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Error("expected error for non-existent product")
	}

	// Test: empty SPUID
	err = service.DeleteProduct(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty SPUID")
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	// Create a product
	created, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Title:      "Original Title",
		CategoryID: "CAT-001",
		SellerID:   "SELLER-001",
		SKUs:       []CreateSKURequest{{Price: 50, Stock: 10}},
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Update it
	updated, err := service.UpdateProduct(context.Background(), created.SPUID, UpdateProductRequest{
		Title: "Updated Title",
	})
	if err != nil {
		t.Fatalf("failed to update product: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated Title")
	}

	// Test: update non-existent
	_, err = service.UpdateProduct(context.Background(), "NONEXISTENT", UpdateProductRequest{Title: "X"})
	if err == nil {
		t.Error("expected error for non-existent product")
	}

	// Test: empty SPUID
	_, err = service.UpdateProduct(context.Background(), "", UpdateProductRequest{Title: "X"})
	if err == nil {
		t.Error("expected error for empty SPUID")
	}
}

func TestProductService_ListProducts(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	// Create multiple products
	for i := 0; i < 5; i++ {
		_, err := service.CreateProduct(context.Background(), CreateProductRequest{
			Title:      "Product",
			CategoryID: "CAT-001",
			SellerID:   "SELLER-001",
			SKUs:       []CreateSKURequest{{Price: 99.99, Stock: 100}},
		})
		if err != nil {
			t.Fatalf("failed to create product: %v", err)
		}
	}

	// List products
	filter := domain.ProductFilter{Page: 1, Size: 10}
	filter.Normalize()

	list, err := service.ListProducts(context.Background(), filter)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if list.Total < 5 {
		t.Errorf("Total = %d, want >= 5", list.Total)
	}
}

func TestProductService_CreateProduct_PublishesEvent(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	_, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Title:      "Event Test",
		CategoryID: "CAT-001",
		SellerID:   "SELLER-001",
		SKUs:       []CreateSKURequest{{Price: 99.99, Stock: 100}},
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	if len(publisher.events) == 0 {
		t.Error("expected event to be published")
	}
	if publisher.events[0].Topic != "product.events" {
		t.Errorf("event topic = %q, want %q", publisher.events[0].Topic, "product.events")
	}
}

func TestProductService_BatchGetSKUs(t *testing.T) {
	repo := newMockProductRepo()
	cache := newMockCache()
	publisher := newMockPublisher()
	service := NewProductService(repo, cache, publisher)

	// Empty SKU IDs
	result, err := service.BatchGetSKUs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty SKU IDs")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure mock errors work
var _ error = errors.New("test")
