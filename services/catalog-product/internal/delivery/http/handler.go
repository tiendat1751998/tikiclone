package http

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"github.com/tikiclone/tiki/services/catalog-product/internal/usecase"
	"go.opentelemetry.io/otel"
)

type Handler struct {
	productUseCase  *usecase.ProductUseCase
	categoryUseCase *usecase.CategoryUseCase
}

func NewHandler(productUC *usecase.ProductUseCase, categoryUC *usecase.CategoryUseCase) *Handler {
	return &Handler{
		productUseCase:  productUC,
		categoryUseCase: categoryUC,
	}
}

var reNonASCII = regexp.MustCompile(`[^\x00-\x7F]+`)

func slugify(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, "đ", "d")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "va")
	slug = reNonASCII.ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func toTimeStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func toProductResponse(p *domain.Product) ProductResponse {
	var brand string
	if p.Attributes != nil {
		brand = p.Attributes["brand"]
	}

	skus := make([]SKUResponse, 0, len(p.SKUs))
	for _, s := range p.SKUs {
		attrs := make(map[string]string)
		for _, v := range s.Variations {
			attrs[v.Name] = v.Value
		}
		var skuName string
		for _, v := range s.Variations {
			skuName = v.Value
			break
		}

		skus = append(skus, SKUResponse{
			ID:            s.SKUID,
			ProductID:     p.SPUID,
			Name:          skuName,
			Price:         s.Price,
			ComparePrice:  s.ComparePrice,
			Currency:      "VND",
			Stock:         s.Stock,
			ReservedStock: 0,
			Weight:        0,
			Dimensions:    "",
			Status:        strings.ToLower(s.Status),
			SortOrder:     0,
			Attributes:    attrs,
			CreatedAt:     toTimeStr(p.CreatedAt),
			UpdatedAt:     toTimeStr(p.UpdatedAt),
		})
	}

	media := make([]MediaResponse, 0, len(p.Images))
	for i, img := range p.Images {
		media = append(media, MediaResponse{
			ProductID:    p.SPUID,
			Type:         "image",
			URL:          img,
			ThumbnailURL: img,
			AltText:      p.Title,
			SortOrder:    int32(i),
			Status:       "active",
		})
	}

	return ProductResponse{
		ID:          p.SPUID,
		Name:        p.Title,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		Brand:       brand,
		Status:      strings.ToLower(p.Status),
		Condition:   "new",
		Weight:      0,
		Dimensions:  "",
		Version:     1,
		CreatedAt:   toTimeStr(p.CreatedAt),
		UpdatedAt:   toTimeStr(p.UpdatedAt),
		ShopID:      p.SellerID,
		SKUs:        skus,
		Media:       media,
		Attributes:  p.Attributes,
		SoldCount:   0,
	}
}

func toCategoryResponse(c domain.Category) CategoryResponse {
 	children := make([]CategoryResponse, 0, len(c.Children))
 	for _, child := range c.Children {
 		children = append(children, toCategoryResponse(child))
 	}
 	slug := c.Slug
 	if slug == "" {
 		slug = slugify(c.Name)
 	}
 	return CategoryResponse{
 		ID:           c.CategoryID,
 		Name:         c.Name,
 		Slug:         slug,
 		ParentID:     c.ParentID,
 		Description:  "",
 		ImageURL:     c.ImageURL,
 		SortOrder:    c.SortOrder,
 		IsActive:     true,
 		Depth:        c.Level,
 		Path:         c.URLPath,
 		Children:     children,
 		ProductCount: c.ProductCount,
 	}
 }

func (h *Handler) GetProduct(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.get")
	defer span.End()

	spuID := c.Param("spu_id")
	if spuID == "" {
		c.Error(domain.ErrInvalidProductData)
		return
	}

	product, err := h.productUseCase.GetByID(ctx, spuID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toProductResponse(product))
}

