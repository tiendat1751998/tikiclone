package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/domain"
)

type FulfillmentRepository struct {
	db *sqlx.DB
}

func NewFulfillmentRepository(db *sqlx.DB) *FulfillmentRepository {
	return &FulfillmentRepository{db: db}
}

func (r *FulfillmentRepository) Create(ctx context.Context, f *domain.FulfillmentOrder) error {
	addrJSON, _ := json.Marshal(f.ShippingAddress)
	_, err := r.db.ExecContext(ctx, `INSERT INTO fulfillment_orders
		(id, order_id, seller_id, warehouse_id, status, priority, shipping_method, shipping_address,
		 estimated_ship_date, estimated_delivery_date, total_items, total_weight_grams, notes, metadata,
		 created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		f.ID, f.OrderID, f.SellerID, f.WarehouseID, f.Status, f.Priority, f.ShippingMethod, string(addrJSON),
		f.EstimatedShipDate, f.EstimatedDeliveryDate, f.TotalItems, f.TotalWeightGrams, f.Notes, f.MetadataJSON,
		f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *FulfillmentRepository) GetByID(ctx context.Context, id string) (*domain.FulfillmentOrder, error) {
	f := &domain.FulfillmentOrder{}
	err := r.db.GetContext(ctx, f, `SELECT * FROM fulfillment_orders WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get fulfillment %s: %w", id, mapError(err))
	}
	if f.ShippingAddressJSON != "" {
		var addr domain.ShippingAddress
		if err := json.Unmarshal([]byte(f.ShippingAddressJSON), &addr); err == nil {
			f.ShippingAddress = &addr
		}
	}
	return f, nil
}

