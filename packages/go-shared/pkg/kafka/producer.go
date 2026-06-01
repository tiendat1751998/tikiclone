package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

type Producer struct {
	writer  *kafka.Writer
	service string
}

type Message struct {
	Key       string
	Topic     string
	Value     interface{}
	Headers   map[string]string
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

func (p *Producer) Publish(ctx context.Context, msg Message) error {
	valueBytes, err := json.Marshal(msg.Value)
	if err != nil {
		return err
	}

	headers := []kafka.Header{
		{Key: "service", Value: []byte(p.service)},
		{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
	}

	for k, v := range msg.Headers {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier{})

	kafkaMsg := kafka.Message{
		Topic:   msg.Topic,
		Key:     []byte(msg.Key),
		Value:   valueBytes,
		Headers: headers,
		Time:    time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		observability.LogWithTrace(ctx).Error("failed to publish kafka message",
			zap.String("topic", msg.Topic),
			zap.Error(err),
		)
		return err
	}

	observability.KafkaMessagesProduced.WithLabelValues(p.service, msg.Topic).Inc()

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
