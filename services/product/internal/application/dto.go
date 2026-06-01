package application

import (
	"time"

	"github.com/tikiclone/tiki/services/product/internal/domain"
)

// -----------------------------------------------------------------------------
// Pagination & Search
// -----------------------------------------------------------------------------

type PaginationParams struct {
	Page          int    `json:"page"           validate:"min=1"`
	Size          int    `json:"size"           validate:"min=1,max=100"`
	SortBy        string `json:"sort_by"        validate:"omitempty,oneof=relevance price created_at updated_at sales rating popularity name"`
	SortDirection string `json:"sort_direction" validate:"omitempty,oneof=ASC DESC"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size < 1 {
		p.Size = 20
	}
	if p.Size > 100 {
		p.Size = 100
	}
	if p.SortDirection != "ASC" && p.SortDirection != "DESC" {
		p.SortDirection = "DESC"
	}
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Size
}

type SearchParams struct {
	PaginationParams
	Query   string            `json:"query"`
	Filters map[string]string `json:"filters,omitempty"`
}

func (sp *SearchParams) ToProductFilter() domain.ProductFilter {
	sp.PaginationParams.Normalize()
	f := domain.ProductFilter{
		Page:      sp.Page,
		Size:      sp.Size,
		Search:    sp.Query,
		SortBy:    domain.SortField(sp.SortBy),
		SortOrder: domain.SortDirection(sp.SortDirection),
	}
	f.Normalize()
	return f
}

// -----------------------------------------------------------------------------
// Product DTOs
// -----------------------------------------------------------------------------

type CreateProductRequest struct {
	Title           string                `json:"title"             validate:"required,max=255"`
	Description     string                `json:"description"       validate:"max=5000"`
	CategoryID      string                `json:"category_id"       validate:"required"`
	BrandID         string                `json:"brand_id"          validate:"omitempty"`
	SellerID        string                `json:"seller_id"         validate:"required"`
	SKUs            []CreateSKURequest    `json:"skus"              validate:"required,min=1,dive"`
	Images          []CreateImageRequest  `json:"images"            validate:"dive"`
	Attributes      []SetAttributeRequest `json:"attributes"        validate:"dive"`
	IdempotencyKey  string                `json:"idempotency_key"   validate:"omitempty,max=128"`
}

type UpdateProductRequest struct {
	Title       string `json:"title"       validate:"omitempty,max=255"`
	Description string `json:"description" validate:"max=5000"`
	CategoryID  string `json:"category_id" validate:"omitempty"`
	BrandID     string `json:"brand_id"    validate:"omitempty"`
	Status      string `json:"status"      validate:"omitempty,oneof=DRAFT PENDING_REVIEW ACTIVE INACTIVE REJECTED DELETED"`
}

type ProductResponse struct {
	ID          int64              `json:"id"`
	SPUID       string             `json:"spu_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	CategoryID  string             `json:"category_id"`
	BrandID     string             `json:"brand_id"`
	SellerID    string             `json:"seller_id"`
	Status      string             `json:"status"`
	SKUs        []SKUResponse      `json:"skus,omitempty"`
	Images      []ImageResponse    `json:"images,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Size       int               `json:"size"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
}

