package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName  string
	Env      string
	Version  string
	HTTPPort string
	GRPCPort string

	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	OTEL     OTELConfig
	SMTP     SMTPConfig
	Twilio   TwilioConfig
	FCM      FCMConfig
	APNs     APNsConfig
	SendGrid SendGridConfig
}

type PostgresConfig struct {
	DSN             string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
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
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=true&loc=UTC&timeout=" + c.Timeout.String()
}

type KafkaConfig struct {
	Brokers          []string
	NotificationTopic string
	ConsumerGroup    string
}

type OTELConfig struct {
	Endpoint string
	Insecure bool
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

type FCMConfig struct {
	ServerKey string
}

type APNsConfig struct {
	KeyID   string
	TeamID  string
	KeyPath string
}

type SendGridConfig struct {
	APIKey string
	From   string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func Load() *Config {
	return &Config{
		AppName:  getEnv("APP_NAME", "notification"),
		Env:      getEnv("ENV", "development"),
		Version:  getEnv("VERSION", "0.0.1"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		GRPCPort: getEnv("GRPC_PORT", "9090"),
		Postgres: PostgresConfig{
			DSN:             getEnv("POSTGRES_DSN", "postgres://notification:notification@localhost:5432/notification?sslmode=disable"),
			MaxConns:        getEnvInt("POSTGRES_MAX_CONNS", 25),
			MinConns:        getEnvInt("POSTGRES_MIN_CONNS", 5),
			MaxConnLifetime: time.Duration(getEnvInt("POSTGRES_MAX_CONN_LIFETIME_MIN", 30)) * time.Minute,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers:           splitEnv(getEnv("KAFKA_BROKERS", "localhost:9092")),
			NotificationTopic: getEnv("KAFKA_NOTIFICATION_TOPIC", "notification.events"),
			ConsumerGroup:     getEnv("KAFKA_CONSUMER_GROUP", "notification-consumer-group"),
		},
		OTEL: OTELConfig{
			Endpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			Insecure: getEnv("OTEL_INSECURE", "true") == "true",
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.example.com"),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@tiki-clone.com"),
		},
		Twilio: TwilioConfig{
			AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
			FromNumber: getEnv("TWILIO_FROM_NUMBER", "+15005550006"),
		},
		FCM: FCMConfig{
			ServerKey: getEnv("FCM_SERVER_KEY", ""),
		},
		APNs: APNsConfig{
			KeyID:   getEnv("APNS_KEY_ID", ""),
			TeamID:  getEnv("APNS_TEAM_ID", ""),
			KeyPath: getEnv("APNS_KEY_PATH", ""),
		},
		SendGrid: SendGridConfig{
			APIKey: getEnv("SENDGRID_API_KEY", ""),
			From:   getEnv("SENDGRID_FROM", "noreply@tiki-clone.com"),
		},
	}
}

func splitEnv(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}
