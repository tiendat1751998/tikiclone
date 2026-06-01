package http

import (
	"github.com/tikiclone/tiki/platforms/advertising/internal/analytics"
	"github.com/tikiclone/tiki/platforms/advertising/internal/bidding"
	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
	"github.com/tikiclone/tiki/platforms/advertising/internal/creative"
	"github.com/tikiclone/tiki/platforms/advertising/internal/events"
)

type Handler struct {
	campaignSvc campaign.Service
	biddingSvc  bidding.Service
	creativeSvc creative.Service
	analyticsSvc analytics.Service
	publisher   events.Publisher
}

func NewHandler(
	campaignSvc campaign.Service,
	biddingSvc bidding.Service,
	creativeSvc creative.Service,
	analyticsSvc analytics.Service,
	pub events.Publisher,
) *Handler {
	return &Handler{
		campaignSvc:  campaignSvc,
		biddingSvc:   biddingSvc,
		creativeSvc:  creativeSvc,
		analyticsSvc: analyticsSvc,
		publisher:    pub,
	}
}
