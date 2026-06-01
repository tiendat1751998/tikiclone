package kafka
import ("context"; "encoding/json"; "time"; "github.com/segmentio/kafka-go"; "github.com/tikiclone/tiki/services/product-catalog/internal/domain"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.uber.org/zap")
const TopicCatalogEvents = "catalog.events"
type Producer struct { writer *kafka.Writer; service string }
func NewProducer(brokers []string, service string) *Producer {
	return &Producer{writer: &kafka.Writer{Addr: kafka.TCP(brokers...), Balancer: &kafka.Hash{}, BatchTimeout: 10 * time.Millisecond, WriteTimeout: 10 * time.Second, BatchSize: 100, Async: false, RequiredAcks: kafka.RequireAll, MaxAttempts: 3}, service: service}
}
func (p *Producer) Publish(ctx context.Context, event *domain.CatalogEvent) error {
	data, _ := json.Marshal(event)
	msg := kafka.Message{Topic: TopicCatalogEvents, Key: []byte(event.AggregateID), Value: data, Headers: []kafka.Header{{Key: "service", Value: []byte(p.service)}, {Key: "event_type", Value: []byte(event.EventType)}}, Time: time.Now()}
	if err := p.writer.WriteMessages(ctx, msg); err != nil { observability.LogWithTrace(ctx).Error("failed to publish", zap.Error(err)); return err }
	observability.KafkaMessagesProduced.WithLabelValues(p.service, TopicCatalogEvents).Inc(); return nil
}
func (p *Producer) Close() error { return p.writer.Close() }
