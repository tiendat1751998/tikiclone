package dispatcher

import (
	"time"

	"github.com/tikiclone/tiki/platforms/notification/internal/notifier"
)

type DispatchJob struct {
	ID             string              `json:"id"`
	NotificationID string              `json:"notification_id"`
	UserID         string              `json:"user_id"`
	Channel        notifier.Channel    `json:"channel"`
	Type           notifier.NotificationType `json:"type"`
	Title          string              `json:"title"`
	Body           string              `json:"body"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Priority       notifier.Priority   `json:"priority"`
	RetryCount     int                 `json:"retry_count"`
	MaxRetries     int                 `json:"max_retries"`
	Status         string              `json:"status"`
	Error          string              `json:"error,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type DispatchResult struct {
	JobID          string `json:"job_id"`
	Success        bool   `json:"success"`
	Channel        string `json:"channel"`
	Error          string `json:"error,omitempty"`
	Attempt        int    `json:"attempt"`
}

type RetryPolicy struct {
	MaxRetries       int           `json:"max_retries"`
	BaseDelay        time.Duration `json:"base_delay"`
	MaxDelay         time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:       3,
		BaseDelay:        1 * time.Second,
		MaxDelay:         30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}
