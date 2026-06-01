package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
)

type InventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// GetStock retrieves stock by SKU and warehouse (no lock).
func (r *InventoryRepository) GetStock(ctx context.Context, skuID, warehouseID string) (*domain.Stock, error) {
	var s domain.Stock
	err := r.db.GetContext(ctx, &s, "SELECT id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level, version, created_at, updated_at FROM stock WHERE sku_id = ? AND warehouse_id = ?", skuID, warehouseID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrStockNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get stock: %w", err)
	}
	return &s, nil
}

// GetStockForUpdate retrieves stock with a row-level lock (SELECT ... FOR UPDATE).
// Must be called within a transaction.
func (r *InventoryRepository) GetStockForUpdate(ctx context.Context, tx *sql.Tx, skuID, warehouseID string) (*domain.Stock, error) {
	var s domain.Stock
	err := tx.QueryRowContext(ctx, "SELECT * FROM stock WHERE sku_id = ? AND warehouse_id = ? FOR UPDATE", skuID, warehouseID).Scan(
		&s.ID, &s.ProductID, &s.SkuID, &s.WarehouseID, &s.Quantity, &s.ReservedQty,
		&s.AvailableQty, &s.Status, &s.ReorderLevel, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrStockNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get stock for update: %w", err)
	}
	return &s, nil
}

