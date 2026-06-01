package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/shipments"
)

func (h *Handler) CreateShipment(c *gin.Context) {
	var req struct {
		ID      string           `json:"id"`
		OrderID string           `json:"order_id"`
		CustomerID string        `json:"customer_id"`
		WarehouseID string       `json:"warehouse_id"`
		Origin  shipments.Address `json:"origin_address"`
		Dest    shipments.Address `json:"destination_address"`
		Packages []shipments.Package `json:"packages"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sh := &shipments.Shipment{
		ID:                 req.ID,
		OrderID:            req.OrderID,
		CustomerID:         req.CustomerID,
		WarehouseID:        req.WarehouseID,
		OriginAddress:      req.Origin,
		DestinationAddress: req.Dest,
		Packages:           req.Packages,
	}
	if err := h.shipments.Create(c.Request.Context(), sh); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sh)
}

func (h *Handler) GetShipment(c *gin.Context) {
	id := c.Param("id")
	sh, err := h.shipments.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shipment not found"})
		return
	}
	c.JSON(http.StatusOK, sh)
}

func (h *Handler) ListShipments(c *gin.Context) {
	filter := shipments.ShipmentFilter{
		Status:      shipments.ShipmentStatus(c.Query("status")),
		CourierID:   c.Query("courier_id"),
		CustomerID:  c.Query("customer_id"),
		WarehouseID: c.Query("warehouse_id"),
		Offset:      0,
		Limit:       20,
	}
	shipments, total, err := h.shipments.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shipments, "total": total})
}

func (h *Handler) UpdateShipmentStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status   shipments.ShipmentStatus `json:"status"`
		Reason   string                   `json:"reason"`
		ReplayID string                   `json:"replay_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.shipments.TransitionStatus(c.Request.Context(), id, req.Status, req.Reason, req.ReplayID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *Handler) GetByOrderID(c *gin.Context) {
	orderID := c.Param("orderID")
	shipments, total, err := h.shipments.List(c.Request.Context(), shipments.ShipmentFilter{OrderID: orderID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shipments, "total": total})
}
