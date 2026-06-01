package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/pickpack"
)

func (h *Handler) CreatePickList(c *gin.Context) {
	var req struct {
		ID          string           `json:"id"`
		OrderID     string           `json:"order_id"`
		Items       []pickpack.PickItem `json:"items"`
		WarehouseID string           `json:"warehouse_id"`
		AssignedTo  string           `json:"assigned_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pl := &pickpack.PickList{
		ID:          req.ID,
		OrderID:     req.OrderID,
		Items:       req.Items,
		WarehouseID: req.WarehouseID,
		AssignedTo:  req.AssignedTo,
	}
	if err := h.pickpack.CreatePickList(c.Request.Context(), pl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pl)
}

func (h *Handler) CompletePicking(c *gin.Context) {
	id := c.Param("id")
	if err := h.pickpack.CompletePick(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "picking completed"})
}

func (h *Handler) CreatePacking(c *gin.Context) {
	var req struct {
		ID         string  `json:"id"`
		PickListID string  `json:"pick_list_id"`
		PackageID  string  `json:"package_id"`
		Weight     float64 `json:"weight"`
		Dimensions string  `json:"dimensions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &pickpack.Packing{
		ID:         req.ID,
		PickListID: req.PickListID,
		PackageID:  req.PackageID,
		Weight:     req.Weight,
		Dimensions: req.Dimensions,
	}
	if err := h.pickpack.CreatePacking(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) CreateShipment(c *gin.Context) {
	var req struct {
		ID             string `json:"id"`
		PackingID      string `json:"packing_id"`
		Carrier        string `json:"carrier"`
		TrackingNumber string `json:"tracking_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sh := &pickpack.Shipment{
		ID:             req.ID,
		PackingID:      req.PackingID,
		Carrier:        req.Carrier,
		TrackingNumber: req.TrackingNumber,
	}
	if err := h.pickpack.CreateShipment(c.Request.Context(), sh); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sh)
}
