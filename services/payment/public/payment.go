package paymentpublic

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	vnpay "github.com/tikiclone/tiki/services/payment/internal/infrastructure/vnpay"
	"github.com/tikiclone/tiki/services/payment/internal/application"
	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
)

type PaymentService = application.PaymentService

type AuthorizePaymentRequest = application.AuthorizePaymentRequest

type Payment = domain.Payment

type PaymentStatus = domain.PaymentStatus

type PaymentMethod = domain.PaymentMethod

const (
	PaymentMethodCreditCard   = domain.PaymentMethodCreditCard
	PaymentStatusAuthorized   = domain.PaymentStatusAuthorized
	PaymentMethodVNPayGateway = domain.PaymentMethodVNPayGateway
)

type FraudDetector = domain.FraudDetector

type FraudCheckResult = domain.FraudCheckResult

func NewFraudCheckResult(paymentID, userID string, riskScore int, isFraud bool) *FraudCheckResult {
	return domain.NewFraudCheckResult(paymentID, userID, riskScore, isFraud)
}

type Config = config.Config

type PaymentConfig = config.PaymentConfig

type RedisConfig = config.RedisConfig

type VNPayPaymentRequest = domain.VNPayPaymentRequest

func VNPayResponseCodes() map[string]string {
	return domain.VNPayResponseCodes
}

type VNPaySimulator = vnpay.Simulator

func NewPaymentRepository(db *sqlx.DB) *mysql.PaymentRepository {
	return mysql.NewPaymentRepository(db)
}

func NewRedisStore(client *redis.Client, cfg config.RedisConfig) *redisinfra.Store {
	return redisinfra.NewStore(client, cfg)
}

func NewPaymentService(cfg *config.Config, paymentRepo *mysql.PaymentRepository, redisStore *redisinfra.Store, kafkaProducer *kafka.Producer, fraudDetector domain.FraudDetector) *PaymentService {
	return application.NewPaymentService(cfg, paymentRepo, redisStore, kafkaProducer, fraudDetector)
}

func NewVNPaySimulator(cfg *config.Config, paymentRepo *mysql.PaymentRepository, redisStore *redisinfra.Store, kafkaProducer *kafka.Producer) *VNPaySimulator {
	return vnpay.NewSimulator(cfg, paymentRepo, redisStore, kafkaProducer)
}

func GenerateVNPaySecureHash(params map[string]string, secretKey string) string {
	return domain.GenerateVNPaySecureHash(params, secretKey)
}

func FormatVNPayDate(t time.Time) string {
	return domain.FormatVNPayDate(t)
}
