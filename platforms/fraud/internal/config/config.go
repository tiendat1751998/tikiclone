package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName  string
	AppEnv   string
	LogLevel string
	HTTPPort int
	GRPCPort int

	Redis          RedisConfig
	Postgres       PostgresConfig
	Kafka          KafkaConfig
	OpenTelemetry  OTELConfig
	Fraud          FraudConfig
	Verification   VerificationConfig
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

type FraudConfig struct {
	DefaultThreshold int
	ScoreWindowSize  int
}

type VerificationConfig struct {
	CodeExpiryMinutes int
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

func Load() *Config {
	return &Config{
		AppName:  env("APP_NAME", "tiki-fraud"),
		AppEnv:   env("APP_ENV", "development"),
		LogLevel: env("LOG_LEVEL", "info"),
		HTTPPort: envInt("HTTP_PORT", 8080),
		GRPCPort: envInt("GRPC_PORT", 9090),
		Redis: RedisConfig{
			Addr:     env("REDIS_ADDR", "localhost:6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
			PoolSize: envInt("REDIS_POOL_SIZE", 100),
		},
		Postgres: PostgresConfig{
			DSN: env("DATABASE_DSN", "postgres://tiki:tiki_dev@localhost:5432/tiki_fraud?sslmode=disable"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		OpenTelemetry: OTELConfig{
			Endpoint:    env("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: env("OTEL_SERVICE_NAME", "tiki-fraud"),
			TraceRatio:  0.1,
		},
		Fraud: FraudConfig{
			DefaultThreshold: envInt("FRAUD_DEFAULT_THRESHOLD", 50),
			ScoreWindowSize:  envInt("FRAUD_SCORE_WINDOW_SIZE", 100),
		},
		Verification: VerificationConfig{
			CodeExpiryMinutes: envInt("VERIFICATION_CODE_EXPIRY_MINUTES", 10),
		},
	}
}
