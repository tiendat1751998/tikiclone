package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/advertising/internal/creative"
)

type createCreativeRequest struct {
	CampaignID     string                    `json:"campaign_id"`
	Name           string                    `json:"name"`
	Format         string                    `json:"format"`
	Content        string                    `json:"content"`
	DestinationURL string                    `json:"destination_url"`
	Sizes          []creative.CreativeSize   `json:"sizes"`
}

type creativeResponse struct {
	ID             string              `json:"id"`
	CampaignID     string              `json:"campaign_id"`
	Name           string              `json:"name"`
	Format         string              `json:"format"`
	Status         string              `json:"status"`
	Content        string              `json:"content"`
	DestinationURL string              `json:"destination_url"`
	Sizes          []creative.CreativeSize `json:"sizes"`
}

func toCreativeResponse(c *creative.Creative) creativeResponse {
	return creativeResponse{
		ID:             c.ID,
		CampaignID:     c.CampaignID,
		Name:           c.Name,
		Format:         string(c.Format),
		Status:         string(c.Status),
		Content:        c.Content,
		DestinationURL: c.DestinationURL,
		Sizes:          c.Sizes,
	}
}

func (h *Handler) CreateCreative(c *gin.Context) {
	var req createCreativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	cr := &creative.Creative{
		CampaignID:     req.CampaignID,
		Name:           req.Name,
		Format:         creative.CreativeFormat(req.Format),
		Content:        req.Content,
		DestinationURL: req.DestinationURL,
		Sizes:          req.Sizes,
	}

	created, err := h.creativeSvc.Create(c.Request.Context(), cr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toCreativeResponse(created))
}

func (h *Handler) ListCreatives(c *gin.Context) {
	status := creative.CreativeStatus(c.Query("status"))
	creatives, err := h.creativeSvc.List(c.Request.Context(), status)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]creativeResponse, 0, len(creatives))
	for _, cr := range creatives {
		responses = append(responses, toCreativeResponse(cr))
	}

	c.JSON(http.StatusOK, responses)
}

func (h *Handler) ApproveCreative(c *gin.Context) {
	id := c.Param("id")
	cr, err := h.creativeSvc.Approve(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toCreativeResponse(cr))
}
