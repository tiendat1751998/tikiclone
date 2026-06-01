package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tikiclone/tiki/services/shipment/internal/config"
	"github.com/tikiclone/tiki/services/shipment/internal/domain"
	"github.com/tikiclone/tiki/services/shipment/internal/metrics"
)

type Producer struct {
	writer *kafka.Writer
	cfg    config.KafkaConfig
}

// [SECURITY] Whitelist of allowed event types to prevent topic injection attacks.
var allowedEventTypes = map[domain.ShipmentEventType]bool{
	domain.EventShipmentCreated:   true,
	domain.EventShipmentPickedUp:  true,
	domain.EventShipmentInTransit: true,
	domain.EventShipmentDelivered: true,
	domain.EventShipmentFailed:    true,
	domain.EventShipmentReturned:  true,
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			WriteTimeout: 10 * time.Second,
		},
		cfg: cfg,
	}
}

func (p *Producer) PublishEvent(ctx context.Context, event *domain.ShipmentEvent) error {
	payload, err := json.Marshal(event)
	if err != nil { return err }

	// [SECURITY] Validate event type against whitelist to prevent topic injection
	if !allowedEventTypes[event.EventType] {
		return fmt.Errorf("invalid event type: %s", event.EventType)
	}

	// [SECURITY] Use validated event type only
	topic := fmt.Sprintf("%s.%s", p.cfg.TopicPrefix, string(event.EventType))

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(event.ShipmentID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(string(event.EventType))},
		},
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
