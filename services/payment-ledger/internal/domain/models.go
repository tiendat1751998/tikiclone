package domain

import "time"

type AccountType string

const (
	AccountAsset    AccountType = "asset"
	AccountLiability AccountType = "liability"
	AccountEquity   AccountType = "equity"
	AccountRevenue  AccountType = "revenue"
	AccountExpense  AccountType = "expense"
)

type LedgerAccount struct {
	ID              string      `db:"id" json:"id"`
	AccountCode     string      `db:"account_code" json:"account_code"`
	AccountName     string      `db:"account_name" json:"account_name"`
	AccountType     AccountType `db:"account_type" json:"account_type"`
	ParentAccountID *string     `db:"parent_account_id" json:"parent_account_id,omitempty"`
	Currency        string      `db:"currency" json:"currency"`
	Balance         int64       `db:"balance" json:"balance"`
	IsActive        bool        `db:"is_active" json:"is_active"`
	Description     *string     `db:"description" json:"description,omitempty"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}

type ReferenceType string

const (
	RefPayment    ReferenceType = "payment"
	RefRefund     ReferenceType = "refund"
	RefSettlement ReferenceType = "settlement"
	RefAdjustment ReferenceType = "adjustment"
	RefFee        ReferenceType = "fee"
	RefCommission ReferenceType = "commission"
	RefWithdrawal ReferenceType = "withdrawal"
	RefDeposit    ReferenceType = "deposit"
	RefTransfer   ReferenceType = "transfer"
)

type EntryStatus string

const (
	EntryPending  EntryStatus = "pending"
	EntryPosted   EntryStatus = "posted"
	EntryReversed EntryStatus = "reversed"
	EntryFailed   EntryStatus = "failed"
)

type JournalEntry struct {
	ID              string        `db:"id" json:"id"`
	EntryNumber     string        `db:"entry_number" json:"entry_number"`
	TransactionID   string        `db:"transaction_id" json:"transaction_id"`
	ReferenceType   ReferenceType `db:"reference_type" json:"reference_type"`
	ReferenceID     string        `db:"reference_id" json:"reference_id"`
	DebitAccountID  string        `db:"debit_account_id" json:"debit_account_id"`
	CreditAccountID string        `db:"credit_account_id" json:"credit_account_id"`
	Amount          int64         `db:"amount" json:"amount"`
	Currency        string        `db:"currency" json:"currency"`
	Description     *string       `db:"description" json:"description,omitempty"`
	Status          EntryStatus   `db:"status" json:"status"`
	PostedAt        *time.Time    `db:"posted_at" json:"posted_at,omitempty"`
	ReversedAt      *time.Time    `db:"reversed_at" json:"reversed_at,omitempty"`
	ReversalEntryID *string       `db:"reversal_entry_id" json:"reversal_entry_id,omitempty"`
	CreatedAt       time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at" json:"updated_at"`
}

type WalletType string

const (
	WalletBuyer     WalletType = "buyer"
	WalletSeller    WalletType = "seller"
	WalletPlatform  WalletType = "platform"
	WalletEscrow    WalletType = "escrow"
	WalletPromotion WalletType = "promotion"
)

type WalletStatus string

const (
	WalletActive   WalletStatus = "active"
	WalletFrozen   WalletStatus = "frozen"
	WalletSuspended WalletStatus = "suspended"
	WalletClosed   WalletStatus = "closed"
)

type WalletAccount struct {
	ID                string       `db:"id" json:"id"`
	UserID            string       `db:"user_id" json:"user_id"`
	WalletType        WalletType   `db:"wallet_type" json:"wallet_type"`
	Currency          string       `db:"currency" json:"currency"`
	Balance           int64        `db:"balance" json:"balance"`
	FrozenBalance     int64        `db:"frozen_balance" json:"frozen_balance"`
	Status            WalletStatus `db:"status" json:"status"`
	LastTransactionAt *time.Time   `db:"last_transaction_at" json:"last_transaction_at,omitempty"`
	CreatedAt         time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time    `db:"updated_at" json:"updated_at"`
}

