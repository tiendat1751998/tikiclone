package indexing

import (
	"time"

	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type IndexStatus string

const (
	StatusPending  IndexStatus = "pending"
	StatusIndexing IndexStatus = "indexing"
	StatusIndexed  IndexStatus = "indexed"
	StatusFailed   IndexStatus = "failed"
)

type IndexTask struct {
	ID             string          `json:"id"`
	DocumentID     string          `json:"document_id"`
	Status         IndexStatus     `json:"status"`
	IdempotencyKey string          `json:"idempotency_key"`
	Document       *search.ProductDocument `json:"document,omitempty"`
	Error          string          `json:"error,omitempty"`
	RetryCount     int             `json:"retry_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type BulkIndexResult struct {
	Total    int    `json:"total"`
	Indexed  int    `json:"indexed"`
	Failed   int    `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
