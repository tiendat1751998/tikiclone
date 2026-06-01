package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type CacheProvider interface {
	GetDashboardSummary(ctx context.Context) ([]byte, error)
	SetDashboardSummary(ctx context.Context, data []byte, ttl time.Duration) error
	InvalidateSummary(ctx context.Context) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event *domain.DashboardEvent) error
}

type DashboardService struct {
	healthRepo    domain.ServiceHealthRepository
	deployRepo    domain.DeploymentRepository
	incidentRepo  domain.IncidentRepository
	alertRepo     domain.AlertRuleRepository
	auditRepo     domain.AuditLogRepository
	depRepo       domain.ServiceDependencyRepository
	capacityRepo  domain.CapacityMetricRepository
	publisher     EventPublisher
	incidentTTL   time.Duration
	cache         CacheProvider
	summaryTTL    time.Duration
}

func NewDashboardService(
	healthRepo domain.ServiceHealthRepository,
	deployRepo domain.DeploymentRepository,
	incidentRepo domain.IncidentRepository,
	alertRepo domain.AlertRuleRepository,
	auditRepo domain.AuditLogRepository,
	depRepo domain.ServiceDependencyRepository,
	capacityRepo domain.CapacityMetricRepository,
	publisher EventPublisher,
	incidentTTL time.Duration,
	cache CacheProvider,
	summaryTTL time.Duration,
) *DashboardService {
	return &DashboardService{
		healthRepo:   healthRepo,
		deployRepo:   deployRepo,
		incidentRepo: incidentRepo,
		alertRepo:    alertRepo,
		auditRepo:    auditRepo,
		depRepo:      depRepo,
		capacityRepo: capacityRepo,
		publisher:    publisher,
		incidentTTL:  incidentTTL,
		cache:        cache,
		summaryTTL:   summaryTTL,
	}
}

func (s *DashboardService) saveAuditLog(ctx context.Context, log *domain.AuditLog) {
	if err := s.auditRepo.Create(ctx, log); err != nil {
		observability.LogWithTrace(ctx).Error("failed to save audit log",
			zap.String("action", log.Action),
			zap.String("resource", log.Resource),
			zap.Error(err))
	}
}

func (s *DashboardService) publishEvent(ctx context.Context, evt *domain.DashboardEvent) {
	if s.publisher != nil {
		if evt.ID == "" {
			evt.ID = uuid.New().String()
		}
		if err := s.publisher.Publish(ctx, evt); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish event",
				zap.String("event_type", evt.EventType),
				zap.Error(err))
		}
	}
}

// === Dashboard Summary ===

func (s *DashboardService) GetSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_summary")
	defer span.End()

	if s.cache != nil {
		if data, err := s.cache.GetDashboardSummary(ctx); err == nil && len(data) > 0 {
			var cached domain.DashboardSummary
			if err := json.Unmarshal(data, &cached); err == nil {
				return &cached, nil
			}
		}
	}

	var (
		services          []*domain.ServiceHealth
		activeIncidents   []*domain.Incident
		activeDeployments []*domain.Deployment
		recentDeployments []*domain.Deployment
		capacity          []*domain.CapacityMetric
	)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		services, err = s.healthRepo.FindAll(gCtx)
		return err
	})
	g.Go(func() (err error) {
		activeIncidents, err = s.incidentRepo.FindActive(gCtx)
		return err
	})
	g.Go(func() (err error) {
		activeDeployments, err = s.deployRepo.FindActive(gCtx)
		return err
	})
	g.Go(func() (err error) {
		recentDeployments, err = s.deployRepo.FindRecent(gCtx, 10)
		return err
	})
	g.Go(func() (err error) {
		capacity, err = s.capacityRepo.FindLatest(gCtx)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("dashboard summary: %w", err)
	}

	summary := &domain.DashboardSummary{
		TotalServices:      len(services),
		ServiceHealthList:  services,
		OpenIncidents:      activeIncidents,
		ActiveIncidents:    len(activeIncidents),
		ActiveDeployments:  len(activeDeployments),
		RecentDeployments:  recentDeployments,
		CapacityOverview:   capacity,
	}

	for _, svc := range services {
		switch svc.Status {
		case domain.ServiceStatusHealthy:
			summary.HealthyServices++
		case domain.ServiceStatusDegraded:
			summary.DegradedServices++
		case domain.ServiceStatusUnhealthy:
			summary.UnhealthyServices++
		}
	}
	for _, inc := range activeIncidents {
		if inc.Severity == domain.IncidentSeverityCritical {
			summary.CriticalIncidents++
		}
	}

	if s.cache != nil && s.summaryTTL > 0 {
		if data, err := json.Marshal(summary); err == nil {
			s.cache.SetDashboardSummary(ctx, data, s.summaryTTL)
		}
	}

	return summary, nil
}

