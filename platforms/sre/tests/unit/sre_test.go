package unit

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/sre/internal/alerting"
	"github.com/tikiclone/tiki/platforms/sre/internal/deployment"
	"github.com/tikiclone/tiki/platforms/sre/internal/healthcheck"
	"github.com/tikiclone/tiki/platforms/sre/internal/incident"
	"github.com/tikiclone/tiki/platforms/sre/internal/runbook"
	"github.com/tikiclone/tiki/platforms/sre/internal/slo"
)

// Incident lifecycle tests

func TestIncidentCreateAndGet(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	inc, err := svc.Create("test outage", incident.SeverityCritical, "api-gateway", "us-east-1", "service is down", "alice")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if inc.ID == "" {
		t.Error("expected non-empty ID")
	}
	if inc.Status != incident.StatusDetected {
		t.Errorf("expected status 'detected', got %s", inc.Status)
	}
	if inc.Severity != incident.SeverityCritical {
		t.Errorf("expected severity 'critical', got %s", inc.Severity)
	}

	got, err := svc.List(incident.Filter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 incident, got %d", len(got))
	}
}

func TestIncidentInvalidSeverity(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	_, err := svc.Create("test", incident.Severity("invalid"), "svc", "reg", "desc", "")
	if err != incident.ErrInvalidSeverity {
		t.Errorf("expected ErrInvalidSeverity, got %v", err)
	}
}

func TestIncidentAcknowledgeAndResolve(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	inc, _ := svc.Create("test", incident.SeverityMajor, "payment", "eu-west-1", "payment failures", "bob")
	inc, err := svc.Acknowledge(inc.ID)
	if err != nil {
		t.Fatalf("Acknowledge failed: %v", err)
	}
	if inc.Status != incident.StatusTriaging {
		t.Errorf("expected 'triaging', got %s", inc.Status)
	}

	inc, err = svc.Resolve(inc.ID)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if inc.Status != incident.StatusResolved {
		t.Errorf("expected 'resolved', got %s", inc.Status)
	}
	if inc.ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
}

