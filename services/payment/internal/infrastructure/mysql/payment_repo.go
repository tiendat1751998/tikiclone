package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, user_id, amount, currency, status, payment_method, psp_transaction_id, psp_provider, idempotency_key, amount_refunded, failure_reason, metadata, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.OrderID, p.UserID, p.Amount, p.Currency, p.Status, p.PaymentMethod, p.PSPTransactionID, p.PSPProvider, p.IdempotencyKey, p.AmountRefunded, p.FailureReason, p.Metadata, p.Version, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	var p domain.Payment
	if err := r.db.GetContext(ctx, &p, "SELECT id, order_id, user_id, amount, currency, status, payment_method, psp_transaction_id, psp_provider, idempotency_key, amount_refunded, failure_reason, metadata, version, authorized_at, captured_at, created_at, updated_at FROM payments WHERE id = ? AND deleted_at IS NULL", id); err != nil {
		if err == sql.ErrNoRows { return nil, domain.ErrPaymentNotFound }
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	var p domain.Payment
	if err := r.db.GetContext(ctx, &p, "SELECT id, order_id, user_id, amount, currency, status, payment_method, psp_transaction_id, psp_provider, idempotency_key, amount_refunded, failure_reason, metadata, version, authorized_at, captured_at, created_at, updated_at FROM payments WHERE order_id = ? AND deleted_at IS NULL", orderID); err != nil {
		if err == sql.ErrNoRows { return nil, domain.ErrPaymentNotFound }
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	var p domain.Payment
	if err := r.db.GetContext(ctx, &p, "SELECT id, order_id, user_id, amount, currency, status, payment_method, psp_transaction_id, psp_provider, idempotency_key, amount_refunded, failure_reason, metadata, version, authorized_at, captured_at, created_at, updated_at FROM payments WHERE idempotency_key = ? AND deleted_at IS NULL", key); err != nil {
		if err == sql.ErrNoRows { return nil, domain.ErrPaymentNotFound }
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, version int) error {
	result, err := r.db.ExecContext(ctx, "UPDATE payments SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL", status, time.Now().UTC(), id, version)
	if err != nil { return err }
	rows, err := result.RowsAffected()
	if err != nil { return fmt.Errorf("rows affected: %w", err) }
	if rows == 0 { return domain.ErrConcurrentModification }
	return nil
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	query := `UPDATE payments SET status = ?, psp_transaction_id = ?, amount_refunded = ?, failure_reason = ?, metadata = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, p.Status, p.PSPTransactionID, p.AmountRefunded, p.FailureReason, p.Metadata, time.Now().UTC(), p.ID, p.Version-1)
	if err != nil { return err }
	rows, err := result.RowsAffected()
	if err != nil { return fmt.Errorf("rows affected: %w", err) }
	if rows == 0 { return domain.ErrConcurrentModification }
	return nil
}

func (r *PaymentRepository) SaveRefund(ctx context.Context, refund *domain.Refund) error {
	query := `INSERT INTO refunds (id, payment_id, order_id, amount, currency, status, reason, psp_refund_id, idempotency_key, metadata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, refund.ID, refund.PaymentID, refund.OrderID, refund.Amount, refund.Currency, refund.Status, refund.Reason, refund.PSPRefundID, refund.IdempotencyKey, refund.Metadata, refund.CreatedAt, refund.UpdatedAt)
	return err
}

func (r *PaymentRepository) SaveWebhookEvent(ctx context.Context, event *domain.WebhookEvent) error {
	query := `INSERT INTO webhook_events (id, psp_provider, event_type, payload, signature, processed, idempotency_key, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.PSPProvider, event.EventType, event.Payload, event.Signature, event.Processed, event.IdempotencyKey, event.CreatedAt)
	return err
}

func (r *PaymentRepository) IsWebhookProcessed(ctx context.Context, idempotencyKey string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM webhook_events WHERE idempotency_key = ?", idempotencyKey)
	return count > 0, err
}

func (r *PaymentRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.AggregateType, event.AggregateID, event.EventType, event.Payload, event.CreatedAt, event.Processed)
	return err
}

func (r *PaymentRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	err := r.db.SelectContext(ctx, &events, "SELECT event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT ?", limit)
	return events, err
}

func (r *PaymentRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE WHERE event_id = ?", eventID)
	return err
}

func (r *PaymentRepository) MarkOutboxEventProcessing(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE, processing_at = NOW() WHERE event_id = ? AND processed = FALSE", eventID)
	return err
}

func (r *PaymentRepository) MarkOutboxEventFailed(ctx context.Context, eventID, reason string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET error_message = ? WHERE event_id = ?", reason, eventID)
	return err
}

func (r *PaymentRepository) SaveFraudCheck(ctx context.Context, result *domain.FraudCheckResult) error {
	query := `INSERT INTO fraud_checks (id, payment_id, user_id, risk_score, risk_level, is_fraud, reasons, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, result.ID, result.PaymentID, result.UserID, result.RiskScore, result.RiskLevel, result.IsFraud, result.Reasons, result.CreatedAt)
	return err
}

func (r *PaymentRepository) SaveIdempotencyKey(ctx context.Context, record *domain.IdempotencyRecord) error {
	query := `INSERT INTO idempotency_keys (` + "`key`" + `, payment_id, expires_at, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, record.Key, record.PaymentID, record.ExpiresAt, record.CreatedAt)
	return err
}

func (r *PaymentRepository) GetIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	if err := r.db.GetContext(ctx, &record, "SELECT `key`, payment_id, expires_at, created_at FROM idempotency_keys WHERE `key` = ?", key); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}
	return &record, nil
}
