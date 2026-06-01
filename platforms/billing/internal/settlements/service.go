package settlements

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/events"
	"github.com/tikiclone/tiki/platforms/billing/internal/metrics"
)

type Service struct {
	settlementRepo domain.SettlementRepository
	publisher      events.Publisher
	postTxn        func(ctx context.Context, txnType domain.TransactionType, debit, credit string, amount int64, currency, description string) (*domain.Transaction, error)
}

func NewService(sr domain.SettlementRepository, pub events.Publisher, postTxn func(ctx context.Context, txnType domain.TransactionType, debit, credit string, amount int64, currency, description string) (*domain.Transaction, error)) *Service {
	return &Service{settlementRepo: sr, publisher: pub, postTxn: postTxn}
}

func (s *Service) CreateSettlement(ctx context.Context, merchantID string, amount, feeAmount int64, currency string, periodStart, periodEnd time.Time) (*domain.Settlement, error) {
	settlement := &domain.Settlement{
		ID:          uuid.New().String(),
		MerchantID:  merchantID,
		Amount:      amount,
		Currency:    currency,
		FeeAmount:   feeAmount,
		NetAmount:   amount - feeAmount,
		Status:      domain.SettlementPending,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   time.Now(),
	}
	if err := s.settlementRepo.Create(ctx, settlement); err != nil {
		return nil, fmt.Errorf("create settlement: %w", err)
	}
	s.publisher.Publish(ctx, events.EventSettlementCreated, settlement)
	return settlement, nil
}

func (s *Service) ProcessSettlement(ctx context.Context, settlementID string) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return err
	}
	if settlement.Status != domain.SettlementPending {
		return domain.ErrSettlementFailed
	}
	if err := s.settlementRepo.UpdateStatus(ctx, settlementID, domain.SettlementProcessing); err != nil {
		return err
	}
	merchantWallet := fmt.Sprintf("merchant_%s_%s", settlement.MerchantID, settlement.Currency)
	platformFee := fmt.Sprintf("platform_fee_%s", settlement.Currency)
	netTxn, err := s.postTxn(ctx, domain.TxnSettlement, platformFee, merchantWallet, settlement.NetAmount, settlement.Currency,
		fmt.Sprintf("Settlement %s net amount", settlementID))
	if err != nil {
		s.settlementRepo.UpdateStatus(ctx, settlementID, domain.SettlementFailed)
		return err
	}
	if settlement.FeeAmount > 0 {
		feeAcct := fmt.Sprintf("platform_revenue_%s", settlement.Currency)
		_, err = s.postTxn(ctx, domain.TxnFee, merchantWallet, feeAcct, settlement.FeeAmount, settlement.Currency,
			fmt.Sprintf("Settlement %s fee", settlementID))
		if err != nil {
			return err
		}
	}
	if err := s.settlementRepo.UpdateStatus(ctx, settlementID, domain.SettlementCompleted); err != nil {
		return err
	}
	now := time.Now()
	settlement.CompletedAt = &now
	_ = netTxn
	metrics.SettlementsTotal.Inc()
	s.publisher.Publish(ctx, events.EventSettlementCompleted, settlement)
	return nil
}
