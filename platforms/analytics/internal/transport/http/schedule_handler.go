package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/analytics/internal/report_scheduler"
)

type createScheduleRequest struct {
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description,omitempty"`
	Frequency       string   `json:"frequency" binding:"required"`
	DeliveryChannel string   `json:"delivery_channel" binding:"required"`
	Recipients      []string `json:"recipients,omitempty"`
	WebhookURL      string   `json:"webhook_url,omitempty"`
	Format          string   `json:"format"`
	TimeZone        string   `json:"time_zone"`
}

func (h *Handler) CreateSchedule(c *gin.Context) {
	var req createScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "csv"
	}
	if req.TimeZone == "" {
		req.TimeZone = "UTC"
	}

	query := map[string]interface{}{
		"metrics": []string{"total_users", "revenue", "orders"},
	}

	r, err := h.scheduleSvc.ScheduleReport(c.Request.Context(), req.Name, req.Description, query,
		report_scheduler.ScheduleFrequency(req.Frequency),
		report_scheduler.DeliveryChannel(req.DeliveryChannel),
		req.Recipients, req.WebhookURL, req.Format, req.TimeZone, "user-1", "org-1")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, r)
}

func (h *Handler) ListSchedules(c *gin.Context) {
	reports, total, err := h.scheduleSvc.ListReports(c.Request.Context(), "", 0, 50)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reports, "total": total})
}
