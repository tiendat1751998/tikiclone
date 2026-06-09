package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/services/payment/internal/config"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/kafka"
	"github.com/tikiclone/tiki/services/payment/internal/infrastructure/mysql"
	"github.com/tikiclone/tiki/services/payment/internal/metrics"
	redisinfra "github.com/tikiclone/tiki/services/payment/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type PaymentService struct {
	cfg           *config.Config
	paymentRepo   *mysql.PaymentRepository
	redisStore    *redisinfra.Store
	kafkaProducer *kafka.Producer
	fraudDetector domain.FraudDetector
}

func NewPaymentService(cfg *config.Config, paymentRepo *mysql.PaymentRepository, redisStore *redisinfra.Store, kafkaProducer *kafka.Producer, fraudDetector domain.FraudDetector) *PaymentService {
	return &PaymentService{cfg: cfg, paymentRepo: paymentRepo, redisStore: redisStore, kafkaProducer: kafkaProducer, fraudDetector: fraudDetector}
}

type AuthorizePaymentRequest struct {
	OrderID        string              `json:"order_id" validate:"required"`
	UserID         string              `json:"user_id" validate:"required"`
	Amount         int64               `json:"amount" validate:"required"`
	Currency       string              `json:"currency"`
	PaymentMethod  domain.PaymentMethod `json:"payment_method" validate:"required"`
	IdempotencyKey string              `json:"idempotency_key" validate:"required"`
	Metadata       json.RawMessage     `json:"metadata,omitempty"`
}

func (s *PaymentService) AuthorizePayment(ctx context.Context, req *AuthorizePaymentRequest) (*domain.Payment, error) {
	ctx, span := otel.Tracer("tiki-payment").Start(ctx, "PaymentService.AuthorizePayment")
	defer span.End()

	start := time.Now()
	defer func() { metrics.PaymentAuthorizationLatency.WithLabelValues(s.cfg.Payment.DefaultPSP).Observe(time.Since(start).Seconds()) }()

	lockToken, locked, err := s.redisStore.AcquirePaymentLock(ctx, req.OrderID, 5*time.Second)
	if err != nil || !locked {
		return nil, fmt.Errorf("failed to acquire payment lock")
	}
	defer s.redisStore.ReleasePaymentLock(ctx, req.OrderID, lockToken)

	// Idempotency check: Redis first, skip DB if Redis hits
	if req.IdempotencyKey != "" {
		existingID, err := s.redisStore.CheckIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existingID != "" {
			metrics.DuplicatePreventionCount.Inc()
			return s.paymentRepo.FindByID(ctx, existingID)
		}
		// Only check DB if Redis miss
		existing, err := s.paymentRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey)
		if err == nil && existing != nil {
			metrics.DuplicatePreventionCount.Inc()
			return existing, nil
		}
	}

	// Check if payment already exists for this order
	existingPayment, err := s.paymentRepo.FindByOrderID(ctx, req.OrderID)
	if err == nil && existingPayment != nil && !existingPayment.IsTerminal() {
		span.SetStatus(codes.Error, "double charge detected")
		return nil, domain.ErrDoubleChargeDetected
	}

	currency := req.Currency
	if currency == "" { currency = "SGD" }

	payment := domain.NewPayment(req.OrderID, req.UserID, req.Amount, currency, req.PaymentMethod, s.cfg.Payment.DefaultPSP, req.IdempotencyKey)
	if len(req.Metadata) > 0 {
		payment.Metadata = &req.Metadata
	}

	// Fraud check (non-blocking save)
	fraudResult, err := s.fraudDetector.Assess(ctx, payment.ID, req.UserID, req.Amount, req.PaymentMethod)
	if err != nil {
		observability.LogWithTrace(ctx).Error("fraud detection failed", zap.Error(err))
	} else {
		// Save fraud check async — don't block payment creation
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := s.paymentRepo.SaveFraudCheck(bgCtx, fraudResult); err != nil {
				zap.L().Warn("failed to save fraud check", zap.Error(err))
			}
		}()
		if fraudResult.IsFraud {
			metrics.FraudDetectedCount.Inc()
			return nil, domain.ErrFraudDetected
		}
	}

	// Simulate PSP authorization
	payment.PSPTransactionID = fmt.Sprintf("psp-tx-%s", payment.ID[:8])
	if err := payment.TransitionTo(domain.PaymentStatusAuthorized); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Store idempotency in background (best-effort)
	if req.IdempotencyKey != "" {
		rec := domain.NewIdempotencyRecord(req.IdempotencyKey, payment.ID, s.cfg.Payment.IdempotencyTTL)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := s.paymentRepo.SaveIdempotencyKey(bgCtx, rec); err != nil {
				zap.L().Warn("failed to save idempotency key", zap.Error(err))
			}
			if err := s.redisStore.StoreIdempotencyKey(bgCtx, req.IdempotencyKey, payment.ID, s.cfg.Payment.IdempotencyTTL); err != nil {
				zap.L().Warn("failed to store idempotency key in redis", zap.Error(err))
			}
		}()
	}

	// Publish event via outbox (reliable) — Kafka is now async so this is fast
	event := domain.NewPaymentEvent(payment, domain.EventPaymentAuthorized, req.Metadata)
	if payload, err := json.Marshal(event); err == nil {
		if err := s.paymentRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("payment", payment.ID, string(domain.EventPaymentAuthorized), payload)); err != nil {
			observability.LogWithTrace(ctx).Error("failed to save outbox event", zap.Error(err))
		}
	}
	// Kafka publish is now async — fire and forget
	if s.kafkaProducer != nil {
		go s.kafkaProducer.PublishEvent(context.Background(), event)
	}

	metrics.PaymentsAuthorizedTotal.WithLabelValues(s.cfg.Payment.DefaultPSP, string(req.PaymentMethod)).Inc()
	metrics.ActivePayments.WithLabelValues(string(domain.PaymentStatusAuthorized)).Inc()

	span.SetAttributes(attribute.String("payment_id", payment.ID), attribute.Int64("amount", payment.Amount))
	zap.L().Info("payment authorized", zap.String("payment_id", payment.ID), zap.String("order_id", req.OrderID))
	return payment, nil
}

