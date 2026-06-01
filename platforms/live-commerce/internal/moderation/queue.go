package moderation

import (
	"context"
	"sync"
	"time"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
	"go.uber.org/zap"
)

type Action struct {
	Action    *domain.ModerationAction
	CreatedAt time.Time
}

type Queue struct {
	mu      sync.Mutex
	items   []*Action
	pending map[string][]*Action
}

func NewQueue() *Queue {
	return &Queue{
		items:   make([]*Action, 0, 100),
		pending: make(map[string][]*Action),
	}
}

func (q *Queue) Enqueue(action *domain.ModerationAction) {
	q.mu.Lock()
	defer q.mu.Unlock()
	entry := &Action{Action: action, CreatedAt: time.Now()}
	q.items = append(q.items, entry)
	q.pending[action.RoomID] = append(q.pending[action.RoomID], entry)
}

func (q *Queue) Dequeue() *domain.ModerationAction {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
		}
	entry := q.items[0]
	q.items = q.items[1:]
	return entry.Action
}

func (q *Queue) DequeueByRoom(roomID string) []*domain.ModerationAction {
	q.mu.Lock()
	defer q.mu.Unlock()
	actions, ok := q.pending[roomID]
	if !ok || len(actions) == 0 {
		return nil
	}
	delete(q.pending, roomID)
	var result []*domain.ModerationAction
	for _, a := range actions {
		result = append(result, a.Action)
	}
	return result
}

func (q *Queue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *Queue) StartWorker(ctx context.Context, processFn func(context.Context, *domain.ModerationAction) error) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				action := q.Dequeue()
				if action == nil {
					continue
				}
				if err := processFn(ctx, action); err != nil {
					observability.LogWithTrace(ctx).Error("moderation worker", zap.Error(err))
				}
			}
		}
	}()
}
