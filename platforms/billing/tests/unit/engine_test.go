package unit

import (
	"context"
	"sync"
	"testing"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
	"github.com/tikiclone/tiki/platforms/billing/internal/ledger"
)

type mockPublisher struct{}
func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload interface{}) error { return nil }

type mockAccountRepo struct {
	accounts map[string]*domain.Account
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{accounts: make(map[string]*domain.Account)}
}

func (r *mockAccountRepo) Create(ctx context.Context, a *domain.Account) error {
	r.accounts[a.ID] = a; return nil
}
func (r *mockAccountRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	a, ok := r.accounts[id]; if !ok { return nil, domain.ErrAccountNotFound }; return a, nil
}
func (r *mockAccountRepo) GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Account, error) {
	for _, a := range r.accounts { if a.UserID == userID && a.Currency == currency { return a, nil } }; return nil, domain.ErrAccountNotFound
}
func (r *mockAccountRepo) UpdateBalance(ctx context.Context, id string, balance, frozen int64) error {
	a, ok := r.accounts[id]; if !ok { return domain.ErrAccountNotFound }; a.Balance = balance; a.Frozen = frozen; return nil
}

type mockTxnRepo struct {
	txns map[string]*domain.Transaction
}

func newMockTxnRepo() *mockTxnRepo { return &mockTxnRepo{txns: make(map[string]*domain.Transaction)} }
func (r *mockTxnRepo) Create(ctx context.Context, t *domain.Transaction) error { r.txns[t.ID] = t; return nil }
func (r *mockTxnRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	t, ok := r.txns[id]; if !ok { return nil, nil }; return t, nil
}
func (r *mockTxnRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	for _, t := range r.txns { if t.IDempotencyKey == key { return t, nil } }; return nil, nil
}
func (r *mockTxnRepo) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	t, ok := r.txns[id]; if !ok { return nil }; t.Status = status; return nil
}
func (r *mockTxnRepo) ListByAccount(ctx context.Context, accountID string, offset, limit int) ([]*domain.Transaction, int64, error) {
	return nil, 0, nil
}

type mockLedgerRepo struct {
	mu      sync.Mutex
	entries []*domain.LedgerEntry
}
func newMockLedgerRepo() *mockLedgerRepo { return &mockLedgerRepo{} }
func (r *mockLedgerRepo) CreateBatch(ctx context.Context, entries []*domain.LedgerEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entries...)
	return nil
}
func (r *mockLedgerRepo) GetByTransaction(ctx context.Context, txnID string) ([]*domain.LedgerEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*domain.LedgerEntry
	for _, e := range r.entries {
		if e.TransactionID == txnID {
			result = append(result, e)
		}
	}
	return result, nil
}
func (r *mockLedgerRepo) GetByAccount(ctx context.Context, accountID string, offset, limit int) ([]*domain.LedgerEntry, int64, error) { return nil, 0, nil }

type mockWalletRepo struct{}
func (r *mockWalletRepo) Create(ctx context.Context, w *domain.Wallet) error { return nil }
func (r *mockWalletRepo) GetByID(ctx context.Context, id string) (*domain.Wallet, error) { return &domain.Wallet{Balance: 1000000, Frozen: 0}, nil }
func (r *mockWalletRepo) GetByUser(ctx context.Context, userID string) ([]*domain.Wallet, error) { return nil, nil }
func (r *mockWalletRepo) GetByUserAndType(ctx context.Context, userID, walletType string) (*domain.Wallet, error) { return &domain.Wallet{ID: "wallet1", Balance: 1000000}, nil }
func (r *mockWalletRepo) UpdateBalance(ctx context.Context, id string, balance, frozen, pending int64) error { return nil }

func TestPostTransaction(t *testing.T) {
	acctRepo := newMockAccountRepo()
	debit := domain.NewAccount("user1", domain.AccountTypeAsset, "VND")
	debit.ID = "debit1"; debit.Balance = 1000
	credit := domain.NewAccount("user2", domain.AccountTypeAsset, "VND")
	credit.ID = "credit1"; credit.Balance = 500

	acctRepo.Create(context.Background(), debit)
	acctRepo.Create(context.Background(), credit)

	txnRepo := newMockTxnRepo()
	ledgerRepo := newMockLedgerRepo()
	walletRepo := &mockWalletRepo{}

	engine := ledger.NewEngine(acctRepo, txnRepo, ledgerRepo, walletRepo, &mockPublisher{})

	txn, err := engine.PostTransaction(context.Background(), domain.TxnTransfer, "debit1", "credit1", 200, "VND", "test transfer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txn.Status != domain.TxnCompleted {
		t.Errorf("expected completed, got %s", txn.Status)
	}
	if txn.Amount != 200 {
		t.Errorf("expected amount 200, got %d", txn.Amount)
	}
}

