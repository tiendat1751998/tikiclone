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

	Redis    RedisConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
	OTEL     OTELConfig

	SFU          SFUConfig
	StreamHealth StreamHealthConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
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
}

type SFUConfig struct {
	HeartbeatInterval time.Duration
	NodeTimeout       time.Duration
}

type StreamHealthConfig struct {
	BitrateThreshold      int
	FrameRateThreshold    float64
	LatencyThresholdMs    int
	PacketLossThreshold   float64
	CheckInterval         time.Duration
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

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func Load() *Config {
	return &Config{
		AppName:  env("APP_NAME", "tiki-live-scale"),
		AppEnv:   env("APP_ENV", "development"),
		LogLevel: env("LOG_LEVEL", "info"),
		HTTPPort: envInt("HTTP_PORT", 8081),
		GRPCPort: envInt("GRPC_PORT", 9091),
		Redis: RedisConfig{
			Addr:     env("REDIS_ADDR", "localhost:6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		Postgres: PostgresConfig{
			DSN: env("DATABASE_DSN", "postgres://tiki:tiki_dev@localhost:5432/tiki_live_scale?sslmode=disable"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		OTEL: OTELConfig{
			Endpoint:    env("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: env("OTEL_SERVICE_NAME", "tiki-live-scale"),
		},
		SFU: SFUConfig{
			HeartbeatInterval: envDuration("SFU_HEARTBEAT_INTERVAL", 10*time.Second),
			NodeTimeout:       envDuration("SFU_NODE_TIMEOUT", 30*time.Second),
		},
		StreamHealth: StreamHealthConfig{
			BitrateThreshold:    envInt("STREAM_BITRATE_THRESHOLD", 500000),
			FrameRateThreshold:  float64(envInt("STREAM_FRAMERATE_THRESHOLD", 24)),
			LatencyThresholdMs:  envInt("STREAM_LATENCY_THRESHOLD_MS", 500),
			PacketLossThreshold: float64(envInt("STREAM_PACKET_LOSS_THRESHOLD", 5)) / 100.0,
			CheckInterval:       envDuration("STREAM_CHECK_INTERVAL", 15*time.Second),
		},
	}
}
