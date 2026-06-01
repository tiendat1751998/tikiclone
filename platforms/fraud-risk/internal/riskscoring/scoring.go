package riskscoring

import (
	"context"
	"math"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
)

type Calculator struct {
	repo Repository
}

func NewCalculator(repo Repository) *Calculator {
	return &Calculator{repo: repo}
}

func (c *Calculator) Calculate(ctx context.Context, factors []RiskFactor) *ScoreCard {
	totalScore := c.BaseScore()
	var normalized []RiskFactor

	for _, f := range factors {
		contrib := f.CurrentValue * f.Weight
		f.Contribution = math.Round(contrib*100) / 100
		totalScore += f.Contribution
		normalized = append(normalized, f)
	}

	if totalScore > 100 {
		totalScore = 100
	}
	if totalScore < 0 {
		totalScore = 0
	}
	totalScore = math.Round(totalScore*100) / 100

	level := c.Classify(totalScore)

	return &ScoreCard{
		Factors:    normalized,
		BaseScore:  c.BaseScore(),
		TotalScore: totalScore,
		MaxScore:   100,
		Level:      level,
	}
}

func (c *Calculator) BaseScore() float64 {
	return 0
}

func (c *Calculator) Classify(score float64) core.RiskLevel {
	thresholds := c.repo.GetThresholds()

	switch {
	case score > thresholds.High:
		return core.RiskCritical
	case score > thresholds.Medium:
		return core.RiskHigh
	case score > thresholds.Low:
		return core.RiskMedium
	case score > thresholds.Safe:
		return core.RiskLow
	default:
		return core.RiskSafe
	}
}

func (c *Calculator) ListFactors(ctx context.Context) []RiskFactor {
	return c.repo.ListFactors()
}
