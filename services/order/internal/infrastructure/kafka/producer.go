package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/order/internal/config"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	"github.com/tikiclone/tiki/services/order/internal/metrics"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	cfg    config.KafkaConfig
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:          kafka.TCP(cfg.Brokers...),
		Balancer:      &kafka.LeastBytes{},
		BatchTimeout:  10 * time.Millisecond,
		WriteTimeout:  10 * time.Second,
		Async:         false,
	}
	return &Producer{writer: writer, cfg: cfg}
}

func (p *Producer) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	topic := fmt.Sprintf("%s.%s", p.cfg.TopicPrefix, event.EventType)
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(event.OrderID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "trace_id", Value: []byte(event.OrderID)},
			{Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
		},
	}

	start := time.Now()
	err = p.writer.WriteMessages(ctx, msg)
	duration := time.Since(start).Seconds()

	if event.EventType != "" {
		metrics.KafkaPublishLatency.WithLabelValues(string(event.EventType)).Observe(duration)
	}

	if err != nil {
		if event.EventType != "" {
			metrics.KafkaPublishErrors.WithLabelValues(string(event.EventType)).Inc()
		}
		return fmt.Errorf("failed to publish event: %w", err)
	}

	zap.L().Info("published kafka event",
		zap.String("topic", topic),
		zap.String("event_type", string(event.EventType)),
		zap.String("order_id", event.OrderID),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
