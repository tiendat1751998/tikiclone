package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LoginsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_logins_total",
		Help: "Total number of successful logins",
	})

	FailedLogins = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_auth_failed_logins_total",
		Help: "Total number of failed login attempts",
	}, []string{"reason"})

	RegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_registrations_total",
		Help: "Total number of user registrations",
	})

	RegistrationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_auth_registration_errors_total",
		Help: "Total number of registration errors",
	}, []string{"reason"})

	TokenRefreshesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_token_refreshes_total",
		Help: "Total number of token refresh operations",
	})

	TokenRefreshErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_auth_token_refresh_errors_total",
		Help: "Total number of token refresh errors",
	}, []string{"reason"})

	AccountLockouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_account_lockouts_total",
		Help: "Total number of account lockouts",
	})

	SuspiciousLogins = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_suspicious_logins_total",
		Help: "Total number of suspicious login detections",
	})

	ActiveSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_auth_active_sessions",
		Help: "Current number of active sessions",
	})

	TokenValidationDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_auth_token_validation_duration_seconds",
		Help:    "Token validation latency",
		Buckets: prometheus.DefBuckets,
	})

	PasswordHashDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_auth_password_hash_duration_seconds",
		Help:    "Password hashing latency",
		Buckets: prometheus.DefBuckets,
	})

	RedisOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_auth_redis_operation_duration_seconds",
		Help:    "Redis operation latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	DatabaseQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tiki_auth_db_query_duration_seconds",
		Help:    "Database query latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"query"})

	PasswordResetsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_password_resets_total",
		Help: "Total number of password reset requests",
	})

	EmailVerificationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_email_verifications_total",
		Help: "Total number of email verification attempts",
	})

	AuditLogsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_auth_audit_logs_total",
		Help: "Total number of audit log entries",
	})

	RateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_auth_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	}, []string{"type"})
)
