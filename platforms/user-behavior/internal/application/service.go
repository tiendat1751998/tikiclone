package application
import ("context"; "fmt"; "time"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/domain"; "github.com/tikiclone/tiki/platforms/user-behavior/internal/metrics"; "github.com/tikiclone/tiki/packages/go-shared/pkg/observability"; "go.opentelemetry.io/otel"; "go.uber.org/zap")

type BehaviorService struct { publisher EventPublisher }
type EventPublisher interface { Publish(ctx context.Context, eventType string, payload interface{}) error }

func NewBehaviorService(pub EventPublisher) *BehaviorService { return &BehaviorService{publisher: pub} }

func (s *BehaviorService) IngestEvent(ctx context.Context, event *domain.ClickEvent) error {
	ctx, span := otel.Tracer("tiki-user-behavior").Start(ctx, "behavior.ingest"); defer span.End()
	if event.UserID == "" || event.EventType == "" { return domain.ErrEventValidation }
	if event.ID == "" { event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano()) }
	if event.Timestamp.IsZero() { event.Timestamp = time.Now() }
	metrics.EventsIngested.Inc()
	if s.publisher != nil { s.publisher.Publish(ctx, "behavior.event", event) }
	observability.LogWithTrace(ctx).Debug("event ingested", zap.String("type", event.EventType), zap.String("user", event.UserID))
	return nil
}

func (s *BehaviorService) IngestBatch(ctx context.Context, events []*domain.ClickEvent) (int, error) {
	success := 0
	for _, e := range events { if err := s.IngestEvent(ctx, e); err == nil { success++ } }
	metrics.BatchIngested.Add(float64(success))
	return success, nil
}

func (s *BehaviorService) GetTrendingProducts(ctx context.Context, limit int) ([]string, error) { return nil, nil }
func (s *BehaviorService) GetUserTimeline(ctx context.Context, userID string, limit int) ([]*domain.ClickEvent, error) { return nil, nil }
