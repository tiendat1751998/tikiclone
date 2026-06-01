package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/advertising/internal/analytics"
	"github.com/tikiclone/tiki/platforms/advertising/internal/events"
	"github.com/tikiclone/tiki/platforms/advertising/internal/metrics"
)

type recordImpressionRequest struct {
	CampaignID string  `json:"campaign_id"`
	CreativeID string  `json:"creative_id"`
	UserID     string  `json:"user_id"`
	Cost       float64 `json:"cost"`
	Device     string  `json:"device"`
	Location   string  `json:"location"`
}

type recordClickRequest struct {
	ImpressionID string  `json:"impression_id"`
	CampaignID   string  `json:"campaign_id"`
	CreativeID   string  `json:"creative_id"`
	UserID       string  `json:"user_id"`
	Cost         float64 `json:"cost"`
}

type recordConversionRequest struct {
	ClickID        string  `json:"click_id"`
	CampaignID     string  `json:"campaign_id"`
	CreativeID     string  `json:"creative_id"`
	UserID         string  `json:"user_id"`
	Revenue        float64 `json:"revenue"`
	ConversionType string  `json:"conversion_type"`
}

func (h *Handler) RecordImpression(c *gin.Context) {
	var req recordImpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	imp := &analytics.Impression{
		CampaignID: req.CampaignID,
		CreativeID: req.CreativeID,
		UserID:     req.UserID,
		Cost:       req.Cost,
		Device:     req.Device,
		Location:   req.Location,
	}

	if err := h.analyticsSvc.RecordImpression(c.Request.Context(), imp); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ImpressionsTotal.Inc()
	metrics.SpendTotal.Add(req.Cost)

	h.publisher.Publish(c.Request.Context(), events.EventImpressionRecorded, events.ImpressionRecorded{
		ImpressionID: imp.ID,
		CampaignID:   req.CampaignID,
		CreativeID:   req.CreativeID,
		UserID:       req.UserID,
		Cost:         req.Cost,
		Timestamp:    time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"id": imp.ID, "message": "impression recorded"})
}

func (h *Handler) RecordClick(c *gin.Context) {
	var req recordClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	click := &analytics.Click{
		ImpressionID: req.ImpressionID,
		CampaignID:   req.CampaignID,
		CreativeID:   req.CreativeID,
		UserID:       req.UserID,
		Cost:         req.Cost,
	}

	if err := h.analyticsSvc.RecordClick(c.Request.Context(), click); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ClicksTotal.Inc()
	metrics.SpendTotal.Add(req.Cost)

	h.publisher.Publish(c.Request.Context(), events.EventClickRecorded, events.ClickRecorded{
		ClickID:      click.ID,
		ImpressionID: req.ImpressionID,
		CampaignID:   req.CampaignID,
		CreativeID:   req.CreativeID,
		UserID:       req.UserID,
		Cost:         req.Cost,
		Timestamp:    time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"id": click.ID, "message": "click recorded"})
}

func (h *Handler) RecordConversion(c *gin.Context) {
	var req recordConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	conv := &analytics.Conversion{
		ClickID:        req.ClickID,
		CampaignID:     req.CampaignID,
		CreativeID:     req.CreativeID,
		UserID:         req.UserID,
		Revenue:        req.Revenue,
		ConversionType: req.ConversionType,
	}

	if err := h.analyticsSvc.RecordConversion(c.Request.Context(), conv); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ConversionsTotal.Inc()

	h.publisher.Publish(c.Request.Context(), events.EventConversionRecorded, conv)

	c.JSON(http.StatusOK, gin.H{"id": conv.ID, "message": "conversion recorded"})
}

func (h *Handler) GetAnalyticsReport(c *gin.Context) {
	filter := &analytics.ReportFilter{
		CampaignID: c.Query("campaign_id"),
		CreativeID: c.Query("creative_id"),
		StartDate:  c.Query("start_date"),
		EndDate:    c.Query("end_date"),
	}

	report, err := h.analyticsSvc.GetReport(c.Request.Context(), filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