func (s *PaymentService) CapturePayment(ctx context.Context, paymentID, actorID string) (*domain.Payment, error) {
	ctx, span := otel.Tracer("tiki-payment").Start(ctx, "PaymentService.CapturePayment")
	defer span.End()

	start := time.Now()
	defer func() { metrics.PaymentCaptureLatency.WithLabelValues(s.cfg.Payment.DefaultPSP).Observe(time.Since(start).Seconds()) }()

	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil { return nil, err }
	if payment.UserID != actorID { return nil, domain.ErrUnauthorized }

	if err := payment.TransitionTo(domain.PaymentStatusCaptured); err != nil {
		return nil, err
	}

	if err := s.paymentRepo.UpdateStatus(ctx, paymentID, domain.PaymentStatusCaptured, payment.Version-1); err != nil {
		return nil, err
	}

	event := domain.NewPaymentEvent(payment, domain.EventPaymentCaptured, nil)
	if payload, err := json.Marshal(event); err == nil {
		if err := s.paymentRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("payment", payment.ID, string(domain.EventPaymentCaptured), payload)); err != nil {
			observability.LogWithTrace(ctx).Error("failed to save capture outbox event",
				zap.String("payment_id", payment.ID), zap.Error(err))
		}
	}
	if s.kafkaProducer != nil {
		go s.kafkaProducer.PublishEvent(context.Background(), event)
	}

	metrics.PaymentsCapturedTotal.WithLabelValues(s.cfg.Payment.DefaultPSP).Inc()
	return payment, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID, reason, idempotencyKey string, amount int64, actorID string) (*domain.Refund, error) {
	ctx, span := otel.Tracer("tiki-payment").Start(ctx, "PaymentService.RefundPayment")
	defer span.End()

	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil { return nil, err }
	if payment.UserID != actorID { return nil, domain.ErrUnauthorized }

	if payment.Status != domain.PaymentStatusCaptured && payment.Status != domain.PaymentStatusPartialRefund {
		return nil, domain.ErrRefundNotAllowed
	}

	// [FIX A4] Validate amount is positive and doesn't exceed remaining
	if amount <= 0 {
		return nil, fmt.Errorf("refund amount must be positive, got %d", amount)
	}
	if amount > payment.RemainingAmount() {
		return nil, domain.ErrRefundAmountExceeded
	}

	refund := domain.NewRefund(paymentID, payment.OrderID, payment.Currency, reason, idempotencyKey, amount)
	if err := s.paymentRepo.SaveRefund(ctx, refund); err != nil {
		return nil, err
	}

	payment.AmountRefunded += amount
	newStatus := domain.PaymentStatusPartialRefund
	if payment.AmountRefunded >= payment.Amount {
		newStatus = domain.PaymentStatusRefunded
	}

	// [FIX A4] MUST check TransitionTo error
	if err := payment.TransitionTo(newStatus); err != nil {
		return nil, fmt.Errorf("failed to transition payment status: %w", err)
	}

	// [FIX A4] MUST check Update error
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	event := domain.NewPaymentEvent(payment, domain.EventPaymentRefunded, nil)
	if payload, err := json.Marshal(event); err == nil {
		if err := s.paymentRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("payment", payment.ID, string(domain.EventPaymentRefunded), payload)); err != nil {
			observability.LogWithTrace(ctx).Error("failed to save refund outbox event",
				zap.String("payment_id", payment.ID), zap.Error(err))
		}
	}
	if s.kafkaProducer != nil {
		go s.kafkaProducer.PublishEvent(context.Background(), event)
	}

	metrics.RefundsProcessed.WithLabelValues("success").Inc()
	return refund, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, paymentID)
}

