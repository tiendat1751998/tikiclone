package http
import ("net/http"; "github.com/gin-gonic/gin"; "github.com/tikiclone/tiki/services/product-catalog/internal/application"; "github.com/tikiclone/tiki/services/product-catalog/internal/domain"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.opentelemetry.io/otel"; "go.uber.org/zap")

type Handler struct { service *application.CatalogService }
func NewHandler(s *application.CatalogService) *Handler { return &Handler{service: s} }

func (h *Handler) CreateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.create_product"); defer span.End()
	var req struct { ShopID string `json:"shop_id" binding:"required"`; Name string `json:"name" binding:"required"`; Description string `json:"description"`; CategoryID string `json:"category_id" binding:"required"`; Currency string `json:"currency" binding:"required"`; IdempotencyKey string `json:"idempotency_key"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	p, err := h.service.CreateProduct(ctx, req.ShopID, req.Name, req.Description, req.CategoryID, req.Currency, req.IdempotencyKey)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetProduct(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.get_product"); defer span.End()
	p, err := h.service.GetProduct(ctx, c.Param("id"))
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.update_product"); defer span.End()
	var req struct { Name string `json:"name"`; Description string `json:"description"`; CategoryID string `json:"category_id"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	if err := h.service.UpdateProduct(ctx, c.Param("id"), req.Name, req.Description, req.CategoryID); err != nil { handleError(c, err); return }
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func (h *Handler) ArchiveProduct(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.archive_product"); defer span.End()
	if err := h.service.ArchiveProduct(ctx, c.Param("id")); err != nil { handleError(c, err); return }
	c.JSON(http.StatusOK, gin.H{"message": "product archived"})
}

func (h *Handler) GetCategories(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.get_categories"); defer span.End()
	cats, err := h.service.GetCategoryTree(ctx)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.create_category"); defer span.End()
	var req struct { ParentID string `json:"parent_id"`; Name string `json:"name" binding:"required"`; Slug string `json:"slug" binding:"required"`; Level int `json:"level" binding:"required"`; SortOrder int `json:"sort_order"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	cat, err := h.service.CreateCategory(ctx, req.ParentID, req.Name, req.Slug, req.Level, req.SortOrder)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusCreated, cat)
}

func (h *Handler) AddSKU(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-catalog").Start(c.Request.Context(), "http.add_sku"); defer span.End()
	var req struct { SKUCode string `json:"sku_code" binding:"required"`; Name string `json:"name"`; Currency string `json:"currency"`; Price int64 `json:"price" binding:"required,min=0"` }
	if err := c.ShouldBindJSON(&req); err != nil { c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()}); return }
	sku, err := h.service.AddSKU(ctx, c.Param("product_id"), req.SKUCode, req.Name, req.Currency, req.Price)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusCreated, sku)
}

var errorStatusMap = map[error]int{domain.ErrProductNotFound: http.StatusNotFound, domain.ErrSKUNotFound: http.StatusNotFound, domain.ErrCategoryNotFound: http.StatusNotFound, domain.ErrInvalidState: http.StatusConflict, domain.ErrDuplicateSKU: http.StatusConflict}

func handleError(c *gin.Context, err error) {
	for e, s := range errorStatusMap {
		if err.Error() == e.Error() || (len(err.Error()) >= len(e.Error()) && err.Error()[:len(e.Error())] == e.Error()) {
			c.AbortWithStatusJSON(s, gin.H{"error_code": e.Error(), "message": err.Error()}); return
		}
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
