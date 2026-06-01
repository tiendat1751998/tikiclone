package ruleengine

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
)

type Engine struct {
	ruleRepo    RuleRepository
	rulesetRepo RuleSetRepository
}

func NewEngine(ruleRepo RuleRepository, rulesetRepo RuleSetRepository) *Engine {
	return &Engine{
		ruleRepo:    ruleRepo,
		rulesetRepo: rulesetRepo,
	}
}

func (e *Engine) CreateRule(ctx context.Context, rule *Rule) error {
	return e.ruleRepo.Create(ctx, rule)
}

func (e *Engine) GetRule(ctx context.Context, id string) (*Rule, error) {
	return e.ruleRepo.Get(ctx, id)
}

func (e *Engine) ListRules(ctx context.Context) ([]*Rule, error) {
	return e.ruleRepo.List(ctx)
}

func (e *Engine) CreateRuleSet(ctx context.Context, rs *RuleSet) error {
	return e.rulesetRepo.Create(ctx, rs)
}

func (e *Engine) GetRuleSet(ctx context.Context, id string) (*RuleSet, error) {
	return e.rulesetRepo.Get(ctx, id)
}

func (e *Engine) EvaluateRule(ctx context.Context, rule *Rule, ev *core.Event) RuleEvaluation {
	eval := RuleEvaluation{
		RuleID:   rule.ID,
		RuleName: rule.Name,
	}

	if !rule.IsActive {
		eval.Triggered = false
		return eval
	}

	triggered, reason := evaluateCondition(rule.ConditionExpr, ev)
	eval.Triggered = triggered
	eval.Reason = reason
	if triggered {
		eval.Score = rule.ScoreDelta
	}

	return eval
}

func (e *Engine) EvaluateEvent(ctx context.Context, ev *core.Event) ([]RuleEvaluation, error) {
	rules, err := e.ruleRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var results []RuleEvaluation
	for _, rule := range rules {
		eval := e.EvaluateRule(ctx, rule, ev)
		results = append(results, eval)
	}

	return results, nil
}

func (e *Engine) EvaluateRuleSet(ctx context.Context, rulesetID string, ev *core.Event) (*RuleSetEvaluation, error) {
	rs, err := e.rulesetRepo.Get(ctx, rulesetID)
	if err != nil {
		return nil, err
	}

	var evals []RuleEvaluation
	for _, rule := range rs.Rules {
		r := rule
		eval := e.EvaluateRule(ctx, &r, ev)
		evals = append(evals, eval)
	}

	totalScore := 0.0
	passed := false

	switch rs.Strategy {
	case StrategyMatchAll:
		passed = true
		for _, e := range evals {
			if !e.Triggered {
				passed = false
			}
			totalScore += e.Score
		}
	case StrategyMatchAny:
		for _, e := range evals {
			totalScore += e.Score
			if e.Triggered {
				passed = true
			}
		}
	case StrategyWeightedSum:
		var sum float64
		for _, e := range evals {
			if e.Triggered {
				sum += e.Score
			}
		}
		totalScore = sum
		passed = totalScore > 0
	}

	totalScore = math.Round(totalScore*100) / 100

	return &RuleSetEvaluation{
		RuleSetID:   rs.ID,
		RuleSetName: rs.Name,
		Strategy:    rs.Strategy,
		Evaluations: evals,
		TotalScore:  totalScore,
		Passed:      passed,
	}, nil
}

func evaluateCondition(condition string, ev *core.Event) (bool, string) {
	switch {
	case condition == "new_device_login":
		if ev.DeviceID != "" && ev.Type == core.EventLogin {
			return true, "new device detected for login"
		}
	case condition == "high_amount":
		if ev.Amount > 10000 {
			return true, "amount exceeds high threshold"
		}
	case condition == "payment_fraud":
		if ev.Type == core.EventPayment {
			return true, "payment event flagged"
		}
	case condition == "foreign_ip":
		if ev.IP != "" && !strings.HasPrefix(ev.IP, "192.168.") && !strings.HasPrefix(ev.IP, "10.") {
			return true, "foreign ip detected"
		}
	case condition == "rapid_succession":
		return true, "velocity check triggered"
	default:
		if strings.HasPrefix(condition, "amount>") {
			parts := strings.SplitN(condition, ">", 2)
			if len(parts) == 2 {
				if threshold, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					if ev.Amount > threshold {
						return true, "amount exceeds threshold"
					}
				}
			}
		}
		if strings.HasPrefix(condition, "age<") {
			parts := strings.SplitN(condition, "<", 2)
			if len(parts) == 2 {
				if maxAge, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					if !ev.Timestamp.IsZero() && time.Since(ev.Timestamp).Seconds() < float64(maxAge) {
						return true, "event age below threshold"
					}
				}
			}
		}
	}

	return false, ""
}
