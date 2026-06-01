package mysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/domain"
)

type ServiceHealthRepository struct {
	db *sqlx.DB
}

func NewServiceHealthRepository(db *sqlx.DB) *ServiceHealthRepository {
	return &ServiceHealthRepository{db: db}
}

func (r *ServiceHealthRepository) FindByID(ctx context.Context, id string) (*domain.ServiceHealth, error) {
	var h domain.ServiceHealth
	err := r.db.GetContext(ctx, &h, "SELECT id, service_name, status, health_url, response_time_ms, last_checked_at, last_error, version, environment, created_at, updated_at FROM service_health WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *ServiceHealthRepository) FindByServiceName(ctx context.Context, name string) (*domain.ServiceHealth, error) {
	var h domain.ServiceHealth
	err := r.db.GetContext(ctx, &h, "SELECT id, service_name, status, health_url, response_time_ms, last_checked_at, last_error, version, environment, created_at, updated_at FROM service_health WHERE service_name = ?", name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *ServiceHealthRepository) FindAll(ctx context.Context) ([]*domain.ServiceHealth, error) {
	var services []*domain.ServiceHealth
	err := r.db.SelectContext(ctx, &services, "SELECT id, service_name, status, health_url, response_time_ms, last_checked_at, last_error, version, environment, created_at, updated_at FROM service_health ORDER BY service_name ASC")
	return services, err
}

func (r *ServiceHealthRepository) FindByStatus(ctx context.Context, status string) ([]*domain.ServiceHealth, error) {
	var services []*domain.ServiceHealth
	err := r.db.SelectContext(ctx, &services, "SELECT id, service_name, status, health_url, response_time_ms, last_checked_at, last_error, version, environment, created_at, updated_at FROM service_health WHERE status = ? ORDER BY service_name ASC", status)
	return services, err
}

func (r *ServiceHealthRepository) Create(ctx context.Context, h *domain.ServiceHealth) error {
	query := `INSERT INTO service_health (id, service_name, status, health_url, response_time_ms, last_checked_at, last_error, version, environment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, h.ID, h.ServiceName, h.Status, h.HealthURL, h.ResponseTime, h.LastCheckedAt, h.LastError, h.Version, h.Environment, h.CreatedAt, h.UpdatedAt)
	return err
}

func (r *ServiceHealthRepository) Update(ctx context.Context, h *domain.ServiceHealth) error {
	query := `UPDATE service_health SET status = ?, response_time_ms = ?, last_checked_at = ?, last_error = ?, version = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, h.Status, h.ResponseTime, h.LastCheckedAt, h.LastError, h.Version, h.UpdatedAt, h.ID)
	return err
}

func (r *ServiceHealthRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM service_health WHERE id = ?", id)
	return err
}

func (r *ServiceHealthRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
}

type DeploymentRepository struct {
	db *sqlx.DB
}

func NewDeploymentRepository(db *sqlx.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) FindByID(ctx context.Context, id string) (*domain.Deployment, error) {
	var d domain.Deployment
	err := r.db.GetContext(ctx, &d, "SELECT id, service_name, version, environment, status, deployed_by, image, replicas, ready_replicas, strategy, started_at, finished_at, rollback_of, notes, created_at, updated_at FROM deployments WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeploymentRepository) FindByServiceName(ctx context.Context, serviceName string, limit int) ([]*domain.Deployment, error) {
	var deployments []*domain.Deployment
	err := r.db.SelectContext(ctx, &deployments, "SELECT id, service_name, version, environment, status, deployed_by, image, replicas, ready_replicas, strategy, started_at, finished_at, rollback_of, notes, created_at, updated_at FROM deployments WHERE service_name = ? ORDER BY created_at DESC LIMIT ?", serviceName, limit)
	return deployments, err
}

func (r *DeploymentRepository) FindActive(ctx context.Context) ([]*domain.Deployment, error) {
	var deployments []*domain.Deployment
	err := r.db.SelectContext(ctx, &deployments, "SELECT id, service_name, version, environment, status, deployed_by, image, replicas, ready_replicas, strategy, started_at, finished_at, rollback_of, notes, created_at, updated_at FROM deployments WHERE status IN ('pending', 'in_progress') ORDER BY started_at DESC")
	return deployments, err
}

func (r *DeploymentRepository) FindRecent(ctx context.Context, limit int) ([]*domain.Deployment, error) {
	var deployments []*domain.Deployment
	err := r.db.SelectContext(ctx, &deployments, "SELECT id, service_name, version, environment, status, deployed_by, image, replicas, ready_replicas, strategy, started_at, finished_at, rollback_of, notes, created_at, updated_at FROM deployments ORDER BY created_at DESC LIMIT ?", limit)
	return deployments, err
}

func (r *DeploymentRepository) Create(ctx context.Context, d *domain.Deployment) error {
	query := `INSERT INTO deployments (id, service_name, version, environment, status, deployed_by, image, replicas, ready_replicas, strategy, started_at, finished_at, rollback_of, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, d.ID, d.ServiceName, d.Version, d.Environment, d.Status, d.DeployedBy, d.Image, d.Replicas, d.ReadyReplicas, d.Strategy, d.StartedAt, d.FinishedAt, d.RollbackOf, d.Notes, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *DeploymentRepository) Update(ctx context.Context, d *domain.Deployment) error {
	query := `UPDATE deployments SET status = ?, ready_replicas = ?, finished_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, d.Status, d.ReadyReplicas, d.FinishedAt, d.UpdatedAt, d.ID)
	return err
}

func (r *DeploymentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM deployments WHERE id = ?", id)
	return err
}

func (r *DeploymentRepository) CountByStatus(ctx context.Context, status string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM deployments WHERE status = ?", status)
	return count, err
}

type IncidentRepository struct {
	db *sqlx.DB
}

func NewIncidentRepository(db *sqlx.DB) *IncidentRepository {
	return &IncidentRepository{db: db}
}

func (r *IncidentRepository) FindByID(ctx context.Context, id string) (*domain.Incident, error) {
	var i domain.Incident
	err := r.db.GetContext(ctx, &i, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *IncidentRepository) FindActive(ctx context.Context) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.SelectContext(ctx, &incidents, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents WHERE status NOT IN ('resolved', 'closed') ORDER BY detected_at DESC")
	return incidents, err
}

func (r *IncidentRepository) FindBySeverity(ctx context.Context, severity string) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.SelectContext(ctx, &incidents, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents WHERE severity = ? AND status NOT IN ('resolved', 'closed') ORDER BY detected_at DESC", severity)
	return incidents, err
}

func (r *IncidentRepository) FindByStatus(ctx context.Context, status string) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.SelectContext(ctx, &incidents, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents WHERE status = ? ORDER BY detected_at DESC", status)
	return incidents, err
}

func (r *IncidentRepository) FindByServiceName(ctx context.Context, serviceName string) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.SelectContext(ctx, &incidents, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents WHERE service_names LIKE ? AND status NOT IN ('resolved', 'closed') ORDER BY detected_at DESC", "%"+serviceName+"%")
	return incidents, err
}

func (r *IncidentRepository) FindRecent(ctx context.Context, limit int) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.db.SelectContext(ctx, &incidents, "SELECT id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at FROM incidents ORDER BY created_at DESC LIMIT ?", limit)
	return incidents, err
}

func (r *IncidentRepository) Create(ctx context.Context, i *domain.Incident) error {
	query := `INSERT INTO incidents (id, title, description, severity, status, service_names, detected_at, acknowledged_at, resolved_at, root_cause, resolution, created_by, assigned_to, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, i.ID, i.Title, i.Description, i.Severity, i.Status, i.ServiceNames, i.DetectedAt, i.AcknowledgedAt, i.ResolvedAt, i.RootCause, i.Resolution, i.CreatedBy, i.AssignedTo, i.CreatedAt, i.UpdatedAt)
	return err
}

func (r *IncidentRepository) Update(ctx context.Context, i *domain.Incident) error {
	query := `UPDATE incidents SET title = ?, description = ?, severity = ?, status = ?, service_names = ?, acknowledged_at = ?, resolved_at = ?, root_cause = ?, resolution = ?, assigned_to = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, i.Title, i.Description, i.Severity, i.Status, i.ServiceNames, i.AcknowledgedAt, i.ResolvedAt, i.RootCause, i.Resolution, i.AssignedTo, i.UpdatedAt, i.ID)
	return err
}

