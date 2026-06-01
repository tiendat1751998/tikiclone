package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/search/internal/events"
	"github.com/tikiclone/tiki/platforms/search/internal/indexing"
	"github.com/tikiclone/tiki/platforms/search/internal/search"
)

func TestDocumentIndexing(t *testing.T) {
	searchRepo := search.NewInMemoryRepository()
	indexingRepo := indexing.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := indexing.NewService(indexingRepo, searchRepo, pub)
	ctx := context.Background()

	doc := &search.ProductDocument{
		ID:          uuid.New().String(),
		Title:       "Test Product",
		Description: "Test description",
		Category:    "Electronics",
		Price:       99.99,
		Rating:      4.5,
		Stock:       100,
	}

	task, err := svc.IndexDocument(ctx, doc, "")
	if err != nil {
		t.Fatalf("IndexDocument failed: %v", err)
	}
	if task.Status != indexing.StatusIndexed {
		t.Errorf("Expected status indexed, got %s", task.Status)
	}

	retrieved, err := searchRepo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Title != "Test Product" {
		t.Errorf("Expected title 'Test Product', got '%s'", retrieved.Title)
	}
}

func TestBulkIndexing(t *testing.T) {
	searchRepo := search.NewInMemoryRepository()
	indexingRepo := indexing.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := indexing.NewService(indexingRepo, searchRepo, pub)
	ctx := context.Background()

	docs := make([]*search.ProductDocument, 5)
	for i := 0; i < 5; i++ {
		docs[i] = &search.ProductDocument{
			ID:    uuid.New().String(),
			Title: "Product",
			Price: float64((i + 1) * 10),
		}
	}

	result, err := svc.BulkIndex(ctx, docs)
	if err != nil {
		t.Fatalf("BulkIndex failed: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}
	if result.Indexed != 5 {
		t.Errorf("Expected 5 indexed, got %d", result.Indexed)
	}
	if result.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", result.Failed)
	}

	for _, doc := range docs {
		_, err := searchRepo.GetByID(ctx, doc.ID)
		if err != nil {
			t.Errorf("Document %s not found after bulk index", doc.ID)
		}
	}
}

func TestIndexingIdempotency(t *testing.T) {
	searchRepo := search.NewInMemoryRepository()
	indexingRepo := indexing.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := indexing.NewService(indexingRepo, searchRepo, pub)
	ctx := context.Background()

	doc := &search.ProductDocument{
		ID:    uuid.New().String(),
		Title: "Original",
		Price: 10.0,
	}

	key := "idempotency-key-123"
	task, err := svc.IndexDocument(ctx, doc, key)
	if err != nil {
		t.Fatalf("First IndexDocument failed: %v", err)
	}
	if task.Status != indexing.StatusIndexed {
		t.Errorf("Expected status indexed, got %s", task.Status)
	}

	_, err = svc.IndexDocument(ctx, doc, key)
	if err == nil {
		t.Error("Expected duplicate idempotency key error, got nil")
	}
}

func TestIndexTaskManagement(t *testing.T) {
	indexingRepo := indexing.NewInMemoryRepository()
	searchRepo := search.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := indexing.NewService(indexingRepo, searchRepo, pub)
	ctx := context.Background()

	doc := &search.ProductDocument{
		ID:    uuid.New().String(),
		Title: "Task Test",
		Price: 25.0,
	}

	task, err := svc.IndexDocument(ctx, doc, "")
	if err != nil {
		t.Fatalf("IndexDocument failed: %v", err)
	}

	retrieved, err := svc.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if retrieved.ID != task.ID {
		t.Errorf("Task ID mismatch: %s vs %s", retrieved.ID, task.ID)
	}
	if retrieved.Status != indexing.StatusIndexed {
		t.Errorf("Expected status indexed, got %s", retrieved.Status)
	}
}

func TestListIndexTasks(t *testing.T) {
	indexingRepo := indexing.NewInMemoryRepository()
	searchRepo := search.NewInMemoryRepository()
	pub := events.NewNoOpPublisher()
	svc := indexing.NewService(indexingRepo, searchRepo, pub)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		doc := &search.ProductDocument{
			ID:    uuid.New().String(),
			Title: "Product",
		}
		svc.IndexDocument(ctx, doc, "")
	}

	tasks, err := svc.ListTasks(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(tasks))
	}

	tasks, err = svc.ListTasks(ctx, 2, 0)
	if err != nil {
		t.Fatalf("ListTasks with limit failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks with limit 2, got %d", len(tasks))
	}
}
