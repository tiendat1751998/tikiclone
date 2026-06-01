package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/aiml/internal/training"
)

func TestTrainingJobCreate(t *testing.T) {
	repo := training.NewInMemoryRepository()
	svc := training.NewService(repo)

	job := &training.TrainingJob{
		ID: uuid.New().String(), Name: "train-1", ModelName: "rec-v2",
		Dataset: "user_events_2024", Hyperparameters: map[string]string{"lr": "0.001"},
	}
	err := svc.Create(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != training.StatusPending {
		t.Errorf("expected pending, got %s", job.Status)
	}
}

func TestTrainingJobLifecycle(t *testing.T) {
	repo := training.NewInMemoryRepository()
	svc := training.NewService(repo)

	job := &training.TrainingJob{ID: uuid.New().String(), Name: "lifecycle", ModelName: "m", Dataset: "d"}
	svc.Create(context.Background(), job)

	err := svc.Start(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	j, _ := svc.Get(context.Background(), job.ID)
	if j.Status != training.StatusRunning {
		t.Errorf("expected running, got %s", j.Status)
	}

	metrics := map[string]float64{"accuracy": 0.95, "loss": 0.05}
	err = svc.Complete(context.Background(), job.ID, metrics)
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	j, _ = svc.Get(context.Background(), job.ID)
	if j.Status != training.StatusCompleted {
		t.Errorf("expected completed, got %s", j.Status)
	}
	if j.Metrics["accuracy"] != 0.95 {
		t.Errorf("expected accuracy 0.95, got %f", j.Metrics["accuracy"])
	}
}

func TestTrainingJobFail(t *testing.T) {
	repo := training.NewInMemoryRepository()
	svc := training.NewService(repo)

	job := &training.TrainingJob{ID: uuid.New().String(), Name: "fail-job", ModelName: "m", Dataset: "d"}
	svc.Create(context.Background(), job)
	svc.Start(context.Background(), job.ID)

	err := svc.Fail(context.Background(), job.ID, "OOM error")
	if err != nil {
		t.Fatalf("fail failed: %v", err)
	}
	j, _ := svc.Get(context.Background(), job.ID)
	if j.Status != training.StatusFailed {
		t.Errorf("expected failed, got %s", j.Status)
	}
	if j.Error != "OOM error" {
		t.Errorf("expected OOM error, got %s", j.Error)
	}
}

func TestTrainingJobInvalidTransition(t *testing.T) {
	svc := training.NewService(training.NewInMemoryRepository())

	job := &training.TrainingJob{ID: uuid.New().String(), Name: "bad", ModelName: "m", Dataset: "d"}
	svc.Create(context.Background(), job)

	err := svc.Complete(context.Background(), job.ID, nil)
	if err != training.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestTrainingJobList(t *testing.T) {
	svc := training.NewService(training.NewInMemoryRepository())
	svc.Create(context.Background(), &training.TrainingJob{ID: uuid.New().String(), Name: "j1", ModelName: "m", Dataset: "d"})
	svc.Create(context.Background(), &training.TrainingJob{ID: uuid.New().String(), Name: "j2", ModelName: "m", Dataset: "d"})

	jobs, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestTrainingJobNotFound(t *testing.T) {
	svc := training.NewService(training.NewInMemoryRepository())
	_, err := svc.Get(context.Background(), "nonexistent")
	if err != training.ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestTrainingJobDuplicateID(t *testing.T) {
	svc := training.NewService(training.NewInMemoryRepository())
	id := uuid.New().String()
	svc.Create(context.Background(), &training.TrainingJob{ID: id, Name: "j1", ModelName: "m", Dataset: "d"})
	err := svc.Create(context.Background(), &training.TrainingJob{ID: id, Name: "j2", ModelName: "m", Dataset: "d"})
	if err != training.ErrJobExists {
		t.Errorf("expected ErrJobExists, got %v", err)
	}
}