// === Service Health Operations ===

type RegisterServiceRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	HealthURL   string `json:"health_url" binding:"required"`
	Environment string `json:"environment"`
}

func (s *DashboardService) RegisterService(ctx context.Context, req RegisterServiceRequest, actor string) (*domain.ServiceHealth, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.register_service")
	defer span.End()

	existing, err := s.healthRepo.FindByServiceName(ctx, req.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("check existing service: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrDuplicateService, req.ServiceName)
	}

	env := req.Environment
	if env == "" {
		env = "production"
	}

	health := domain.NewServiceHealth(req.ServiceName, req.HealthURL, env)
	if err := s.healthRepo.Create(ctx, health); err != nil {
		return nil, fmt.Errorf("create service health: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionCreate, "service", health.ID, fmt.Sprintf("Registered service %s", req.ServiceName), "", ""))

	return health, nil
}

type UpdateServiceHealthRequest struct {
	Status       string `json:"status" binding:"required"`
	ResponseTime int    `json:"response_time_ms"`
	Version      string `json:"version"`
	LastError    string `json:"last_error"`
}

func (s *DashboardService) UpdateServiceHealth(ctx context.Context, serviceID string, req UpdateServiceHealthRequest) (*domain.ServiceHealth, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.update_service_health")
	defer span.End()

	health, err := s.healthRepo.FindByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, domain.ErrServiceNotFound
	}

	previousStatus := health.Status

	switch req.Status {
	case domain.ServiceStatusHealthy:
		health.MarkHealthy(req.ResponseTime, req.Version)
	case domain.ServiceStatusDegraded:
		health.MarkDegraded(req.ResponseTime, req.LastError)
	case domain.ServiceStatusUnhealthy:
		health.MarkUnhealthy(req.LastError)
	default:
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	if err := s.healthRepo.Update(ctx, health); err != nil {
		return nil, fmt.Errorf("update service health: %w", err)
	}

	if previousStatus != req.Status && s.publisher != nil {
		s.publisher.Publish(ctx, &domain.DashboardEvent{
			ID:            uuid.New().String(),
			EventType:     domain.EventServiceStatusChanged,
			AggregateType: "service_health",
			AggregateID:   health.ID,
			Payload: domain.ServiceStatusChangedPayload{
				ServiceName:    health.ServiceName,
				PreviousStatus: previousStatus,
				CurrentStatus:  req.Status,
			},
			CreatedAt: time.Now(),
		})
	}

	return health, nil
}

func (s *DashboardService) GetServiceHealth(ctx context.Context, serviceID string) (*domain.ServiceHealth, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_service_health")
	defer span.End()

	health, err := s.healthRepo.FindByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, domain.ErrServiceNotFound
	}
	return health, nil
}

func (s *DashboardService) ListServices(ctx context.Context) ([]*domain.ServiceHealth, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_services")
	defer span.End()

	return s.healthRepo.FindAll(ctx)
}

// === Deployment Operations ===

type CreateDeploymentRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Version     string `json:"version" binding:"required"`
	Environment string `json:"environment" binding:"required"`
	Image       string `json:"image" binding:"required"`
	Replicas    int    `json:"replicas" binding:"required,min=1"`
	Strategy    string `json:"strategy"`
	Notes       string `json:"notes"`
}

func (s *DashboardService) CreateDeployment(ctx context.Context, req CreateDeploymentRequest, actor string) (*domain.Deployment, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.create_deployment")
	defer span.End()

	strategy := req.Strategy
	if strategy == "" {
		strategy = "rolling"
	}

	deployment := domain.NewDeployment(req.ServiceName, req.Version, req.Environment, actor, req.Image, req.Replicas, strategy)
	deployment.MarkInProgress()

	if err := s.deployRepo.Create(ctx, deployment); err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionDeploy, "deployment", deployment.ID,
		fmt.Sprintf("Deployed %s v%s to %s", req.ServiceName, req.Version, req.Environment), "", ""))
	s.publishEvent(ctx, &domain.DashboardEvent{
		ID:            uuid.New().String(),
		EventType:     domain.EventDeploymentStarted,
		AggregateType: "deployment",
		AggregateID:   deployment.ID,
		Payload: domain.DeploymentEventPayload{
			DeploymentID: deployment.ID,
			ServiceName:  req.ServiceName,
			Version:      req.Version,
			Environment:  req.Environment,
		},
		CreatedAt: time.Now(),
	})

	return deployment, nil
}

