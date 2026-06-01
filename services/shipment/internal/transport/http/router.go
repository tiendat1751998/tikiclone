package http

import (
	"github.com/gin-gonic/gin"
	deliveryhttp "github.com/tikiclone/tiki/services/shipment/internal/transport/http/delivery"
	"github.com/tikiclone/tiki/services/shipment/internal/transport/http/middleware"
)

type Router struct {
	handler       *Handler
	deliveryGeo   *deliveryhttp.GeoHandler
	deliveryOrder *deliveryhttp.OrderHandler
	deliveryWS    *deliveryhttp.WSHandler
	authMw        gin.HandlerFunc
}

func NewRouter(handler *Handler, deliveryGeo *deliveryhttp.GeoHandler, deliveryOrder *deliveryhttp.OrderHandler, deliveryWS *deliveryhttp.WSHandler, authMw gin.HandlerFunc) *Router {
	return &Router{handler: handler, deliveryGeo: deliveryGeo, deliveryOrder: deliveryOrder, deliveryWS: deliveryWS, authMw: authMw}
}

func (r *Router) Setup(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	v1.Use(middleware.RequestID())
	v1.Use(middleware.Recovery())

	// Shipment routes (existing)
	shipments := v1.Group("/shipments")
	if r.authMw != nil {
		shipments.Use(r.authMw)
	}
	{
		shipments.POST("", r.handler.CreateShipment)
		shipments.GET("/:id", r.handler.GetShipment)
		shipments.POST("/:id/status", r.handler.UpdateStatus)
		shipments.GET("/:id/tracking", r.handler.GetTracking)
	}

	// Delivery geo routes
	delivery := v1.Group("/delivery")
	{
		delivery.GET("/search", r.deliveryGeo.SearchAddress)
		delivery.GET("/reverse", r.deliveryGeo.ReverseGeocode)
		delivery.POST("/route", r.deliveryGeo.CalculateRoute)
		delivery.POST("/drivers/location", r.deliveryGeo.UpdateDriverLocation)
		delivery.GET("/drivers/nearby", r.deliveryGeo.FindNearbyDrivers)
	}

	// Delivery order routes
	orders := v1.Group("/orders")
	{
		orders.POST("", r.deliveryOrder.CreateOrder)
		orders.GET("/:id", r.deliveryOrder.GetOrder)
		orders.POST("/:id/assign", r.deliveryOrder.AssignDriver)
		orders.PUT("/:id/status", r.deliveryOrder.UpdateStatus)
		orders.POST("/:id/cancel", r.deliveryOrder.CancelOrder)
	}

	// WebSocket
	v1.GET("/ws", r.deliveryWS.HandleWS)

	v1.POST("/webhooks/:provider", r.handler.HandleWebhook)
}
