package bidding

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/tikiclone/tiki/platforms/advertising/internal/budget"
	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
	"github.com/tikiclone/tiki/platforms/advertising/internal/creative"
	"github.com/tikiclone/tiki/platforms/advertising/internal/targeting"
)

type Service interface {
	RunAuction(ctx context.Context, req *BidRequest) (*AuctionResult, error)
	CalculateBid(ctx context.Context, c *campaign.Campaign, strategy BidStrategy) (float64, error)
	BudgetCheck(ctx context.Context, campaignID string, bidAmount float64) (bool, error)
}

type service struct {
	campaignSvc  campaign.Service
	budgetSvc    budget.Service
	targetingSvc targeting.Service
	creativeSvc  creative.Service
	repo         Repository
}

func NewService(
	campaignSvc campaign.Service,
	budgetSvc budget.Service,
	targetingSvc targeting.Service,
	creativeSvc creative.Service,
	repo Repository,
) Service {
	return &service{
		campaignSvc:  campaignSvc,
		budgetSvc:    budgetSvc,
		targetingSvc: targetingSvc,
		creativeSvc:  creativeSvc,
		repo:         repo,
	}
}

func (s *service) RunAuction(ctx context.Context, req *BidRequest) (*AuctionResult, error) {
	camps, err := s.campaignSvc.List(ctx, campaign.CampaignStatusActive)
	if err != nil {
		return nil, err
	}

	var eligible []campaign.Campaign
	for _, c := range camps {
		if !c.DateRange.End.IsZero() && time.Now().After(c.DateRange.End) {
			continue
		}
		if !c.DateRange.Start.IsZero() && time.Now().Before(c.DateRange.Start) {
			continue
		}
		eligible = append(eligible, *c)
	}

	var responses []BidResponse
	for _, c := range eligible {
		ok, err := s.budgetSvc.CheckBudget(ctx, c.ID, c.BidAmount)
		if err != nil || !ok {
			continue
		}

		profile := &targeting.UserProfile{
			UserID:   req.UserID,
			Location: req.Context.Location,
			Devices:  []string{req.Context.Device},
		}
		match, err := s.targetingSvc.MatchAudience(ctx, &c.Targeting, profile)
		if err != nil || !match {
			continue
		}

		creatives, err := s.creativeSvc.GetByCampaign(ctx, c.ID)
		if err != nil || len(creatives) == 0 {
			continue
		}

		selectedCreative, err := s.creativeSvc.RotateCreative(ctx, creatives)
		if err != nil {
			continue
		}

		bidAmount, err := s.CalculateBid(ctx, &c, BidStrategyManual)
		if err != nil {
			continue
		}

		qs := s.calculateQualityScore(ctx, &c)
		adRank := bidAmount * qs

		responses = append(responses, BidResponse{
			CampaignID:   c.ID,
			CreativeID:   selectedCreative.ID,
			BidAmount:    bidAmount,
			AdRank:       adRank,
			QualityScore: qs,
		})
	}

	if len(responses) == 0 {
		return &AuctionResult{AllBids: []BidResponse{}}, nil
	}

	sort.Slice(responses, func(i, j int) bool {
		return responses[i].AdRank > responses[j].AdRank
	})

	winner := responses[0]
	secondPrice := 0.0
	if len(responses) > 1 {
		secondPrice = responses[1].BidAmount
	}

	s.repo.StoreBidHistory(ctx, &BidHistory{
		CampaignID: winner.CampaignID,
		UserID:     req.UserID,
		BidAmount:  winner.BidAmount,
		Won:        true,
		Timestamp:  time.Now(),
	})
	for i := 1; i < len(responses); i++ {
		s.repo.StoreBidHistory(ctx, &BidHistory{
			CampaignID: responses[i].CampaignID,
			UserID:     req.UserID,
			BidAmount:  responses[i].BidAmount,
			Won:        false,
			Timestamp:  time.Now(),
		})
	}

	s.budgetSvc.DeductSpend(ctx, winner.CampaignID, secondPrice)

	return &AuctionResult{
		Winner:      &winner,
		SecondPrice: secondPrice,
		AllBids:     responses,
	}, nil
}

func (s *service) CalculateBid(ctx context.Context, c *campaign.Campaign, strategy BidStrategy) (float64, error) {
	switch strategy {
	case BidStrategyAuto:
		if c.TargetCPA > 0 {
			return c.TargetCPA * 0.5, nil
		}
		return c.BidAmount, nil
	case BidStrategyEnhancedCPC:
		enhanced := c.BidAmount * 1.2
		return math.Min(enhanced, c.BidAmount*1.5), nil
	default:
		return c.BidAmount, nil
	}
}

func (s *service) BudgetCheck(ctx context.Context, campaignID string, bidAmount float64) (bool, error) {
	return s.budgetSvc.CheckBudget(ctx, campaignID, bidAmount)
}

func (s *service) calculateQualityScore(ctx context.Context, c *campaign.Campaign) float64 {
	base := c.QualityScore
	if base <= 0 {
		base = 1.0
	}
	return base
}
