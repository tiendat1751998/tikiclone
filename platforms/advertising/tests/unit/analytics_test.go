package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/advertising/internal/analytics"
)

func TestRecordImpression(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	imp := &analytics.Impression{
		CampaignID: "camp-1",
		CreativeID: "cre-1",
		UserID:     "user1",
		Cost:       0.05,
		Device:     "mobile",
		Location:   "US",
	}

	err := svc.RecordImpression(ctx, imp)
	if err != nil {
		t.Fatalf("RecordImpression failed: %v", err)
	}
	if imp.ID == "" {
		t.Fatal("Expected non-empty impression ID")
	}
	if imp.Timestamp.IsZero() {
		t.Fatal("Expected non-zero timestamp")
	}
}

func TestRecordClick(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	click := &analytics.Click{
		ImpressionID: "imp-1",
		CampaignID:   "camp-1",
		CreativeID:   "cre-1",
		UserID:       "user1",
		Cost:         0.10,
	}

	err := svc.RecordClick(ctx, click)
	if err != nil {
		t.Fatalf("RecordClick failed: %v", err)
	}
	if click.ID == "" {
		t.Fatal("Expected non-empty click ID")
	}
}

func TestRecordConversion(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	conv := &analytics.Conversion{
		ClickID:        "click-1",
		CampaignID:     "camp-1",
		CreativeID:     "cre-1",
		UserID:         "user1",
		Revenue:        29.99,
		ConversionType: "purchase",
	}

	err := svc.RecordConversion(ctx, conv)
	if err != nil {
		t.Fatalf("RecordConversion failed: %v", err)
	}
	if conv.ID == "" {
		t.Fatal("Expected non-empty conversion ID")
	}
}

func TestAnalyticsReportCalculation(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-1", Cost: 0.05, Timestamp: time.Now()})
	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-1", Cost: 0.05, Timestamp: time.Now()})
	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-1", Cost: 0.05, Timestamp: time.Now()})
	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-1", Cost: 0.05, Timestamp: time.Now()})

	svc.RecordClick(ctx, &analytics.Click{CampaignID: "camp-1", Cost: 0.10, Timestamp: time.Now()})
	svc.RecordClick(ctx, &analytics.Click{CampaignID: "camp-1", Cost: 0.10, Timestamp: time.Now()})

	svc.RecordConversion(ctx, &analytics.Conversion{CampaignID: "camp-1", Revenue: 50.00, Timestamp: time.Now()})

	report, err := svc.GetReport(ctx, &analytics.ReportFilter{CampaignID: "camp-1"})
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	if report.Impressions != 4 {
		t.Errorf("Expected 4 impressions, got %d", report.Impressions)
	}
	if report.Clicks != 2 {
		t.Errorf("Expected 2 clicks, got %d", report.Clicks)
	}
	if report.Conversions != 1 {
		t.Errorf("Expected 1 conversion, got %d", report.Conversions)
	}
	if report.Spend != 0.40 {
		t.Errorf("Expected spend 0.40, got %f", report.Spend)
	}
	if report.Revenue != 50.00 {
		t.Errorf("Expected revenue 50.00, got %f", report.Revenue)
	}
	if report.CTR != 50.00 {
		t.Errorf("Expected CTR 50.00, got %f", report.CTR)
	}
	if report.CVR != 50.00 {
		t.Errorf("Expected CVR 50.00, got %f", report.CVR)
	}
	if report.CPC != 0.20 {
		t.Errorf("Expected CPC 0.20, got %f", report.CPC)
	}
	if report.CPM != 100.00 {
		t.Errorf("Expected CPM 100.00, got %f", report.CPM)
	}
	if report.ROAS != 125.00 {
		t.Errorf("Expected ROAS 125.00, got %f", report.ROAS)
	}
}

func TestAnalyticsReportByDateRange(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	svc.RecordImpression(ctx, &analytics.Impression{
		CampaignID: "camp-1", Cost: 0.05,
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	})
	svc.RecordImpression(ctx, &analytics.Impression{
		CampaignID: "camp-1", Cost: 0.05,
		Timestamp: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
	})
	svc.RecordImpression(ctx, &analytics.Impression{
		CampaignID: "camp-1", Cost: 0.05,
		Timestamp: time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
	})

	report, err := svc.GetReport(ctx, &analytics.ReportFilter{
		CampaignID: "camp-1",
		StartDate:  "2025-01-01",
		EndDate:    "2025-01-31",
	})
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}
	if report.Impressions != 2 {
		t.Errorf("Expected 2 impressions in date range, got %d", report.Impressions)
	}
}

func TestAnalyticsReportNoData(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	report, err := svc.GetReport(ctx, &analytics.ReportFilter{CampaignID: "nonexistent"})
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}
	if report.Impressions != 0 {
		t.Errorf("Expected 0 impressions for no data, got %d", report.Impressions)
	}
	if report.Clicks != 0 {
		t.Errorf("Expected 0 clicks for no data, got %d", report.Clicks)
	}
}

func TestAnalyticsReportMultiCampaign(t *testing.T) {
	repo := analytics.NewInMemoryRepository()
	svc := analytics.NewService(repo)
	ctx := context.Background()

	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-1", Cost: 0.05, Timestamp: time.Now()})
	svc.RecordImpression(ctx, &analytics.Impression{CampaignID: "camp-2", Cost: 0.10, Timestamp: time.Now()})
	svc.RecordClick(ctx, &analytics.Click{CampaignID: "camp-1", Cost: 0.10, Timestamp: time.Now()})

	report1, _ := svc.GetReport(ctx, &analytics.ReportFilter{CampaignID: "camp-1"})
	if report1.Impressions != 1 || report1.Clicks != 1 {
		t.Errorf("camp-1: expected 1 impression, 1 click; got %d, %d", report1.Impressions, report1.Clicks)
	}

	report2, _ := svc.GetReport(ctx, &analytics.ReportFilter{CampaignID: "camp-2"})
	if report2.Impressions != 1 || report2.Clicks != 0 {
		t.Errorf("camp-2: expected 1 impression, 0 clicks; got %d, %d", report2.Impressions, report2.Clicks)
	}
}
