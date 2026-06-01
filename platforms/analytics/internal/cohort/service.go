package cohort

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

func (s *Service) BuildCohort(ctx context.Context, definition *CohortDefinition) (*CohortAnalysis, error) {
	if definition.ID == "" {
		definition.ID = uuid.New().String()
	}
	if definition.CreatedAt.IsZero() {
		definition.CreatedAt = time.Now()
	}
	s.repo.StoreDefinition(ctx, definition)

	startTime := time.Now().AddDate(0, 0, -90)
	endTime := time.Now()

	signupEvents, _, err := s.eventSvc.ListEvents(ctx, events.EventSignup, startTime, endTime, 0, 100000)
	if err != nil {
		return nil, err
	}

	cohortUserMap := make(map[string][]string)
	for _, e := range signupEvents {
		dateKey := e.Timestamp.Format("2006-01-02")
		cohortUserMap[dateKey] = append(cohortUserMap[dateKey], e.UserID)
	}

	var periods []int
	switch definition.Period {
	case CohortDay:
		periods = []int{0, 1, 2, 3, 7, 14, 21, 30}
	case CohortWeek:
		periods = []int{0, 1, 2, 3, 4, 8, 12}
	default:
		periods = []int{0, 1, 2, 3, 6, 12}
	}

	var cohorts []CohortRow
	matrix := make([][]CohortCell, 0)

	for periodStart, userIDs := range cohortUserMap {
		startDate, _ := time.Parse("2006-01-02", periodStart)
		userCount := int64(len(userIDs))
		userSet := make(map[string]bool)
		for _, uid := range userIDs {
			userSet[uid] = true
		}

		row := CohortRow{
			ID:          uuid.New().String(),
			PeriodStart: periodStart,
			UserCount:   userCount,
		}

		var periodCells []CohortCell
		for _, offset := range periods {
			var periodEnd time.Time
			switch definition.Period {
			case CohortDay:
				periodEnd = startDate.AddDate(0, 0, offset)
			case CohortWeek:
				periodEnd = startDate.AddDate(0, 0, offset*7)
			default:
				periodEnd = startDate.AddDate(0, offset, 0)
			}

			retainedUsers := s.countRetainedUsers(ctx, userSet, startDate, periodEnd)
			var rate float64
			if userCount > 0 {
				rate = math.Round(float64(retainedUsers)/float64(userCount)*10000) / 100
			}

			periodCells = append(periodCells, CohortCell{
				UserCount:     retainedUsers,
				RetentionRate: rate,
			})
			row.Retention = append(row.Retention, rate)
		}

		cohorts = append(cohorts, row)
		matrix = append(matrix, periodCells)
	}

	return &CohortAnalysis{
		ID:          uuid.New().String(),
		Name:        definition.Name,
		Period:      definition.Period,
		PeriodLabel: string(definition.Period),
		Cohorts:     cohorts,
		Matrix:      matrix,
		Periods:     periods,
		AnalyzedAt:  time.Now(),
	}, nil
}

func (s *Service) countRetainedUsers(ctx context.Context, userSet map[string]bool, startDate, endDate time.Time) int64 {
	events_list, _, _ := s.eventSvc.ListEvents(ctx, "", startDate, endDate, 0, 100000)
	retained := make(map[string]bool)
	for _, e := range events_list {
		if userSet[e.UserID] {
			retained[e.UserID] = true
		}
	}
	return int64(len(retained))
}

func (s *Service) CalculateRetention(ctx context.Context, cohortID string, periodOffset int) (*RetentionPoint, error) {
	analysis, err := s.repo.GetAnalysis(ctx, cohortID)
	if err != nil || analysis == nil {
		return nil, ErrCohortNotFound
	}

	var totalUsers int64
	var totalRetained int64
	for _, cohort := range analysis.Cohorts {
		totalUsers += cohort.UserCount
		for i, r := range cohort.Retention {
			if i == periodOffset {
				totalRetained += int64(float64(cohort.UserCount) * r / 100)
			}
		}
	}

	var rate float64
	if totalUsers > 0 {
		rate = math.Round(float64(totalRetained)/float64(totalUsers)*10000) / 100
	}

	return &RetentionPoint{
		PeriodOffset: periodOffset,
		UserCount:    totalRetained,
		Rate:         rate,
	}, nil
}

func (s *Service) GetCohortReport(ctx context.Context, cohortID string) (*CohortAnalysis, error) {
	return s.repo.GetAnalysis(ctx, cohortID)
}
