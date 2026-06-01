package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/application"
	"github.com/tikiclone/tiki/platforms/production-dashboard/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type Handler struct {
	service *application.DashboardService
}

func NewHandler(service *application.DashboardService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSummary(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_summary")
	defer span.End()

	summary, err := h.service.GetSummary(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) RegisterService(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.register_service")
	defer span.End()

	var req application.RegisterServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	health, err := h.service.RegisterService(ctx, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, health)
}

func (h *Handler) UpdateServiceHealth(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.update_service_health")
	defer span.End()

	serviceID := c.Param("service_id")
	var req application.UpdateServiceHealthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	health, err := h.service.UpdateServiceHealth(ctx, serviceID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, health)
}

func (h *Handler) GetServiceHealth(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_service_health")
	defer span.End()

	serviceID := c.Param("service_id")
	health, err := h.service.GetServiceHealth(ctx, serviceID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, health)
}

func (h *Handler) ListServices(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_services")
	defer span.End()

	services, err := h.service.ListServices(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *Handler) CreateDeployment(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.create_deployment")
	defer span.End()

	var req application.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	deployment, err := h.service.CreateDeployment(ctx, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, deployment)
}

func (h *Handler) UpdateDeployment(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.update_deployment")
	defer span.End()

	deploymentID := c.Param("deployment_id")
	var req application.UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	deployment, err := h.service.UpdateDeployment(ctx, deploymentID, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deployment)
}

func (h *Handler) GetDeployment(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_deployment")
	defer span.End()

	deploymentID := c.Param("deployment_id")
	deployment, err := h.service.GetDeployment(ctx, deploymentID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deployment)
}

func (h *Handler) ListDeployments(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_deployments")
	defer span.End()

	serviceName := c.Query("service_name")
	limit := 50
	deployments, err := h.service.ListDeployments(ctx, serviceName, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deployments)
}

func (h *Handler) ListActiveDeployments(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_active_deployments")
	defer span.End()

	deployments, err := h.service.ListActiveDeployments(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deployments)
}

func (h *Handler) CreateIncident(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.create_incident")
	defer span.End()

	var req application.CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	incident, err := h.service.CreateIncident(ctx, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	span.SetAttributes(attribute.String("incident_id", incident.ID))
	c.JSON(http.StatusCreated, incident)
}

func (h *Handler) AcknowledgeIncident(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.acknowledge_incident")
	defer span.End()

	incidentID := c.Param("incident_id")
	var req application.AcknowledgeIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	incident, err := h.service.AcknowledgeIncident(ctx, incidentID, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

func (h *Handler) ResolveIncident(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.resolve_incident")
	defer span.End()

	incidentID := c.Param("incident_id")
	var req application.ResolveIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	incident, err := h.service.ResolveIncident(ctx, incidentID, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

func (h *Handler) CloseIncident(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.close_incident")
	defer span.End()

	incidentID := c.Param("incident_id")
	actor := c.GetString("user_id")
	incident, err := h.service.CloseIncident(ctx, incidentID, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

func (h *Handler) GetIncident(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_incident")
	defer span.End()

	incidentID := c.Param("incident_id")
	incident, err := h.service.GetIncident(ctx, incidentID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incident)
}

func (h *Handler) ListActiveIncidents(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_active_incidents")
	defer span.End()

	incidents, err := h.service.ListActiveIncidents(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (h *Handler) ListRecentIncidents(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_recent_incidents")
	defer span.End()

	limit := 50
	incidents, err := h.service.ListRecentIncidents(ctx, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, incidents)
}

func (h *Handler) CreateAlertRule(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.create_alert_rule")
	defer span.End()

	var req application.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	rule, err := h.service.CreateAlertRule(ctx, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) UpdateAlertRule(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.update_alert_rule")
	defer span.End()

	ruleID := c.Param("rule_id")
	var req application.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	rule, err := h.service.UpdateAlertRule(ctx, ruleID, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *Handler) DeleteAlertRule(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.delete_alert_rule")
	defer span.End()

	ruleID := c.Param("rule_id")
	actor := c.GetString("user_id")
	if err := h.service.DeleteAlertRule(ctx, ruleID, actor); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert rule deleted"})
}

func (h *Handler) GetAlertRule(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_alert_rule")
	defer span.End()

	ruleID := c.Param("rule_id")
	rule, err := h.service.GetAlertRule(ctx, ruleID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *Handler) ListAlertRules(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_alert_rules")
	defer span.End()

	rules, err := h.service.ListAlertRules(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.list_audit_logs")
	defer span.End()

	limit := 100
	logs, err := h.service.ListAuditLogs(ctx, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *Handler) AddServiceDependency(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.add_dependency")
	defer span.End()

	var req application.AddDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	actor := c.GetString("user_id")
	dep, err := h.service.AddServiceDependency(ctx, req, actor)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dep)
}

func (h *Handler) GetServiceDependencies(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_dependencies")
	defer span.End()

	serviceName := c.Param("service_name")
	deps, err := h.service.GetServiceDependencies(ctx, serviceName)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deps)
}

func (h *Handler) GetAllDependencies(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_all_dependencies")
	defer span.End()

	deps, err := h.service.GetAllDependencies(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, deps)
}

func (h *Handler) RecordCapacityMetric(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.record_capacity")
	defer span.End()

	var req application.RecordCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	metric, err := h.service.RecordCapacityMetric(ctx, req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, metric)
}

func (h *Handler) GetCapacityMetrics(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_capacity")
	defer span.End()

	serviceName := c.Param("service_name")
	metrics, err := h.service.GetCapacityMetrics(ctx, serviceName)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) GetLatestCapacity(c *gin.Context) {
	ctx, span := otel.Tracer("tiki-dashboard").Start(c.Request.Context(), "http.get_latest_capacity")
	defer span.End()

	metrics, err := h.service.GetLatestCapacity(ctx)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, metrics)
}

var errorStatusMap = map[error]int{
	domain.ErrServiceNotFound:    http.StatusNotFound,
	domain.ErrDeploymentNotFound: http.StatusNotFound,
	domain.ErrIncidentNotFound:   http.StatusNotFound,
	domain.ErrAlertRuleNotFound:  http.StatusNotFound,
	domain.ErrInvalidIncidentState:   http.StatusConflict,
	domain.ErrInvalidDeploymentState: http.StatusConflict,
	domain.ErrDuplicateService:   http.StatusConflict,
	domain.ErrDuplicateAlertRule: http.StatusConflict,
}

func handleError(c *gin.Context, err error) {
	for domainErr, status := range errorStatusMap {
		if errors.Is(err, domainErr) {
			c.AbortWithStatusJSON(status, gin.H{"error_code": domainErr.Error(), "message": err.Error()})
			return
		}
	}
	observability.LogWithTrace(c.Request.Context()).Error("unhandled error", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error_code": "INTERNAL_ERROR", "message": "An unexpected error occurred"})
}
