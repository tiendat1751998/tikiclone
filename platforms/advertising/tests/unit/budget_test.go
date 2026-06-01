package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/advertising/internal/budget"
)

func TestCheckBudgetWithinLimits(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID:     "camp-1",
		DailyBudget:    100,
		LifetimeBudget: 1000,
		SpentToday:     0,
		LifetimeSpent:     0,
		LastResetDate:  time.Now().Format("2006-01-02"),
		IsActive:       true,
	})

	ok, err := svc.CheckBudget(ctx, "camp-1", 50)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if !ok {
		t.Error("Expected budget check to pass")
	}
}

func TestCheckBudgetExceedsDaily(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID:     "camp-1",
		DailyBudget:    100,
		LifetimeBudget: 1000,
		SpentToday:     80,
		LifetimeSpent:     200,
		LastResetDate:  time.Now().Format("2006-01-02"),
		IsActive:       true,
	})

	ok, err := svc.CheckBudget(ctx, "camp-1", 30)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if ok {
		t.Error("Expected budget check to fail (80+30 > 100 daily)")
	}

	ok, err = svc.CheckBudget(ctx, "camp-1", 20)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if !ok {
		t.Error("Expected budget check to pass (80+20 <= 100 daily)")
	}
}

func TestCheckBudgetExceedsLifetime(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID:     "camp-1",
		DailyBudget:    100,
		LifetimeBudget: 200,
		SpentToday:     10,
		LifetimeSpent:     190,
		LastResetDate:  time.Now().Format("2006-01-02"),
		IsActive:       true,
	})

	ok, err := svc.CheckBudget(ctx, "camp-1", 20)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if ok {
		t.Error("Expected budget check to fail (190+20 > 200 lifetime)")
	}

	ok, err = svc.CheckBudget(ctx, "camp-1", 10)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if !ok {
		t.Error("Expected budget check to pass (190+10 <= 200 lifetime)")
	}
}

func TestBudgetDeductAndRemaining(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")
	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID:     "camp-1",
		DailyBudget:    100,
		LifetimeBudget: 1000,
		SpentToday:     0,
		LifetimeSpent:     0,
		LastResetDate:  today,
		IsActive:       true,
	})

	svc.DeductSpend(ctx, "camp-1", 25.50)

	daily, lifetime, _ := svc.GetRemainingBudget(ctx, "camp-1")
	if daily != 74.50 {
		t.Errorf("Expected daily remaining 74.50, got %f", daily)
	}
	if lifetime != 974.50 {
		t.Errorf("Expected lifetime remaining 974.50, got %f", lifetime)
	}

	svc.DeductSpend(ctx, "camp-1", 74.50)
	daily, _, _ = svc.GetRemainingBudget(ctx, "camp-1")
	if daily != 0 {
		t.Errorf("Expected daily remaining 0, got %f", daily)
	}
}

func TestDailyBudgetReset(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID:     "camp-1",
		DailyBudget:    100,
		LifetimeBudget: 1000,
		SpentToday:     90,
		LifetimeSpent:     500,
		LastResetDate:  yesterday,
		IsActive:       true,
	})

	ok, err := svc.CheckBudget(ctx, "camp-1", 50)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if !ok {
		t.Error("Expected budget check to pass after daily reset")
	}

	tracker, _ := repo.GetTracker(ctx, "camp-1")
	if tracker.SpentToday != 0 {
		t.Errorf("Expected spent_today reset to 0, got %f", tracker.SpentToday)
	}
	if tracker.LastResetDate == yesterday {
		t.Error("Expected last_reset_date to be updated")
	}
}

func TestInactiveCampaignBudget(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	repo.CreateTracker(ctx, &budget.BudgetPlan{
		CampaignID: "camp-1",
		IsActive:   false,
	})

	ok, _ := svc.CheckBudget(ctx, "camp-1", 10)
	if ok {
		t.Error("Expected budget check to fail for inactive campaign")
	}
}

func TestBudgetDefaultValues(t *testing.T) {
	repo := budget.NewInMemoryRepository()
	svc := budget.NewService(repo)
	ctx := context.Background()

	ok, err := svc.CheckBudget(ctx, "new-campaign", 10)
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}
	if !ok {
		t.Error("Expected budget check to pass with default values")
	}
}
