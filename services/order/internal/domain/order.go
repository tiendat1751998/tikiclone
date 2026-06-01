package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusProcessing      OrderStatus = "processing"
	OrderStatusPacked          OrderStatus = "packed"
	OrderStatusShipped         OrderStatus = "shipped"
	OrderStatusDelivered       OrderStatus = "delivered"
	OrderStatusCompleted       OrderStatus = "completed"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusRefunded        OrderStatus = "refunded"
)

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = OrderStatus(v)
	case []byte:
		*s = OrderStatus(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into OrderStatus", value)
	}
	return nil
}

type Address struct {
	Name         string `json:"name,omitempty"`
	Street1      string `json:"street1"`
	AddressLine1 string `json:"address_line1,omitempty"`
	Street2      string `json:"street2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone,omitempty"`
}

// MarshalJSON handles both new and old field names for frontend compatibility
func (a Address) MarshalJSON() ([]byte, error) {
	type Alias Address
	aux := &struct {
		AddressLine1 string `json:"address_line1"`
		Alias
	}{
		AddressLine1: a.Street1,
		Alias:        (Alias)(a),
	}
	if aux.AddressLine1 == "" {
		aux.AddressLine1 = a.AddressLine1
	}
	if aux.Name == "" {
		aux.Name = a.Phone
	}
	return json.Marshal(aux)
}

// UnmarshalJSON accepts both "address_line1" and "street1" field names
func (a *Address) UnmarshalJSON(data []byte) error {
	type Alias Address
	aux := &struct {
		AddressLine1 string `json:"address_line1"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.AddressLine1 != "" && a.Street1 == "" {
		a.Street1 = aux.AddressLine1
	}
	if a.Name == "" && a.Phone != "" {
		a.Name = a.Phone
	}
	return nil
}

func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan type %T into Address", value)
	}
	return json.Unmarshal(b, a)
}

type Order struct {
	ID               string          `db:"id" json:"id"`
	OrderNumber      string          `db:"order_number" json:"order_number"`
	UserID           string          `db:"user_id" json:"user_id"`
	SellerID         string          `db:"seller_id" json:"seller_id"`
	Status           OrderStatus     `db:"status" json:"status"`
	TotalAmount      int64           `db:"total_amount" json:"total_amount,omitempty"`
	Subtotal         int64           `db:"-" json:"subtotal"`
	ShippingFee      int64           `db:"-" json:"shipping_fee"`
	Discount         int64           `db:"-" json:"discount"`
	Total            int64           `db:"-" json:"total"`
	Currency         string          `db:"currency" json:"currency"`
	ShippingAddress  Address         `db:"shipping_address" json:"shipping_address"`
	BillingAddress   Address         `db:"billing_address" json:"billing_address"`
	IdempotencyKey   string          `db:"idempotency_key" json:"idempotency_key"`
	SnapshotID       string          `db:"snapshot_id" json:"snapshot_id"`
	ParentOrderID    *string         `db:"parent_order_id" json:"parent_order_id,omitempty"`
	Metadata         *json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	Version             int             `db:"version" json:"version"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
	Items            []OrderItem     `json:"items,omitempty"`
}

func EnrichOrderItems(items []OrderItem) {
	for i := range items {
		if items[i].Snapshot != nil {
			var snap SnapshotItem
			if err := json.Unmarshal(items[i].Snapshot, &snap); err == nil {
				if items[i].Name == "" {
					items[i].Name = snap.Name
				}
				if items[i].Price == 0 {
					items[i].Price = snap.UnitPrice
				}
				items[i].Total = items[i].TotalPrice
				if items[i].Total == 0 && items[i].Price > 0 {
					items[i].Total = items[i].Price * int64(items[i].Quantity)
				}
			}
		}
		if items[i].Price == 0 {
			items[i].Price = items[i].UnitPrice
		}
		if items[i].Total == 0 {
			items[i].Total = items[i].TotalPrice
		}
	}
}

func EnrichOrder(order *Order) {
	EnrichOrderItems(order.Items)
	order.Subtotal = order.TotalAmount
	order.Total = order.TotalAmount
	order.ShippingFee = 0
	order.Discount = 0
}

func NewOrder(userID, sellerID, currency, idempotencyKey string, shippingAddr, billingAddr Address, items []OrderItem) *Order {
	now := time.Now().UTC()
	total := int64(0)
	for i := range items {
		items[i].TotalPrice = int64(items[i].Quantity) * items[i].UnitPrice
		total += items[i].TotalPrice
	}
	return &Order{
		ID:              uuid.New().String(),
		UserID:          userID,
		SellerID:        sellerID,
		Status:          OrderStatusPending,
		TotalAmount:     total,
		Currency:        currency,
		ShippingAddress: shippingAddr,
		BillingAddress:  billingAddr,
		IdempotencyKey:  idempotencyKey,
		Items:           items,
		Version:            1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (o *Order) CanTransitionTo(target OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:         {OrderStatusAwaitingPayment, OrderStatusCancelled},
		OrderStatusAwaitingPayment: {OrderStatusPaid, OrderStatusCancelled},
		OrderStatusPaid:            {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing:      {OrderStatusPacked, OrderStatusCancelled},
		OrderStatusPacked:          {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:         {OrderStatusDelivered, OrderStatusRefunded},
		OrderStatusDelivered:       {OrderStatusCompleted, OrderStatusRefunded},
		OrderStatusCompleted:       {OrderStatusRefunded},
		OrderStatusCancelled:       {},
		OrderStatusRefunded:        {},
	}
	allowed, ok := validTransitions[o.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

func (o *Order) TransitionTo(target OrderStatus, actorID, actorType, reason string) (*LifecycleEvent, error) {
	if !o.CanTransitionTo(target) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, o.Status, target)
	}
	now := time.Now().UTC()
	fromStatus := o.Status
	o.Status = target
	o.Version++
	o.UpdatedAt = now
	event := &LifecycleEvent{
		ID:               uuid.New().String(),
		OrderID:          o.ID,
		FromStatus:       fromStatus,
		ToStatus:         target,
		TransitionReason: reason,
		ActorID:          actorID,
		ActorType:        actorType,
		CreatedAt:        now,
	}
	return event, nil
}

func (o *Order) IsCancellable() bool {
	return o.Status != OrderStatusCancelled &&
		o.Status != OrderStatusRefunded &&
		o.Status != OrderStatusCompleted &&
		o.Status != OrderStatusDelivered &&
		o.Status != OrderStatusShipped
}

func (o *Order) IsTerminal() bool {
	return o.Status == OrderStatusCancelled ||
		o.Status == OrderStatusRefunded ||
		o.Status == OrderStatusCompleted
}
