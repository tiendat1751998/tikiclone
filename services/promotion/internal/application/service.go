package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/promotion/internal/domain"
	"github.com/tikiclone/tiki/services/promotion/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/promotion/internal/metrics"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type PromotionService struct {
	voucherRepo    domain.VoucherRepository
	redemptionRepo domain.VoucherRedemptionRepository
	campaignRepo   domain.CampaignRepository
	pricingRepo    domain.PricingRuleRepository
	eligibilityRepo domain.EligibilityRuleRepository
	stackingRepo   domain.StackingRuleRepository
	redis          *redis.Store
	publisher      EventPublisher
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}

func NewPromotionService(
	voucherRepo domain.VoucherRepository,
	redemptionRepo domain.VoucherRedemptionRepository,
	campaignRepo domain.CampaignRepository,
	pricingRepo domain.PricingRuleRepository,
	eligibilityRepo domain.EligibilityRuleRepository,
	stackingRepo domain.StackingRuleRepository,
	redisStore *redis.Store,
	publisher EventPublisher,
) *PromotionService {
	return &PromotionService{
		voucherRepo: voucherRepo, redemptionRepo: redemptionRepo,
		campaignRepo: campaignRepo, pricingRepo: pricingRepo,
		eligibilityRepo: eligibilityRepo, stackingRepo: stackingRepo,
		redis: redisStore, publisher: publisher,
	}
}

// ValidateVoucher checks if a voucher is valid for the given context
func (s *PromotionService) ValidateVoucher(ctx context.Context, code, userID string, subtotal int64, shopID, categoryID, sku, region, paymentMethod string) (*domain.Voucher, error) {
	ctx, span := otel.Tracer("tiki-promotion").Start(ctx, "promotion.validate_voucher")
	defer span.End()

	voucher, err := s.voucherRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if voucher == nil {
		return nil, domain.ErrVoucherInvalid
	}

	if err := voucher.CanRedeem(userID, subtotal, shopID, categoryID, sku, region, paymentMethod); err != nil {
		return nil, err
	}

	// Check per-user limit
	userCount, err := s.redemptionRepo.CountByUserAndVoucher(ctx, userID, voucher.ID)
	if err != nil {
		observability.LogWithTrace(ctx).Warn("failed to check user redemption count", zap.Error(err))
	} else if userCount >= voucher.PerUserLimit {
		return nil, fmt.Errorf("%w: user limit %d reached", domain.ErrVoucherUserLimit, voucher.PerUserLimit)
	}

	metrics.ValidationsTotal.WithLabelValues("success").Inc()
	return voucher, nil
}

