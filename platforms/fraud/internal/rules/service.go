package rules

import (
	"context"
	"strconv"
	"strings"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRule(ctx context.Context, rule *RuleDefinition) error {
	return s.repo.Create(ctx, rule)
}

func (s *Service) UpdateRule(ctx context.Context, rule *RuleDefinition) error {
	return s.repo.Update(ctx, rule)
}

func (s *Service) GetRule(ctx context.Context, id string) (*RuleDefinition, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) ListRules(ctx context.Context) ([]*RuleDefinition, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListActive(ctx context.Context) []RuleDefinition {
	rules, err := s.repo.List(ctx)
	if err != nil {
		return nil
	}
	var active []RuleDefinition
	for _, r := range rules {
		if r.IsActive {
			active = append(active, *r)
		}
	}
	return active
}

func (s *Service) ToggleRule(ctx context.Context, id string) (*RuleDefinition, error) {
	rule, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	rule.IsActive = !rule.IsActive
	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) EvaluateRule(ctx context.Context, rule *RuleDefinition, event *core.FraudEvent) RuleEvaluation {
	eval := RuleEvaluation{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Severity: rule.Severity,
		Weight:   rule.Weight,
	}

	if !rule.IsActive {
		eval.Triggered = false
		return eval
	}

	triggered, reason := evaluateCondition(rule.Condition, event)
	eval.Triggered = triggered
	eval.Reason = reason
	if triggered {
		eval.Score = float64(rule.Severity) * rule.Weight
	}

	return eval
}

func evaluateCondition(condition string, event *core.FraudEvent) (bool, string) {
	switch condition {
	case "new_device_login":
		if event.DeviceID != "" && event.Type == core.EventLogin {
			return true, "new device detected for login"
		}
	case "high_velocity":
		return true, "velocity check triggered"
	case "amount_anomaly":
		if event.Amount > 10000 {
			return true, "amount exceeds anomaly threshold"
		}
	case "payment_fraud":
		if event.Type == core.EventPayment {
			return true, "payment event flagged for review"
		}
	default:
		if strings.HasPrefix(condition, "amount>") {
			parts := strings.SplitN(condition, ">", 2)
			if len(parts) == 2 {
				if threshold, err := strconv.ParseFloat(parts[1], 64); err == nil {
					if event.Amount > threshold {
						return true, "amount exceeds threshold"
					}
				}
			}
		}
	}
	return false, ""
}
