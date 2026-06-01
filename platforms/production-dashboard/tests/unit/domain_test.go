package domain_test

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/domain"
)

func TestNewServiceHealth(t *testing.T) {
	h := domain.NewServiceHealth("auth-service", "http://auth:8080/health", "production")
	if h.ServiceName != "auth-service" {
		t.Errorf("expected auth-service, got %s", h.ServiceName)
	}
	if h.HealthURL != "http://auth:8080/health" {
		t.Errorf("expected health URL, got %s", h.HealthURL)
	}
	if h.Status != domain.ServiceStatusUnknown {
		t.Errorf("expected unknown status, got %s", h.Status)
	}
	if h.Environment != "production" {
		t.Errorf("expected production env, got %s", h.Environment)
	}
	if h.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestServiceHealth_MarkHealthy(t *testing.T) {
	h := domain.NewServiceHealth("test", "http://test/health", "staging")
	h.MarkHealthy(45, "1.2.3")
	if h.Status != domain.ServiceStatusHealthy {
		t.Errorf("expected healthy, got %s", h.Status)
	}
	if h.ResponseTime != 45 {
		t.Errorf("expected 45ms, got %d", h.ResponseTime)
	}
	if h.Version != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", h.Version)
	}
	if h.LastError != "" {
		t.Errorf("expected empty error, got %s", h.LastError)
	}
	if !h.IsHealthy() {
		t.Error("expected IsHealthy() to be true")
	}
}

func TestServiceHealth_MarkUnhealthy(t *testing.T) {
	h := domain.NewServiceHealth("test", "http://test/health", "staging")
	h.MarkUnhealthy("connection refused")
	if h.Status != domain.ServiceStatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", h.Status)
	}
	if h.LastError != "connection refused" {
		t.Errorf("expected error message, got %s", h.LastError)
	}
	if h.ResponseTime != 0 {
		t.Errorf("expected 0 response time, got %d", h.ResponseTime)
	}
}

func TestNewDeployment(t *testing.T) {
	d := domain.NewDeployment("auth", "2.0.0", "production", "admin@example.com", "ghcr.io/tiki/auth:2.0.0", 3, "rolling")
	if d.ServiceName != "auth" {
		t.Errorf("expected auth, got %s", d.ServiceName)
	}
	if d.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", d.Version)
	}
	if d.Status != domain.DeploymentStatusPending {
		t.Errorf("expected pending, got %s", d.Status)
	}
	if d.Replicas != 3 {
		t.Errorf("expected 3 replicas, got %d", d.Replicas)
	}
	if d.Strategy != "rolling" {
		t.Errorf("expected rolling strategy, got %s", d.Strategy)
	}
}

func TestDeployment_MarkSucceeded(t *testing.T) {
	d := domain.NewDeployment("auth", "2.0.0", "production", "admin@example.com", "ghcr.io/tiki/auth:2.0.0", 3, "rolling")
	d.MarkInProgress()
	if d.Status != domain.DeploymentStatusInProgress {
		t.Errorf("expected in_progress, got %s", d.Status)
	}
	d.MarkSucceeded(3)
	if d.Status != domain.DeploymentStatusSucceeded {
		t.Errorf("expected succeeded, got %s", d.Status)
	}
	if d.ReadyReplicas != 3 {
		t.Errorf("expected 3 ready replicas, got %d", d.ReadyReplicas)
	}
	if d.FinishedAt.IsZero() {
		t.Error("expected finished_at to be set")
	}
}

func TestDeployment_MarkRolledBack(t *testing.T) {
	d := domain.NewDeployment("auth", "2.0.0", "production", "admin@example.com", "ghcr.io/tiki/auth:2.0.0", 3, "rolling")
	d.MarkRolledBack()
	if d.Status != domain.DeploymentStatusRolledBack {
		t.Errorf("expected rolled_back, got %s", d.Status)
	}
}

func TestNewIncident(t *testing.T) {
	i := domain.NewIncident("Auth service down", "All auth endpoints returning 503", domain.IncidentSeverityCritical, "auth-service", "system")
	if i.Title != "Auth service down" {
		t.Errorf("expected title, got %s", i.Title)
	}
	if i.Severity != domain.IncidentSeverityCritical {
		t.Errorf("expected critical, got %s", i.Severity)
	}
	if i.Status != domain.IncidentStatusOpen {
		t.Errorf("expected open, got %s", i.Status)
	}
	if i.CreatedBy != "system" {
		t.Errorf("expected system, got %s", i.CreatedBy)
	}
}

