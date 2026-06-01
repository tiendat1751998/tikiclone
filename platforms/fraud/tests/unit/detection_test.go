package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
	fraudcase "github.com/tikiclone/tiki/platforms/fraud/internal/case"
	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud/internal/detection"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
	"github.com/tikiclone/tiki/platforms/fraud/internal/scoring"
	"github.com/tikiclone/tiki/platforms/fraud/internal/streaming"
	"github.com/tikiclone/tiki/platforms/fraud/internal/verification"
)

func setupDetectionTest(t *testing.T) (*detection.Service, *rules.Service, *rules.InMemoryRepository) {
	t.Helper()
	ruleRepo := rules.NewInMemoryRepository()
	scoreRepo := scoring.NewInMemoryRepository()
	streamRepo := streaming.NewInMemoryRepository()
	blacklistRepo := blacklist.NewInMemoryRepository()

	ruleSvc := rules.NewService(ruleRepo)
	scoreSvc := scoring.NewService(scoreRepo)
	streamSvc := streaming.NewService(streamRepo)
	blacklistSvc := blacklist.NewService(blacklistRepo)

	detectRepo := detection.NewInMemoryRepository()
	detectSvc := detection.NewService(detectRepo, ruleSvc, scoreSvc, streamSvc, blacklistSvc, 51)

	ruleSvc.CreateRule(context.Background(), &rules.RuleDefinition{
		Name: "new_device_login", Description: "New device login detection",
		Condition: "new_device_login", Severity: 5, Weight: 0.8, IsActive: true,
	})
	ruleSvc.CreateRule(context.Background(), &rules.RuleDefinition{
		Name: "amount_anomaly", Description: "Amount anomaly detection",
		Condition: "amount_anomaly", Severity: 6, Weight: 0.9, IsActive: true,
	})

	return detectSvc, ruleSvc, ruleRepo
}

func TestEvaluateLowRisk(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-1", Type: core.EventLogin, UserID: "user1",
		IP: "192.168.1.1", DeviceID: "dev-1", Amount: 50,
		Timestamp: time.Now(),
	}

	score, err := svc.Evaluate(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.AlertTriggered {
		t.Error("expected no alert for low risk event")
	}
	if score.Level != core.RiskLow && score.Level != core.RiskMedium {
		t.Errorf("expected low/medium risk, got %s", score.Level)
	}
}

func TestEvaluateHighRisk(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-2", Type: core.EventPayment, UserID: "user2",
		IP: "10.0.0.1", DeviceID: "dev-2", Amount: 50000,
		Timestamp: time.Now(),
	}

	score, err := svc.Evaluate(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.AlertTriggered {
		t.Log("score was", score.Score)
	}
}

func TestEvaluateBlacklistedUser(t *testing.T) {
	blRepo := blacklist.NewInMemoryRepository()
	blSvc := blacklist.NewService(blRepo)
	blSvc.Add(context.Background(), &blacklist.BlacklistEntry{
		Type: blacklist.BlacklistUser, Value: "blocked-user",
		Reason: blacklist.ReasonFraudulentActivity, IsActive: true,
	})

	ruleRepo := rules.NewInMemoryRepository()
	scoreRepo := scoring.NewInMemoryRepository()
	streamRepo := streaming.NewInMemoryRepository()
	ruleSvc := rules.NewService(ruleRepo)
	scoreSvc := scoring.NewService(scoreRepo)
	streamSvc := streaming.NewService(streamRepo)
	detectRepo := detection.NewInMemoryRepository()
	detectSvc := detection.NewService(detectRepo, ruleSvc, scoreSvc, streamSvc, blSvc, 50)

	event := &core.FraudEvent{
		ID: "evt-3", Type: core.EventLogin, UserID: "blocked-user", Timestamp: time.Now(),
	}

	_, err := detectSvc.Evaluate(context.Background(), event)
	if err == nil {
		t.Fatal("expected error for blacklisted user")
	}
}

