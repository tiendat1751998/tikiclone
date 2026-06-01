package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/estimations"
)

func (h *Handler) CalculateEstimation(c *gin.Context) {
	var req estimations.EstimationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	est, err := h.estimations.Calculate(c.Request.Context(), &req, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, est)
}

func (h *Handler) GetEstimationByShipment(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	est, err := h.estimations.GetByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "estimation not found"})
		return
	}
	c.JSON(http.StatusOK, est)
}
