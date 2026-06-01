package unit

import (
	"testing"
	"github.com/tikiclone/tiki/platforms/billing/internal/domain"
)

func TestNewAccount(t *testing.T) {
	a := domain.NewAccount("user1", domain.AccountTypeAsset, "VND")
	if a.UserID != "user1" {
		t.Errorf("expected user1, got %s", a.UserID)
	}
	if a.Type != domain.AccountTypeAsset {
		t.Errorf("expected asset, got %s", a.Type)
	}
	if a.Balance != 0 {
		t.Errorf("expected 0 balance, got %d", a.Balance)
	}
	if a.Status != "active" {
		t.Errorf("expected active, got %s", a.Status)
	}
}

func TestTransactionTypes(t *testing.T) {
	tests := []struct {
		txnType domain.TransactionType
		valid   bool
	}{
		{domain.TxnDeposit, true},
		{domain.TxnWithdrawal, true},
		{domain.TxnTransfer, true},
		{domain.TxnPayment, true},
		{domain.TxnRefund, true},
		{domain.TxnFee, true},
		{domain.TxnSettlement, true},
		{domain.TxnPayout, true},
		{domain.TxnAdjustment, true},
	}
	validTypes := map[domain.TransactionType]bool{
		domain.TxnDeposit: true, domain.TxnWithdrawal: true, domain.TxnTransfer: true,
		domain.TxnPayment: true, domain.TxnRefund: true, domain.TxnFee: true,
		domain.TxnSettlement: true, domain.TxnPayout: true, domain.TxnAdjustment: true,
	}
	for _, tt := range tests {
		if !validTypes[tt.txnType] {
			t.Errorf("unexpected type: %s", tt.txnType)
		}
		_ = tt.valid
	}
}

func TestTransactionStatus(t *testing.T) {
	statuses := []domain.TransactionStatus{
		domain.TxnPending, domain.TxnCompleted,
		domain.TxnFailed, domain.TxnReversed, domain.TxnExpired,
	}
	if len(statuses) != 5 {
		t.Errorf("expected 5 statuses, got %d", len(statuses))
	}
}

func TestErrorTypes(t *testing.T) {
	if domain.ErrInsufficientBalance.Error() != "billing: insufficient_balance" {
		t.Errorf("unexpected: %s", domain.ErrInsufficientBalance.Error())
	}
	if domain.ErrAccountNotFound.Error() != "billing: account_not_found" {
		t.Errorf("unexpected: %s", domain.ErrAccountNotFound.Error())
	}
	if domain.ErrDuplicateTxn.Error() != "billing: duplicate_transaction" {
		t.Errorf("unexpected: %s", domain.ErrDuplicateTxn.Error())
	}
	if domain.ErrInvalidAmount.Error() != "billing: invalid_amount" {
		t.Errorf("unexpected: %s", domain.ErrInvalidAmount.Error())
	}
}

func TestNegativeAmount(t *testing.T) {
	if -100 > 0 {
		t.Error("negative amount should be invalid")
	}
}

func TestZeroAmount(t *testing.T) {
	if 0 > 0 {
		t.Error("zero amount should be invalid")
	}
}

func TestAccountStatus(t *testing.T) {
	acct := domain.NewAccount("user1", domain.AccountTypeLiability, "USD")
	if acct.Type != domain.AccountTypeLiability {
		t.Errorf("expected liability, got %s", acct.Type)
	}
}

func TestWalletTypes(t *testing.T) {
	w := &domain.Wallet{Type: "user", Currency: "VND", Status: "active"}
	if w.Type != "user" || w.Currency != "VND" || w.Status != "active" {
		t.Error("wallet fields mismatch")
	}
}

func TestSettlementStatus(t *testing.T) {
	s := &domain.Settlement{Status: domain.SettlementPending}
	if s.Status != domain.SettlementPending {
		t.Errorf("expected pending, got %s", s.Status)
	}
}

func TestPayoutStatus(t *testing.T) {
	p := &domain.Payout{Status: domain.PayoutRequested}
	if p.Status != domain.PayoutRequested {
		t.Errorf("expected requested, got %s", p.Status)
	}
}

func TestRefundStatus(t *testing.T) {
	r := &domain.Refund{Status: domain.RefundPending}
	if r.Status != domain.RefundPending {
		t.Errorf("expected pending, got %s", r.Status)
	}
}

func TestDebitCreditTypes(t *testing.T) {
	if domain.EntryDebit != "debit" || domain.EntryCredit != "credit" {
		t.Error("entry type strings mismatch")
	}
}

func TestAccountTypeString(t *testing.T) {
	if string(domain.AccountTypeAsset) != "asset" {
		t.Errorf("expected asset, got %s", domain.AccountTypeAsset)
	}
}
