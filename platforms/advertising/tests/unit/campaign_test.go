package unit

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
)

func TestCampaignCRUD(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)

	ctx := context.Background()

	cm := &campaign.Campaign{
		Name:      "Test Campaign",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.50,
		Budget:    campaign.Budget{Daily: 100, Lifetime: 1000},
		DateRange: campaign.DateRange{Start: time.Now(), End: time.Now().Add(30 * 24 * time.Hour)},
	}

	created, err := svc.Create(ctx, cm)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatal("Expected non-empty ID")
	}
	if created.Status != campaign.CampaignStatusDraft {
		t.Errorf("Expected draft status, got %s", created.Status)
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "Test Campaign" {
		t.Errorf("Expected 'Test Campaign', got '%s'", got.Name)
	}

	camps, err := svc.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(camps) != 1 {
		t.Errorf("Expected 1 campaign, got %d", len(camps))
	}

	created.Name = "Updated Campaign"
	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated Campaign" {
		t.Errorf("Expected 'Updated Campaign', got '%s'", updated.Name)
	}
}

func TestCampaignStatusTransitions(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)

	ctx := context.Background()

	cm := &campaign.Campaign{
		Name:      "Status Test",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.00,
		Budget:    campaign.Budget{Daily: 50, Lifetime: 500},
	}

	created, _ := svc.Create(ctx, cm)

	paused, err := svc.Pause(ctx, created.ID)
	if err == nil {
		t.Errorf("Expected error pausing draft campaign, got status %s", paused.Status)
	}

	created.Status = campaign.CampaignStatusActive
	svc.Update(ctx, created)

	paused, err = svc.Pause(ctx, created.ID)
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	if paused.Status != campaign.CampaignStatusPaused {
		t.Errorf("Expected paused, got %s", paused.Status)
	}

	resumed, err := svc.Resume(ctx, created.ID)
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if resumed.Status != campaign.CampaignStatusActive {
		t.Errorf("Expected active, got %s", resumed.Status)
	}

	ended, err := svc.End(ctx, created.ID)
	if err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if ended.Status != campaign.CampaignStatusEnded {
		t.Errorf("Expected ended, got %s", ended.Status)
	}

	_, err = svc.Resume(ctx, created.ID)
	if err == nil {
		t.Error("Expected error resuming ended campaign")
	}
}

func TestCampaignValidation(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)
	ctx := context.Background()

	_, err := svc.Create(ctx, &campaign.Campaign{Type: campaign.CampaignTypeCPC})
	if err != campaign.ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}

	_, err = svc.Create(ctx, &campaign.Campaign{
		Name: "Bad Type",
		Type: "INVALID",
	})
	if err != campaign.ErrInvalidType {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}

	_, err = svc.Create(ctx, &campaign.Campaign{
		Name:      "Bad Budget",
		Type:      campaign.CampaignTypeCPC,
		Budget:    campaign.Budget{Daily: -1, Lifetime: 10},
	})
	if err != campaign.ErrInvalidBudget {
		t.Errorf("Expected ErrInvalidBudget, got %v", err)
	}

	_, err = svc.Create(ctx, &campaign.Campaign{
		Name:      "Bad Date Range",
		Type:      campaign.CampaignTypeCPC,
		Budget:    campaign.Budget{Daily: 10, Lifetime: 100},
		DateRange: campaign.DateRange{Start: time.Now().Add(24 * time.Hour), End: time.Now()},
	})
	if err != campaign.ErrInvalidDateRange {
		t.Errorf("Expected ErrInvalidDateRange, got %v", err)
	}
}

func TestCampaignNotFound(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)
	ctx := context.Background()

	_, err := svc.Get(ctx, "nonexistent")
	if err != campaign.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestCampaignListByStatus(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		svc.Create(ctx, &campaign.Campaign{
			Name:      "Draft Campaign",
			Type:      campaign.CampaignTypeCPM,
			BidAmount: 2.00,
			Budget:    campaign.Budget{Daily: 200, Lifetime: 2000},
		})
	}

	cm := &campaign.Campaign{
		Name:      "Active Campaign",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.00,
		Budget:    campaign.Budget{Daily: 50, Lifetime: 500},
	}
	created, _ := svc.Create(ctx, cm)
	created.Status = campaign.CampaignStatusActive
	svc.Update(ctx, created)

	drafts, _ := svc.List(ctx, campaign.CampaignStatusDraft)
	if len(drafts) != 3 {
		t.Errorf("Expected 3 draft campaigns, got %d", len(drafts))
	}

	active, _ := svc.List(ctx, campaign.CampaignStatusActive)
	if len(active) != 1 {
		t.Errorf("Expected 1 active campaign, got %d", len(active))
	}
}
