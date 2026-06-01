package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/search/internal/health"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery(), middleware.ErrorHandler(), middleware.RequestID(), middleware.CORS(), middleware.OTelMiddleware("tiki-search"), observability.ObserveHTTPMetrics("tiki-search"))

	e.GET("/health", r.health.LivenessHandler())
	e.GET("/ready", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	api := e.Group("/api/v1")
	{
		api.POST("/search", r.handler.Search)
		api.POST("/search/faceted", r.handler.FacetedSearch)
		api.GET("/autocomplete", r.handler.Autocomplete)
		api.POST("/index/documents", r.handler.IndexDocument)
		api.POST("/index/bulk", r.handler.BulkIndex)
		api.GET("/index/tasks", r.handler.ListIndexTasks)
	}
}
