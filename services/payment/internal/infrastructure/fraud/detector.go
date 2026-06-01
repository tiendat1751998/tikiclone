package fraud

import (
	"context"
	"encoding/json"

	"github.com/tikiclone/tiki/services/payment/internal/domain"
)

type DetectorConfig struct {
	RiskThreshold int
}

type Detector struct {
	cfg DetectorConfig
}

func NewDetector(cfg DetectorConfig) *Detector {
	return &Detector{cfg: cfg}
}

func (d *Detector) Assess(ctx context.Context, paymentID, userID string, amount int64, method domain.PaymentMethod) (*domain.FraudCheckResult, error) {
	riskScore := 0
	reasons := make([]string, 0)

	if amount > 10000000 {
		riskScore += 30
		reasons = append(reasons, "high_amount")
	}
	if method == domain.PaymentMethodCOD {
		riskScore += 15
		reasons = append(reasons, "high_risk_method")
	}
	if method == domain.PaymentMethodBankTransfer {
		riskScore += 5
		reasons = append(reasons, "bank_transfer_risk")
	}

	isFraud := riskScore >= d.cfg.RiskThreshold

	result := domain.NewFraudCheckResult(paymentID, userID, riskScore, isFraud)
	reasonsBytes, _ := json.Marshal(reasons)
	result.Reasons = reasonsBytes
	return result, nil
}