func (r *FulfillmentRepository) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]domain.FulfillmentOrder, int, error) {
	query := "SELECT * FROM fulfillment_orders"
	countQuery := "SELECT COUNT(*) FROM fulfillment_orders"
	var args []interface{}
	var where string

	if status, ok := filter["status"]; ok {
		where = " WHERE status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query += where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	var items []domain.FulfillmentOrder
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *FulfillmentRepository) UpdateStatus(ctx context.Context, id string, status domain.FulfillmentStatus) error {
	res, err := r.db.ExecContext(ctx, "UPDATE fulfillment_orders SET status = ?, updated_at = ? WHERE id = ?", status, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrFulfillmentNotFound
	}
	return nil
}

func (r *FulfillmentRepository) CreateItem(ctx context.Context, item *domain.FulfillmentItem) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO fulfillment_items
		(id, fulfillment_id, order_item_id, product_id, sku_id, quantity,
		 picked_quantity, packed_quantity, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.FulfillmentID, item.OrderItemID, item.ProductID, item.SKUID, item.Quantity,
		item.PickedQuantity, item.PackedQuantity, item.Status, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *FulfillmentRepository) GetItems(ctx context.Context, fulfillmentID string) ([]domain.FulfillmentItem, error) {
	var items []domain.FulfillmentItem
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM fulfillment_items WHERE fulfillment_id = ? ORDER BY created_at", fulfillmentID)
	return items, err
}

func (r *FulfillmentRepository) SaveEvent(ctx context.Context, event *domain.FulfillmentEvent) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO fulfillment_events
		(id, fulfillment_id, event_type, description, actor_id, created_at)
		VALUES (?,?,?,?,?,?)`,
		event.ID, event.FulfillmentID, event.EventType, event.Description, event.ActorID, time.Now().UTC())
	return err
}

func (r *FulfillmentRepository) GetEvents(ctx context.Context, fulfillmentID string) ([]domain.FulfillmentEvent, error) {
	var events []domain.FulfillmentEvent
	err := r.db.SelectContext(ctx, &events, "SELECT * FROM fulfillment_events WHERE fulfillment_id = ? ORDER BY created_at", fulfillmentID)
	return events, err
}

func (r *FulfillmentRepository) CreateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO warehouses
		(id, name, code, address, contact_phone, contact_email, capacity_total, capacity_used,
		 operating_hours, timezone, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		w.ID, w.Name, w.Code, w.AddressJSON, w.ContactPhone, w.ContactEmail, w.CapacityTotal, w.CapacityUsed,
		nil, "UTC", w.IsActive, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *FulfillmentRepository) GetWarehouse(ctx context.Context, id string) (*domain.Warehouse, error) {
	w := &domain.Warehouse{}
	err := r.db.GetContext(ctx, w, "SELECT * FROM warehouses WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return w, nil
}

func (r *FulfillmentRepository) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	var warehouses []domain.Warehouse
	err := r.db.SelectContext(ctx, &warehouses, "SELECT * FROM warehouses ORDER BY name")
	return warehouses, err
}

func (r *FulfillmentRepository) GetWarehouseInventory(ctx context.Context, warehouseID string) ([]domain.WarehouseInventory, error) {
	var inv []domain.WarehouseInventory
	err := r.db.SelectContext(ctx, &inv, "SELECT * FROM warehouse_inventory WHERE warehouse_id = ?", warehouseID)
	return inv, err
}

func (r *FulfillmentRepository) CreateReturn(ctx context.Context, ret *domain.ReturnRequest) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO return_requests
		(id, fulfillment_id, order_id, user_id, return_type, reason, reason_detail, status,
		 refund_amount, currency, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		ret.ID, ret.FulfillmentID, ret.OrderID, ret.UserID, ret.ReturnType, ret.Reason, ret.ReasonDetail,
		ret.Status, ret.RefundAmount, ret.Currency, ret.CreatedAt, ret.UpdatedAt)
	return err
}

func (r *FulfillmentRepository) GetReturn(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	ret := &domain.ReturnRequest{}
	err := r.db.GetContext(ctx, ret, "SELECT * FROM return_requests WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return ret, nil
}

func (r *FulfillmentRepository) ListReturns(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]domain.ReturnRequest, int, error) {
	query := "SELECT * FROM return_requests"
	countQuery := "SELECT COUNT(*) FROM return_requests"
	var args []interface{}
	var where string

	if status, ok := filter["status"]; ok {
		where = " WHERE status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query += where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	var items []domain.ReturnRequest
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *FulfillmentRepository) UpdateReturnStatus(ctx context.Context, id string, status domain.ReturnStatus, approvedBy *string) error {
	now := time.Now().UTC()
	var err error
	switch status {
	case domain.ReturnStatusApproved:
		_, err = r.db.ExecContext(ctx, "UPDATE return_requests SET status = ?, approved_by = ?, approved_at = ?, updated_at = ? WHERE id = ?",
			status, approvedBy, now, now, id)
	case domain.ReturnStatusRejected:
		_, err = r.db.ExecContext(ctx, "UPDATE return_requests SET status = ?, approved_by = ?, updated_at = ? WHERE id = ?",
			status, approvedBy, now, id)
	default:
		_, err = r.db.ExecContext(ctx, "UPDATE return_requests SET status = ?, updated_at = ? WHERE id = ?",
			status, now, id)
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *FulfillmentRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO outbox_events
		(id, event_type, payload, status, retry_count, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`,
		event.ID, event.EventType, event.Payload, domain.OutboxStatusPending, 0, time.Now().UTC(), time.Now().UTC())
	return err
}

func (r *FulfillmentRepository) GetUnprocessedOutboxEvents(ctx context.Context) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT * FROM outbox_events WHERE status IN (?, ?) AND retry_count < 3 ORDER BY created_at LIMIT 100`,
		domain.OutboxStatusPending, domain.OutboxStatusFailed)
	return events, err
}

func (r *FulfillmentRepository) MarkOutboxEventProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET status = ?, updated_at = ? WHERE id = ?",
		domain.OutboxStatusProcessed, time.Now().UTC(), id)
	return err
}

func (r *FulfillmentRepository) MarkOutboxEventFailed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET status = ?, retry_count = retry_count + 1, updated_at = ? WHERE id = ?",
		domain.OutboxStatusFailed, time.Now().UTC(), id)
	return err
}

func ExecInTx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original: %w)", rbErr, err)
		}
		return err
	}
	return tx.Commit()
}

func mapError(err error) error {
	if err != nil && err.Error() == "sql: no rows in result set" {
		return domain.ErrFulfillmentNotFound
	}
	return err
}
