package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsIngestedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_events_ingested_total", Help: "Total analytics events ingested",
	})
	ReportsGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_reports_generated_total", Help: "Total analytics reports generated",
	})
	DashboardViewsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_dashboard_views_total", Help: "Total dashboard views",
	})
	QueryLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_analytics_query_latency_seconds",
		Help:    "Query latency distribution",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})
	ScheduledReportsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_scheduled_reports_total", Help: "Total scheduled reports generated",
	})
	ActiveDashboards = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_analytics_active_dashboards", Help: "Number of active dashboards",
	})
	EventProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_analytics_event_processing_duration_seconds",
		Help:    "Event processing duration distribution",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
	})
	FunnelAnalysesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_funnel_analyses_total", Help: "Total funnel analyses performed",
	})
	CohortAnalysesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_analytics_cohort_analyses_total", Help: "Total cohort analyses performed",
	})
	ActiveSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_analytics_active_sessions", Help: "Number of active sessions",
	})
)
