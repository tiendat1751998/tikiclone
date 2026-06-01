package ledger

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/events"
	"github.com/tikiclone/tiki/platforms/billing/internal/metrics"
	"go.uber.org/zap"
)

type Engine struct {
	accountRepo  domain.AccountRepository
	txnRepo      domain.TransactionRepository
	ledgerRepo   domain.LedgerEntryRepository
	walletRepo   domain.WalletRepository
	publisher    events.Publisher
}

func NewEngine(
	ar domain.AccountRepository,
	tr domain.TransactionRepository,
	lr domain.LedgerEntryRepository,
	wr domain.WalletRepository,
	pub events.Publisher,
) *Engine {
	return &Engine{
		accountRepo: ar, txnRepo: tr, ledgerRepo: lr,
		walletRepo: wr, publisher: pub,
	}
}

func (e *Engine) PostTransaction(ctx context.Context, txnType domain.TransactionType, debitAccountID, creditAccountID string, amount int64, currency, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	debit, err := e.accountRepo.GetByID(ctx, debitAccountID)
	if err != nil {
		return nil, fmt.Errorf("debit account: %w", err)
	}
	credit, err := e.accountRepo.GetByID(ctx, creditAccountID)
	if err != nil {
		return nil, fmt.Errorf("credit account: %w", err)
	}
	if debit.Currency != currency || credit.Currency != currency {
		return nil, domain.ErrCurrencyMismatch
	}
	if debit.Status != "active" || credit.Status != "active" {
		return nil, domain.ErrAccountFrozen
	}
	available := debit.Balance - debit.Frozen
	if available < amount {
		return nil, domain.ErrInsufficientBalance
	}

	txn := &domain.Transaction{
		ID:             uuid.New().String(),
		IDempotencyKey: uuid.New().String(),
		Type:           txnType,
		Status:         domain.TxnPending,
		Amount:         amount,
		Currency:       currency,
		Description:    description,
		DebitAccountID:  debitAccountID,
		CreditAccountID: creditAccountID,
		CreatedAt:      time.Now(),
	}
	if err := e.txnRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("create txn: %w", err)
	}

	debitEntry := &domain.LedgerEntry{
		ID:            uuid.New().String(),
		TransactionID: txn.ID,
		AccountID:     debitAccountID,
		Type:          domain.EntryDebit,
		Amount:        amount,
		Currency:      currency,
		BalanceBefore: debit.Balance,
		BalanceAfter:  debit.Balance - amount,
		Description:   description,
		Reference:     txn.ID,
		CreatedAt:     time.Now(),
	}
	creditEntry := &domain.LedgerEntry{
		ID:            uuid.New().String(),
		TransactionID: txn.ID,
		AccountID:     creditAccountID,
		Type:          domain.EntryCredit,
		Amount:        amount,
		Currency:      currency,
		BalanceBefore: credit.Balance,
		BalanceAfter:  credit.Balance + amount,
		Description:   description,
		Reference:     txn.ID,
		CreatedAt:     time.Now(),
	}
	if err := e.ledgerRepo.CreateBatch(ctx, []*domain.LedgerEntry{debitEntry, creditEntry}); err != nil {
		return nil, fmt.Errorf("create ledger: %w", err)
	}
	if err := e.accountRepo.UpdateBalance(ctx, debitAccountID, debitEntry.BalanceAfter, debit.Frozen); err != nil {
		return nil, fmt.Errorf("update debit: %w", err)
	}
	if err := e.accountRepo.UpdateBalance(ctx, creditAccountID, creditEntry.BalanceAfter, credit.Frozen); err != nil {
		return nil, fmt.Errorf("update credit: %w", err)
	}

	debitTotal := debitEntry.Amount
	creditTotal := creditEntry.Amount
	_ = debitTotal
	_ = creditTotal

	txn.Status = domain.TxnCompleted
	now := time.Now()
	txn.CompletedAt = &now
	if err := e.txnRepo.UpdateStatus(ctx, txn.ID, domain.TxnCompleted); err != nil {
		observability.LogWithTrace(ctx).Error("update txn status", zap.Error(err))
	}

	metrics.TransactionsTotal.WithLabelValues(string(txnType)).Inc()
	e.publishEvent(ctx, events.EventTransactionCompleted, txn)
	return txn, nil
}

func (e *Engine) ReverseTransaction(ctx context.Context, originalTxnID string) (*domain.Transaction, error) {
	orig, err := e.txnRepo.GetByID(ctx, originalTxnID)
	if err != nil {
		return nil, fmt.Errorf("original txn: %w", err)
	}
	if orig.Status != domain.TxnCompleted {
		return nil, fmt.Errorf("cannot reverse non-completed txn")
	}
	return e.PostTransaction(ctx, domain.TxnAdjustment,
		orig.CreditAccountID, orig.DebitAccountID,
		orig.Amount, orig.Currency,
		fmt.Sprintf("Reversal of %s", originalTxnID))
}

func (e *Engine) GetBalance(ctx context.Context, accountID string) (int64, int64, error) {
	acct, err := e.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, 0, err
	}
	return acct.Balance, acct.Frozen, nil
}

func (e *Engine) GetTransaction(ctx context.Context, txnID string) (*domain.Transaction, error) {
	return e.txnRepo.GetByID(ctx, txnID)
}

func (e *Engine) GetLedgerEntries(ctx context.Context, accountID string, offset, limit int) ([]*domain.LedgerEntry, int64, error) {
	return e.ledgerRepo.GetByAccount(ctx, accountID, offset, limit)
}

func (e *Engine) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if e.publisher != nil {
		e.publisher.Publish(ctx, eventType, payload)
	}
}
