package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

const (
	TopicInventoryEvents = "inventory.events"
	TopicStockMovements  = "inventory.stock-movements"
)

type Producer struct {
	writer  *kafka.Writer
	service string
}

func NewProducer(brokers []string, service string) *Producer {
	writer := &kafka.Writer{
		Addr:          kafka.TCP(brokers...),
		Balancer:      &kafka.Hash{},
		BatchTimeout:  10 * time.Millisecond,
		WriteTimeout:  10 * time.Second,
		BatchSize:     100,
		Async:         false,
		RequiredAcks:  kafka.RequireAll,
		MaxAttempts:   3,
	}

	return &Producer{
		writer:  writer,
		service: service,
	}
}

func (p *Producer) Publish(ctx context.Context, event *domain.InventoryEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := TopicInventoryEvents
	switch event.EventType {
	case domain.EventStockReserved, domain.EventStockReleased, domain.EventStockDeducted,
		domain.EventReservationExpired:
		topic = TopicInventoryEvents
	case domain.EventStockReplenished, domain.EventReconciliationTriggered:
		topic = TopicStockMovements
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(event.SkuID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "service", Value: []byte(p.service)},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
		Time: time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		observability.LogWithTrace(ctx).Error("failed to publish event",
			zap.String("event_type", string(event.EventType)),
			zap.String("topic", topic),
			zap.Error(err),
		)
		return err
	}

	observability.KafkaMessagesProduced.WithLabelValues(p.service, topic).Inc()
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
