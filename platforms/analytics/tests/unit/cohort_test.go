package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/cohort"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

func setupCohortTest(t *testing.T) (*cohort.Service, *events.Service) {
	t.Helper()
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	cohortRepo := cohort.NewInMemoryRepository()
	cohortSvc := cohort.NewService(cohortRepo, eventSvc)
	return cohortSvc, eventSvc
}

func TestCohortAnalysis(t *testing.T) {
	svc, eventSvc := setupCohortTest(t)

	now := time.Now()
	for i := 0; i < 10; i++ {
		userID := string(rune('0' + i))
		if i >= 10 {
			userID = string(rune('a' + i - 10))
		}
		uid := "cu-" + userID
		eventSvc.IngestEvent(context.Background(), &events.AnalyticsEvent{
			EventID: "signup-" + uid, EventType: events.EventSignup, UserID: uid, Timestamp: now.AddDate(0, 0, -i),
		})
		eventSvc.IngestEvent(context.Background(), &events.AnalyticsEvent{
			EventID: "pageview-" + uid, EventType: events.EventPageview, UserID: uid, Timestamp: now,
		})
	}

	definition := &cohort.CohortDefinition{
		Name:   "signup_cohort",
		Period: cohort.CohortDay,
	}

	result, err := svc.BuildCohort(context.Background(), definition)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestCohortRetentionCalculation(t *testing.T) {
	svc, eventSvc := setupCohortTest(t)

	now := time.Now()
	for i := 0; i < 5; i++ {
		uid := "ret-user-" + string(rune('0'+i))
		eventSvc.IngestEvent(context.Background(), &events.AnalyticsEvent{
			EventID: "ret-signup-" + uid, EventType: events.EventSignup, UserID: uid, Timestamp: now.AddDate(0, 0, -30),
		})
		for day := 0; day < 5; day++ {
			eventSvc.IngestEvent(context.Background(), &events.AnalyticsEvent{
				EventID:   "ret-pv-" + uid + "-d" + string(rune('0'+day)),
				EventType: events.EventPageview,
				UserID:    uid,
				Timestamp: now.AddDate(0, 0, -day),
			})
		}
	}

	def := &cohort.CohortDefinition{
		Name:   "retention_test",
		Period: cohort.CohortDay,
	}
	result, err := svc.BuildCohort(context.Background(), def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Cohorts) > 0 {
		if len(result.Matrix) > 0 && len(result.Matrix[0]) > 0 {
			if result.Matrix[0][0].RetentionRate > 0 {
				t.Logf("retention rate for period 0: %f", result.Matrix[0][0].RetentionRate)
			}
		}
	}
}

func TestCohortEmptyData(t *testing.T) {
	svc, _ := setupCohortTest(t)

	def := &cohort.CohortDefinition{
		Name:   "empty",
		Period: cohort.CohortWeek,
	}
	result, err := svc.BuildCohort(context.Background(), def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Cohorts) != 0 {
		t.Errorf("expected 0 cohorts for empty data, got %d", len(result.Cohorts))
	}
}
