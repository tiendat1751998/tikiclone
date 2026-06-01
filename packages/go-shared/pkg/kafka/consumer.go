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

type Consumer struct {
	reader  *kafka.Reader
	service string
}

type HandlerFunc func(ctx context.Context, msg Message) error

type ConsumerConfig struct {
	Brokers       []string
	Topic         string
	GroupID       string
	Service       string
	MinBytes      int
	MaxBytes      int
	MaxWait       time.Duration
	RetryMax      int
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	if cfg.MinBytes == 0 {
		cfg.MinBytes = 10e3
	}
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 10e6
	}
	if cfg.MaxWait == 0 {
		cfg.MaxWait = 1 * time.Second
	}
	if cfg.RetryMax == 0 {
		cfg.RetryMax = 3
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 1 * time.Second,
		RetentionTime:  24 * time.Hour,
		WatchPartitionChanges: true,
	})

	return &Consumer{
		reader:  reader,
		service: cfg.Service,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler HandlerFunc) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			kafkaMsg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				observability.GetLogger().Error("kafka read error", zap.Error(err))
				continue
			}

			msgCtx := c.extractContext(kafkaMsg)

			var payload interface{}
			if err := json.Unmarshal(kafkaMsg.Value, &payload); err != nil {
				observability.LogWithTrace(msgCtx).Error("failed to unmarshal message",
					zap.Error(err),
					zap.String("topic", kafkaMsg.Topic),
				)
				continue
			}

			msg := Message{
				Key:   string(kafkaMsg.Key),
				Topic: kafkaMsg.Topic,
				Value: payload,
			}

			if err := c.retryWithBackoff(msgCtx, handler, msg); err != nil {
				observability.KafkaMessagesConsumed.WithLabelValues(c.service, kafkaMsg.Topic, "failed").Inc()
				observability.LogWithTrace(msgCtx).Error("message processing failed after retries",
					zap.Error(err),
					zap.String("topic", kafkaMsg.Topic),
					zap.String("key", msg.Key),
				)
			} else {
				observability.KafkaMessagesConsumed.WithLabelValues(c.service, kafkaMsg.Topic, "success").Inc()
			}
		}
	}
}

func (c *Consumer) retryWithBackoff(ctx context.Context, handler HandlerFunc, msg Message) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := handler(ctx, msg); err != nil {
			lastErr = err
			time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Consumer) extractContext(msg kafka.Message) context.Context {
	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	ctx := context.Background()
	carrier := propagation.MapCarrier(headers)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	return ctx
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
