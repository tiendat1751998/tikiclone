package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, eventType EventType, data []byte) error

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type KafkaConsumer struct {
	reader  *kafka.Reader
	handler EventHandler
	service string
}

func NewKafkaConsumer(brokers []string, groupID string, handler EventHandler, service string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          TopicRecEvents,
			GroupID:        groupID,
			MinBytes:       10,
			MaxBytes:       10e6,
			MaxWait:        1 * time.Second,
			StartOffset:    kafka.LastOffset,
			CommitInterval: 1 * time.Second,
		}),
		handler: handler,
		service: service,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					observability.LogWithTrace(ctx).Error("failed to read message", zap.Error(err))
					continue
				}

				eventType := EventType(msg.Key)
				if c.handler != nil {
					if err := c.handler(ctx, eventType, msg.Value); err != nil {
						observability.LogWithTrace(ctx).Error("failed to handle event",
							zap.String("event_type", string(eventType)),
							zap.Error(err))
					}
				}
			}
		}
	}()
	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

type NoOpConsumer struct{}

func NewNoOpConsumer() *NoOpConsumer {
	return &NoOpConsumer{}
}

func (n *NoOpConsumer) Start(ctx context.Context) error {
	return nil
}

func (n *NoOpConsumer) Close() error {
	return nil
}

type FeedbackIngestionHandler struct {
	processFunc func(ctx context.Context, eventType EventType, data []byte) error
}

func NewFeedbackIngestionHandler(processFunc func(ctx context.Context, eventType EventType, data []byte) error) *FeedbackIngestionHandler {
	return &FeedbackIngestionHandler{processFunc: processFunc}
}

func (h *FeedbackIngestionHandler) Handle(ctx context.Context, eventType EventType, data []byte) error {
	if h.processFunc != nil {
		return h.processFunc(ctx, eventType, data)
	}

	switch eventType {
	case EventItemClicked:
		var event ItemClicked
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		observability.LogWithTrace(ctx).Info("click event received",
			zap.String("user_id", event.UserID),
			zap.String("product_id", event.ProductID))
	case EventItemPurchased:
		var event ItemPurchased
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		observability.LogWithTrace(ctx).Info("purchase event received",
			zap.String("user_id", event.UserID),
			zap.String("product_id", event.ProductID))
	}

	return nil
}