// UpdateStock updates stock with optimistic locking (version check).
func (r *InventoryRepository) UpdateStock(ctx context.Context, stock *domain.Stock) error {
	query := `UPDATE stock SET quantity = ?, reserved_qty = ?, available_qty = ?, status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`
	result, err := r.db.ExecContext(ctx, query, stock.Quantity, stock.ReservedQty, stock.AvailableQty, stock.Status, time.Now().UTC(), stock.ID, stock.Version)
	if err != nil {
		return fmt.Errorf("update stock: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrConcurrentModification
	}
	stock.Version++
	return nil
}

// UpdateStockInTx updates stock within a transaction.
func (r *InventoryRepository) UpdateStockInTx(ctx context.Context, tx *sql.Tx, stock *domain.Stock) error {
	query := `UPDATE stock SET quantity = ?, reserved_qty = ?, available_qty = ?, status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`
	result, err := tx.ExecContext(ctx, query, stock.Quantity, stock.ReservedQty, stock.AvailableQty, stock.Status, time.Now().UTC(), stock.ID, stock.Version)
	if err != nil {
		return fmt.Errorf("update stock in tx: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrConcurrentModification
	}
	stock.Version++
	return nil
}

// CreateStock inserts a new stock record.
func (r *InventoryRepository) CreateStock(ctx context.Context, stock *domain.Stock) error {
	query := `INSERT INTO stock (id, product_id, sku_id, warehouse_id, quantity, reserved_qty, available_qty, status, reorder_level, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, stock.ID, stock.ProductID, stock.SkuID, stock.WarehouseID, stock.Quantity, stock.ReservedQty, stock.AvailableQty, stock.Status, stock.ReorderLevel, stock.Version, stock.CreatedAt, stock.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create stock: %w", err)
	}
	return nil
}

// SaveReservation inserts a new reservation record.
func (r *InventoryRepository) SaveReservation(ctx context.Context, res *domain.Reservation) error {
	query := `INSERT INTO reservations (id, order_id, user_id, product_id, sku_id, warehouse_id, quantity, status, expires_at, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, res.ID, res.OrderID, res.UserID, res.ProductID, res.SkuID, res.WarehouseID, res.Quantity, res.Status, res.ExpiresAt, res.IdempotencyKey, res.CreatedAt, res.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save reservation: %w", err)
	}
	return nil
}

// SaveReservationInTx inserts a reservation within a transaction.
func (r *InventoryRepository) SaveReservationInTx(ctx context.Context, tx *sql.Tx, res *domain.Reservation) error {
	query := `INSERT INTO reservations (id, order_id, user_id, product_id, sku_id, warehouse_id, quantity, status, expires_at, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, res.ID, res.OrderID, res.UserID, res.ProductID, res.SkuID, res.WarehouseID, res.Quantity, res.Status, res.ExpiresAt, res.IdempotencyKey, res.CreatedAt, res.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save reservation in tx: %w", err)
	}
	return nil
}

// GetReservation retrieves a reservation by ID.
func (r *InventoryRepository) GetReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	var res domain.Reservation
	err := r.db.GetContext(ctx, &res, "SELECT id, order_id, user_id, product_id, sku_id, warehouse_id, quantity, status, expires_at, idempotency_key, created_at, updated_at FROM reservations WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, domain.ErrReservationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	return &res, nil
}

// GetReservationForUpdate retrieves a reservation with a row-level lock.
func (r *InventoryRepository) GetReservationForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Reservation, error) {
	var res domain.Reservation
	err := tx.QueryRowContext(ctx, "SELECT * FROM reservations WHERE id = ? FOR UPDATE", id).Scan(
		&res.ID, &res.OrderID, &res.UserID, &res.ProductID, &res.SkuID, &res.WarehouseID,
		&res.Quantity, &res.Status, &res.ExpiresAt, &res.IdempotencyKey, &res.CreatedAt, &res.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrReservationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reservation for update: %w", err)
	}
	return &res, nil
}

// UpdateReservationStatus updates the status of a reservation.
func (r *InventoryRepository) UpdateReservationStatus(ctx context.Context, id string, status domain.ReservationStatus) error {
	_, err := r.db.ExecContext(ctx, "UPDATE reservations SET status = ?, updated_at = ? WHERE id = ?", status, time.Now().UTC(), id)
	return err
}

// UpdateReservationStatusInTx updates reservation status within a transaction.
func (r *InventoryRepository) UpdateReservationStatusInTx(ctx context.Context, tx *sql.Tx, id string, status domain.ReservationStatus) error {
	_, err := tx.ExecContext(ctx, "UPDATE reservations SET status = ?, updated_at = ? WHERE id = ?", status, time.Now().UTC(), id)
	return err
}

// GetExpiredReservations finds reservations that have expired.
func (r *InventoryRepository) GetExpiredReservations(ctx context.Context, limit int) ([]*domain.Reservation, error) {
	var res []*domain.Reservation
	err := r.db.SelectContext(ctx, &res, "SELECT id, order_id, user_id, product_id, sku_id, warehouse_id, quantity, status, expires_at, idempotency_key, created_at, updated_at FROM reservations WHERE status = 'active' AND expires_at < NOW() LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("get expired reservations: %w", err)
	}
	return res, nil
}

// FindByIdempotencyKey finds a reservation by its idempotency key.
func (r *InventoryRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Reservation, error) {
	var res domain.Reservation
	err := r.db.GetContext(ctx, &res, "SELECT id, order_id, user_id, product_id, sku_id, warehouse_id, quantity, status, expires_at, idempotency_key, created_at, updated_at FROM reservations WHERE idempotency_key = ?", key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find by idempotency key: %w", err)
	}
	return &res, nil
}

// SaveOutboxEvent stores an event in the outbox table.
func (r *InventoryRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.AggregateType, event.AggregateID, event.EventType, event.Payload, event.CreatedAt, event.Processed)
	if err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}
	return nil
}

// GetUnprocessedOutboxEvents fetches events that haven't been published yet.
func (r *InventoryRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	err := r.db.SelectContext(ctx, &events, "SELECT event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("get unprocessed outbox events: %w", err)
	}
	return events, nil
}

// MarkOutboxEventProcessed marks an event as successfully published.
func (r *InventoryRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE WHERE event_id = ?", eventID)
	return err
}

// MarkOutboxEventsProcessed marks multiple events as successfully published in one query.
func (r *InventoryRepository) MarkOutboxEventsProcessed(ctx context.Context, eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In("UPDATE outbox_events SET processed = TRUE WHERE event_id IN (?)", eventIDs)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// MarkOutboxEventProcessing marks an event as being processed (prevents duplicate processing).
func (r *InventoryRepository) MarkOutboxEventProcessing(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE, processing_at = NOW() WHERE event_id = ? AND processed = FALSE", eventID)
	return err
}

// MarkOutboxEventFailed marks an event as failed with error message.
func (r *InventoryRepository) MarkOutboxEventFailed(ctx context.Context, eventID, errorMsg string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET error_message = ? WHERE event_id = ?", errorMsg, eventID)
	return err
}

// SaveIdempotencyKey stores an idempotency record.
func (r *InventoryRepository) SaveIdempotencyKey(ctx context.Context, record *domain.IdempotencyRecord) error {
	query := "INSERT INTO idempotency_keys (`key`, reservation_id, expires_at, created_at) VALUES (?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, record.Key, record.ReservationID, record.ExpiresAt, record.CreatedAt)
	if err != nil {
		return fmt.Errorf("save idempotency key: %w", err)
	}
	return nil
}

// GetIdempotencyKey retrieves an idempotency record by key.
func (r *InventoryRepository) GetIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	err := r.db.GetContext(ctx, &record, "SELECT `key`, reservation_id, expires_at, created_at FROM idempotency_keys WHERE `key` = ?", key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get idempotency key: %w", err)
	}
	return &record, nil
}

// ExecInTx executes a function within a database transaction.
// This is the key method that prevents race conditions in ReserveStock/ReleaseStock.
func (r *InventoryRepository) ExecInTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
