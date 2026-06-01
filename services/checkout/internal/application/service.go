package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/checkout/internal/domain"
	"github.com/tikiclone/tiki/services/checkout/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/checkout/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// CheckoutService orchestrates the checkout process with saga-like coordination
type CheckoutService struct {
	checkoutRepo    domain.CheckoutRepository
	stepLogRepo     domain.CheckoutStepLogRepository
	pricingRepo     domain.PricingSnapshotRepository
	reservationRepo domain.ReservationOrchestrationRepository
	reconcileRepo   domain.ReconciliationJobRepository
	redis           *redis.Store
	snapshotTTL     time.Duration
	reservationTTL  time.Duration
	idempotencyTTL  time.Duration
	maxRetries      int
	publisher       EventPublisher
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}

func NewCheckoutService(
	checkoutRepo domain.CheckoutRepository,
	stepLogRepo domain.CheckoutStepLogRepository,
	pricingRepo domain.PricingSnapshotRepository,
	reservationRepo domain.ReservationOrchestrationRepository,
	reconcileRepo domain.ReconciliationJobRepository,
	redisStore *redis.Store,
	snapshotTTL, reservationTTL, idempotencyTTL time.Duration,
	maxRetries int,
	publisher EventPublisher,
) *CheckoutService {
	return &CheckoutService{
		checkoutRepo: checkoutRepo, stepLogRepo: stepLogRepo,
		pricingRepo: pricingRepo, reservationRepo: reservationRepo,
		reconcileRepo: reconcileRepo, redis: redisStore,
		snapshotTTL: snapshotTTL, reservationTTL: reservationTTL,
		idempotencyTTL: idempotencyTTL, maxRetries: maxRetries, publisher: publisher,
	}
}

// InitiateCheckout starts the checkout orchestration process
func (s *CheckoutService) InitiateCheckout(ctx context.Context, req InitiateRequest) (*domain.Checkout, error) {
	ctx, span := otel.Tracer("tiki-checkout").Start(ctx, "checkout.initiate")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", req.UserID),
		attribute.String("cart_id", req.CartId),
	)

	// Idempotency check
	if req.IdempotencyKey != "" {
		existing, err := s.checkoutRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil && existing.Status != domain.CheckoutStatusFailed && existing.Status != domain.CheckoutStatusRolledBack {
			metrics.IdempotentRequests.Inc()
			observability.LogWithTrace(ctx).Info("idempotent checkout request",
				zap.String("checkout_id", existing.ID))
			return existing, nil
		}
	}

	checkout := domain.NewCheckout(req.UserID, req.CartId, req.IdempotencyKey, s.snapshotTTL)
	checkout.Currency = req.Currency

	if err := s.checkoutRepo.Create(ctx, checkout); err != nil {
		return nil, fmt.Errorf("create checkout: %w", err)
	}

	// Execute saga steps asynchronously with timeout.
	// Use context.WithoutCancel to keep tracing context but remain alive after client disconnect.
	// The saga gets its own timeout and is not affected by client cancellation.
	sagaCtx, sagaCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	go func() {
		defer sagaCancel()
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error("panic in checkout saga", zap.Any("recover", r), zap.String("checkout_id", checkout.ID))
			}
		}()
		s.executeSaga(sagaCtx, checkout.ID, req)
	}()

	metrics.CheckoutsInitiated.Inc()
	return checkout, nil
}

// executeSaga runs the checkout orchestration steps sequentially with rollback support
func (s *CheckoutService) executeSaga(ctx context.Context, checkoutID string, req InitiateRequest) {
	logger := observability.LogWithTrace(ctx)
	checkout, err := s.checkoutRepo.FindByID(ctx, checkoutID)
	if err != nil {
		logger.Error("saga: failed to load checkout", zap.Error(err))
		return
	}

	// Step 1: Validate
	if err := s.stepValidate(ctx, checkout, req); err != nil {
		s.handleFailure(ctx, checkout, domain.StepValidate, err)
		return
	}

	// Step 2: Freeze Pricing
	if err := s.stepFreezePricing(ctx, checkout, req); err != nil {
		s.handleFailure(ctx, checkout, domain.StepFreezePricing, err)
		return
	}

	// Step 3: Reserve Inventory
	reservationKeys, err := s.stepReserveInventory(ctx, checkout, req)
	if err != nil {
		s.handleFailure(ctx, checkout, domain.StepReserve, err)
		return
	}

	// Step 4: Process (create order, trigger payment)
	if err := s.stepProcess(ctx, checkout, req); err != nil {
		// Rollback reservations
		s.rollbackReservations(ctx, checkout, reservationKeys)
		s.handleFailure(ctx, checkout, domain.StepProcess, err)
		return
	}

	// Step 5: Complete
	s.stepComplete(ctx, checkout, reservationKeys)

	metrics.CheckoutsCompleted.Inc()
}

