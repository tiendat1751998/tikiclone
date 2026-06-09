package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/domain"
)

type Repository interface {
	CreateAccount(ctx context.Context, a *domain.LedgerAccount) error
	GetAccount(ctx context.Context, id string) (*domain.LedgerAccount, error)
	ListAccounts(ctx context.Context) ([]domain.LedgerAccount, error)

	CreateEntry(ctx context.Context, e *domain.JournalEntry) error
	GetEntry(ctx context.Context, id string) (*domain.JournalEntry, error)
	ListEntries(ctx context.Context, referenceType, referenceID string, limit, offset int) ([]domain.JournalEntry, error)
	UpdateEntryStatus(ctx context.Context, id string, status domain.EntryStatus) error

	UpdateAccountBalance(ctx context.Context, id string, balance int64) error

	CreateWallet(ctx context.Context, w *domain.WalletAccount) error
	GetWallet(ctx context.Context, id string) (*domain.WalletAccount, error)
	GetWalletByUserType(ctx context.Context, userID string, walletType domain.WalletType) (*domain.WalletAccount, error)
	ListWallets(ctx context.Context, userID string) ([]domain.WalletAccount, error)
	UpdateWalletBalance(ctx context.Context, id string, balance, frozenBalance int64) error

	CreateWalletTransaction(ctx context.Context, t *domain.WalletTransaction) error
	ListWalletTransactions(ctx context.Context, walletID string, limit, offset int) ([]domain.WalletTransaction, error)

	CreateBatch(ctx context.Context, b *domain.SettlementBatch) error
	GetBatch(ctx context.Context, id string) (*domain.SettlementBatch, error)
	ListBatches(ctx context.Context, sellerID string, limit, offset int) ([]domain.SettlementBatch, error)
	UpdateBatchStatus(ctx context.Context, id string, status domain.SettlementStatus, paymentMethod, paymentReference *string) error

	CreateReconciliation(ctx context.Context, rec *domain.ReconciliationRecord) error
	ListReconciliations(ctx context.Context, accountID string, limit, offset int) ([]domain.ReconciliationRecord, error)
}

type LedgerService struct {
	repo Repository
}

