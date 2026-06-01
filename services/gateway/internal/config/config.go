package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName  string
	AppEnv   string
	LogLevel string
	HTTPPort int
	GRPCPort int

	Redis RedisConfig

	RateLimit RateLimitConfig

	Auth AuthConfig

	OpenTelemetry OTELConfig

	Upstreams UpstreamConfig

	CORS CORSConfig

	Server ServerConfig
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxRetries   int
}

type RateLimitConfig struct {
	Enabled          bool
	GlobalMaxRPS     int
	DefaultMaxRPS    int
	IPMaxRPS         int
	AuthenticatedRPS int
	LoginMaxRPS      int
	CheckoutMaxRPS   int
	WindowSize       time.Duration
}

type AuthConfig struct {
	JWKSEndpoint      string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	EnableRBAC        bool
	TokenBlacklistTTL time.Duration
	AccessTokenKey    string
	RefreshTokenKey   string
}

type OTELConfig struct {
	Endpoint      string
	ServiceName   string
	TraceRatio    float64
	MetricsPrefix string
}

type CircuitBreakerConfig struct {
	Enabled      bool
	MaxRequests  int
	Interval     time.Duration
	Timeout      time.Duration
	FailureRatio float64
	MinSamples   int
}

type UpstreamConfig struct {
	AuthService           string
	CatalogService        string
	CatalogServiceReplica2 string
	CatalogServiceReplica3 string
	CartService           string
	OrderService          string
	InventoryService      string
	PaymentService        string
	SearchService         string
	RecommendationService string
	DeliveryService       string
	DefaultTimeout        time.Duration
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	MaxRetries            int
	CircuitBreaker        CircuitBreakerConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	MaxAge           time.Duration
	AllowCredentials bool
}

type ServerConfig struct {
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	MaxBodySize       int64
	EnablePprof       bool
	TrustedProxies    []string
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "tiki-gateway"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("GATEWAY_HTTP_PORT", 8080),
		GRPCPort: getEnvInt("GATEWAY_GRPC_PORT", 9090),

		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE", 20),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},

		RateLimit: RateLimitConfig{
			Enabled:          getEnvBool("RATE_LIMIT_ENABLED", false),
			GlobalMaxRPS:     getEnvInt("RATE_LIMIT_GLOBAL_RPS", 100000),
			DefaultMaxRPS:    getEnvInt("RATE_LIMIT_DEFAULT_RPS", 10000),
			IPMaxRPS:         getEnvInt("RATE_LIMIT_IP_RPS", 5000),
			AuthenticatedRPS: getEnvInt("RATE_LIMIT_AUTH_RPS", 5000),
			LoginMaxRPS:      getEnvInt("RATE_LIMIT_LOGIN_RPS", 5),
			CheckoutMaxRPS:   getEnvInt("RATE_LIMIT_CHECKOUT_RPS", 1),
			WindowSize:       getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Second),
		},

		Auth: AuthConfig{
			JWKSEndpoint:      getEnv("JWKS_ENDPOINT", ""),
			AccessTTL:         getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:        getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
			EnableRBAC:        getEnvBool("RBAC_ENABLED", true),
			TokenBlacklistTTL: getEnvDuration("TOKEN_BLACKLIST_TTL", 24*time.Hour),
			AccessTokenKey:    getEnv("JWT_ACCESS_SECRET", ""),
			RefreshTokenKey:   getEnv("JWT_REFRESH_SECRET", ""),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:      getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName:   getEnv("OTEL_SERVICE_NAME", "tiki-gateway"),
			TraceRatio:    getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
			MetricsPrefix: getEnv("OTEL_METRICS_PREFIX", "tiki_gateway"),
		},

		Upstreams: UpstreamConfig{
			AuthService:           getEnv("UPSTREAM_AUTH_SERVICE", "identity-auth:8080"),
			CatalogService:        getEnv("UPSTREAM_CATALOG_SERVICE", "catalog-product:8080"),
			CatalogServiceReplica2: getEnv("UPSTREAM_CATALOG_SERVICE_REPLICA2", "catalog-product-2:8088"),
			CatalogServiceReplica3: getEnv("UPSTREAM_CATALOG_SERVICE_REPLICA3", "catalog-product-3:8088"),
			CartService:           getEnv("UPSTREAM_CART_SERVICE", "shopping-cart:8080"),
			OrderService:          getEnv("UPSTREAM_ORDER_SERVICE", "order-processing:8080"),
			InventoryService:      getEnv("UPSTREAM_INVENTORY_SERVICE", "inventory-flashsale:8080"),
			PaymentService:        getEnv("UPSTREAM_PAYMENT_SERVICE", "payment-ledger:8080"),
			SearchService:         getEnv("UPSTREAM_SEARCH_SERVICE", "search-indexing:8080"),
			RecommendationService: getEnv("UPSTREAM_RECOMMENDATION_SERVICE", "recommendation-ml:8080"),
			DeliveryService:       getEnv("UPSTREAM_DELIVERY_SERVICE", "shipment:8092"),
			DefaultTimeout:        getEnvDuration("UPSTREAM_DEFAULT_TIMEOUT", 5*time.Second),
			MaxIdleConns:          getEnvInt("UPSTREAM_MAX_IDLE_CONNS", 5000),
			IdleConnTimeout:       getEnvDuration("UPSTREAM_IDLE_CONN_TIMEOUT", 300*time.Second),
			MaxRetries:            getEnvInt("UPSTREAM_MAX_RETRIES", 0),
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:      getEnvBool("CIRCUIT_BREAKER_ENABLED", false),
				MaxRequests:  getEnvInt("CIRCUIT_BREAKER_MAX_REQUESTS", 5),
				Interval:     getEnvDuration("CIRCUIT_BREAKER_INTERVAL", 60*time.Second),
				Timeout:      getEnvDuration("CIRCUIT_BREAKER_TIMEOUT", 30*time.Second),
				FailureRatio: getEnvFloat("CIRCUIT_BREAKER_FAILURE_RATIO", 0.6),
				MinSamples:   getEnvInt("CIRCUIT_BREAKER_MIN_SAMPLES", 5),
			},
		},

		CORS: CORSConfig{
			AllowedOrigins:   splitEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://api.tiki-clone.com"),
			AllowedMethods:   strings.Split(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"), ","),
			AllowedHeaders:   strings.Split(getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Correlation-ID"), ","),
			ExposedHeaders:   strings.Split(getEnv("CORS_EXPOSED_HEADERS", "X-Request-ID,X-Correlation-ID,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset"), ","),
			MaxAge:           getEnvDuration("CORS_MAX_AGE", 86400*time.Second),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
		},

		Server: ServerConfig{
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 60*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxHeaderBytes:  getEnvInt("SERVER_MAX_HEADER_BYTES", 1<<20),
			MaxBodySize:     int64(getEnvInt("SERVER_MAX_BODY_SIZE_MB", 10)) * 1024 * 1024,
			EnablePprof:     getEnvBool("SERVER_ENABLE_PPROF", false),
			TrustedProxies:  strings.Split(getEnv("SERVER_TRUSTED_PROXIES", "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"), ","),
		},
	}
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}

func splitEnv(key, fallback string) []string {
	val := getEnv(key, fallback)
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
