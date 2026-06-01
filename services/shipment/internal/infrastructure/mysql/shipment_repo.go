package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/shipment/internal/domain"
)

type ShipmentRepository struct {
	db *sqlx.DB
}

func NewShipmentRepository(db *sqlx.DB) *ShipmentRepository { return &ShipmentRepository{db: db} }

func (r *ShipmentRepository) Create(ctx context.Context, s *domain.Shipment) error {
	query := `INSERT INTO shipments (id, order_id, user_id, carrier_id, tracking_number, status, origin_address, destination_address, weight, dimensions, label_url, cost, currency, idempotency_key, metadata, version, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.OrderID, s.UserID, s.CarrierID, s.TrackingNumber, s.Status, s.OriginAddress, s.DestAddress, s.Weight, s.Dimensions, s.LabelURL, s.Cost, s.Currency, s.IdempotencyKey, s.Metadata, s.Version, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *ShipmentRepository) FindByID(ctx context.Context, id string) (*domain.Shipment, error) {
	var s domain.Shipment
	if err := r.db.GetContext(ctx, &s, "SELECT id, order_id, user_id, carrier_id, tracking_number, status, origin_address, destination_address, weight, dimensions, label_url, cost, currency, idempotency_key, metadata, version, created_at, updated_at, deleted_at FROM shipments WHERE id = ? AND deleted_at IS NULL", id); err != nil {
		if err == sql.ErrNoRows { return nil, domain.ErrShipmentNotFound }
		return nil, err
	}
	return &s, nil
}

func (r *ShipmentRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Shipment, error) {
	var s domain.Shipment
	if err := r.db.GetContext(ctx, &s, "SELECT id, order_id, user_id, carrier_id, tracking_number, status, origin_address, destination_address, weight, dimensions, label_url, cost, currency, idempotency_key, metadata, version, created_at, updated_at, deleted_at FROM shipments WHERE order_id = ? AND deleted_at IS NULL", orderID); err != nil {
		if err == sql.ErrNoRows { return nil, domain.ErrShipmentNotFound }
		return nil, err
	}
	return &s, nil
}

func (r *ShipmentRepository) UpdateStatus(ctx context.Context, id string, status domain.ShipmentStatus, version int) error {
	result, err := r.db.ExecContext(ctx, "UPDATE shipments SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?", status, time.Now().UTC(), id, version)
	if err != nil { return err }
	rows, _ := result.RowsAffected()
	if rows == 0 { return domain.ErrConcurrentModification }
	return nil
}

func (r *ShipmentRepository) Update(ctx context.Context, s *domain.Shipment) error {
	query := `UPDATE shipments SET status = ?, tracking_number = ?, label_url = ?, cost = ?, metadata = ?, version = version + 1, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, s.Status, s.TrackingNumber, s.LabelURL, s.Cost, s.Metadata, time.Now().UTC(), s.ID)
	return err
}

func (r *ShipmentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Shipment, error) {
	var s domain.Shipment
	if err := r.db.GetContext(ctx, &s, "SELECT id, order_id, user_id, carrier_id, tracking_number, status, origin_address, destination_address, weight, dimensions, label_url, cost, currency, idempotency_key, metadata, version, created_at, updated_at, deleted_at FROM shipments WHERE idempotency_key = ?", key); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}
	return &s, nil
}

func (r *ShipmentRepository) SaveTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	query := `INSERT INTO tracking_events (id, shipment_id, status, location, description, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.ShipmentID, event.Status, event.Location, event.Description, event.Timestamp, event.CreatedAt)
	return err
}

func (r *ShipmentRepository) GetTrackingHistory(ctx context.Context, shipmentID string) ([]*domain.TrackingEvent, error) {
	var events []*domain.TrackingEvent
	err := r.db.SelectContext(ctx, &events, "SELECT id, shipment_id, status, location, description, timestamp, created_at FROM tracking_events WHERE shipment_id = ? ORDER BY timestamp ASC LIMIT 200", shipmentID)
	return events, err
}

func (r *ShipmentRepository) SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	query := `INSERT INTO outbox_events (event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.AggregateType, event.AggregateID, event.EventType, event.Payload, event.CreatedAt, event.Processed)
	return err
}

func (r *ShipmentRepository) GetUnprocessedOutboxEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent
	err := r.db.SelectContext(ctx, &events, "SELECT event_id, aggregate_type, aggregate_id, event_type, payload, created_at, processed FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT ?", limit)
	return events, err
}

func (r *ShipmentRepository) MarkOutboxEventProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox_events SET processed = TRUE WHERE event_id = ?", eventID)
	return err
}

func (r *ShipmentRepository) SaveIdempotencyKey(ctx context.Context, record *domain.IdempotencyRecord) error {
	query := `INSERT INTO idempotency_keys (` + "`key`" + `, shipment_id, expires_at, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, record.Key, record.ShipmentID, record.ExpiresAt, record.CreatedAt)
	return err
}

func (r *ShipmentRepository) GetIdempotencyKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var record domain.IdempotencyRecord
	if err := r.db.GetContext(ctx, &record, "SELECT `key`, shipment_id, expires_at, created_at FROM idempotency_keys WHERE `key` = ?", key); err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}
	return &record, nil
}

func (r *ShipmentRepository) IsWebhookProcessed(ctx context.Context, key string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM webhook_events WHERE idempotency_key = ?", key)
	return count > 0, err
}

func (r *ShipmentRepository) SaveWebhookEvent(ctx context.Context, provider, eventType string, payload []byte, signature, idempotencyKey string) error {
	query := `INSERT INTO webhook_events (id, provider, event_type, payload, signature, idempotency_key, created_at) VALUES (?, ?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, idempotencyKey[:36], provider, eventType, payload, signature, idempotencyKey)
	return err
}

