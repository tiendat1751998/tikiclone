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

func NewRouter(h *Handler, hc *health.Checker) *Router {
	return &Router{handler: h, health: hc}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery(), middleware.ErrorHandler(), middleware.RequestID(), middleware.CORS(), observability.ObserveHTTPMetrics("tiki-sre"))

	e.GET("/health", r.health.LivenessHandler())
	e.GET("/ready", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())

	api := e.Group("/api/v1")
	{
		api.POST("/incidents", r.handler.CreateIncident)
		api.GET("/incidents", r.handler.ListIncidents)
		api.PUT("/incidents/:id/ack", r.handler.AcknowledgeIncident)
		api.PUT("/incidents/:id/resolve", r.handler.ResolveIncident)

		api.POST("/alerts/rules", r.handler.CreateAlertRule)
		api.GET("/alerts/rules", r.handler.ListAlertRules)
		api.POST("/alerts/evaluate", r.handler.EvaluateAlerts)
		api.GET("/alerts", r.handler.ListAlerts)

		api.POST("/healthchecks", r.handler.CreateHealthCheck)
		api.POST("/healthchecks/run", r.handler.RunHealthChecks)
		api.GET("/healthchecks/results", r.handler.GetHealthCheckResults)

		api.POST("/slos", r.handler.CreateSLO)
		api.GET("/slos", r.handler.ListSLOs)
		api.GET("/slos/:id/report", r.handler.GetSLOReport)

		api.POST("/deployments", r.handler.CreateDeployment)
		api.GET("/deployments", r.handler.ListDeployments)
		api.POST("/deployments/:id/approve", r.handler.ApproveDeployment)
		api.POST("/deployments/:id/rollback", r.handler.RollbackDeployment)

		api.POST("/runbooks", r.handler.CreateRunbook)
		api.GET("/runbooks", r.handler.ListRunbooks)
	}
}
