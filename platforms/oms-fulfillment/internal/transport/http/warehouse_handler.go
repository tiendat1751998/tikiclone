package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/warehouse"
)

func (h *Handler) ListWarehouses(c *gin.Context) {
	warehouses, err := h.warehouse.ListWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": warehouses})
}

func (h *Handler) RecordMovement(c *gin.Context) {
	var req struct {
		ID          string                 `json:"id"`
		ProductID   string                 `json:"product_id"`
		WarehouseID string                 `json:"warehouse_id"`
		FromZone    string                 `json:"from_zone"`
		ToZone      string                 `json:"to_zone"`
		Quantity    int                    `json:"quantity"`
		Type        warehouse.MovementType `json:"type"`
		Reference   string                 `json:"reference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &warehouse.InventoryMovement{
		ID:          req.ID,
		ProductID:   req.ProductID,
		WarehouseID: req.WarehouseID,
		FromZone:    req.FromZone,
		ToZone:      req.ToZone,
		Quantity:    req.Quantity,
		Type:        req.Type,
		Reference:   req.Reference,
	}
	if err := h.warehouse.RecordMovement(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}