// --- QR Code methods ---

func (r *ShipmentRepository) CreateQRCode(ctx context.Context, qr *domain.QRCode) error {
	query := `INSERT INTO qr_codes (id, shipment_id, type, code, status, signed_token, expires_at, scanned_at, scanned_by, scan_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, qr.ID, qr.ShipmentID, qr.Type, qr.Code, qr.Status, qr.SignedToken, qr.ExpiresAt, qr.ScannedAt, qr.ScannedBy, qr.ScanCount, qr.CreatedAt, qr.UpdatedAt)
	return err
}

func (r *ShipmentRepository) FindQRCodeByCode(ctx context.Context, code string) (*domain.QRCode, error) {
	var qr domain.QRCode
	if err := r.db.GetContext(ctx, &qr, "SELECT id, shipment_id, type, code, status, signed_token, expires_at, scanned_at, scanned_by, scan_count, created_at, updated_at FROM qr_codes WHERE code = ?", code); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrQRCodeNotFound
		}
		return nil, err
	}
	return &qr, nil
}

func (r *ShipmentRepository) FindQRCodeByShipmentAndType(ctx context.Context, shipmentID string, qrType domain.QRCodeType) (*domain.QRCode, error) {
	var qr domain.QRCode
	if err := r.db.GetContext(ctx, &qr, "SELECT id, shipment_id, type, code, status, signed_token, expires_at, scanned_at, scanned_by, scan_count, created_at, updated_at FROM qr_codes WHERE shipment_id = ? AND type = ?", shipmentID, qrType); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrQRCodeNotFound
		}
		return nil, err
	}
	return &qr, nil
}

func (r *ShipmentRepository) FindQRCodesByShipment(ctx context.Context, shipmentID string) ([]*domain.QRCode, error) {
	var codes []*domain.QRCode
	err := r.db.SelectContext(ctx, &codes, "SELECT id, shipment_id, type, code, status, signed_token, expires_at, scanned_at, scanned_by, scan_count, created_at, updated_at FROM qr_codes WHERE shipment_id = ? ORDER BY created_at DESC", shipmentID)
	return codes, err
}

func (r *ShipmentRepository) UpdateQRCodeStatus(ctx context.Context, id string, status domain.QRCodeStatus, scannedAt *time.Time, scannedBy string, scanCount int32) error {
	query := `UPDATE qr_codes SET status = ?, scanned_at = ?, scanned_by = ?, scan_count = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, scannedAt, scannedBy, scanCount, time.Now().UTC(), id)
	return err
}

func (r *ShipmentRepository) RevokeQRCode(ctx context.Context, id string) error {
	query := `UPDATE qr_codes SET status = 'revoked', updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *ShipmentRepository) ExpireOldQRCodes(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE qr_codes SET status = 'expired', updated_at = ? WHERE expires_at < ? AND status = 'active'", time.Now().UTC(), before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// --- Scan Event methods ---

func (r *ShipmentRepository) SaveScanEvent(ctx context.Context, event *domain.ScanEvent) error {
	query := `INSERT INTO scan_events (id, qr_code_id, shipment_id, shipper_id, shipper_name, shipper_role, scan_type, latitude, longitude, device_info, ip_address, is_valid, fail_reason, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.QRCodeID, event.ShipmentID, event.ShipperID, event.ShipperName, event.ShipperRole, event.ScanType, event.Latitude, event.Longitude, event.DeviceInfo, event.IPAddress, event.IsValid, event.FailReason, event.CreatedAt)
	return err
}

func (r *ShipmentRepository) GetScanHistory(ctx context.Context, shipmentID string) ([]*domain.ScanEvent, error) {
	var events []*domain.ScanEvent
	err := r.db.SelectContext(ctx, &events, "SELECT id, qr_code_id, shipment_id, shipper_id, shipper_name, shipper_role, scan_type, latitude, longitude, device_info, ip_address, is_valid, fail_reason, created_at FROM scan_events WHERE shipment_id = ? ORDER BY created_at DESC LIMIT 200", shipmentID)
	return events, err
}

func (r *ShipmentRepository) GetScanHistoryByShipper(ctx context.Context, shipperID string, limit int) ([]*domain.ScanEvent, error) {
	var events []*domain.ScanEvent
	err := r.db.SelectContext(ctx, &events, "SELECT id, qr_code_id, shipment_id, shipper_id, shipper_name, shipper_role, scan_type, latitude, longitude, device_info, ip_address, is_valid, fail_reason, created_at FROM scan_events WHERE shipper_id = ? ORDER BY created_at DESC LIMIT ?", shipperID, limit)
	return events, err
}
