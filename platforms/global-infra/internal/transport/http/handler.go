package http

import (
	"github.com/tikiclone/tiki/platforms/global-infra/internal/configmanager"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/featureflag"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/multiregion"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/ratelimit"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/registry"
	"github.com/tikiclone/tiki/platforms/global-infra/internal/secrets"
)

type Handler struct {
	ConfigSvc     *configmanager.Service
	FeatureFlagSvc *featureflag.Service
	MultiRegionSvc *multiregion.Service
	RegistrySvc   *registry.Service
	SecretSvc     *secrets.Service
	RateLimiter   *ratelimit.RateLimiter
}

func NewHandler(
	configSvc *configmanager.Service,
	featureFlagSvc *featureflag.Service,
	multiRegionSvc *multiregion.Service,
	registrySvc *registry.Service,
	secretSvc *secrets.Service,
	rateLimiter *ratelimit.RateLimiter,
) *Handler {
	return &Handler{
		ConfigSvc:      configSvc,
		FeatureFlagSvc: featureFlagSvc,
		MultiRegionSvc: multiRegionSvc,
		RegistrySvc:    registrySvc,
		SecretSvc:      secretSvc,
		RateLimiter:    rateLimiter,
	}
}
