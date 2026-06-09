package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/transport/http/middleware"
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
	engine.Use(middleware.Recovery())

	api := engine.Group("/api/v1/oms")
	if r.authMw != nil {
		api.Use(r.authMw)
	}
	{
		api.POST("/fulfillments", r.handler.CreateFulfillment)
		api.GET("/fulfillments", r.handler.ListFulfillments)
		api.GET("/fulfillments/:id", r.handler.GetFulfillment)
		api.PUT("/fulfillments/:id/status", r.handler.TransitionStatus)
		api.GET("/fulfillments/:id/events", r.handler.GetFulfillmentEvents)

		api.GET("/warehouses", r.handler.ListWarehouses)
		api.GET("/warehouses/:id", r.handler.GetWarehouse)
		api.GET("/warehouses/:id/inventory", r.handler.GetWarehouseInventory)

		api.POST("/returns", r.handler.CreateReturn)
		api.GET("/returns", r.handler.ListReturns)
		api.GET("/returns/:id", r.handler.GetReturn)
		api.POST("/returns/:id/approve", r.handler.ApproveReturn)
		api.POST("/returns/:id/reject", r.handler.RejectReturn)
	}
}
