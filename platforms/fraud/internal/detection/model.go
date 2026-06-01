package detection

import (
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
)

type AlertType string

const (
	AlertNewDeviceLogin       AlertType = "NEW_DEVICE_LOGIN"
	AlertHighValueTransaction AlertType = "HIGH_VALUE_TRANSACTION"
	AlertRapidFireOrders      AlertType = "RAPID_FIRE_ORDERS"
	AlertAccountTakeover      AlertType = "ACCOUNT_TAKEOVER"
	AlertPaymentFraud         AlertType = "PAYMENT_FRAUD"
)

type RiskScore struct {
	Score          float64      `json:"score"`
	Level          core.RiskLevel `json:"level"`
	MaxScore       float64      `json:"max_score"`
	RuleResults    []RuleResult `json:"rule_results"`
	EvaluatedAt    time.Time    `json:"evaluated_at"`
	AlertTriggered bool         `json:"alert_triggered"`
	AlertID        string       `json:"alert_id,omitempty"`
}

type RuleResult struct {
	RuleName  string  `json:"rule_name"`
	Severity  int     `json:"severity"`
	Weight    float64 `json:"weight"`
	Score     float64 `json:"score"`
	Triggered bool    `json:"triggered"`
	Reason    string  `json:"reason,omitempty"`
}

type FraudAlert struct {
	ID          string        `json:"id"`
	EventID     string        `json:"event_id"`
	UserID      string        `json:"user_id"`
	Type        AlertType     `json:"type"`
	RiskScore   float64       `json:"risk_score"`
	RiskLevel   core.RiskLevel `json:"risk_level"`
	Description string        `json:"description"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
	ResolvedBy  string        `json:"resolved_by,omitempty"`
	Resolution  string        `json:"resolution,omitempty"`
}

var _ = core.FraudEvent{}
