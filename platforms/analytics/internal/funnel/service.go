package funnel

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

type Service struct {
	repo     Repository
	eventSvc *events.Service
}

func NewService(repo Repository, eventSvc *events.Service) *Service {
	return &Service{repo: repo, eventSvc: eventSvc}
}

func (s *Service) BuildFunnel(ctx context.Context, definition *FunnelDefinition) (*FunnelResult, error) {
	if len(definition.Steps) < 2 {
		return nil, ErrFunnelInvalid
	}

	if definition.ID == "" {
		definition.ID = uuid.New().String()
	}
	if definition.CreatedAt.IsZero() {
		definition.CreatedAt = time.Now()
	}
	s.repo.StoreDefinition(ctx, definition)

	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()

	var stepResults []FunnelStepResult
	var previousUserCount int64
	var startCount int64

	for i, step := range definition.Steps {
		eventsInStep, _, err := s.eventSvc.ListEvents(ctx, events.EventType(step.EventType), startTime, endTime, 0, 100000)
		if err != nil {
			return nil, err
		}

		usersInStep := make(map[string]bool)
		for _, e := range eventsInStep {
			usersInStep[e.UserID] = true
		}

		userCount := int64(len(usersInStep))
		if i == 0 {
			startCount = userCount
			previousUserCount = userCount
		}

		var stepRate float64
		if previousUserCount > 0 {
			stepRate = math.Round(float64(userCount)/float64(previousUserCount)*10000) / 100
		}

		var overallRate float64
		if startCount > 0 {
			overallRate = math.Round(float64(userCount)/float64(startCount)*10000) / 100
		}

		dropOff := previousUserCount - userCount
		var dropOffRate float64
		if previousUserCount > 0 {
			dropOffRate = math.Round(float64(dropOff)/float64(previousUserCount)*10000) / 100
		}

		stepResults = append(stepResults, FunnelStepResult{
			StepName:    step.Name,
			EventType:   step.EventType,
			Order:       step.Order,
			UserCount:   userCount,
			StepRate:    stepRate,
			OverallRate: overallRate,
			DropOff:     dropOff,
			DropOffRate: dropOffRate,
		})

		previousUserCount = userCount
	}

	var overallRate float64
	if startCount > 0 && len(stepResults) > 0 {
		last := stepResults[len(stepResults)-1]
		overallRate = math.Round(float64(last.UserCount)/float64(startCount)*10000) / 100
	}

	result := &FunnelResult{
		ID:          uuid.New().String(),
		FunnelName:  definition.Name,
		Steps:       stepResults,
		OverallRate: overallRate,
		StartCount:  startCount,
		EndCount:    previousUserCount,
		AnalyzedAt:  time.Now(),
	}

	s.repo.StoreResult(ctx, result)
	return result, nil
}

func (s *Service) CalculateConversion(ctx context.Context, fromStep, toStep string, funnelID string) (*ConversionRate, error) {
	result, err := s.repo.GetResult(ctx, funnelID)
	if err != nil || result == nil {
		return nil, ErrFunnelNotFound
	}

	var fromCount, toCount int64
	for _, step := range result.Steps {
		if step.EventType == fromStep {
			fromCount = step.UserCount
		}
		if step.EventType == toStep {
			toCount = step.UserCount
		}
	}

	var rate float64
	if fromCount > 0 {
		rate = math.Round(float64(toCount)/float64(fromCount)*10000) / 100
	}

	return &ConversionRate{
		StepFrom:  fromStep,
		StepTo:    toStep,
		Rate:      rate,
		FromCount: fromCount,
		ToCount:   toCount,
	}, nil
}

func (s *Service) AnalyzeDropoff(ctx context.Context, funnelID string) ([]FunnelStepResult, error) {
	result, err := s.repo.GetResult(ctx, funnelID)
	if err != nil || result == nil {
		return nil, ErrFunnelNotFound
	}

	maxDropOff := int64(0)
	for _, step := range result.Steps {
		if step.DropOff > maxDropOff {
			maxDropOff = step.DropOff
		}
	}

	return result.Steps, nil
}