type UpdateDeploymentRequest struct {
	Status        string `json:"status" binding:"required"`
	ReadyReplicas int    `json:"ready_replicas"`
}

func (s *DashboardService) UpdateDeployment(ctx context.Context, deploymentID string, req UpdateDeploymentRequest, actor string) (*domain.Deployment, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.update_deployment")
	defer span.End()

	deployment, err := s.deployRepo.FindByID(ctx, deploymentID)
	if err != nil {
		return nil, err
	}
	if deployment == nil {
		return nil, domain.ErrDeploymentNotFound
	}

	var eventType string
	switch req.Status {
	case domain.DeploymentStatusSucceeded:
		deployment.MarkSucceeded(req.ReadyReplicas)
		eventType = domain.EventDeploymentSucceeded
	case domain.DeploymentStatusFailed:
		deployment.MarkFailed()
		eventType = domain.EventDeploymentFailed
	case domain.DeploymentStatusRolledBack:
		deployment.MarkRolledBack()
		eventType = domain.EventDeploymentRolledBack
	default:
		return nil, fmt.Errorf("invalid deployment status: %s", req.Status)
	}

	if err := s.deployRepo.Update(ctx, deployment); err != nil {
		return nil, fmt.Errorf("update deployment: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionUpdate, "deployment", deployment.ID,
		fmt.Sprintf("Deployment %s -> %s", deploymentID, req.Status), "", ""))
	if eventType != "" {
		s.publishEvent(ctx, &domain.DashboardEvent{
			ID:            uuid.New().String(),
			EventType:     eventType,
			AggregateType: "deployment",
			AggregateID:   deployment.ID,
			Payload: domain.DeploymentEventPayload{
				DeploymentID: deployment.ID,
				ServiceName:  deployment.ServiceName,
				Version:      deployment.Version,
				Environment:  deployment.Environment,
			},
			CreatedAt: time.Now(),
		})
	}

	return deployment, nil
}

func (s *DashboardService) GetDeployment(ctx context.Context, deploymentID string) (*domain.Deployment, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_deployment")
	defer span.End()

	deployment, err := s.deployRepo.FindByID(ctx, deploymentID)
	if err != nil {
		return nil, err
	}
	if deployment == nil {
		return nil, domain.ErrDeploymentNotFound
	}
	return deployment, nil
}

