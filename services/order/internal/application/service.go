package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/order/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/order/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/order/internal/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type OrderService struct {
	cfg           *config.Config
	orderRepo     *mysql.OrderRepository
	outboxRepo    *mysql.OutboxRepository
	redisStore    *redisinfra.Store
	kafkaProducer *kafka.Producer
	numberGen     *domain.OrderNumberGenerator
	stateMachine  *domain.StateMachine
}

func NewOrderService(
	cfg *config.Config,
	orderRepo *mysql.OrderRepository,
	outboxRepo *mysql.OutboxRepository,
	redisStore *redisinfra.Store,
	kafkaProducer *kafka.Producer,
) *OrderService {
	return &OrderService{
		cfg:           cfg,
		orderRepo:     orderRepo,
		outboxRepo:    outboxRepo,
		redisStore:    redisStore,
		kafkaProducer: kafkaProducer,
		numberGen:     domain.NewOrderNumberGenerator(),
		stateMachine:  domain.NewStateMachine(),
	}
}

type CreateOrderRequest struct {
	UserID          string                `json:"user_id" validate:"required"`
	SellerID        string                `json:"seller_id" validate:"required"`
	Currency        string                `json:"currency"`
	IdempotencyKey  string                `json:"idempotency_key" validate:"required"`
	ShippingAddress domain.Address        `json:"shipping_address" validate:"required"`
	BillingAddress  domain.Address        `json:"billing_address" validate:"required"`
	Items           []domain.SnapshotItem `json:"items" validate:"required,min=1"`
	Metadata        json.RawMessage       `json:"metadata,omitempty"`
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*domain.Order, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", req.UserID),
		attribute.String("seller_id", req.SellerID),
	)

	start := time.Now()
	defer func() {
		metrics.OrderCreationLatency.Observe(time.Since(start).Seconds())
	}()

	// Check idempotency
	if req.IdempotencyKey != "" {
		existingID, err := s.redisStore.CheckIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			zap.L().Warn("idempotency check failed", zap.Error(err))
		} else if existingID != "" {
			metrics.IdempotencyHits.Inc()
			existingOrder, err := s.orderRepo.FindByID(ctx, existingID)
			if err == nil && existingOrder != nil {
				span.SetAttributes(attribute.Bool("idempotent", true))
				return existingOrder, nil
			}
		}

		// Also check DB for idempotency
		existingOrder, err := s.orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			zap.L().Warn("db idempotency check failed", zap.Error(err))
		} else if existingOrder != nil {
			metrics.IdempotencyHits.Inc()
			return existingOrder, nil
		}
	}

	currency := req.Currency
	if currency == "" {
		currency = s.cfg.Order.DefaultCurrency
	}

	// Build order items
	orderItems := make([]domain.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		snapshot, _ := json.Marshal(item)
		orderItems = append(orderItems, *domain.NewOrderItem(
			"", item.ProductID, item.SkuID, item.ShopID,
			item.Quantity, item.UnitPrice, snapshot,
		))
	}

	order := domain.NewOrder(
		req.UserID, req.SellerID, currency,
		req.IdempotencyKey, req.ShippingAddress, req.BillingAddress,
		orderItems,
	)
	order.OrderNumber = s.numberGen.Generate()
	if len(req.Metadata) > 0 {
		order.Metadata = &req.Metadata
	}

	// Create snapshot
	cartSnapshot := &domain.CartSnapshot{
		Items:       req.Items,
		TotalAmount: order.TotalAmount,
		Currency:    currency,
	}
	snapshot, err := domain.NewOrderSnapshot(order.ID, cartSnapshot)
	if err != nil {
		span.SetStatus(codes.Error, "snapshot creation failed")
		return nil, fmt.Errorf("failed to create order snapshot: %w", err)
	}
	order.SnapshotID = snapshot.ID

	// Publish event via outbox (saved atomically with order creation)
	orderEvent := domain.NewOrderEvent(order, domain.EventOrderCreated, req.Metadata)
	eventPayload, _ := json.Marshal(orderEvent)
	outboxEvent := domain.NewOutboxEvent("order", order.ID, string(domain.EventOrderCreated), eventPayload)

	// Persist order + outbox in single transaction
	if err := s.orderRepo.Create(ctx, order, outboxEvent); err != nil {
		span.SetStatus(codes.Error, "order creation failed")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Save snapshot (best-effort after successful transaction)
	if err := s.orderRepo.SaveSnapshot(ctx, snapshot); err != nil {
		zap.L().Warn("failed to save snapshot", zap.Error(err))
	}

	// Save idempotency key (best-effort after successful transaction)
	if req.IdempotencyKey != "" {
		rec := domain.NewIdempotencyRecord(order.ID, s.cfg.Order.IdempotencyKeyTTL)
		rec.Key = req.IdempotencyKey
		if err := s.orderRepo.SaveIdempotencyKey(ctx, rec); err != nil {
			zap.L().Warn("failed to save idempotency key", zap.Error(err))
		}
		s.redisStore.StoreIdempotencyKey(ctx, req.IdempotencyKey, order.ID, s.cfg.Order.IdempotencyKeyTTL)
	}

	// Save lifecycle event (best-effort after successful transaction)
	lifeEvent := domain.NewLifecycleEvent(order.ID, "", domain.OrderStatusPending, "order_created", req.UserID, "user")
	if err := s.orderRepo.SaveLifecycleEvent(ctx, lifeEvent); err != nil {
		zap.L().Warn("failed to save lifecycle event", zap.Error(err))
	}

	// Publish to Kafka (best-effort after successful transaction)
	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.PublishEvent(ctx, orderEvent); err != nil {
			zap.L().Warn("failed to publish kafka event", zap.Error(err))
		}
	}

	// Populate computed fields from snapshot
	domain.EnrichOrder(order)

	// Cache the order
	s.redisStore.CacheOrder(ctx, order, 5*time.Minute)

	// Update metrics
	metrics.OrdersCreatedTotal.WithLabelValues(currency).Inc()
	metrics.ActiveOrdersByStatus.WithLabelValues(string(domain.OrderStatusPending)).Inc()

	span.SetAttributes(
		attribute.String("order_id", order.ID),
		attribute.String("order_number", order.OrderNumber),
	)

	zap.L().Info("order created",
		zap.String("order_id", order.ID),
		zap.String("order_number", order.OrderNumber),
		zap.String("user_id", req.UserID),
		zap.Int64("total_amount", order.TotalAmount),
	)

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.GetOrder")
	defer span.End()

	// Try cache first
	cached, err := s.redisStore.GetCachedOrder(ctx, orderID)
	if err == nil && cached != nil {
		metrics.CacheHits.WithLabelValues("get_order").Inc()
		return cached, nil
	}
	metrics.CacheMisses.WithLabelValues("get_order").Inc()

	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		span.SetStatus(codes.Error, "order not found")
		return nil, err
	}

	domain.EnrichOrder(order)

	// Cache for next time
	s.redisStore.CacheOrder(ctx, order, 5*time.Minute)

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*domain.Order, int, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.ListOrders")
	defer span.End()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	orders, err := s.orderRepo.FindByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	for _, o := range orders {
		domain.EnrichOrder(o)
	}

	total, err := s.orderRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (s *OrderService) TransitionStatus(ctx context.Context, orderID string, target domain.OrderStatus, actorID, actorType, reason string) (*domain.Order, error) {
	ctx, span := otel.Tracer("tiki-order").Start(ctx, "OrderService.TransitionStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", orderID),
		attribute.String("target_status", string(target)),
	)

	start := time.Now()
	defer func() {
		metrics.OrderTransitionLatency.WithLabelValues("", string(target)).Observe(time.Since(start).Seconds())
	}()

	// Acquire distributed lock
	locked, err := s.redisStore.AcquireTransitionLock(ctx, orderID, 10*time.Second)
	if err != nil || !locked {
		return nil, fmt.Errorf("failed to acquire transition lock: %w", err)
	}
	defer s.redisStore.ReleaseTransitionLock(ctx, orderID)

	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Validate ownership — user owns the order, or actor is admin/seller/system
	if order.UserID != actorID && actorType != "admin" && actorType != "system" && actorType != "seller" {
		return nil, domain.ErrUnauthorized
	}

	// Perform transition
	lifecycleEvent, err := order.TransitionTo(target, actorID, actorType, reason)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Persist status change with optimistic locking
	// TransitionTo already incremented Version, so we use the current Version as expected
	if err := s.orderRepo.UpdateStatus(ctx, orderID, target, order.Version); err != nil {
		return nil, err
	}

	// Save lifecycle event
	if err := s.orderRepo.SaveLifecycleEvent(ctx, lifecycleEvent); err != nil {
		zap.L().Warn("failed to save lifecycle event", zap.Error(err))
	}

	// Invalidate cache
	s.redisStore.InvalidateOrderCache(ctx, orderID)

	// Publish event via outbox (best-effort - outbox processor will retry)
	transitionEvent := domain.NewOrderEvent(order, domain.EventType("order."+string(target)), nil)
	eventPayload, _ := json.Marshal(transitionEvent)
	outboxEvent := domain.NewOutboxEvent("order", order.ID, string(domain.EventType("order."+string(target))), eventPayload)
	if err := s.outboxRepo.SaveOutboxEvent(ctx, outboxEvent); err != nil {
		zap.L().Warn("failed to save outbox event", zap.Error(err))
	}

	// Best-effort Kafka publish (outbox pattern is the reliable path)
	if s.kafkaProducer != nil {
		if err := s.kafkaProducer.PublishEvent(ctx, transitionEvent); err != nil {
			zap.L().Warn("failed to publish kafka event", zap.Error(err))
		}
	}

	// Update metrics
	metrics.ActiveOrdersByStatus.WithLabelValues(string(lifecycleEvent.FromStatus)).Dec()
	metrics.ActiveOrdersByStatus.WithLabelValues(string(target)).Inc()

	zap.L().Info("order status transitioned",
		zap.String("order_id", orderID),
		zap.String("from", string(lifecycleEvent.FromStatus)),
		zap.String("to", string(target)),
	)

	return order, nil
}

