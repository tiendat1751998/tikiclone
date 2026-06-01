package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/ruleengine"
)

func TestCreateRule(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "test_rule",
		ConditionExpr: "amount>1000",
		Priority:      1,
		Weight:        0.5,
		ScoreDelta:    10,
		IsActive:      true,
	}

	if err := eng.CreateRule(context.Background(), rule); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGetRule(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{Name: "get_test", ConditionExpr: "new_device_login", ScoreDelta: 5, IsActive: true}
	eng.CreateRule(context.Background(), rule)

	fetched, err := eng.GetRule(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "get_test" {
		t.Errorf("expected get_test, got %s", fetched.Name)
	}
}

func TestListRules(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	eng.CreateRule(context.Background(), &ruleengine.Rule{Name: "r1", ConditionExpr: "a", ScoreDelta: 1, IsActive: true})
	eng.CreateRule(context.Background(), &ruleengine.Rule{Name: "r2", ConditionExpr: "b", ScoreDelta: 2, IsActive: true})

	list, err := eng.ListRules(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 rules, got %d", len(list))
	}
}

func TestEvaluateRuleNewDeviceLogin(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "new_device_login",
		ConditionExpr: "new_device_login",
		ScoreDelta:    15,
		IsActive:      true,
	}
	eng.CreateRule(context.Background(), rule)

	ev := &core.Event{
		Type:     core.EventLogin,
		UserID:   "user1",
		DeviceID: "new-device",
	}

	eval := eng.EvaluateRule(context.Background(), rule, ev)
	if !eval.Triggered {
		t.Error("expected rule to trigger for new device login")
	}
	if eval.Score != 15 {
		t.Errorf("expected score 15, got %f", eval.Score)
	}
}

func TestEvaluateRuleInactive(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "inactive_rule",
		ConditionExpr: "new_device_login",
		ScoreDelta:    10,
		IsActive:      false,
	}
	eng.CreateRule(context.Background(), rule)

	ev := &core.Event{Type: core.EventLogin, UserID: "user1"}
	eval := eng.EvaluateRule(context.Background(), rule, ev)
	if eval.Triggered {
		t.Error("expected inactive rule not to trigger")
	}
}

func TestEvaluateRuleAmountThreshold(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "amount_test",
		ConditionExpr: "amount>10000",
		ScoreDelta:    20,
		IsActive:      true,
	}
	eng.CreateRule(context.Background(), rule)

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
		ev := &core.Event{Type: core.EventOrder, Amount: tt.amount}
		eval := eng.EvaluateRule(context.Background(), rule, ev)
		if eval.Triggered != tt.triggered {
			t.Errorf("amount %.0f: expected triggered=%v, got %v", tt.amount, tt.triggered, eval.Triggered)
		}
	}
}