func ToProductResponse(p *domain.Product) *ProductResponse {
	if p == nil {
		return nil
	}
	skus := make([]SKUResponse, 0, len(p.SKUs))
	for _, sku := range p.SKUs {
		if resp := ToSKUResponse(&sku); resp != nil {
			skus = append(skus, *resp)
		}
	}
	images := make([]ImageResponse, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, ImageResponse{
			ID:        img.ID,
			URL:       img.URL,
			AltText:   img.AltText,
			SortOrder: img.SortOrder,
			IsPrimary: img.IsPrimary,
		})
	}
	return &ProductResponse{
		ID:          p.ID,
		SPUID:       p.SPUID,
		Title:       p.Title,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		BrandID:     p.BrandID,
		SellerID:    p.SellerID,
		Status:      string(p.Status),
		SKUs:        skus,
		Images:      images,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToProductListResponse(pl *domain.ProductList) *ProductListResponse {
	products := make([]ProductResponse, 0, len(pl.Products))
	for i := range pl.Products {
		if resp := ToProductResponse(&pl.Products[i]); resp != nil {
			products = append(products, *resp)
		}
	}
	return &ProductListResponse{
		Products:   products,
		Total:      pl.Total,
		Page:       pl.Page,
		Size:       pl.Size,
		TotalPages: pl.TotalPages(),
		HasNext:    pl.HasNext(),
		HasPrev:    pl.HasPrevious(),
	}
}

// -----------------------------------------------------------------------------
// SKU DTOs
// -----------------------------------------------------------------------------

type CreateSKURequest struct {
	Price     float64 `json:"price"       validate:"required,gt=0"`
	SalePrice float64 `json:"sale_price"  validate:"min=0"`
	Stock     int32   `json:"stock"       validate:"min=0"`
	Weight    float64 `json:"weight"      validate:"min=0"`
	Length    float64 `json:"length"      validate:"min=0"`
	Width     float64 `json:"width"       validate:"min=0"`
	Height    float64 `json:"height"      validate:"min=0"`
}

type UpdateSKURequest struct {
	Price     float64 `json:"price"      validate:"min=0"`
	SalePrice float64 `json:"sale_price" validate:"min=0"`
	Stock     int32   `json:"stock"      validate:"min=0"`
	Status    string  `json:"status"     validate:"omitempty,oneof=ACTIVE INACTIVE OUT_OF_STOCK"`
}

type SKUResponse struct {
	ID             int64     `json:"id"`
	SPUID          string    `json:"spu_id"`
	SKUID          string    `json:"sku_id"`
	Price          float64   `json:"price"`
	SalePrice      float64   `json:"sale_price"`
	Stock          int32     `json:"stock"`
	Status         string    `json:"status"`
	EffectivePrice float64   `json:"effective_price"`
	IsAvailable    bool      `json:"is_available"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func ToSKUResponse(s *domain.SKU) *SKUResponse {
	if s == nil {
		return nil
	}
	return &SKUResponse{
		ID:             s.ID,
		SPUID:          s.SPUID,
		SKUID:          s.SKUID,
		Price:          s.Price,
		SalePrice:      s.SalePrice,
		Stock:          s.Stock,
		Status:         string(s.Status),
		EffectivePrice: s.EffectivePrice(),
		IsAvailable:    s.IsAvailable(),
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

// -----------------------------------------------------------------------------
// Image DTOs
// -----------------------------------------------------------------------------

type CreateImageRequest struct {
	URL       string `json:"url"        validate:"required,url"`
	AltText   string `json:"alt_text"   validate:"max=255"`
	SortOrder int    `json:"sort_order" validate:"min=0"`
	IsPrimary bool   `json:"is_primary"`
}

type ImageResponse struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

// -----------------------------------------------------------------------------
// Category DTOs
// -----------------------------------------------------------------------------

type CreateCategoryRequest struct {
	Name      string `json:"name"       validate:"required,max=255"`
	Slug      string `json:"slug"       validate:"required,max=255"`
	ParentID  string `json:"parent_id,omitempty" validate:"omitempty"`
	SortOrder int    `json:"sort_order" validate:"min=0"`
	ImageURL  string `json:"image_url,omitempty"  validate:"omitempty,url"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name"       validate:"omitempty,max=255"`
	Slug      string `json:"slug"       validate:"omitempty,max=255"`
	ParentID  string `json:"parent_id,omitempty" validate:"omitempty"`
	SortOrder int    `json:"sort_order" validate:"min=0"`
	ImageURL  string `json:"image_url,omitempty"  validate:"omitempty,url"`
	IsActive  bool   `json:"is_active"`
}

type CategoryResponse struct {
	ID         int64     `json:"id"`
	CategoryID string    `json:"category_id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentID   string    `json:"parent_id,omitempty"`
	Level      int       `json:"level"`
	SortOrder  int       `json:"sort_order"`
	ImageURL   string    `json:"image_url,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CategoryTreeNode struct {
	CategoryResponse
	Children []CategoryTreeNode `json:"children,omitempty"`
	Depth    int                `json:"depth"`
}

type CategoryTreeResponse struct {
	Categories []CategoryTreeNode `json:"categories"`
}

func ToCategoryResponse(c *domain.Category) *CategoryResponse {
	if c == nil {
		return nil
	}
	return &CategoryResponse{
		ID:         c.ID,
		CategoryID: c.CategoryID,
		Name:       c.Name,
		Slug:       c.Slug,
		ParentID:   c.ParentID,
		Level:      c.Level,
		SortOrder:  c.SortOrder,
		ImageURL:   c.ImageURL,
		IsActive:   c.IsActive,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func ToCategoryTreeResponse(tree *domain.CategoryTree) *CategoryTreeResponse {
	if tree == nil {
		return &CategoryTreeResponse{}
	}
	nodes := make([]CategoryTreeNode, len(tree.Roots))
	for i, root := range tree.Roots {
		nodes[i] = toCategoryTreeNode(root)
	}
	return &CategoryTreeResponse{Categories: nodes}
}

func toCategoryTreeNode(node *domain.CategoryTreeNode) CategoryTreeNode {
	if node == nil {
		return CategoryTreeNode{}
	}
	children := make([]CategoryTreeNode, len(node.Children))
	for i, child := range node.Children {
		children[i] = toCategoryTreeNode(child)
	}
	catResp := ToCategoryResponse(&node.Category)
	if catResp == nil {
		return CategoryTreeNode{}
	}
	return CategoryTreeNode{
		CategoryResponse: *catResp,
		Children:         children,
		Depth:            node.Category.Level,
	}
}

// -----------------------------------------------------------------------------
// Attribute DTOs
// -----------------------------------------------------------------------------

type CreateAttributeRequest struct {
	Name         string `json:"name"          validate:"required,max=255"`
	Type         string `json:"type"          validate:"required,oneof=TEXT NUMBER BOOLEAN SELECT MULTI_SELECT COLOR"`
	CategoryID   string `json:"category_id"   validate:"required"`
	IsRequired   bool   `json:"is_required"`
	IsFilterable bool   `json:"is_filterable"`
	IsSearchable bool   `json:"is_searchable"`
}

type UpdateAttributeRequest struct {
	Name         string `json:"name"          validate:"omitempty,max=255"`
	Type         string `json:"type"          validate:"omitempty,oneof=TEXT NUMBER BOOLEAN SELECT MULTI_SELECT COLOR"`
	IsRequired   bool   `json:"is_required"`
	IsFilterable bool   `json:"is_filterable"`
	IsSearchable bool   `json:"is_searchable"`
}

type SetAttributeRequest struct {
	AttributeID string `json:"attribute_id" validate:"required"`
	ValueID     string `json:"value_id,omitempty" validate:"omitempty"`
	CustomValue string `json:"custom_value,omitempty" validate:"max=500"`
}

type CreateAttributeValueRequest struct {
	AttributeID  string `json:"attribute_id"   validate:"required"`
	Value        string `json:"value"          validate:"required,max=255"`
	DisplayValue string `json:"display_value,omitempty" validate:"max=255"`
	SortOrder    int    `json:"sort_order"     validate:"min=0"`
}

type AttributeResponse struct {
	ID           string                   `json:"id"`
	CategoryID   string                   `json:"category_id"`
	Name         string                   `json:"name"`
	Type         string                   `json:"type"`
	IsRequired   bool                     `json:"is_required"`
	IsFilterable bool                     `json:"is_filterable"`
	IsSearchable bool                     `json:"is_searchable"`
	SortOrder    int                      `json:"sort_order"`
	Values       []AttributeValueResponse `json:"values,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type AttributeValueResponse struct {
	ID           string `json:"id"`
	AttributeID  string `json:"attribute_id"`
	Value        string `json:"value"`
	DisplayValue string `json:"display_value,omitempty"`
	SortOrder    int    `json:"sort_order"`
}

func ToAttributeResponse(a *domain.Attribute) *AttributeResponse {
	if a == nil {
		return nil
	}
	vals := make([]AttributeValueResponse, 0, len(a.Values))
	for _, v := range a.Values {
		vals = append(vals, AttributeValueResponse{
			ID:           v.ID,
			AttributeID:  v.AttributeID,
			Value:        v.Value,
			DisplayValue: v.DisplayValue,
			SortOrder:    v.SortOrder,
		})
	}
	return &AttributeResponse{
		ID:           a.ID,
		CategoryID:   a.CategoryID,
		Name:         a.Name,
		Type:         string(a.Type),
		IsRequired:   a.IsRequired,
		IsFilterable: a.IsFilterable,
		IsSearchable: a.IsSearchable,
		SortOrder:    a.SortOrder,
		Values:       vals,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// -----------------------------------------------------------------------------
// Moderation DTOs
// -----------------------------------------------------------------------------

type ModerationRequest struct {
	Action     string `json:"action"      validate:"required,oneof=approve reject flag"`
	Reason     string `json:"reason"      validate:"required_if=Action reject,omitempty,max=500"`
	ReviewerID string `json:"reviewer_id,omitempty" validate:"omitempty"`
}

type ModerationResponse struct {
	ID         string    `json:"id"`
	SPUID      string    `json:"spu_id"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason,omitempty"`
	ReviewerID string    `json:"reviewer_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ToModerationResponse(mr *domain.ModerationRecord) *ModerationResponse {
	if mr == nil {
		return nil
	}
	return &ModerationResponse{
		ID:         mr.ID,
		SPUID:      mr.SPUID,
		Status:     string(mr.Status),
		Reason:     mr.Reason,
		ReviewerID: mr.ReviewerID,
		CreatedAt:  mr.CreatedAt,
		UpdatedAt:  mr.UpdatedAt,
	}
}