func NewLedgerService(repo Repository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*domain.LedgerAccount, error) {
	now := time.Now().UTC()
	a := &domain.LedgerAccount{
		ID:          uuid.New().String(),
		AccountCode: req.AccountCode,
		AccountName: req.AccountName,
		AccountType: req.AccountType,
		Currency:    "SGD",
		Balance:     0,
		IsActive:    true,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.Currency != "" {
		a.Currency = req.Currency
	}
	if req.ParentAccountID != "" {
		a.ParentAccountID = &req.ParentAccountID
	}
	if err := s.repo.CreateAccount(ctx, a); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return a, nil
}

func (s *LedgerService) GetAccount(ctx context.Context, id string) (*domain.LedgerAccount, error) {
	return s.repo.GetAccount(ctx, id)
}

func (s *LedgerService) ListAccounts(ctx context.Context) ([]domain.LedgerAccount, error) {
	return s.repo.ListAccounts(ctx)
}

func (s *LedgerService) PostEntry(ctx context.Context, req *PostEntryRequest) (*domain.JournalEntry, error) {
	if req.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if req.DebitAccountID == req.CreditAccountID {
		return nil, domain.ErrDebitCreditMismatch
	}

	now := time.Now().UTC()
	e := &domain.JournalEntry{
		ID:              uuid.New().String(),
		EntryNumber:     uuid.New().String()[:8],
		TransactionID:   req.TransactionID,
		ReferenceType:   req.ReferenceType,
		ReferenceID:     req.ReferenceID,
		DebitAccountID:  req.DebitAccountID,
		CreditAccountID: req.CreditAccountID,
		Amount:          req.Amount,
		Currency:        "SGD",
		Description:     req.Description,
		Status:          domain.EntryPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateEntry(ctx, e); err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}
	return e, nil
}

func (s *LedgerService) PostAndSettleEntry(ctx context.Context, req *PostEntryRequest) (*domain.JournalEntry, error) {
	e, err := s.PostEntry(ctx, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	e.Status = domain.EntryPosted
	e.PostedAt = &now
	if err := s.repo.UpdateEntryStatus(ctx, e.ID, domain.EntryPosted); err != nil {
		return nil, fmt.Errorf("post entry: %w", err)
	}

	debit, err := s.repo.GetAccount(ctx, req.DebitAccountID)
	if err != nil {
		return nil, err
	}
	credit, err := s.repo.GetAccount(ctx, req.CreditAccountID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAccountBalance(ctx, debit.ID, debit.Balance-req.Amount); err != nil {
		return nil, fmt.Errorf("update debit balance: %w", err)
	}
	if err := s.repo.UpdateAccountBalance(ctx, credit.ID, credit.Balance+req.Amount); err != nil {
		return nil, fmt.Errorf("update credit balance: %w", err)
	}

	return e, nil
}

func (s *LedgerService) GetEntry(ctx context.Context, id string) (*domain.JournalEntry, error) {
	return s.repo.GetEntry(ctx, id)
}

func (s *LedgerService) ListEntries(ctx context.Context, referenceType, referenceID string, limit, offset int) ([]domain.JournalEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListEntries(ctx, referenceType, referenceID, limit, offset)
}

func (s *LedgerService) CreateWallet(ctx context.Context, req *CreateWalletRequest) (*domain.WalletAccount, error) {
	now := time.Now().UTC()
	w := &domain.WalletAccount{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		WalletType:    req.WalletType,
		Currency:      "SGD",
		Balance:       0,
		FrozenBalance: 0,
		Status:        domain.WalletActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateWallet(ctx, w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (s *LedgerService) GetWallet(ctx context.Context, id string) (*domain.WalletAccount, error) {
	return s.repo.GetWallet(ctx, id)
}

func (s *LedgerService) GetWalletByUserType(ctx context.Context, userID string, walletType domain.WalletType) (*domain.WalletAccount, error) {
	return s.repo.GetWalletByUserType(ctx, userID, walletType)
}

func (s *LedgerService) ListWallets(ctx context.Context, userID string) ([]domain.WalletAccount, error) {
	return s.repo.ListWallets(ctx, userID)
}

func (s *LedgerService) ListWalletTransactions(ctx context.Context, walletID string, limit, offset int) ([]domain.WalletTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListWalletTransactions(ctx, walletID, limit, offset)
}

func (s *LedgerService) CreateBatch(ctx context.Context, req *CreateBatchRequest) (*domain.SettlementBatch, error) {
	now := time.Now().UTC()
	b := &domain.SettlementBatch{
		ID:               uuid.New().String(),
		BatchNumber:      uuid.New().String()[:8],
		SellerID:         req.SellerID,
		PeriodStart:      req.PeriodStart,
		PeriodEnd:        req.PeriodEnd,
		TotalSales:       req.TotalSales,
		TotalFees:        req.TotalFees,
		TotalCommissions: req.TotalCommissions,
		TotalAdjustments: req.TotalAdjustments,
		NetSettlement:    req.TotalSales - req.TotalFees - req.TotalCommissions + req.TotalAdjustments,
		Currency:         "SGD",
		Status:           domain.SettlementPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.repo.CreateBatch(ctx, b); err != nil {
		return nil, fmt.Errorf("create batch: %w", err)
	}
	return b, nil
}

func (s *LedgerService) GetBatch(ctx context.Context, id string) (*domain.SettlementBatch, error) {
	return s.repo.GetBatch(ctx, id)
}

func (s *LedgerService) ListBatches(ctx context.Context, sellerID string, limit, offset int) ([]domain.SettlementBatch, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListBatches(ctx, sellerID, limit, offset)
}

func (s *LedgerService) CreateReconciliation(ctx context.Context, req *CreateReconciliationRequest) (*domain.ReconciliationRecord, error) {
	now := time.Now().UTC()
	rec := &domain.ReconciliationRecord{
		ID:                 uuid.New().String(),
		ReconciliationDate: req.ReconciliationDate,
		AccountID:          req.AccountID,
		LedgerBalance:      req.LedgerBalance,
		ExternalBalance:    req.ExternalBalance,
		Difference:         req.ExternalBalance - req.LedgerBalance,
		Status:             domain.ReconcileUnmatched,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if rec.Difference == 0 {
		rec.Status = domain.ReconcileMatched
	}
	if err := s.repo.CreateReconciliation(ctx, rec); err != nil {
		return nil, fmt.Errorf("create reconciliation: %w", err)
	}
	return rec, nil
}

func (s *LedgerService) ListReconciliations(ctx context.Context, accountID string, limit, offset int) ([]domain.ReconciliationRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListReconciliations(ctx, accountID, limit, offset)
}

type CreateAccountRequest struct {
	AccountCode     string            `json:"account_code" binding:"required"`
	AccountName     string            `json:"account_name" binding:"required"`
	AccountType     domain.AccountType `json:"account_type" binding:"required"`
	ParentAccountID string            `json:"parent_account_id"`
	Currency        string            `json:"currency"`
	Description     *string           `json:"description"`
}

type PostEntryRequest struct {
	TransactionID   string              `json:"transaction_id" binding:"required"`
	ReferenceType   domain.ReferenceType `json:"reference_type" binding:"required"`
	ReferenceID     string              `json:"reference_id" binding:"required"`
	DebitAccountID  string              `json:"debit_account_id" binding:"required"`
	CreditAccountID string              `json:"credit_account_id" binding:"required"`
	Amount          int64               `json:"amount" binding:"required"`
	Description     *string             `json:"description"`
}

type CreateWalletRequest struct {
	UserID     string           `json:"user_id" binding:"required"`
	WalletType domain.WalletType `json:"wallet_type" binding:"required"`
}

type CreateBatchRequest struct {
	SellerID         string    `json:"seller_id" binding:"required"`
	PeriodStart      time.Time `json:"period_start" binding:"required"`
	PeriodEnd        time.Time `json:"period_end" binding:"required"`
	TotalSales       int64     `json:"total_sales"`
	TotalFees        int64     `json:"total_fees"`
	TotalCommissions int64     `json:"total_commissions"`
	TotalAdjustments int64     `json:"total_adjustments"`
}

type CreateReconciliationRequest struct {
	ReconciliationDate time.Time `json:"reconciliation_date" binding:"required"`
	AccountID          string    `json:"account_id" binding:"required"`
	LedgerBalance      int64     `json:"ledger_balance"`
	ExternalBalance    int64     `json:"external_balance"`
}
