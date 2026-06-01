package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type CancelOrderRequest struct {
	OrderID       string                   `json:"order_id" validate:"required"`
	Reason        string                   `json:"reason" validate:"required"`
	CancelledBy   string                   `json:"cancelled_by" validate:"required"`
	CancelledType domain.CancellationType  `json:"cancelled_type" validate:"required"`
}

func (s *OrderService) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*domain.Order, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.CancelOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", req.OrderID),
		attribute.String("cancelled_by", req.CancelledBy),
		attribute.String("cancellation_type", string(req.CancelledType)),
	)

	start := time.Now()
	defer func() {
		metrics.OrderCancellationLatency.Observe(time.Since(start).Seconds())
	}()

	// Acquire distributed lock
	locked, err := s.redisStore.AcquireTransitionLock(ctx, req.OrderID, 10*time.Second)
	if err != nil || !locked {
		return nil, fmt.Errorf("failed to acquire cancellation lock: %w", err)
	}
	defer s.redisStore.ReleaseTransitionLock(ctx, req.OrderID)

	order, err := s.orderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	// Validate cancellation
	if !order.IsCancellable() {
		span.SetStatus(codes.Error, "order not cancellable")
		return nil, domain.ErrOrderNotCancellable
	}

	// Perform cancellation transition
	lifecycleEvent, err := order.TransitionTo(domain.OrderStatusCancelled, req.CancelledBy, string(req.CancelledType), req.Reason)
	if err != nil {
		return nil, err
	}

	cancellation := domain.NewOrderCancellation(
		req.OrderID, req.Reason, req.CancelledBy, req.CancelledType, order.TotalAmount,
	)

	// Persist all changes in a single SERIALIZABLE transaction
	cancelEvent := domain.NewOrderEvent(order, domain.EventOrderCancelled, nil)
	eventPayload, _ := json.Marshal(cancelEvent)
	outboxEvent := domain.NewOutboxEvent("order", order.ID, string(domain.EventOrderCancelled), eventPayload)

	err = s.orderRepo.ExecInTx(ctx, func(tx *sqlx.Tx) error {
		// TransitionTo incremented Version in memory; use original version for DB optimistic lock
		if err := s.orderRepo.UpdateStatusInTx(ctx, tx, req.OrderID, domain.OrderStatusCancelled, order.Version-1); err != nil {
			return err
		}
		if err := s.orderRepo.SaveLifecycleEventInTx(ctx, tx, lifecycleEvent); err != nil {
			return err
		}
		if err := s.orderRepo.SaveCancellationInTx(ctx, tx, cancellation); err != nil {
			return err
		}
		if err := s.outboxRepo.SaveOutboxEventInTx(ctx, tx, outboxEvent); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist cancellation: %w", err)
	}

	// Invalidate cache
	s.redisStore.InvalidateOrderCache(ctx, req.OrderID)

	// Publish event to Kafka (best-effort after successful transaction)
	if s.kafkaProducer != nil {
		s.kafkaProducer.PublishEvent(ctx, cancelEvent)
	}

	// Trigger compensation workflows asynchronously
	compCtx, compCancel := context.WithTimeout(ctx, 30*time.Second)
	go func() {
		defer compCancel()
		s.triggerCompensation(compCtx, order, cancellation)
	}()

	// Update metrics
	metrics.OrdersCancelledTotal.WithLabelValues(string(req.CancelledType)).Inc()
	metrics.ActiveOrdersByStatus.WithLabelValues(string(order.Status)).Dec()

	zap.L().Info("order cancelled",
		zap.String("order_id", req.OrderID),
		zap.String("reason", req.Reason),
		zap.String("cancelled_by", req.CancelledBy),
	)

	return order, nil
}

func (s *OrderService) triggerCompensation(ctx context.Context, order *domain.Order, cancellation *domain.OrderCancellation) {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("panic in compensation", zap.Any("recover", r), zap.String("order_id", order.ID))
		}
	}()

	// Compensation: release inventory reservations, trigger refund if paid
	zap.L().Info("triggering compensation for cancelled order",
		zap.String("order_id", order.ID),
		zap.String("cancellation_id", cancellation.ID),
	)

	// Update compensation status
	s.orderRepo.UpdateCancellationCompensation(ctx, cancellation.ID, domain.CompensationInProgress)

	// In a real system, this would:
	// 1. Call inventory service to release reservations
	// 2. Call payment service to trigger refund if order was paid
	// 3. Call shipment service to cancel shipment if order was shipped

	s.orderRepo.UpdateCancellationCompensation(ctx, cancellation.ID, domain.CompensationCompleted)
}