func TestPostTransactionInsufficientBalance(t *testing.T) {
	acctRepo := newMockAccountRepo()
	debit := domain.NewAccount("user1", domain.AccountTypeAsset, "VND")
	debit.ID = "debit1"; debit.Balance = 50
	credit := domain.NewAccount("user2", domain.AccountTypeAsset, "VND")
	credit.ID = "credit1"; credit.Balance = 500
	acctRepo.Create(context.Background(), debit)
	acctRepo.Create(context.Background(), credit)

	engine := ledger.NewEngine(acctRepo, newMockTxnRepo(), newMockLedgerRepo(), &mockWalletRepo{}, &mockPublisher{})
	_, err := engine.PostTransaction(context.Background(), domain.TxnTransfer, "debit1", "credit1", 200, "VND", "test")
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestPostTransactionInvalidAmount(t *testing.T) {
	engine := ledger.NewEngine(newMockAccountRepo(), newMockTxnRepo(), newMockLedgerRepo(), &mockWalletRepo{}, &mockPublisher{})
	_, err := engine.PostTransaction(context.Background(), domain.TxnTransfer, "a", "b", 0, "VND", "test")
	if err == nil {
		t.Fatal("expected invalid amount error")
	}
}

func TestLedgerEntryBalancing(t *testing.T) {
	acctRepo := newMockAccountRepo()
	debit := domain.NewAccount("user1", domain.AccountTypeAsset, "VND")
	debit.ID = "debit1"; debit.Balance = 1000
	credit := domain.NewAccount("user2", domain.AccountTypeAsset, "VND")
	credit.ID = "credit1"; credit.Balance = 500
	acctRepo.Create(context.Background(), debit)
	acctRepo.Create(context.Background(), credit)

	ledgerRepo := newMockLedgerRepo()
	engine := ledger.NewEngine(acctRepo, newMockTxnRepo(), ledgerRepo, &mockWalletRepo{}, &mockPublisher{})
	txn, err := engine.PostTransaction(context.Background(), domain.TxnTransfer, "debit1", "credit1", 300, "VND", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, _ := ledgerRepo.GetByTransaction(context.Background(), txn.ID)
	if len(entries) != 2 {
		t.Errorf("expected 2 ledger entries, got %d", len(entries))
	}
	totalDebit := int64(0)
	totalCredit := int64(0)
	for _, e := range entries {
		if e.Type == domain.EntryDebit {
			totalDebit += e.Amount
		} else {
			totalCredit += e.Amount
		}
	}
	if totalDebit != totalCredit {
		t.Errorf("ledger imbalanced: debit=%d credit=%d", totalDebit, totalCredit)
	}
}

func TestReverseTransaction(t *testing.T) {
	acctRepo := newMockAccountRepo()
	user1 := domain.NewAccount("user1", domain.AccountTypeAsset, "VND")
	user1.ID = "u1"; user1.Balance = 1000
	user2 := domain.NewAccount("user2", domain.AccountTypeAsset, "VND")
	user2.ID = "u2"; user2.Balance = 0
	acctRepo.Create(context.Background(), user1)
	acctRepo.Create(context.Background(), user2)

	txnRepo := newMockTxnRepo()
	ledgerRepo := newMockLedgerRepo()
	engine := ledger.NewEngine(acctRepo, txnRepo, ledgerRepo, &mockWalletRepo{}, &mockPublisher{})

	txn, _ := engine.PostTransaction(context.Background(), domain.TxnTransfer, "u1", "u2", 200, "VND", "test")
	rev, err := engine.ReverseTransaction(context.Background(), txn.ID)
	if err != nil {
		t.Fatalf("reverse failed: %v", err)
	}
	if rev.Type != domain.TxnAdjustment {
		t.Errorf("expected adjustment, got %s", rev.Type)
	}
	if rev.Amount != 200 {
		t.Errorf("expected amount 200, got %d", rev.Amount)
	}
}
