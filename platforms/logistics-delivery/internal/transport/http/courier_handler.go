package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/couriers"
)

func (h *Handler) CreateCourier(c *gin.Context) {
	var req struct {
		ID          string                `json:"id"`
		Name        string                `json:"name"`
		Phone       string                `json:"phone"`
		Provider    couriers.CourierProvider `json:"provider"`
		ZoneID      string                `json:"zone_id"`
		MaxCapacity int                   `json:"max_capacity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cr := &couriers.Courier{
		ID:          req.ID,
		Name:        req.Name,
		Phone:       req.Phone,
		Provider:    req.Provider,
		ZoneID:      req.ZoneID,
		MaxCapacity: req.MaxCapacity,
		Status:      couriers.CourierAvailable,
		IsActive:    true,
	}
	if err := h.couriers.Create(c.Request.Context(), cr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cr)
}

func (h *Handler) GetCourier(c *gin.Context) {
	id := c.Param("id")
	cr, err := h.couriers.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "courier not found"})
		return
	}
	c.JSON(http.StatusOK, cr)
}

func (h *Handler) ProcessCourierWebhook(c *gin.Context) {
	var payload couriers.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.couriers.ProcessWebhook(c.Request.Context(), &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
}

func (h *Handler) UpdateCourierLocation(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.couriers.UpdateLocation(c.Request.Context(), id, req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "location updated"})
}
