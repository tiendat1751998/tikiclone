package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/health"
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
	engine.Use(middleware.OTelMiddleware("logistics-delivery"))
	engine.Use(observability.ObserveHTTPMetrics("logistics-delivery"))

	api := engine.Group("/api/v1")

	shipments := api.Group("/shipments")
	shipments.POST("", r.handler.CreateShipment)
	shipments.GET("/:id", r.handler.GetShipment)
	shipments.GET("", r.handler.ListShipments)
	shipments.PUT("/:id/status", r.handler.UpdateShipmentStatus)
	shipments.GET("/order/:orderID", r.handler.GetByOrderID)

	tracking := api.Group("/tracking")
	tracking.POST("/events", r.handler.AppendTrackingEvent)
	tracking.GET("/:shipmentID/timeline", r.handler.GetTrackingTimeline)
	tracking.GET("/:shipmentID/last", r.handler.GetLastTrackingEvent)

	routes := api.Group("/routing")
	routes.POST("/assign", r.handler.AssignRoute)
	routes.POST("/optimize", r.handler.OptimizeWaypoints)
	routes.GET("/shipment/:shipmentID", r.handler.GetShipmentRoutes)

	dispatch := api.Group("/dispatch")
	dispatch.POST("", r.handler.CreateDispatch)
	dispatch.PUT("/:id/assign", r.handler.AssignDispatchCourier)
	dispatch.PUT("/:id/enroute", r.handler.MarkDispatchEnRoute)
	dispatch.PUT("/:id/complete", r.handler.MarkDispatchComplete)
	dispatch.GET("/shipment/:shipmentID", r.handler.GetDispatchByShipment)

	couriers := api.Group("/couriers")
	couriers.POST("", r.handler.CreateCourier)
	couriers.GET("/:id", r.handler.GetCourier)
	couriers.POST("/webhook", r.handler.ProcessCourierWebhook)
	couriers.PUT("/:id/location", r.handler.UpdateCourierLocation)

	fulfillment := api.Group("/fulfillment")
	fulfillment.POST("", r.handler.CreateFulfillment)
	fulfillment.PUT("/:id/pack", r.handler.MarkFulfillmentPacked)
	fulfillment.PUT("/:id/ship", r.handler.MarkFulfillmentShipped)
	fulfillment.GET("/shipment/:shipmentID", r.handler.GetFulfillmentByShipment)

	pickups := api.Group("/pickups")
	pickups.POST("", r.handler.CreatePickup)
	pickups.PUT("/:id/complete", r.handler.MarkPickupComplete)
	pickups.PUT("/:id/fail", r.handler.MarkPickupFailed)
	pickups.GET("/shipment/:shipmentID", r.handler.GetPickupByShipment)

	estimations := api.Group("/estimations")
	estimations.POST("/calculate", r.handler.CalculateEstimation)
	estimations.GET("/shipment/:shipmentID", r.handler.GetEstimationByShipment)

	engine.GET("/health/live", r.health.LivenessHandler())
	engine.GET("/health/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())
}
