package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/payment/internal/config"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, eventType string, payload []byte) error

type Consumer struct {
	reader   *kafka.Reader
	dlqWriter *kafka.Writer
	cfg      config.KafkaConfig
	handler  EventHandler
}

func NewConsumer(cfg config.KafkaConfig, topics []string, handler EventHandler) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers, GroupID: cfg.ConsumerGroup, GroupTopics: topics,
		MinBytes: 1e3, MaxBytes: 10e6, MaxWait: 500 * time.Millisecond,
	})
	var dlqWriter *kafka.Writer
	if cfg.DLQTopic != "" {
		dlqWriter = &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.DLQTopic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			WriteTimeout: 10 * time.Second,
			Async:        false,
		}
	}
	return &Consumer{reader: reader, dlqWriter: dlqWriter, cfg: cfg, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil { return ctx.Err() }
			zap.L().Error("failed to fetch kafka message", zap.Error(err))
			continue
		}
		var eventType string
		for _, h := range msg.Headers {
			if h.Key == "event_type" { eventType = string(h.Value); break }
		}
		if err := c.handler(ctx, eventType, msg.Value); err != nil {
			zap.L().Error("failed to handle kafka message", zap.String("event_type", eventType), zap.Error(err))
			c.sendToDLQ(ctx, msg, eventType, err)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				zap.L().Error("failed to commit kafka message after DLQ", zap.Error(err))
			}
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			zap.L().Error("failed to commit kafka message", zap.Error(err))
		}
	}
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, eventType string, handlerErr error) {
	if c.dlqWriter == nil {
		zap.L().Warn("DLQ not configured, discarding failed message", zap.String("event_type", eventType))
		return
	}
	dlqMsg := kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers,
			kafka.Header{Key: "original_error", Value: []byte(handlerErr.Error())},
			kafka.Header{Key: "original_topic", Value: []byte(msg.Topic)},
			kafka.Header{Key: "original_partition", Value: []byte(time.Now().String())},
		),
		Time: time.Now(),
	}
	dlqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.dlqWriter.WriteMessages(dlqCtx, dlqMsg); err != nil {
		zap.L().Error("failed to write to DLQ", zap.Error(err))
	}
}

func (c *Consumer) Close() error {
	if c.dlqWriter != nil {
		c.dlqWriter.Close()
	}
	return c.reader.Close()
}
