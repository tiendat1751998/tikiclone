package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/developer/internal/cicd"
)

type Router struct {
	handler *Handler
}

func NewRouter(h *Handler) *Router {
	return &Router{handler: h}
}

func (r *Router) Setup(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	{
		apiKeys := v1.Group("/api-keys")
		{
			apiKeys.POST("", r.handler.GenerateAPIKey)
			apiKeys.GET("", r.handler.ListAPIKeys)
			apiKeys.POST("/validate", r.handler.ValidateAPIKey)
			apiKeys.POST("/:id/revoke", r.handler.RevokeAPIKey)
		}

		docs := v1.Group("/docs")
		{
			docs.POST("", r.handler.CreateDoc)
			docs.GET("", r.handler.ListDocs)
			docs.GET("/search", r.handler.SearchDocs)
			docs.PUT("/:id", r.handler.UpdateDoc)
		}

		sdks := v1.Group("/sdks")
		{
			sdks.POST("", r.handler.RegisterSDK)
			sdks.GET("", r.handler.ListSDKs)
			sdks.POST("/:id/latest", r.handler.MarkSDKLatest)
		}

		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("", r.handler.RegisterWebhook)
			webhooks.GET("", r.handler.ListWebhooks)
			webhooks.PUT("/:id", r.handler.UpdateWebhook)
			webhooks.POST("/trigger", r.handler.TriggerWebhook)
			webhooks.GET("/deliveries", r.handler.ListDeliveries)
		}

		pipelines := v1.Group("/pipelines")
		{
			pipelines.POST("", r.handler.CreatePipeline)
			pipelines.POST("/:id/trigger", r.handler.TriggerPipeline)
			pipelines.GET("", r.handler.ListPipelines)
			pipelines.GET("/:id", r.handler.GetPipelineStatus)
		}

		onboarding := v1.Group("/onboarding")
		{
			onboarding.GET("/templates", r.handler.ListTemplates)
			onboarding.POST("/tasks/:id/complete", r.handler.CompleteTask)
			onboarding.GET("/progress", r.handler.GetProgress)
		}
	}
}

func (h *Handler) GenerateAPIKey(c *gin.Context) {
	var req struct {
		Name         string   `json:"name" binding:"required"`
		Permissions  []string `json:"permissions"`
		ServiceName  string   `json:"service_name" binding:"required"`
		ExpiresIn    string   `json:"expires_in"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	if req.ExpiresIn != "" {
		if d, err := time.ParseDuration(req.ExpiresIn); err == nil {
			expiresAt = time.Now().Add(d)
		}
	}
	key, rawKey, err := h.apikeysSvc.Generate(c.Request.Context(), req.Name, req.Permissions, req.ServiceName, expiresAt)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_key": key, "raw_key": rawKey})
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	keys, err := h.apikeysSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *Handler) ValidateAPIKey(c *gin.Context) {
	var req struct {
		Key string `json:"key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key, valid := h.apikeysSvc.Validate(c.Request.Context(), req.Key)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true, "api_key": key})
}

func (h *Handler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.apikeysSvc.Revoke(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

func (h *Handler) CreateDoc(c *gin.Context) {
	var req struct {
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Service  string   `json:"service"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Version  string   `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := h.docsSvc.Create(c.Request.Context(), req.Title, req.Content, req.Service, req.Category, req.Tags, req.Version)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) ListDocs(c *gin.Context) {
	service := c.Query("service")
	category := c.Query("category")
	docs, err := h.docsSvc.List(c.Request.Context(), service, category)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"docs": docs})
}

func (h *Handler) SearchDocs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}
	docs, err := h.docsSvc.Search(c.Request.Context(), query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"docs": docs})
}

