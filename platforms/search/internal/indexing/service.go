package indexing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/search/internal/events"
	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

type Service interface {
	IndexDocument(ctx context.Context, doc *search.ProductDocument, idempotencyKey string) (*IndexTask, error)
	BulkIndex(ctx context.Context, docs []*search.ProductDocument) (*BulkIndexResult, error)
	GetTask(ctx context.Context, id string) (*IndexTask, error)
	ListTasks(ctx context.Context, limit, offset int) ([]*IndexTask, error)
}

type service struct {
	repo     Repository
	search   search.Repository
	publisher events.Publisher
}

func NewService(repo Repository, search search.Repository, publisher events.Publisher) Service {
	return &service{repo: repo, search: search, publisher: publisher}
}

func (s *service) IndexDocument(ctx context.Context, doc *search.ProductDocument, idempotencyKey string) (*IndexTask, error) {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}

	if idempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, idempotencyKey)
		if err == nil && existing != nil {
			return existing, ErrDuplicateIdempotency
		}
	}

	task := &IndexTask{
		ID:             uuid.New().String(),
		DocumentID:     doc.ID,
		Status:         StatusPending,
		IdempotencyKey: idempotencyKey,
		Document:       doc,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		s.publisher.Publish(ctx, events.EventIndexTaskCreated, &events.IndexTaskCreatedEvent{
			TaskID:         task.ID,
			DocumentID:     doc.ID,
			IdempotencyKey: idempotencyKey,
			CreatedAt:      task.CreatedAt,
		})
	}

	task.Status = StatusIndexing
	s.repo.UpdateTask(ctx, task)

	if err := s.search.Index(ctx, doc); err != nil {
		task.Status = StatusFailed
		task.Error = err.Error()
		s.repo.UpdateTask(ctx, task)
		return task, ErrIndexingFailed
	}

	task.Status = StatusIndexed
	s.repo.UpdateTask(ctx, task)

	if s.publisher != nil {
		s.publisher.Publish(ctx, events.EventDocumentIndexed, &events.DocumentIndexedEvent{
			DocumentID: doc.ID,
			Title:      doc.Title,
			Category:   doc.Category,
			SellerID:   doc.SellerID,
			IndexedAt:  time.Now(),
		})
	}

	return task, nil
}

func (s *service) BulkIndex(ctx context.Context, docs []*search.ProductDocument) (*BulkIndexResult, error) {
	result := &BulkIndexResult{Total: len(docs)}

	for _, doc := range docs {
		_, err := s.IndexDocument(ctx, doc, "")
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Indexed++
		}
	}

	return result, nil
}

func (s *service) GetTask(ctx context.Context, id string) (*IndexTask, error) {
	return s.repo.GetTask(ctx, id)
}

func (s *service) ListTasks(ctx context.Context, limit, offset int) ([]*IndexTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListTasks(ctx, limit, offset)
}
