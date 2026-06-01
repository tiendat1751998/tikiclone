package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/checkout/internal/domain"
)

type CheckoutRepository struct{ db *sqlx.DB }
func NewCheckoutRepository(db *sqlx.DB) *CheckoutRepository { return &CheckoutRepository{db: db} }

func (r *CheckoutRepository) FindByID(ctx context.Context, id string) (*domain.Checkout, error) {
	var c domain.Checkout
	err := r.db.GetContext(ctx, &c, `SELECT id, user_id, cart_id, order_id, status, idempotency_key, current_step, failure_reason, attempt_count, reservation_keys, pricing_snapshot, promotion_results, subtotal, discount_total, shipping_total, grand_total, currency, expires_at, completed_at, created_at, updated_at FROM checkouts WHERE id = ?`, id)
	if err == sql.ErrNoRows { return nil, nil }
	return &c, err
}

func (r *CheckoutRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Checkout, error) {
	var c domain.Checkout
	err := r.db.GetContext(ctx, &c, `SELECT id, user_id, cart_id, order_id, status, idempotency_key, current_step, failure_reason, attempt_count, reservation_keys, pricing_snapshot, promotion_results, subtotal, discount_total, shipping_total, grand_total, currency, expires_at, completed_at, created_at, updated_at FROM checkouts WHERE idempotency_key = ?`, key)
	if err == sql.ErrNoRows { return nil, nil }
	return &c, err
}

func (r *CheckoutRepository) Create(ctx context.Context, c *domain.Checkout) error {
	query := `INSERT INTO checkouts (id, user_id, cart_id, status, idempotency_key, current_step, attempt_count, subtotal, discount_total, shipping_total, grand_total, currency, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.UserID, c.CartID, c.Status, c.IdempotencyKey, c.CurrentStep, c.AttemptCount, c.Subtotal, c.DiscountTotal, c.ShippingTotal, c.GrandTotal, c.Currency, c.ExpiresAt, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *CheckoutRepository) Update(ctx context.Context, c *domain.Checkout) error {
	query := `UPDATE checkouts SET status = ?, current_step = ?, failure_reason = ?, attempt_count = ?, reservation_keys = ?, order_id = ?, subtotal = ?, discount_total = ?, shipping_total = ?, grand_total = ?, completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, c.Status, c.CurrentStep, c.FailureReason, c.AttemptCount, c.ReservationKeys, c.OrderID, c.Subtotal, c.DiscountTotal, c.ShippingTotal, c.GrandTotal, c.CompletedAt, c.UpdatedAt, c.ID)
	return err
}

func (r *CheckoutRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE checkouts SET status = ?, updated_at = NOW() WHERE id = ?", status, id)
	return err
}

func (r *CheckoutRepository) FindExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Checkout, error) {
	var checkouts []*domain.Checkout
	err := r.db.SelectContext(ctx, &checkouts, `SELECT id, user_id, cart_id, order_id, status, idempotency_key, current_step, failure_reason, attempt_count, reservation_keys, pricing_snapshot, promotion_results, subtotal, discount_total, shipping_total, grand_total, currency, expires_at, completed_at, created_at, updated_at FROM checkouts WHERE status IN ('pending','validating','pricing_frozen','reserving_inventory') AND expires_at < ? LIMIT ?`, before, limit)
	return checkouts, err
}

