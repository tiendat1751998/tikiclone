package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

const TopicRecEvents = "recommendation.events"

type Publisher interface {
	Publish(ctx context.Context, eventType EventType, payload interface{}) error
	Close() error
}

type KafkaProducer struct {
	writer  *kafka.Writer
	service string
}

func NewKafkaProducer(brokers []string, service string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        TopicRecEvents,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    100,
			Async:        false,
			RequiredAcks: kafka.RequireAll,
			MaxAttempts:  3,
		},
		service: service,
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, eventType EventType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(string(eventType)),
		Value: data,
		Headers: []kafka.Header{
			{Key: "service", Value: []byte(p.service)},
			{Key: "event_type", Value: []byte(string(eventType))},
		},
		Time: time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		observability.LogWithTrace(ctx).Error("failed to publish event", zap.Error(err))
		return err
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

type NoOpPublisher struct{}

func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

func (n *NoOpPublisher) Publish(ctx context.Context, eventType EventType, payload interface{}) error {
	return nil
}

func (n *NoOpPublisher) Close() error {
	return nil
}
