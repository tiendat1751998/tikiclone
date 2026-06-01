package scoring

import (
	"context"
	"math"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CalculateScore(ctx context.Context, evaluations []rules.RuleEvaluation) float64 {
	if len(evaluations) == 0 {
		return 0
	}

	var totalScore float64
	var totalWeight float64

	for _, eval := range evaluations {
		if eval.Triggered {
			contrib := float64(eval.Severity) * eval.Weight
			totalScore += contrib
			totalWeight += eval.Weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	score := (totalScore / (10 * totalWeight)) * 100

	return math.Round(score*100) / 100
}

func (s *Service) ClassifyRisk(ctx context.Context, score float64) core.RiskLevel {
	thresholds := s.repo.GetThresholds()

	switch {
	case score >= thresholds.Critical:
		return core.RiskCritical
	case score >= thresholds.High:
		return core.RiskHigh
	case score >= thresholds.Medium:
		return core.RiskMedium
	default:
		return core.RiskLow
	}
}

func (s *Service) GetScoreCard(ctx context.Context, evaluations []rules.RuleEvaluation) *ScoreCard {
	score := s.CalculateScore(ctx, evaluations)
	level := s.ClassifyRisk(ctx, score)

	var factors []WeightedFactor
	for _, eval := range evaluations {
		if eval.Triggered {
			factors = append(factors, WeightedFactor{
				Name:   eval.RuleName,
				Score:  float64(eval.Severity),
				Weight: eval.Weight,
			})
		}
	}

	return &ScoreCard{
		TotalScore: score,
		Level:      level,
		MaxScore:   100,
		Factors:    factors,
	}
}
