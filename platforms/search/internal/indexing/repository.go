package indexing

import (
	"context"
	"sync"
	"time"

	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type Repository interface {
	CreateTask(ctx context.Context, task *IndexTask) error
	GetTask(ctx context.Context, id string) (*IndexTask, error)
	UpdateTask(ctx context.Context, task *IndexTask) error
	ListTasks(ctx context.Context, limit, offset int) ([]*IndexTask, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*IndexTask, error)
	DeleteTask(ctx context.Context, id string) error
}

type InMemoryRepository struct {
	mu    sync.RWMutex
	tasks map[string]*IndexTask
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		tasks: make(map[string]*IndexTask),
	}
}

func (r *InMemoryRepository) CreateTask(ctx context.Context, task *IndexTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task.IdempotencyKey != "" {
		for _, existing := range r.tasks {
			if existing.IdempotencyKey == task.IdempotencyKey {
				return ErrDuplicateIdempotency
			}
		}
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Status == "" {
		task.Status = StatusPending
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryRepository) GetTask(ctx context.Context, id string) (*IndexTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (r *InMemoryRepository) UpdateTask(ctx context.Context, task *IndexTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return ErrTaskNotFound
	}
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryRepository) ListTasks(ctx context.Context, limit, offset int) ([]*IndexTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*IndexTask
	for _, task := range r.tasks {
		all = append(all, task)
	}

	if offset >= len(all) {
		return []*IndexTask{}, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], nil
}

func (r *InMemoryRepository) FindByIdempotencyKey(ctx context.Context, key string) (*IndexTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, task := range r.tasks {
		if task.IdempotencyKey == key {
			return task, nil
		}
	}
	return nil, ErrTaskNotFound
}

func (r *InMemoryRepository) DeleteTask(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}

func NewSearchRepositoryFromIndexing(searchRepo search.Repository) search.Repository {
	return searchRepo
}
