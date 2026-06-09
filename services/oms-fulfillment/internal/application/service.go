package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/domain"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/infrastructure/kafka"
	redisinfra "github.com/tikiclone/tiki/services/oms-fulfillment/internal/infrastructure/redis"
	"go.uber.org/zap"
)

type FulfillmentRepository interface {
	Create(ctx context.Context, f *domain.FulfillmentOrder) error
	GetByID(ctx context.Context, id string) (*domain.FulfillmentOrder, error)
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]domain.FulfillmentOrder, int, error)
	UpdateStatus(ctx context.Context, id string, status domain.FulfillmentStatus) error
	CreateItem(ctx context.Context, item *domain.FulfillmentItem) error
	GetItems(ctx context.Context, fulfillmentID string) ([]domain.FulfillmentItem, error)
	SaveEvent(ctx context.Context, event *domain.FulfillmentEvent) error
	GetEvents(ctx context.Context, fulfillmentID string) ([]domain.FulfillmentEvent, error)

	CreateWarehouse(ctx context.Context, w *domain.Warehouse) error
	GetWarehouse(ctx context.Context, id string) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]domain.Warehouse, error)
	GetWarehouseInventory(ctx context.Context, warehouseID string) ([]domain.WarehouseInventory, error)

	CreateReturn(ctx context.Context, r *domain.ReturnRequest) error
	GetReturn(ctx context.Context, id string) (*domain.ReturnRequest, error)
	ListReturns(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]domain.ReturnRequest, int, error)
	UpdateReturnStatus(ctx context.Context, id string, status domain.ReturnStatus, approvedBy *string) error

	SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error
	GetUnprocessedOutboxEvents(ctx context.Context) ([]domain.OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id string) error
	MarkOutboxEventFailed(ctx context.Context, id string) error
}

type FulfillmentService struct {
	repo         FulfillmentRepository
	redisStore   *redisinfra.Store
	kafkaProducer *kafka.Producer
	cfg          interface{ GetDefaultCarrier() string }
}

func NewFulfillmentService(repo FulfillmentRepository, redisStore *redisinfra.Store, kafkaProducer *kafka.Producer, carrier string) *FulfillmentService {
	return &FulfillmentService{
		repo:         repo,
		redisStore:   redisStore,
		kafkaProducer: kafkaProducer,
		cfg:          &carrierConfig{defaultCarrier: carrier},
	}
}

type carrierConfig struct {
	defaultCarrier string
}

func (c *carrierConfig) GetDefaultCarrier() string { return c.defaultCarrier }

func (s *FulfillmentService) CreateFulfillment(ctx context.Context, req *CreateFulfillmentRequest) (*domain.FulfillmentOrder, error) {
	f := domain.NewFulfillmentOrder(req.OrderID, req.SellerID, req.ShippingMethod, req.ShippingAddress, req.TotalItems, req.TotalWeightGrams)
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("create fulfillment: %w", err)
	}
	for i := range req.Items {
		item := domain.NewFulfillmentItem(f.ID, req.Items[i].OrderItemID, req.Items[i].ProductID, req.Items[i].SKUID, req.Items[i].Quantity)
		if err := s.repo.CreateItem(ctx, item); err != nil {
			return nil, fmt.Errorf("create fulfillment item: %w", err)
		}
		f.Items = append(f.Items, *item)
	}
	event := &domain.FulfillmentEvent{
		ID:            uuid.New().String(),
		FulfillmentID: f.ID,
		EventType:     domain.EventCreated,
		ActorID:       &req.SellerID,
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		zap.L().Warn("failed to save event", zap.Error(err))
	}
	s.publishOutbox(ctx, domain.EventFulfillmentCreated, f)
	return f, nil
}

func (s *FulfillmentService) GetFulfillment(ctx context.Context, id string) (*domain.FulfillmentOrder, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.GetItems(ctx, id)
	if err == nil {
		f.Items = items
	}
	return f, nil
}

