package http

import (
	"github.com/tikiclone/tiki/platforms/sre/internal/alerting"
	"github.com/tikiclone/tiki/platforms/sre/internal/deployment"
	"github.com/tikiclone/tiki/platforms/sre/internal/healthcheck"
	"github.com/tikiclone/tiki/platforms/sre/internal/incident"
	"github.com/tikiclone/tiki/platforms/sre/internal/runbook"
	"github.com/tikiclone/tiki/platforms/sre/internal/slo"
)

type Handler struct {
	incidentSvc   *incident.Service
	alertingSvc   *alerting.Service
	healthcheckSvc *healthcheck.Service
	sloSvc        *slo.Service
	deploymentSvc *deployment.Service
	runbookSvc    *runbook.Service
}

func NewHandler(
	incidentSvc *incident.Service,
	alertingSvc *alerting.Service,
	healthcheckSvc *healthcheck.Service,
	sloSvc *slo.Service,
	deploymentSvc *deployment.Service,
	runbookSvc *runbook.Service,
) *Handler {
	return &Handler{
		incidentSvc:   incidentSvc,
		alertingSvc:   alertingSvc,
		healthcheckSvc: healthcheckSvc,
		sloSvc:        sloSvc,
		deploymentSvc: deploymentSvc,
		runbookSvc:    runbookSvc,
	}
}