func (s *CheckoutService) stepValidate(ctx context.Context, checkout *domain.Checkout, req InitiateRequest) error {
	start := time.Now()
	defer func() { metrics.CheckoutLatency.WithLabelValues(domain.StepValidate).Observe(time.Since(start).Seconds()) }()
	checkout.AdvanceStep(domain.StepValidate)
	checkout.Status = domain.CheckoutStatusValidating
	s.checkoutRepo.Update(ctx, checkout)

	// Validate cart exists and is active
	// In production: call Cart Service gRPC to validate
	if req.CartId == "" || req.UserID == "" {
		s.logStep(ctx, checkout.ID, domain.StepValidate, "failed", "missing cart_id or user_id", time.Since(start).Milliseconds())
		return fmt.Errorf("invalid request: cart_id and user_id required")
	}

	s.logStep(ctx, checkout.ID, domain.StepValidate, "success", "", time.Since(start).Milliseconds())
	metrics.CheckoutLatency.WithLabelValues(domain.StepValidate).Observe(time.Since(start).Seconds())
	return nil
}

func (s *CheckoutService) stepFreezePricing(ctx context.Context, checkout *domain.Checkout, req InitiateRequest) error {
	start := time.Now()
	defer func() { metrics.CheckoutLatency.WithLabelValues(domain.StepFreezePricing).Observe(time.Since(start).Seconds()) }()
	checkout.AdvanceStep(domain.StepFreezePricing)
	checkout.Status = domain.CheckoutStatusPricingFrozen
	s.checkoutRepo.Update(ctx, checkout)

	itemsJSON, err := mustJSON(req.Items)
	if err != nil {
		return fmt.Errorf("marshal items: %w", err)
	}
	sellerGroupsJSON, err := mustJSON(req.SellerGroups)
	if err != nil {
		return fmt.Errorf("marshal seller groups: %w", err)
	}
	promotionsJSON, err := mustJSON(req.Promotions)
	if err != nil {
		return fmt.Errorf("marshal promotions: %w", err)
	}
	snapshot := &domain.PricingSnapshot{
		ID:                fmt.Sprintf("snap_%d", time.Now().UnixNano()),
		CheckoutID:        checkout.ID,
		Items:             itemsJSON,
		SellerGroups:      sellerGroupsJSON,
		Subtotal:          req.Subtotal,
		DiscountTotal:     req.DiscountTotal,
		ShippingTotal:     req.ShippingTotal,
		GrandTotal:        req.GrandTotal,
		Currency:          checkout.Currency,
		PromotionsApplied: promotionsJSON,
		CreatedAt:         time.Now(),
	}

	if err := s.pricingRepo.Create(ctx, snapshot); err != nil {
		s.logStep(ctx, checkout.ID, domain.StepFreezePricing, "failed", err.Error(), time.Since(start).Milliseconds())
		return fmt.Errorf("freeze pricing: %w", err)
	}

	checkout.PricingSnapshot = snapshot.ID
	checkout.Subtotal = req.Subtotal
	checkout.DiscountTotal = req.DiscountTotal
	checkout.ShippingTotal = req.ShippingTotal
	checkout.GrandTotal = req.GrandTotal
	s.checkoutRepo.Update(ctx, checkout)

	s.logStep(ctx, checkout.ID, domain.StepFreezePricing, "success", "", time.Since(start).Milliseconds())
	metrics.CheckoutLatency.WithLabelValues(domain.StepFreezePricing).Observe(time.Since(start).Seconds())
	return nil
}

