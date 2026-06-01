package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/services/product/internal/application"
	"github.com/tikiclone/tiki/services/product/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Handler handles HTTP requests for the product service
type Handler struct {
	productService    *application.ProductService
	categoryService   *application.CategoryService
	attributeService  *application.AttributeService
}

// NewHandler creates a new HTTP handler
func NewHandler(productService *application.ProductService, categoryService *application.CategoryService, attributeService *application.AttributeService) *Handler {
	return &Handler{
		productService:   productService,
		categoryService:  categoryService,
		attributeService: attributeService,
	}
}

// RegisterRoutes registers all product service routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		products.POST("", h.CreateProduct)
		products.GET("", h.ListProducts)
		products.GET("/search", h.SearchProducts)
		products.GET("/:spu_id", h.GetProduct)
		products.PUT("/:spu_id", h.UpdateProduct)
		products.DELETE("/:spu_id", h.DeleteProduct)
	}

	categories := router.Group("/categories")
	{
		categories.POST("", h.CreateCategory)
		categories.GET("", h.ListCategories)
		categories.GET("/tree", h.GetCategoryTree)
		categories.GET("/:category_id", h.GetCategory)
		categories.PUT("/:category_id", h.UpdateCategory)
		categories.DELETE("/:category_id", h.DeleteCategory)
	}

	attributes := router.Group("/attributes")
	{
		attributes.POST("", h.CreateAttribute)
		attributes.GET("/:attribute_id", h.GetAttribute)
		attributes.GET("/by-category/:category_id", h.ListAttributesByCategory)
		attributes.PUT("/:attribute_id", h.UpdateAttribute)
		attributes.DELETE("/:attribute_id", h.DeleteAttribute)
		attributes.POST("/values", h.CreateAttributeValue)
		attributes.GET("/:attribute_id/values", h.ListAttributeValues)
	}
}

// CreateProduct handles POST /api/v1/products
func (h *Handler) CreateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.create_product")
	defer span.End()

	// [SECURITY] Limit request body size to prevent memory exhaustion
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1MB max

	var req application.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.productService.CreateProduct(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("spu_id", resp.SPUID))
	c.JSON(http.StatusCreated, resp)
}

// GetProduct handles GET /api/v1/products/:spu_id
func (h *Handler) GetProduct(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.get_product")
	defer span.End()

	spuID := c.Param("spu_id")
	resp, err := h.productService.GetProduct(ctx, spuID)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("spu_id", resp.SPUID))
	c.JSON(http.StatusOK, resp)
}

