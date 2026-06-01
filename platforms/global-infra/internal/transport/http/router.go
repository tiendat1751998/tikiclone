package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/health"
	httpmiddleware "github.com/tikiclone/tiki/platforms/global-infra/internal/middleware"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery(), middleware.ErrorHandler(), middleware.RequestID(), middleware.CORS(), observability.ObserveHTTPMetrics("global-infra"))

	e.GET("/health", r.health.LivenessHandler())
	e.GET("/ready", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	secrets := e.Group("/api/v1/secrets")
	secrets.Use(httpmiddleware.APIKeyAuth())
	{
		secrets.POST("", r.handler.CreateSecret)
		secrets.GET("", r.handler.ListSecrets)
		secrets.POST("/:id/rotate", r.handler.RotateSecret)
	}
}
