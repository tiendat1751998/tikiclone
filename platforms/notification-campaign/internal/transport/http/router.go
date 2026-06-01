package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/health"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.OTelMiddleware("notification-campaign"))
	engine.Use(observability.ObserveHTTPMetrics("notification-campaign"))

	api := engine.Group("/api/v1")

	campaigns := api.Group("/campaigns")
	campaigns.POST("", r.handler.CreateCampaign)
	campaigns.GET("", r.handler.ListCampaigns)
	campaigns.GET("/:id", r.handler.GetCampaign)
	campaigns.PUT("/:id", r.handler.UpdateCampaign)
	campaigns.POST("/:id/start", r.handler.StartCampaign)
	campaigns.POST("/:id/pause", r.handler.PauseCampaign)
	campaigns.POST("/:id/cancel", r.handler.CancelCampaign)

	segments := api.Group("/segments")
	segments.POST("", r.handler.CreateSegment)
	segments.GET("", r.handler.ListSegments)
	segments.POST("/evaluate", r.handler.EvaluateSegment)

	templates := api.Group("/templates")
	templates.POST("", r.handler.CreateTemplate)
	templates.GET("", r.handler.ListTemplates)
	templates.POST("/render", r.handler.RenderTemplate)
	templates.POST("/variants", r.handler.CreateVariant)
	templates.GET("/:id/variants", r.handler.ListVariants)

	delivery := api.Group("/delivery")
	delivery.POST("/optimize-time", r.handler.OptimizeSendTime)
	delivery.POST("/send", r.handler.SendMessage)

	reporting := api.Group("/reporting")
	reporting.POST("/track-open", r.handler.TrackOpen)
	reporting.POST("/track-click", r.handler.TrackClick)
	reporting.GET("/campaigns/:id", r.handler.GetCampaignReport)
	reporting.GET("/aggregated", r.handler.GetAggregatedReport)

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())
}