func (r *IncidentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM incidents WHERE id = ?", id)
	return err
}

func (r *IncidentRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM incidents WHERE status NOT IN ('resolved', 'closed')")
	return count, err
}

func (r *IncidentRepository) CountBySeverity(ctx context.Context, severity string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM incidents WHERE severity = ? AND status NOT IN ('resolved', 'closed')", severity)
	return count, err
}

func (r *IncidentRepository) CountActiveBySeverity(ctx context.Context) (map[string]int, error) {
	type SeverityCount struct {
		Severity string `db:"severity"`
		Count    int    `db:"count"`
	}
	var results []SeverityCount
	err := r.db.SelectContext(ctx, &results, "SELECT severity, COUNT(*) AS count FROM incidents WHERE status NOT IN ('resolved', 'closed') GROUP BY severity")
	if err != nil {
		return nil, err
	}
	m := make(map[string]int)
	for _, r := range results {
		m[r.Severity] = r.Count
	}
	return m, nil
}

type AlertRuleRepository struct {
	db *sqlx.DB
}

func NewAlertRuleRepository(db *sqlx.DB) *AlertRuleRepository {
	return &AlertRuleRepository{db: db}
}

func (r *AlertRuleRepository) FindByID(ctx context.Context, id string) (*domain.AlertRule, error) {
	var a domain.AlertRule
	err := r.db.GetContext(ctx, &a, "SELECT id, name, description, service_name, metric_name, condition, threshold, duration, severity, enabled, notify_channels, created_by, created_at, updated_at FROM alert_rules WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AlertRuleRepository) FindAll(ctx context.Context) ([]*domain.AlertRule, error) {
	var rules []*domain.AlertRule
	err := r.db.SelectContext(ctx, &rules, "SELECT id, name, description, service_name, metric_name, condition, threshold, duration, severity, enabled, notify_channels, created_by, created_at, updated_at FROM alert_rules ORDER BY service_name ASC, name ASC")
	return rules, err
}

func (r *AlertRuleRepository) FindByServiceName(ctx context.Context, serviceName string) ([]*domain.AlertRule, error) {
	var rules []*domain.AlertRule
	err := r.db.SelectContext(ctx, &rules, "SELECT id, name, description, service_name, metric_name, condition, threshold, duration, severity, enabled, notify_channels, created_by, created_at, updated_at FROM alert_rules WHERE service_name = ? ORDER BY name ASC", serviceName)
	return rules, err
}

func (r *AlertRuleRepository) FindEnabled(ctx context.Context) ([]*domain.AlertRule, error) {
	var rules []*domain.AlertRule
	err := r.db.SelectContext(ctx, &rules, "SELECT id, name, description, service_name, metric_name, condition, threshold, duration, severity, enabled, notify_channels, created_by, created_at, updated_at FROM alert_rules WHERE enabled = TRUE ORDER BY service_name ASC")
	return rules, err
}

func (r *AlertRuleRepository) Create(ctx context.Context, a *domain.AlertRule) error {
	query := `INSERT INTO alert_rules (id, name, description, service_name, metric_name, condition, threshold, duration, severity, enabled, notify_channels, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.Name, a.Description, a.ServiceName, a.MetricName, a.Condition, a.Threshold, a.Duration, a.Severity, a.Enabled, a.NotifyChannels, a.CreatedBy, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AlertRuleRepository) Update(ctx context.Context, a *domain.AlertRule) error {
	query := `UPDATE alert_rules SET name = ?, description = ?, service_name = ?, metric_name = ?, condition = ?, threshold = ?, duration = ?, severity = ?, enabled = ?, notify_channels = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, a.Name, a.Description, a.ServiceName, a.MetricName, a.Condition, a.Threshold, a.Duration, a.Severity, a.Enabled, a.NotifyChannels, a.UpdatedAt, a.ID)
	return err
}