// VNPayAuthorizePayment creates a payment and returns VNPay payment URL for frontend redirect
func (s *PaymentService) VNPayAuthorizePayment(ctx context.Context, orderID, userID string, amount int64, txnRef, clientIP, returnURL string) (*domain.VNPayPaymentRequest, string, error) {
	currency := "VND"
	paymentMethod := domain.PaymentMethodVNPayGateway

	payment := domain.NewPayment(orderID, userID, amount, currency, paymentMethod, "vnpay", txnRef)

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, "", fmt.Errorf("failed to create payment: %w", err)
	}

	req := &domain.VNPayPaymentRequest{
		Version:    domain.VNPayVersion,
		Command:    "pay",
		TmnCode:    s.cfg.Payment.VNPayTmnCode,
		Amount:     amount,
		CurrCode:   currency,
		Locale:     "vn",
		TxnRef:     txnRef,
		OrderInfo:  fmt.Sprintf("Order %s", orderID),
		ReturnUrl:  returnURL,
		IpAddr:     clientIP,
		CreateDate: domain.FormatVNPayDate(time.Now()),
	}

	return req, payment.ID, nil
}

// ProcessVNPayCallback handles the VNPay return callback
func (s *PaymentService) ProcessVNPayCallback(ctx context.Context, params map[string]string) error {
	txnRef := params["vnp_TxnRef"]
	responseCode := params["vnp_ResponseCode"]
	transactionNo := params["vnp_TransactionNo"]
	transactionStatus := params["vnp_TransactionStatus"]

	payment, err := s.paymentRepo.FindByOrderID(ctx, txnRef)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	success := responseCode == "00" && transactionStatus == "00"

	var newStatus domain.PaymentStatus
	if success {
		newStatus = domain.PaymentStatusCaptured
		payment.PSPTransactionID = transactionNo
	} else {
		newStatus = domain.PaymentStatusFailed
		if msg, ok := domain.VNPayResponseCodes[responseCode]; ok {
			payment.FailureReason = msg
		}
	}

	if err := payment.TransitionTo(newStatus); err != nil {
		return fmt.Errorf("failed to transition payment: %w", err)
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	eventType := domain.EventPaymentCaptured
	if !success {
		eventType = domain.EventPaymentFailed
	}

	event := domain.NewPaymentEvent(payment, eventType, nil)
	payload, _ := json.Marshal(event)
	s.paymentRepo.SaveOutboxEvent(ctx, domain.NewOutboxEvent("payment", payment.ID, string(eventType), payload))
	if s.kafkaProducer != nil {
		s.kafkaProducer.PublishEvent(ctx, event)
	}

	return nil
}

// [FIX A3] Webhook handler - now properly processes PSP events
func (s *PaymentService) HandleWebhook(ctx context.Context, pspProvider, eventType string, payload []byte, signature, idempotencyKey string) error {
	ctx, span := otel.Tracer("tiki-payment").Start(ctx, "PaymentService.HandleWebhook")
	defer span.End()

	start := time.Now()
	defer func() { metrics.WebhookLatency.WithLabelValues(pspProvider, eventType).Observe(time.Since(start).Seconds()) }()

	// Replay protection
	isReplay, err := s.redisStore.CheckWebhookReplay(ctx, idempotencyKey)
	if err == nil && isReplay {
		metrics.ReplayAttackCount.Inc()
		return domain.ErrWebhookReplayDetected
	}

	// Verify signature
	if !domain.VerifyWebhookSignature(payload, signature, s.cfg.Payment.WebhookSecret) {
		return domain.ErrInvalidWebhookSignature
	}

	// Store webhook event
	webhookEvent := domain.NewWebhookEvent(pspProvider, eventType, payload, signature, idempotencyKey)
	if err := s.paymentRepo.SaveWebhookEvent(ctx, webhookEvent); err != nil {
		return err
	}
	s.redisStore.MarkWebhookProcessed(ctx, idempotencyKey, 24*time.Hour)

	// [FIX A3] Actually process the webhook event
	switch eventType {
	case "payment.authorized":
		var eventData map[string]interface{}
		if err := json.Unmarshal(payload, &eventData); err != nil {
			return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
		}
		paymentID, _ := eventData["payment_id"].(string)
		if paymentID != "" {
			if err := s.markPaymentAuthorized(ctx, paymentID); err != nil {
				return fmt.Errorf("failed to mark payment authorized: %w", err)
			}
		}
	case "payment.captured":
		var eventData map[string]interface{}
		if err := json.Unmarshal(payload, &eventData); err != nil {
			return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
		}
		paymentID, _ := eventData["payment_id"].(string)
		if paymentID != "" {
			if _, err := s.CapturePayment(ctx, paymentID, "psp_webhook"); err != nil {
				return fmt.Errorf("failed to capture payment: %w", err)
			}
		}
	case "payment.failed":
		var eventData map[string]interface{}
		if err := json.Unmarshal(payload, &eventData); err != nil {
			return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
		}
		paymentID, _ := eventData["payment_id"].(string)
		if paymentID != "" {
			if err := s.markPaymentFailed(ctx, paymentID); err != nil {
				return fmt.Errorf("failed to mark payment failed: %w", err)
			}
		}
	default:
		observability.LogWithTrace(ctx).Warn("unknown webhook event type", zap.String("type", eventType))
	}

	metrics.WebhookProcessed.WithLabelValues(pspProvider, eventType).Inc()
	return nil
}

// markPaymentAuthorized updates payment status from PSP webhook
func (s *PaymentService) markPaymentAuthorized(ctx context.Context, paymentID string) error {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil { return err }
	if payment == nil { return fmt.Errorf("payment not found: %s", paymentID) }
	if err := payment.TransitionTo(domain.PaymentStatusAuthorized); err != nil { return err }
	return s.paymentRepo.UpdateStatus(ctx, paymentID, domain.PaymentStatusAuthorized, payment.Version-1)
}

// markPaymentFailed updates payment status from PSP webhook
func (s *PaymentService) markPaymentFailed(ctx context.Context, paymentID string) error {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil { return err }
	if payment == nil { return fmt.Errorf("payment not found: %s", paymentID) }
	if err := payment.TransitionTo(domain.PaymentStatusFailed); err != nil { return err }
	return s.paymentRepo.UpdateStatus(ctx, paymentID, domain.PaymentStatusFailed, payment.Version-1)
}

// [FIX A1] ProcessOutboxEvents - now properly logs errors and tracks failed events
func (s *PaymentService) ProcessOutboxEvents(ctx context.Context) error {
	events, err := s.paymentRepo.GetUnprocessedOutboxEvents(ctx, 100)
	if err != nil { return err }

	for _, event := range events {
		// Mark as processing first (prevents duplicate processing)
		if err := s.paymentRepo.MarkOutboxEventProcessing(ctx, event.ID); err != nil {
			observability.LogWithTrace(ctx).Error("failed to mark outbox event as processing",
				zap.String("event_id", event.ID), zap.Error(err))
			continue
		}

		var paymentEvent domain.PaymentEvent
		if err := json.Unmarshal(event.Payload, &paymentEvent); err != nil {
			observability.LogWithTrace(ctx).Error("failed to unmarshal outbox event payload",
				zap.String("event_id", event.ID), zap.Error(err))
			s.paymentRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			continue
		}

		if err := s.kafkaProducer.PublishEvent(ctx, &paymentEvent); err != nil {
			observability.LogWithTrace(ctx).Error("failed to publish outbox event to Kafka",
				zap.String("event_id", event.ID), zap.Error(err))
			s.paymentRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error())
			continue
		}

		if err := s.paymentRepo.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			observability.LogWithTrace(ctx).Error("failed to mark outbox event as processed",
				zap.String("event_id", event.ID), zap.Error(err))
		}
	}
	return nil
}
