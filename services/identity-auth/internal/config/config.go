package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName string
	AppEnv  string
	LogLevel string
	HTTPPort int

	MySQL MySQLConfig
	Redis RedisConfig

	JWT JWTConfig
	RateLimit RateLimitConfig
	AccountLockout AccountLockoutConfig
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
	Addr string
	Password string
	DB int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	RSAPrivateKey string
	RSAPublicKey  string
}

type RateLimitConfig struct {
	LoginMaxAttempts    int
	LoginWindow         time.Duration
	IPMaxAttempts       int
	IPWindow            time.Duration
}

type AccountLockoutConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "identity-auth"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("AUTH_HTTP_PORT", 8080),

		MySQL: MySQLConfig{
			Host:         requireEnv("MYSQL_HOST"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         requireEnv("MYSQL_USER"),
			Password:     requireEnv("MYSQL_PASSWORD"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_auth"),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvDuration("MYSQL_MAX_LIFETIME", 5*time.Minute),
		},

		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},

		JWT: JWTConfig{
			AccessSecret:  requireEnv("JWT_ACCESS_SECRET"),
			RefreshSecret: requireEnv("JWT_REFRESH_SECRET"),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
			RSAPrivateKey: getEnv("JWT_RSA_PRIVATE_KEY", ""),
			RSAPublicKey:  getEnv("JWT_RSA_PUBLIC_KEY", ""),
		},

		RateLimit: RateLimitConfig{
			LoginMaxAttempts: getEnvInt("RATE_LIMIT_LOGIN_MAX", 5),
			LoginWindow:      getEnvDuration("RATE_LIMIT_LOGIN_WINDOW", 5*time.Minute),
			IPMaxAttempts:    getEnvInt("RATE_LIMIT_IP_MAX", 20),
			IPWindow:         getEnvDuration("RATE_LIMIT_IP_WINDOW", 1*time.Minute),
		},

		AccountLockout: AccountLockoutConfig{
			MaxFailedAttempts: getEnvInt("ACCOUNT_LOCKOUT_MAX", 5),
			LockoutDuration:   getEnvDuration("ACCOUNT_LOCKOUT_DURATION", 15*time.Minute),
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