type WalletTxType string

const (
	WalletTxCredit   WalletTxType = "credit"
	WalletTxDebit    WalletTxType = "debit"
	WalletTxFreeze   WalletTxType = "freeze"
	WalletTxUnfreeze WalletTxType = "unfreeze"
	WalletTxAdjust   WalletTxType = "adjustment"
)

type WalletTxStatus string

const (
	WalletTxPending   WalletTxStatus = "pending"
	WalletTxCompleted WalletTxStatus = "completed"
	WalletTxFailed    WalletTxStatus = "failed"
	WalletTxReversed  WalletTxStatus = "reversed"
)

type WalletTransaction struct {
	ID              string        `db:"id" json:"id"`
	WalletID        string        `db:"wallet_id" json:"wallet_id"`
	TransactionType WalletTxType  `db:"transaction_type" json:"transaction_type"`
	Amount          int64         `db:"amount" json:"amount"`
	BalanceBefore   int64         `db:"balance_before" json:"balance_before"`
	BalanceAfter    int64         `db:"balance_after" json:"balance_after"`
	ReferenceType   ReferenceType `db:"reference_type" json:"reference_type"`
	ReferenceID     *string       `db:"reference_id" json:"reference_id,omitempty"`
	JournalEntryID  *string       `db:"journal_entry_id" json:"journal_entry_id,omitempty"`
	Description     *string       `db:"description" json:"description,omitempty"`
	Status          WalletTxStatus `db:"status" json:"status"`
	CreatedAt       time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at" json:"updated_at"`
}

type SettlementStatus string

const (
	SettlementPending    SettlementStatus = "pending"
	SettlementProcessing SettlementStatus = "processing"
	SettlementCompleted  SettlementStatus = "completed"
	SettlementFailed     SettlementStatus = "failed"
	SettlementCancelled  SettlementStatus = "cancelled"
)

type SettlementBatch struct {
	ID               string           `db:"id" json:"id"`
	BatchNumber      string           `db:"batch_number" json:"batch_number"`
	SellerID         string           `db:"seller_id" json:"seller_id"`
	PeriodStart      time.Time        `db:"period_start" json:"period_start"`
	PeriodEnd        time.Time        `db:"period_end" json:"period_end"`
	TotalSales       int64            `db:"total_sales" json:"total_sales"`
	TotalFees        int64            `db:"total_fees" json:"total_fees"`
	TotalCommissions int64            `db:"total_commissions" json:"total_commissions"`
	TotalAdjustments int64            `db:"total_adjustments" json:"total_adjustments"`
	NetSettlement    int64            `db:"net_settlement" json:"net_settlement"`
	Currency         string           `db:"currency" json:"currency"`
	Status           SettlementStatus `db:"status" json:"status"`
	PaymentMethod    *string          `db:"payment_method" json:"payment_method,omitempty"`
	PaymentReference *string          `db:"payment_reference" json:"payment_reference,omitempty"`
	PaidAt           *time.Time       `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at" json:"updated_at"`
}

type ReconcileStatus string

const (
	ReconcileMatched       ReconcileStatus = "matched"
	ReconcileUnmatched     ReconcileStatus = "unmatched"
	ReconcileInvestigating ReconcileStatus = "investigating"
	ReconcileResolved      ReconcileStatus = "resolved"
)

type ReconciliationRecord struct {
	ID                 string          `db:"id" json:"id"`
	ReconciliationDate time.Time       `db:"reconciliation_date" json:"reconciliation_date"`
	AccountID          string          `db:"account_id" json:"account_id"`
	LedgerBalance      int64           `db:"ledger_balance" json:"ledger_balance"`
	ExternalBalance    int64           `db:"external_balance" json:"external_balance"`
	Difference         int64           `db:"difference" json:"difference"`
	Status             ReconcileStatus `db:"status" json:"status"`
	ResolvedBy         *string         `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolvedAt         *time.Time      `db:"resolved_at" json:"resolved_at,omitempty"`
	Notes              *string         `db:"notes" json:"notes,omitempty"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at" json:"updated_at"`
}