func (h *Handler) UpdateDoc(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Service  string   `json:"service"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Version  string   `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, err := h.docsSvc.Update(c.Request.Context(), id, req.Title, req.Content, req.Service, req.Category, req.Tags, req.Version)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doc == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "doc not found"})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) RegisterSDK(c *gin.Context) {
	var req struct {
		Name             string `json:"name" binding:"required"`
		Language         string `json:"language" binding:"required"`
		Version          string `json:"version" binding:"required"`
		RepositoryURL    string `json:"repository_url"`
		DocumentationURL string `json:"documentation_url"`
		Compatibility    string `json:"compatibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sdk, err := h.sdkSvc.Register(c.Request.Context(), req.Name, req.Language, req.Version, req.RepositoryURL, req.DocumentationURL, req.Compatibility)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sdk)
}

func (h *Handler) ListSDKs(c *gin.Context) {
	language := c.Query("language")
	sdks, err := h.sdkSvc.List(c.Request.Context(), language)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sdks": sdks})
}

func (h *Handler) MarkSDKLatest(c *gin.Context) {
	id := c.Param("id")
	sdk, err := h.sdkSvc.MarkLatest(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sdk == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "sdk not found"})
		return
	}
	c.JSON(http.StatusOK, sdk)
}

func (h *Handler) RegisterWebhook(c *gin.Context) {
	var req struct {
		Name           string   `json:"name" binding:"required"`
		URL            string   `json:"url" binding:"required"`
		Secret         string   `json:"secret"`
		Events         []string `json:"events" binding:"required"`
		RetryCount     int      `json:"retry_count"`
		TimeoutSeconds int      `json:"timeout_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.webhookSvc.Register(c.Request.Context(), req.Name, req.URL, req.Secret, req.Events, req.RetryCount, req.TimeoutSeconds)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *Handler) ListWebhooks(c *gin.Context) {
	webhooks, err := h.webhookSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}

func (h *Handler) UpdateWebhook(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name           *string  `json:"name"`
		URL            *string  `json:"url"`
		Secret         *string  `json:"secret"`
		Events         []string `json:"events"`
		IsActive       *bool    `json:"is_active"`
		RetryCount     *int     `json:"retry_count"`
		TimeoutSeconds *int     `json:"timeout_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.webhookSvc.Update(c.Request.Context(), id,
		strPtrOrEmpty(req.Name), strPtrOrEmpty(req.URL), strPtrOrEmpty(req.Secret),
		req.Events, req.IsActive, req.RetryCount, req.TimeoutSeconds)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if w == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func strPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *Handler) TriggerWebhook(c *gin.Context) {
	var req struct {
		Event   string      `json:"event" binding:"required"`
		Payload interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deliveries, err := h.webhookSvc.TriggerEvent(c.Request.Context(), req.Event, req.Payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deliveries": deliveries})
}

func (h *Handler) ListDeliveries(c *gin.Context) {
	webhookID := c.Query("webhook_id")
	deliveries, err := h.webhookSvc.ListDeliveries(c.Request.Context(), webhookID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deliveries": deliveries})
}

func (h *Handler) CreatePipeline(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Service   string `json:"service" binding:"required"`
		Trigger   string `json:"trigger" binding:"required"`
		CommitSHA string `json:"commit_sha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trigger := cicd.TriggerType(req.Trigger)
	p, err := h.cicdSvc.Create(c.Request.Context(), req.Name, req.Service, trigger, req.CommitSHA)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) TriggerPipeline(c *gin.Context) {
	id := c.Param("id")
	p, err := h.cicdSvc.Trigger(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPipelines(c *gin.Context) {
	pipelines, err := h.cicdSvc.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pipelines": pipelines})
}

func (h *Handler) GetPipelineStatus(c *gin.Context) {
	id := c.Param("id")
	p, err := h.cicdSvc.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.onboardingSvc.ListTemplates(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *Handler) CompleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.onboardingSvc.CompleteTask(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (h *Handler) GetProgress(c *gin.Context) {
	progress, err := h.onboardingSvc.GetProgress(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *Handler) RotateAPIKey(c *gin.Context) {
	id := c.Param("id")
	key, rawKey, err := h.apikeysSvc.Rotate(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if key == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "api key not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_key": key, "raw_key": rawKey})
}

func parseInt32(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}

func (h *Handler) DeleteWebhook(c *gin.Context) {
	id := c.Param("id")
	if err := h.webhookSvc.Delete(c.Request.Context(), id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
