package detection

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
	"github.com/tikiclone/tiki/platforms/fraud/internal/scoring"
	"github.com/tikiclone/tiki/platforms/fraud/internal/streaming"
)

type Service struct {
	repo           Repository
	ruleService    *rules.Service
	scoringService *scoring.Service
	streamService  *streaming.Service
	blacklistSvc   *blacklist.Service
	threshold      float64
}

func NewService(repo Repository, rs *rules.Service, ss *scoring.Service, sts *streaming.Service, bs *blacklist.Service, threshold float64) *Service {
	return &Service{
		repo:           repo,
		ruleService:    rs,
		scoringService: ss,
		streamService:  sts,
		blacklistSvc:   bs,
		threshold:      threshold,
	}
}

func (s *Service) Evaluate(ctx context.Context, event *core.FraudEvent) (*RiskScore, error) {
	if event == nil || event.Type == "" {
		return nil, ErrEventInvalid
	}

	blacklistCheck := &blacklist.CheckRequest{
		UserID:   event.UserID,
		IP:       event.IP,
		DeviceID: event.DeviceID,
	}
	if hit, err := s.blacklistSvc.Check(ctx, blacklistCheck); err == nil && hit.Blocked {
		return nil, fmt.Errorf("fraud: entity blacklisted: %v", hit.Reasons)
	}

	allRules := s.ruleService.ListActive(ctx)
	var triggeredRules []rules.RuleEvaluation

	for _, rule := range allRules {
		s.streamService.ProcessEvent(ctx, event)
		eval := s.ruleService.EvaluateRule(ctx, &rule, event)
		triggeredRules = append(triggeredRules, eval)
	}

	score := s.scoringService.CalculateScore(ctx, triggeredRules)
	level := s.scoringService.ClassifyRisk(ctx, score)
	alertTriggered := score >= s.threshold

	var alertID string
	if alertTriggered {
		alertType := s.determineAlertType(event, triggeredRules)
		alert := &FraudAlert{
			ID:          uuid.New().String(),
			EventID:     event.ID,
			UserID:      event.UserID,
			Type:        alertType,
			RiskScore:   score,
			RiskLevel:   level,
			Description: s.buildAlertDescription(alertType, score, triggeredRules),
			Status:      "open",
			CreatedAt:   time.Now(),
		}
		if err := s.repo.SaveAlert(ctx, alert); err != nil {
			return nil, err
		}
		alertID = alert.ID
	}

	var results []RuleResult
	for _, tr := range triggeredRules {
		results = append(results, RuleResult{
			RuleName:  tr.RuleName,
			Severity:  tr.Severity,
			Weight:    tr.Weight,
			Score:     tr.Score,
			Triggered: tr.Triggered,
			Reason:    tr.Reason,
		})
	}

	return &RiskScore{
		Score:          score,
		Level:          level,
		MaxScore:       100,
		RuleResults:    results,
		EvaluatedAt:    time.Now(),
		AlertTriggered: alertTriggered,
		AlertID:        alertID,
	}, nil
}

func (s *Service) GetAlert(ctx context.Context, id string) (*FraudAlert, error) {
	return s.repo.GetAlert(ctx, id)
}

func (s *Service) ListAlerts(ctx context.Context, status string, riskLevel core.RiskLevel, offset, limit int) ([]*FraudAlert, int, error) {
	return s.repo.ListAlerts(ctx, status, riskLevel, offset, limit)
}

func (s *Service) ResolveAlert(ctx context.Context, id, resolvedBy, resolution string) error {
	alert, err := s.repo.GetAlert(ctx, id)
	if err != nil {
		return err
	}
	if alert.Status == "resolved" {
		return ErrAlertResolved
	}
	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy
	alert.Resolution = resolution
	return s.repo.UpdateAlert(ctx, alert)
}

func (s *Service) determineAlertType(event *core.FraudEvent, triggeredRules []rules.RuleEvaluation) AlertType {
	for _, tr := range triggeredRules {
		if tr.Triggered {
			switch tr.RuleName {
			case "new_device_login":
				return AlertNewDeviceLogin
			case "high_velocity":
				return AlertRapidFireOrders
			case "amount_anomaly":
				return AlertHighValueTransaction
			case "payment_fraud":
				return AlertPaymentFraud
			}
		}
	}
	if event.Type == core.EventLogin {
		return AlertAccountTakeover
	}
	return AlertPaymentFraud
}

func (s *Service) buildAlertDescription(alertType AlertType, score float64, triggeredRules []rules.RuleEvaluation) string {
	return fmt.Sprintf("Alert %s triggered with score %.1f", alertType, score)
}

func (s *Service) GetThreshold() float64 {
	return s.threshold
}
