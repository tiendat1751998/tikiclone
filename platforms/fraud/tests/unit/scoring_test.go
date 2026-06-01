package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
	"github.com/tikiclone/tiki/platforms/fraud/internal/scoring"
)

func TestCalculateScoreNoRules(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	score := svc.CalculateScore(context.Background(), nil)
	if score != 0 {
		t.Errorf("expected 0, got %f", score)
	}
}

func TestCalculateScoreNoTriggered(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	evaluations := []rules.RuleEvaluation{
		{RuleName: "r1", Severity: 5, Weight: 0.8, Triggered: false},
		{RuleName: "r2", Severity: 3, Weight: 0.5, Triggered: false},
	}

	score := svc.CalculateScore(context.Background(), evaluations)
	if score != 0 {
		t.Errorf("expected 0, got %f", score)
	}
}

func TestCalculateScoreSingleRule(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	evaluations := []rules.RuleEvaluation{
		{RuleName: "r1", Severity: 5, Weight: 1.0, Triggered: true, Score: 5.0},
	}

	score := svc.CalculateScore(context.Background(), evaluations)
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestCalculateScoreMultipleRules(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	evaluations := []rules.RuleEvaluation{
		{RuleName: "high_severity", Severity: 10, Weight: 1.0, Triggered: true, Score: 10.0},
		{RuleName: "medium_severity", Severity: 5, Weight: 0.8, Triggered: true, Score: 4.0},
		{RuleName: "not_triggered", Severity: 8, Weight: 0.9, Triggered: false},
	}

	score := svc.CalculateScore(context.Background(), evaluations)
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestClassifyRiskLow(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	level := svc.ClassifyRisk(context.Background(), 10)
	if level != core.RiskLow {
		t.Errorf("expected low, got %s", level)
	}
}

func TestClassifyRiskMedium(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	level := svc.ClassifyRisk(context.Background(), 30)
	if level != core.RiskMedium {
		t.Errorf("expected medium, got %s", level)
	}
}

func TestClassifyRiskHigh(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	level := svc.ClassifyRisk(context.Background(), 60)
	if level != core.RiskHigh {
		t.Errorf("expected high, got %s", level)
	}
}

func TestClassifyRiskCritical(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	level := svc.ClassifyRisk(context.Background(), 85)
	if level != core.RiskCritical {
		t.Errorf("expected critical, got %s", level)
	}
}

func TestClassifyRiskBoundaries(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	tests := []struct {
		score float64
		level core.RiskLevel
	}{
		{0, core.RiskLow},
		{25, core.RiskLow},
		{26, core.RiskMedium},
		{50, core.RiskMedium},
		{51, core.RiskHigh},
		{75, core.RiskHigh},
		{76, core.RiskCritical},
		{100, core.RiskCritical},
	}

	for _, tt := range tests {
		level := svc.ClassifyRisk(context.Background(), tt.score)
		if level != tt.level {
			t.Errorf("score %.0f: expected %s, got %s", tt.score, tt.level, level)
		}
	}
}

func TestGetScoreCard(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	evaluations := []rules.RuleEvaluation{
		{RuleName: "r1", Severity: 4, Weight: 0.5, Triggered: true, Score: 2.0},
	}

	card := svc.GetScoreCard(context.Background(), evaluations)
	if card.TotalScore <= 0 {
		t.Errorf("expected positive score, got %f", card.TotalScore)
	}
	if len(card.Factors) != 1 {
		t.Errorf("expected 1 factor, got %d", len(card.Factors))
	}
}

func TestSetCustomThresholds(t *testing.T) {
	repo := scoring.NewInMemoryRepository()
	svc := scoring.NewService(repo)

	repo.SetThresholds(&scoring.ScoreThreshold{Low: 0, Medium: 10, High: 30, Critical: 60})

	level := svc.ClassifyRisk(context.Background(), 50)
	if level != core.RiskHigh {
		t.Errorf("expected high with custom thresholds, got %s", level)
	}

	level = svc.ClassifyRisk(context.Background(), 70)
	if level != core.RiskCritical {
		t.Errorf("expected critical with custom thresholds, got %s", level)
	}
}
