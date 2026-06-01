package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/pickups"
)

func (h *Handler) CreatePickup(c *gin.Context) {
	var req struct {
		ID            string  `json:"id"`
		ShipmentID    string  `json:"shipment_id"`
		FulfillmentID string  `json:"fulfillment_id"`
		Address       string  `json:"address"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &pickups.Pickup{
		ID:            req.ID,
		ShipmentID:    req.ShipmentID,
		FulfillmentID: req.FulfillmentID,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
	}
	if err := h.pickups.Create(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) MarkPickupComplete(c *gin.Context) {
	id := c.Param("id")
	if err := h.pickups.MarkCompleted(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pickup completed"})
}

func (h *Handler) MarkPickupFailed(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pickups.MarkFailed(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pickup failed"})
}

func (h *Handler) GetPickupByShipment(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	p, err := h.pickups.GetByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pickup not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}
