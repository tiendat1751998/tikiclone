package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
)

type createCampaignRequest struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	BidAmount float64         `json:"bid_amount"`
	TargetCPA float64         `json:"target_cpa"`
	Budget    campaign.Budget `json:"budget"`
	DateRange struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"date_range"`
	Targeting campaign.Targeting `json:"targeting"`
}

type campaignResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Type         string            `json:"type"`
	BidAmount    float64           `json:"bid_amount"`
	Budget       campaign.Budget   `json:"budget"`
	DateRange    campaign.DateRange `json:"date_range"`
	Targeting    campaign.Targeting `json:"targeting"`
	QualityScore float64           `json:"quality_score"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func toCampaignResponse(c *campaign.Campaign) campaignResponse {
	return campaignResponse{
		ID:           c.ID,
		Name:         c.Name,
		Status:       string(c.Status),
		Type:         string(c.Type),
		BidAmount:    c.BidAmount,
		Budget:       c.Budget,
		DateRange:    c.DateRange,
		Targeting:    c.Targeting,
		QualityScore: c.QualityScore,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func (h *Handler) CreateCampaign(c *gin.Context) {
	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	cm := &campaign.Campaign{
		Name:      req.Name,
		Type:      campaign.CampaignType(req.Type),
		BidAmount: req.BidAmount,
		TargetCPA: req.TargetCPA,
		Budget:    req.Budget,
		Targeting: req.Targeting,
	}

	if req.DateRange.Start != "" {
		start, err := time.Parse(time.RFC3339, req.DateRange.Start)
		if err == nil {
			cm.DateRange.Start = start
		}
	}
	if req.DateRange.End != "" {
		end, err := time.Parse(time.RFC3339, req.DateRange.End)
		if err == nil {
			cm.DateRange.End = end
		}
	}

	created, err := h.campaignSvc.Create(c.Request.Context(), cm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toCampaignResponse(created))
}

func (h *Handler) ListCampaigns(c *gin.Context) {
	status := campaign.CampaignStatus(c.Query("status"))
	camps, err := h.campaignSvc.List(c.Request.Context(), status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]campaignResponse, 0, len(camps))
	for _, cm := range camps {
		responses = append(responses, toCampaignResponse(cm))
	}

	c.JSON(http.StatusOK, responses)
}

func (h *Handler) GetCampaign(c *gin.Context) {
	id := c.Param("id")
	cm, err := h.campaignSvc.Get(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCampaignResponse(cm))
}

func (h *Handler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")

	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	cm := &campaign.Campaign{
		ID:        id,
		Name:      req.Name,
		Type:      campaign.CampaignType(req.Type),
		BidAmount: req.BidAmount,
		TargetCPA: req.TargetCPA,
		Budget:    req.Budget,
		Targeting: req.Targeting,
	}

	updated, err := h.campaignSvc.Update(c.Request.Context(), cm)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCampaignResponse(updated))
}

func (h *Handler) PauseCampaign(c *gin.Context) {
	id := c.Param("id")
	cm, err := h.campaignSvc.Pause(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCampaignResponse(cm))
}

func (h *Handler) ResumeCampaign(c *gin.Context) {
	id := c.Param("id")
	cm, err := h.campaignSvc.Resume(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCampaignResponse(cm))
}
