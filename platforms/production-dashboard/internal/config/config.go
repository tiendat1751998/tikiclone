package config

import (
	"os"
	"strconv"
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

	Dashboard DashboardConfig

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

type KafkaConfig struct {
	Brokers []string
}

type DashboardConfig struct {
	ServicesPollInterval  time.Duration
	IncidentRetentionDays int
	MetricsRetentionDays  int
	MaxAlertRules         int
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "production-dashboard"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("DASHBOARD_HTTP_PORT", 8080),
		GRPCPort: getEnvInt("DASHBOARD_GRPC_PORT", 9090),

		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "tiki"),
			Password:     getEnv("MYSQL_PASSWORD", "tiki_dev"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_dashboard"),
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

		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		},

		Dashboard: DashboardConfig{
			ServicesPollInterval:  getEnvDuration("SERVICES_POLL_INTERVAL", 30*time.Second),
			IncidentRetentionDays: getEnvInt("INCIDENT_RETENTION_DAYS", 90),
			MetricsRetentionDays:  getEnvInt("METRICS_RETENTION_DAYS", 30),
			MaxAlertRules:         getEnvInt("MAX_ALERT_RULES", 500),
		},

		JWT: JWTConfig{
			AccessSecret: getEnv("JWT_ACCESS_SECRET", ""),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "production-dashboard"),
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
