package streaming

import (
	"context"
	"math"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ProcessEvent(ctx context.Context, event *core.FraudEvent) {
	now := time.Now()

	if event.UserID != "" {
		s.repo.AddEvent(ctx, "user", event.UserID, now)
	}
	if event.IP != "" {
		s.repo.AddEvent(ctx, "ip", event.IP, now)
	}
	if event.DeviceID != "" {
		s.repo.AddEvent(ctx, "device", event.DeviceID, now)
	}
}

func (s *Service) CountInWindow(ctx context.Context, entityType, entityID string, window time.Duration) int {
	since := time.Now().Add(-window)
	return s.repo.CountSince(ctx, entityType, entityID, since)
}

func (s *Service) DetectBurst(ctx context.Context, entityType, entityID string) *EventWindow {
	now := time.Now()
	window5Min := now.Add(-5 * time.Minute)
	window1Hour := now.Add(-1 * time.Hour)

	recentCount := s.repo.CountSince(ctx, entityType, entityID, window5Min)
	historicalCount := s.repo.CountSince(ctx, entityType, entityID, window1Hour)

	historicalAvg := float64(historicalCount-recentCount) / 55.0 * 5.0
	if historicalAvg <= 0 {
		historicalAvg = 1
	}

	ratio := float64(recentCount) / historicalAvg
	isBurst := ratio >= 3.0 && recentCount > 5

	return &EventWindow{
		WindowSize: 5 * time.Minute,
		Count:      recentCount,
		Threshold:  int(math.Ceil(3 * historicalAvg)),
		IsBurst:    isBurst,
		AvgCount:   math.Round(historicalAvg*100) / 100,
	}
}

func (s *Service) GetAggregated(ctx context.Context, entityType, entityID string) *AggregatedCount {
	now := time.Now()
	return &AggregatedCount{
		EntityID:    entityID,
		EntityType:  entityType,
		Window1Min:  s.repo.CountSince(ctx, entityType, entityID, now.Add(-1*time.Minute)),
		Window5Min:  s.repo.CountSince(ctx, entityType, entityID, now.Add(-5*time.Minute)),
		Window1Hour: s.repo.CountSince(ctx, entityType, entityID, now.Add(-1*time.Hour)),
		LastUpdated: now,
	}
}