func (s *OrderService) GetOrderHistory(ctx context.Context, orderID string) ([]*domain.LifecycleEvent, error) {
	return s.orderRepo.GetLifecycleHistory(ctx, orderID)
}

func (s *OrderService) GetReconciliationStatus(ctx context.Context, orderID string) ([]*domain.OrderReconciliation, error) {
	// This would query the reconciliation table
	// For now return empty
	return nil, nil
}

// ProcessOutboxEvents polls the outbox table and publishes to Kafka with the three-state pattern:
// pending → processing → processed/failed. This prevents duplicate messages if the publish
// succeeds but the DB update fails (event will retry from 'processing' state).
func (s *OrderService) ProcessOutboxEvents(ctx context.Context) error {
	if s.kafkaProducer == nil {
		return nil
	}

	events, err := s.outboxRepo.GetUnprocessedOutboxEvents(ctx, 100)
	if err != nil {
		return err
	}

	var processedIDs []string
	for _, event := range events {
		if err := s.outboxRepo.MarkOutboxEventProcessing(ctx, event.ID); err != nil {
			zap.L().Warn("failed to mark outbox event as processing", zap.Error(err), zap.String("event_id", event.ID))
			continue
		}

		var orderEvent domain.OrderEvent
		if err := json.Unmarshal(event.Payload, &orderEvent); err != nil {
			s.outboxRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			zap.L().Warn("failed to unmarshal outbox event", zap.Error(err), zap.String("event_id", event.ID))
			continue
		}

		if err := s.kafkaProducer.PublishEvent(ctx, &orderEvent); err != nil {
			s.outboxRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			zap.L().Warn("failed to publish outbox event", zap.Error(err), zap.String("event_id", event.ID))
			continue
		}

		processedIDs = append(processedIDs, event.ID)
	}

	if len(processedIDs) > 0 {
		if err := s.outboxRepo.MarkOutboxEventsProcessed(ctx, processedIDs); err != nil {
			zap.L().Error("failed to batch mark outbox events processed", zap.Error(err))
		}
	}

	return nil
}
