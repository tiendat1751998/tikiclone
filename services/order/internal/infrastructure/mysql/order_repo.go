package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/order/internal/domain"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order, outboxEvent *domain.OutboxEvent) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO orders (id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, query,
		order.ID, order.OrderNumber, order.UserID, order.SellerID, order.Status,
		order.TotalAmount, order.Currency, order.ShippingAddress, order.BillingAddress,
		order.IdempotencyKey, order.SnapshotID, order.ParentOrderID,
		order.Metadata, order.Version, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		itemQuery := `INSERT INTO order_items (id, order_id, product_id, sku_id, shop_id, quantity, unit_price, total_price, snapshot, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, itemQuery,
			order.Items[i].ID, order.Items[i].OrderID, order.Items[i].ProductID,
			order.Items[i].SkuID, order.Items[i].ShopID, order.Items[i].Quantity,
			order.Items[i].UnitPrice, order.Items[i].TotalPrice, order.Items[i].Snapshot,
			order.Items[i].CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Save outbox event in the same transaction (atomic with order creation)
	if outboxEvent != nil {
		oq := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, status, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		if _, err = tx.ExecContext(ctx, oq,
			outboxEvent.ID, outboxEvent.AggregateType, outboxEvent.AggregateID,
			outboxEvent.EventType, outboxEvent.Status, outboxEvent.Payload,
			outboxEvent.CreatedAt, outboxEvent.Processed,
		); err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE id = ? AND deleted_at IS NULL`
	var order domain.Order
	err := r.db.GetContext(ctx, &order, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find order by id: %w", err)
	}

	items, err := r.FindItemsByOrderID(ctx, id)
	if err == nil {
		order.Items = items
	}

	return &order, nil
}

func (r *OrderRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*domain.Order, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var orders []*domain.Order
	query, args, err := sqlx.In("SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE id IN (?) AND deleted_at IS NULL", ids)
	if err != nil {
		return nil, fmt.Errorf("build find by ids: %w", err)
	}
	query = r.db.Rebind(query)
	if err := r.db.SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, fmt.Errorf("find orders by ids: %w", err)
	}
	result := make(map[string]*domain.Order, len(orders))
	for _, o := range orders {
		result[o.ID] = o
	}
	return result, nil
}

func (r *OrderRepository) FindByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	var order domain.Order
	query := `SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE order_number = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &order, query, orderNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order by number: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	var orders []*domain.Order
	query := `SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`
	if err := r.db.SelectContext(ctx, &orders, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	for _, o := range orders {
		items, err := r.FindItemsByOrderID(ctx, o.ID)
		if err == nil {
			o.Items = items
		}
	}
	return orders, nil
}

func (r *OrderRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE user_id = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}
	return count, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus, version int) error {
	query := `UPDATE orders SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), id, version)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrConcurrentModification
	}
	return nil
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `UPDATE orders SET status = ?, total_amount = ?, shipping_address = ?, billing_address = ?, metadata = ?, version = version + 1, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query,
		order.Status, order.TotalAmount, order.ShippingAddress, order.BillingAddress,
		order.Metadata, time.Now().UTC(), order.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

func (r *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	var order domain.Order
	query := `SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE idempotency_key = ? AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &order, query, key); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by idempotency key: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) FindItemsByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	query := `SELECT id, order_id, product_id, sku_id, shop_id, quantity, unit_price, total_price, snapshot, created_at FROM order_items WHERE order_id = ? ORDER BY id ASC LIMIT 500`
	if err := r.db.SelectContext(ctx, &items, query, orderID); err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}
	return items, nil
}

func (r *OrderRepository) FindByParentOrderID(ctx context.Context, parentOrderID string) ([]*domain.Order, error) {
	var orders []*domain.Order
	query := `SELECT id, order_number, user_id, seller_id, status, total_amount, currency, shipping_address, billing_address, idempotency_key, snapshot_id, parent_order_id, metadata, version, created_at, updated_at, deleted_at FROM orders WHERE parent_order_id = ? AND deleted_at IS NULL`
	if err := r.db.SelectContext(ctx, &orders, query, parentOrderID); err != nil {
		return nil, fmt.Errorf("failed to find sub-orders: %w", err)
	}
	return orders, nil
}

