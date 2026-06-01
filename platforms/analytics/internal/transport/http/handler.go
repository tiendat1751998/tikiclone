package http

import (
	analyticsSvc "github.com/tikiclone/tiki/platforms/analytics/internal/analytics"
	"github.com/tikiclone/tiki/platforms/analytics/internal/cohort"
	"github.com/tikiclone/tiki/platforms/analytics/internal/dashboard"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
	"github.com/tikiclone/tiki/platforms/analytics/internal/funnel"
	"github.com/tikiclone/tiki/platforms/analytics/internal/report_scheduler"
	"github.com/tikiclone/tiki/platforms/analytics/internal/session"
)

type Handler struct {
	analyticsSvc *analyticsSvc.Service
	eventSvc     *events.Service
	funnelSvc    *funnel.Service
	cohortSvc    *cohort.Service
	sessionSvc   *session.Service
	dashboardSvc *dashboard.Service
	scheduleSvc  *report_scheduler.Service
	publisher    events.Publisher
}

func NewHandler(
	analyticsSvc *analyticsSvc.Service,
	eventSvc *events.Service,
	funnelSvc *funnel.Service,
	cohortSvc *cohort.Service,
	sessionSvc *session.Service,
	dashboardSvc *dashboard.Service,
	scheduleSvc *report_scheduler.Service,
	pub events.Publisher,
) *Handler {
	return &Handler{
		analyticsSvc: analyticsSvc,
		eventSvc:     eventSvc,
		funnelSvc:    funnelSvc,
		cohortSvc:    cohortSvc,
		sessionSvc:   sessionSvc,
		dashboardSvc: dashboardSvc,
		scheduleSvc:  scheduleSvc,
		publisher:    pub,
	}
}
