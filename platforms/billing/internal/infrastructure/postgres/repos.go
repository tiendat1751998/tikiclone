package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
)

type AccountRepo struct{ pool *Pool }

func NewAccountRepo(pool *Pool) *AccountRepo { return &AccountRepo{pool: pool} }

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO accounts (id,user_id,type,currency,balance,frozen,status,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.UserID, a.Type, a.Currency, a.Balance, a.Frozen, a.Status, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AccountRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	a := &domain.Account{}
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,type,currency,balance,frozen,status,created_at,updated_at
		FROM accounts WHERE id=$1`, id).Scan(&a.ID, &a.UserID, &a.Type, &a.Currency, &a.Balance, &a.Frozen, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	return a, err
}

func (r *AccountRepo) GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Account, error) {
	a := &domain.Account{}
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,type,currency,balance,frozen,status,created_at,updated_at
		FROM accounts WHERE user_id=$1 AND currency=$2`, userID, currency).
		Scan(&a.ID, &a.UserID, &a.Type, &a.Currency, &a.Balance, &a.Frozen, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	return a, err
}

func (r *AccountRepo) UpdateBalance(ctx context.Context, id string, balance, frozen int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE accounts SET balance=$2, frozen=$3, updated_at=NOW() WHERE id=$1`, id, balance, frozen)
	return err
}

type TransactionRepo struct{ pool *Pool }

func NewTransactionRepo(pool *Pool) *TransactionRepo { return &TransactionRepo{pool: pool} }

func (r *TransactionRepo) Create(ctx context.Context, txn *domain.Transaction) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO transactions (id,idempotency_key,type,status,amount,currency,description,debit_account_id,credit_account_id,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		txn.ID, txn.IDempotencyKey, txn.Type, txn.Status, txn.Amount, txn.Currency, txn.Description, txn.DebitAccountID, txn.CreditAccountID, txn.CreatedAt)
	return err
}

func (r *TransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	t := &domain.Transaction{}
	err := r.pool.QueryRow(ctx, `SELECT id,idempotency_key,type,status,amount,currency,description,debit_account_id,credit_account_id,created_at,completed_at
		FROM transactions WHERE id=$1`, id).Scan(&t.ID, &t.IDempotencyKey, &t.Type, &t.Status, &t.Amount, &t.Currency, &t.Description, &t.DebitAccountID, &t.CreditAccountID, &t.CreatedAt, &t.CompletedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	return t, err
}

func (r *TransactionRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	t := &domain.Transaction{}
	err := r.pool.QueryRow(ctx, `SELECT id,idempotency_key,type,status,amount,currency,description,debit_account_id,credit_account_id,created_at,completed_at
		FROM transactions WHERE idempotency_key=$1`, key).Scan(&t.ID, &t.IDempotencyKey, &t.Type, &t.Status, &t.Amount, &t.Currency, &t.Description, &t.DebitAccountID, &t.CreditAccountID, &t.CreatedAt, &t.CompletedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *TransactionRepo) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE transactions SET status=$2, completed_at=CASE WHEN $2='completed' THEN NOW() ELSE completed_at END WHERE id=$1`, id, status)
	return err
}

