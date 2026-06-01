package riskscoring

import "github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"

type RiskFactor struct {
	Name         string  `json:"name"`
	Weight       float64 `json:"weight"`
	CurrentValue float64 `json:"current_value"`
	Contribution float64 `json:"contribution"`
}

type ScoreCard struct {
	Factors    []RiskFactor    `json:"factors"`
	BaseScore  float64         `json:"base_score"`
	TotalScore float64         `json:"total_score"`
	MaxScore   float64         `json:"max_score"`
	Level      core.RiskLevel  `json:"level"`
}

type RiskLevelThresholds struct {
	Safe     float64 `json:"safe"`
	Low      float64 `json:"low"`
	Medium   float64 `json:"medium"`
	High     float64 `json:"high"`
	Critical float64 `json:"critical"`
}