// Outbox methods
func (r *OrderRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.AggregateType, event.AggregateID, event.EventType,
		event.Payload, event.CreatedAt, event.Processed,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	query := `SELECT event_id, aggregate_type, aggregate_id, event_type, status, error_message, retries, payload, created_at, processed FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT ?`
	if err := r.db.SelectContext(ctx, &events, query, limit); err != nil {
		return nil, fmt.Errorf("failed to get outbox events: %w", err)
	}
	return events, nil
}

func (r *OrderRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	query := `UPDATE outbox_events SET processed = TRUE WHERE event_id = ?`
	_, err := r.db.ExecContext(ctx, query, eventID)
	return err
}

// Lifecycle history
func (r *OrderRepository) SaveLifecycleEvent(ctx context.Context, event *domain.LifecycleEvent) error {
	query := `INSERT INTO order_lifecycle_history (id, order_id, from_state, to_state, transition_reason, actor_id, actor_type, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.OrderID, event.FromStatus, event.ToStatus,
		event.TransitionReason, event.ActorID, event.ActorType, event.Metadata, event.CreatedAt,
	)
	return err
}

func (r *OrderRepository) GetLifecycleHistory(ctx context.Context, orderID string) ([]*domain.LifecycleEvent, error) {
	var events []*domain.LifecycleEvent
	query := `SELECT id, order_id, from_state, to_state, transition_reason, actor_id, actor_type, metadata, created_at FROM order_lifecycle_history WHERE order_id = ? ORDER BY created_at ASC LIMIT 1000`
	if err := r.db.SelectContext(ctx, &events, query, orderID); err != nil {
		return nil, fmt.Errorf("failed to get lifecycle history: %w", err)
	}
	return events, nil
}

// Snapshot methods
func (r *OrderRepository) SaveSnapshot(ctx context.Context, snapshot *domain.OrderSnapshot) error {
	query := `INSERT INTO order_snapshots (id, order_id, snapshot_data, checksum, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		snapshot.ID, snapshot.OrderID, snapshot.SnapshotData, snapshot.Checksum, snapshot.CreatedAt,
	)
	return err
}

func (r *OrderRepository) GetSnapshot(ctx context.Context, snapshotID string) (*domain.OrderSnapshot, error) {
	var snapshot domain.OrderSnapshot
	query := `SELECT id, order_id, snapshot_data, checksum, created_at FROM order_snapshots WHERE id = ?`
	if err := r.db.GetContext(ctx, &snapshot, query, snapshotID); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return &snapshot, nil
}

// Cancellation methods
// ExecInTx executes a function within a database transaction with SERIALIZABLE isolation.
func (r *OrderRepository) ExecInTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func (r *OrderRepository) UpdateStatusInTx(ctx context.Context, tx *sqlx.Tx, id string, status domain.OrderStatus, version int) error {
	query := `UPDATE orders SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ? AND deleted_at IS NULL`
	result, err := tx.ExecContext(ctx, query, status, time.Now().UTC(), id, version)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrConcurrentModification
	}
	return nil
}

func (r *OrderRepository) SaveLifecycleEventInTx(ctx context.Context, tx *sqlx.Tx, event *domain.LifecycleEvent) error {
	query := `INSERT INTO order_lifecycle_history (id, order_id, from_state, to_state, transition_reason, actor_id, actor_type, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query,
		event.ID, event.OrderID, event.FromStatus, event.ToStatus,
		event.TransitionReason, event.ActorID, event.ActorType, event.Metadata, event.CreatedAt,
	)
	return err
}

func (r *OrderRepository) SaveCancellationInTx(ctx context.Context, tx *sqlx.Tx, cancellation *domain.OrderCancellation) error {
	query := `INSERT INTO order_cancellations (id, order_id, reason, cancelled_by, cancelled_by_type, compensation_status, refund_amount, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query,
		cancellation.ID, cancellation.OrderID, cancellation.Reason,
		cancellation.CancelledBy, cancellation.CancelledByType,
		cancellation.CompensationStatus, cancellation.RefundAmount,
		cancellation.Metadata, cancellation.CreatedAt,
	)
	return err
}