func TestIncident_Acknowledge(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	err := i.Acknowledge("oncall@example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i.Status != domain.IncidentStatusAcknowledged {
		t.Errorf("expected acknowledged, got %s", i.Status)
	}
	if i.AssignedTo != "oncall@example.com" {
		t.Errorf("expected assigned to, got %s", i.AssignedTo)
	}
	if i.AcknowledgedAt.IsZero() {
		t.Error("expected acknowledged_at to be set")
	}
}

func TestIncident_Acknowledge_InvalidState(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	i.Acknowledge("oncall@example.com")
	err := i.Acknowledge("other@example.com")
	if err == nil {
		t.Error("expected error for acknowledging non-open incident")
	}
}

func TestIncident_Resolve(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	i.Acknowledge("oncall@example.com")
	i.StartInvestigation()
	i.Mitigate()
	err := i.Resolve("database connection pool exhausted", "Restarted connection pool")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i.Status != domain.IncidentStatusResolved {
		t.Errorf("expected resolved, got %s", i.Status)
	}
	if i.RootCause != "database connection pool exhausted" {
		t.Errorf("expected root cause, got %s", i.RootCause)
	}
	if i.ResolvedAt.IsZero() {
		t.Error("expected resolved_at to be set")
	}
}

func TestIncident_Close(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	i.Acknowledge("oncall@example.com")
	i.StartInvestigation()
	i.Mitigate()
	i.Resolve("root cause", "resolution")
	err := i.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i.Status != domain.IncidentStatusClosed {
		t.Errorf("expected closed, got %s", i.Status)
	}
}

func TestIncident_Close_NotResolved(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	err := i.Close()
	if err == nil {
		t.Error("expected error for closing unresolved incident")
	}
}

func TestIncident_Duration(t *testing.T) {
	i := domain.NewIncident("Test", "Desc", domain.IncidentSeverityHigh, "auth", "system")
	duration := i.Duration()
	if duration < 0 {
		t.Error("expected non-negative duration")
	}
}

func TestNewAlertRule(t *testing.T) {
	r := domain.NewAlertRule("High CPU", "CPU > 80%", "auth", "cpu_usage", domain.AlertConditionGT, 80.0, "5m", "high", "slack://alerts", "admin")
	if r.Name != "High CPU" {
		t.Errorf("expected High CPU, got %s", r.Name)
	}
	if r.Condition != domain.AlertConditionGT {
		t.Errorf("expected gt, got %s", r.Condition)
	}
	if r.Threshold != 80.0 {
		t.Errorf("expected 80.0, got %f", r.Threshold)
	}
	if !r.Enabled {
		t.Error("expected enabled by default")
	}
}

func TestAlertRule_Evaluate(t *testing.T) {
	r := domain.NewAlertRule("High CPU", "CPU > 80%", "auth", "cpu_usage", domain.AlertConditionGT, 80.0, "5m", "high", "slack://alerts", "admin")
	if !r.Evaluate(85.0) {
		t.Error("expected alert to fire for 85 > 80")
	}
	if r.Evaluate(75.0) {
		t.Error("expected no alert for 75 < 80")
	}
	if r.Evaluate(80.0) {
		t.Error("expected no alert for 80 = 80 (gt, not gte)")
	}
}

func TestAlertRule_Evaluate_LessThan(t *testing.T) {
	r := domain.NewAlertRule("Low Memory", "Memory < 10%", "auth", "memory_free_pct", domain.AlertConditionLT, 10.0, "5m", "critical", "pagerduty://oncall", "admin")
	if !r.Evaluate(5.0) {
		t.Error("expected alert to fire for 5 < 10")
	}
	if r.Evaluate(15.0) {
		t.Error("expected no alert for 15 > 10")
	}
}

func TestAlertRule_Evaluate_Disabled(t *testing.T) {
	r := domain.NewAlertRule("Test", "Desc", "auth", "metric", domain.AlertConditionGT, 80.0, "5m", "high", "", "admin")
	r.Enabled = false
	if r.Evaluate(100.0) {
		t.Error("disabled rule should not fire")
	}
}

