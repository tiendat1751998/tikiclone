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
	JWT   JWTConfig

	Recommendation RecConfig
	OpenTelemetry  OTELConfig
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

type JWTConfig struct {
	AccessSecret string
	AccessTTL    time.Duration
	Issuer       string
	Audience     string
}

type RecConfig struct {
	DefaultLimit   int
	MaxLimit       int
	TrendingWindow time.Duration
	CacheTTL       time.Duration
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "tiki-rec-vector"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("REC_HTTP_PORT", 8096),
		GRPCPort: getEnvInt("REC_GRPC_PORT", 9096),

		MySQL: MySQLConfig{
			Host:         getEnv("MYSQL_HOST", "localhost"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "tiki"),
			Password:     requireEnv("MYSQL_PASSWORD"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_recommendation"),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 50),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 25),
			MaxLifetime:  getEnvDuration("MYSQL_MAX_LIFETIME", 30*time.Minute),
			Timeout:      getEnvDuration("MYSQL_TIMEOUT", 2*time.Second),
		},

		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 8),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 200),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE", 30),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 150*time.Millisecond),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 150*time.Millisecond),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},

		JWT: JWTConfig{
			AccessSecret: requireEnv("JWT_ACCESS_SECRET"),
			AccessTTL:    getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			Issuer:       getEnv("JWT_ISSUER", "tiki-auth"),
			Audience:     getEnv("JWT_AUDIENCE", "tiki-clone"),
		},

		Recommendation: RecConfig{
			DefaultLimit:   getEnvInt("REC_DEFAULT_LIMIT", 20),
			MaxLimit:       getEnvInt("REC_MAX_LIMIT", 100),
			TrendingWindow: getEnvDuration("REC_TRENDING_WINDOW", 24*time.Hour),
			CacheTTL:       getEnvDuration("REC_CACHE_TTL", 5*time.Minute),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-rec-vector"),
			TraceRatio:  getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
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
		return nil
	}
	return strings.Split(val, sep)
}