func (s *CheckoutService) stepReserveInventory(ctx context.Context, checkout *domain.Checkout, req InitiateRequest) ([]string, error) {
	start := time.Now()
	defer func() { metrics.CheckoutLatency.WithLabelValues(domain.StepReserve).Observe(time.Since(start).Seconds()) }()
	checkout.AdvanceStep(domain.StepReserve)
	checkout.Status = domain.CheckoutStatusReserving
	s.checkoutRepo.Update(ctx, checkout)

	var reservationKeys []string
	for i, item := range req.Items {
		resKey := fmt.Sprintf("res_%s_%s", checkout.ID, item.SKU)
		res := &domain.ReservationOrchestration{
			ID:             fmt.Sprintf("ro_%d_%d", time.Now().UnixNano(), i),
			CheckoutID:     checkout.ID,
			ReservationKey: resKey,
			SKU:            item.SKU,
			WarehouseID:    item.WarehouseID,
			Quantity:       int64(item.Quantity),
			Status:         domain.ReservationStatusReserved,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := s.reservationRepo.Create(ctx, res); err != nil {
			s.logStep(ctx, checkout.ID, domain.StepReserve, "failed", err.Error(), time.Since(start).Milliseconds())
			return reservationKeys, fmt.Errorf("reserve inventory for SKU %s: %w", item.SKU, err)
		}
		reservationKeys = append(reservationKeys, resKey)
	}

	checkout.Status = domain.CheckoutStatusReserved
	resKeysJSON, err := mustJSON(reservationKeys)
	if err != nil {
		return reservationKeys, fmt.Errorf("marshal reservation keys: %w", err)
	}
	checkout.ReservationKeys = resKeysJSON
	s.checkoutRepo.Update(ctx, checkout)

	s.logStep(ctx, checkout.ID, domain.StepReserve, "success", fmt.Sprintf("reserved %d items", len(reservationKeys)), time.Since(start).Milliseconds())
	metrics.CheckoutLatency.WithLabelValues(domain.StepReserve).Observe(time.Since(start).Seconds())
	return reservationKeys, nil
}

func (s *CheckoutService) stepProcess(ctx context.Context, checkout *domain.Checkout, req InitiateRequest) error {
	start := time.Now()
	defer func() { metrics.CheckoutLatency.WithLabelValues(domain.StepProcess).Observe(time.Since(start).Seconds()) }()
	checkout.AdvanceStep(domain.StepProcess)
	checkout.Status = domain.CheckoutStatusProcessing
	s.checkoutRepo.Update(ctx, checkout)

	// In production: call Order Service to create order, then Payment Service
	orderID := fmt.Sprintf("ORD-%s", checkout.ID[:8])

	s.logStep(ctx, checkout.ID, domain.StepProcess, "success", "order_created:"+orderID, time.Since(start).Milliseconds())
	metrics.CheckoutLatency.WithLabelValues(domain.StepProcess).Observe(time.Since(start).Seconds())
	return nil
}

func (s *CheckoutService) stepComplete(ctx context.Context, checkout *domain.Checkout, reservationKeys []string) {
	start := time.Now()
	defer func() { metrics.CheckoutLatency.WithLabelValues(domain.StepComplete).Observe(time.Since(start).Seconds()) }()
	checkout.AdvanceStep(domain.StepComplete)
	checkout.MarkCompleted(fmt.Sprintf("ORD-%s", checkout.ID[:8]))
	s.checkoutRepo.Update(ctx, checkout)

	// Confirm all reservations asynchronously
	for i, key := range reservationKeys {
		job := &domain.ReconciliationJob{
			ID:          fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), i),
			CheckoutID:  checkout.ID,
			JobType:     domain.JobTypeConfirmReservation,
			Status:      domain.JobStatusPending,
			MaxAttempts: 3,
			NextRetryAt: time.Now(),
			Metadata:    fmt.Sprintf(`{"reservation_key":"%s"}`, key),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.reconcileRepo.Create(ctx, job); err != nil {
			observability.LogWithTrace(ctx).Error("failed to create reconciliation job",
				zap.String("checkout_id", checkout.ID), zap.Error(err))
		}
	}

	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, "checkout.completed", map[string]interface{}{
			"checkout_id": checkout.ID, "order_id": checkout.OrderID,
			"grand_total": checkout.GrandTotal, "currency": checkout.Currency,
		}); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish checkout.completed event",
				zap.String("checkout_id", checkout.ID), zap.Error(err))
		}
	}
}