func TestNewAuditLog(t *testing.T) {
	l := domain.NewAuditLog("admin@example.com", domain.ActionCreate, "incident", "INC-001", "Created incident", "10.0.0.1", "Mozilla/5.0")
	if l.Actor != "admin@example.com" {
		t.Errorf("expected actor, got %s", l.Actor)
	}
	if l.Action != domain.ActionCreate {
		t.Errorf("expected create action, got %s", l.Action)
	}
	if l.Resource != "incident" {
		t.Errorf("expected incident resource, got %s", l.Resource)
	}
	if l.IPAddress != "10.0.0.1" {
		t.Errorf("expected IP, got %s", l.IPAddress)
	}
}

func TestNewServiceDependency(t *testing.T) {
	d := domain.NewServiceDependency("order", "payment", domain.DependencyTypeSync, true)
	if d.ServiceName != "order" {
		t.Errorf("expected order, got %s", d.ServiceName)
	}
	if d.DependsOn != "payment" {
		t.Errorf("expected payment, got %s", d.DependsOn)
	}
	if d.DependencyType != domain.DependencyTypeSync {
		t.Errorf("expected sync, got %s", d.DependencyType)
	}
	if !d.Critical {
		t.Error("expected critical dependency")
	}
}

func TestNewCapacityMetric(t *testing.T) {
	m := domain.NewCapacityMetric("auth", domain.ResourceCPU, 45.5, 100.0, "percent")
	if m.ServiceName != "auth" {
		t.Errorf("expected auth, got %s", m.ServiceName)
	}
	if m.ResourceType != domain.ResourceCPU {
		t.Errorf("expected cpu, got %s", m.ResourceType)
	}
	if m.CurrentValue != 45.5 {
		t.Errorf("expected 45.5, got %f", m.CurrentValue)
	}
	if m.Unit != "percent" {
		t.Errorf("expected percent, got %s", m.Unit)
	}
}

func TestCapacityMetric_UtilizationPercent(t *testing.T) {
	m := domain.NewCapacityMetric("auth", domain.ResourceCPU, 45.0, 100.0, "percent")
	if m.UtilizationPercent() != 45.0 {
		t.Errorf("expected 45%%, got %f", m.UtilizationPercent())
	}

	m2 := domain.NewCapacityMetric("auth", domain.ResourceMemory, 0, 0, "bytes")
	if m2.UtilizationPercent() != 0 {
		t.Errorf("expected 0 for zero max, got %f", m2.UtilizationPercent())
	}
}

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{domain.ErrServiceNotFound, "dashboard: service_not_found"},
		{domain.ErrDeploymentNotFound, "dashboard: deployment_not_found"},
		{domain.ErrIncidentNotFound, "dashboard: incident_not_found"},
		{domain.ErrAlertRuleNotFound, "dashboard: alert_rule_not_found"},
		{domain.ErrInvalidIncidentState, "dashboard: invalid_incident_state"},
		{domain.ErrDuplicateService, "dashboard: duplicate_service"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.err.Error())
		}
	}
}

func TestIncident_FullLifecycle(t *testing.T) {
	i := domain.NewIncident("DB outage", "Primary DB unreachable", domain.IncidentSeverityCritical, "order,payment", "system")

	// Open -> Acknowledged
	if err := i.Acknowledge("sre@example.com"); err != nil {
		t.Fatalf("acknowledge: %v", err)
	}

	// Acknowledged -> Investigating
	if err := i.StartInvestigation(); err != nil {
		t.Fatalf("investigate: %v", err)
	}
	if i.Status != domain.IncidentStatusInvestigating {
		t.Errorf("expected investigating, got %s", i.Status)
	}

	// Investigating -> Mitigating
	if err := i.Mitigate(); err != nil {
		t.Fatalf("mitigate: %v", err)
	}
	if i.Status != domain.IncidentStatusMitigating {
		t.Errorf("expected mitigating, got %s", i.Status)
	}

	// Mitigating -> Resolved
	if err := i.Resolve("DB failover triggered", "Promoted replica to primary"); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if i.Status != domain.IncidentStatusResolved {
		t.Errorf("expected resolved, got %s", i.Status)
	}

	// Resolved -> Closed
	if err := i.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if i.Status != domain.IncidentStatusClosed {
		t.Errorf("expected closed, got %s", i.Status)
	}

	// Duration should be reasonable
	duration := i.Duration()
	if duration < 0 || duration > time.Hour {
		t.Errorf("unexpected duration: %v", duration)
	}
}
