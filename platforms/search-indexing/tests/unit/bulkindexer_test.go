package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search-indexing/internal/bulkindexer"
)

func TestCreateBulkJob(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, err := svc.CreateJob(ctx, "products", 100)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}
	if job.ID == "" {
		t.Error("expected job ID to be set")
	}
	if job.Status != bulkindexer.JobStatusPending {
		t.Errorf("expected status pending, got %s", job.Status)
	}
	if job.TotalDocuments != 100 {
		t.Errorf("expected total 100, got %d", job.TotalDocuments)
	}
}

func TestSubmitBatch(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, _ := svc.CreateJob(ctx, "products", 10)
	docs := []map[string]interface{}{
		{"id": "1", "title": "Product 1"},
		{"id": "2", "title": "Product 2"},
	}

	batch, err := svc.SubmitBatch(ctx, job.ID, docs, 1)
	if err != nil {
		t.Fatalf("SubmitBatch failed: %v", err)
	}
	if batch.JobID != job.ID {
		t.Errorf("expected job ID %s, got %s", job.ID, batch.JobID)
	}
	if len(batch.Documents) != 2 {
		t.Errorf("expected 2 documents, got %d", len(batch.Documents))
	}

	updatedJob, _ := svc.GetJobProgress(ctx, job.ID)
	if updatedJob.Status != bulkindexer.JobStatusRunning {
		t.Errorf("expected job status running, got %s", updatedJob.Status)
	}
}

func TestMarkBatchProcessed(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, _ := svc.CreateJob(ctx, "products", 4)
	docs := []map[string]interface{}{
		{"id": "1", "title": "A"},
		{"id": "2", "title": "B"},
	}
	batch1, _ := svc.SubmitBatch(ctx, job.ID, docs, 1)

	if err := svc.MarkBatchProcessed(ctx, batch1.ID, 0, nil); err != nil {
		t.Fatalf("MarkBatchProcessed failed: %v", err)
	}

	progress, _ := svc.GetJobProgress(ctx, job.ID)
	if progress.ProcessedCount != 2 {
		t.Errorf("expected 2 processed, got %d", progress.ProcessedCount)
	}
}

func TestBulkJobLifecycle(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, _ := svc.CreateJob(ctx, "products", 6)

	batch1, _ := svc.SubmitBatch(ctx, job.ID, []map[string]interface{}{{"id": "1"}, {"id": "2"}, {"id": "3"}}, 1)
	svc.MarkBatchProcessed(ctx, batch1.ID, 0, nil)

	batch2, _ := svc.SubmitBatch(ctx, job.ID, []map[string]interface{}{{"id": "4"}, {"id": "5"}, {"id": "6"}}, 2)
	svc.MarkBatchProcessed(ctx, batch2.ID, 0, nil)

	progress, _ := svc.GetJobProgress(ctx, job.ID)
	if progress.Status != bulkindexer.JobStatusCompleted {
		t.Errorf("expected job completed, got %s", progress.Status)
	}
	if progress.ProcessedCount != 6 {
		t.Errorf("expected 6 processed, got %d", progress.ProcessedCount)
	}
	if progress.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestGetJobProgress(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, _ := svc.CreateJob(ctx, "products", 10)

	progress, err := svc.GetJobProgress(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobProgress failed: %v", err)
	}
	if progress.ID != job.ID {
		t.Errorf("expected job ID %s, got %s", job.ID, progress.ID)
	}
}

func TestListBulkJobs(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	svc.CreateJob(ctx, "products", 10)
	svc.CreateJob(ctx, "orders", 20)
	svc.CreateJob(ctx, "users", 5)

	jobs, err := svc.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d", len(jobs))
	}
}

func TestSubmitBatchCompletedJob(t *testing.T) {
	repo := bulkindexer.NewInMemoryRepository()
	svc := bulkindexer.NewService(repo)
	ctx := context.Background()

	job, _ := svc.CreateJob(ctx, "products", 1)
	batch, _ := svc.SubmitBatch(ctx, job.ID, []map[string]interface{}{{"id": "1"}}, 1)
	svc.MarkBatchProcessed(ctx, batch.ID, 0, nil)

	_, err := svc.SubmitBatch(ctx, job.ID, []map[string]interface{}{{"id": "2"}}, 2)
	if err != bulkindexer.ErrJobCompleted {
		t.Errorf("expected ErrJobCompleted, got %v", err)
	}
}
