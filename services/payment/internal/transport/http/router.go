package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/payment/internal/transport/http/middleware"
)

type Router struct {
	handler            *Handler
	authMw             gin.HandlerFunc
	webhookMiddlewares []gin.HandlerFunc
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc, webhookMiddlewares ...gin.HandlerFunc) *Router {
	return &Router{handler: handler, authMw: authMw, webhookMiddlewares: webhookMiddlewares}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())

	if r.authMw != nil { engine.Use(r.authMw) }
	{
		engine.POST("/", r.handler.AuthorizePayment)
		engine.GET("/:id", r.handler.GetPayment)
		engine.POST("/:id/capture", r.handler.CapturePayment)
		engine.POST("/:id/refund", r.handler.RefundPayment)
	}

	// Webhook endpoint (no auth, uses signature verification)
	webhooks := engine.Group("/webhooks")
	for _, mw := range r.webhookMiddlewares {
		webhooks.Use(mw)
	}
	webhooks.POST("/:provider", r.handler.HandleWebhook)
}