type CheckoutStepLogRepository struct{ db *sqlx.DB }
func NewCheckoutStepLogRepository(db *sqlx.DB) *CheckoutStepLogRepository { return &CheckoutStepLogRepository{db: db} }
func (r *CheckoutStepLogRepository) Create(ctx context.Context, log *domain.CheckoutStepLog) error {
	query := `INSERT INTO checkout_step_logs (id, checkout_id, step, status, error_message, metadata, duration_ms, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.CheckoutID, log.Step, log.Status, log.ErrorMsg, log.Metadata, log.DurationMs, log.CreatedAt)
	return err
}
func (r *CheckoutStepLogRepository) FindByCheckoutID(ctx context.Context, checkoutID string) ([]*domain.CheckoutStepLog, error) {
	var logs []*domain.CheckoutStepLog
	err := r.db.SelectContext(ctx, &logs, `SELECT id, checkout_id, step, status, error_message, metadata, duration_ms, created_at FROM checkout_step_logs WHERE checkout_id = ? ORDER BY created_at ASC`, checkoutID)
	return logs, err
}

type PricingSnapshotRepository struct{ db *sqlx.DB }
func NewPricingSnapshotRepository(db *sqlx.DB) *PricingSnapshotRepository { return &PricingSnapshotRepository{db: db} }
func (r *PricingSnapshotRepository) FindByCheckoutID(ctx context.Context, checkoutID string) (*domain.PricingSnapshot, error) {
	var snap domain.PricingSnapshot
	err := r.db.GetContext(ctx, &snap, `SELECT id, checkout_id, items, seller_groups, subtotal, discount_total, shipping_total, grand_total, currency, promotions_applied, created_at FROM pricing_snapshots WHERE checkout_id = ?`, checkoutID)
	if err == sql.ErrNoRows { return nil, nil }
	return &snap, err
}
func (r *PricingSnapshotRepository) Create(ctx context.Context, snap *domain.PricingSnapshot) error {
	query := `INSERT INTO pricing_snapshots (id, checkout_id, items, seller_groups, subtotal, discount_total, shipping_total, grand_total, currency, promotions_applied, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, snap.ID, snap.CheckoutID, snap.Items, snap.SellerGroups, snap.Subtotal, snap.DiscountTotal, snap.ShippingTotal, snap.GrandTotal, snap.Currency, snap.PromotionsApplied, snap.CreatedAt)
	return err
}

type ReservationOrchestrationRepository struct{ db *sqlx.DB }
func NewReservationOrchestrationRepository(db *sqlx.DB) *ReservationOrchestrationRepository {
	return &ReservationOrchestrationRepository{db: db}
}
func (r *ReservationOrchestrationRepository) FindByCheckoutID(ctx context.Context, checkoutID string) ([]*domain.ReservationOrchestration, error) {
	var res []*domain.ReservationOrchestration
	err := r.db.SelectContext(ctx, &res, `SELECT id, checkout_id, reservation_key, sku, warehouse_id, quantity, status, error_message, created_at, updated_at FROM reservation_orchestrations WHERE checkout_id = ?`, checkoutID)
	return res, err
}
func (r *ReservationOrchestrationRepository) Create(ctx context.Context, res *domain.ReservationOrchestration) error {
	query := `INSERT INTO reservation_orchestrations (id, checkout_id, reservation_key, sku, warehouse_id, quantity, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, res.ID, res.CheckoutID, res.ReservationKey, res.SKU, res.WarehouseID, res.Quantity, res.Status, res.CreatedAt, res.UpdatedAt)
	return err
}
func (r *ReservationOrchestrationRepository) UpdateStatus(ctx context.Context, key, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE reservation_orchestrations SET status = ?, updated_at = NOW() WHERE reservation_key = ?", status, key)
	return err
}

type ReconciliationJobRepository struct{ db *sqlx.DB }
func NewReconciliationJobRepository(db *sqlx.DB) *ReconciliationJobRepository { return &ReconciliationJobRepository{db: db} }
func (r *ReconciliationJobRepository) Create(ctx context.Context, job *domain.ReconciliationJob) error {
	query := `INSERT INTO reconciliation_jobs (id, checkout_id, job_type, status, attempt_count, max_attempts, next_retry_at, metadata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, job.ID, job.CheckoutID, job.JobType, job.Status, job.AttemptCount, job.MaxAttempts, job.NextRetryAt, job.Metadata, job.CreatedAt, job.UpdatedAt)
	return err
}
func (r *ReconciliationJobRepository) FindPending(ctx context.Context, limit int) ([]*domain.ReconciliationJob, error) {
	var jobs []*domain.ReconciliationJob
	err := r.db.SelectContext(ctx, &jobs, `SELECT id, checkout_id, job_type, status, attempt_count, max_attempts, next_retry_at, metadata, created_at, updated_at FROM reconciliation_jobs WHERE status = 'pending' AND next_retry_at <= NOW() ORDER BY next_retry_at ASC LIMIT ?`, limit)
	return jobs, err
}
func (r *ReconciliationJobRepository) Update(ctx context.Context, job *domain.ReconciliationJob) error {
	query := `UPDATE reconciliation_jobs SET status = ?, attempt_count = ?, next_retry_at = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, job.Status, job.AttemptCount, job.NextRetryAt, job.ID)
	return err
}
func (r *ReconciliationJobRepository) IncrementAttempt(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE reconciliation_jobs SET attempt_count = attempt_count + 1, updated_at = NOW() WHERE id = ?", id)
	return err
}
