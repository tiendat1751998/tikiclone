package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RulesEvaluatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_rules_evaluated_total", Help: "Total rules evaluated",
	})
	AlertsGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_alerts_generated_total", Help: "Total fraud alerts generated",
	})
	AlertsResolvedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_alerts_resolved_total", Help: "Total fraud alerts resolved",
	})
	FalsePositiveRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_fraud_false_positive_rate", Help: "False positive rate",
	})
	AverageRiskScore = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_fraud_average_risk_score", Help: "Average risk score",
	})
	BlacklistHitsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_blacklist_hits_total", Help: "Total blacklist hits",
	})
	VerificationInitiatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_verification_initiated_total", Help: "Total verifications initiated",
	})
	VerificationSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_verification_success_total", Help: "Total successful verifications",
	})
	CasesOpenedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_cases_opened_total", Help: "Total fraud cases opened",
	})
	ScoreHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_fraud_risk_score",
		Help:    "Distribution of risk scores",
		Buckets: []float64{10, 25, 50, 75, 90, 100},
	})

	FraudScoresComputed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_scores_computed_total", Help: "Total fraud scores computed",
	})

	FraudCasesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_fraud_cases_created_total", Help: "Total fraud cases created",
	})
)
