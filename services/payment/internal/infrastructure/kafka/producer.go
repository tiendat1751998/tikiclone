package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	"github.com/tikiclone/tiki/services/payment/internal/metrics"
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

func (p *Producer) PublishEvent(ctx context.Context, event *domain.PaymentEvent) error {
	payload, err := json.Marshal(event)
	if err != nil { return err }
	topic := fmt.Sprintf("%s.%s", p.cfg.TopicPrefix, string(event.EventType))
	msg := kafka.Message{
		Topic:   topic,
		Key:     []byte(event.PaymentID),
		Value:   payload,
		Headers: []kafka.Header{{Key: "event_type", Value: []byte(string(event.EventType))}},
	}
	start := time.Now()
	err = p.writer.WriteMessages(ctx, msg)
	metrics.KafkaPublishLatency.WithLabelValues(string(event.EventType)).Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.KafkaPublishErrors.WithLabelValues(string(event.EventType)).Inc()
		return err
	}
	return nil
}

func (p *Producer) Close() error { return p.writer.Close() }