func (s *FulfillmentService) ListFulfillments(ctx context.Context, status string, page, pageSize int) ([]domain.FulfillmentOrder, int, error) {
	filter := map[string]interface{}{}
	if status != "" {
		filter["status"] = status
	}
	return s.repo.List(ctx, filter, page, pageSize)
}

func (s *FulfillmentService) TransitionStatus(ctx context.Context, id string, target domain.FulfillmentStatus, actorID string) (*domain.FulfillmentOrder, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := f.TransitionTo(target); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, id, target); err != nil {
		return nil, err
	}
	eventType := mapStatusToEvent(target)
	event := &domain.FulfillmentEvent{
		ID:            uuid.New().String(),
		FulfillmentID: f.ID,
		EventType:     eventType,
		ActorID:       &actorID,
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		zap.L().Warn("failed to save event", zap.Error(err))
	}
	items, _ := s.repo.GetItems(ctx, id)
	f.Items = items

	eventName := mapStatusToEventName(target)
	s.publishOutbox(ctx, eventName, f)
	return f, nil
}

func mapStatusToEvent(target domain.FulfillmentStatus) domain.FulfillmentEventType {
	m := map[domain.FulfillmentStatus]domain.FulfillmentEventType{
		domain.FulfillmentStatusConfirmed: domain.EventConfirmed,
		domain.FulfillmentStatusPicking:   domain.EventPickStarted,
		domain.FulfillmentStatusPacked:    domain.EventPackCompleted,
		domain.FulfillmentStatusShipped:   domain.EventShipped,
		domain.FulfillmentStatusDelivered: domain.EventDelivered,
		domain.FulfillmentStatusReturned:  domain.EventReturned,
		domain.FulfillmentStatusCancelled: domain.EventCancelled,
		domain.FulfillmentStatusFailed:    domain.EventException,
	}
	if t, ok := m[target]; ok {
		return t
	}
	return domain.EventCreated
}

func mapStatusToEventName(target domain.FulfillmentStatus) domain.FulfillmentEventName {
	m := map[domain.FulfillmentStatus]domain.FulfillmentEventName{
		domain.FulfillmentStatusConfirmed: domain.EventFulfillmentConfirmed,
		domain.FulfillmentStatusPicking:   domain.EventFulfillmentPicking,
		domain.FulfillmentStatusPacked:    domain.EventFulfillmentPacked,
		domain.FulfillmentStatusShipped:   domain.EventFulfillmentShipped,
		domain.FulfillmentStatusDelivered: domain.EventFulfillmentDelivered,
		domain.FulfillmentStatusReturned:  domain.EventFulfillmentReturned,
		domain.FulfillmentStatusCancelled: domain.EventFulfillmentCancelled,
		domain.FulfillmentStatusFailed:    domain.EventFulfillmentFailed,
	}
	if t, ok := m[target]; ok {
		return t
	}
	return domain.EventFulfillmentCreated
}

func (s *FulfillmentService) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	return s.repo.ListWarehouses(ctx)
}

func (s *FulfillmentService) GetWarehouse(ctx context.Context, id string) (*domain.Warehouse, error) {
	return s.repo.GetWarehouse(ctx, id)
}

func (s *FulfillmentService) GetWarehouseInventory(ctx context.Context, warehouseID string) ([]domain.WarehouseInventory, error) {
	return s.repo.GetWarehouseInventory(ctx, warehouseID)
}

func (s *FulfillmentService) CreateReturn(ctx context.Context, req *CreateReturnRequest) (*domain.ReturnRequest, error) {
	r := domain.NewReturnRequest(req.FulfillmentID, req.OrderID, req.UserID, req.ReturnType, req.Reason, req.RefundAmount)
	if req.ReasonDetail != nil {
		r.ReasonDetail = req.ReasonDetail
	}
	if err := s.repo.CreateReturn(ctx, r); err != nil {
		return nil, fmt.Errorf("create return: %w", err)
	}
	return r, nil
}

func (s *FulfillmentService) GetReturn(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	return s.repo.GetReturn(ctx, id)
}

