package kafka

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/application"
	"go.uber.org/zap"
)

type Consumer struct {
	reader  *kafka.Reader
	service *application.LiveCommerceService
}

func NewConsumer(brokers []string, groupID, topic string, service *application.LiveCommerceService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			MinBytes:       10,
			MaxBytes:       10e6,
			MaxWait:        500 * time.Millisecond,
			StartOffset:    kafka.LastOffset,
			CommitInterval: time.Second,
		}),
		service: service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // goroutine exits cleanly on context cancellation
				}
				observability.LogWithTrace(ctx).Error("consumer read", zap.Error(err))
				continue
			}
			c.processMessage(ctx, msg)
		}
	}()
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var event struct {
		Type    string          `json:"type"`
		Payload sonic.RawMessage `json:"payload"`
	}
	if err := sonic.Unmarshal(msg.Value, &event); err != nil {
		observability.LogWithTrace(ctx).Warn("invalid event", zap.Error(err))
		return
	}
	logger := observability.LogWithTrace(ctx).With(
		zap.String("event_type", event.Type),
		zap.String("topic", msg.Topic),
	)
	logger.Info("processing event")
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
