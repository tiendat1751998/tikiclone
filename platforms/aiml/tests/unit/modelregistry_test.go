package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/aiml/internal/modelregistry"
)

func TestModelRegistration(t *testing.T) {
	repo := modelregistry.NewInMemoryRepository()
	svc := modelregistry.NewService(repo)

	m := &modelregistry.Model{
		ID: "m1", Name: "rec-model", Version: "1.0",
		Type: modelregistry.TypeRecommendation, Framework: modelregistry.FrameworkPyTorch,
	}
	err := svc.Register(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.Get(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "rec-model" {
		t.Errorf("expected rec-model, got %s", got.Name)
	}
}

func TestModelDuplicateRegistration(t *testing.T) {
	svc := modelregistry.NewService(modelregistry.NewInMemoryRepository())
	m := &modelregistry.Model{ID: "m1", Name: "dup", Version: "1", Type: modelregistry.TypeRanking, Framework: modelregistry.FrameworkONNX}
	svc.Register(context.Background(), m)
	err := svc.Register(context.Background(), m)
	if err != modelregistry.ErrModelExists {
		t.Errorf("expected ErrModelExists, got %v", err)
	}
}

func TestModelNotFound(t *testing.T) {
	svc := modelregistry.NewService(modelregistry.NewInMemoryRepository())
	_, err := svc.Get(context.Background(), "nonexistent")
	if err != modelregistry.ErrModelNotFound {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

func TestModelPromotion(t *testing.T) {
	repo := modelregistry.NewInMemoryRepository()
	svc := modelregistry.NewService(repo)

	svc.Register(context.Background(), &modelregistry.Model{
		ID: "m1", Name: "m", Version: "1", Type: modelregistry.TypeFraud, Framework: modelregistry.FrameworkTensorFlow,
	})

	err := svc.Promote(context.Background(), "m1", modelregistry.StageProduction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	models, _ := svc.ListByStage(context.Background(), modelregistry.StageProduction)
	if len(models) != 1 {
		t.Errorf("expected 1 production model, got %d", len(models))
	}
}

func TestModelArchive(t *testing.T) {
	svc := modelregistry.NewService(modelregistry.NewInMemoryRepository())
	svc.Register(context.Background(), &modelregistry.Model{
		ID: "m1", Name: "m", Version: "1", Type: modelregistry.TypeSearch, Framework: modelregistry.FrameworkONNX,
	})

	err := svc.Archive(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, _ := svc.Get(context.Background(), "m1")
	if m.Status != modelregistry.StageArchived {
		t.Errorf("expected archived, got %s", m.Status)
	}
}

func TestModelPromoteArchivedFails(t *testing.T) {
	svc := modelregistry.NewService(modelregistry.NewInMemoryRepository())
	svc.Register(context.Background(), &modelregistry.Model{
		ID: "m1", Name: "m", Version: "1", Type: modelregistry.TypeRecommendation, Framework: modelregistry.FrameworkPyTorch,
	})
	svc.Archive(context.Background(), "m1")
	err := svc.Promote(context.Background(), "m1", modelregistry.StageProduction)
	if err != modelregistry.ErrInvalidStage {
		t.Errorf("expected ErrInvalidStage, got %v", err)
	}
}

func TestModelList(t *testing.T) {
	svc := modelregistry.NewService(modelregistry.NewInMemoryRepository())
	svc.Register(context.Background(), &modelregistry.Model{ID: "m1", Name: "a", Version: "1", Type: modelregistry.TypeRecommendation, Framework: modelregistry.FrameworkPyTorch})
	svc.Register(context.Background(), &modelregistry.Model{ID: "m2", Name: "b", Version: "1", Type: modelregistry.TypeFraud, Framework: modelregistry.FrameworkTensorFlow})

	models, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}
