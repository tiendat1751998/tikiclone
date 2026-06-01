package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

type ingestEventRequest struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type" binding:"required"`
	UserID     string                 `json:"user_id" binding:"required"`
	SessionID  string                 `json:"session_id"`
	Timestamp  string                 `json:"timestamp"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Context    struct {
		IP        string `json:"ip,omitempty"`
		UserAgent string `json:"user_agent,omitempty"`
		Device    string `json:"device,omitempty"`
		Referrer  string `json:"referrer,omitempty"`
	} `json:"context,omitempty"`
	Source   string  `json:"source,omitempty"`
	Country  string  `json:"country,omitempty"`
	Device   string  `json:"device,omitempty"`
	Campaign string  `json:"campaign,omitempty"`
	Revenue  float64 `json:"revenue,omitempty"`
}

type batchIngestRequest struct {
	Events []ingestEventRequest `json:"events" binding:"required"`
}

func (h *Handler) IngestEvent(c *gin.Context) {
	var req ingestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ts := time.Now()
	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			ts = t
		}
	}

	event := &events.AnalyticsEvent{
		EventID:   req.EventID,
		EventType: events.EventType(req.EventType),
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Timestamp: ts,
		Properties: req.Properties,
		Source:    req.Source,
		Country:   req.Country,
		Device:    req.Device,
		Campaign:  req.Campaign,
		Revenue:   req.Revenue,
		Context: events.EventContext{
			IP:        req.Context.IP,
			UserAgent: req.Context.UserAgent,
			Device:    req.Context.Device,
			Referrer:  req.Context.Referrer,
		},
	}

	if err := h.eventSvc.IngestEvent(c.Request.Context(), event); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event_id": event.EventID, "status": "ingested"})
}

func (h *Handler) BatchIngest(c *gin.Context) {
	var req batchIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analyticsEvents := make([]events.AnalyticsEvent, len(req.Events))
	for i, e := range req.Events {
		ts := time.Now()
		if e.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
				ts = t
			}
		}
		analyticsEvents[i] = events.AnalyticsEvent{
			EventID:   e.EventID,
			EventType: events.EventType(e.EventType),
			UserID:    e.UserID,
			SessionID: e.SessionID,
			Timestamp: ts,
			Properties: e.Properties,
			Source:    e.Source,
			Country:   e.Country,
			Device:    e.Device,
			Campaign:  e.Campaign,
			Revenue:   e.Revenue,
		}
	}

	ingested, err := h.eventSvc.BatchIngest(c.Request.Context(), analyticsEvents)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ingested": ingested, "total": len(req.Events)})
}
