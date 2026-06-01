package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/fraud/internal/health"
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
	engine.Use(middleware.OTelMiddleware("tiki-fraud"))
	engine.Use(observability.ObserveHTTPMetrics("tiki-fraud"))

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	v1 := engine.Group("/api/v1/fraud")
	{
		v1.POST("/evaluate", r.handler.Evaluate)
		v1.GET("/alerts", r.handler.ListAlerts)
		v1.PUT("/alerts/:id/resolve", r.handler.ResolveAlert)
	}

	rules := engine.Group("/api/v1/rules")
	{
		rules.POST("", r.handler.CreateRule)
		rules.GET("", r.handler.ListRules)
		rules.PUT("/:id", r.handler.UpdateRule)
		rules.POST("/:id/toggle", r.handler.ToggleRule)
	}

	bl := engine.Group("/api/v1/blacklist")
	{
		bl.POST("/check", r.handler.CheckBlacklist)
		bl.POST("/add", r.handler.AddToBlacklist)
		bl.POST("/remove", r.handler.RemoveFromBlacklist)
	}

	cases := engine.Group("/api/v1/cases")
	{
		cases.POST("", r.handler.CreateCase)
		cases.GET("", r.handler.ListCases)
		cases.PUT("/:id", r.handler.UpdateCase)
	}

	verify := engine.Group("/api/v1/verify")
	{
		verify.POST("/initiate", r.handler.InitiateVerification)
		verify.POST("/check", r.handler.VerifyCode)
	}
}