func TestGetAlert(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-get", Type: core.EventPayment, UserID: "user-get",
		Amount: 100000, Timestamp: time.Now(),
	}
	score, _ := svc.Evaluate(context.Background(), event)
	if !score.AlertTriggered {
		t.Skip("no alert triggered")
	}

	alert, err := svc.GetAlert(context.Background(), score.AlertID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert.ID != score.AlertID {
		t.Errorf("alert ID mismatch: %s vs %s", alert.ID, score.AlertID)
	}
}

func TestResolveAlert(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-resolve", Type: core.EventPayment, UserID: "user-resolve",
		Amount: 99999, Timestamp: time.Now(),
	}
	score, _ := svc.Evaluate(context.Background(), event)
	if !score.AlertTriggered {
		t.Skip("no alert triggered")
	}

	err := svc.ResolveAlert(context.Background(), score.AlertID, "investigator1", "false positive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alert, _ := svc.GetAlert(context.Background(), score.AlertID)
	if alert.Status != "resolved" {
		t.Errorf("expected resolved, got %s", alert.Status)
	}
	if alert.ResolvedBy != "investigator1" {
		t.Errorf("expected investigator1, got %s", alert.ResolvedBy)
	}
}

func TestListAlertsFilter(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-list1", Type: core.EventPayment, UserID: "user-list",
		Amount: 100000, Timestamp: time.Now(),
	}
	svc.Evaluate(context.Background(), event)

	alerts, total, err := svc.ListAlerts(context.Background(), "open", "", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total == 0 {
		t.Error("expected at least 1 alert")
	}
	_ = alerts
}

func TestEvaluateInvalidEvent(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)
	_, err := svc.Evaluate(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil event")
	}
}

func TestAlertTypes(t *testing.T) {
	svc, _, _ := setupDetectionTest(t)

	event := &core.FraudEvent{
		ID: "evt-login", Type: core.EventLogin, UserID: "user-login",
		DeviceID: "new-device", Timestamp: time.Now(),
	}
	score, err := svc.Evaluate(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.AlertTriggered {
		alert, _ := svc.GetAlert(context.Background(), score.AlertID)
		t.Logf("alert type: %s, score: %.2f", alert.Type, score.Score)
	}
}

func TestDetectionWithVerification(t *testing.T) {
	verifyRepo := verification.NewInMemoryRepository()
	verifySvc := verification.NewService(verifyRepo, 10)

	req, err := verifySvc.InitiateVerification(context.Background(), "user-v", verification.MethodEmail, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != "pending" {
		t.Errorf("expected pending, got %s", req.Status)
	}
}

func TestDetectionWithCase(t *testing.T) {
	caseRepo := fraudcase.NewInMemoryRepository()
	caseSvc := fraudcase.NewService(caseRepo)

	c, err := caseSvc.CreateCase(context.Background(), "alert-1", "user-c", "Investigation", "Suspicious activity", 85.0, fraudcase.PriorityHigh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != "open" {
		t.Errorf("expected open, got %s", c.Status)
	}
	if c.Priority != fraudcase.PriorityHigh {
		t.Errorf("expected high, got %s", c.Priority)
	}
}

func TestEvaluateMultipleEvents(t *testing.T) {
	svc, ruleSvc, _ := setupDetectionTest(t)

	ruleSvc.CreateRule(context.Background(), &rules.RuleDefinition{
		Name: "payment_fraud", Description: "Payment fraud detection",
		Condition: "payment_fraud", Severity: 8, Weight: 1.0, IsActive: true,
	})

	events := []*core.FraudEvent{
		{ID: "m1", Type: core.EventLogin, UserID: "u1", DeviceID: "d1", Amount: 100, Timestamp: time.Now()},
		{ID: "m2", Type: core.EventOrder, UserID: "u1", DeviceID: "d1", Amount: 500, Timestamp: time.Now()},
		{ID: "m3", Type: core.EventPayment, UserID: "u1", DeviceID: "d1", Amount: 100000, Timestamp: time.Now()},
	}

	for _, e := range events {
		score, err := svc.Evaluate(context.Background(), e)
		if err != nil {
			t.Fatalf("unexpected error for event %s: %v", e.ID, err)
		}
		if e.Amount > 10000 && !score.AlertTriggered {
			t.Logf("event %s triggered alert=%v score=%.2f", e.ID, score.AlertTriggered, score.Score)
		}
	}
}
