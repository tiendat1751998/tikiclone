package domain

import (
	"testing"

	"github.com/tikiclone/tiki/services/product-catalog/internal/domain"
)

func TestNewProduct(t *testing.T) {
	p := domain.NewProduct("SHOP-001", "Test Product", "Description", "CAT-001", "SGD")
	if p.Name != "Test Product" {
		t.Errorf("expected Test Product, got %s", p.Name)
	}
	if p.Status != domain.ProductStatusDraft {
		t.Errorf("expected draft, got %s", p.Status)
	}
	if p.ShopID != "SHOP-001" {
		t.Errorf("expected SHOP-001, got %s", p.ShopID)
	}
}

func TestProduct_Activate(t *testing.T) {
	p := domain.NewProduct("SHOP-001", "Test", "Desc", "CAT-001", "SGD")
	if err := p.Activate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p.Status != domain.ProductStatusActive {
		t.Errorf("expected active, got %s", p.Status)
	}
}

func TestProduct_Activate_InvalidState(t *testing.T) {
	p := domain.NewProduct("SHOP-001", "Test", "Desc", "CAT-001", "SGD")
	p.Activate()
	if err := p.Activate(); err == nil {
		t.Error("expected error for double activate")
	}
}

func TestProduct_Archive(t *testing.T) {
	p := domain.NewProduct("SHOP-001", "Test", "Desc", "CAT-001", "SGD")
	if err := p.Archive(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p.Status != domain.ProductStatusArchived {
		t.Errorf("expected archived, got %s", p.Status)
	}
}

func TestProduct_Update(t *testing.T) {
	p := domain.NewProduct("SHOP-001", "Test", "Desc", "CAT-001", "SGD")
	p.Update("Updated Name", "Updated Desc")
	if p.Name != "Updated Name" {
		t.Errorf("expected Updated Name, got %s", p.Name)
	}
	if p.Version != 2 {
		t.Errorf("expected version 2, got %d", p.Version)
	}
}

func TestNewSKU(t *testing.T) {
	sku := domain.NewSKU("PROD-001", "SKU-001", "color:red", "SGD", 5000)
	if sku.SkuCode != "SKU-001" {
		t.Errorf("expected SKU-001, got %s", sku.SkuCode)
	}
	if sku.Price != 5000 {
		t.Errorf("expected 5000, got %d", sku.Price)
	}
	if sku.Status != domain.SKUStatusActive {
		t.Errorf("expected active, got %s", sku.Status)
	}
}

func TestNewCategory(t *testing.T) {
	c := domain.NewCategory("Electronics", "electronics", "", nil, 0)
	if c.Name != "Electronics" {
		t.Errorf("expected Electronics, got %s", c.Name)
	}
	if c.Depth != 0 {
		t.Errorf("expected depth 0, got %d", c.Depth)
	}
	if !c.IsActive {
		t.Error("expected active")
	}
}

func TestNewCategory_WithParent(t *testing.T) {
	parentID := "CAT-001"
	c := domain.NewCategory("Phones", "phones", "Smartphones", &parentID, 1)
	if c.ParentID == nil || *c.ParentID != "CAT-001" {
		t.Errorf("expected CAT-001, got %v", c.ParentID)
	}
	if c.Depth != 1 {
		t.Errorf("expected depth 1, got %d", c.Depth)
	}
}

func TestNewAttribute(t *testing.T) {
	a := domain.NewAttribute("CAT-001", "Color", "color", domain.AttributeTypeSelect)
	if a.Type != domain.AttributeTypeSelect {
		t.Errorf("expected select type, got %s", a.Type)
	}
	if a.CategoryID != "CAT-001" {
		t.Errorf("expected CAT-001, got %s", a.CategoryID)
	}
}

func TestNewProductMedia(t *testing.T) {
	m := &domain.Media{ID: "MEDIA-001", ProductID: "PROD-001", Type: domain.MediaTypeImage, URL: "http://img.url"}
	if m.Type != domain.MediaTypeImage {
		t.Errorf("expected image type, got %s", m.Type)
	}
	if m.ProductID != "PROD-001" {
		t.Errorf("expected PROD-001, got %s", m.ProductID)
	}
}
