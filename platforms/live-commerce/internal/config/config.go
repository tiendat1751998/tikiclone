package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName     string
	AppEnv      string
	LogLevel    string
	HTTPPort    int
	GRPCPort    int
	Redis       RedisConfig
	Postgres    PostgresConfig
	ClickHouse  ClickHouseConfig
	Kafka       KafkaConfig
	OpenTelemetry OTELConfig
}

type RedisConfig struct {
	Addr          string
	Password      string
	DB            int
	PoolSize      int
	MinIdleConns  int
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	MaxRetries    int
}

type PostgresConfig struct {
	DSN string
}

type ClickHouseConfig struct {
	DSN  string
	Enabled bool
}

type KafkaConfig struct {
	Brokers []string
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
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

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return fallback
}

func Load() *Config {
	return &Config{
		AppName:  env("APP_NAME", "tiki-live-commerce"),
		AppEnv:   env("APP_ENV", "development"),
		LogLevel: env("LOG_LEVEL", "info"),
		HTTPPort: envInt("HTTP_PORT", 8080),
		GRPCPort: envInt("GRPC_PORT", 9090),
		Redis: RedisConfig{
			Addr:         env("REDIS_ADDR", "localhost:6379"),
			Password:     env("REDIS_PASSWORD", ""),
			DB:           envInt("REDIS_DB", 0),
			PoolSize:     envInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: envInt("REDIS_MIN_IDLE", 20),
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			MaxRetries:   3,
		},
		Postgres: PostgresConfig{
			DSN: env("DATABASE_DSN", "postgres://tiki:tiki_dev@localhost:5432/tiki_live?sslmode=disable"),
		},
		ClickHouse: ClickHouseConfig{
			DSN:     env("CLICKHOUSE_DSN", "clickhouse://localhost:9000/tiki_analytics?sslmode=disable"),
			Enabled: envBool("CLICKHOUSE_ENABLED", true),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		OpenTelemetry: OTELConfig{
			Endpoint:    env("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: env("OTEL_SERVICE_NAME", "tiki-live-commerce"),
			TraceRatio:  0.1,
		},
	}
}
