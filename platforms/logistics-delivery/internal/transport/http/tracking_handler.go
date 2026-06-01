package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/tracking"
)

func (h *Handler) AppendTrackingEvent(c *gin.Context) {
	var req struct {
		ID          string                 `json:"id"`
		ShipmentID  string                 `json:"shipment_id"`
		EventType   tracking.TrackingEventType `json:"event_type"`
		Location    tracking.Location       `json:"location"`
		Description string                 `json:"description"`
		CourierData map[string]any         `json:"courier_data"`
		ReplayID    string                 `json:"replay_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event := &tracking.TrackingEvent{
		ID:          req.ID,
		ShipmentID:  req.ShipmentID,
		EventType:   req.EventType,
		Location:    req.Location,
		Description: req.Description,
		CourierData: req.CourierData,
		ReplayID:    req.ReplayID,
	}
	if err := h.tracking.AppendEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *Handler) GetTrackingTimeline(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	timeline, err := h.tracking.GetTimeline(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
		return
	}
	c.JSON(http.StatusOK, timeline)
}

func (h *Handler) GetLastTrackingEvent(c *gin.Context) {
	shipmentID := c.Param("shipmentID")
	event, err := h.tracking.GetLastEvent(c.Request.Context(), shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no events found"})
		return
	}
	c.JSON(http.StatusOK, event)
}
