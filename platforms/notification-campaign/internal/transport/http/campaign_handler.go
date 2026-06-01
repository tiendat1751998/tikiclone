package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/campaign"
)

type createCampaignRequest struct {
	Name            string            `json:"name" binding:"required"`
	Type            string            `json:"type" binding:"required"`
	Channel         string            `json:"channel" binding:"required"`
	Schedule        campaign.Schedule `json:"schedule"`
	AudienceQuery   string            `json:"audience_query"`
	ContentTemplate string            `json:"content_template"`
}

type updateCampaignRequest struct {
	Name            *string            `json:"name,omitempty"`
	Type            *string            `json:"type,omitempty"`
	Channel         *string            `json:"channel,omitempty"`
	Schedule        *campaign.Schedule `json:"schedule,omitempty"`
	AudienceQuery   *string            `json:"audience_query,omitempty"`
	ContentTemplate *string            `json:"content_template,omitempty"`
}

func (h *Handler) CreateCampaign(c *gin.Context) {
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cm, err := h.campaignSvc.Create(c.Request.Context(), &campaign.CreateCampaignRequest{
		Name:            req.Name,
		Type:            campaign.CampaignType(req.Type),
		Channel:         campaign.Channel(req.Channel),
		Schedule:        req.Schedule,
		AudienceQuery:   req.AudienceQuery,
		ContentTemplate: req.ContentTemplate,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cm)
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	campaigns, err := h.campaignSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"campaigns": campaigns})
}

func (h *Handler) GetCampaign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	cm, err := h.campaignSvc.Get(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cm)
}

func (h *Handler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req updateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &campaign.UpdateCampaignRequest{
		Name:            req.Name,
		Schedule:        req.Schedule,
		AudienceQuery:   req.AudienceQuery,
		ContentTemplate: req.ContentTemplate,
	}
	if req.Type != nil {
		t := campaign.CampaignType(*req.Type)
		updateReq.Type = &t
	}
	if req.Channel != nil {
		ch := campaign.Channel(*req.Channel)
		updateReq.Channel = &ch
	}

	cm, err := h.campaignSvc.Update(c.Request.Context(), id, updateReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cm)
}

func (h *Handler) StartCampaign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.campaignSvc.Start(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

func (h *Handler) PauseCampaign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.campaignSvc.Pause(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func (h *Handler) CancelCampaign(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.campaignSvc.Cancel(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}