// RedeemVoucher redeems a voucher for a user/order
func (s *PromotionService) RedeemVoucher(ctx context.Context, code, userID, orderID, idempotencyKey string, subtotal int64, shopID, categoryID, sku, region, paymentMethod string) (*domain.PromotionResult, error) {
	ctx, span := otel.Tracer("tiki-promotion").Start(ctx, "promotion.redeem_voucher")
	defer span.End()

	tx, err := s.voucherRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Idempotency check (inside tx with FOR UPDATE)
	if idempotencyKey != "" {
		existing, err := s.redemptionRepo.FindByIdempotencyKeyInTx(ctx, tx, idempotencyKey)
		if err == nil && existing != nil {
			metrics.IdempotentRequests.Inc()
			tx.Commit()
			return &domain.PromotionResult{
				VoucherID: existing.VoucherID, DiscountAmt: existing.DiscountAmt,
			}, nil
		}
	}

	// Lock voucher row for atomic read-check-write
	voucher, err := s.voucherRepo.FindByCodeForUpdate(ctx, tx, code)
	if err != nil {
		return nil, fmt.Errorf("find voucher: %w", err)
	}
	if voucher == nil {
		return nil, domain.ErrVoucherInvalid
	}

	if err := voucher.CanRedeem(userID, subtotal, shopID, categoryID, sku, region, paymentMethod); err != nil {
		return nil, err
	}

	// Check per-user limit inside transaction
	userCount, err := s.redemptionRepo.CountByUserAndVoucherInTx(ctx, tx, userID, voucher.ID)
	if err != nil {
		return nil, fmt.Errorf("per-user limit check: %w", err)
	}
	if userCount >= voucher.PerUserLimit {
		return nil, fmt.Errorf("%w: user limit %d reached", domain.ErrVoucherUserLimit, voucher.PerUserLimit)
	}

	discount := voucher.CalculateDiscount(subtotal)

	redemption := &domain.VoucherRedemption{
		ID:             uuid.New().String(),
		VoucherID:      voucher.ID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmt:    discount,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}

	if err := s.redemptionRepo.CreateInTx(ctx, tx, redemption); err != nil {
		return nil, fmt.Errorf("create redemption: %w", err)
	}

	// Atomic increment with exhaustion guard
	if err := s.voucherRepo.IncrementUsageInTx(ctx, tx, voucher.ID, voucher.UsageLimit); err != nil {
		return nil, fmt.Errorf("increment usage: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Post-commit side-effects (best-effort)
	if s.redis != nil {
		s.redis.IncrementVoucherUsage(ctx, voucher.ID)
		s.redis.IncrementUserVoucherUsage(ctx, userID, voucher.ID)
	}
	if s.publisher != nil {
		s.publisher.Publish(ctx, "voucher.redeemed", map[string]interface{}{
			"voucher_id": voucher.ID, "user_id": userID, "order_id": orderID, "discount": discount,
		})
	}

	metrics.RedeemTotal.Inc()
	return &domain.PromotionResult{
		VoucherID: voucher.ID, Type: voucher.Type, Description: voucher.Title,
		DiscountAmt: discount, FinalAmount: subtotal - discount,
	}, nil
}

// EvaluatePromotions evaluates all applicable promotions for a cart
func (s *PromotionService) EvaluatePromotions(ctx context.Context, userID string, subtotal int64, shopID, categoryID, sku, region, paymentMethod string) ([]*domain.PromotionResult, error) {
	ctx, span := otel.Tracer("tiki-promotion").Start(ctx, "promotion.evaluate")
	defer span.End()

	var results []*domain.PromotionResult

	// Get active campaigns
	campaigns, err := s.campaignRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	// Batch load pricing rules for all flash sale campaigns
	var flashIDs []string
	for _, c := range campaigns {
		if c.Type == domain.CampaignTypeFlashSale {
			flashIDs = append(flashIDs, c.ID)
		}
	}
	rulesByCampaign, err := s.pricingRepo.FindByCampaigns(ctx, flashIDs)
	if err != nil {
		observability.LogWithTrace(ctx).Warn("failed to batch load pricing rules", zap.Error(err))
	}

	for _, campaign := range campaigns {
		if campaign.Type == domain.CampaignTypeFlashSale {
			rules := rulesByCampaign[campaign.ID]
			for _, rule := range rules {
				if rule.IsActive {
					discount := s.evaluatePricingRule(rule, subtotal)
					if discount > 0 {
						results = append(results, &domain.PromotionResult{
							CampaignID: campaign.ID, Type: campaign.Type,
							Description: campaign.Name, DiscountAmt: discount,
							FinalAmount: subtotal - discount,
						})
					}
				}
			}
		}
	}

	metrics.EvaluationsTotal.Inc()
	return results, nil
}

func (s *PromotionService) evaluatePricingRule(rule *domain.PricingRule, subtotal int64) int64 {
	// Simplified: in production, parse condition JSON and evaluate
	switch rule.Action {
	case "percentage_discount":
		return subtotal * 10 / 100 // 10% default
	case "fixed_discount":
		return 5000 // $50 default
	}
	return 0
}

// GetActiveCampaigns returns all active campaigns
func (s *PromotionService) GetActiveCampaigns(ctx context.Context) ([]*domain.Campaign, error) {
	return s.campaignRepo.ListActive(ctx)
}

// CreateVoucher creates a new voucher
func (s *PromotionService) CreateVoucher(ctx context.Context, v *domain.Voucher) error {
	return s.voucherRepo.Create(ctx, v)
}

// CreateCampaign creates a new campaign
func (s *PromotionService) CreateCampaign(ctx context.Context, c *domain.Campaign) error {
	return s.campaignRepo.Create(ctx, c)
}
