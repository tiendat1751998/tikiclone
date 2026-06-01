package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the product service
type Config struct {
	ServiceName string
	AppEnv      string
	LogLevel    string
	HTTPPort    int
	GRPCPort    int

	MySQL MySQLConfig
	Redis RedisConfig
	Kafka KafkaConfig
	OpenSearch OpenSearchConfig
	OpenTelemetry OTELConfig
	Server ServerConfig
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
}

func (c MySQLConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=true&loc=UTC"
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

type KafkaConfig struct {
	Brokers []string
}

type OpenSearchConfig struct {
	Addresses []string
	Username  string
	Password  string
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServiceName: getEnv("APP_NAME", "product-service"),
		AppEnv:      getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HTTPPort:    getEnvInt("HTTP_PORT", 8080),
		GRPCPort:    getEnvInt("GRPC_PORT", 9090),

		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "tiki"),
			Password:     getEnv("MYSQL_PASSWORD", "tiki_dev"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_product"),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvDuration("MYSQL_MAX_LIFETIME", 5*time.Minute),
		},

		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE", 20),
		},

		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		},

		OpenSearch: OpenSearchConfig{
			Addresses: []string{getEnv("OPENSEARCH_ADDR", "http://localhost:9200")},
			Username:  getEnv("OPENSEARCH_USER", "admin"),
			Password:  getEnv("OPENSEARCH_PASSWORD", "admin"),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "product-service"),
			TraceRatio:  getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},

		Server: ServerConfig{
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
	}
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

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
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
