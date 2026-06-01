package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/decisionlog"
)

func TestLogDecision(t *testing.T) {
	repo := decisionlog.NewInMemoryRepository()
	logger := decisionlog.NewLogger(repo)

	log := &decisionlog.DecisionLog{
		EventID:       "evt-1",
		EventType:     "login",
		UserID:        "user1",
		Decision:      decisionlog.DecisionAllow,
		RiskScore:     15.0,
		TriggeredRules: []string{"rule-1"},
	}

	if err := logger.LogDecision(context.Background(), log); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGetDecisionHistory(t *testing.T) {
	repo := decisionlog.NewInMemoryRepository()
	logger := decisionlog.NewLogger(repo)

	logger.LogDecision(context.Background(), &decisionlog.DecisionLog{
		EventID: "evt-1", EventType: "login", UserID: "u1", Decision: decisionlog.DecisionAllow,
	})
	logger.LogDecision(context.Background(), &decisionlog.DecisionLog{
		EventID: "evt-2", EventType: "payment", UserID: "u2", Decision: decisionlog.DecisionBlock,
	})

	logs, err := logger.GetDecisionHistory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}

func TestGetDecisionStats(t *testing.T) {
	repo := decisionlog.NewInMemoryRepository()
	logger := decisionlog.NewLogger(repo)

	logger.LogDecision(context.Background(), &decisionlog.DecisionLog{
		EventID: "e1", EventType: "login", UserID: "u1", Decision: decisionlog.DecisionAllow, RiskScore: 10,
	})
	logger.LogDecision(context.Background(), &decisionlog.DecisionLog{
		EventID: "e2", EventType: "payment", UserID: "u2", Decision: decisionlog.DecisionBlock, RiskScore: 80,
	})
	logger.LogDecision(context.Background(), &decisionlog.DecisionLog{
		EventID: "e3", EventType: "order", UserID: "u3", Decision: decisionlog.DecisionReview, RiskScore: 50,
	})

	stats, err := logger.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalDecisions != 3 {
		t.Errorf("expected 3 total, got %d", stats.TotalDecisions)
	}
	if stats.AllowCount != 1 {
		t.Errorf("expected 1 allow, got %d", stats.AllowCount)
	}
	if stats.BlockCount != 1 {
		t.Errorf("expected 1 block, got %d", stats.BlockCount)
	}
	if stats.ReviewCount != 1 {
		t.Errorf("expected 1 review, got %d", stats.ReviewCount)
	}
	if stats.AvgRiskScore <= 0 {
		t.Errorf("expected positive avg risk score, got %f", stats.AvgRiskScore)
	}
}

func TestGetDecisionByID(t *testing.T) {
	repo := decisionlog.NewInMemoryRepository()
	logger := decisionlog.NewLogger(repo)

	log := &decisionlog.DecisionLog{
		EventID: "evt-99", EventType: "register", UserID: "u99", Decision: decisionlog.DecisionAllow,
	}
	logger.LogDecision(context.Background(), log)

	fetched, err := logger.GetDecisionByID(context.Background(), log.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.EventID != "evt-99" {
		t.Errorf("expected evt-99, got %s", fetched.EventID)
	}
}

func TestGetNonexistentDecision(t *testing.T) {
	repo := decisionlog.NewInMemoryRepository()
	logger := decisionlog.NewLogger(repo)

	_, err := logger.GetDecisionByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent decision")
	}
}
