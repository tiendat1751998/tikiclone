package http

import (
	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
	fraudcase "github.com/tikiclone/tiki/platforms/fraud/internal/case"
	"github.com/tikiclone/tiki/platforms/fraud/internal/detection"
	"github.com/tikiclone/tiki/platforms/fraud/internal/rules"
	"github.com/tikiclone/tiki/platforms/fraud/internal/verification"
)

type Handler struct {
	detectSvc       *detection.Service
	ruleSvc         *rules.Service
	blacklistSvc    *blacklist.Service
	caseSvc         *fraudcase.Service
	verificationSvc *verification.Service
}

func NewHandler(
	ds *detection.Service,
	rs *rules.Service,
	bs *blacklist.Service,
	cs *fraudcase.Service,
	vs *verification.Service,
) *Handler {
	return &Handler{
		detectSvc:       ds,
		ruleSvc:         rs,
		blacklistSvc:    bs,
		caseSvc:         cs,
		verificationSvc: vs,
	}
}
