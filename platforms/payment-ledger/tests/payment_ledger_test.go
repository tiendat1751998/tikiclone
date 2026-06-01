package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/dispute"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/ledger"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payment"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/payout"
	"github.com/tikiclone/tiki/platforms/payment-ledger/internal/reconciliation"
)

func ptr(i int64) *int64 { return &i }

func TestPaymentProcess(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, err := svc.Process(context.Background(), "order1", "user1", 1000, 50, "VND", payment.MethodCreditCard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != payment.StatusPending {
		t.Errorf("expected pending, got %s", p.Status)
	}
	if p.Amount != 1000 {
		t.Errorf("expected 1000, got %d", p.Amount)
	}
	if p.NetAmount != 950 {
		t.Errorf("expected net 950, got %d", p.NetAmount)
	}
	if p.TransactionID == "" {
		t.Error("expected transaction_id")
	}
}

func TestPaymentProcessInvalidAmount(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	_, err := svc.Process(context.Background(), "order1", "user1", 0, 0, "VND", payment.MethodCreditCard)
	if err != payment.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestPaymentProcessInvalidMethod(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	_, err := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", "bitcoin")
	if err != payment.ErrInvalidMethod {
		t.Errorf("expected ErrInvalidMethod, got %v", err)
	}
}

func TestPaymentNetAmountWithFee(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, err := svc.Process(context.Background(), "order1", "user1", 500, 100, "VND", payment.MethodCreditCard)
	if err != nil {
		t.Fatal(err)
	}
	if p.NetAmount != 400 {
		t.Errorf("expected net 400, got %d", p.NetAmount)
	}
}

func TestPaymentNetAmountFeeExceedsAmount(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, err := svc.Process(context.Background(), "order1", "user1", 100, 200, "VND", payment.MethodCreditCard)
	if err != nil {
		t.Fatal(err)
	}
	if p.NetAmount != 0 {
		t.Errorf("expected net 0, got %d", p.NetAmount)
	}
}

func TestPaymentThreePhaseLifecycle(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 1000, 50, "VND", payment.MethodWallet)

	p, err := svc.Authorize(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("authorize failed: %v", err)
	}
	if p.Status != payment.StatusProcessing {
		t.Errorf("expected processing after authorize, got %s", p.Status)
	}

	p, err = svc.Capture(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if p.Status != payment.StatusCompleted {
		t.Errorf("expected completed after capture, got %s", p.Status)
	}

	p, err = svc.Settle(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("settle failed: %v", err)
	}
	if p.SettledAt == "" {
		t.Error("expected settled_at timestamp")
	}
	if p.Status != payment.StatusCompleted {
		t.Errorf("expected completed after settle, got %s", p.Status)
	}
}

func TestPaymentAuthorizeInvalidStatus(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodBankTransfer)
	p, _ = svc.Authorize(context.Background(), p.ID)
	p, _ = svc.Capture(context.Background(), p.ID)
	_, err := svc.Authorize(context.Background(), p.ID)
	if err != payment.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestPaymentCaptureWithoutAuthorize(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodCreditCard)
	_, err := svc.Capture(context.Background(), p.ID)
	if err != payment.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestPaymentSettleWithoutCapture(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodCreditCard)
	_, err := svc.Settle(context.Background(), p.ID)
	if err != payment.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestPaymentRefund(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 1000, 0, "VND", payment.MethodCreditCard)
	p, _ = svc.Authorize(context.Background(), p.ID)
	p, _ = svc.Capture(context.Background(), p.ID)

	p, err := svc.Refund(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if p.Status != payment.StatusRefunded {
		t.Errorf("expected refunded, got %s", p.Status)
	}
}

func TestPaymentRefundNotCaptured(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodWallet)
	_, err := svc.Refund(context.Background(), p.ID)
	if err != payment.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestPaymentDoubleRefund(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodCOD)
	p, _ = svc.Authorize(context.Background(), p.ID)
	_, _ = svc.Capture(context.Background(), p.ID)
	_, _ = svc.Refund(context.Background(), p.ID)
	_, err := svc.Refund(context.Background(), p.ID)
	if err != payment.ErrAlreadyRefunded {
		t.Errorf("expected ErrAlreadyRefunded, got %v", err)
	}
}

func TestPaymentPartialRefund(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 1000, 0, "VND", payment.MethodCreditCard)
	p, _ = svc.Authorize(context.Background(), p.ID)
	p, _ = svc.Capture(context.Background(), p.ID)

	p, err := svc.PartialRefund(context.Background(), p.ID, 300)
	if err != nil {
		t.Fatalf("partial refund failed: %v", err)
	}
	if p.Status != payment.StatusPartiallyRefunded {
		t.Errorf("expected partially_refunded, got %s", p.Status)
	}
}

