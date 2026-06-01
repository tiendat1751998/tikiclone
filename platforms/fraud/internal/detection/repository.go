package detection

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/fraud/internal/core"
)

type Repository interface {
	SaveAlert(ctx context.Context, alert *FraudAlert) error
	GetAlert(ctx context.Context, id string) (*FraudAlert, error)
	ListAlerts(ctx context.Context, status string, riskLevel core.RiskLevel, offset, limit int) ([]*FraudAlert, int, error)
	UpdateAlert(ctx context.Context, alert *FraudAlert) error
}

type InMemoryRepository struct {
	mu     sync.RWMutex
	alerts map[string]*FraudAlert
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		alerts: make(map[string]*FraudAlert),
	}
}

func (r *InMemoryRepository) SaveAlert(ctx context.Context, alert *FraudAlert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}
	r.alerts[alert.ID] = alert
	return nil
}

func (r *InMemoryRepository) GetAlert(ctx context.Context, id string) (*FraudAlert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	alert, ok := r.alerts[id]
	if !ok {
		return nil, ErrAlertNotFound
	}
	return alert, nil
}

func (r *InMemoryRepository) ListAlerts(ctx context.Context, status string, riskLevel core.RiskLevel, offset, limit int) ([]*FraudAlert, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*FraudAlert
	for _, a := range r.alerts {
		if status != "" && a.Status != status {
			continue
		}
		if riskLevel != "" && a.RiskLevel != riskLevel {
			continue
		}
		filtered = append(filtered, a)
	}

	total := len(filtered)
	if offset >= total {
		return []*FraudAlert{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (r *InMemoryRepository) UpdateAlert(ctx context.Context, alert *FraudAlert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alerts[alert.ID] = alert
	return nil
}
