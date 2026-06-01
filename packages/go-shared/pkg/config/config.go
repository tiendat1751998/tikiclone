package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	AppName  string
	AppEnv   string
	LogLevel string
	Port     int

	Postgres PostgresConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	MinIO    MinIOConfig
	JWT      JWTConfig

	OpenTelemetry OTELConfig
	RateLimit     RateLimitConfig
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	MaxConns int
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode, c.MaxConns,
	)
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers []string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	JWKSEndpoint  string
}

type OTELConfig struct {
	Endpoint string
	Protocol string
}

type RateLimitConfig struct {
	LoginMax        int
	LoginWindow     time.Duration
	CheckoutMax     int
	CheckoutWindow  time.Duration
	RedisURL        string
}

func Load() *AppConfig {
	return &AppConfig{
		AppName:  getEnv("APP_NAME", "tiki-clone"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Port:     getEnvInt("PORT", 8080),

		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "tiki"),
			Password: getEnv("POSTGRES_PASSWORD", "tiki_dev"),
			Database: getEnv("POSTGRES_DB", "tiki"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getEnvInt("POSTGRES_MAX_CONNS", 25),
		},

		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "tiki_catalog"),
		},

		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},

		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		},

		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "tiki_access"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "tiki_secret"),
			Bucket:    getEnv("MINIO_BUCKET", "tiki-assets"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
		},

		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "change-me-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-me-too-in-production"),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			JWKSEndpoint:  getEnv("JWKS_ENDPOINT", ""),
		},

		OpenTelemetry: OTELConfig{
			Endpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			Protocol: getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf"),
		},

		RateLimit: RateLimitConfig{
			LoginMax:       getEnvInt("RATE_LIMIT_LOGIN_MAX", 5),
			LoginWindow:    getEnvDuration("RATE_LIMIT_LOGIN_WINDOW", 5*time.Minute),
			CheckoutMax:    getEnvInt("RATE_LIMIT_CHECKOUT_MAX", 1),
			CheckoutWindow: getEnvDuration("RATE_LIMIT_CHECKOUT_WINDOW", 5*time.Second),
			RedisURL:       getEnv("RATE_LIMIT_REDIS_URL", "localhost:6379"),
		},
	}
}

func (c *AppConfig) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func (c *AppConfig) IsProduction() bool {
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
