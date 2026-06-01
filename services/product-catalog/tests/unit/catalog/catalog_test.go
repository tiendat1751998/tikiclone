package unit

import (
	"testing"

	"github.com/tikiclone/tiki/services/product-catalog/internal/domain"
)

func TestNewProduct(t *testing.T) {
	p := domain.NewProduct("shop-1", "Test Product", "A great product", "cat-1", "SGD")
	if p.ShopID != "shop-1" {
		t.Errorf("expected shop-1, got %s", p.ShopID)
	}
	if p.Name != "Test Product" {
		t.Errorf("expected Test Product, got %s", p.Name)
	}
	if p.Status != domain.ProductStatusDraft {
		t.Errorf("expected draft, got %s", p.Status)
	}
	if p.Version != 1 {
		t.Errorf("expected version 1, got %d", p.Version)
	}
}

func TestProduct_Activate(t *testing.T) {
	p := domain.NewProduct("shop-1", "Test Product", "A great product", "cat-1", "SGD")
	if err := p.Activate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != domain.ProductStatusActive {
		t.Errorf("expected active, got %s", p.Status)
	}
}

func TestProduct_Activate_InvalidState(t *testing.T) {
	p := domain.NewProduct("shop-1", "Test Product", "A great product", "cat-1", "SGD")
	p.Activate()
	if err := p.Activate(); err == nil {
		t.Error("expected error for double activate")
	}
}

func TestProduct_Archive(t *testing.T) {
	p := domain.NewProduct("shop-1", "Test Product", "A great product", "cat-1", "SGD")
	if err := p.Archive(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != domain.ProductStatusArchived {
		t.Errorf("expected archived, got %s", p.Status)
	}
}

func TestProduct_Update(t *testing.T) {
	p := domain.NewProduct("shop-1", "Test Product", "A great product", "cat-1", "SGD")
	p.Update("Updated Name", "Updated Desc")
	if p.Name != "Updated Name" {
		t.Errorf("expected Updated Name, got %s", p.Name)
	}
	if p.Version != 2 {
		t.Errorf("expected version 2, got %d", p.Version)
	}
}

func TestNewSKU(t *testing.T) {
	sku := domain.NewSKU("prod-1", "SKU-001", "Red / Large", "SGD", 2999)
	if sku.ProductID != "prod-1" {
		t.Errorf("expected prod-1, got %s", sku.ProductID)
	}
	if sku.Price != 2999 {
		t.Errorf("expected 2999, got %d", sku.Price)
	}
	sku.Stock = 10
	if !sku.IsAvailable() {
		t.Error("expected SKU to be available")
	}
}

func TestSKU_Reserve(t *testing.T) {
	sku := domain.NewSKU("prod-1", "SKU-001", "Red", "SGD", 100)
	sku.Stock = 10
	if err := sku.Reserve(3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sku.ReservedStock != 3 {
		t.Errorf("expected reserved 3, got %d", sku.ReservedStock)
	}
	if sku.AvailableStock() != 7 {
		t.Errorf("expected available 7, got %d", sku.AvailableStock())
	}
}

func TestSKU_ReserveInsufficient(t *testing.T) {
	sku := domain.NewSKU("prod-1", "SKU-001", "Red", "SGD", 100)
	sku.Stock = 5
	if err := sku.Reserve(10); err == nil {
		t.Error("expected error for insufficient stock")
	}
}

func TestNewCategory(t *testing.T) {
	c := domain.NewCategory("Electronics", "electronics", "All electronics", nil, 0)
	if c.Name != "Electronics" {
		t.Errorf("expected Electronics, got %s", c.Name)
	}
	if c.Slug != "electronics" {
		t.Errorf("expected electronics, got %s", c.Slug)
	}
	if !c.IsRoot() {
		t.Error("expected root category")
	}
	if c.Depth != 0 {
		t.Errorf("expected depth 0, got %d", c.Depth)
	}
}

func TestCategory_WithParent(t *testing.T) {
	parentID := "parent-1"
	c := domain.NewCategory("Phones", "phones", "Smartphones", &parentID, 1)
	if c.IsRoot() {
		t.Error("expected non-root category")
	}
	if c.Depth != 1 {
		t.Errorf("expected depth 1, got %d", c.Depth)
	}
}
