package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IncidentsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_dashboard_incidents_created_total",
		Help: "Total number of incidents created",
	}, []string{"severity"})

	IncidentsResolved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_dashboard_incidents_resolved_total",
		Help: "Total number of incidents resolved",
	})

	DeploymentsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_dashboard_deployments_created_total",
		Help: "Total number of deployments created",
	}, []string{"environment"})

	DeploymentsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_dashboard_deployments_failed_total",
		Help: "Total number of failed deployments",
	}, []string{"environment"})

	AlertRulesFired = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_dashboard_alert_rules_fired_total",
		Help: "Total number of alert rules triggered",
	}, []string{"severity"})

	ServiceHealthChecks = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_dashboard_service_health_checks_total",
		Help: "Total number of service health checks performed",
	}, []string{"status"})

	IncidentResolutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_dashboard_incident_resolution_duration_seconds",
		Help: "Incident resolution time in seconds",
		Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400},
	}, []string{"severity"})

	DashboardOperationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_dashboard_operation_duration_seconds",
		Help: "Dashboard operation latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"operation"})
)
