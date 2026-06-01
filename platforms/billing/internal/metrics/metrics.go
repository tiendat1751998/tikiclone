package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TransactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_billing_transactions_total", Help: "Total transactions by type",
	}, []string{"type"})
	TransactionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "tiki_billing_transaction_duration_ms", Help: "Transaction duration",
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
	LedgerEntriesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_ledger_entries_total", Help: "Total ledger entries",
	})
	SettlementsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_settlements_total", Help: "Total settlements completed",
	})
	PayoutsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_payouts_total", Help: "Total payouts",
	})
	RefundsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_refunds_total", Help: "Total refunds",
	})
	ReconciliationMismatches = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_reconciliation_mismatches_total", Help: "Total mismatches",
	})
	WalletBalanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_billing_wallet_balance", Help: "Wallet balance by type",
	}, []string{"wallet_type", "currency"})
	IdempotencyHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_billing_idempotency_hits_total", Help: "Idempotency key reuse count",
	})
)
