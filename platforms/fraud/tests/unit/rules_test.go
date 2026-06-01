package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
)

func TestCreateRule(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{
		Name: "test_rule", Description: "Test rule",
		Condition: "amount>1000", Severity: 5, Weight: 0.5, IsActive: true,
	}

	if err := svc.CreateRule(context.Background(), rule); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGetRule(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{Name: "get_test", Condition: "new_device_login", Severity: 3, Weight: 0.4}
	svc.CreateRule(context.Background(), rule)

	fetched, err := svc.GetRule(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "get_test" {
		t.Errorf("expected get_test, got %s", fetched.Name)
	}
}

func TestUpdateRule(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{Name: "update_test", Condition: "amount_anomaly", Severity: 4, Weight: 0.6}
	svc.CreateRule(context.Background(), rule)

	rule.Severity = 9
	rule.Weight = 1.0
	if err := svc.UpdateRule(context.Background(), rule); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fetched, _ := svc.GetRule(context.Background(), rule.ID)
	if fetched.Severity != 9 {
		t.Errorf("expected severity 9, got %d", fetched.Severity)
	}
	if fetched.Weight != 1.0 {
		t.Errorf("expected weight 1.0, got %f", fetched.Weight)
	}
}

func TestListActiveRules(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rulesList := []*rules.RuleDefinition{
		{Name: "active1", Condition: "a", Severity: 1, Weight: 0.1, IsActive: true},
		{Name: "inactive1", Condition: "b", Severity: 1, Weight: 0.1, IsActive: false},
		{Name: "active2", Condition: "c", Severity: 1, Weight: 0.1, IsActive: true},
	}
	for _, r := range rulesList {
		svc.CreateRule(context.Background(), r)
	}

	active := svc.ListActive(context.Background())
	if len(active) != 2 {
		t.Errorf("expected 2 active rules, got %d", len(active))
	}
}

func TestToggleRule(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{Name: "toggle_test", Condition: "a", Severity: 1, Weight: 0.1, IsActive: true}
	svc.CreateRule(context.Background(), rule)

	toggled, err := svc.ToggleRule(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if toggled.IsActive {
		t.Error("expected rule to be inactive after toggle")
	}

	toggled2, _ := svc.ToggleRule(context.Background(), rule.ID)
	if !toggled2.IsActive {
		t.Error("expected rule to be active after second toggle")
	}
}

func TestEvaluateRuleNewDeviceLogin(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{
		Name: "new_device_login", Condition: "new_device_login",
		Severity: 5, Weight: 0.8, IsActive: true,
	}
	svc.CreateRule(context.Background(), rule)

	event := &core.FraudEvent{
		Type: core.EventLogin, UserID: "user1",
		DeviceID: "new-device", Timestamp: time.Now(),
	}

	eval := svc.EvaluateRule(context.Background(), rule, event)
	if !eval.Triggered {
		t.Error("expected rule to trigger for new device login")
	}
	if eval.Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestEvaluateRuleInactive(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{
		Name: "inactive_rule", Condition: "new_device_login",
		Severity: 5, Weight: 0.8, IsActive: false,
	}
	svc.CreateRule(context.Background(), rule)

	event := &core.FraudEvent{Type: core.EventLogin, Timestamp: time.Now()}
	eval := svc.EvaluateRule(context.Background(), rule, event)
	if eval.Triggered {
		t.Error("expected inactive rule not to trigger")
	}
}

func TestEvaluateRuleAmountThreshold(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	rule := &rules.RuleDefinition{
		Name: "amount_test", Condition: "amount>10000",
		Severity: 7, Weight: 0.9, IsActive: true,
	}
	svc.CreateRule(context.Background(), rule)

	tests := []struct {
		amount    float64
		triggered bool
	}{
		{5000, false},
		{15000, true},
		{100, false},
		{10001, true},
	}

	for _, tt := range tests {
		event := &core.FraudEvent{Type: core.EventOrder, Amount: tt.amount, Timestamp: time.Now()}
		eval := svc.EvaluateRule(context.Background(), rule, event)
		if eval.Triggered != tt.triggered {
			t.Errorf("amount %.0f: expected triggered=%v, got %v", tt.amount, tt.triggered, eval.Triggered)
		}
	}
}

func TestListRules(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	svc.CreateRule(context.Background(), &rules.RuleDefinition{Name: "r1", Condition: "a", Severity: 1, Weight: 0.1})
	svc.CreateRule(context.Background(), &rules.RuleDefinition{Name: "r2", Condition: "b", Severity: 1, Weight: 0.1})

	list, err := svc.ListRules(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 rules, got %d", len(list))
	}
}

func TestGetNonexistentRule(t *testing.T) {
	repo := rules.NewInMemoryRepository()
	svc := rules.NewService(repo)

	_, err := svc.GetRule(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent rule")
	}
}
