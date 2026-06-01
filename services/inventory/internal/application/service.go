package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/services/inventory/internal/config"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
	"github.com/tikiclone/tiki/services/inventory/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/inventory/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/inventory/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/inventory/internal/metrics"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type InventoryService struct {
	cfg           *config.Config
	db            *sql.DB
	invRepo       *mysql.InventoryRepository
	redisStore    *redisinfra.Store
	kafkaProducer *kafka.Producer
}

func NewInventoryService(cfg *config.Config, db *sql.DB, repo *mysql.InventoryRepository, store *redisinfra.Store, producer *kafka.Producer) *InventoryService {
	return &InventoryService{cfg: cfg, db: db, invRepo: repo, redisStore: store, kafkaProducer: producer}
}

type ReserveStockRequest struct {
	OrderID        string `json:"order_id" validate:"required"`
	UserID         string `json:"user_id" validate:"required"`
	ProductID      string `json:"product_id" validate:"required"`
	SkuID          string `json:"sku_id" validate:"required"`
	WarehouseID    string `json:"warehouse_id"`
	Quantity       int    `json:"quantity" validate:"required,min=1"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

// ReserveStock reserves inventory for an order with full protection against:
// - Oversell (DB transaction with SERIALIZABLE isolation + SELECT FOR UPDATE)
// - Double-spend (idempotency keys)
// - Race conditions (distributed lock + DB lock)
// - Data inconsistency (atomic stock update + reservation creation in same tx)
func (s *InventoryService) ReserveStock(ctx context.Context, req *ReserveStockRequest) (*domain.Reservation, error) {
	ctx, span := otel.Tracer("tiki-inventory").Start(ctx, "InventoryService.ReserveStock")
	defer span.End()

	start := time.Now()
	defer func() { metrics.ReservationLatency.Observe(time.Since(start).Seconds()) }()

	// Validate input
	if err := validateReserveRequest(req); err != nil {
		metrics.ReservationFailures.WithLabelValues("validation_failed").Inc()
		return nil, err
	}

	// Idempotency check (fast path - no lock needed)
	if req.IdempotencyKey != "" {
		existing, err := s.invRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			metrics.IdempotentRequests.Inc()
			return existing, nil
		}
	}

	// Acquire distributed lock for this SKU (prevents concurrent reservations across instances)
	lockToken, locked, err := s.redisStore.AcquireStockLock(ctx, req.SkuID, 10*time.Second)
	if err != nil || !locked {
		metrics.ReservationFailures.WithLabelValues("lock_failed").Inc()
		return nil, fmt.Errorf("failed to acquire stock lock for SKU %s", req.SkuID)
	}
	defer s.redisStore.ReleaseStockLock(ctx, req.SkuID, lockToken)

	// Execute stock update + reservation creation in a single DB transaction
	reservation, err := s.executeReservationInTx(ctx, req)
	if err != nil {
		metrics.ReservationFailures.WithLabelValues("transaction_failed").Inc()
		return nil, err
	}

	// Store idempotency key only AFTER successful transaction
	if req.IdempotencyKey != "" {
		rec := domain.NewIdempotencyRecord(reservation.ID, s.cfg.Inventory.IdempotencyTTL)
		rec.Key = req.IdempotencyKey
		if err := s.invRepo.SaveIdempotencyKey(ctx, rec); err != nil {
			observability.LogWithTrace(ctx).Error("failed to save idempotency key", zap.Error(err))
		}
		if err := s.redisStore.StoreIdempotencyKey(ctx, req.IdempotencyKey, reservation.ID, s.cfg.Inventory.IdempotencyTTL); err != nil {
			observability.LogWithTrace(ctx).Error("failed to store idempotency key in Redis", zap.Error(err))
		}
	}

	// Invalidate cache (don't update - prevents stale data on cache write failure)
	s.redisStore.InvalidateStockCache(ctx, req.SkuID)

	// Publish event via outbox pattern
	event := &domain.InventoryEvent{
		ProductID: req.ProductID, SkuID: req.SkuID, WarehouseID: req.WarehouseID,
		Quantity: req.Quantity, EventType: domain.EventStockReserved, Timestamp: time.Now().UTC(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		observability.LogWithTrace(ctx).Error("failed to marshal reserve event", zap.Error(err))
	} else {
		s.invRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("inventory", reservation.ID, string(domain.EventStockReserved), payload))
	}
	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.PublishEvent(ctx, event); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish reserve event", zap.Error(err))
		}
	}

	span.SetAttributes(
		attribute.String("reservation_id", reservation.ID),
		attribute.String("sku_id", req.SkuID),
		attribute.Int("quantity", req.Quantity),
	)
	zap.L().Info("stock reserved",
		zap.String("reservation_id", reservation.ID),
		zap.String("sku_id", req.SkuID),
		zap.Int("qty", req.Quantity),
	)
	return reservation, nil
}

// executeReservationInTx performs the actual stock reservation within a DB transaction.
// This is the critical section that prevents oversell.
func (s *InventoryService) executeReservationInTx(ctx context.Context, req *ReserveStockRequest) (*domain.Reservation, error) {
	var reservation *domain.Reservation

	err := s.invRepo.ExecInTx(ctx, func(tx *sql.Tx) error {
		// Get stock with row-level lock (SELECT ... FOR UPDATE)
		stock, err := s.invRepo.GetStockForUpdate(ctx, tx, req.SkuID, req.WarehouseID)
		if err != nil {
			return err
		}

		// Check available quantity (oversell prevention)
		if stock.AvailableQty < req.Quantity {
			metrics.OversellPreventionCount.Inc()
			return domain.ErrInsufficientStock
		}

		// Reserve stock in memory
		if err := stock.Reserve(req.Quantity); err != nil {
			return err
		}
		stock.Version++

		// Persist stock update within transaction
		if err := s.invRepo.UpdateStockInTx(ctx, tx, stock); err != nil {
			return fmt.Errorf("update stock: %w", err)
		}

		// Create reservation record within SAME transaction
		reservation = domain.NewReservation(req.OrderID, req.UserID, req.ProductID, req.SkuID, req.WarehouseID, req.Quantity, s.cfg.Inventory.ReservationTTL, req.IdempotencyKey)
		if err := s.invRepo.SaveReservationInTx(ctx, tx, reservation); err != nil {
			return fmt.Errorf("save reservation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return reservation, nil
}

// ReleaseStock releases a reservation and returns stock to available pool.
// Uses DB transaction to ensure atomicity.
func (s *InventoryService) ReleaseStock(ctx context.Context, reservationID string) error {
	ctx, span := otel.Tracer("tiki-inventory").Start(ctx, "InventoryService.ReleaseStock")
	defer span.End()

	// Get reservation first to have SkuID for cache invalidation after tx
	reservation, err := s.invRepo.GetReservation(ctx, reservationID)
	if err != nil {
		return err
	}

	err = s.invRepo.ExecInTx(ctx, func(tx *sql.Tx) error {
		// Get reservation with lock
		res, err := s.invRepo.GetReservationForUpdate(ctx, tx, reservationID)
		if err != nil {
			return err
		}

		// Validate state transition
		if err := res.Release(); err != nil {
			return err
		}

		// Update reservation status
		if err := s.invRepo.UpdateReservationStatusInTx(ctx, tx, reservationID, domain.ReservationStatusReleased); err != nil {
			return err
		}

		// Get stock with lock and release
		stock, err := s.invRepo.GetStockForUpdate(ctx, tx, res.SkuID, res.WarehouseID)
		if err != nil {
			return err
		}
		stock.Release(res.Quantity)
		stock.Version++
		if err := s.invRepo.UpdateStockInTx(ctx, tx, stock); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Invalidate cache after successful commit
	s.redisStore.InvalidateStockCache(ctx, reservation.SkuID)

	// Publish event
	event := &domain.InventoryEvent{
		SkuID: reservation.SkuID, EventType: domain.EventStockReleased, Timestamp: time.Now().UTC(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		observability.LogWithTrace(ctx).Error("failed to marshal release event", zap.Error(err))
	} else {
		s.invRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("inventory", reservationID, string(domain.EventStockReleased), payload))
	}
	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.PublishEvent(ctx, event); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish release event", zap.Error(err))
		}
	}

	return nil
}

func (s *InventoryService) GetStock(ctx context.Context, skuID, warehouseID string) (*domain.Stock, error) {
	return s.invRepo.GetStock(ctx, skuID, warehouseID)
}

// ExpireReservations processes expired reservations with per-reservation timeout.
// Uses errgroup for bounded concurrency and proper goroutine lifecycle management.
func (s *InventoryService) ExpireReservations(ctx context.Context) error {
	reservations, err := s.invRepo.GetExpiredReservations(ctx, 100)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrent goroutines to prevent resource exhaustion

	for _, res := range reservations {
		res := res // capture range variable
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					zap.L().Error("panic in ExpireReservations", zap.Any("recover", r), zap.String("id", res.ID))
				}
			}()
			resCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := s.ReleaseStock(resCtx, res.ID); err != nil {
				zap.L().Warn("failed to expire reservation",
					zap.String("id", res.ID), zap.Error(err))
				return nil // Don't fail entire batch for single reservation failure
			}
			return nil
		})
	}
	return g.Wait()
}

// ProcessOutboxEvents publishes pending outbox events to Kafka with idempotency.
func (s *InventoryService) ProcessOutboxEvents(ctx context.Context) error {
	events, err := s.invRepo.GetUnprocessedOutboxEvents(ctx, 100)
	if err != nil {
		return err
	}

	var processedIDs []string
	for _, event := range events {
		// Mark as processing first (prevents duplicate processing)
		if err := s.invRepo.MarkOutboxEventProcessing(ctx, event.ID); err != nil {
			observability.LogWithTrace(ctx).Error("failed to mark outbox event as processing",
				zap.String("event_id", event.ID), zap.Error(err))
			continue
		}

		var invEvent domain.InventoryEvent
		if err := json.Unmarshal(event.Payload, &invEvent); err != nil {
			s.invRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			continue
		}

		if err := s.kafkaProducer.PublishEvent(ctx, &invEvent); err != nil {
			s.invRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			continue
		}

		processedIDs = append(processedIDs, event.ID)
	}
	// Batch-mark all processed events in a single query
	if len(processedIDs) > 0 {
		if err := s.invRepo.MarkOutboxEventsProcessed(ctx, processedIDs); err != nil {
			zap.L().Error("failed to batch mark outbox events processed", zap.Error(err))
		}
	}
	return nil
}

func validateReserveRequest(req *ReserveStockRequest) error {
	if req.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.SkuID == "" {
		return fmt.Errorf("sku_id is required")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive, got %d", req.Quantity)
	}
	if req.IdempotencyKey == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	return nil
}