func TestEvaluateEventAllRules(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	eng.CreateRule(context.Background(), &ruleengine.Rule{Name: "r1", ConditionExpr: "high_amount", ScoreDelta: 10, IsActive: true})
	eng.CreateRule(context.Background(), &ruleengine.Rule{Name: "r2", ConditionExpr: "new_device_login", ScoreDelta: 15, IsActive: true})

	ev := &core.Event{
		Type:     core.EventLogin,
		UserID:   "user1",
		DeviceID: "new-device",
		Amount:   500,
	}

	results, err := eng.EvaluateEvent(context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var triggered int
	for _, r := range results {
		if r.Triggered {
			triggered++
		}
	}
	if triggered != 1 {
		t.Errorf("expected 1 triggered rule, got %d", triggered)
	}
}

func TestGetNonexistentRule(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	_, err := eng.GetRule(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent rule")
	}
}

func TestCreateRuleSet(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rs := &ruleengine.RuleSet{
		Name:     "test_ruleset",
		Strategy: ruleengine.StrategyMatchAll,
		Rules: []ruleengine.Rule{
			{Name: "r1", ConditionExpr: "high_amount", IsActive: true, ScoreDelta: 10},
		},
	}

	if err := eng.CreateRuleSet(context.Background(), rs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rs.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestEvaluateRuleSetMatchAllPass(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rs := &ruleengine.RuleSet{
		Name:     "match_all_test",
		Strategy: ruleengine.StrategyMatchAll,
		Rules: []ruleengine.Rule{
			{Name: "high_amount", ConditionExpr: "high_amount", IsActive: true, ScoreDelta: 10},
		},
	}
	eng.CreateRuleSet(context.Background(), rs)

	ev := &core.Event{Type: core.EventOrder, Amount: 20000}
	result, err := eng.EvaluateRuleSet(context.Background(), rs.ID, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected ruleset to pass")
	}
}

func TestEvaluateRuleSetMatchAnyPass(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rs := &ruleengine.RuleSet{
		Name:     "match_any_test",
		Strategy: ruleengine.StrategyMatchAny,
		Rules: []ruleengine.Rule{
			{Name: "high_amount", ConditionExpr: "high_amount", IsActive: true, ScoreDelta: 10},
			{Name: "never_trigger", ConditionExpr: "new_device_login", IsActive: true, ScoreDelta: 5},
		},
	}
	eng.CreateRuleSet(context.Background(), rs)

	ev := &core.Event{Type: core.EventOrder, Amount: 20000}
	result, err := eng.EvaluateRuleSet(context.Background(), rs.ID, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Error("expected ruleset to pass with match_any")
	}
}

func TestEvaluateRuleSetWeightedSum(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rs := &ruleengine.RuleSet{
		Name:     "weighted_test",
		Strategy: ruleengine.StrategyWeightedSum,
		Rules: []ruleengine.Rule{
			{Name: "high_amount", ConditionExpr: "high_amount", IsActive: true, ScoreDelta: 10},
		},
	}
	eng.CreateRuleSet(context.Background(), rs)

	ev := &core.Event{Type: core.EventOrder, Amount: 20000}
	result, err := eng.EvaluateRuleSet(context.Background(), rs.ID, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalScore <= 0 {
		t.Errorf("expected positive total score, got %f", result.TotalScore)
	}
}

func TestEvaluateRuleSetMatchAllFails(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rs := &ruleengine.RuleSet{
		Name:     "match_all_fail",
		Strategy: ruleengine.StrategyMatchAll,
		Rules: []ruleengine.Rule{
			{Name: "high_amount", ConditionExpr: "high_amount", IsActive: true, ScoreDelta: 10},
			{Name: "login_check", ConditionExpr: "new_device_login", IsActive: true, ScoreDelta: 5},
		},
	}
	eng.CreateRuleSet(context.Background(), rs)

	ev := &core.Event{Type: core.EventOrder, Amount: 20000, DeviceID: ""}
	result, err := eng.EvaluateRuleSet(context.Background(), rs.ID, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Error("expected ruleset to fail match_all when one rule doesn't trigger")
	}
}

func TestEvaluateRuleForeignIP(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "foreign_ip",
		ConditionExpr: "foreign_ip",
		ScoreDelta:    25,
		IsActive:      true,
	}
	eng.CreateRule(context.Background(), rule)

	ev := &core.Event{Type: core.EventLogin, IP: "203.0.113.1"}
	eval := eng.EvaluateRule(context.Background(), rule, ev)
	if !eval.Triggered {
		t.Error("expected foreign_ip to trigger")
	}
}

func TestEvaluateRulePaymentFraud(t *testing.T) {
	repo := ruleengine.NewInMemoryRuleRepository()
	rulesetRepo := ruleengine.NewInMemoryRuleSetRepository()
	eng := ruleengine.NewEngine(repo, rulesetRepo)

	rule := &ruleengine.Rule{
		Name:          "payment_fraud",
		ConditionExpr: "payment_fraud",
		ScoreDelta:    30,
		IsActive:      true,
	}
	eng.CreateRule(context.Background(), rule)

	ev := &core.Event{Type: core.EventPayment}
	eval := eng.EvaluateRule(context.Background(), rule, ev)
	if !eval.Triggered {
		t.Error("expected payment_fraud to trigger")
	}
}
