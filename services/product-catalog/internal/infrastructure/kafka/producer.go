package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/product-catalog/internal/config"
	"github.com/tikiclone/tiki/services/product-catalog/internal/domain"
	"github.com/tikiclone/tiki/services/product-catalog/internal/metrics"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	cfg    config.KafkaConfig
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:          kafka.TCP(cfg.Brokers...),
			Balancer:      &kafka.LeastBytes{},
			BatchTimeout:  10 * time.Millisecond,
			WriteTimeout:  10 * time.Second,
		},
		cfg: cfg,
	}
}

func (p *Producer) PublishCatalogEvent(ctx context.Context, event *domain.CatalogEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := fmt.Sprintf("%s.%s", p.cfg.TopicPrefix, event.EventType)
	msg := kafka.Message{
		Topic:   topic,
		Key:     []byte(event.AggregateID),
		Value:   payload,
		Headers: []kafka.Header{{Key: "event_type", Value: []byte(event.EventType)}},
	}

	start := time.Now()
	err = p.writer.WriteMessages(ctx, msg)
	metrics.KafkaPublishLatency.WithLabelValues(string(event.EventType)).Observe(time.Since(start).Seconds())

	if err != nil {
		metrics.KafkaPublishErrors.WithLabelValues(string(event.EventType)).Inc()
		return fmt.Errorf("failed to publish catalog event: %w", err)
	}

	zap.L().Info("published catalog event",
		zap.String("topic", topic),
		zap.String("event_type", string(event.EventType)),
		zap.String("aggregate_id", event.AggregateID),
	)
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
