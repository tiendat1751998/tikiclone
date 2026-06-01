package http

import (
	"net/http"

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

func dashboardAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "UNAUTHORIZED",
				"message":    "authentication required",
			})
			return
		}
		c.Set("user_id", userID)
		c.Set("role", c.GetHeader("X-User-Roles"))
		c.Set("email", c.GetHeader("X-User-Email"))
		c.Next()
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(
		middleware.Recovery(),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.CORS(),
		middleware.OTelMiddleware("production-dashboard"),
		observability.ObserveHTTPMetrics("production-dashboard"),
	)

	engine.GET("/health", r.health.LivenessHandler())
	engine.GET("/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	api := engine.Group("/api/v1")
	{
		// Public dashboard summary
		api.GET("/summary", r.handler.GetSummary)

		// Protected routes
		authorized := api.Group("")
		authorized.Use(dashboardAuth())
		{
			// Service health
			authorized.GET("/services", r.handler.ListServices)
			authorized.POST("/services", r.handler.RegisterService)
			authorized.GET("/services/:service_id", r.handler.GetServiceHealth)
			authorized.PATCH("/services/:service_id/health", r.handler.UpdateServiceHealth)

			// Deployments
			authorized.GET("/deployments", r.handler.ListDeployments)
			authorized.GET("/deployments/active", r.handler.ListActiveDeployments)
			authorized.POST("/deployments", r.handler.CreateDeployment)
			authorized.GET("/deployments/:deployment_id", r.handler.GetDeployment)
			authorized.PATCH("/deployments/:deployment_id", r.handler.UpdateDeployment)

			// Incidents
			authorized.GET("/incidents/active", r.handler.ListActiveIncidents)
			authorized.GET("/incidents/recent", r.handler.ListRecentIncidents)
			authorized.POST("/incidents", r.handler.CreateIncident)
			authorized.GET("/incidents/:incident_id", r.handler.GetIncident)
			authorized.POST("/incidents/:incident_id/acknowledge", r.handler.AcknowledgeIncident)
			authorized.POST("/incidents/:incident_id/resolve", r.handler.ResolveIncident)
			authorized.POST("/incidents/:incident_id/close", r.handler.CloseIncident)

			// Alert rules
			authorized.GET("/alert-rules", r.handler.ListAlertRules)
			authorized.POST("/alert-rules", r.handler.CreateAlertRule)
			authorized.GET("/alert-rules/:rule_id", r.handler.GetAlertRule)
			authorized.PATCH("/alert-rules/:rule_id", r.handler.UpdateAlertRule)
			authorized.DELETE("/alert-rules/:rule_id", r.handler.DeleteAlertRule)

			// Audit logs
			authorized.GET("/audit-logs", r.handler.ListAuditLogs)

			// Service dependencies
			authorized.GET("/dependencies", r.handler.GetAllDependencies)
			authorized.GET("/dependencies/:service_name", r.handler.GetServiceDependencies)
			authorized.POST("/dependencies", r.handler.AddServiceDependency)

			// Capacity metrics
			authorized.GET("/capacity", r.handler.GetLatestCapacity)
			authorized.GET("/capacity/:service_name", r.handler.GetCapacityMetrics)
			authorized.POST("/capacity", r.handler.RecordCapacityMetric)
		}
	}
}