func (r *OrderRepository) SaveOutboxEventInTx(ctx context.Context, tx *sqlx.Tx, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query,
		event.ID, event.AggregateType, event.AggregateID, event.EventType,
		event.Payload, event.CreatedAt, event.Processed,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}

func (r *OrderRepository) SaveCancellation(ctx context.Context, cancellation *domain.OrderCancellation) error {
	query := `INSERT INTO order_cancellations (id, order_id, reason, cancelled_by, cancelled_by_type, compensation_status, refund_amount, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		cancellation.ID, cancellation.OrderID, cancellation.Reason,
		cancellation.CancelledBy, cancellation.CancelledByType,
		cancellation.CompensationStatus, cancellation.RefundAmount,
		cancellation.Metadata, cancellation.CreatedAt,
	)
	return err
}

func (r *OrderRepository) UpdateCancellationCompensation(ctx context.Context, cancellationID string, status domain.CompensationStatus) error {
	query := `UPDATE order_cancellations SET compensation_status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, cancellationID)
	return err
}

// Reconciliation methods
func (r *OrderRepository) SaveReconciliation(ctx context.Context, rec *domain.OrderReconciliation) error {
	query := `INSERT INTO order_reconciliation (id, order_id, reconciliation_type, status, last_checked_at, retry_count, max_retries, metadata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		rec.ID, rec.OrderID, rec.ReconciliationType, rec.Status,
		rec.LastCheckedAt, rec.RetryCount, rec.MaxRetries, rec.Metadata,
		rec.CreatedAt, rec.UpdatedAt,
	)
	return err
}

func (r *OrderRepository) UpdateReconciliationStatus(ctx context.Context, id string, status domain.ReconciliationStatus, retryCount int) error {
	query := `UPDATE order_reconciliation SET status = ?, retry_count = ?, last_checked_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, retryCount, time.Now().UTC(), time.Now().UTC(), id)
	return err
}

func (r *OrderRepository) GetPendingReconciliations(ctx context.Context, limit int) ([]*domain.OrderReconciliation, error) {
	var recs []*domain.OrderReconciliation
	query := `SELECT id, order_id, reconciliation_type, status, last_checked_at, retry_count, max_retries, metadata, created_at, updated_at FROM order_reconciliation WHERE status IN ('pending', 'in_progress') AND retry_count < max_retries ORDER BY created_at ASC LIMIT ?`
	if err := r.db.SelectContext(ctx, &recs, query, limit); err != nil {
		return nil, err
	}
	return recs, nil
}

// Seller split methods
func (r *OrderRepository) SaveSellerSplit(ctx context.Context, split *domain.SellerSplit) error {
	query := `INSERT INTO order_seller_splits (id, parent_order_id, seller_id, sub_order_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		split.ID, split.ParentOrderID, split.SellerID, split.SubOrderID,
		split.Status, split.CreatedAt, split.UpdatedAt,
	)
	return err
}

// Idempotency methods
func (r *OrderRepository) SaveIdempotencyKey(ctx context.Context, record *domain.IdempotencyRecord) error {
	query := "INSERT INTO idempotency_keys (`key`, order_id, expires_at, created_at) VALUES (?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, record.Key, record.OrderID, record.ExpiresAt, record.CreatedAt)
	return err
}

func (r *OrderRepository) GetIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	query := "SELECT `key`, order_id, expires_at, created_at FROM idempotency_keys WHERE `key` = ?"
	if err := r.db.GetContext(ctx, &record, query, key); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}