func (s *DashboardService) ListDeployments(ctx context.Context, serviceName string, limit int) ([]*domain.Deployment, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_deployments")
	defer span.End()

	if serviceName != "" {
		return s.deployRepo.FindByServiceName(ctx, serviceName, limit)
	}
	return s.deployRepo.FindRecent(ctx, limit)
}

func (s *DashboardService) ListActiveDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_active_deployments")
	defer span.End()

	return s.deployRepo.FindActive(ctx)
}

// === Incident Operations ===

type CreateIncidentRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description" binding:"required"`
	Severity     string `json:"severity" binding:"required"`
	ServiceNames string `json:"service_names" binding:"required"`
}

func (s *DashboardService) CreateIncident(ctx context.Context, req CreateIncidentRequest, actor string) (*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.create_incident")
	defer span.End()

	span.SetAttributes(
		attribute.String("severity", req.Severity),
		attribute.String("services", req.ServiceNames),
	)

	incident := domain.NewIncident(req.Title, req.Description, req.Severity, req.ServiceNames, actor)
	if err := s.incidentRepo.Create(ctx, incident); err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionCreate, "incident", incident.ID,
		fmt.Sprintf("Created incident: %s", req.Title), "", ""))
	s.publishEvent(ctx, &domain.DashboardEvent{
		ID:            uuid.New().String(),
		EventType:     domain.EventIncidentCreated,
		AggregateType: "incident",
		AggregateID:   incident.ID,
		Payload: domain.IncidentCreatedPayload{
			IncidentID: incident.ID,
			Title:      req.Title,
			Severity:   req.Severity,
			Services:   req.ServiceNames,
		},
		CreatedAt: time.Now(),
	})

	return incident, nil
}

type AcknowledgeIncidentRequest struct {
	AssignedTo string `json:"assigned_to"`
}

func (s *DashboardService) AcknowledgeIncident(ctx context.Context, incidentID string, req AcknowledgeIncidentRequest, actor string) (*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.acknowledge_incident")
	defer span.End()

	incident, err := s.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrIncidentNotFound
	}

	if err := incident.Acknowledge(req.AssignedTo); err != nil {
		return nil, err
	}

	if err := s.incidentRepo.Update(ctx, incident); err != nil {
		return nil, fmt.Errorf("update incident: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionAcknowledge, "incident", incident.ID,
		fmt.Sprintf("Acknowledged incident, assigned to %s", req.AssignedTo), "", ""))
	s.publishEvent(ctx, &domain.DashboardEvent{
		ID:            uuid.New().String(),
		EventType:     domain.EventIncidentAcknowledged,
		AggregateType: "incident",
		AggregateID:   incident.ID,
		CreatedAt:     time.Now(),
	})

	return incident, nil
}

type ResolveIncidentRequest struct {
	RootCause  string `json:"root_cause" binding:"required"`
	Resolution string `json:"resolution" binding:"required"`
}

func (s *DashboardService) ResolveIncident(ctx context.Context, incidentID string, req ResolveIncidentRequest, actor string) (*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.resolve_incident")
	defer span.End()

	incident, err := s.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrIncidentNotFound
	}

	if err := incident.Resolve(req.RootCause, req.Resolution); err != nil {
		return nil, err
	}

	if err := s.incidentRepo.Update(ctx, incident); err != nil {
		return nil, fmt.Errorf("update incident: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionResolve, "incident", incident.ID,
		fmt.Sprintf("Resolved incident: %s", req.RootCause), "", ""))
	s.publishEvent(ctx, &domain.DashboardEvent{
		ID:            uuid.New().String(),
		EventType:     domain.EventIncidentResolved,
		AggregateType: "incident",
		AggregateID:   incident.ID,
		Payload: domain.IncidentResolvedPayload{
			IncidentID: incident.ID,
			RootCause:  req.RootCause,
			Duration:   incident.Duration().String(),
		},
		CreatedAt: time.Now(),
	})

	return incident, nil
}

