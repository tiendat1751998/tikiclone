package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/routing"
)

func (h *Handler) AssignRoute(c *gin.Context) {
	var req struct {
		ShipmentID string `json:"shipment_id"`
		City       string `json:"city"`
		State      string `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assignment, err := h.routing.AssignWarehouse(c.Request.Context(), req.ShipmentID, req.City, req.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *Handler) OptimizeWaypoints(c *gin.Context) {
	var req struct {
		Waypoints []routing.Waypoint `json:"waypoints"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	optimized, err := h.routing.OptimizeWaypoints(c.Request.Context(), req.Waypoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"waypoints": optimized})
}

func (h *Handler) GetShipmentRoutes(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	routes, err := h.routing.GetRoutesByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no routes found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": routes})
}
