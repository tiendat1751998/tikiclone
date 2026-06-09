package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAccountNotFound    = errors.New("ledger account not found")
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrEntryNotFound      = errors.New("journal entry not found")
	ErrBatchNotFound      = errors.New("settlement batch not found")
	ErrReconcileNotFound  = errors.New("reconciliation record not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrWalletFrozen       = errors.New("wallet is frozen")
	ErrWalletClosed       = errors.New("wallet is closed")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrEntryAlreadyPosted = errors.New("entry already posted")
	ErrEntryAlreadyReversed = errors.New("entry already reversed")
	ErrDebitCreditMismatch = errors.New("debit and credit must reference different accounts")
	ErrUnauthorized       = errors.New("unauthorized")
)
