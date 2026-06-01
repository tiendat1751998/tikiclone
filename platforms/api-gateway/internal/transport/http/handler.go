package http

import (
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/auth"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/circuitbreaker"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/edgecache"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/ratelimit"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/routes"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/transform"
)

type Handler struct {
	RouteService    *routes.Service
	RateLimiter     *ratelimit.RateLimiter
	APIKeyStore     *auth.APIKeyStore
	APIKeyValidator *auth.APIKeyValidator
	JWTHandler      *auth.JWTHandler
	KeyRateLimiter  *auth.KeyRateLimiter
	Transformer     *transform.Transformer
	Composer        *transform.Composer
	CBSvc           *circuitbreaker.Service
	EdgeCache       *edgecache.Cache
}

func NewHandler(
	routeService *routes.Service,
	rateLimiter *ratelimit.RateLimiter,
	apiKeyStore *auth.APIKeyStore,
	apiKeyValidator *auth.APIKeyValidator,
	jwtHandler *auth.JWTHandler,
	keyRateLimiter *auth.KeyRateLimiter,
	transformer *transform.Transformer,
	composer *transform.Composer,
	cbSvc *circuitbreaker.Service,
	edgeCache *edgecache.Cache,
) *Handler {
	return &Handler{
		RouteService:    routeService,
		RateLimiter:     rateLimiter,
		APIKeyStore:     apiKeyStore,
		APIKeyValidator: apiKeyValidator,
		JWTHandler:      jwtHandler,
		KeyRateLimiter:  keyRateLimiter,
		Transformer:     transformer,
		Composer:        composer,
		CBSvc:           cbSvc,
		EdgeCache:       edgeCache,
	}
}
