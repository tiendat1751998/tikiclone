package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/order/internal/application"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	orderService *application.OrderService
}

func NewHandler(orderService *application.OrderService) *Handler {
	return &Handler{orderService: orderService}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req application.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	req.UserID = uid

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Validate ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	if order.UserID != uid && r != "admin" && r != "seller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}

	// Admin can list all orders, users can only list their own
	queryUserID := c.Query("user_id")
	if queryUserID == "" || r != "admin" {
		queryUserID = uid
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), queryUserID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"items":      orders,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": totalPages,
	})
}

func (h *Handler) GetOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Validate ownership
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	if order.UserID != uid && r != "admin" && r != "seller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": order.ID,
		"status":   order.Status,
		"version":  order.Version,
	})
}

func (h *Handler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}

	cancelReq := &application.CancelOrderRequest{
		OrderID:       orderID,
		Reason:        req.Reason,
		CancelledBy:   uid,
		CancelledType: domain.CancellationTypeUser,
	}

	if r == "admin" {
		cancelReq.CancelledType = domain.CancellationTypeSystem
	}

	order, err := h.orderService.CancelOrder(c.Request.Context(), cancelReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetOrderHistory(c *gin.Context) {
	orderID := c.Param("id")

	// Validate ownership: fetch order first to check user
	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	if order.UserID != uid && r != "admin" && r != "seller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	history, err := h.orderService.GetOrderHistory(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"history":  history,
	})
}

func (h *Handler) GetReconciliationStatus(c *gin.Context) {
	orderID := c.Param("id")

	// Validate ownership
	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	role, _ := c.Get("role")
	r := ""
	if role != nil {
		r, _ = role.(string)
	}
	if order.UserID != uid && r != "admin" && r != "seller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	status, err := h.orderService.GetReconciliationStatus(c.Request.Context(), orderID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":       orderID,
		"reconciliation": status,
	})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case errors.Is(err, domain.ErrOrderNotCancellable):
		c.JSON(http.StatusConflict, gin.H{"error": "order cannot be cancelled"})
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid state transition"})
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInsufficientPermissions):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	case errors.Is(err, domain.ErrConcurrentModification):
		c.JSON(http.StatusConflict, gin.H{"error": "concurrent modification detected, please retry"})
	case errors.Is(err, domain.ErrIdempotencyKeyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate request"})
	case errors.Is(err, domain.ErrOrderExpired):
		c.JSON(http.StatusGone, gin.H{"error": "order expired"})
	case errors.Is(err, domain.ErrOrderAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "order already exists"})
	case errors.Is(err, domain.ErrOrderNotModifiable):
		c.JSON(http.StatusConflict, gin.H{"error": "order cannot be modified"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
