package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/auth"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/product-catalog/internal/health"
)

type Router struct {
	handler   *Handler
	health    *health.Checker
	jwtSecret string
}

func NewRouter(h *Handler, hc *health.Checker, jwtSecret string) *Router {
	return &Router{handler: h, health: hc, jwtSecret: jwtSecret}
}

func (r *Router) Setup(e *gin.Engine) {
	e.Use(middleware.Recovery(), middleware.ErrorHandler(), middleware.RequestID(), middleware.CORS(), middleware.OTelMiddleware("tiki-product-catalog"), observability.ObserveHTTPMetrics("tiki-product-catalog"))
	e.GET("/health", r.health.LivenessHandler())
	e.GET("/ready", r.health.ReadinessHandler())
	e.GET("/metrics", observability.MetricsHandler())
	api := e.Group("/api/v1")
	if r.jwtSecret != "" {
		api.Use(auth.GinJWTAuth(r.jwtSecret))
	}
	{
		api.POST("/products", r.handler.CreateProduct)
		api.GET("/products/:id", r.handler.GetProduct)
		api.PUT("/products/:id", r.handler.UpdateProduct)
		api.DELETE("/products/:id", r.handler.ArchiveProduct)
		api.GET("/products/:id/skus", r.handler.AddSKU)
		api.POST("/products/:id/skus", r.handler.AddSKU)
		api.GET("/categories", r.handler.GetCategories)
		api.POST("/categories", r.handler.CreateCategory)
	}
}
