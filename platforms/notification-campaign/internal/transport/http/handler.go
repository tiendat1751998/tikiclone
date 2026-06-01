package http

import (
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/audience"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/campaign"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/content"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/deliveryopt"
	"github.com/tikiclone/tiki/platforms/notification-campaign/internal/reporting"
)

type Handler struct {
	campaignSvc  campaign.Service
	audienceSvc  audience.Service
	contentSvc   content.Service
	deliverySvc  deliveryopt.Service
	reportingSvc reporting.Service
}

func NewHandler(
	campaignSvc campaign.Service,
	audienceSvc audience.Service,
	contentSvc content.Service,
	deliverySvc deliveryopt.Service,
	reportingSvc reporting.Service,
) *Handler {
	return &Handler{
		campaignSvc:  campaignSvc,
		audienceSvc:  audienceSvc,
		contentSvc:   contentSvc,
		deliverySvc:  deliverySvc,
		reportingSvc: reportingSvc,
	}
}
