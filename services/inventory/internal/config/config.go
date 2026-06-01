package config

import (
	"log"
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
	MySQL    MySQLConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Inventory InventoryConfig
	Idempotency IdempotencyConfig
	OpenTelemetry OTELConfig
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

type KafkaConfig struct {
	Brokers       []string
	TopicPrefix   string
	ConsumerGroup string
	DLQTopic      string
}

type JWTConfig struct {
	AccessSecret string
	AccessTTL    time.Duration
	Issuer       string
	Audience     string
}

type InventoryConfig struct {
	ReservationTTL       time.Duration
	ReorderLevel         int
	IdempotencyTTL       time.Duration
	ReconciliationInterval time.Duration
	FlashSaleRateLimit   int
}

type IdempotencyConfig struct {
	TTL time.Duration
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

func Load() *Config {
	return &Config{
		AppName: getEnv("APP_NAME", "tiki-inventory"), AppEnv: getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"), HTTPPort: getEnvInt("INVENTORY_HTTP_PORT", 8086),
		GRPCPort: getEnvInt("INVENTORY_GRPC_PORT", 9096),
		MySQL: MySQLConfig{
			Host: getEnv("MYSQL_HOST", "localhost"), Port: getEnvInt("MYSQL_PORT", 3306),
			User: getEnv("MYSQL_USER", "tiki"), Password: requireEnv("MYSQL_PASSWORD"),
			Database: getEnv("MYSQL_DATABASE", "tiki_inventory"), MaxOpenConns: 25, MaxIdleConns: 10,
			MaxLifetime: 5 * time.Minute, Timeout: 5 * time.Second,
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", ""),
			DB: getEnvInt("REDIS_DB", 6), PoolSize: 100, MinIdleConns: 20,
			DialTimeout: 5 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, MaxRetries: 3,
		},
		Kafka: KafkaConfig{
			Brokers: getEnvSlice("KAFKA_BROKERS", ","), TopicPrefix: getEnv("KAFKA_TOPIC_PREFIX", "tiki.inventory"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "tiki-inventory-service"), DLQTopic: "tiki.inventory.dlq",
		},
		JWT: JWTConfig{
			AccessSecret: requireEnv("JWT_ACCESS_SECRET"),
			AccessTTL: getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute), Issuer: "tiki-auth", Audience: "tiki-clone",
		},
		Inventory: InventoryConfig{
			ReservationTTL: getEnvDuration("INVENTORY_RESERVATION_TTL", 30*time.Minute),
			ReorderLevel: getEnvInt("INVENTORY_REORDER_LEVEL", 10),
			IdempotencyTTL: getEnvDuration("INVENTORY_IDEMPOTENCY_TTL", 24*time.Hour),
			ReconciliationInterval: getEnvDuration("INVENTORY_RECONCILIATION_INTERVAL", 5*time.Minute),
			FlashSaleRateLimit: getEnvInt("FLASH_SALE_RATE_LIMIT", 1000),
		},
		Idempotency: IdempotencyConfig{TTL: getEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour)},
		OpenTelemetry: OTELConfig{
			Endpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-inventory"),
			TraceRatio: getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},
	}
}

func getEnv(key, fallback string) string { if v := os.Getenv(key); v != "" { return v }; return fallback }
func requireEnv(key string) string { if v := os.Getenv(key); v != "" { return v }; log.Fatalf("required environment variable %s is not set", key); return "" }
func getEnvInt(key string, fallback int) int { if v := os.Getenv(key); v != "" { if i, err := strconv.Atoi(v); err == nil { return i } }; return fallback }
func getEnvDuration(key string, fallback time.Duration) time.Duration { if v := os.Getenv(key); v != "" { if d, err := time.ParseDuration(v); err == nil { return d } }; return fallback }
func getEnvFloat(key string, fallback float64) float64 { if v := os.Getenv(key); v != "" { if f, err := strconv.ParseFloat(v, 64); err == nil { return f } }; return fallback }
func getEnvSlice(key, sep string) []string { v := getEnv(key, ""); if v == "" { return []string{"localhost:9092"} }; return strings.Split(v, sep) }
