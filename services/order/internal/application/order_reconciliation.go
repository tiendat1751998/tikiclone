package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func (s *OrderService) TriggerReconciliation(ctx context.Context, orderID string, rtype domain.ReconciliationType) (*domain.OrderReconciliation, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.TriggerReconciliation")
	defer span.End()

	rec := domain.NewOrderReconciliation(orderID, rtype)
	if err := s.orderRepo.SaveReconciliation(ctx, rec); err != nil {
		return nil, err
	}

	// Publish reconciliation event
	event := &domain.OrderEvent{
		OrderID:   orderID,
		EventType: domain.EventOrderReconciliationTriggered,
		Timestamp: time.Now().UTC(),
	}
	if s.kafkaProducer != nil {
		s.kafkaProducer.PublishEvent(ctx, event)
	}

	return rec, nil
}

func (s *OrderService) RunReconciliation(ctx context.Context) error {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.RunReconciliation")
	defer span.End()

	recs, err := s.orderRepo.GetPendingReconciliations(ctx, 100)
	if err != nil {
		return err
	}

	// Batch-load all referenced orders to avoid N+1
	orderIDs := make([]string, 0, len(recs))
	orderMap := make(map[string]*domain.Order, len(recs))
	for _, rec := range recs {
		orderIDs = append(orderIDs, rec.OrderID)
	}
	if len(orderIDs) > 0 {
		loaded, err := s.orderRepo.FindByIDs(ctx, orderIDs)
		if err != nil {
			return fmt.Errorf("batch load orders for reconciliation: %w", err)
		}
		orderMap = loaded
	}

	for _, rec := range recs {
		start := time.Now()
		s.reconcileOrder(ctx, rec, orderMap[rec.OrderID])
		metrics.ReconciliationLatency.WithLabelValues(string(rec.ReconciliationType)).Observe(time.Since(start).Seconds())
	}

	return nil
}

func (s *OrderService) reconcileOrder(ctx context.Context, rec *domain.OrderReconciliation, order *domain.Order) {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("panic in reconciliation", zap.Any("recover", r))
		}
	}()

	rec.IncrementRetry()
	rec.MarkChecked()

	if order == nil {
		zap.L().Warn("order not found for reconciliation", zap.String("order_id", rec.OrderID))
		s.orderRepo.UpdateReconciliationStatus(ctx, rec.ID, domain.ReconciliationStatusFailed, rec.RetryCount)
		metrics.ReconciliationFailures.WithLabelValues(string(rec.ReconciliationType)).Inc()
		return
	}

	var status domain.ReconciliationStatus

	switch rec.ReconciliationType {
	case domain.ReconciliationTypePayment:
		status = s.reconcilePayment(ctx, order)
	case domain.ReconciliationTypeInventory:
		status = s.reconcileInventory(ctx, order)
	case domain.ReconciliationTypeShipment:
		status = s.reconcileShipment(ctx, order)
	default:
		status = domain.ReconciliationStatusFailed
	}

	s.orderRepo.UpdateReconciliationStatus(ctx, rec.ID, status, rec.RetryCount)

	if status == domain.ReconciliationStatusFailed && rec.CanRetry() {
		metrics.ReconciliationFailures.WithLabelValues(string(rec.ReconciliationType)).Inc()
	}
}

func (s *OrderService) reconcilePayment(ctx context.Context, order *domain.Order) domain.ReconciliationStatus {
	// In production: call payment service to verify payment status
	// For now, mark as matched if order is paid or beyond
	if order.Status == domain.OrderStatusPaid ||
		order.Status == domain.OrderStatusProcessing ||
		order.Status == domain.OrderStatusPacked ||
		order.Status == domain.OrderStatusShipped ||
		order.Status == domain.OrderStatusDelivered ||
		order.Status == domain.OrderStatusCompleted {
		return domain.ReconciliationStatusMatched
	}
	if order.Status == domain.OrderStatusCancelled || order.Status == domain.OrderStatusRefunded {
		return domain.ReconciliationStatusMatched
	}
	return domain.ReconciliationStatusPending
}

func (s *OrderService) reconcileInventory(ctx context.Context, order *domain.Order) domain.ReconciliationStatus {
	// In production: call inventory service to verify stock deductions
	// Check if order has reached paid status or beyond (using valid state progression)
	if order.Status == domain.OrderStatusPaid ||
		order.Status == domain.OrderStatusProcessing ||
		order.Status == domain.OrderStatusPacked ||
		order.Status == domain.OrderStatusShipped ||
		order.Status == domain.OrderStatusDelivered ||
		order.Status == domain.OrderStatusCompleted {
		return domain.ReconciliationStatusMatched
	}
	return domain.ReconciliationStatusPending
}

func (s *OrderService) reconcileShipment(ctx context.Context, order *domain.Order) domain.ReconciliationStatus {
	// In production: call shipment service to verify shipment status
	if order.Status == domain.OrderStatusShipped ||
		order.Status == domain.OrderStatusDelivered ||
		order.Status == domain.OrderStatusCompleted {
		return domain.ReconciliationStatusMatched
	}
	if order.Status == domain.OrderStatusCancelled || order.Status == domain.OrderStatusRefunded {
		return domain.ReconciliationStatusMatched
	}
	return domain.ReconciliationStatusPending
}
