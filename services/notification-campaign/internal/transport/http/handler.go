package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/application"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	svc *application.NotificationService
}

func NewHandler(svc *application.NotificationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateCampaign(c *gin.Context) {
	var req application.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := h.svc.CreateCampaign(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, campaign)
}

func (h *Handler) GetCampaign(c *gin.Context) {
	campaign, err := h.svc.GetCampaign(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	campaignType := c.Query("campaign_type")
	items, err := h.svc.ListCampaigns(c.Request.Context(), status, campaignType, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) UpdateCampaign(c *gin.Context) {
	var req application.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := h.svc.UpdateCampaign(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *Handler) DeleteCampaign(c *gin.Context) {
	if err := h.svc.DeleteCampaign(c.Request.Context(), c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) LaunchCampaign(c *gin.Context) {
	campaign, err := h.svc.LaunchCampaign(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *Handler) PauseCampaign(c *gin.Context) {
	campaign, err := h.svc.PauseCampaign(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req application.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetTemplate(c *gin.Context) {
	t, err := h.svc.GetTemplate(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	channel := c.Query("channel")
	items, err := h.svc.ListTemplates(c.Request.Context(), campaignID, channel)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	var req application.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.UpdateTemplate(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) AddRecipient(c *gin.Context) {
	var req application.AddRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.AddRecipient(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *Handler) ListRecipients(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	items, err := h.svc.ListRecipients(c.Request.Context(), c.Param("id"), status, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetPreference(c *gin.Context) {
	p, err := h.svc.GetPreference(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdatePreference(c *gin.Context) {
	var req application.UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.UpdatePreference(c.Request.Context(), c.Param("user_id"), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) GetDeliveryLog(c *gin.Context) {
	l, err := h.svc.GetDeliveryLog(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *Handler) ListDeliveryLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	campaignID := c.Query("campaign_id")
	userID := c.Query("user_id")
	items, err := h.svc.ListDeliveryLogs(c.Request.Context(), campaignID, userID, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrCampaignNotFound),
		errors.Is(err, domain.ErrTemplateNotFound),
		errors.Is(err, domain.ErrPreferenceNotFound),
		errors.Is(err, domain.ErrDeliveryLogNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrCampaignNotDraft),
		errors.Is(err, domain.ErrCampaignNotActive):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	default:
		zap.L().Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
