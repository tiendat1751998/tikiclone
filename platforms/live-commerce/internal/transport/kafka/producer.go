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

const (
	TopicLiveEvents   = "live-commerce.events"
	TopicChatMessages = "live-commerce.chat"
	TopicAnalytics    = "live-commerce.analytics"
)

type Producer struct {
	eventsWriter    *kafka.Writer
	chatWriter      *kafka.Writer
	analyticsWriter *kafka.Writer
	service         string
}

func NewProducer(brokers []string, service string) *Producer {
	return &Producer{
		eventsWriter:    newWriter(brokers, TopicLiveEvents),
		chatWriter:      newWriter(brokers, TopicChatMessages),
		analyticsWriter: newWriter(brokers, TopicAnalytics),
		service:         service,
	}
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 5 * time.Millisecond,
		WriteTimeout: 10 * time.Second,
		BatchSize:    100,
		Async:        false,
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  3,
		// Tuned: increased buffer for flash-sale traffic spikes
		BatchSize:    256,
		BatchTimeout: 2 * time.Millisecond,
	}
}

func (p *Producer) Publish(ctx context.Context, eventType string, payload interface{}) error {
	body, err := sonic.Marshal(map[string]interface{}{
		"type":    eventType,
		"payload": payload,
		"time":    time.Now().UnixMilli(),
	})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(eventType),
		Value: body,
		Headers: []kafka.Header{
			{Key: "service", Value: []byte(p.service)},
			{Key: "event_type", Value: []byte(eventType)},
		},
		Time: time.Now(),
	}
	var writer *kafka.Writer
	switch eventType {
	case "chat.message.sent", "chat.message.deleted":
		writer = p.chatWriter
	default:
		writer = p.eventsWriter
	}
	if err := writer.WriteMessages(ctx, msg); err != nil {
		observability.LogWithTrace(ctx).Error("kafka publish failed",
			zap.String("event", eventType), zap.Error(err))
		return err
	}
	observability.KafkaMessagesProduced.WithLabelValues(p.service, writer.Topic).Inc()
	return nil
}

func (p *Producer) PublishAnalytics(ctx context.Context, payload interface{}) error {
	body, err := sonic.Marshal(payload)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Key:   []byte("analytics"),
		Value: body,
		Headers: []kafka.Header{
			{Key: "service", Value: []byte(p.service)},
		},
		Time: time.Now(),
	}
	return p.analyticsWriter.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	if err := p.eventsWriter.Close(); err != nil {
		return err
	}
	if err := p.chatWriter.Close(); err != nil {
		return err
	}
	return p.analyticsWriter.Close()
}
