package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/dispatch"
)

func (h *Handler) CreateDispatch(c *gin.Context) {
	var req struct {
		ID         string `json:"id"`
		ShipmentID string `json:"shipment_id"`
		CourierID  string `json:"courier_id"`
		ZoneID     string `json:"zone_id"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d := &dispatch.Dispatch{
		ID:         req.ID,
		ShipmentID: req.ShipmentID,
		CourierID:  req.CourierID,
		ZoneID:     req.ZoneID,
		Notes:      req.Notes,
	}
	if err := h.dispatch.CreateDispatch(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) AssignDispatchCourier(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CourierID string `json:"courier_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.dispatch.AssignCourier(c.Request.Context(), id, req.CourierID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "courier assigned"})
}

func (h *Handler) MarkDispatchEnRoute(c *gin.Context) {
	id := c.Param("id")
	if err := h.dispatch.MarkEnRoute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "en route"})
}

func (h *Handler) MarkDispatchComplete(c *gin.Context) {
	id := c.Param("id")
	if err := h.dispatch.MarkCompleted(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "completed"})
}

func (h *Handler) GetDispatchByShipment(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	d, err := h.dispatch.GetByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dispatch not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}
