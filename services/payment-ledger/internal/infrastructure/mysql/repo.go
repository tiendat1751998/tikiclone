package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/config"
	"github.com/tikiclone/tiki/services/payment-ledger/internal/domain"
)

type Repository struct {
	db *sqlx.DB
}

func NewDB(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var accountCols = "id, account_code, account_name, account_type, parent_account_id, currency, balance, is_active, description, created_at, updated_at"

func (r *Repository) CreateAccount(ctx context.Context, a *domain.LedgerAccount) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO ledger_accounts
		(id, account_code, account_name, account_type, parent_account_id, currency, balance, is_active, description, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.AccountCode, a.AccountName, a.AccountType, a.ParentAccountID, a.Currency, a.Balance, a.IsActive, a.Description, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *Repository) GetAccount(ctx context.Context, id string) (*domain.LedgerAccount, error) {
	a := &domain.LedgerAccount{}
	err := r.db.GetContext(ctx, a, "SELECT "+accountCols+" FROM ledger_accounts WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return a, nil
}

func (r *Repository) ListAccounts(ctx context.Context) ([]domain.LedgerAccount, error) {
	var items []domain.LedgerAccount
	err := r.db.SelectContext(ctx, &items, "SELECT "+accountCols+" FROM ledger_accounts ORDER BY account_code")
	return items, err
}

var entryCols = "id, entry_number, transaction_id, reference_type, reference_id, debit_account_id, credit_account_id, amount, currency, description, status, posted_at, reversed_at, reversal_entry_id, created_at, updated_at"

func (r *Repository) CreateEntry(ctx context.Context, e *domain.JournalEntry) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO journal_entries
		(id, entry_number, transaction_id, reference_type, reference_id, debit_account_id, credit_account_id, amount, currency, description, status, posted_at, reversed_at, reversal_entry_id, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.EntryNumber, e.TransactionID, e.ReferenceType, e.ReferenceID, e.DebitAccountID, e.CreditAccountID, e.Amount, e.Currency, e.Description, e.Status, e.PostedAt, e.ReversedAt, e.ReversalEntryID, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *Repository) GetEntry(ctx context.Context, id string) (*domain.JournalEntry, error) {
	e := &domain.JournalEntry{}
	err := r.db.GetContext(ctx, e, "SELECT "+entryCols+" FROM journal_entries WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return e, nil
}

func (r *Repository) ListEntries(ctx context.Context, referenceType, referenceID string, limit, offset int) ([]domain.JournalEntry, error) {
	var entries []domain.JournalEntry
	q := "SELECT " + entryCols + " FROM journal_entries WHERE 1=1"
	args := []interface{}{}
	if referenceType != "" {
		q += " AND reference_type = ?"
		args = append(args, referenceType)
	}
	if referenceID != "" {
		q += " AND reference_id = ?"
		args = append(args, referenceID)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &entries, q, args...)
	return entries, err
}

func (r *Repository) UpdateEntryStatus(ctx context.Context, id string, status domain.EntryStatus) error {
	_, err := r.db.ExecContext(ctx, "UPDATE journal_entries SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *Repository) UpdateAccountBalance(ctx context.Context, id string, balance int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE ledger_accounts SET balance = ? WHERE id = ?", balance, id)
	return err
}

var walletCols = "id, user_id, wallet_type, currency, balance, frozen_balance, status, last_transaction_at, created_at, updated_at"

func (r *Repository) CreateWallet(ctx context.Context, w *domain.WalletAccount) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO wallet_accounts
		(id, user_id, wallet_type, currency, balance, frozen_balance, status, last_transaction_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		w.ID, w.UserID, w.WalletType, w.Currency, w.Balance, w.FrozenBalance, w.Status, w.LastTransactionAt, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *Repository) GetWallet(ctx context.Context, id string) (*domain.WalletAccount, error) {
	w := &domain.WalletAccount{}
	err := r.db.GetContext(ctx, w, "SELECT "+walletCols+" FROM wallet_accounts WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return w, nil
}

func (r *Repository) GetWalletByUserType(ctx context.Context, userID string, walletType domain.WalletType) (*domain.WalletAccount, error) {
	w := &domain.WalletAccount{}
	err := r.db.GetContext(ctx, w, "SELECT "+walletCols+" FROM wallet_accounts WHERE user_id = ? AND wallet_type = ?", userID, walletType)
	if err != nil {
		return nil, mapError(err)
	}
	return w, nil
}

func (r *Repository) ListWallets(ctx context.Context, userID string) ([]domain.WalletAccount, error) {
	var items []domain.WalletAccount
	err := r.db.SelectContext(ctx, &items, "SELECT "+walletCols+" FROM wallet_accounts WHERE user_id = ?", userID)
	return items, err
}

func (r *Repository) UpdateWalletBalance(ctx context.Context, id string, balance, frozenBalance int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE wallet_accounts SET balance = ?, frozen_balance = ? WHERE id = ?", balance, frozenBalance, id)
	return err
}

var walletTxCols = "id, wallet_id, transaction_type, amount, balance_before, balance_after, reference_type, reference_id, journal_entry_id, description, status, created_at, updated_at"

func (r *Repository) CreateWalletTransaction(ctx context.Context, t *domain.WalletTransaction) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO wallet_transactions
		(id, wallet_id, transaction_type, amount, balance_before, balance_after, reference_type, reference_id, journal_entry_id, description, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.WalletID, t.TransactionType, t.Amount, t.BalanceBefore, t.BalanceAfter, t.ReferenceType, t.ReferenceID, t.JournalEntryID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Repository) ListWalletTransactions(ctx context.Context, walletID string, limit, offset int) ([]domain.WalletTransaction, error) {
	var items []domain.WalletTransaction
	err := r.db.SelectContext(ctx, &items, "SELECT "+walletTxCols+" FROM wallet_transactions WHERE wallet_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", walletID, limit, offset)
	return items, err
}

var batchCols = "id, batch_number, seller_id, period_start, period_end, total_sales, total_fees, total_commissions, total_adjustments, net_settlement, currency, status, payment_method, payment_reference, paid_at, created_at, updated_at"

func (r *Repository) CreateBatch(ctx context.Context, b *domain.SettlementBatch) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO settlement_batches
		(id, batch_number, seller_id, period_start, period_end, total_sales, total_fees, total_commissions, total_adjustments, net_settlement, currency, status, payment_method, payment_reference, paid_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		b.ID, b.BatchNumber, b.SellerID, b.PeriodStart, b.PeriodEnd, b.TotalSales, b.TotalFees, b.TotalCommissions, b.TotalAdjustments, b.NetSettlement, b.Currency, b.Status, b.PaymentMethod, b.PaymentReference, b.PaidAt, b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *Repository) GetBatch(ctx context.Context, id string) (*domain.SettlementBatch, error) {
	b := &domain.SettlementBatch{}
	err := r.db.GetContext(ctx, b, "SELECT "+batchCols+" FROM settlement_batches WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return b, nil
}

func (r *Repository) ListBatches(ctx context.Context, sellerID string, limit, offset int) ([]domain.SettlementBatch, error) {
	var items []domain.SettlementBatch
	q := "SELECT " + batchCols + " FROM settlement_batches"
	args := []interface{}{}
	if sellerID != "" {
		q += " WHERE seller_id = ?"
		args = append(args, sellerID)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *Repository) UpdateBatchStatus(ctx context.Context, id string, status domain.SettlementStatus, paymentMethod, paymentReference *string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE settlement_batches SET status = ?, payment_method = ?, payment_reference = ? WHERE id = ?", status, paymentMethod, paymentReference, id)
	return err
}

var reconcileCols = "id, reconciliation_date, account_id, ledger_balance, external_balance, difference, status, resolved_by, resolved_at, notes, created_at, updated_at"

func (r *Repository) CreateReconciliation(ctx context.Context, rec *domain.ReconciliationRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO reconciliation_records
		(id, reconciliation_date, account_id, ledger_balance, external_balance, difference, status, resolved_by, resolved_at, notes, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		rec.ID, rec.ReconciliationDate, rec.AccountID, rec.LedgerBalance, rec.ExternalBalance, rec.Difference, rec.Status, rec.ResolvedBy, rec.ResolvedAt, rec.Notes, rec.CreatedAt, rec.UpdatedAt)
	return err
}

func (r *Repository) ListReconciliations(ctx context.Context, accountID string, limit, offset int) ([]domain.ReconciliationRecord, error) {
	var items []domain.ReconciliationRecord
	q := "SELECT " + reconcileCols + " FROM reconciliation_records"
	args := []interface{}{}
	if accountID != "" {
		q += " WHERE account_id = ?"
		args = append(args, accountID)
	}
	q += " ORDER BY reconciliation_date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
