package routing

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"

	"github.com/tikiclone/tiki/services/gateway/internal/auth"
	"github.com/tikiclone/tiki/services/gateway/internal/cache"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"github.com/tikiclone/tiki/services/gateway/internal/discovery"
	gwHealth "github.com/tikiclone/tiki/services/gateway/internal/health"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
	"github.com/tikiclone/tiki/services/gateway/internal/ratelimit"
	"github.com/tikiclone/tiki/services/gateway/internal/transport"
	sharedMiddleware "github.com/tikiclone/tiki/packages/go-shared/pkg/middleware"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/health"
)

type Router struct {
	cfg           *config.Config
	proxy         *transport.Proxy
	grpcProxy     *transport.GRPCProxy
	rateLimiter   *ratelimit.RateLimiter
	authMW        *auth.AuthMiddleware
	discovery     *discovery.ServiceDiscovery
	health        *health.Checker
	rdb           *redis.Client
	cacheMW       *cache.CacheMiddleware
}

func NewRouter(
	cfg *config.Config,
	proxy *transport.Proxy,
	grpcProxy *transport.GRPCProxy,
	rateLimiter *ratelimit.RateLimiter,
	authMW *auth.AuthMiddleware,
	svcDiscovery *discovery.ServiceDiscovery,
	healthChecker *health.Checker,
	rdb *redis.Client,
) *Router {
	return &Router{
		cfg:         cfg,
		proxy:       proxy,
		grpcProxy:   grpcProxy,
		rateLimiter: rateLimiter,
		authMW:      authMW,
		discovery:   svcDiscovery,
		health:      healthChecker,
		rdb:         rdb,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(
		sharedMiddleware.Recovery(),
		sharedMiddleware.ErrorHandler(),
		middleware.CorrelationID(),
		middleware.SecurityHeaders(),
		sharedMiddleware.OTelMiddleware(r.cfg.OpenTelemetry.ServiceName),
		middleware.RequestLogger(),
		middleware.RequestSanitizer(),
		middleware.BodySizeLimiter(r.cfg.Server.MaxBodySize),
		middleware.CORS(r.cfg.CORS),
		middleware.DeviceMetadata(),
		middleware.AntiAbuse(),
		middleware.RequestValidation(),
	)

	// Response cache for read-heavy endpoints (products, categories, search)
	cacheMW := cache.NewCacheMiddleware(r.rdb, 60*time.Second)
	engine.Use(cacheMW.CacheByPath(60*time.Second,
		"/api/v1/products",
		"/api/v1/categories",
	))

	r.setupSystemEndpoints(engine)
	r.setupUpstreamRoutes(engine)
	r.setupGRPCRoutes(engine)
	r.setupFallback(engine)
}

func (r *Router) setupSystemEndpoints(engine *gin.Engine) {
	engine.GET("/health", r.health.LivenessHandler())
	engine.GET("/ready", r.health.ReadinessHandler())
	engine.GET("/metrics", observability.MetricsHandler())

	engine.GET("/upstreams", func(c *gin.Context) {
		services := r.discovery.GetAllServices()
		c.JSON(http.StatusOK, gin.H{
			"services": services,
			"count":    len(services),
		})
	})

	gatewayHealth := gwHealth.NewGatewayHealth(r.discovery, r.rdb)
	engine.GET("/health/upstreams", gatewayHealth.UpstreamsHandler())

	if r.cfg.Server.EnablePprof {
		observability.GetLogger().Warn("pprof endpoints enabled, not recommended for production")
	}
}

func (r *Router) setupUpstreamRoutes(engine *gin.Engine) {
	for _, route := range RouteTable {
		if route.Protocol == "grpc" {
			continue
		}
		r.registerRouteGroup(engine, route)
	}

}

func (r *Router) setupGRPCRoutes(engine *gin.Engine) {
	for _, route := range RouteTable {
		if route.Protocol != "grpc" {
			continue
		}
		r.registerGRPCRoute(engine, route)
	}
}

func (r *Router) registerRouteGroup(engine *gin.Engine, route RouteGroup) {
	var authMW gin.HandlerFunc
	if route.Auth && len(route.Roles) > 0 {
		authMW = r.authMW.RequireRoles(route.Roles...)
	} else if route.Auth {
		authMW = r.authMW.RequireAuth()
	} else {
		authMW = func(c *gin.Context) { c.Next() }
	}

	rateLimitMW := r.rateLimiter.AuthenticatedRateLimit()
	endpointRateLimit := r.rateLimiter.PerEndpointRateLimit(route.RateLimit)
	handler := r.proxy.ReverseProxy(route.ToProxyTarget())

	group := engine.Group(route.Prefix)
	group.Use(endpointRateLimit)
	if route.Auth {
		group.Use(authMW)
	}
	group.Use(rateLimitMW)

	group.Any("/*path", handler)
	group.Any("", handler)

	observability.GetLogger().Info("registered route",
		zap.String("prefix", route.Prefix),
		zap.String("target", route.Target),
		zap.Bool("auth", route.Auth),
		zap.Int("rate_limit", route.RateLimit),
	)
}

func (r *Router) registerGRPCRoute(engine *gin.Engine, route RouteGroup) {
	observability.GetLogger().Info("registered gRPC route",
		zap.String("prefix", route.Prefix),
		zap.String("target", route.Target),
		zap.String("grpc_method", route.GRPCMethod),
	)
}

func (r *Router) setupFallback(engine *gin.Engine) {
	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		for _, route := range RouteTable {
			if strings.HasPrefix(path, route.Prefix) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"error_code": "UPSTREAM_NOT_FOUND",
					"message":    fmt.Sprintf("route %s not found in upstream service %s", path, route.Target),
					"path":       path,
					"service":    route.Target,
				})
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error_code": "ROUTE_NOT_FOUND",
			"message":    fmt.Sprintf("no route configured for %s", path),
			"path":       path,
		})
	})

	engine.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
			"error_code": "METHOD_NOT_ALLOWED",
			"message":    fmt.Sprintf("method %s not allowed for %s", c.Request.Method, c.Request.URL.Path),
		})
	})
}