// ListProducts handles GET /api/v1/products
func (h *Handler) ListProducts(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.list_products")
	defer span.End()

	filter := parseProductFilter(c)
	resp, err := h.productService.ListProducts(ctx, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchProducts handles GET /api/v1/products/search
func (h *Handler) SearchProducts(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.search_products")
	defer span.End()

	query := c.Query("q")
	filter := parseProductFilter(c)
	resp, err := h.productService.SearchProducts(ctx, query, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateProduct handles PUT /api/v1/products/:spu_id
func (h *Handler) UpdateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.update_product")
	defer span.End()

	spuID := c.Param("spu_id")
	var req application.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.productService.UpdateProduct(ctx, spuID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("spu_id", resp.SPUID))
	c.JSON(http.StatusOK, resp)
}

// DeleteProduct handles DELETE /api/v1/products/:spu_id
func (h *Handler) DeleteProduct(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.delete_product")
	defer span.End()

	spuID := c.Param("spu_id")
	if err := h.productService.DeleteProduct(ctx, spuID); err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("spu_id", spuID))
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

// CreateCategory handles POST /api/v1/categories
func (h *Handler) CreateCategory(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.create_category")
	defer span.End()

	var req application.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.categoryService.CreateCategory(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("category_id", resp.CategoryID))
	c.JSON(http.StatusCreated, resp)
}

// GetCategory handles GET /api/v1/categories/:category_id
func (h *Handler) GetCategory(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.get_category")
	defer span.End()

	categoryID := c.Param("category_id")
	resp, err := h.categoryService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("category_id", resp.CategoryID))
	c.JSON(http.StatusOK, resp)
}

// GetCategoryTree handles GET /api/v1/categories/tree
func (h *Handler) GetCategoryTree(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.get_category_tree")
	defer span.End()

	resp, err := h.categoryService.GetCategoryTree(ctx)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListCategories handles GET /api/v1/categories
func (h *Handler) ListCategories(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.list_categories")
	defer span.End()

	parentID := c.Query("parent_id")
	resp, err := h.categoryService.ListCategories(ctx, parentID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateCategory handles PUT /api/v1/categories/:category_id
func (h *Handler) UpdateCategory(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.update_category")
	defer span.End()

	categoryID := c.Param("category_id")
	var req application.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.categoryService.UpdateCategory(ctx, categoryID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("category_id", resp.CategoryID))
	c.JSON(http.StatusOK, resp)
}

// DeleteCategory handles DELETE /api/v1/categories/:category_id
func (h *Handler) DeleteCategory(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.delete_category")
	defer span.End()

	categoryID := c.Param("category_id")
	if err := h.categoryService.DeleteCategory(ctx, categoryID); err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("category_id", categoryID))
	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

// CreateAttribute handles POST /api/v1/attributes
func (h *Handler) CreateAttribute(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.create_attribute")
	defer span.End()

	var req application.CreateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.attributeService.CreateAttribute(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("attribute.id", resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// GetAttribute handles GET /api/v1/attributes/:attribute_id
func (h *Handler) GetAttribute(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.get_attribute")
	defer span.End()

	attributeID := c.Param("attribute_id")
	resp, err := h.attributeService.GetAttribute(ctx, attributeID)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("attribute.id", resp.ID))
	c.JSON(http.StatusOK, resp)
}

// ListAttributesByCategory handles GET /api/v1/attributes/by-category/:category_id
func (h *Handler) ListAttributesByCategory(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.list_attributes_by_category")
	defer span.End()

	categoryID := c.Param("category_id")
	resp, err := h.attributeService.ListAttributesByCategory(ctx, categoryID)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("category.id", categoryID))
	c.JSON(http.StatusOK, resp)
}

// UpdateAttribute handles PUT /api/v1/attributes/:attribute_id
func (h *Handler) UpdateAttribute(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.update_attribute")
	defer span.End()

	attributeID := c.Param("attribute_id")
	var req application.UpdateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.attributeService.UpdateAttribute(ctx, attributeID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("attribute.id", resp.ID))
	c.JSON(http.StatusOK, resp)
}

// DeleteAttribute handles DELETE /api/v1/attributes/:attribute_id
func (h *Handler) DeleteAttribute(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.delete_attribute")
	defer span.End()

	attributeID := c.Param("attribute_id")
	if err := h.attributeService.DeleteAttribute(ctx, attributeID); err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("attribute.id", attributeID))
	c.JSON(http.StatusOK, gin.H{"message": "attribute deleted"})
}

// CreateAttributeValue handles POST /api/v1/attributes/values
func (h *Handler) CreateAttributeValue(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.create_attribute_value")
	defer span.End()

	var req application.CreateAttributeValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.attributeService.CreateAttributeValue(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListAttributeValues handles GET /api/v1/attributes/:attribute_id/values
func (h *Handler) ListAttributeValues(c *gin.Context) {
	ctx, span := otel.Tracer("product-service").Start(c.Request.Context(), "http.list_attribute_values")
	defer span.End()

	attributeID := c.Param("attribute_id")
	resp, err := h.attributeService.ListAttributeValues(ctx, attributeID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// parseProductFilter parses query parameters into a ProductFilter
func parseProductFilter(c *gin.Context) domain.ProductFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	minPrice, _ := strconv.ParseFloat(c.DefaultQuery("min_price", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("max_price", "0"), 64)

	return domain.ProductFilter{
		Page:       page,
		Size:       size,
		CategoryID: c.Query("category_id"),
		SellerID:   c.Query("seller_id"),
		BrandID:    c.Query("brand_id"),
		Status:     c.Query("status"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		SortBy:     domain.SortField(c.DefaultQuery("sort_by", "created_at")),
		SortOrder:  domain.SortDirection(c.DefaultQuery("sort_order", "DESC")),
		Search:     c.Query("q"),
	}
}

// handleError maps domain/application errors to HTTP responses
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check for AppError
	if appErr, ok := err.(*errors.AppError); ok {
		c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
			"error_code": appErr.Code,
			"message":    appErr.Message,
			"details":    appErr.Details,
		})
		return
	}

	// Check for domain errors
	if domain.IsNotFound(err) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check for domain validation errors
	if domain.IsDomainError(err, "INVALID_PRICE") ||
		domain.IsDomainError(err, "INVALID_STOCK") ||
		domain.IsDomainError(err, "INVALID_CATEGORY") {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Check for duplicate errors
	if domain.IsDomainError(err, "DUPLICATE_PRODUCT") ||
		domain.IsDomainError(err, "DUPLICATE_SKU") {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Check for unauthorized errors
	if domain.IsDomainError(err, "UNAUTHORIZED") {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})
}