func (s *FulfillmentService) ListReturns(ctx context.Context, status string, page, pageSize int) ([]domain.ReturnRequest, int, error) {
	filter := map[string]interface{}{}
	if status != "" {
		filter["status"] = status
	}
	return s.repo.ListReturns(ctx, filter, page, pageSize)
}

func (s *FulfillmentService) ApproveReturn(ctx context.Context, id, approvedBy string) (*domain.ReturnRequest, error) {
	if err := s.repo.UpdateReturnStatus(ctx, id, domain.ReturnStatusApproved, &approvedBy); err != nil {
		return nil, err
	}
	return s.repo.GetReturn(ctx, id)
}

func (s *FulfillmentService) RejectReturn(ctx context.Context, id, rejectedBy string) (*domain.ReturnRequest, error) {
	if err := s.repo.UpdateReturnStatus(ctx, id, domain.ReturnStatusRejected, &rejectedBy); err != nil {
		return nil, err
	}
	return s.repo.GetReturn(ctx, id)
}

func (s *FulfillmentService) GetFulfillmentEvents(ctx context.Context, fulfillmentID string) ([]domain.FulfillmentEvent, error) {
	return s.repo.GetEvents(ctx, fulfillmentID)
}

func (s *FulfillmentService) ProcessOutboxEvents(ctx context.Context) error {
	events, err := s.repo.GetUnprocessedOutboxEvents(ctx)
	if err != nil {
		return err
	}
	for _, event := range events {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			zap.L().Warn("invalid outbox payload, marking failed", zap.String("id", event.ID), zap.Error(err))
			_ = s.repo.MarkOutboxEventFailed(ctx, event.ID)
			continue
		}
		if s.kafkaProducer != nil {
			if err := s.kafkaProducer.Publish(ctx, event.EventType, event.Payload); err != nil {
				zap.L().Warn("kafka publish failed", zap.String("id", event.ID), zap.Error(err))
				_ = s.repo.MarkOutboxEventFailed(ctx, event.ID)
				continue
			}
		}
		if err := s.repo.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			zap.L().Warn("failed to mark outbox processed", zap.String("id", event.ID), zap.Error(err))
		}
	}
	return nil
}

func (s *FulfillmentService) publishOutbox(ctx context.Context, eventType domain.FulfillmentEventName, f *domain.FulfillmentOrder) {
	payload, _ := json.Marshal(map[string]interface{}{
		"event_type":     string(eventType),
		"fulfillment_id": f.ID,
		"order_id":       f.OrderID,
		"status":         f.Status,
		"timestamp":      time.Now().UTC(),
	})
	_ = s.repo.SaveOutboxEvent(ctx, &domain.OutboxEvent{
		ID:        uuid.New().String(),
		EventType: string(eventType),
		Payload:   string(payload),
		Status:    domain.OutboxStatusPending,
		CreatedAt: time.Now().UTC(),
	})
}

type CreateFulfillmentRequest struct {
	OrderID          string                  `json:"order_id" binding:"required"`
	SellerID         string                  `json:"seller_id" binding:"required"`
	ShippingMethod   string                  `json:"shipping_method" binding:"required"`
	ShippingAddress  *domain.ShippingAddress `json:"shipping_address" binding:"required"`
	TotalItems       int                     `json:"total_items"`
	TotalWeightGrams int                     `json:"total_weight_grams"`
	Items            []CreateFulfillmentItem `json:"items"`
}

type CreateFulfillmentItem struct {
	OrderItemID string `json:"order_item_id" binding:"required"`
	ProductID   string `json:"product_id" binding:"required"`
	SKUID       string `json:"sku_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required"`
}

type CreateReturnRequest struct {
	FulfillmentID string            `json:"fulfillment_id" binding:"required"`
	OrderID       string            `json:"order_id" binding:"required"`
	UserID        string            `json:"-"`
	ReturnType    domain.ReturnType `json:"return_type" binding:"required"`
	Reason        domain.ReturnReason `json:"reason" binding:"required"`
	ReasonDetail  *string           `json:"reason_detail,omitempty"`
	RefundAmount  int64             `json:"refund_amount" binding:"required"`
}