func (s *DashboardService) CloseIncident(ctx context.Context, incidentID string, actor string) (*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.close_incident")
	defer span.End()

	incident, err := s.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrIncidentNotFound
	}

	if err := incident.Close(); err != nil {
		return nil, err
	}

	if err := s.incidentRepo.Update(ctx, incident); err != nil {
		return nil, fmt.Errorf("update incident: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionClose, "incident", incident.ID, "Closed incident", "", ""))

	return incident, nil
}

func (s *DashboardService) GetIncident(ctx context.Context, incidentID string) (*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_incident")
	defer span.End()

	incident, err := s.incidentRepo.FindByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrIncidentNotFound
	}
	return incident, nil
}

func (s *DashboardService) ListActiveIncidents(ctx context.Context) ([]*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_active_incidents")
	defer span.End()

	return s.incidentRepo.FindActive(ctx)
}

func (s *DashboardService) ListRecentIncidents(ctx context.Context, limit int) ([]*domain.Incident, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_recent_incidents")
	defer span.End()

	return s.incidentRepo.FindRecent(ctx, limit)
}

// === Alert Rule Operations ===

type CreateAlertRuleRequest struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	ServiceName    string  `json:"service_name" binding:"required"`
	MetricName     string  `json:"metric_name" binding:"required"`
	Condition      string  `json:"condition" binding:"required"`
	Threshold      float64 `json:"threshold" binding:"required"`
	Duration       string  `json:"duration"`
	Severity       string  `json:"severity" binding:"required"`
	NotifyChannels string  `json:"notify_channels"`
}

func (s *DashboardService) CreateAlertRule(ctx context.Context, req CreateAlertRuleRequest, actor string) (*domain.AlertRule, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.create_alert_rule")
	defer span.End()

	count, err := s.alertRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count alert rules: %w", err)
	}
	if count >= 500 {
		return nil, fmt.Errorf("%w: max 500 alert rules", domain.ErrDuplicateAlertRule)
	}

	duration := req.Duration
	if duration == "" {
		duration = "5m"
	}

	rule := domain.NewAlertRule(req.Name, req.Description, req.ServiceName, req.MetricName, req.Condition, req.Threshold, duration, req.Severity, req.NotifyChannels, actor)
	if err := s.alertRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create alert rule: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionCreate, "alert_rule", rule.ID,
		fmt.Sprintf("Created alert rule: %s", req.Name), "", ""))

	return rule, nil
}

type UpdateAlertRuleRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	MetricName     string  `json:"metric_name"`
	Condition      string  `json:"condition"`
	Threshold      float64 `json:"threshold"`
	Duration       string  `json:"duration"`
	Severity       string  `json:"severity"`
	Enabled        *bool   `json:"enabled"`
	NotifyChannels string  `json:"notify_channels"`
}

func (s *DashboardService) UpdateAlertRule(ctx context.Context, ruleID string, req UpdateAlertRuleRequest, actor string) (*domain.AlertRule, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.update_alert_rule")
	defer span.End()

	rule, err := s.alertRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, domain.ErrAlertRuleNotFound
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.MetricName != "" {
		rule.MetricName = req.MetricName
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Threshold != 0 {
		rule.Threshold = req.Threshold
	}
	if req.Duration != "" {
		rule.Duration = req.Duration
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.NotifyChannels != "" {
		rule.NotifyChannels = req.NotifyChannels
	}
	rule.UpdatedAt = time.Now()

	if err := s.alertRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("update alert rule: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionUpdate, "alert_rule", rule.ID, "Updated alert rule", "", ""))

	return rule, nil
}

func (s *DashboardService) DeleteAlertRule(ctx context.Context, ruleID string, actor string) error {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.delete_alert_rule")
	defer span.End()

	rule, err := s.alertRepo.FindByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if rule == nil {
		return domain.ErrAlertRuleNotFound
	}

	if err := s.alertRepo.Delete(ctx, ruleID); err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionDelete, "alert_rule", rule.ID, fmt.Sprintf("Deleted alert rule: %s", rule.Name), "", ""))

	return nil
}

func (s *DashboardService) GetAlertRule(ctx context.Context, ruleID string) (*domain.AlertRule, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_alert_rule")
	defer span.End()

	rule, err := s.alertRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, domain.ErrAlertRuleNotFound
	}
	return rule, nil
}

