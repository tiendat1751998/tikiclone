package wallets

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/events"
)

type Service struct {
	walletRepo    domain.WalletRepository
	accountRepo   domain.AccountRepository
	ledger        *LedgerBridge
	publisher     events.Publisher
}

type LedgerBridge struct {
	PostTransaction func(ctx context.Context, txnType domain.TransactionType, debit, credit string, amount int64, currency, description string) (*domain.Transaction, error)
}

func NewService(
	wr domain.WalletRepository,
	ar domain.AccountRepository,
	lb *LedgerBridge,
	pub events.Publisher,
) *Service {
	return &Service{
		walletRepo: wr, accountRepo: ar,
		ledger: lb, publisher: pub,
	}
}

func (s *Service) CreateWallet(ctx context.Context, userID, walletType, currency string) (*domain.Wallet, error) {
	w := &domain.Wallet{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      walletType,
		Currency:  currency,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.walletRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	acct := domain.NewAccount(userID, domain.AccountTypeAsset, currency)
	acct.ID = w.ID
	if err := s.accountRepo.Create(ctx, acct); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return w, nil
}

func (s *Service) Deposit(ctx context.Context, userID, walletType string, amount int64, currency, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	wallet, err := s.walletRepo.GetByUserAndType(ctx, userID, walletType)
	if err != nil {
		return nil, err
	}
	platformAcct := fmt.Sprintf("platform_%s", currency)
	return s.ledger.PostTransaction(ctx, domain.TxnDeposit, platformAcct, wallet.ID, amount, currency, description)
}

func (s *Service) Withdraw(ctx context.Context, userID, walletType string, amount int64, currency, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	wallet, err := s.walletRepo.GetByUserAndType(ctx, userID, walletType)
	if err != nil {
		return nil, err
	}
	platformAcct := fmt.Sprintf("platform_%s", currency)
	return s.ledger.PostTransaction(ctx, domain.TxnWithdrawal, wallet.ID, platformAcct, amount, currency, description)
}

func (s *Service) Transfer(ctx context.Context, fromUserID, toUserID string, amount int64, currency, description string) (*domain.Transaction, error) {
	fromWallet, err := s.walletRepo.GetByUserAndType(ctx, fromUserID, "user")
	if err != nil {
		return nil, err
	}
	toWallet, err := s.walletRepo.GetByUserAndType(ctx, toUserID, "user")
	if err != nil {
		return nil, err
	}
	return s.ledger.PostTransaction(ctx, domain.TxnTransfer, fromWallet.ID, toWallet.ID, amount, currency, description)
}

func (s *Service) GetBalance(ctx context.Context, walletID string) (int64, int64, int64, error) {
	w, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return 0, 0, 0, err
	}
	return w.Balance, w.Frozen, w.Pending, nil
}
