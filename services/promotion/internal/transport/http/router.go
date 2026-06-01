package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

type Router struct {
	handler *Handler
	health  *health.Checker
}

func NewRouter(handler *Handler, healthChecker *health.Checker) *Router {
	return &Router{handler: handler, health: healthChecker}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(
		middleware.Recovery(), middleware.ErrorHandler(),
		middleware.RequestID(), middleware.CORS(),
		middleware.OTelMiddleware("tiki-promotion"),
		observability.ObserveHTTPMetrics("tiki-promotion"),
	)

	engine.GET("/health", r.health.LivenessHandler())
	engine.GET("/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	api := engine.Group("/api/v1")
	{
		api.POST("/vouchers/validate", r.handler.ValidateVoucher)
		api.POST("/vouchers/redeem", r.handler.RedeemVoucher)
		api.POST("/promotions/evaluate", r.handler.EvaluatePromotions)
		api.GET("/campaigns", r.handler.GetActiveCampaigns)
		api.POST("/vouchers", r.handler.CreateVoucher)
	}
}
