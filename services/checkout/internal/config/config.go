package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName  string
	AppEnv   string
	LogLevel string
	HTTPPort int

	MySQL MySQLConfig
	Redis RedisConfig

	SnapshotTTL        time.Duration
	ReservationTimeout time.Duration
	IdempotencyTTL     time.Duration
	MaxRetries         int

	InventoryServiceAddr string
	PromotionServiceAddr string
	OrderServiceAddr     string

	JWTConfig     JWTConfig
	OpenTelemetry OTELConfig
}

type JWTConfig struct {
	AccessSecret string
}

type MySQLConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	Timeout      time.Duration
}

func (c MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=true&loc=UTC&timeout=" + strconv.Itoa(int(c.Timeout.Milliseconds())) + "ms"
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

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "tiki-checkout"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("CHECKOUT_HTTP_PORT", 8080),

		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "tiki"),
			Password:     requireEnv("MYSQL_PASSWORD"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_checkout"),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvDuration("MYSQL_MAX_LIFETIME", 5*time.Minute),
			Timeout:      getEnvDuration("MYSQL_TIMEOUT", 5*time.Second),
		},

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

		JWTConfig: JWTConfig{
			AccessSecret: requireEnv("JWT_ACCESS_SECRET"),
		},

		SnapshotTTL:        getEnvDuration("SNAPSHOT_TTL", 30*time.Minute),
		ReservationTimeout: getEnvDuration("RESERVATION_TIMEOUT", 15*time.Minute),
		IdempotencyTTL:     getEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour),
		MaxRetries:         getEnvInt("MAX_RETRIES", 3),

		InventoryServiceAddr: getEnv("INVENTORY_SERVICE_ADDR", "tiki-inventory:9090"),
		PromotionServiceAddr: getEnv("PROMOTION_SERVICE_ADDR", "tiki-promotion:9090"),
		OrderServiceAddr:     getEnv("ORDER_SERVICE_ADDR", "tiki-order:9090"),

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-checkout"),
			TraceRatio:  getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},
	}
}

func (c *Config) IsDevelopment() bool { return c.AppEnv == "development" }
func (c *Config) IsProduction() bool  { return c.AppEnv == "production" }

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

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return val
}
