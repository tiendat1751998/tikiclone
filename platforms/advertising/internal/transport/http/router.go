package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/advertising/internal/health"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery(), middleware.ErrorHandler(), middleware.RequestID(), middleware.CORS(), middleware.OTelMiddleware("tiki-advertising"), observability.ObserveHTTPMetrics("tiki-advertising"))

	e.GET("/health", r.health.LivenessHandler())
	e.GET("/ready", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	api := e.Group("/api/v1")
	{
		api.POST("/campaigns", r.handler.CreateCampaign)
		api.GET("/campaigns", r.handler.ListCampaigns)
		api.GET("/campaigns/:id", r.handler.GetCampaign)
		api.PUT("/campaigns/:id", r.handler.UpdateCampaign)
		api.POST("/campaigns/:id/pause", r.handler.PauseCampaign)
		api.POST("/campaigns/:id/resume", r.handler.ResumeCampaign)

		api.POST("/auction", r.handler.RunAuction)

		api.POST("/creatives", r.handler.CreateCreative)
		api.GET("/creatives", r.handler.ListCreatives)
		api.PUT("/creatives/:id/approve", r.handler.ApproveCreative)

		api.POST("/analytics/impression", r.handler.RecordImpression)
		api.POST("/analytics/click", r.handler.RecordClick)
		api.POST("/analytics/conversion", r.handler.RecordConversion)
		api.GET("/analytics/report", r.handler.GetAnalyticsReport)
	}
}
