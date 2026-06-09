package domain

import "time"

type FulfillmentEventName string

const (
	EventFulfillmentCreated   FulfillmentEventName = "fulfillment.created"
	EventFulfillmentConfirmed FulfillmentEventName = "fulfillment.confirmed"
	EventFulfillmentPicking   FulfillmentEventName = "fulfillment.picking"
	EventFulfillmentPacked    FulfillmentEventName = "fulfillment.packed"
	EventFulfillmentShipped   FulfillmentEventName = "fulfillment.shipped"
	EventFulfillmentDelivered FulfillmentEventName = "fulfillment.delivered"
	EventFulfillmentReturned  FulfillmentEventName = "fulfillment.returned"
	EventFulfillmentCancelled FulfillmentEventName = "fulfillment.cancelled"
	EventFulfillmentFailed    FulfillmentEventName = "fulfillment.failed"
)

type FulfillmentEventType string

const (
	EventCreated        FulfillmentEventType = "created"
	EventConfirmed      FulfillmentEventType = "confirmed"
	EventPickStarted    FulfillmentEventType = "pick_started"
	EventPickCompleted  FulfillmentEventType = "pick_completed"
	EventPackStarted    FulfillmentEventType = "pack_started"
	EventPackCompleted  FulfillmentEventType = "pack_completed"
	EventShipped        FulfillmentEventType = "shipped"
	EventInTransit      FulfillmentEventType = "in_transit"
	EventDelivered      FulfillmentEventType = "delivered"
	EventReturned       FulfillmentEventType = "returned"
	EventCancelled      FulfillmentEventType = "cancelled"
	EventException      FulfillmentEventType = "exception"
)

type OutboxEventStatus string

const (
	OutboxStatusPending    OutboxEventStatus = "pending"
	OutboxStatusProcessing OutboxEventStatus = "processing"
	OutboxStatusProcessed  OutboxEventStatus = "processed"
	OutboxStatusFailed     OutboxEventStatus = "failed"
)

type OutboxEvent struct {
	ID         string            `db:"id"`
	EventType  string            `db:"event_type"`
	Payload    string            `db:"payload"`
	Status     OutboxEventStatus `db:"status"`
	RetryCount int               `db:"retry_count"`
	CreatedAt  time.Time         `db:"created_at"`
	UpdatedAt  time.Time         `db:"updated_at"`
}

type FulfillmentEvent struct {
	ID            string              `db:"id" json:"id"`
	FulfillmentID string              `db:"fulfillment_id" json:"fulfillment_id"`
	EventType     FulfillmentEventType `db:"event_type" json:"event_type"`
	Description   *string             `db:"description" json:"description,omitempty"`
	ActorID       *string             `db:"actor_id" json:"actor_id,omitempty"`
	CreatedAt     time.Time           `db:"created_at" json:"created_at"`
}

type FulfillmentEventMsg struct {
	EventType     FulfillmentEventName `json:"event_type"`
	FulfillmentID string               `json:"fulfillment_id"`
	OrderID       string               `json:"order_id"`
	Status        FulfillmentStatus     `json:"status"`
	Timestamp     time.Time            `json:"timestamp"`
}

func NewFulfillmentEventMsg(eventType FulfillmentEventName, fulfillmentID, orderID string, status FulfillmentStatus) *FulfillmentEventMsg {
	return &FulfillmentEventMsg{
		EventType:     eventType,
		FulfillmentID: fulfillmentID,
		OrderID:       orderID,
		Status:        status,
		Timestamp:     time.Now().UTC(),
	}
}
