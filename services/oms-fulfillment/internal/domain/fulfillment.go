package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FulfillmentStatus string

const (
	FulfillmentStatusPending   FulfillmentStatus = "pending"
	FulfillmentStatusConfirmed FulfillmentStatus = "confirmed"
	FulfillmentStatusPicking   FulfillmentStatus = "picking"
	FulfillmentStatusPacked    FulfillmentStatus = "packed"
	FulfillmentStatusShipped   FulfillmentStatus = "shipped"
	FulfillmentStatusDelivered FulfillmentStatus = "delivered"
	FulfillmentStatusReturned  FulfillmentStatus = "returned"
	FulfillmentStatusCancelled FulfillmentStatus = "cancelled"
	FulfillmentStatusFailed    FulfillmentStatus = "failed"
)

func (s FulfillmentStatus) Valid() bool {
	switch s {
	case FulfillmentStatusPending, FulfillmentStatusConfirmed, FulfillmentStatusPicking,
		FulfillmentStatusPacked, FulfillmentStatusShipped, FulfillmentStatusDelivered,
		FulfillmentStatusReturned, FulfillmentStatusCancelled, FulfillmentStatusFailed:
		return true
	}
	return false
}

func (s FulfillmentStatus) CanTransitionTo(target FulfillmentStatus) bool {
	transitions := map[FulfillmentStatus][]FulfillmentStatus{
		FulfillmentStatusPending:   {FulfillmentStatusConfirmed, FulfillmentStatusCancelled, FulfillmentStatusFailed},
		FulfillmentStatusConfirmed: {FulfillmentStatusPicking, FulfillmentStatusCancelled, FulfillmentStatusFailed},
		FulfillmentStatusPicking:   {FulfillmentStatusPacked, FulfillmentStatusCancelled, FulfillmentStatusFailed},
		FulfillmentStatusPacked:    {FulfillmentStatusShipped, FulfillmentStatusCancelled, FulfillmentStatusFailed},
		FulfillmentStatusShipped:   {FulfillmentStatusDelivered, FulfillmentStatusReturned, FulfillmentStatusFailed},
		FulfillmentStatusDelivered: {FulfillmentStatusReturned},
	}
	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

type FulfillmentItemStatus string

const (
	ItemStatusPending  FulfillmentItemStatus = "pending"
	ItemStatusPicked   FulfillmentItemStatus = "picked"
	ItemStatusPacked   FulfillmentItemStatus = "packed"
	ItemStatusShipped  FulfillmentItemStatus = "shipped"
	ItemStatusReturned FulfillmentItemStatus = "returned"
	ItemStatusCancelled FulfillmentItemStatus = "cancelled"
)

type ShippingAddress struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type FulfillmentOrder struct {
	ID                    string            `db:"id" json:"id"`
	OrderID               string            `db:"order_id" json:"order_id"`
	SellerID             string            `db:"seller_id" json:"seller_id"`
	WarehouseID          *string           `db:"warehouse_id" json:"warehouse_id,omitempty"`
	Status               FulfillmentStatus `db:"status" json:"status"`
	Priority             string            `db:"priority" json:"priority"`
	ShippingMethod       string            `db:"shipping_method" json:"shipping_method"`
	ShippingAddressJSON  string            `db:"shipping_address" json:"-"`
	ShippingAddress      *ShippingAddress  `db:"-" json:"shipping_address"`
	EstimatedShipDate    *time.Time        `db:"estimated_ship_date" json:"estimated_ship_date,omitempty"`
	EstimatedDeliveryDate *time.Time       `db:"estimated_delivery_date" json:"estimated_delivery_date,omitempty"`
	ActualShippedAt      *time.Time        `db:"actual_shipped_at" json:"actual_shipped_at,omitempty"`
	ActualDeliveredAt    *time.Time        `db:"actual_delivered_at" json:"actual_delivered_at,omitempty"`
	TrackingNumber       *string           `db:"tracking_number" json:"tracking_number,omitempty"`
	Carrier              *string           `db:"carrier" json:"carrier,omitempty"`
	TotalItems           int               `db:"total_items" json:"total_items"`
	TotalWeightGrams     int               `db:"total_weight_grams" json:"total_weight_grams"`
	Notes                *string           `db:"notes" json:"notes,omitempty"`
	MetadataJSON         *string           `db:"metadata" json:"-"`
	CreatedAt            time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time         `db:"updated_at" json:"updated_at"`

	Items []FulfillmentItem `db:"-" json:"items,omitempty"`
}

func NewFulfillmentOrder(orderID, sellerID, shippingMethod string, address *ShippingAddress, totalItems, totalWeight int) *FulfillmentOrder {
	now := time.Now().UTC()
	addrJSON, _ := json.Marshal(address)
	return &FulfillmentOrder{
		ID:                   uuid.New().String(),
		OrderID:              orderID,
		SellerID:            sellerID,
		Status:               FulfillmentStatusPending,
		Priority:             "normal",
		ShippingMethod:       shippingMethod,
		ShippingAddressJSON:  string(addrJSON),
		ShippingAddress:      address,
		TotalItems:           totalItems,
		TotalWeightGrams:     totalWeight,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func (f *FulfillmentOrder) TransitionTo(target FulfillmentStatus) error {
	if !f.Status.CanTransitionTo(target) {
		return ErrInvalidStateTransition
	}
	f.Status = target
	f.UpdatedAt = time.Now().UTC()
	return nil
}

type FulfillmentItem struct {
	ID              string              `db:"id" json:"id"`
	FulfillmentID   string              `db:"fulfillment_id" json:"fulfillment_id"`
	OrderItemID     string              `db:"order_item_id" json:"order_item_id"`
	ProductID       string              `db:"product_id" json:"product_id"`
	SKUID           string              `db:"sku_id" json:"sku_id"`
	Quantity        int                 `db:"quantity" json:"quantity"`
	PickedQuantity  int                 `db:"picked_quantity" json:"picked_quantity"`
	PackedQuantity  int                 `db:"packed_quantity" json:"packed_quantity"`
	Status          FulfillmentItemStatus `db:"status" json:"status"`
	PickedAt        *time.Time          `db:"picked_at" json:"picked_at,omitempty"`
	PackedAt        *time.Time          `db:"packed_at" json:"packed_at,omitempty"`
	PickedBy        *string             `db:"picked_by" json:"picked_by,omitempty"`
	PackedBy        *string             `db:"packed_by" json:"packed_by,omitempty"`
	CreatedAt       time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `db:"updated_at" json:"updated_at"`
}

func NewFulfillmentItem(fulfillmentID, orderItemID, productID, skuID string, quantity int) *FulfillmentItem {
	now := time.Now().UTC()
	return &FulfillmentItem{
		ID:            uuid.New().String(),
		FulfillmentID: fulfillmentID,
		OrderItemID:   orderItemID,
		ProductID:     productID,
		SKUID:         skuID,
		Quantity:      quantity,
		Status:        ItemStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
