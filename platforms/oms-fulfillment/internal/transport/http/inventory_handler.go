package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/inventory"
)

func (h *Handler) ReserveInventory(c *gin.Context) {
	var req struct {
		OrderID   string `json:"order_id"`
		ProductID string `json:"product_id"`
		SKU       string `json:"sku"`
		Quantity  int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.inventory.Reserve(c.Request.Context(), inventory.ReserveRequest{
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
		SKU:       req.SKU,
		Quantity:  req.Quantity,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) ReleaseInventory(c *gin.Context) {
	var req struct {
		ReservationID string `json:"reservation_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.inventory.Release(c.Request.Context(), req.ReservationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reservation released"})
}

func (h *Handler) CheckStock(c *gin.Context) {
	productID := c.Query("product_id")
	if productID == "" {
		stocks, err := h.inventory.ListStock(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": stocks})
		return
	}
	stock, err := h.inventory.GetStock(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)
}
