package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/analytics/internal/analytics"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

func setupAnalyticsTest(t *testing.T) (*analytics.Service, *events.Service) {
	t.Helper()
	eventRepo := events.NewInMemoryRepository()
	eventSvc := events.NewService(eventRepo)
	analyticsRepo := analytics.NewInMemoryRepository()
	analyticsSvc := analytics.NewService(analyticsRepo, eventSvc)
	return analyticsSvc, eventSvc
}

func seedEvents(t *testing.T, svc *events.Service) {
	t.Helper()
	now := time.Now()
	events_list := []events.AnalyticsEvent{
		{EventID: "e1", EventType: events.EventPageview, UserID: "u1", SessionID: "s1", Timestamp: now, Source: "web", Device: "desktop", Country: "US"},
		{EventID: "e2", EventType: events.EventPurchase, UserID: "u1", SessionID: "s1", Timestamp: now, Revenue: 100, Source: "web", Device: "desktop", Country: "US"},
		{EventID: "e3", EventType: events.EventPageview, UserID: "u2", SessionID: "s2", Timestamp: now, Source: "mobile", Device: "mobile", Country: "ID"},
		{EventID: "e4", EventType: events.EventAddToCart, UserID: "u2", SessionID: "s2", Timestamp: now, Source: "mobile", Device: "mobile", Country: "ID"},
		{EventID: "e5", EventType: events.EventPurchase, UserID: "u2", SessionID: "s2", Timestamp: now, Revenue: 50, Source: "mobile", Device: "mobile", Country: "ID"},
		{EventID: "e6", EventType: events.EventSearch, UserID: "u1", SessionID: "s3", Timestamp: now, Source: "web", Device: "desktop", Country: "US"},
		{EventID: "e7", EventType: events.EventLogin, UserID: "u3", SessionID: "s4", Timestamp: now, Source: "web", Device: "tablet", Country: "SG"},
	}
	for i := range events_list {
		svc.IngestEvent(context.Background(), &events_list[i])
	}
}

func TestAnalyticsQuery(t *testing.T) {
	svc, eventSvc := setupAnalyticsTest(t)
	seedEvents(t, eventSvc)

	query := &analytics.AnalyticsQuery{
		Metrics: []analytics.Metric{
			{Name: analytics.MetricTotalUsers, Aggregation: analytics.AggSum, Alias: "total_users"},
			{Name: analytics.MetricRevenue, Aggregation: analytics.AggSum, Alias: "revenue"},
			{Name: analytics.MetricOrders, Aggregation: analytics.AggSum, Alias: "orders"},
		},
		TimeRange: analytics.TimeRange{Type: analytics.TimeToday},
	}

	result, err := svc.RunQuery(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected at least 1 row")
	}
	row := result.Rows[0]
	if row["total_users"] != float64(3) {
		t.Errorf("expected 3 total users, got %v", row["total_users"])
	}
	if row["revenue"] != float64(150) {
		t.Errorf("expected 150 revenue, got %v", row["revenue"])
	}
	if row["orders"] != float64(2) {
		t.Errorf("expected 2 orders, got %v", row["orders"])
	}
}

func TestAnalyticsQueryWithGroupBy(t *testing.T) {
	svc, eventSvc := setupAnalyticsTest(t)
	seedEvents(t, eventSvc)

	query := &analytics.AnalyticsQuery{
		Metrics: []analytics.Metric{
			{Name: analytics.MetricRevenue, Aggregation: analytics.AggSum, Alias: "revenue"},
			{Name: analytics.MetricOrders, Aggregation: analytics.AggSum, Alias: "orders"},
		},
		TimeRange: analytics.TimeRange{Type: analytics.TimeToday},
		GroupBy:   []string{"source"},
	}

	result, err := svc.RunQuery(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected grouped rows")
	}
	for _, row := range result.Rows {
		source, ok := row["source"].(string)
		if !ok || source == "" {
			t.Errorf("expected non-empty source, got %v", row["source"])
		}
	}
}

func TestAnalyticsQueryEmptyData(t *testing.T) {
	svc, _ := setupAnalyticsTest(t)

	query := &analytics.AnalyticsQuery{
		Metrics: []analytics.Metric{
			{Name: analytics.MetricTotalUsers, Aggregation: analytics.AggSum},
		},
		TimeRange: analytics.TimeRange{Type: analytics.TimeToday},
	}

	result, err := svc.RunQuery(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected at least 1 row")
	}
}

func TestAnalyticsQueryInvalid(t *testing.T) {
	svc, _ := setupAnalyticsTest(t)
	_, err := svc.RunQuery(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil query")
	}
}

func TestAnalyticsQueryNoMetrics(t *testing.T) {
	svc, _ := setupAnalyticsTest(t)
	_, err := svc.RunQuery(context.Background(), &analytics.AnalyticsQuery{
		TimeRange: analytics.TimeRange{Type: analytics.TimeToday},
	})
	if err == nil {
		t.Fatal("expected error for empty metrics")
	}
}

func TestGetKeyMetrics(t *testing.T) {
	svc, eventSvc := setupAnalyticsTest(t)
	seedEvents(t, eventSvc)

	metrics, err := svc.GetKeyMetrics(context.Background(), analytics.TimeRange{Type: analytics.TimeToday})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metrics) == 0 {
		t.Fatal("expected metrics")
	}
	metricMap := make(map[analytics.MetricType]float64)
	for _, m := range metrics {
		metricMap[m.Name] = m.Value
	}
	if metricMap[analytics.MetricTotalUsers] != 3 {
		t.Errorf("expected 3 total users, got %f", metricMap[analytics.MetricTotalUsers])
	}
	if metricMap[analytics.MetricRevenue] != 150 {
		t.Errorf("expected 150 revenue, got %f", metricMap[analytics.MetricRevenue])
	}
}
