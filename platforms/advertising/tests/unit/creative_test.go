package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/advertising/internal/creative"
)

func TestCreativeCRUD(t *testing.T) {
	repo := creative.NewInMemoryRepository()
	svc := creative.NewService(repo)
	ctx := context.Background()

	cr := &creative.Creative{
		CampaignID:     "camp-1",
		Name:           "Test Banner",
		Format:         creative.CreativeFormatBanner,
		Content:        "https://example.com/banner.jpg",
		DestinationURL: "https://example.com",
		Sizes: []creative.CreativeSize{
			{Width: 300, Height: 250, Label: "medium_rectangle"},
			{Width: 728, Height: 90, Label: "leaderboard"},
		},
	}

	created, err := svc.Create(ctx, cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatal("Expected non-empty ID")
	}
	if created.Status != creative.CreativeStatusDraft {
		t.Errorf("Expected draft status, got %s", created.Status)
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil || got == nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "Test Banner" {
		t.Errorf("Expected 'Test Banner', got '%s'", got.Name)
	}

	campaignCreatives, err := svc.GetByCampaign(ctx, "camp-1")
	if err != nil {
		t.Fatalf("GetByCampaign failed: %v", err)
	}
	if len(campaignCreatives) != 1 {
		t.Errorf("Expected 1 creative for camp-1, got %d", len(campaignCreatives))
	}
}

func TestCreativeApprovalWorkflow(t *testing.T) {
	repo := creative.NewInMemoryRepository()
	svc := creative.NewService(repo)
	ctx := context.Background()

	cr, _ := svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Approval Test",
		Format:     creative.CreativeFormatText,
	})

	approved, err := svc.Approve(ctx, cr.ID)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if approved.Status != creative.CreativeStatusApproved {
		t.Errorf("Expected approved status, got %s", approved.Status)
	}

	rejected, err := svc.Reject(ctx, cr.ID)
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if rejected.Status != creative.CreativeStatusRejected {
		t.Errorf("Expected rejected status, got %s", rejected.Status)
	}
}

func TestCreativeServeOnlyApproved(t *testing.T) {
	repo := creative.NewInMemoryRepository()
	svc := creative.NewService(repo)
	ctx := context.Background()

	draft, _ := svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Draft",
		Format:     creative.CreativeFormatBanner,
	})

	svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Pending",
		Format:     creative.CreativeFormatBanner,
	})

	cr3, _ := svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Approved",
		Format:     creative.CreativeFormatBanner,
	})
	svc.Approve(ctx, cr3.ID)

	served, err := svc.ServeCreative(ctx, "camp-1")
	if err != nil {
		t.Fatalf("ServeCreative failed: %v", err)
	}
	if served == nil {
		t.Fatal("Expected a creative to serve")
	}
	if served.ID != cr3.ID {
		t.Errorf("Expected approved creative, got %s", served.ID)
	}

	svc.Reject(ctx, draft.ID)

	served2, err := svc.ServeCreative(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("ServeCreative failed: %v", err)
	}
	if served2 != nil {
		t.Error("Expected nil for nonexistent campaign")
	}
}

func TestCreativeRotation(t *testing.T) {
	svc := creative.NewService(creative.NewInMemoryRepository())
	ctx := context.Background()

	creatives := []*creative.Creative{
		{ID: "c1", Name: "High CTR", Performance: creative.CreativePerformance{CTR: 5.0}},
		{ID: "c2", Name: "Low CTR", Performance: creative.CreativePerformance{CTR: 1.0}},
		{ID: "c3", Name: "Medium CTR", Performance: creative.CreativePerformance{CTR: 3.0}},
	}

	selected, err := svc.RotateCreative(ctx, creatives)
	if err != nil {
		t.Fatalf("RotateCreative failed: %v", err)
	}
	if selected == nil {
		t.Fatal("Expected a creative")
	}

	_, err = svc.RotateCreative(ctx, []*creative.Creative{})
	if err != nil {
		t.Fatalf("RotateCreative with empty slice failed: %v", err)
	}
}

func TestCreativeListByStatus(t *testing.T) {
	repo := creative.NewInMemoryRepository()
	svc := creative.NewService(repo)
	ctx := context.Background()

	cr1, _ := svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Banner 1",
		Format:     creative.CreativeFormatBanner,
	})
	svc.Create(ctx, &creative.Creative{
		CampaignID: "camp-1",
		Name:       "Banner 2",
		Format:     creative.CreativeFormatBanner,
	})
	svc.Approve(ctx, cr1.ID)

	all, _ := svc.List(ctx, "")
	if len(all) != 2 {
		t.Errorf("Expected 2 total creatives, got %d", len(all))
	}

	drafts, _ := svc.List(ctx, creative.CreativeStatusDraft)
	if len(drafts) != 1 {
		t.Errorf("Expected 1 draft, got %d", len(drafts))
	}

	approved, _ := svc.List(ctx, creative.CreativeStatusApproved)
	if len(approved) != 1 {
		t.Errorf("Expected 1 approved, got %d", len(approved))
	}
}
