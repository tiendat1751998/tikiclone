package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/auth"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
)

type Router struct {
	handler  *Handler
	health   *health.Checker
	jwtAuth  gin.HandlerFunc
}

// NewRouter creates a new router with mandatory JWT authentication.
// Unlike other services, cart auth is REQUIRED - fail fast if not configured.
func NewRouter(handler *Handler, healthChecker *health.Checker, jwtSecret string) *Router {
	if jwtSecret == "" {
		panic("cart: JWT_ACCESS_SECRET is required - cannot start without authentication")
	}
	return &Router{handler: handler, health: healthChecker, jwtAuth: auth.GinJWTAuth(jwtSecret)}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(
		middleware.Recovery(),
		middleware.ErrorHandler(),
		middleware.RequestID(),
		middleware.CORS(),
		middleware.OTelMiddleware("tiki-cart"),
		observability.ObserveHTTPMetrics("tiki-cart"),
	)

	engine.GET("/health", r.health.LivenessHandler())
	engine.GET("/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	api := engine.Group("/")
	api.Use(r.jwtAuth)
	{
		api.GET("/", r.handler.GetCart)
		api.POST("/items", r.handler.AddItem)
		api.PATCH("/items/:item_id", r.handler.UpdateItem)
		api.DELETE("/items/:item_id", r.handler.RemoveItem)
		api.DELETE("/", r.handler.ClearCart)
		api.POST("/merge", r.handler.MergeCarts)
		api.POST("/checkout-preview", r.handler.CheckoutPreview)
	}
}
