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
	Kafka    KafkaConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
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
		AppName:  env("APP_NAME", "tiki-global-infra"),
		AppEnv:   env("APP_ENV", "development"),
		LogLevel: env("LOG_LEVEL", "info"),
		HTTPPort: envInt("GLOBAL_INFRA_HTTP_PORT", 8085),
		Kafka: KafkaConfig{
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
			Topic:   env("KAFKA_CONFIG_TOPIC", "config-changes"),
		},
	}
}
