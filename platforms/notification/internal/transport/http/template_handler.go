package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/notification/internal/template"
)

type createTemplateRequest struct {
	Name      string   `json:"name" binding:"required"`
	Subject   string   `json:"subject" binding:"required"`
	Body      string   `json:"body" binding:"required"`
	Variables []string `json:"variables"`
}

type updateTemplateRequest struct {
	Subject   *string   `json:"subject,omitempty"`
	Body      *string   `json:"body,omitempty"`
	Variables *[]string `json:"variables,omitempty"`
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.template.CreateTemplate(c.Request.Context(), &template.CreateTemplateRequest{
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: req.Variables,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.template.ListTemplates(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	tmpl, err := h.template.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &template.UpdateTemplateRequest{
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: req.Variables,
	}

	tmpl, err := h.template.UpdateTemplate(c.Request.Context(), id, updateReq)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h *Handler) ListTemplateVersions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	versions, err := h.template.ListVersions(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}
