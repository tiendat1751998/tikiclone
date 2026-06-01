package mysql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/order/internal/domain"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	query := `SELECT event_id, aggregate_type, aggregate_id, event_type, status, error_message, retries, payload, created_at, processed FROM outbox_events WHERE status = 'pending' ORDER BY created_at ASC LIMIT ?`
	if err := r.db.SelectContext(ctx, &events, query, limit); err != nil {
		return nil, fmt.Errorf("failed to get outbox events: %w", err)
	}
	return events, nil
}

func (r *OutboxRepository) MarkOutboxEventProcessing(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'processing' WHERE event_id = ? AND status = 'pending'`, eventID)
	return err
}

func (r *OutboxRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'processed', processed = TRUE WHERE event_id = ? AND status = 'processing'`, eventID)
	return err
}

func (r *OutboxRepository) MarkOutboxEventsProcessed(ctx context.Context, eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In(
		`UPDATE outbox_events SET status = 'processed', processed = TRUE WHERE event_id IN (?) AND status = 'processing'`, eventIDs)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *OutboxRepository) MarkOutboxEventFailed(ctx context.Context, eventID, errorMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'failed', error_message = ?, retries = retries + 1 WHERE event_id = ?`, errorMsg, eventID)
	return err
}

func (r *OutboxRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, status, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.AggregateType, event.AggregateID, event.EventType,
		event.Status, event.Payload, event.CreatedAt, event.Processed,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}

func (r *OutboxRepository) SaveOutboxEventInTx(ctx context.Context, tx *sqlx.Tx, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, status, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query,
		event.ID, event.AggregateType, event.AggregateID, event.EventType,
		event.Status, event.Payload, event.CreatedAt, event.Processed,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}
