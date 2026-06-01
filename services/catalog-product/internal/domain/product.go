package domain

import "time"

type Product struct {
	SPUID       string            `json:"spu_id" bson:"spu_id"`
	Title       string            `json:"title" bson:"title"`
	Description string            `json:"description" bson:"description"`
	CategoryID  string            `json:"category_id" bson:"category_id"`
	SKUs        []SKU             `json:"skus" bson:"skus"`
	Attributes  map[string]string `json:"attributes,omitempty" bson:"attributes,omitempty"`
	Images      []string          `json:"images,omitempty" bson:"images,omitempty"`
	SellerID    string            `json:"seller_id" bson:"seller_id"`
	Status      string            `json:"status" bson:"status"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" bson:"updated_at"`
}

type SKU struct {
	SKUID        string      `json:"sku_id" bson:"sku_id"`
	SPUID        string      `json:"spu_id" bson:"spu_id"`
	Price        float64     `json:"price" bson:"price"`
	ComparePrice float64     `json:"compare_price" bson:"compare_price"`
	Stock        int32       `json:"stock" bson:"stock"`
	Variations   []Variation `json:"variations" bson:"variations"`
	Image        string      `json:"image,omitempty" bson:"image,omitempty"`
	Status       string      `json:"status" bson:"status"`
}

type Variation struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type Category struct {
 	CategoryID string     `json:"category_id" bson:"category_id"`
 	Name       string     `json:"name" bson:"name"`
 	Slug       string     `json:"slug,omitempty" bson:"slug,omitempty"`
 	ParentID   string     `json:"parent_id,omitempty" bson:"parent_id,omitempty"`
 	Level      int32      `json:"level" bson:"level"`
 	SortOrder  int32      `json:"sort_order" bson:"sort_order"`
 	Children   []Category `json:"children,omitempty" bson:"children,omitempty"`
 	ProductCount int64    `json:"product_count" bson:"product_count"`
 	ImageURL   string     `json:"image_url,omitempty" bson:"image_url,omitempty"`
 	URLPath    string     `json:"url_path,omitempty" bson:"url_path,omitempty"`
 }

type ProductFilter struct {
	Page       int
	Size       int
	CategoryID string
	SellerID   string
	Search     string
	MinPrice   float64
	MaxPrice   float64
	SortBy     string
	SortOrder  string
	IsDeal     bool
}

type ProductList struct {
	Products []Product
	Total    int64
	Page     int
	Size     int
}

const (
	ProductStatusActive   = "ACTIVE"
	ProductStatusInactive = "INACTIVE"
	ProductStatusDraft    = "DRAFT"
	SKUStatusActive       = "ACTIVE"
	SKUStatusInactive     = "INACTIVE"
)