func (s *CheckoutService) handleFailure(ctx context.Context, checkout *domain.Checkout, step string, err error) {
	logger := observability.LogWithTrace(ctx)
	logger.Error("checkout failed",
		zap.String("checkout_id", checkout.ID),
		zap.String("step", step),
		zap.Error(err),
	)

	checkout.MarkFailed(fmt.Sprintf("[%s] %s", step, err.Error()))
	s.checkoutRepo.Update(ctx, checkout)

	// Create reconciliation job for cleanup
	job := &domain.ReconciliationJob{
		ID:          fmt.Sprintf("job_%d", time.Now().UnixNano()),
		CheckoutID:  checkout.ID,
		JobType:     domain.JobTypeReleaseReservation,
		Status:      domain.JobStatusPending,
		MaxAttempts: 3,
		NextRetryAt: time.Now().Add(100 * time.Millisecond),
		Metadata:    fmt.Sprintf(`{"failure_step":"%s","error":"%s"}`, step, err.Error()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if createErr := s.reconcileRepo.Create(ctx, job); createErr != nil {
		logger.Error("failed to create reconciliation job", zap.Error(createErr))
	}

	metrics.CheckoutsFailed.Inc()

	if s.publisher != nil {
		if pubErr := s.publisher.Publish(ctx, "checkout.failed", map[string]interface{}{
			"checkout_id": checkout.ID, "step": step, "error": err.Error(),
		}); pubErr != nil {
			logger.Error("failed to publish checkout.failed event", zap.Error(pubErr))
		}
	}
}

func (s *CheckoutService) rollbackReservations(ctx context.Context, checkout *domain.Checkout, keys []string) {
	checkout.MarkRollingBack()
	s.checkoutRepo.Update(ctx, checkout)

	for _, key := range keys {
		if err := s.reservationRepo.UpdateStatus(ctx, key, domain.ReservationStatusReleased); err != nil {
			observability.LogWithTrace(ctx).Error("failed to rollback reservation",
				zap.String("checkout_id", checkout.ID),
				zap.String("reservation_key", key),
				zap.Error(err))
		}
	}

	checkout.MarkRolledBack()
	s.checkoutRepo.Update(ctx, checkout)
}

func (s *CheckoutService) logStep(ctx context.Context, checkoutID, step, status, errMsg string, durationMs int64) {
	log := domain.NewCheckoutStepLog(checkoutID, step, status, durationMs, errMsg, "")
	if err := s.stepLogRepo.Create(ctx, log); err != nil {
		observability.LogWithTrace(ctx).Error("failed to log checkout step",
			zap.String("checkout_id", checkoutID),
			zap.String("step", step),
			zap.Error(err))
	}
}

// GetCheckoutStatus returns the current status of a checkout
func (s *CheckoutService) GetCheckoutStatus(ctx context.Context, checkoutID string) (*domain.Checkout, error) {
	checkout, err := s.checkoutRepo.FindByID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}
	if checkout == nil {
		return nil, domain.ErrCheckoutNotFound
	}
	return checkout, nil
}

// RetryCheckout retries a failed checkout
func (s *CheckoutService) RetryCheckout(ctx context.Context, checkoutID, requestingUserID string) error {
	checkout, err := s.checkoutRepo.FindByID(ctx, checkoutID)
	if err != nil {
		return err
	}
	if checkout == nil {
		return domain.ErrCheckoutNotFound
	}
	if checkout.UserID != requestingUserID {
		return domain.ErrUnauthorized
	}
	if !checkout.CanRetry() {
		return fmt.Errorf("%w: status=%s attempts=%d", domain.ErrMaxRetriesExceeded, checkout.Status, checkout.AttemptCount)
	}

	checkout.IncrementAttempt()
	checkout.Status = domain.CheckoutStatusPending
	checkout.FailureReason = ""
	s.checkoutRepo.Update(ctx, checkout)

	// Re-trigger saga (in production, reload original request from snapshot)
	metrics.CheckoutRetries.Inc()
	return nil
}

// Request types

type InitiateRequest struct {
	UserID         string               `json:"user_id"`
	CartId         string               `json:"cart_id"`
	IdempotencyKey string               `json:"idempotency_key"`
	Currency       string               `json:"currency"`
	Items          []ItemRequest        `json:"items"`
	SellerGroups   []SellerGroupRequest `json:"seller_groups"`
	Subtotal       int64                `json:"subtotal"`
	DiscountTotal  int64                `json:"discount_total"`
	ShippingTotal  int64                `json:"shipping_total"`
	GrandTotal     int64                `json:"grand_total"`
	Promotions     interface{}          `json:"promotions"`
}

type ItemRequest struct {
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
	ShopID      string `json:"shop_id"`
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
}

type SellerGroupRequest struct {
	ShopID      string        `json:"shop_id"`
	ShopName    string        `json:"shop_name"`
	Items       []ItemRequest `json:"items"`
	Subtotal    int64         `json:"subtotal"`
	ShippingFee int64         `json:"shipping_fee"`
}

func mustJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(b), nil
}