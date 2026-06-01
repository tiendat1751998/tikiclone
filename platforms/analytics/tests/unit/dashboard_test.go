package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/analytics/internal/dashboard"
)

func TestDashboardCRUD(t *testing.T) {
	repo := dashboard.NewInMemoryRepository()
	svc := dashboard.NewService(repo)

	d, err := svc.CreateDashboard(context.Background(), "Test Dashboard", "A test dashboard", "org-1", "user-1", false, []string{"test", "analytics"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Title != "Test Dashboard" {
		t.Errorf("expected Test Dashboard, got %s", d.Title)
	}
	if d.ID == "" {
		t.Error("expected non-empty ID")
	}

	retrieved, err := svc.GetDashboard(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Title != d.Title {
		t.Errorf("title mismatch: %s vs %s", retrieved.Title, d.Title)
	}

	updated, err := svc.UpdateDashboard(context.Background(), d.ID, "Updated Dashboard", "Updated description", true, []string{"updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "Updated Dashboard" {
		t.Errorf("expected Updated Dashboard, got %s", updated.Title)
	}
	if !updated.IsPublic {
		t.Error("expected dashboard to be public")
	}

	list, total, err := svc.ListDashboards(context.Background(), "", 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 dashboard, got %d", total)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 in list, got %d", len(list))
	}
}

func TestDashboardWidget(t *testing.T) {
	repo := dashboard.NewInMemoryRepository()
	svc := dashboard.NewService(repo)

	d, _ := svc.CreateDashboard(context.Background(), "Widget Test", "", "org-1", "user-1", false, nil)

	w, err := svc.AddWidget(context.Background(), d.ID, "Revenue Chart", "line", 6, 4, 0, 0, dashboard.DataSource{Type: "metric", Metric: "revenue"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Type != dashboard.ChartLine {
		t.Errorf("expected line chart, got %s", w.Type)
	}
	if w.DashboardID != d.ID {
		t.Errorf("dashboard ID mismatch")
	}
}

func TestDashboardNotFound(t *testing.T) {
	repo := dashboard.NewInMemoryRepository()
	svc := dashboard.NewService(repo)

	_, err := svc.GetDashboard(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent dashboard")
	}
}

func TestDashboardListEmpty(t *testing.T) {
	repo := dashboard.NewInMemoryRepository()
	svc := dashboard.NewService(repo)

	list, total, err := svc.ListDashboards(context.Background(), "", 0, 10)
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
