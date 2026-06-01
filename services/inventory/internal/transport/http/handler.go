package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/inventory/internal/application"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	inventoryService *application.InventoryService
}

func NewHandler(svc *application.InventoryService) *Handler {
	return &Handler{inventoryService: svc}
}

// ReserveStock handles stock reservation requests.
// [SECURITY] user_id is extracted from JWT context (set by auth middleware), NOT from request body.
// This prevents users from reserving stock on behalf of other users.
func (h *Handler) ReserveStock(c *gin.Context) {
	var req application.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// [SECURITY] Get user_id from JWT context, not from request body
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}
	req.UserID = userIDStr

	reservation, err := h.inventoryService.ReserveStock(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reservation)
}

func (h *Handler) ReleaseStock(c *gin.Context) {
	reservationID := c.Param("id")
	if err := h.inventoryService.ReleaseStock(c.Request.Context(), reservationID); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "released"})
}

func (h *Handler) GetStock(c *gin.Context) {
	skuID := c.Param("sku_id")
	warehouseID := c.Query("warehouse_id")
	stock, err := h.inventoryService.GetStock(c.Request.Context(), skuID, warehouseID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, stock)
}

// handleError maps domain errors to HTTP status codes.
// [SECURITY] Never leaks internal error details to clients.
func handleError(c *gin.Context, err error) {
	switch {
	case err == domain.ErrStockNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
	case err == domain.ErrInsufficientStock:
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})
	case err == domain.ErrReservationNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
	case err == domain.ErrOversellPrevented:
		c.JSON(http.StatusConflict, gin.H{"error": "oversell prevented"})
	case err == domain.ErrUnauthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	case err == domain.ErrConcurrentModification:
		c.JSON(http.StatusConflict, gin.H{"error": "concurrent modification, please retry"})
	case err == domain.ErrReservationExpired:
		c.JSON(http.StatusGone, gin.H{"error": "reservation expired"})
	case err == domain.ErrReservationExists:
		c.JSON(http.StatusConflict, gin.H{"error": "reservation already exists"})
	case err == domain.ErrInvalidStockOperation:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock operation"})
	case err == domain.ErrWarehouseNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "warehouse not found"})
	case err == domain.ErrIdempotencyKeyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate request"})
	default:
		// [SECURITY] Log full error internally, return generic message to client
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