func (h *Handler) ListProducts(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.list")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

categoryID := c.Query("category_id")
 	categorySlug := c.Query("category_slug")
 	if categoryID == "" && categorySlug != "" {
 		cat, err := h.categoryUseCase.GetBySlug(ctx, categorySlug)
 		if err != nil {
 			_ = c.Error(err)
 			return
 		}
 		if cat != nil {
 			categoryID = cat.CategoryID
 		}
 	}

	filter := domain.ProductFilter{
		Page:       page,
		Size:       size,
		CategoryID: categoryID,
		SellerID:   c.Query("seller_id"),
		Search:     c.Query("search"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	}

	result, err := h.productUseCase.List(ctx, filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	products := make([]ProductResponse, 0, len(result.Products))
	for _, p := range result.Products {
		products = append(products, toProductResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    result.Total,
		"page":     result.Page,
		"size":     result.Size,
	})
}

func (h *Handler) SearchProducts(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.search")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "24"))
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

	filter := domain.ProductFilter{
		Page:       page,
		Size:       pageSize,
		CategoryID: c.Query("category_id"),
		SellerID:   c.Query("shop_id"),
		Search:     c.Query("q"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	}

	result, err := h.productUseCase.List(ctx, filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	products := make([]ProductResponse, 0, len(result.Products))
	for _, p := range result.Products {
		products = append(products, toProductResponse(&p))
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(pageSize)))

	c.JSON(http.StatusOK, SearchResultResponse{
		Products:   products,
		Total:      result.Total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetFeaturedProducts(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.featured")
	defer span.End()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := domain.ProductFilter{
		Page: 1,
		Size: limit,
	}

	result, err := h.productUseCase.List(ctx, filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	products := make([]ProductResponse, 0, len(result.Products))
	for _, p := range result.Products {
		products = append(products, toProductResponse(&p))
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) GetDealsProducts(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.deals")
	defer span.End()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filter := domain.ProductFilter{
		Page:     1,
		Size:     50,
		IsDeal:   true,
	}

	result, err := h.productUseCase.List(ctx, filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if len(result.Products) > limit {
		result.Products = result.Products[:limit]
	}

	products := make([]ProductResponse, 0, len(result.Products))
	for _, p := range result.Products {
		products = append(products, toProductResponse(&p))
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) GetFlashSale(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.flashsale")
	defer span.End()

	result, err := h.productUseCase.List(ctx, domain.ProductFilter{Page: 1, Size: 20, IsDeal: true})
	if err != nil {
		_ = c.Error(err)
		return
	}

	flashProducts := make([]map[string]any, 0, len(result.Products))
	for _, p := range result.Products {
		resp := toProductResponse(&p)
		var price, originalPrice float64
		var imageURL string
		if len(resp.SKUs) > 0 {
			price = resp.SKUs[0].Price
			originalPrice = resp.SKUs[0].ComparePrice
		}
		if len(resp.Media) > 0 {
			imageURL = resp.Media[0].URL
		}
		flashProducts = append(flashProducts, map[string]any{
			"id":             resp.ID,
			"name":           resp.Name,
			"image_url":      imageURL,
			"price":          price,
			"original_price": originalPrice,
		})
	}

	endTime := time.Now().Add(3 * time.Hour)

	c.JSON(http.StatusOK, gin.H{
		"end_time": endTime.Format(time.RFC3339),
		"products": flashProducts,
	})
}

func (h *Handler) CreateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.create")
	defer span.End()

	var req struct {
		Title       string              `json:"title" binding:"required"`
		Description string              `json:"description"`
		CategoryID  string              `json:"category_id" binding:"required"`
		SKUs        []domain.SKU        `json:"skus" binding:"required"`
		Attributes  map[string]string   `json:"attributes"`
		Images      []string            `json:"images"`
		SellerID    string              `json:"seller_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("Invalid request body").WithDetail("body", err.Error()))
		return
	}

	product := &domain.Product{
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		SKUs:        req.SKUs,
		Attributes:  req.Attributes,
		Images:      req.Images,
		SellerID:    req.SellerID,
	}

	created, err := h.productUseCase.Create(ctx, product)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, toProductResponse(created))
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.update")
	defer span.End()

	spuID := c.Param("spu_id")
	if spuID == "" {
		c.Error(domain.ErrInvalidProductData)
		return
	}

	var req struct {
		Title       string              `json:"title"`
		Description string              `json:"description"`
		CategoryID  string              `json:"category_id"`
		Attributes  map[string]string   `json:"attributes"`
		Images      []string            `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidation("Invalid request body"))
		return
	}

	product := &domain.Product{
		SPUID:       spuID,
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Attributes:  req.Attributes,
		Images:      req.Images,
	}

	if err := h.productUseCase.Update(ctx, product); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated"})
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.product.delete")
	defer span.End()

	spuID := c.Param("spu_id")
	if spuID == "" {
		c.Error(domain.ErrInvalidProductData)
		return
	}

	if err := h.productUseCase.Delete(ctx, spuID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func (h *Handler) ListCategories(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.category.list")
	defer span.End()

	parentID := c.Query("parent_id")
	levelStr := c.Query("level")
	var level int32
	if levelStr != "" {
		l, err := strconv.Atoi(levelStr)
		if err == nil {
			level = int32(l)
		}
	}

	categories, err := h.categoryUseCase.List(ctx, parentID, level)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := make([]CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, toCategoryResponse(cat))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCategory(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.category.get")
	defer span.End()

	categoryID := c.Param("category_id")
	if categoryID == "" {
		c.Error(domain.ErrInvalidCategory)
		return
	}

	category, err := h.categoryUseCase.GetByID(ctx, categoryID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toCategoryResponse(*category))
}

func (h *Handler) GetCategoryBySlug(c *gin.Context) {
	ctx, span := otel.Tracer("catalog-product").Start(c.Request.Context(), "handler.category.get_by_slug")
	defer span.End()

	slug := c.Param("slug")
	if slug == "" {
		c.Error(domain.ErrInvalidCategory)
		return
	}

	// OPTIMIZATION: Direct DB lookup by slug instead of loading all categories
	category, err := h.categoryUseCase.GetBySlug(ctx, slug)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if category == nil {
		c.Error(domain.ErrCategoryNotFound)
		return
	}

	c.JSON(http.StatusOK, toCategoryResponse(*category))
}
