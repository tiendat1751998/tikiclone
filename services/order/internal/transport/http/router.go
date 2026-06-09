package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/order/internal/transport/http/middleware"
)

type Router struct {
	handler *Handler
	authMw  gin.HandlerFunc
}

func NewRouter(handler *Handler, authMw gin.HandlerFunc) *Router {
	return &Router{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())

	protected := engine.Group("/")
	if r.authMw != nil { protected.Use(r.authMw) }
	protected.POST("/", r.handler.CreateOrder)
	protected.GET("/", r.handler.ListOrders)
	protected.GET("/:id", r.handler.GetOrder)
	protected.GET("/:id/status", r.handler.GetOrderStatus)
	protected.POST("/:id/cancel", r.handler.CancelOrder)
	protected.GET("/:id/history", r.handler.GetOrderHistory)
	protected.GET("/:id/reconciliation", r.handler.GetReconciliationStatus)
}
