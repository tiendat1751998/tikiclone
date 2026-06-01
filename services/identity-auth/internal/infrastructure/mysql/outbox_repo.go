package mysql

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type OutboxEventRepository struct {
	db *sqlx.DB
}

func NewOutboxEventRepository(db *sqlx.DB) *OutboxEventRepository {
	return &OutboxEventRepository{db: db}
}

func (r *OutboxEventRepository) Create(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, processed, retry_count, created_at)
		VALUES (:event_id, :aggregate_type, :aggregate_id, :event_type, :payload, :processed, :retry_count, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, event)
	return err
}

func (r *OutboxEventRepository) MarkProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	query := `UPDATE outbox_events SET processed = TRUE, processed_at = ? WHERE event_id = ?`
	_, err := r.db.ExecContext(ctx, query, now, eventID)
	return err
}

func (r *OutboxEventRepository) MarkFailed(ctx context.Context, eventID, errorMsg string) error {
	query := `UPDATE outbox_events SET retry_count = retry_count + 1, last_error = ? WHERE event_id = ?`
	_, err := r.db.ExecContext(ctx, query, errorMsg, eventID)
	return err
}