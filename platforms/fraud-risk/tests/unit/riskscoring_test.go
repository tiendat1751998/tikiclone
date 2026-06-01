package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/riskscoring"
)

func TestCalculateRiskScoreNoFactors(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	card := calc.Calculate(context.Background(), nil)
	if card.TotalScore != 0 {
		t.Errorf("expected 0, got %f", card.TotalScore)
	}
	if card.Level != core.RiskSafe {
		t.Errorf("expected safe, got %s", card.Level)
	}
}

func TestCalculateRiskScoreSingleFactor(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	factors := []riskscoring.RiskFactor{
		{Name: "device_risk", Weight: 1.0, CurrentValue: 50},
	}

	card := calc.Calculate(context.Background(), factors)
	if card.TotalScore <= 0 {
		t.Errorf("expected positive score, got %f", card.TotalScore)
	}
	if len(card.Factors) != 1 {
		t.Errorf("expected 1 factor, got %d", len(card.Factors))
	}
}

func TestCalculateRiskScoreMultipleFactors(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	factors := []riskscoring.RiskFactor{
		{Name: "device_risk", Weight: 0.5, CurrentValue: 80},
		{Name: "location_risk", Weight: 0.3, CurrentValue: 60},
		{Name: "behavior_risk", Weight: 0.2, CurrentValue: 40},
	}

	card := calc.Calculate(context.Background(), factors)
	if card.TotalScore <= 0 {
		t.Errorf("expected positive score, got %f", card.TotalScore)
	}
	if len(card.Factors) != 3 {
		t.Errorf("expected 3 factors, got %d", len(card.Factors))
	}
	if card.TotalScore > 100 {
		t.Errorf("expected score capped at 100, got %f", card.TotalScore)
	}
}

func TestCalculateRiskScoreCapsAt100(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	factors := []riskscoring.RiskFactor{
		{Name: "extreme", Weight: 2.0, CurrentValue: 100},
	}

	card := calc.Calculate(context.Background(), factors)
	if card.TotalScore > 100 {
		t.Errorf("expected score capped at 100, got %f", card.TotalScore)
	}
}

func TestClassifyRiskSafe(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	level := calc.Classify(10)
	if level != core.RiskSafe {
		t.Errorf("expected safe, got %s", level)
	}
}

func TestClassifyRiskLow(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	level := calc.Classify(30)
	if level != core.RiskLow {
		t.Errorf("expected low, got %s", level)
	}
}

func TestClassifyRiskMedium(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	level := calc.Classify(50)
	if level != core.RiskMedium {
		t.Errorf("expected medium, got %s", level)
	}
}

func TestClassifyRiskHigh(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	level := calc.Classify(70)
	if level != core.RiskHigh {
		t.Errorf("expected high, got %s", level)
	}
}

func TestClassifyRiskCritical(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	level := calc.Classify(90)
	if level != core.RiskCritical {
		t.Errorf("expected critical, got %s", level)
	}
}

func TestClassifyRiskBoundaries(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	tests := []struct {
		score float64
		level core.RiskLevel
	}{
		{0, core.RiskSafe},
		{20, core.RiskSafe},
		{21, core.RiskLow},
		{40, core.RiskLow},
		{41, core.RiskMedium},
		{60, core.RiskMedium},
		{61, core.RiskHigh},
		{80, core.RiskHigh},
		{81, core.RiskCritical},
		{100, core.RiskCritical},
	}

	for _, tt := range tests {
		level := calc.Classify(tt.score)
		if level != tt.level {
			t.Errorf("score %.0f: expected %s, got %s", tt.score, tt.level, level)
		}
	}
}

func TestUpsertAndListFactors(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()

	repo.UpsertFactor(riskscoring.RiskFactor{Name: "f1", Weight: 0.5, CurrentValue: 30})
	repo.UpsertFactor(riskscoring.RiskFactor{Name: "f2", Weight: 0.5, CurrentValue: 70})

	factors := repo.ListFactors()
	if len(factors) != 2 {
		t.Errorf("expected 2 factors, got %d", len(factors))
	}
}

func TestSetCustomThresholds(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	repo.SetThresholds(&riskscoring.RiskLevelThresholds{
		Safe: 10, Low: 50, Medium: 70, High: 90, Critical: 100,
	})

	level := calc.Classify(60)
	if level != core.RiskMedium {
		t.Errorf("expected medium with custom thresholds (60 > 50, 60 <= 70), got %s", level)
	}

	level = calc.Classify(75)
	if level != core.RiskHigh {
		t.Errorf("expected high with custom thresholds (75 > 70, 75 <= 90), got %s", level)
	}
}

func TestListRiskFactors(t *testing.T) {
	repo := riskscoring.NewInMemoryRepository()
	calc := riskscoring.NewCalculator(repo)

	repo.UpsertFactor(riskscoring.RiskFactor{Name: "factor_a", Weight: 1.0, CurrentValue: 50})
	factors := calc.ListFactors(context.Background())
	if len(factors) != 1 {
		t.Errorf("expected 1 factor, got %d", len(factors))
	}
}
