package config
import ("os"; "strconv"; "time")
type Config struct { AppName string; AppEnv string; LogLevel string; HTTPPort int; GRPCPort int; Redis RedisConfig; Kafka KafkaConfig; Elasticsearch ESConfig; OpenTelemetry OTELConfig }
type RedisConfig struct { Addr string; Password string; DB int; PoolSize int; MinIdleConns int; DialTimeout time.Duration; ReadTimeout time.Duration; WriteTimeout time.Duration; MaxRetries int }
type KafkaConfig struct { Brokers []string }
type ESConfig struct { Addresses []string; Username string; Password string }
type OTELConfig struct { Endpoint string; ServiceName string; TraceRatio float64 }
func Load() *Config {
	return &Config{AppName: getEnv("APP_NAME", "tiki-search"), AppEnv: getEnv("APP_ENV", "development"), LogLevel: getEnv("LOG_LEVEL", "info"), HTTPPort: getEnvInt("SEARCH_HTTP_PORT", 8080), GRPCPort: getEnvInt("SEARCH_GRPC_PORT", 9090),
		Redis: RedisConfig{Addr: getEnv("REDIS_ADDR", "localhost:6379"), Password: getEnv("REDIS_PASSWORD", ""), DB: getEnvInt("REDIS_DB", 0), PoolSize: getEnvInt("REDIS_POOL_SIZE", 100), MinIdleConns: getEnvInt("REDIS_MIN_IDLE", 20), DialTimeout: getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second), ReadTimeout: getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second), WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second), MaxRetries: getEnvInt("REDIS_MAX_RETRIES", 3)},
		Kafka: KafkaConfig{Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")}},
		Elasticsearch: ESConfig{Addresses: []string{getEnv("ES_ADDRESSES", "http://localhost:9200")}, Username: getEnv("ES_USERNAME", ""), Password: getEnv("ES_PASSWORD", "")},
		OpenTelemetry: OTELConfig{Endpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"), ServiceName: getEnv("OTEL_SERVICE_NAME", "tiki-search"), TraceRatio: getEnvFloat("OTEL_TRACES_SAMPLER_ARG", 0.1)}}
}
func getEnv(k, f string) string { if v := os.Getenv(k); v != "" { return v }; return f }
func getEnvInt(k string, f int) int { if v := os.Getenv(k); v != "" { if i, e := strconv.Atoi(v); e == nil { return i } }; return f }
func getEnvDuration(k string, f time.Duration) time.Duration { if v := os.Getenv(k); v != "" { if d, e := time.ParseDuration(v); e == nil { return d } }; return f }
func getEnvFloat(k string, f float64) float64 { if v := os.Getenv(k); v != "" { if fl, e := strconv.ParseFloat(v, 64); e == nil { return fl } }; return f }
