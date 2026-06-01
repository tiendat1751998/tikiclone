package scoring

import "github.com/tikiclone/tiki/platforms/fraud/internal/core"

type ScoreCard struct {
	TotalScore float64          `json:"total_score"`
	Level      core.RiskLevel   `json:"level"`
	MaxScore   float64          `json:"max_score"`
	Factors    []WeightedFactor `json:"factors"`
}

type WeightedFactor struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
}

type ScoreThreshold struct {
	Low      float64 `json:"low"`
	Medium   float64 `json:"medium"`
	High     float64 `json:"high"`
	Critical float64 `json:"critical"`
}
