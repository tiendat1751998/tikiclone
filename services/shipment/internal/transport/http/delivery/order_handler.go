package delivery

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/shipment/internal/domain/delivery"
	"github.com/tikiclone/tiki/services/shipment/internal/infrastructure/geo"
	infraredis "github.com/tikiclone/tiki/services/shipment/internal/infrastructure/redis"
	"go.uber.org/zap"
)

type OrderHandler struct {
	geoService *geo.Service
	redis      *infraredis.Store
	logger     *zap.Logger
	orders     map[string]*delivery.Order // in-memory for demo; use MySQL in production
}

func NewOrderHandler(geoService *geo.Service, redis *infraredis.Store, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		geoService: geoService,
		redis:      redis,
		logger:     logger.Named("order_handler"),
		orders:     make(map[string]*delivery.Order),
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req delivery.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Calculate route
	route, err := h.geoService.CalculateRoute(c.Request.Context(),
		req.Pickup.Lat, req.Pickup.Lng, req.Dropoff.Lat, req.Dropoff.Lng)
	if err != nil {
		h.logger.Warn("route calculation failed", zap.Error(err))
	}

	order := &delivery.Order{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		Pickup:     req.Pickup,
		Dropoff:    req.Dropoff,
		Status:     delivery.OrderStatusPending,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if route != nil {
		order.DistanceMeters = route.DistanceMeters
		order.DurationSeconds = route.DurationSeconds
		order.Polyline = route.Polyline
	}

	h.orders[order.ID] = order
	h.logger.Info("order created", zap.String("order_id", order.ID))

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	order, ok := h.orders[orderID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) AssignDriver(c *gin.Context) {
	orderID := c.Param("id")
	var req delivery.AssignDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	order, ok := h.orders[orderID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if !order.Status.CanTransitionTo(delivery.OrderStatusDriverAssigned) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot assign driver: order is in " + string(order.Status) + " status"})
		return
	}

	now := time.Now().UTC()
	order.DriverID = &req.DriverID
	order.Status = delivery.OrderStatusDriverAssigned
	order.AssignedAt = &now
	order.UpdatedAt = now

	h.logger.Info("driver assigned", zap.String("order_id", orderID), zap.String("driver_id", req.DriverID))
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	orderID := c.Param("id")
	var req delivery.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	newStatus := delivery.OrderStatus(req.Status)
	order, ok := h.orders[orderID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if !order.Status.CanTransitionTo(newStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid status transition from " + string(order.Status) + " to " + string(newStatus),
		})
		return
	}

	now := time.Now().UTC()
	order.Status = newStatus
	order.UpdatedAt = now

	if newStatus == delivery.OrderStatusPickedUp {
		order.PickedUpAt = &now
	}
	if newStatus == delivery.OrderStatusCompleted {
		order.DeliveredAt = &now
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	var req delivery.CancelOrderRequest
	c.ShouldBindJSON(&req)

	order, ok := h.orders[orderID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if !order.Status.CanTransitionTo(delivery.OrderStatusCancelled) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order cannot be cancelled in " + string(order.Status) + " status"})
		return
	}

	now := time.Now().UTC()
	order.Status = delivery.OrderStatusCancelled
	order.CancelledReason = req.Reason
	order.CancelledAt = &now
	order.UpdatedAt = now

	c.JSON(http.StatusOK, order)
}
