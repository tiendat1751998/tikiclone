package config
import ("os"; "strconv"; "time")
type Config struct { AppName string; AppEnv string; LogLevel string; HTTPPort int; GRPCPort int; Redis RedisConfig; Kafka KafkaConfig; OpenTelemetry OTELConfig }
type RedisConfig struct { Addr string; Password string; DB int; PoolSize int; MinIdleConns int; DialTimeout time.Duration; ReadTimeout time.Duration; WriteTimeout time.Duration; MaxRetries int }
type KafkaConfig struct { Brokers []string }
type OTELConfig struct { Endpoint string; ServiceName string; TraceRatio float64 }
func Load() *Config {
	return &Config{AppName: getEnv("APP_NAME", "tiki-user-behavior"), AppEnv: getEnv("APP_ENV", "development"), LogLevel: getEnv("LOG_LEVEL", "info"), HTTPPort: getEnvInt("UB_HTTP_PORT", 8080), GRPCPort: getEnvInt("UB_GRPC_PORT", 9090),
		Redis: RedisConfig{Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", ""), DB: 0, PoolSize: 100, MinIdleConns: 20, DialTimeout: 5 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, MaxRetries: 3},
		Kafka: KafkaConfig{Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")}},
		OpenTelemetry: OTELConfig{Endpoint: "http://localhost:4318", ServiceName: "tiki-user-behavior", TraceRatio: 0.1}}
}
func getEnv(k, f string) string { if v := os.Getenv(k); v != "" { return v }; return f }
func getEnvInt(k string, f int) int { if v := os.Getenv(k); v != "" { if i, e := strconv.Atoi(v); e == nil { return i } }; return f }
