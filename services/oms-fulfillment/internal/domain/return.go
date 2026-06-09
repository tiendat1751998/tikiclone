package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReturnType string

const (
	ReturnTypeFull    ReturnType = "full"
	ReturnTypePartial ReturnType = "partial"
	ReturnTypeExchange ReturnType = "exchange"
)

type ReturnReason string

const (
	ReturnReasonDefective       ReturnReason = "defective"
	ReturnReasonWrongItem       ReturnReason = "wrong_item"
	ReturnReasonNotAsDescribed  ReturnReason = "not_as_described"
	ReturnReasonChangedMind     ReturnReason = "changed_mind"
	ReturnReasonDamagedShipping ReturnReason = "damaged_in_shipping"
	ReturnReasonLateDelivery    ReturnReason = "late_delivery"
	ReturnReasonOther           ReturnReason = "other"
)

type ReturnStatus string

const (
	ReturnStatusRequested     ReturnStatus = "requested"
	ReturnStatusApproved      ReturnStatus = "approved"
	ReturnStatusRejected      ReturnStatus = "rejected"
	ReturnStatusItemsShipped  ReturnStatus = "items_shipped"
	ReturnStatusItemsReceived ReturnStatus = "items_received"
	ReturnStatusRefunded      ReturnStatus = "refunded"
	ReturnStatusCompleted     ReturnStatus = "completed"
	ReturnStatusCancelled     ReturnStatus = "cancelled"
)

type ReturnRequest struct {
	ID             string       `db:"id" json:"id"`
	FulfillmentID  string       `db:"fulfillment_id" json:"fulfillment_id"`
	OrderID        string       `db:"order_id" json:"order_id"`
	UserID         string       `db:"user_id" json:"user_id"`
	ReturnType     ReturnType   `db:"return_type" json:"return_type"`
	Reason         ReturnReason `db:"reason" json:"reason"`
	ReasonDetail   *string      `db:"reason_detail" json:"reason_detail,omitempty"`
	Status         ReturnStatus `db:"status" json:"status"`
	RefundAmount   int64        `db:"refund_amount" json:"refund_amount"`
	Currency       string       `db:"currency" json:"currency"`
	TrackingNumber *string      `db:"tracking_number" json:"tracking_number,omitempty"`
	ApprovedBy     *string      `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt     *time.Time   `db:"approved_at" json:"approved_at,omitempty"`
	ReceivedAt     *time.Time   `db:"received_at" json:"received_at,omitempty"`
	RefundedAt     *time.Time   `db:"refunded_at" json:"refunded_at,omitempty"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at" json:"updated_at"`
	Items          []ReturnItem `db:"-" json:"items,omitempty"`
}

func NewReturnRequest(fulfillmentID, orderID, userID string, returnType ReturnType, reason ReturnReason, refundAmount int64) *ReturnRequest {
	return &ReturnRequest{
		ID:            uuid.New().String(),
		FulfillmentID: fulfillmentID,
		OrderID:       orderID,
		UserID:        userID,
		ReturnType:    returnType,
		Reason:        reason,
		Status:        ReturnStatusRequested,
		RefundAmount:  refundAmount,
		Currency:      "SGD",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

type ReturnItem struct {
	ID                string `db:"id" json:"id"`
	ReturnID          string `db:"return_id" json:"return_id"`
	FulfillmentItemID string `db:"fulfillment_item_id" json:"fulfillment_item_id"`
	SKUID             string `db:"sku_id" json:"sku_id"`
	Quantity          int    `db:"quantity" json:"quantity"`
	Condition         string `db:"condition" json:"condition"`
	RefundAmount      int64  `db:"refund_amount" json:"refund_amount"`
	Restockable       bool   `db:"restockable" json:"restockable"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