func TestPaymentPartialRefundExceedsAmount(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodWallet)
	p, _ = svc.Authorize(context.Background(), p.ID)
	p, _ = svc.Capture(context.Background(), p.ID)
	_, err := svc.PartialRefund(context.Background(), p.ID, 500)
	if err != payment.ErrRefundExceeds {
		t.Errorf("expected ErrRefundExceeds, got %v", err)
	}
}

func TestPaymentGetByID(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	p, _ := svc.Process(context.Background(), "order1", "user1", 100, 0, "VND", payment.MethodWallet)
	got, err := svc.GetByID(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %s, got %s", p.ID, got.ID)
	}
}

func TestPaymentGetByIDNotFound(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err != payment.ErrPaymentNotFound {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestPaymentList(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	svc.Process(context.Background(), "o1", "u1", 100, 0, "VND", payment.MethodWallet)
	svc.Process(context.Background(), "o2", "u2", 200, 0, "VND", payment.MethodWallet)
	items, total, err := svc.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestPaymentGetByOrder(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	svc.Process(context.Background(), "orderX", "user1", 100, 0, "VND", payment.MethodWallet)
	svc.Process(context.Background(), "orderX", "user1", 200, 0, "VND", payment.MethodWallet)
	items, err := svc.GetByOrder(context.Background(), "orderX")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 payments for orderX, got %d", len(items))
	}
}

func TestLedgerCreateAccount(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	a, err := svc.CreateAccount(context.Background(), "acct1", "Cash", ledger.AccountTypeAsset, "1000", "VND", 5000)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}
	if a.Balance != 5000 {
		t.Errorf("expected balance 5000, got %d", a.Balance)
	}
}

func TestLedgerPostTransaction(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "debit1", "Checking", ledger.AccountTypeAsset, "1001", "VND", 1000)
	svc.CreateAccount(context.Background(), "credit1", "Savings", ledger.AccountTypeAsset, "1002", "VND", 500)

	txn, err := svc.PostTransaction(context.Background(), "debit1", "credit1", 200, "transfer", "order", "ref1", "test transfer")
	if err != nil {
		t.Fatalf("PostTransaction failed: %v", err)
	}
	if txn.Status != ledger.TxnStatusPosted {
		t.Errorf("expected posted, got %s", txn.Status)
	}
	if len(txn.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(txn.Entries))
	}

	debit, _ := svc.GetAccountBalance(context.Background(), "debit1")
	if debit.Balance != 800 {
		t.Errorf("expected debit balance 800, got %d", debit.Balance)
	}
	credit, _ := svc.GetAccountBalance(context.Background(), "credit1")
	if credit.Balance != 700 {
		t.Errorf("expected credit balance 700, got %d", credit.Balance)
	}
}

func TestLedgerDoubleEntryBalancing(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "a1", "Account1", ledger.AccountTypeAsset, "1001", "VND", 10000)
	svc.CreateAccount(context.Background(), "a2", "Account2", ledger.AccountTypeLiability, "2001", "VND", 0)

	txn, _ := svc.PostTransaction(context.Background(), "a1", "a2", 5000, "transfer", "order", "ref1", "test")

	totalDebit := int64(0)
	totalCredit := int64(0)
	for _, e := range txn.Entries {
		if e.EntryType == ledger.EntryDebit {
			totalDebit += e.DebitAmount
		} else {
			totalCredit += e.CreditAmount
		}
	}
	if totalDebit != totalCredit {
		t.Errorf("ledger imbalanced: debit=%d credit=%d", totalDebit, totalCredit)
	}
	if totalDebit != 5000 {
		t.Errorf("expected 5000, got %d", totalDebit)
	}
}