func (r *AlertRuleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM alert_rules WHERE id = ?", id)
	return err
}

func (r *AlertRuleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM alert_rules")
	return count, err
}

type AuditLogRepository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) FindByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	var a domain.AuditLog
	err := r.db.GetContext(ctx, &a, "SELECT id, actor, action, resource, resource_id, details, ip_address, user_agent, created_at FROM audit_logs WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AuditLogRepository) FindByActor(ctx context.Context, actor string, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, "SELECT id, actor, action, resource, resource_id, details, ip_address, user_agent, created_at FROM audit_logs WHERE actor = ? ORDER BY created_at DESC LIMIT ?", actor, limit)
	return logs, err
}

func (r *AuditLogRepository) FindByResource(ctx context.Context, resource, resourceID string, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, "SELECT id, actor, action, resource, resource_id, details, ip_address, user_agent, created_at FROM audit_logs WHERE resource = ? AND resource_id = ? ORDER BY created_at DESC LIMIT ?", resource, resourceID, limit)
	return logs, err
}

func (r *AuditLogRepository) FindRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, "SELECT id, actor, action, resource, resource_id, details, ip_address, user_agent, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ?", limit)
	return logs, err
}

func (r *AuditLogRepository) Create(ctx context.Context, a *domain.AuditLog) error {
	query := `INSERT INTO audit_logs (id, actor, action, resource, resource_id, details, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.Actor, a.Action, a.Resource, a.ResourceID, a.Details, a.IPAddress, a.UserAgent, a.CreatedAt)
	return err
}

type ServiceDependencyRepository struct {
	db *sqlx.DB
}

func NewServiceDependencyRepository(db *sqlx.DB) *ServiceDependencyRepository {
	return &ServiceDependencyRepository{db: db}
}

func (r *ServiceDependencyRepository) FindByService(ctx context.Context, serviceName string) ([]*domain.ServiceDependency, error) {
	var deps []*domain.ServiceDependency
	err := r.db.SelectContext(ctx, &deps, "SELECT id, service_name, depends_on, dependency_type, critical, created_at FROM service_dependencies WHERE service_name = ? ORDER BY depends_on ASC", serviceName)
	return deps, err
}

func (r *ServiceDependencyRepository) FindAll(ctx context.Context) ([]*domain.ServiceDependency, error) {
	var deps []*domain.ServiceDependency
	err := r.db.SelectContext(ctx, &deps, "SELECT id, service_name, depends_on, dependency_type, critical, created_at FROM service_dependencies ORDER BY service_name ASC, depends_on ASC")
	return deps, err
}

func (r *ServiceDependencyRepository) FindDependents(ctx context.Context, serviceName string) ([]*domain.ServiceDependency, error) {
	var deps []*domain.ServiceDependency
	err := r.db.SelectContext(ctx, &deps, "SELECT id, service_name, depends_on, dependency_type, critical, created_at FROM service_dependencies WHERE depends_on = ? ORDER BY service_name ASC", serviceName)
	return deps, err
}

func (r *ServiceDependencyRepository) Create(ctx context.Context, d *domain.ServiceDependency) error {
	query := `INSERT INTO service_dependencies (id, service_name, depends_on, dependency_type, critical, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, d.ID, d.ServiceName, d.DependsOn, d.DependencyType, d.Critical, d.CreatedAt)
	return err
}

