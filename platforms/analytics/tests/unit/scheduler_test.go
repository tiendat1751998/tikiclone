package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/report_scheduler"
)

func TestScheduleReport(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	r, err := svc.ScheduleReport(context.Background(), "Weekly Report", "Weekly analytics report",
		map[string]interface{}{"metrics": []string{"revenue", "orders"}},
		report_scheduler.FreqWeekly, report_scheduler.ChannelEmail,
		[]string{"admin@tiki.com"}, "", "csv", "UTC", "user-1", "org-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "Weekly Report" {
		t.Errorf("expected Weekly Report, got %s", r.Name)
	}
	if !r.IsActive {
		t.Error("expected active report")
	}
	if r.Frequency != report_scheduler.FreqWeekly {
		t.Errorf("expected weekly, got %s", r.Frequency)
	}
}

func TestReportGeneration(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	r, _ := svc.ScheduleReport(context.Background(), "Daily Report", "", map[string]interface{}{},
		report_scheduler.FreqDaily, report_scheduler.ChannelDownload, nil, "", "csv", "UTC", "user-1", "org-1")

	gen, err := svc.GenerateReport(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen.Status != "completed" {
		t.Errorf("expected completed, got %s", gen.Status)
	}

	err = svc.DeliverReport(context.Background(), gen.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	storedGen, _ := repo.GetGeneration(context.Background(), gen.ID)
	if storedGen.Status != "delivered" {
		t.Errorf("expected delivered, got %s", storedGen.Status)
	}
}

func TestReportDeliveryFailure(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	r, _ := svc.ScheduleReport(context.Background(), "Fail Report", "", map[string]interface{}{},
		report_scheduler.FreqDaily, report_scheduler.ChannelWebhook, nil, "http://example.com/hook", "csv", "UTC", "user-1", "org-1")

	gen, _ := svc.GenerateReport(context.Background(), r.ID)

	deliveryErr := fmt.Errorf("connection timeout")
	err := svc.DeliverReport(context.Background(), gen.ID, deliveryErr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	storedGen, _ := repo.GetGeneration(context.Background(), gen.ID)
	if storedGen.Status != "delivery_failed" {
		t.Errorf("expected delivery_failed, got %s", storedGen.Status)
	}
}

func TestScheduleNotFound(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	_, err := svc.GetReport(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent report")
	}
}

func TestScheduleListEmpty(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	list, total, err := svc.ListReports(context.Background(), "", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 total, got %d", total)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestNextRunCalculation(t *testing.T) {
	repo := report_scheduler.NewInMemoryRepository()
	svc := report_scheduler.NewService(repo)

	daily, _ := svc.ScheduleReport(context.Background(), "Daily", "", map[string]interface{}{},
		report_scheduler.FreqDaily, report_scheduler.ChannelDownload, nil, "", "csv", "UTC", "user-1", "org-1")

	if daily.NextRunAt.Before(time.Now()) {
		t.Error("next run should be in the future")
	}

	monthly, _ := svc.ScheduleReport(context.Background(), "Monthly", "", map[string]interface{}{},
		report_scheduler.FreqMonthly, report_scheduler.ChannelDownload, nil, "", "csv", "UTC", "user-1", "org-1")

	if monthly.NextRunAt.Before(time.Now()) {
		t.Error("monthly next run should be in the future")
	}
}