func TestLedgerInsufficientBalance(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "poor", "Poor", ledger.AccountTypeAsset, "1001", "VND", 50)
	svc.CreateAccount(context.Background(), "rich", "Rich", ledger.AccountTypeAsset, "1002", "VND", 1000)
	_, err := svc.PostTransaction(context.Background(), "poor", "rich", 200, "transfer", "order", "ref1", "test")
	if err != ledger.ErrInsufficientBalance {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestLedgerPostZeroAmount(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "a1", "A1", ledger.AccountTypeAsset, "1001", "VND", 100)
	svc.CreateAccount(context.Background(), "a2", "A2", ledger.AccountTypeAsset, "1002", "VND", 100)
	_, err := svc.PostTransaction(context.Background(), "a1", "a2", 0, "transfer", "order", "ref1", "test")
	if err != ledger.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestLedgerPostNegativeAmount(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "a1", "A1", ledger.AccountTypeAsset, "1001", "VND", 100)
	svc.CreateAccount(context.Background(), "a2", "A2", ledger.AccountTypeAsset, "1002", "VND", 100)
	_, err := svc.PostTransaction(context.Background(), "a1", "a2", -100, "transfer", "order", "ref1", "test")
	if err != ledger.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestLedgerGetAccountStatement(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "a1", "A1", ledger.AccountTypeAsset, "1001", "VND", 1000)
	svc.CreateAccount(context.Background(), "a2", "A2", ledger.AccountTypeAsset, "1002", "VND", 0)
	svc.PostTransaction(context.Background(), "a1", "a2", 300, "transfer", "order", "ref1", "test")

	entries, err := svc.GetAccountStatement(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestLedgerAccountNotFound(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	_, err := svc.GetAccountBalance(context.Background(), "nonexistent")
	if err != ledger.ErrAccountNotFound {
		t.Errorf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestLedgerGetAccount(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "acct1", "TestAcct", ledger.AccountTypeAsset, "1000", "VND", 999)
	a, err := svc.GetAccount(context.Background(), "acct1")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "TestAcct" {
		t.Errorf("expected TestAcct, got %s", a.Name)
	}
}

func TestLedgerMultipleTransactionsBalanceSum(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "a1", "A1", ledger.AccountTypeAsset, "1001", "VND", 10000)
	svc.CreateAccount(context.Background(), "a2", "A2", ledger.AccountTypeAsset, "1002", "VND", 0)
	svc.PostTransaction(context.Background(), "a1", "a2", 1000, "transfer", "order", "ref1", "")
	svc.PostTransaction(context.Background(), "a1", "a2", 2000, "transfer", "order", "ref2", "")
	svc.PostTransaction(context.Background(), "a1", "a2", 500, "transfer", "order", "ref3", "")

	a1, _ := svc.GetAccountBalance(context.Background(), "a1")
	if a1.Balance != 6500 {
		t.Errorf("expected 6500, got %d", a1.Balance)
	}
	a2, _ := svc.GetAccountBalance(context.Background(), "a2")
	if a2.Balance != 3500 {
		t.Errorf("expected 3500, got %d", a2.Balance)
	}
}

func TestPayoutCreate(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	p, err := ps.CreatePayout(context.Background(), "seller1", 1000, 50, payout.MethodBankTransfer, "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("CreatePayout failed: %v", err)
	}
	if p.Status != payout.StatusPending {
		t.Errorf("expected pending, got %s", p.Status)
	}
	if p.NetAmount != 950 {
		t.Errorf("expected net 950, got %d", p.NetAmount)
	}
}

func TestPayoutCreateInvalidAmount(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	_, err := ps.CreatePayout(context.Background(), "seller1", 0, 0, payout.MethodBankTransfer, "2026-01-01", "2026-01-31")
	if err != payout.ErrInvalidPayoutAmount {
		t.Errorf("expected ErrInvalidPayoutAmount, got %v", err)
	}
}

func TestPayoutProcess(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	p, _ := ps.CreatePayout(context.Background(), "seller1", 500, 0, payout.MethodWallet, "2026-01-01", "2026-01-31")
	p, err := ps.ProcessPayout(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("ProcessPayout failed: %v", err)
	}
	if p.Status != payout.StatusCompleted {
		t.Errorf("expected completed, got %s", p.Status)
	}
	if p.CompletedAt == "" {
		t.Error("expected completed_at timestamp")
	}
}

func TestPayoutProcessInvalidStatus(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	p, _ := ps.CreatePayout(context.Background(), "seller1", 500, 0, payout.MethodWallet, "", "")
	p, _ = ps.ProcessPayout(context.Background(), p.ID)
	_, err := ps.ProcessPayout(context.Background(), p.ID)
	if err != payout.ErrInvalidPayoutStatus {
		t.Errorf("expected ErrInvalidPayoutStatus, got %v", err)
	}
}

func TestPayoutBatch(t *testing.T) {
	pr := payout.NewInMemoryPayoutRepo()
	br := payout.NewInMemoryBatchRepo()
	ps := payout.NewService(pr, br)

	p1, _ := ps.CreatePayout(context.Background(), "seller1", 1000, 50, payout.MethodBankTransfer, "2026-01-01", "2026-01-31")
	p2, _ := ps.CreatePayout(context.Background(), "seller2", 2000, 100, payout.MethodBankTransfer, "2026-01-01", "2026-01-31")

	p1r, _ := pr.GetByID(context.Background(), p1.ID)
	p2r, _ := pr.GetByID(context.Background(), p2.ID)

	batch, err := ps.BatchPayout(context.Background(), []*payout.Payout{p1r, p2r})
	if err != nil {
		t.Fatalf("BatchPayout failed: %v", err)
	}
	if batch.Count != 2 {
		t.Errorf("expected count 2, got %d", batch.Count)
	}
	if batch.Status != payout.StatusCompleted {
		t.Errorf("expected batch completed, got %s", batch.Status)
	}

	completed1, _ := pr.GetByID(context.Background(), p1.ID)
	if completed1.Status != payout.StatusCompleted {
		t.Errorf("expected payout completed, got %s", completed1.Status)
	}
}

func TestPayoutGetByID(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	p, _ := ps.CreatePayout(context.Background(), "seller1", 100, 0, payout.MethodWallet, "", "")
	got, err := ps.GetByID(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %s, got %s", p.ID, got.ID)
	}
}

func TestPayoutGetByIDNotFound(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	_, err := ps.GetByID(context.Background(), "nonexistent")
	if err != payout.ErrPayoutNotFound {
		t.Errorf("expected ErrPayoutNotFound, got %v", err)
	}
}

func TestPayoutList(t *testing.T) {
	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	ps.CreatePayout(context.Background(), "s1", 100, 0, payout.MethodWallet, "", "")
	ps.CreatePayout(context.Background(), "s2", 200, 0, payout.MethodWallet, "", "")
	items, total, err := ps.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestReconciliationRun(t *testing.T) {
	svc := reconciliation.NewService(reconciliation.NewInMemoryRepository(), nil, nil)
	payments := []reconciliation.PaymentRecord{
		{ID: "pay1", Amount: 1000},
		{ID: "pay2", Amount: 2000},
	}
	ledgers := []reconciliation.LedgerRecord{
		{TransactionID: "pay1", Amount: 1000},
		{TransactionID: "pay2", Amount: 2000},
	}
	run, err := svc.RunReconciliation(context.Background(), "2026-01-15", payments, ledgers)
	if err != nil {
		t.Fatalf("RunReconciliation failed: %v", err)
	}
	if run.Status != reconciliation.RunStatusCompleted {
		t.Errorf("expected completed, got %s", run.Status)
	}
	if run.MatchedCount != 2 {
		t.Errorf("expected 2 matched, got %d", run.MatchedCount)
	}
	if run.UnmatchedCount != 0 {
		t.Errorf("expected 0 unmatched, got %d", run.UnmatchedCount)
	}
}

func TestReconciliationMismatch(t *testing.T) {
	svc := reconciliation.NewService(reconciliation.NewInMemoryRepository(), nil, nil)
	payments := []reconciliation.PaymentRecord{
		{ID: "pay1", Amount: 1000},
	}
	ledgers := []reconciliation.LedgerRecord{
		{TransactionID: "pay1", Amount: 900},
	}
	run, err := svc.RunReconciliation(context.Background(), "2026-01-15", payments, ledgers)
	if err != nil {
		t.Fatal(err)
	}
	items, _ := svc.GetUnmatched(context.Background(), run.ID)
	if len(items) != 1 {
		t.Errorf("expected 1 unmatched, got %d", len(items))
	}
	if items[0].Status != reconciliation.ItemStatusDiscrepancy {
		t.Errorf("expected discrepancy, got %s", items[0].Status)
	}
}

func TestReconciliationPaymentMissingInLedger(t *testing.T) {
	svc := reconciliation.NewService(reconciliation.NewInMemoryRepository(), nil, nil)
	payments := []reconciliation.PaymentRecord{
		{ID: "pay1", Amount: 1000},
		{ID: "pay2", Amount: 500},
	}
	ledgers := []reconciliation.LedgerRecord{
		{TransactionID: "pay1", Amount: 1000},
	}
	run, err := svc.RunReconciliation(context.Background(), "2026-01-15", payments, ledgers)
	if err != nil {
		t.Fatal(err)
	}
	items, _ := svc.GetUnmatched(context.Background(), run.ID)
	if len(items) != 1 {
		t.Errorf("expected 1 unmatched, got %d", len(items))
	}
	if items[0].Status != reconciliation.ItemStatusUnmatched {
		t.Errorf("expected unmatched, got %s", items[0].Status)
	}
}

func TestReconciliationLedgerMissingInSource(t *testing.T) {
	svc := reconciliation.NewService(reconciliation.NewInMemoryRepository(), nil, nil)
	payments := []reconciliation.PaymentRecord{
		{ID: "pay1", Amount: 1000},
	}
	ledgers := []reconciliation.LedgerRecord{
		{TransactionID: "pay1", Amount: 1000},
		{TransactionID: "pay_extra", Amount: 200},
	}
	run, err := svc.RunReconciliation(context.Background(), "2026-01-15", payments, ledgers)
	if err != nil {
		t.Fatal(err)
	}
	unmatched, _ := svc.GetUnmatched(context.Background(), run.ID)
	if len(unmatched) != 1 {
		t.Errorf("expected 1 unmatched (extra ledger entry), got %d", len(unmatched))
	}
}

func TestReconciliationReport(t *testing.T) {
	svc := reconciliation.NewService(reconciliation.NewInMemoryRepository(), nil, nil)
	run, _ := svc.RunReconciliation(context.Background(), "2026-01-15", nil, nil)
	report, err := svc.GenerateReport(context.Background(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.ID != run.ID {
		t.Errorf("expected id %s, got %s", run.ID, report.ID)
	}
}

func TestDisputeOpen(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, err := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "item not received", 500)
	if err != nil {
		t.Fatalf("OpenDispute failed: %v", err)
	}
	if d.Status != dispute.StatusOpened {
		t.Errorf("expected opened, got %s", d.Status)
	}
	if d.OpenedAt == "" {
		t.Error("expected opened_at timestamp")
	}
}

func TestDisputeSubmitEvidence(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "defective", 300)
	d, err := svc.SubmitEvidence(context.Background(), d.ID, "photo_of_damage.jpg")
	if err != nil {
		t.Fatalf("SubmitEvidence failed: %v", err)
	}
	if d.Status != dispute.StatusUnderReview {
		t.Errorf("expected under_review, got %s", d.Status)
	}
	if len(d.Evidence) != 1 {
		t.Errorf("expected 1 evidence item, got %d", len(d.Evidence))
	}
}

func TestDisputeResolveFullRefund(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "wrong item", 200)
	d, err := svc.Resolve(context.Background(), d.ID, dispute.ResolutionFullRefund, "accepted")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if d.Status != dispute.StatusResolved {
		t.Errorf("expected resolved, got %s", d.Status)
	}
	if d.Resolution != dispute.ResolutionFullRefund {
		t.Errorf("expected full_refund, got %s", d.Resolution)
	}
	if d.ResolvedAt == "" {
		t.Error("expected resolved_at timestamp")
	}
}

func TestDisputeResolveNoRefund(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "changed mind", 100)
	d, err := svc.Resolve(context.Background(), d.ID, dispute.ResolutionNoRefund, "not valid")
	if err != nil {
		t.Fatal(err)
	}
	if d.Resolution != dispute.ResolutionNoRefund {
		t.Errorf("expected no_refund, got %s", d.Resolution)
	}
}

func TestDisputeAppeal(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "defective", 150)
	d, _ = svc.Resolve(context.Background(), d.ID, dispute.ResolutionNoRefund, "rejected")
	d, err := svc.Appeal(context.Background(), d.ID, "new evidence available")
	if err != nil {
		t.Fatalf("Appeal failed: %v", err)
	}
	if d.Status != dispute.StatusUnderReview {
		t.Errorf("expected under_review after appeal, got %s", d.Status)
	}
}

