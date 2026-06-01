package application
import ("context"; "fmt"; "time"; "github.com/tikiclone/tiki/platforms/fraud/internal/domain"; "github.com/tikiclone/tiki/platforms/fraud/internal/metrics"; "go.opentelemetry.io/otel")

type FraudService struct { publisher EventPublisher }
type EventPublisher interface { Publish(ctx context.Context, eventType string, payload interface{}) error }
func NewFraudService(pub EventPublisher) *FraudService { return &FraudService{publisher: pub} }

func (s *FraudService) ScoreTransaction(ctx context.Context, userID, orderID string, amount int64, deviceIP string) (*domain.FraudScore, error) {
	ctx, span := otel.Tracer("tiki-fraud").Start(ctx, "fraud.score"); defer span.End()
	score := &domain.FraudScore{
		ID: fmt.Sprintf("fraud_%d", time.Now().UnixNano()), UserID: userID, OrderID: orderID,
		Score: 0.1, RiskLevel: domain.RiskLow, Signals: []string{}, CreatedAt: time.Now(),
	}
	metrics.FraudScoresComputed.Inc()
	if s.publisher != nil { s.publisher.Publish(ctx, "fraud.scored", score) }
	return score, nil
}

func (s *FraudService) CreateCase(ctx context.Context, userID, orderID string, score float64) (*domain.FraudCase, error) {
	c := &domain.FraudCase{
		ID: fmt.Sprintf("case_%d", time.Now().UnixNano()), UserID: userID, OrderID: orderID,
		Status: domain.CaseStatusOpen, Score: score, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	metrics.FraudCasesCreated.Inc()
	return c, nil
}

func (s *FraudService) EvaluateRules(ctx context.Context, userID string, eventData map[string]interface{}) ([]string, error) { return nil, nil }
