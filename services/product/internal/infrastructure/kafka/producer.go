package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

// Producer wraps kafka.Writer for publishing events with retry and DLQ support
type Producer struct {
	writer    *kafka.Writer
	dlqWriter *kafka.Writer
}

// NewProducer creates a new Kafka producer with DLQ support
func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 10 * time.Second,
		Async:        false,
		// [RELIABILITY] Require all in-sync replicas for durability
		RequiredAcks: kafka.RequireAll,
		// [RELIABILITY] Retry failed writes
		MaxAttempts: 3,
	}

	// DLQ writer for failed messages
	dlqWriter := &kafka.Writer{
		Addr:            kafka.TCP(brokers...),
		Balancer:        &kafka.LeastBytes{},
		BatchTimeout:    10 * time.Millisecond,
		WriteTimeout:    10 * time.Second,
		Async:           false,
		RequiredAcks:    kafka.RequireAll,
		MaxAttempts:     1, // Don't retry DLQ writes
	}

	return &Producer{
		writer:    writer,
		dlqWriter: dlqWriter,
	}
}

// Publish publishes a message to a Kafka topic with retry and DLQ fallback
func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("product-service")},
			{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	// Attempt to write to main topic (with built-in retries)
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		observability.GetLogger().Error("failed to publish to main topic, sending to DLQ",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		// Send to DLQ
		dlqTopic := topic + ".dlq"
		dlqMsg := kafka.Message{
			Topic: dlqTopic,
			Key:   []byte(key),
			Value: payload,
			Time:  time.Now(),
			Headers: append(msg.Headers,
				kafka.Header{Key: "dlq_reason", Value: []byte(err.Error())},
				kafka.Header{Key: "original_topic", Value: []byte(topic)},
			),
		}
		if dlqErr := p.dlqWriter.WriteMessages(ctx, dlqMsg); dlqErr != nil {
			observability.GetLogger().Error("failed to send to DLQ",
				zap.String("dlq_topic", dlqTopic),
				zap.Error(dlqErr),
			)
			return fmt.Errorf("main topic: %w, dlq: %v", err, dlqErr)
		}
		return fmt.Errorf("message sent to DLQ %s: %w", dlqTopic, err)
	}

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return err
	}
	return p.dlqWriter.Close()
}

// EventPublisher interface for domain events
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

// Ensure interface is satisfied
var _ EventPublisher = (*Producer)(nil)

// MarshalEvent is a helper to marshal events
func MarshalEvent(event interface{}) ([]byte, error) {
	return sonic.Marshal(event)
}
