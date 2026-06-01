package kafka
import ("context"; "encoding/json"; "time"; "github.com/segmentio/kafka-go"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.uber.org/zap")
const TopicSearchEvents = "search.events"
type Producer struct { writer *kafka.Writer; service string }
func NewProducer(brokers []string, service string) *Producer {
	return &Producer{writer: &kafka.Writer{Addr: kafka.TCP(brokers...), Balancer: &kafka.Hash{}, BatchTimeout: 10 * time.Millisecond, WriteTimeout: 10 * time.Second, BatchSize: 100, Async: false, RequiredAcks: kafka.RequireAll, MaxAttempts: 3}, service: service}
}
func (p *Producer) Publish(ctx context.Context, eventType string, payload interface{}) error {
	data, _ := json.Marshal(payload)
	msg := kafka.Message{Topic: TopicSearchEvents, Key: []byte(eventType), Value: data, Headers: []kafka.Header{{Key: "service", Value: []byte(p.service)}, {Key: "event_type", Value: []byte(eventType)}}, Time: time.Now()}
	if err := p.writer.WriteMessages(ctx, msg); err != nil { observability.LogWithTrace(ctx).Error("failed to publish", zap.Error(err)); return err }
	observability.KafkaMessagesProduced.WithLabelValues(p.service, TopicSearchEvents).Inc(); return nil
}
func (p *Producer) Close() error { return p.writer.Close() }
