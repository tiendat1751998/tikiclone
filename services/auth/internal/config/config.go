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

	JWT JWTConfig

	Password PasswordConfig

	RateLimit RateLimitConfig

	Session SessionConfig

	Security SecurityConfig

	OpenTelemetry OTELConfig

	Audit AuditConfig
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
	AccessSecret     string
	RefreshSecret    string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	Issuer           string
	Audience         string
	ClockSkew        time.Duration
	RotationEnabled  bool
	BlacklistEnabled bool
}

type PasswordConfig struct {
	Algorithm string
	Cost      int
	SaltLen   int
	KeyLen    uint32
	Memory    uint32
	Time      uint32
	Threads   uint8
}

type RateLimitConfig struct {
	LoginMaxAttempts    int
	LoginWindow         time.Duration
	RegisterMaxPerIP    int
	RegisterWindow      time.Duration
	PasswordResetMax    int
	PasswordResetWindow time.Duration
	AccountLockout      int
	LockoutDuration     time.Duration
}

type SessionConfig struct {
	Driver             string
	MaxSessionsPerUser int
	SessionTTL         time.Duration
	IdleTimeout        time.Duration
	RefreshRotation    bool
}

type SecurityConfig struct {
	Argon2Memory         uint32
	Argon2Time           uint32
	Argon2Threads        uint8
	Argon2KeyLen         uint32
	SuspiciousIPTTL      time.Duration
	SuspiciousLoginCount int
	DeviceFingerprinting bool
	MaxDevicesPerUser    int
}

type OTELConfig struct {
	Endpoint    string
	ServiceName string
	TraceRatio  float64
}

type AuditConfig struct {
	Enabled       bool
	Driver        string
	LogLevel      string
	FlushInterval time.Duration
	BatchSize     int
	MaxBufferSize int
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "tiki-auth"),
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTPPort: getEnvInt("AUTH_HTTP_PORT", 8080),
		GRPCPort: getEnvInt("AUTH_GRPC_PORT", 9090),

		MySQL: MySQLConfig{
			Host:         requireEnv("MYSQL_HOST"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         requireEnv("MYSQL_USER"),
			Password:     requireEnv("MYSQL_PASSWORD"),
			Database:     getEnv("MYSQL_DATABASE", "tiki_auth"),
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

		JWT: JWTConfig{
			AccessSecret:     requireEnv("JWT_ACCESS_SECRET"),
			RefreshSecret:    requireEnv("JWT_REFRESH_SECRET"),
			AccessTTL:        getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:       getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
			Issuer:           getEnv("JWT_ISSUER", "tiki-auth"),
			Audience:         getEnv("JWT_AUDIENCE", "tiki-clone"),
			ClockSkew:        getEnvDuration("JWT_CLOCK_SKEW", 30*time.Second),
			RotationEnabled:  getEnvBool("JWT_ROTATION_ENABLED", true),
			BlacklistEnabled: getEnvBool("JWT_BLACKLIST_ENABLED", true),
		},

		Password: PasswordConfig{
			Algorithm: getEnv("PASSWORD_ALGORITHM", "bcrypt"),
			Cost:      getEnvInt("BCRYPT_COST", 12),
			SaltLen:   getEnvInt("PASSWORD_SALT_LEN", 16),
			KeyLen:    uint32(getEnvInt("ARGON2_KEY_LEN", 32)),
			Memory:    uint32(getEnvInt("ARGON2_MEMORY", 64*1024)),
			Time:      uint32(getEnvInt("ARGON2_TIME", 3)),
			Threads:   uint8(getEnvInt("ARGON2_THREADS", 4)),
		},

		RateLimit: RateLimitConfig{
			LoginMaxAttempts:    getEnvInt("LOGIN_MAX_ATTEMPTS", 5),
			LoginWindow:         getEnvDuration("LOGIN_WINDOW", 5*time.Minute),
			RegisterMaxPerIP:    getEnvInt("REGISTER_MAX_PER_IP", 3),
			RegisterWindow:      getEnvDuration("REGISTER_WINDOW", 1*time.Hour),
			PasswordResetMax:    getEnvInt("PASSWORD_RESET_MAX", 3),
			PasswordResetWindow: getEnvDuration("PASSWORD_RESET_WINDOW", 1*time.Hour),
			AccountLockout:      getEnvInt("ACCOUNT_LOCKOUT_THRESHOLD", 10),
			LockoutDuration:     getEnvDuration("ACCOUNT_LOCKOUT_DURATION", 30*time.Minute),
		},

		Session: SessionConfig{
			Driver:             getEnv("SESSION_DRIVER", "redis"),
			MaxSessionsPerUser: getEnvInt("MAX_SESSIONS_PER_USER", 10),
			SessionTTL:         getEnvDuration("SESSION_TTL", 24*time.Hour),
			IdleTimeout:        getEnvDuration("SESSION_IDLE_TIMEOUT", 30*time.Minute),
			RefreshRotation:    getEnvBool("SESSION_REFRESH_ROTATION", true),
		},

		Security: SecurityConfig{
			Argon2Memory:         uint32(getEnvInt("ARGON2_MEMORY", 64*1024)),
			Argon2Time:           uint32(getEnvInt("ARGON2_TIME", 3)),
			Argon2Threads:        uint8(getEnvInt("ARGON2_THREADS", 4)),
			Argon2KeyLen:         uint32(getEnvInt("ARGON2_KEY_LEN", 32)),
			SuspiciousIPTTL:      getEnvDuration("SUSPICIOUS_IP_TTL", 24*time.Hour),
			SuspiciousLoginCount: getEnvInt("SUSPICIOUS_LOGIN_COUNT", 3),
			DeviceFingerprinting: getEnvBool("DEVICE_FINGERPRINTING", true),
			MaxDevicesPerUser:    getEnvInt("MAX_DEVICES_PER_USER", 20),
		},

		OpenTelemetry: OTELConfig{
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-auth"),
			TraceRatio:  getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1),
		},

		Audit: AuditConfig{
			Enabled:       getEnvBool("AUDIT_ENABLED", true),
			Driver:        getEnv("AUDIT_DRIVER", "mysql"),
			LogLevel:      getEnv("AUDIT_LOG_LEVEL", "info"),
			FlushInterval: getEnvDuration("AUDIT_FLUSH_INTERVAL", 5*time.Second),
			BatchSize:     getEnvInt("AUDIT_BATCH_SIZE", 100),
			MaxBufferSize: getEnvInt("AUDIT_MAX_BUFFER_SIZE", 10000),
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
		return nil
	}
	return strings.Split(val, sep)
}
