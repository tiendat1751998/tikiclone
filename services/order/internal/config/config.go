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

	MySQL MySQLConfig
	Redis RedisConfig
	Kafka KafkaConfig
	JWT   JWTConfig

	Order         OrderConfig
	Idempotency   IdempotencyConfig
	OpenTelemetry OTELConfig
	Audit         AuditConfig
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

type OrderConfig struct {
	DefaultCurrency        string
	PaymentTimeout         time.Duration
	IdempotencyKeyTTL      time.Duration
	SnapshotRetentionDays  int
	ReconciliationInterval time.Duration
	MaxCancelRetryCount    int
}

type IdempotencyConfig struct {
	TTL time.Duration
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

type AuditConfig struct {
	Enabled       bool
	FlushInterval time.Duration
	BatchSize     int
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "tiki-order"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("ORDER_HTTP_PORT", 8083),
		GRPCPort: getEnvInt("ORDER_GRPC_PORT", 9093),

		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "tiki"),
			Password:     requireEnv("MYSQL_PASSWORD"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_orders"),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvDuration("MYSQL_MAX_LIFETIME", 5*time.Minute),
			Timeout:      getEnvDuration("MYSQL_TIMEOUT", 5*time.Second),
		},

		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 3),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE", 20),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},

		Kafka: KafkaConfig{
			Brokers:       getEnvSlice("KAFKA_BROKERS", ","),
			TopicPrefix:   getEnv("KAFKA_TOPIC_PREFIX", "tiki.orders"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "tiki-order-service"),
			DLQTopic:      getEnv("KAFKA_DLQ_TOPIC", "tiki.orders.dlq"),
		},

		JWT: JWTConfig{
			AccessSecret: requireEnv("JWT_ACCESS_SECRET"),
			AccessTTL:    getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			Issuer:       getEnv("JWT_ISSUER", "tiki-auth"),
			Audience:     getEnv("JWT_AUDIENCE", "tiki-clone"),
		},

		Order: OrderConfig{
			DefaultCurrency:        getEnv("ORDER_DEFAULT_CURRENCY", "SGD"),
			PaymentTimeout:         getEnvDuration("ORDER_PAYMENT_TIMEOUT", 30*time.Minute),
			IdempotencyKeyTTL:      getEnvDuration("ORDER_IDEMPOTENCY_TTL", 24*time.Hour),
			SnapshotRetentionDays:  getEnvInt("ORDER_SNAPSHOT_RETENTION_DAYS", 365),
			ReconciliationInterval: getEnvDuration("ORDER_RECONCILIATION_INTERVAL", 5*time.Minute),
			MaxCancelRetryCount:    getEnvInt("ORDER_MAX_CANCEL_RETRY", 3),
		},

		Idempotency: IdempotencyConfig{
			TTL: getEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-order"),
			TraceRatio:  getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},

		Audit: AuditConfig{
			Enabled:       getEnvBool("AUDIT_ENABLED", true),
			FlushInterval: getEnvDuration("AUDIT_FLUSH_INTERVAL", 5*time.Second),
			BatchSize:     getEnvInt("AUDIT_BATCH_SIZE", 100),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func requireEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	log.Fatalf("required environment variable %s is not set", key)
	return ""
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

func getEnvSlice(key, sep string) []string {
	val := getEnv(key, "")
	if val == "" {
		return []string{"localhost:9092"}
	}
	return strings.Split(val, sep)
}
