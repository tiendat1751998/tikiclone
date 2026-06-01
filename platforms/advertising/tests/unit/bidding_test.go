package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/advertising/internal/bidding"
	"github.com/tikiclone/tiki/platforms/advertising/internal/budget"
	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
	"github.com/tikiclone/tiki/platforms/advertising/internal/creative"
	"github.com/tikiclone/tiki/platforms/advertising/internal/targeting"
)

func setupBidding() (bidding.Service, campaign.Service, creative.Service, *budget.InMemoryRepository, *targeting.InMemoryRepository) {
	campaignRepo := campaign.NewInMemoryRepository()
	campaignSvc := campaign.NewService(campaignRepo)

	budgetRepo := budget.NewInMemoryRepository()
	budgetSvc := budget.NewService(budgetRepo)

	targetingRepo := targeting.NewInMemoryRepository()
	targetingSvc := targeting.NewService(targetingRepo)

	creativeRepo := creative.NewInMemoryRepository()
	creativeSvc := creative.NewService(creativeRepo)

	biddingRepo := bidding.NewInMemoryRepository()
	biddingSvc := bidding.NewService(campaignSvc, budgetSvc, targetingSvc, creativeSvc, biddingRepo)

	return biddingSvc, campaignSvc, creativeSvc, budgetRepo, targetingRepo
}

func TestSecondPriceAuction(t *testing.T) {
	ctx := context.Background()
	biddingSvc, campaignSvc, creativeSvc, _, _ := setupBidding()

	cam1, _ := campaignSvc.Create(ctx, &campaign.Campaign{
		Name:      "Campaign A",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 2.00,
		Budget:    campaign.Budget{Daily: 100, Lifetime: 1000},
	})
	cam1.Status = campaign.CampaignStatusActive
	campaignSvc.Update(ctx, cam1)

	cam2, _ := campaignSvc.Create(ctx, &campaign.Campaign{
		Name:      "Campaign B",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.50,
		Budget:    campaign.Budget{Daily: 100, Lifetime: 1000},
	})
	cam2.Status = campaign.CampaignStatusActive
	campaignSvc.Update(ctx, cam2)

	cam3, _ := campaignSvc.Create(ctx, &campaign.Campaign{
		Name:      "Campaign C",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.00,
		Budget:    campaign.Budget{Daily: 100, Lifetime: 1000},
	})
	cam3.Status = campaign.CampaignStatusActive
	campaignSvc.Update(ctx, cam3)

	cr1, _ := creativeSvc.Create(ctx, &creative.Creative{
		CampaignID:     cam1.ID,
		Name:           "Ad A",
		Format:         creative.CreativeFormatBanner,
		Content:        "https://example.com/ad-a.jpg",
		DestinationURL: "https://example.com",
	})
	creativeSvc.Approve(ctx, cr1.ID)

	cr2, _ := creativeSvc.Create(ctx, &creative.Creative{
		CampaignID:     cam2.ID,
		Name:           "Ad B",
		Format:         creative.CreativeFormatBanner,
		Content:        "https://example.com/ad-b.jpg",
		DestinationURL: "https://example.com",
	})
	creativeSvc.Approve(ctx, cr2.ID)

	cr3, _ := creativeSvc.Create(ctx, &creative.Creative{
		CampaignID:     cam3.ID,
		Name:           "Ad C",
		Format:         creative.CreativeFormatBanner,
		Content:        "https://example.com/ad-c.jpg",
		DestinationURL: "https://example.com",
	})
	creativeSvc.Approve(ctx, cr3.ID)

	result, err := biddingSvc.RunAuction(ctx, &bidding.BidRequest{
		UserID: "user1",
		Context: bidding.BidContext{
			Device:   "mobile",
			Location: "US",
		},
	})
	if err != nil {
		t.Fatalf("RunAuction failed: %v", err)
	}

	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}
	if result.Winner.CampaignID != cam1.ID {
		t.Errorf("Expected winner %s (highest bid), got %s", cam1.ID, result.Winner.CampaignID)
	}
	if result.SecondPrice != 1.50 {
		t.Errorf("Expected second price 1.50, got %f", result.SecondPrice)
	}
	if len(result.AllBids) != 3 {
		t.Errorf("Expected 3 bids, got %d", len(result.AllBids))
	}
}

func TestAuctionNoEligibleBidders(t *testing.T) {
	ctx := context.Background()
	biddingSvc, _, _, _, _ := setupBidding()

	result, err := biddingSvc.RunAuction(ctx, &bidding.BidRequest{
		UserID: "user1",
	})
	if err != nil {
		t.Fatalf("RunAuction failed: %v", err)
	}
	if result.Winner != nil {
		t.Error("Expected no winner for empty auction")
	}
	if len(result.AllBids) != 0 {
		t.Errorf("Expected 0 bids, got %d", len(result.AllBids))
	}
}

func TestAuctionSingleBidder(t *testing.T) {
	ctx := context.Background()
	biddingSvc, campaignSvc, creativeSvc, _, _ := setupBidding()

	cam1, _ := campaignSvc.Create(ctx, &campaign.Campaign{
		Name:      "Solo",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 2.00,
		Budget:    campaign.Budget{Daily: 100, Lifetime: 1000},
	})
	cam1.Status = campaign.CampaignStatusActive
	campaignSvc.Update(ctx, cam1)

	cr1, _ := creativeSvc.Create(ctx, &creative.Creative{
		CampaignID: cam1.ID,
		Name:       "Solo Ad",
		Format:     creative.CreativeFormatText,
	})
	creativeSvc.Approve(ctx, cr1.ID)

	result, err := biddingSvc.RunAuction(ctx, &bidding.BidRequest{
		UserID: "user1",
	})
	if err != nil {
		t.Fatalf("RunAuction failed: %v", err)
	}
	if result.Winner == nil {
		t.Fatal("Expected a winner")
	}
	if result.SecondPrice != 0 {
		t.Errorf("Expected second price 0 for single bidder, got %f", result.SecondPrice)
	}
}

func TestCalculateBidStrategies(t *testing.T) {
	repo := campaign.NewInMemoryRepository()
	svc := campaign.NewService(repo)
	budgetSvc := budget.NewService(budget.NewInMemoryRepository())
	targetingSvc := targeting.NewService(targeting.NewInMemoryRepository())
	creativeSvc := creative.NewService(creative.NewInMemoryRepository())
	biddingRepo := bidding.NewInMemoryRepository()
	biddingSvc := bidding.NewService(svc, budgetSvc, targetingSvc, creativeSvc, biddingRepo)

	ctx := context.Background()
	cam, _ := svc.Create(ctx, &campaign.Campaign{
		Name:      "Bid Test",
		Type:      campaign.CampaignTypeCPC,
		BidAmount: 1.00,
		TargetCPA: 10.00,
		Budget:    campaign.Budget{Daily: 50, Lifetime: 500},
	})

	manual, _ := biddingSvc.CalculateBid(ctx, cam, bidding.BidStrategyManual)
	if manual != 1.00 {
		t.Errorf("Expected manual bid 1.00, got %f", manual)
	}

	auto, _ := biddingSvc.CalculateBid(ctx, cam, bidding.BidStrategyAuto)
	if auto != 5.00 {
		t.Errorf("Expected auto bid 5.00 (50%% of CPA), got %f", auto)
	}

	enhanced, _ := biddingSvc.CalculateBid(ctx, cam, bidding.BidStrategyEnhancedCPC)
	if enhanced <= 1.00 || enhanced > 1.50 {
		t.Errorf("Expected enhanced bid between 1.00 and 1.50, got %f", enhanced)
	}
}
