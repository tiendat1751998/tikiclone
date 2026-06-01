package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/fulfillment"
)

func (h *Handler) CreateFulfillment(c *gin.Context) {
	var req struct {
		ID          string                `json:"id"`
		ShipmentID  string                `json:"shipment_id"`
		OrderID     string                `json:"order_id"`
		WarehouseID string                `json:"warehouse_id"`
		Items       []fulfillment.FulfillmentItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f := &fulfillment.Fulfillment{
		ID:          req.ID,
		ShipmentID:  req.ShipmentID,
		OrderID:     req.OrderID,
		WarehouseID: req.WarehouseID,
		Items:       req.Items,
	}
	if err := h.fulfillment.Create(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, f)
}

func (h *Handler) MarkFulfillmentPacked(c *gin.Context) {
	id := c.Param("id")
	if err := h.fulfillment.MarkPacked(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "packed"})
}

func (h *Handler) MarkFulfillmentShipped(c *gin.Context) {
	id := c.Param("id")
	if err := h.fulfillment.MarkShipped(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "shipped"})
}

func (h *Handler) GetFulfillmentByShipment(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	f, err := h.fulfillment.GetByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fulfillment not found"})
		return
	}
	c.JSON(http.StatusOK, f)
}