func (r *TransactionRepo) ListByAccount(ctx context.Context, accountID string, offset, limit int) ([]*domain.Transaction, int64, error) {
	var total int64
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE debit_account_id=$1 OR credit_account_id=$1`, accountID).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,idempotency_key,type,status,amount,currency,description,debit_account_id,credit_account_id,created_at,completed_at
		FROM transactions WHERE debit_account_id=$1 OR credit_account_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := rows.Scan(&t.ID, &t.IDempotencyKey, &t.Type, &t.Status, &t.Amount, &t.Currency, &t.Description, &t.DebitAccountID, &t.CreditAccountID, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, t)
	}
	return result, total, nil
}

type LedgerEntryRepo struct{ pool *Pool }

func NewLedgerEntryRepo(pool *Pool) *LedgerEntryRepo { return &LedgerEntryRepo{pool: pool} }

func (r *LedgerEntryRepo) CreateBatch(ctx context.Context, entries []*domain.LedgerEntry) error {
	for _, e := range entries {
		_, err := r.pool.Exec(ctx, `INSERT INTO ledger_entries (id,transaction_id,account_id,type,amount,currency,balance_before,balance_after,description,reference,created_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			e.ID, e.TransactionID, e.AccountID, e.Type, e.Amount, e.Currency, e.BalanceBefore, e.BalanceAfter, e.Description, e.Reference, e.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *LedgerEntryRepo) GetByTransaction(ctx context.Context, txnID string) ([]*domain.LedgerEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,transaction_id,account_id,type,amount,currency,balance_before,balance_after,description,reference,created_at
		FROM ledger_entries WHERE transaction_id=$1 ORDER BY type`, txnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.LedgerEntry
	for rows.Next() {
		e := &domain.LedgerEntry{}
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.AccountID, &e.Type, &e.Amount, &e.Currency, &e.BalanceBefore, &e.BalanceAfter, &e.Description, &e.Reference, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *LedgerEntryRepo) GetByAccount(ctx context.Context, accountID string, offset, limit int) ([]*domain.LedgerEntry, int64, error) {
	var total int64
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE account_id=$1`, accountID).Scan(&total)
	rows, err := r.pool.Query(ctx, `SELECT id,transaction_id,account_id,type,amount,currency,balance_before,balance_after,description,reference,created_at
		FROM ledger_entries WHERE account_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*domain.LedgerEntry
	for rows.Next() {
		e := &domain.LedgerEntry{}
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.AccountID, &e.Type, &e.Amount, &e.Currency, &e.BalanceBefore, &e.BalanceAfter, &e.Description, &e.Reference, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, e)
	}
	return result, total, nil
}

type WalletRepo struct{ pool *Pool }

func NewWalletRepo(pool *Pool) *WalletRepo { return &WalletRepo{pool: pool} }

func (r *WalletRepo) Create(ctx context.Context, w *domain.Wallet) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO wallets (id,user_id,type,currency,balance,frozen,pending,status,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		w.ID, w.UserID, w.Type, w.Currency, w.Balance, w.Frozen, w.Pending, w.Status, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WalletRepo) GetByID(ctx context.Context, id string) (*domain.Wallet, error) {
	w := &domain.Wallet{}
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,type,currency,balance,frozen,pending,status,created_at,updated_at
		FROM wallets WHERE id=$1`, id).Scan(&w.ID, &w.UserID, &w.Type, &w.Currency, &w.Balance, &w.Frozen, &w.Pending, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (r *WalletRepo) GetByUser(ctx context.Context, userID string) ([]*domain.Wallet, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,type,currency,balance,frozen,pending,status,created_at,updated_at
		FROM wallets WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Wallet
	for rows.Next() {
		w := &domain.Wallet{}
		if err := rows.Scan(&w.ID, &w.UserID, &w.Type, &w.Currency, &w.Balance, &w.Frozen, &w.Pending, &w.Status, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, nil
}

func (r *WalletRepo) GetByUserAndType(ctx context.Context, userID, walletType string) (*domain.Wallet, error) {
	w := &domain.Wallet{}
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,type,currency,balance,frozen,pending,status,created_at,updated_at
		FROM wallets WHERE user_id=$1 AND type=$2`, userID, walletType).
		Scan(&w.ID, &w.UserID, &w.Type, &w.Currency, &w.Balance, &w.Frozen, &w.Pending, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (r *WalletRepo) UpdateBalance(ctx context.Context, id string, balance, frozen, pending int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE wallets SET balance=$2, frozen=$3, pending=$4, updated_at=NOW() WHERE id=$1`, id, balance, frozen, pending)
	return err
}

type SettlementRepo struct{ pool *Pool }

func NewSettlementRepo(pool *Pool) *SettlementRepo { return &SettlementRepo{pool: pool} }

func (r *SettlementRepo) Create(ctx context.Context, s *domain.Settlement) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO settlements (id,merchant_id,amount,currency,fee_amount,net_amount,status,period_start,period_end,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		s.ID, s.MerchantID, s.Amount, s.Currency, s.FeeAmount, s.NetAmount, s.Status, s.PeriodStart, s.PeriodEnd, s.CreatedAt)
	return err
}

func (r *SettlementRepo) GetByID(ctx context.Context, id string) (*domain.Settlement, error) {
	s := &domain.Settlement{}
	err := r.pool.QueryRow(ctx, `SELECT id,merchant_id,amount,currency,fee_amount,net_amount,status,period_start,period_end,scheduled_date,completed_at,created_at
		FROM settlements WHERE id=$1`, id).Scan(&s.ID, &s.MerchantID, &s.Amount, &s.Currency, &s.FeeAmount, &s.NetAmount, &s.Status, &s.PeriodStart, &s.PeriodEnd, &s.ScheduledDate, &s.CompletedAt, &s.CreatedAt)
	return s, err
}

func (r *SettlementRepo) UpdateStatus(ctx context.Context, id string, status domain.SettlementStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE settlements SET status=$2, completed_at=CASE WHEN $2='completed' THEN NOW() ELSE completed_at END WHERE id=$1`, id, status)
	return err
}

func (r *SettlementRepo) ListPending(ctx context.Context, limit int) ([]*domain.Settlement, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,merchant_id,amount,currency,fee_amount,net_amount,status,period_start,period_end,scheduled_date,completed_at,created_at
		FROM settlements WHERE status='pending' ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Settlement
	for rows.Next() {
		s := &domain.Settlement{}
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.Amount, &s.Currency, &s.FeeAmount, &s.NetAmount, &s.Status, &s.PeriodStart, &s.PeriodEnd, &s.ScheduledDate, &s.CompletedAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}