func TestDisputeAppealNotResolved(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "user1", "test", 100)
	_, err := svc.Appeal(context.Background(), d.ID, "appealing from opened")
	if err != dispute.ErrInvalidDisputeStatus {
		t.Errorf("expected ErrInvalidDisputeStatus, got %v", err)
	}
}

func TestDisputeList(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	svc.OpenDispute(context.Background(), "txn1", "pay1", "u1", "reason1", 100)
	svc.OpenDispute(context.Background(), "txn2", "pay2", "u2", "reason2", 200)
	items, total, err := svc.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestDisputeGetByIDNotFound(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	_, _, err := svc.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDisputeResolveAlreadyClosed(t *testing.T) {
	svc := dispute.NewService(dispute.NewInMemoryRepository())
	d, _ := svc.OpenDispute(context.Background(), "txn1", "pay1", "u1", "reason", 100)
	_, _ = svc.Resolve(context.Background(), d.ID, dispute.ResolutionFullRefund, "")
	_, _ = svc.Resolve(context.Background(), d.ID, dispute.ResolutionNoRefund, "")
}

func bootstrapLedgerData(t *testing.T) *ledger.Service {
	t.Helper()
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	svc.CreateAccount(context.Background(), "asset1", "Cash", ledger.AccountTypeAsset, "1001", "VND", 5000)
	svc.CreateAccount(context.Background(), "liab1", "Payable", ledger.AccountTypeLiability, "2001", "VND", 0)
	svc.CreateAccount(context.Background(), "rev1", "Revenue", ledger.AccountTypeRevenue, "3001", "VND", 0)
	svc.CreateAccount(context.Background(), "exp1", "Expenses", ledger.AccountTypeExpense, "4001", "VND", 10000)
	return svc
}

func TestLedgerTransactionTypes(t *testing.T) {
	svc := bootstrapLedgerData(t)

	_, err := svc.PostTransaction(context.Background(), "asset1", "rev1", 1000, "revenue", "order", "ord1", "sale")
	if err != nil {
		t.Fatalf("revenue transaction failed: %v", err)
	}

	_, err = svc.PostTransaction(context.Background(), "asset1", "liab1", 500, "liability", "order", "ord2", "fee hold")
	if err != nil {
		t.Fatalf("liability transaction failed: %v", err)
	}

	_, err = svc.PostTransaction(context.Background(), "exp1", "asset1", 200, "expense", "order", "ord3", "refund fee")
	if err != nil {
		t.Fatalf("expense transaction failed: %v", err)
	}

	a1, _ := svc.GetAccountBalance(context.Background(), "asset1")
	rev, _ := svc.GetAccountBalance(context.Background(), "rev1")
	liab, _ := svc.GetAccountBalance(context.Background(), "liab1")
	exp, _ := svc.GetAccountBalance(context.Background(), "exp1")

	if a1.Balance != 3700 {
		t.Errorf("asset expected 3700, got %d", a1.Balance)
	}
	if rev.Balance != 1000 {
		t.Errorf("revenue expected 1000, got %d", rev.Balance)
	}
	if liab.Balance != 500 {
		t.Errorf("liability expected 500, got %d", liab.Balance)
	}
	if exp.Balance != 9800 {
		t.Errorf("expense expected 9800, got %d", exp.Balance)
	}
}

func TestLedgerVoidTransaction(t *testing.T) {
	svc := bootstrapLedgerData(t)
	txn, _ := svc.PostTransaction(context.Background(), "asset1", "rev1", 500, "revenue", "order", "ord1", "sale")

	voided, err := svc.VoidTransaction(context.Background(), txn.ID)
	if err != nil {
		t.Fatalf("VoidTransaction failed: %v", err)
	}
	if voided.Status != ledger.TxnStatusVoided {
		t.Errorf("expected voided, got %s", voided.Status)
	}

	a, _ := svc.GetAccountBalance(context.Background(), "asset1")
	if a.Balance != 5000 {
		t.Errorf("asset expected back to 5000 after void, got %d", a.Balance)
	}
}

func TestLedgerGetTransaction(t *testing.T) {
	svc := bootstrapLedgerData(t)
	txn, _ := svc.PostTransaction(context.Background(), "asset1", "liab1", 300, "liability", "order", "ord1", "fee hold")
	got, err := svc.GetTransaction(context.Background(), txn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != txn.ID {
		t.Errorf("expected id %s, got %s", txn.ID, got.ID)
	}
}

func TestLedgerListTransactions(t *testing.T) {
	svc := bootstrapLedgerData(t)
	svc.PostTransaction(context.Background(), "asset1", "rev1", 100, "revenue", "order", "ord1", "")
	svc.PostTransaction(context.Background(), "asset1", "rev1", 200, "revenue", "order", "ord2", "")
	items, total, err := svc.ListTransactions(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestLedgerInvalidAccountPosting(t *testing.T) {
	svc := ledger.NewService(
		ledger.NewInMemoryAccountRepo(),
		ledger.NewInMemoryTransactionRepo(),
		ledger.NewInMemoryEntryRepo(),
	)
	_, err := svc.PostTransaction(context.Background(), "nonexistent", "also_missing", 100, "transfer", "order", "ref1", "")
	if err != ledger.ErrAccountNotFound {
		t.Errorf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestLedgerCreateDuplicateAccount(t *testing.T) {
	repo := ledger.NewInMemoryAccountRepo()
	repo.Create(context.Background(), &ledger.Account{ID: "dup1", Name: "Dup", Type: ledger.AccountTypeAsset, Code: "1000", Currency: "VND"})
	err := repo.Create(context.Background(), &ledger.Account{ID: "dup1", Name: "Dup2", Type: ledger.AccountTypeAsset, Code: "1000", Currency: "VND"})
	if err != ledger.ErrAccountAlreadyExists {
		t.Errorf("expected ErrAccountAlreadyExists, got %v", err)
	}
}

func TestAllServicesCreateUniqueIDs(t *testing.T) {
	pr := payment.NewInMemoryRepository()
	p1, _ := payment.NewService(pr).Process(context.Background(), "o1", "u1", 100, 0, "VND", payment.MethodWallet)
	p2, _ := payment.NewService(pr).Process(context.Background(), "o2", "u2", 200, 0, "VND", payment.MethodWallet)
	if p1.ID == p2.ID {
		t.Error("expected unique payment IDs")
	}

	ps := payout.NewService(payout.NewInMemoryPayoutRepo(), payout.NewInMemoryBatchRepo())
	po1, _ := ps.CreatePayout(context.Background(), "s1", 100, 0, payout.MethodWallet, "", "")
	po2, _ := ps.CreatePayout(context.Background(), "s2", 200, 0, payout.MethodWallet, "", "")
	if po1.ID == po2.ID {
		t.Error("expected unique payout IDs")
	}

	ds := dispute.NewService(dispute.NewInMemoryRepository())
	d1, _ := ds.OpenDispute(context.Background(), "txn1", "pay1", "u1", "reason1", 100)
	d2, _ := ds.OpenDispute(context.Background(), "txn2", "pay2", "u2", "reason2", 200)
	if d1.ID == d2.ID {
		t.Error("expected unique dispute IDs")
	}
}

func TestPaymentListPagination(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	for i := 0; i < 10; i++ {
		svc.Process(context.Background(), fmt.Sprintf("o%d", i), "u1", 100, 0, "VND", payment.MethodWallet)
	}
	page1, total, err := svc.List(context.Background(), 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(page1) != 3 {
		t.Errorf("expected 3 items on page 1, got %d", len(page1))
	}
	page2, _, _ := svc.List(context.Background(), 3, 3)
	if len(page2) != 3 {
		t.Errorf("expected 3 items on page 2, got %d", len(page2))
	}
}

func TestPaymentProcessAmountEdgeCases(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())

	p, err := svc.Process(context.Background(), "o1", "u1", 1, 0, "VND", payment.MethodWallet)
	if err != nil {
		t.Fatalf("minimum amount failed: %v", err)
	}
	if p.Amount != 1 {
		t.Errorf("expected 1, got %d", p.Amount)
	}

	_, err = svc.Process(context.Background(), "o2", "u1", -1, 0, "VND", payment.MethodWallet)
	if err != payment.ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount for negative, got %v", err)
	}
}

func TestPaymentGetByOrderEmpty(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	items, err := svc.GetByOrder(context.Background(), "no_such_order")
	if err != nil {
		t.Fatal(err)
	}
	if items != nil && len(items) != 0 {
		t.Errorf("expected empty, got %d", len(items))
	}
}

func TestPaymentAllMethods(t *testing.T) {
	svc := payment.NewService(payment.NewInMemoryRepository())
	methods := []payment.PaymentMethod{payment.MethodCreditCard, payment.MethodBankTransfer, payment.MethodWallet, payment.MethodCOD}
	for _, m := range methods {
		_, err := svc.Process(context.Background(), "o1", "u1", 100, 0, "VND", m)
		if err != nil {
			t.Errorf("method %s should be valid, got error: %v", m, err)
		}
	}
}
