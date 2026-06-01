package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/content"
)

type createTemplateRequest struct {
	Name        string   `json:"name" binding:"required"`
	Channel     string   `json:"channel" binding:"required"`
	Subject     string   `json:"subject" binding:"required"`
	Body        string   `json:"body" binding:"required"`
	Variables   []string `json:"variables"`
	PreviewText string   `json:"preview_text"`
}

type renderTemplateRequest struct {
	TemplateID string                 `json:"template_id" binding:"required"`
	Variables  map[string]interface{} `json:"variables"`
}

type createVariantRequest struct {
	TemplateID        string            `json:"template_id" binding:"required"`
	Name              string            `json:"name" binding:"required"`
	Modifications     map[string]string `json:"modifications"`
	TrafficPercentage int               `json:"traffic_percentage"`
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.contentSvc.CreateTemplate(c.Request.Context(), &content.CreateTemplateRequest{
		Name:        req.Name,
		Channel:     req.Channel,
		Subject:     req.Subject,
		Body:        req.Body,
		Variables:   req.Variables,
		PreviewText: req.PreviewText,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.contentSvc.ListTemplates(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *Handler) RenderTemplate(c *gin.Context) {
	var req renderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject, body, err := h.contentSvc.Render(c.Request.Context(), &content.RenderRequest{
		TemplateID: req.TemplateID,
		Variables:  req.Variables,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subject": subject, "body": body})
}

func (h *Handler) CreateVariant(c *gin.Context) {
	var req createVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v, err := h.contentSvc.CreateVariant(c.Request.Context(), &content.CreateVariantRequest{
		TemplateID:        req.TemplateID,
		Name:              req.Name,
		Modifications:     req.Modifications,
		TrafficPercentage: req.TrafficPercentage,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

func (h *Handler) ListVariants(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "template id is required"})
		return
	}

	variants, err := h.contentSvc.ListVariants(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variants": variants})
}
