package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/oms-fulfillment/internal/config"
)

type Producer struct {
	writer *kafka.Writer
	prefix string
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			Async:        true,
			RequiredAcks: kafka.RequireOne,
		},
		prefix: cfg.TopicPrefix,
	}
}

func (p *Producer) Publish(ctx context.Context, eventType, payload string) error {
	topic := p.prefix + "." + eventType
	msg := kafka.Message{
		Topic: topic,
		Value: []byte(payload),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(eventType)},
			{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type EventHandler interface {
	HandleEvent(ctx context.Context, eventType string, payload []byte) error
}

type Consumer struct {
	reader      *kafka.Reader
	dlqWriter   *kafka.Writer
	eventHandler EventHandler
}

func NewConsumer(cfg config.KafkaConfig, topics []string, handler EventHandler) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Brokers,
			GroupID:     cfg.ConsumerGroup,
			GroupTopics: topics,
			MinBytes:    1,
			MaxBytes:    10e6,
			MaxWait:     1 * time.Second,
			StartOffset: kafka.LastOffset,
		}),
		dlqWriter: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    cfg.DLQTopic,
			Balancer: &kafka.LeastBytes{},
			Async:    true,
		},
		eventHandler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		eventType := ""
		for _, h := range msg.Headers {
			if h.Key == "event_type" {
				eventType = string(h.Value)
				break
			}
		}

		if err := c.eventHandler.HandleEvent(ctx, eventType, msg.Value); err != nil {
			c.dlqWriter.WriteMessages(ctx, msg)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.dlqWriter.Close()
}