func (r *ServiceDependencyRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM service_dependencies WHERE id = ?", id)
	return err
}

func (r *ServiceDependencyRepository) DeleteByService(ctx context.Context, serviceName string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM service_dependencies WHERE service_name = ?", serviceName)
	return err
}

type CapacityMetricRepository struct {
	db *sqlx.DB
}

func NewCapacityMetricRepository(db *sqlx.DB) *CapacityMetricRepository {
	return &CapacityMetricRepository{db: db}
}

func (r *CapacityMetricRepository) FindByService(ctx context.Context, serviceName string) ([]*domain.CapacityMetric, error) {
	var metrics []*domain.CapacityMetric
	err := r.db.SelectContext(ctx, &metrics, "SELECT id, service_name, resource_type, current_value, max_value, unit, recorded_at FROM capacity_metrics WHERE service_name = ? ORDER BY resource_type ASC, recorded_at DESC", serviceName)
	return metrics, err
}

func (r *CapacityMetricRepository) FindByResource(ctx context.Context, serviceName, resourceType string, limit int) ([]*domain.CapacityMetric, error) {
	var metrics []*domain.CapacityMetric
	err := r.db.SelectContext(ctx, &metrics, "SELECT id, service_name, resource_type, current_value, max_value, unit, recorded_at FROM capacity_metrics WHERE service_name = ? AND resource_type = ? ORDER BY recorded_at DESC LIMIT ?", serviceName, resourceType, limit)
	return metrics, err
}

func (r *CapacityMetricRepository) FindLatest(ctx context.Context) ([]*domain.CapacityMetric, error) {
	var metrics []*domain.CapacityMetric
	err := r.db.SelectContext(ctx, &metrics, `SELECT cm.id, cm.service_name, cm.resource_type, cm.current_value, cm.max_value, cm.unit, cm.recorded_at
		FROM capacity_metrics cm
		INNER JOIN (
			SELECT service_name, resource_type, MAX(recorded_at) AS max_recorded_at
			FROM capacity_metrics
			GROUP BY service_name, resource_type
		) latest ON cm.service_name = latest.service_name AND cm.resource_type = latest.resource_type AND cm.recorded_at = latest.max_recorded_at
		ORDER BY cm.service_name ASC, cm.resource_type ASC`)
	return metrics, err
}

func (r *CapacityMetricRepository) Create(ctx context.Context, m *domain.CapacityMetric) error {
	query := `INSERT INTO capacity_metrics (id, service_name, resource_type, current_value, max_value, unit, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.ServiceName, m.ResourceType, m.CurrentValue, m.MaxValue, m.Unit, m.RecordedAt)
	return err
}
