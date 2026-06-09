package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusPartialRefund PaymentStatus = "partial_refund"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

func (s PaymentStatus) String() string { return string(s) }

func (s PaymentStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *PaymentStatus) Scan(value interface{}) error {
	if value == nil { return nil }
	switch v := value.(type) {
	case string: *s = PaymentStatus(v)
	case []byte: *s = PaymentStatus(string(v))
	default: return fmt.Errorf("cannot scan %T into PaymentStatus", value)
	}
	return nil
}

type PaymentMethod string

const (
	PaymentMethodCreditCard    PaymentMethod = "credit_card"
	PaymentMethodDebitCard     PaymentMethod = "debit_card"
	PaymentMethodEWallet       PaymentMethod = "e_wallet"
	PaymentMethodBankTransfer  PaymentMethod = "bank_transfer"
	PaymentMethodCOD           PaymentMethod = "cod"
	PaymentMethodVNPayQR       PaymentMethod = "vnpay_qr"
	PaymentMethodVNPayToken      PaymentMethod = "vnpay_token"
	PaymentMethodVNPayGateway    PaymentMethod = "vnpay_gateway"
)

type Payment struct {
	ID                string            `db:"id" json:"id"`
	OrderID           string            `db:"order_id" json:"order_id"`
	UserID            string            `db:"user_id" json:"user_id"`
	Amount            int64             `db:"amount" json:"amount"`
	Currency          string            `db:"currency" json:"currency"`
	Status            PaymentStatus     `db:"status" json:"status"`
	PaymentMethod     PaymentMethod     `db:"payment_method" json:"payment_method"`
	PSPTransactionID  string            `db:"psp_transaction_id" json:"psp_transaction_id,omitempty"`
	PSPProvider       string            `db:"psp_provider" json:"psp_provider"`
	IdempotencyKey    string            `db:"idempotency_key" json:"idempotency_key"`
	AmountRefunded    int64             `db:"amount_refunded" json:"amount_refunded"`
	FailureReason     string            `db:"failure_reason" json:"failure_reason,omitempty"`
	Metadata          *json.RawMessage  `db:"metadata" json:"metadata,omitempty"`
	Version           int               `db:"version" json:"version"`
	AuthorizedAt      *time.Time        `db:"authorized_at" json:"authorized_at,omitempty"`
	CapturedAt        *time.Time        `db:"captured_at" json:"captured_at,omitempty"`
	CreatedAt         time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`
}

func NewPayment(orderID, userID string, amount int64, currency string, method PaymentMethod, pspProvider, idempotencyKey string) *Payment {
	now := time.Now().UTC()
	return &Payment{
		ID:             uuid.New().String(),
		OrderID:        orderID,
		UserID:         userID,
		Amount:         amount,
		Currency:       currency,
		Status:         PaymentStatusPending,
		PaymentMethod:  method,
		PSPProvider:    pspProvider,
		IdempotencyKey: idempotencyKey,
		AmountRefunded: 0,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (p *Payment) CanTransitionTo(target PaymentStatus) bool {
	validTransitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusPending:    {PaymentStatusAuthorized, PaymentStatusFailed, PaymentStatusExpired, PaymentStatusCancelled},
		PaymentStatusAuthorized: {PaymentStatusCaptured, PaymentStatusFailed, PaymentStatusCancelled},
		PaymentStatusCaptured:   {PaymentStatusRefunded, PaymentStatusPartialRefund},
		PaymentStatusPartialRefund: {PaymentStatusRefunded},
		PaymentStatusFailed:     {},
		PaymentStatusExpired:    {},
		PaymentStatusRefunded:   {},
		PaymentStatusCancelled:  {},
	}
	allowed, ok := validTransitions[p.Status]
	if !ok { return false }
	for _, s := range allowed {
		if s == target { return true }
	}
	return false
}

func (p *Payment) TransitionTo(target PaymentStatus) error {
	if !p.CanTransitionTo(target) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidPaymentState, p.Status, target)
	}
	now := time.Now().UTC()
	p.Status = target
	p.Version++
	p.UpdatedAt = now
	switch target {
	case PaymentStatusAuthorized:
		p.AuthorizedAt = &now
	case PaymentStatusCaptured:
		p.CapturedAt = &now
	}
	return nil
}

func (p *Payment) IsTerminal() bool {
	return p.Status == PaymentStatusFailed || p.Status == PaymentStatusExpired ||
		p.Status == PaymentStatusRefunded || p.Status == PaymentStatusCancelled
}

func (p *Payment) RemainingAmount() int64 {
	return p.Amount - p.AmountRefunded
}
