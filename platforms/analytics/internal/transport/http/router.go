package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/analytics/internal/health"
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
	engine.Use(middleware.OTelMiddleware("tiki-analytics"))
	engine.Use(observability.ObserveHTTPMetrics("tiki-analytics"))

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	v1 := engine.Group("/api/v1")
	{
		analytics := v1.Group("/analytics")
		{
			analytics.POST("/query", r.handler.RunQuery)
			analytics.GET("/metrics", r.handler.GetMetrics)
		}

		events := v1.Group("/events")
		{
			events.POST("/ingest", r.handler.IngestEvent)
			events.POST("/batch", r.handler.BatchIngest)
		}

		v1.POST("/funnel/analyze", r.handler.AnalyzeFunnel)

		v1.POST("/cohort/analyze", r.handler.AnalyzeCohort)

		v1.GET("/sessions", r.handler.GetSessions)
	}

	dashboards := v1.Group("/dashboards")
	{
		dashboards.POST("", r.handler.CreateDashboard)
		dashboards.GET("", r.handler.ListDashboards)
		dashboards.GET("/:id", r.handler.GetDashboard)
		dashboards.PUT("/:id", r.handler.UpdateDashboard)
		dashboards.POST("/:id/widgets", r.handler.AddWidget)
	}

	schedules := v1.Group("/schedules")
	{
		schedules.POST("", r.handler.CreateSchedule)
		schedules.GET("", r.handler.ListSchedules)
	}
}
