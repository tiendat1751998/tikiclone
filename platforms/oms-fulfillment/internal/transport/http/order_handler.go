package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/ordermanagement"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		ID              string                    `json:"id"`
		UserID          string                    `json:"user_id"`
		Items           []ordermanagement.OrderItem `json:"items"`
		ShippingAddress ordermanagement.Address    `json:"shipping_address"`
		BillingAddress  ordermanagement.Address    `json:"billing_address"`
		Notes           string                    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var totalAmount float64
	for i := range req.Items {
		req.Items[i].TotalPrice = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		totalAmount += req.Items[i].TotalPrice
	}
	order := &ordermanagement.Order{
		ID:              req.ID,
		UserID:          req.UserID,
		Items:           req.Items,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Notes:           req.Notes,
		PaymentStatus:   "unpaid",
	}
	if err := h.orders.Create(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orders.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	filter := ordermanagement.OrderFilter{
		Status: ordermanagement.OrderStatus(c.Query("status")),
		UserID: c.Query("user_id"),
	}
	orders, total, err := h.orders.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders, "total": total})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status ordermanagement.OrderStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.orders.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