func TestIncidentAcknowledgeResolvedFails(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	inc, _ := svc.Create("test", incident.SeverityMinor, "svc", "reg", "desc", "")
	svc.Resolve(inc.ID)
	_, err := svc.Acknowledge(inc.ID)
	if err != incident.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestIncidentUpdate(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	inc, _ := svc.Create("test", incident.SeverityMinor, "svc", "reg", "desc", "")
	updated, err := svc.Update(inc.ID, map[string]interface{}{"title": "updated title", "assignee": "charlie"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Title != "updated title" {
		t.Errorf("expected 'updated title', got '%s'", updated.Title)
	}
	if updated.Assignee != "charlie" {
		t.Errorf("expected 'charlie', got '%s'", updated.Assignee)
	}
}

func TestIncidentFilterBySeverity(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	svc.Create("critical issue", incident.SeverityCritical, "svc1", "reg", "", "")
	svc.Create("major issue", incident.SeverityMajor, "svc2", "reg", "", "")
	svc.Create("minor issue", incident.SeverityMinor, "svc3", "reg", "", "")

	result, _ := svc.List(incident.Filter{Severity: incident.SeverityCritical})
	if len(result) != 1 {
		t.Errorf("expected 1 critical incident, got %d", len(result))
	}

	result, _ = svc.List(incident.Filter{Service: "svc2"})
	if len(result) != 1 {
		t.Errorf("expected 1 incident for svc2, got %d", len(result))
	}
}

// Alert rule tests

func TestCreateAlertRule(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{
		Name:       "high-cpu",
		MetricName: "cpu_usage",
		Operator:   ">",
		Threshold:  90,
	}
	err := svc.CreateRule(rule)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	rules, _ := svc.ListRules()
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
}

func TestCreateDuplicateRuleFails(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	svc.CreateRule(&alerting.Rule{Name: "high-cpu", MetricName: "cpu", Operator: ">", Threshold: 90})
	err := svc.CreateRule(&alerting.Rule{Name: "high-cpu", MetricName: "mem", Operator: ">", Threshold: 80})
	if err != alerting.ErrRuleExists {
		t.Errorf("expected ErrRuleExists, got %v", err)
	}
}

func TestEvaluateRuleFiresAlert(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{Name: "high-cpu", MetricName: "cpu_usage", Operator: ">", Threshold: 90}
	svc.CreateRule(rule)

	metrics := []alerting.MetricValue{
		{Name: "cpu_usage", Value: 95},
	}
	alerts, err := svc.Evaluate([]*alerting.Rule{rule}, metrics)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Status != alerting.AlertFiring {
		t.Errorf("expected 'firing', got %s", alerts[0].Status)
	}
}

func TestEvaluateRuleNoAlertBelowThreshold(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{Name: "high-cpu", MetricName: "cpu_usage", Operator: ">", Threshold: 90}
	svc.CreateRule(rule)

	metrics := []alerting.MetricValue{
		{Name: "cpu_usage", Value: 50},
	}
	alerts, err := svc.Evaluate([]*alerting.Rule{rule}, metrics)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestEvaluateRuleLessThanOperator(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{Name: "low-memory", MetricName: "mem_free", Operator: "<", Threshold: 256}
	svc.CreateRule(rule)

	metrics := []alerting.MetricValue{
		{Name: "mem_free", Value: 128},
	}
	alerts, _ := svc.Evaluate([]*alerting.Rule{rule}, metrics)
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert for low memory, got %d", len(alerts))
	}
}

func TestEvaluateRuleCooldown(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{Name: "high-cpu", MetricName: "cpu_usage", Operator: ">", Threshold: 90, CooldownSeconds: 3600}
	svc.CreateRule(rule)

	metrics := []alerting.MetricValue{
		{Name: "cpu_usage", Value: 95},
	}

	alerts1, _ := svc.Evaluate([]*alerting.Rule{rule}, metrics)
	alerts2, _ := svc.Evaluate([]*alerting.Rule{rule}, metrics)

	if len(alerts1) != 1 {
		t.Errorf("expected 1 alert on first eval, got %d", len(alerts1))
	}
	if len(alerts2) != 0 {
		t.Errorf("expected 0 alerts during cooldown, got %d", len(alerts2))
	}
}

// Health check tests

func TestCreateHealthCheck(t *testing.T) {
	repo := healthcheck.NewInMemoryRepository()
	svc := healthcheck.NewService(repo)

	hc, err := svc.CreateCheck("api-health", "https://api.example.com/health", 30, 5, "GET", 200)
	if err != nil {
		t.Fatalf("CreateCheck failed: %v", err)
	}
	if hc.ID == "" {
		t.Error("expected non-empty ID")
	}

	checks, _ := svc.ListChecks()
	if len(checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(checks))
	}
}

func TestRunChecksProducesResults(t *testing.T) {
	repo := healthcheck.NewInMemoryRepository()
	svc := healthcheck.NewService(repo)

	svc.CreateCheck("api-health", "https://api.example.com/health", 30, 5, "GET", 200)
	svc.CreateCheck("db-health", "https://db.example.com/health", 30, 5, "GET", 200)

	results, err := svc.RunChecks()
	if err != nil {
		t.Fatalf("RunChecks failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.CheckName == "" {
			t.Error("expected check name in result")
		}
		if r.Status != healthcheck.StatusPass && r.Status != healthcheck.StatusFail && r.Status != healthcheck.StatusDegraded {
			t.Errorf("unexpected status: %s", r.Status)
		}
	}
}

func TestGetResultsReturnsResults(t *testing.T) {
	repo := healthcheck.NewInMemoryRepository()
	svc := healthcheck.NewService(repo)

	svc.CreateCheck("test-check", "https://example.com", 30, 5, "GET", 200)
	svc.RunChecks()

	results, err := svc.GetResults()
	if err != nil {
		t.Fatalf("GetResults failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least 1 result")
	}
	if results[0].CheckedAt.IsZero() {
		t.Error("expected checked_at to be set")
	}
}

func TestMultipleRunChecksAppendResults(t *testing.T) {
	repo := healthcheck.NewInMemoryRepository()
	svc := healthcheck.NewService(repo)

	svc.CreateCheck("test-check", "https://example.com", 30, 5, "GET", 200)
	svc.RunChecks()
	svc.RunChecks()

	results, _ := svc.GetResults()
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// SLO/SLI tests

func TestCreateSLO(t *testing.T) {
	repo := slo.NewInMemoryRepository()
	svc := slo.NewService(repo)

	sl, err := svc.CreateSLO("api-availability", "api-gateway", "http_requests_total", 99.9, slo.Window28d)
	if err != nil {
		t.Fatalf("CreateSLO failed: %v", err)
	}
	if sl.ID == "" {
		t.Error("expected non-empty ID")
	}
	if sl.CurrentValue != 100.0 {
		t.Errorf("expected initial value 100, got %f", sl.CurrentValue)
	}
}

func TestCalculateSLI(t *testing.T) {
	repo := slo.NewInMemoryRepository()
	svc := slo.NewService(repo)

	sli := svc.CalculateSLI("http_requests_total", 950, 1000)
	if sli.Ratio != 95.0 {
		t.Errorf("expected ratio 95, got %f", sli.Ratio)
	}

	sli = svc.CalculateSLI("http_requests_total", 0, 0)
	if sli.Ratio != 0 {
		t.Errorf("expected ratio 0 for zero total, got %f", sli.Ratio)
	}
}

func TestEvaluateSLOUpdatesValue(t *testing.T) {
	repo := slo.NewInMemoryRepository()
	svc := slo.NewService(repo)

	sl, _ := svc.CreateSLO("api-availability", "api", "requests", 99.9, slo.Window28d)
	sli := svc.CalculateSLI("requests", 999, 1000)

	updated, err := svc.EvaluateSLO(sl.ID, sli)
	if err != nil {
		t.Fatalf("EvaluateSLO failed: %v", err)
	}
	if updated.CurrentValue != 99.9 {
		t.Errorf("expected current_value 99.9, got %f", updated.CurrentValue)
	}
}

func TestGetSLOReport(t *testing.T) {
	repo := slo.NewInMemoryRepository()
	svc := slo.NewService(repo)

	sl, _ := svc.CreateSLO("api-availability", "api", "requests", 99.9, slo.Window28d)

	report, err := svc.GetSLOReport(sl.ID)
	if err != nil {
		t.Fatalf("GetSLOReport failed: %v", err)
	}
	if report.SLO.Name != "api-availability" {
		t.Errorf("expected 'api-availability', got '%s'", report.SLO.Name)
	}
	if !report.Compliant {
		t.Error("expected compliant to be true for initial 100%")
	}
}

func TestSLOList(t *testing.T) {
	repo := slo.NewInMemoryRepository()
	svc := slo.NewService(repo)

	svc.CreateSLO("slo1", "svc1", "metric1", 99.9, slo.Window28d)
	svc.CreateSLO("slo2", "svc2", "metric2", 99.99, slo.Window7d)

	slos, _ := svc.ListSLOs()
	if len(slos) != 2 {
		t.Errorf("expected 2 SLOs, got %d", len(slos))
	}
}

// Deployment strategy tests

func TestCreateDeployment(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, err := svc.Create("api-gateway", "v2.0.0", deployment.StrategyRolling)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if d.ID == "" {
		t.Error("expected non-empty ID")
	}
	if d.Status != deployment.DeployPending {
		t.Errorf("expected 'pending', got %s", d.Status)
	}
}

func TestDeploymentRollingStrategy(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("api-gateway", "v2.0.0", deployment.StrategyRolling)

	d, _ = svc.Approve(d.ID)
	if d.ProgressPercentage != 25 {
		t.Errorf("expected 25%% after first approve, got %d%%", d.ProgressPercentage)
	}

	d, _ = svc.Approve(d.ID)
	if d.ProgressPercentage != 50 {
		t.Errorf("expected 50%% after second approve, got %d%%", d.ProgressPercentage)
	}

	d, _ = svc.Approve(d.ID)
	if d.ProgressPercentage != 75 {
		t.Errorf("expected 75%% after third approve, got %d%%", d.ProgressPercentage)
	}

	d, _ = svc.Approve(d.ID)
	if d.ProgressPercentage != 100 {
		t.Errorf("expected 100%% after fourth approve, got %d%%", d.ProgressPercentage)
	}
	if d.Status != deployment.DeploySucceeded {
		t.Errorf("expected 'succeeded', got %s", d.Status)
	}
}

func TestDeploymentCanaryStrategy(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("api-gateway", "v2.0.0", deployment.StrategyCanary)

	expected := []int{10, 25, 50, 100}
	for i, exp := range expected {
		d, _ = svc.Approve(d.ID)
		if d.ProgressPercentage != exp {
			t.Errorf("step %d: expected %d%%, got %d%%", i+1, exp, d.ProgressPercentage)
		}
	}
	if d.Status != deployment.DeploySucceeded {
		t.Errorf("expected 'succeeded', got %s", d.Status)
	}
}

func TestDeploymentBlueGreenStrategy(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("api-gateway", "v2.0.0", deployment.StrategyBlueGreen)

	d, _ = svc.Approve(d.ID)
	if d.ProgressPercentage != 100 {
		t.Errorf("expected 100%% for blue/green, got %d%%", d.ProgressPercentage)
	}
	if d.Status != deployment.DeploySucceeded {
		t.Errorf("expected 'succeeded', got %s", d.Status)
	}
}

func TestDeploymentRollback(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("api-gateway", "v2.0.0", deployment.StrategyRolling)
	d, _ = svc.Approve(d.ID)

	d, err := svc.Rollback(d.ID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if d.Status != deployment.DeployRolledBack {
		t.Errorf("expected 'rolled_back', got %s", d.Status)
	}
	if d.ProgressPercentage != 0 {
		t.Errorf("expected 0%% after rollback, got %d%%", d.ProgressPercentage)
	}
}

func TestDeploymentDoubleRollbackFails(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("api-gateway", "v2.0.0", deployment.StrategyRolling)
	svc.Rollback(d.ID)
	_, err := svc.Rollback(d.ID)
	if err != deployment.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestDeploymentList(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	svc.Create("svc1", "v1", deployment.StrategyRolling)
	svc.Create("svc2", "v2", deployment.StrategyCanary)

	deployments, _ := svc.List()
	if len(deployments) != 2 {
		t.Errorf("expected 2 deployments, got %d", len(deployments))
	}
}

func TestDeploymentGetStatus(t *testing.T) {
	repo := deployment.NewInMemoryRepository()
	svc := deployment.NewService(repo)

	d, _ := svc.Create("svc1", "v1", deployment.StrategyRolling)
	got, err := svc.GetStatus(d.ID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if got.ID != d.ID {
		t.Errorf("expected id %s, got %s", d.ID, got.ID)
	}
}

// Runbook tests

func TestCreateRunbook(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	steps := []runbook.Step{
		{Title: "Check logs", Command: "kubectl logs -n production pod/api-0", ExpectedResult: "no errors"},
		{Title: "Restart service", Command: "kubectl rollout restart deployment/api", ExpectedResult: "restarted"},
	}
	rb, err := svc.Create("API Outage Response", "api-gateway", "outage", steps)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if rb.ID == "" {
		t.Error("expected non-empty ID")
	}
	if len(rb.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(rb.Steps))
	}
}

func TestGetRunbook(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	rb, _ := svc.Create("test", "svc", "outage", []runbook.Step{{Title: "Check", Command: "cmd", ExpectedResult: "ok"}})
	got, err := svc.Get(rb.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Title != "test" {
		t.Errorf("expected 'test', got '%s'", got.Title)
	}
}

func TestUpdateRunbook(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	rb, _ := svc.Create("test", "svc", "outage", []runbook.Step{{Title: "Check", Command: "cmd", ExpectedResult: "ok"}})
	updated, err := svc.Update(rb.ID, "updated title", []runbook.Step{{Title: "New step", Command: "new", ExpectedResult: "done"}})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Title != "updated title" {
		t.Errorf("expected 'updated title', got '%s'", updated.Title)
	}
	if len(updated.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(updated.Steps))
	}
}

func TestDeleteRunbook(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	rb, _ := svc.Create("test", "svc", "outage", []runbook.Step{{Title: "Check", Command: "cmd", ExpectedResult: "ok"}})
	err := svc.Delete(rb.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = svc.Get(rb.ID)
	if err != runbook.ErrRunbookNotFound {
		t.Errorf("expected ErrRunbookNotFound, got %v", err)
	}
}

func TestListRunbooksFilterByService(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	svc.Create("rb1", "api", "outage", []runbook.Step{{Title: "Step", Command: "cmd", ExpectedResult: "ok"}})
	svc.Create("rb2", "db", "outage", []runbook.Step{{Title: "Step", Command: "cmd", ExpectedResult: "ok"}})
	svc.Create("rb3", "api", "latency", []runbook.Step{{Title: "Step", Command: "cmd", ExpectedResult: "ok"}})

	result, _ := svc.List(runbook.Filter{Service: "api"})
	if len(result) != 2 {
		t.Errorf("expected 2 runbooks for api, got %d", len(result))
	}

	result, _ = svc.List(runbook.Filter{IncidentType: "latency"})
	if len(result) != 1 {
		t.Errorf("expected 1 runbook for latency, got %d", len(result))
	}
}

func TestListRunbooksNoFilter(t *testing.T) {
	repo := runbook.NewInMemoryRepository()
	svc := runbook.NewService(repo)

	svc.Create("rb1", "api", "outage", []runbook.Step{{Title: "Step", Command: "cmd", ExpectedResult: "ok"}})
	svc.Create("rb2", "db", "outage", []runbook.Step{{Title: "Step", Command: "cmd", ExpectedResult: "ok"}})

	result, _ := svc.List(runbook.Filter{})
	if len(result) != 2 {
		t.Errorf("expected 2 runbooks, got %d", len(result))
	}
}

// Incident timing test

func TestIncidentResolvedAtTime(t *testing.T) {
	repo := incident.NewInMemoryRepository()
	svc := incident.NewService(repo)

	inc, _ := svc.Create("test", incident.SeverityCritical, "svc", "reg", "", "")
	time.Sleep(10 * time.Millisecond)
	svc.Resolve(inc.ID)

	list, _ := svc.List(incident.Filter{})
	if list[0].ResolvedAt == nil {
		t.Fatal("expected resolved_at to be set")
	}
	if !list[0].ResolvedAt.After(list[0].DetectedAt) {
		t.Error("expected resolved_at after detected_at")
	}
}

// Alert list test

func TestListAlerts(t *testing.T) {
	repo := alerting.NewInMemoryRepository()
	svc := alerting.NewService(repo)

	rule := &alerting.Rule{Name: "cpu", MetricName: "cpu_usage", Operator: ">", Threshold: 90}
	svc.CreateRule(rule)
	svc.Evaluate([]*alerting.Rule{rule}, []alerting.MetricValue{{Name: "cpu_usage", Value: 95}})

	alerts, _ := svc.ListAlerts()
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
}