func (s *DashboardService) ListAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_alert_rules")
	defer span.End()

	return s.alertRepo.FindAll(ctx)
}

// === Audit Log Operations ===

func (s *DashboardService) ListAuditLogs(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_audit_logs")
	defer span.End()

	return s.auditRepo.FindRecent(ctx, limit)
}

func (s *DashboardService) ListAuditLogsByActor(ctx context.Context, actor string, limit int) ([]*domain.AuditLog, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.list_audit_logs_by_actor")
	defer span.End()

	return s.auditRepo.FindByActor(ctx, actor, limit)
}

// === Service Dependency Operations ===

type AddDependencyRequest struct {
	ServiceName    string `json:"service_name" binding:"required"`
	DependsOn      string `json:"depends_on" binding:"required"`
	DependencyType string `json:"dependency_type"`
	Critical       bool   `json:"critical"`
}

func (s *DashboardService) AddServiceDependency(ctx context.Context, req AddDependencyRequest, actor string) (*domain.ServiceDependency, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.add_dependency")
	defer span.End()

	depType := req.DependencyType
	if depType == "" {
		depType = domain.DependencyTypeSync
	}

	dep := domain.NewServiceDependency(req.ServiceName, req.DependsOn, depType, req.Critical)
	if err := s.depRepo.Create(ctx, dep); err != nil {
		return nil, fmt.Errorf("create dependency: %w", err)
	}

	s.saveAuditLog(ctx, domain.NewAuditLog(actor, domain.ActionCreate, "service_dependency", dep.ID,
		fmt.Sprintf("Added dependency: %s -> %s", req.ServiceName, req.DependsOn), "", ""))

	return dep, nil
}

func (s *DashboardService) GetServiceDependencies(ctx context.Context, serviceName string) ([]*domain.ServiceDependency, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_dependencies")
	defer span.End()

	return s.depRepo.FindByService(ctx, serviceName)
}

func (s *DashboardService) GetAllDependencies(ctx context.Context) ([]*domain.ServiceDependency, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_all_dependencies")
	defer span.End()

	return s.depRepo.FindAll(ctx)
}

// === Capacity Operations ===

type RecordCapacityRequest struct {
	ServiceName  string  `json:"service_name" binding:"required"`
	ResourceType string  `json:"resource_type" binding:"required"`
	CurrentValue float64 `json:"current_value" binding:"required"`
	MaxValue     float64 `json:"max_value" binding:"required"`
	Unit         string  `json:"unit" binding:"required"`
}

func (s *DashboardService) RecordCapacityMetric(ctx context.Context, req RecordCapacityRequest) (*domain.CapacityMetric, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.record_capacity")
	defer span.End()

	metric := domain.NewCapacityMetric(req.ServiceName, req.ResourceType, req.CurrentValue, req.MaxValue, req.Unit)
	if err := s.capacityRepo.Create(ctx, metric); err != nil {
		return nil, fmt.Errorf("create capacity metric: %w", err)
	}
	return metric, nil
}

func (s *DashboardService) GetCapacityMetrics(ctx context.Context, serviceName string) ([]*domain.CapacityMetric, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_capacity")
	defer span.End()

	return s.capacityRepo.FindByService(ctx, serviceName)
}

func (s *DashboardService) GetLatestCapacity(ctx context.Context) ([]*domain.CapacityMetric, error) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(ctx, "dashboard.get_latest_capacity")
	defer span.End()

	return s.capacityRepo.FindLatest(ctx)
}
