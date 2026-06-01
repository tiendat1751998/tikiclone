package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/cart/internal/application"
	"github.com/tikiclone/tiki/services/cart/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type Handler struct {
	service *application.CartService
}

func NewHandler(service *application.CartService) *Handler {
	return &Handler{service: service}
}

// cartResponse maps backend cart data to the frontend Cart type
type cartResponse struct {
	Items      []cartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	Subtotal   float64            `json:"subtotal"`
	Currency   string             `json:"currency"`
}

type cartItemResponse struct {
	ID             string  `json:"id"`
	ProductID      string  `json:"product_id"`
	SkuID          string  `json:"sku_id"`
	ShopID         string  `json:"shop_id"`
	Name           string  `json:"name"`
	ImageURL       string  `json:"image_url"`
	Price          float64 `json:"price"`
	OriginalPrice  float64 `json:"original_price,omitempty"`
	Currency       string  `json:"currency"`
	Quantity       int     `json:"quantity"`
	Stock          int     `json:"stock"`
	SkuName        string  `json:"sku_name"`
	IsSelected     bool    `json:"is_selected"`
	ShopName       string  `json:"shop_name"`
}

func toCartResponse(cart *domain.Cart, items []*domain.CartItem) cartResponse {
	resp := cartResponse{
		Items:      make([]cartItemResponse, 0, len(items)),
		Subtotal:   0,
		Currency:   cart.Currency,
		TotalItems: len(items),
	}
	for _, item := range items {
		cur := cart.Currency
		if cur == "" {
			cur = "VND"
		}
		resp.Items = append(resp.Items, cartItemResponse{
			ID:         item.ID,
			ProductID:  item.SKU,
			SkuID:      item.SKU,
			ShopID:     item.ShopID,
			Name:       item.ProductName,
			ImageURL:   item.ImageURL,
			Price:      float64(item.UnitPrice),
			Quantity:   item.Quantity,
			Stock:      0, // Stock should be fetched from inventory service via gRPC integration
			SkuName:    item.ProductName,
			IsSelected: item.IsSelected,
			ShopName:   item.ShopName,
			Currency:   cur,
		})
		if item.IsSelected {
			resp.Subtotal += float64(item.UnitPrice) * float64(item.Quantity)
		}
	}
	return resp
}

func (h *Handler) getUserID(c *gin.Context) string {
	userID := c.GetString("user_id")
	if userID != "" {
		return userID
	}
	// Fallback: read from gateway-injected header
	userID = c.GetHeader("X-User-ID")
	return userID
}

func (h *Handler) getOrCreateUserCart(c *gin.Context) (*domain.Cart, error) {
	userID := h.getUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "UNAUTHORIZED", "message": "authentication required"})
		return nil, errors.New("unauthorized")
	}
	cart, err := h.service.GetOrCreateCart(c.Request.Context(), userID, "", "VND")
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (h *Handler) GetCart(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.get_cart")
	defer span.End()

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	_, items, err := h.service.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCartResponse(cart, items))
}

func (h *Handler) AddItem(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.add_item")
	defer span.End()

	var req struct {
		ProductID string  `json:"product_id" binding:"required"`
		SkuID     string  `json:"sku_id" binding:"required"`
		Quantity  int     `json:"quantity" binding:"required,min=1"`
		Name      string  `json:"name"`
		Price     float64 `json:"price"`
		ShopID    string  `json:"shop_id"`
		ShopName  string  `json:"shop_name"`
		ImageURL  string  `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	item, err := h.service.AddItem(ctx, cart.ID, application.AddItemRequest{
		SKU:         req.SkuID,
		ProductName: req.Name,
		ShopID:      req.ShopID,
		ShopName:    req.ShopName,
		Quantity:    req.Quantity,
		UnitPrice:   int64(req.Price),
		ImageURL:    req.ImageURL,
		Attributes:  "",
	})
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("item_id", item.ID))

	// Return the updated cart so the frontend can reconcile
	_, updatedItems, err := h.service.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCartResponse(cart, updatedItems))
}

func (h *Handler) UpdateItem(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.update_item")
	defer span.End()

	itemID := c.Param("item_id")

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	if err := h.service.UpdateItemQuantity(ctx, cart.ID, itemID, req.Quantity); err != nil {
		handleError(c, err)
		return
	}

	updatedCart, updatedItems, err := h.service.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCartResponse(updatedCart, updatedItems))
}

func (h *Handler) RemoveItem(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.remove_item")
	defer span.End()

	itemID := c.Param("item_id")

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	if err := h.service.RemoveItem(ctx, cart.ID, itemID); err != nil {
		handleError(c, err)
		return
	}

	updatedCart, updatedItems, err := h.service.GetCartWithItems(ctx, cart.ID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCartResponse(updatedCart, updatedItems))
}

func (h *Handler) ClearCart(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.clear_cart")
	defer span.End()

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	if err := h.service.ClearCart(ctx, cart.ID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCartResponse(cart, nil))
}

func (h *Handler) MergeCarts(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.merge_carts")
	defer span.End()

	userID := h.getUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	var req struct {
		SourceCartID string `json:"source_cart_id" binding:"required"`
		TargetCartID string `json:"target_cart_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	if err := h.service.MergeCarts(ctx, req.SourceCartID, req.TargetCartID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "carts merged"})
}

func (h *Handler) CheckoutPreview(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-cart").Start(c.Request.Context(), "http.checkout_preview")
	defer span.End()

	userID := h.getUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "UNAUTHORIZED", "message": "authentication required"})
		return
	}

	cart, err := h.getOrCreateUserCart(c)
	if err != nil {
		return
	}

	var req struct {
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	preview, err := h.service.PrepareCheckout(ctx, cart.ID, userID, req.IdempotencyKey)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, preview)
}

var errorStatusMap = map[error]int{
	domain.ErrCartNotFound:        http.StatusNotFound,
	domain.ErrCartExpired:         http.StatusGone,
	domain.ErrCartFull:            http.StatusConflict,
	domain.ErrItemNotFound:        http.StatusNotFound,
	domain.ErrInvalidQuantity:     http.StatusBadRequest,
	domain.ErrInvalidCartState:    http.StatusConflict,
	domain.ErrDuplicateItem:       http.StatusConflict,
	domain.ErrMergeConflict:       http.StatusConflict,
	domain.ErrSnapshotNotFound:    http.StatusNotFound,
	domain.ErrIdempotencyConflict: http.StatusConflict,
}

func handleError(c *gin.Context, err error) {
	for domainErr, status := range errorStatusMap {
		if errors.Is(err, domainErr) {
			c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
			return
		}
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
