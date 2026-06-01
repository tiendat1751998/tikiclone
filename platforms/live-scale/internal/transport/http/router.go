package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/health"
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
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS())
	engine.Use(middleware.OTelMiddleware("tiki-live-scale"))
	engine.Use(observability.ObserveHTTPMetrics("tiki-live-scale"))

	engine.GET("/healthz", r.health.LivenessHandler())
	engine.GET("/readyz", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	v1 := engine.Group("/api/v1")
	{
		sfu := v1.Group("/sfu")
		{
			sfu.POST("/register", r.handler.RegisterSFUNode)
			sfu.GET("/optimal", r.handler.GetOptimalSFUNode)
		}

		cdn := v1.Group("/cdn")
		{
			cdn.POST("/purge", r.handler.PurgeCDNCache)
		}

		cluster := v1.Group("/cluster")
		{
			cluster.POST("/nodes", r.handler.RegisterWSNode)
			cluster.POST("/assign", r.handler.AssignRoom)
			cluster.POST("/broadcast", r.handler.BroadcastMessage)
		}

		streams := v1.Group("/streams")
		{
			streams.POST("/health", r.handler.ReportStreamHealth)
			streams.GET("/:id/health", r.handler.GetStreamHealth)
		}

		v1.GET("/region/nearest", r.handler.GetNearestRegion)

		transcode := v1.Group("/transcode")
		{
			transcode.POST("", r.handler.CreateTranscodeJob)
			transcode.GET("/:id", r.handler.GetTranscodeJob)
		}
	}
}
