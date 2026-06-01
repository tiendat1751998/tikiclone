package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName       string
	AppEnv        string
	LogLevel      string
	HTTPPort      int
	GRPCPort      int
	Redis         RedisConfig
	Postgres      PostgresConfig
	Kafka         KafkaConfig
	OpenTelemetry OTELConfig
	Analytics     AnalyticsConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type PostgresConfig struct {
	DSN string
}

type KafkaConfig struct {
	Brokers []string
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

type AnalyticsConfig struct {
	SessionTimeoutMinutes int
	DefaultPageSize       int
	MaxQueryWindowDays    int
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func Load() *Config {
	return &Config{
		AppName:  env("APP_NAME", "tiki-analytics"),
		AppEnv:   env("APP_ENV", "development"),
		LogLevel: env("LOG_LEVEL", "info"),
		HTTPPort: envInt("ANALYTICS_HTTP_PORT", 8080),
		GRPCPort: envInt("ANALYTICS_GRPC_PORT", 9090),
		Redis: RedisConfig{
			Addr:     env("REDIS_ADDR", "localhost:6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
			PoolSize: envInt("REDIS_POOL_SIZE", 100),
		},
		Postgres: PostgresConfig{
			DSN: env("DATABASE_DSN", "postgres://tiki:tiki_dev@localhost:5432/tiki_analytics?sslmode=disable"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		OpenTelemetry: OTELConfig{
			Endpoint:    env("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: env("OTEL_SERVICE_NAME", "tiki-analytics"),
			TraceRatio:  envFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},
		Analytics: AnalyticsConfig{
			SessionTimeoutMinutes: envInt("SESSION_TIMEOUT_MINUTES", 30),
			DefaultPageSize:       envInt("DEFAULT_PAGE_SIZE", 50),
			MaxQueryWindowDays:    envInt("MAX_QUERY_WINDOW_DAYS", 365),
		},
	}
}
