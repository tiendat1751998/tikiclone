package delivery

import "time"

// GeoSearchResult represents a Nominatim address search result
type GeoSearchResult struct {
	Address string  `json:"address"`
	Name    string  `json:"name,omitempty"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// ReverseGeocodeResult represents a Nominatim reverse geocoding result
type ReverseGeocodeResult struct {
	Address  string  `json:"address"`
	Name     string  `json:"name,omitempty"`
	Street   string  `json:"street,omitempty"`
	City     string  `json:"city,omitempty"`
	District string  `json:"district,omitempty"`
	Ward     string  `json:"ward,omitempty"`
	Country  string  `json:"country,omitempty"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// RouteResult represents an OSRM routing result
type RouteResult struct {
	DistanceMeters  int    `json:"distance_meters"`
	DurationSeconds int    `json:"duration_seconds"`
	Polyline        string `json:"polyline"`
}

// DriverStatus represents driver availability
type DriverStatus string

const (
	DriverStatusOffline DriverStatus = "offline"
	DriverStatusOnline  DriverStatus = "online"
	DriverStatusBusy    DriverStatus = "busy"
)

// DriverLocationUpdate is sent by drivers to update their position
type DriverLocationUpdate struct {
	DriverID string  `json:"driver_id" binding:"required"`
	Lat      float64 `json:"lat" binding:"required"`
	Lng      float64 `json:"lng" binding:"required"`
}

// NearbyDriver represents a driver found near a location
type NearbyDriver struct {
	DriverID string  `json:"driver_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Distance float64 `json:"distance_meters"`
	Status   string  `json:"status"`
}

// OrderStatus represents delivery order state
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusSearchingDriver OrderStatus = "searching_driver"
	OrderStatusDriverAssigned  OrderStatus = "driver_assigned"
	OrderStatusPickedUp        OrderStatus = "picked_up"
	OrderStatusDelivering      OrderStatus = "delivering"
	OrderStatusCompleted       OrderStatus = "completed"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// Valid order status transitions
var ValidTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending:         {OrderStatusSearchingDriver, OrderStatusCancelled},
	OrderStatusSearchingDriver: {OrderStatusDriverAssigned, OrderStatusCancelled},
	OrderStatusDriverAssigned:  {OrderStatusPickedUp, OrderStatusCancelled},
	OrderStatusPickedUp:        {OrderStatusDelivering},
	OrderStatusDelivering:      {OrderStatusCompleted, OrderStatusCancelled},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	valid, ok := ValidTransitions[s]
	if !ok {
		return false
	}
	for _, v := range valid {
		if v == next {
			return true
		}
	}
	return false
}

// Location represents a geographic point with optional address
type Location struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
}

// CreateOrderRequest is the request to create a delivery order
type CreateOrderRequest struct {
	CustomerID string   `json:"customer_id" binding:"required"`
	Pickup     Location `json:"pickup" binding:"required"`
	Dropoff    Location `json:"dropoff" binding:"required"`
	Note       string   `json:"note"`
}

// AssignDriverRequest assigns a driver to an order
type AssignDriverRequest struct {
	DriverID string `json:"driver_id" binding:"required"`
}

// UpdateStatusRequest updates order status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// CancelOrderRequest cancels an order
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

// TrackingUpdate is broadcast via WebSocket
type TrackingUpdate struct {
	OrderID   string    `json:"order_id"`
	DriverID  string    `json:"driver_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// WSEvent is a WebSocket event message
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// DriverAssignedEvent is broadcast when a driver is assigned
type DriverAssignedEvent struct {
	OrderID   string    `json:"order_id"`
	DriverID  string    `json:"driver_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Order represents a delivery order
type Order struct {
	ID              string     `json:"id"`
	CustomerID      string     `json:"customer_id"`
	DriverID        *string    `json:"driver_id,omitempty"`
	Pickup          Location   `json:"pickup"`
	Dropoff         Location   `json:"dropoff"`
	Status          OrderStatus `json:"status"`
	DistanceMeters  int        `json:"distance_meters"`
	DurationSeconds int        `json:"duration_seconds"`
	Polyline        string     `json:"polyline,omitempty"`
	Note            string     `json:"note,omitempty"`
	CancelledReason string     `json:"cancelled_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	PickedUpAt      *time.Time `json:"picked_up_at,omitempty"`
	DeliveredAt     *time.Time `json:"delivered_at,omitempty"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
}
