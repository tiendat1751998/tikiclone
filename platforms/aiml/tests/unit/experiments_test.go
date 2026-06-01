package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/aiml/internal/experiments"
)

func TestCreateExperiment(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	exp := &experiments.Experiment{
		ID: "exp-1", Name: "ctr-test", ModelA: "model-a", ModelB: "model-b",
		TrafficPct: 50, Metric: experiments.MetricCTR,
	}
	err := svc.CreateExperiment(context.Background(), exp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetExperiment(context.Background(), "exp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != experiments.StatusRunning {
		t.Errorf("expected running, got %s", got.Status)
	}
}

func TestExperimentAssignVariant(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	exp := &experiments.Experiment{
		ID: "exp-1", Name: "test", ModelA: "a", ModelB: "b", TrafficPct: 50,
		Metric: experiments.MetricConversion,
	}
	svc.CreateExperiment(context.Background(), exp)

	assignment, err := svc.AssignVariant(context.Background(), "exp-1", "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.ExperimentID != "exp-1" {
		t.Errorf("expected exp-1, got %s", assignment.ExperimentID)
	}
	if assignment.Variant != "a" && assignment.Variant != "b" {
		t.Errorf("expected a or b, got %s", assignment.Variant)
	}
}

func TestExperimentConsistentAssignment(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	exp := &experiments.Experiment{
		ID: "exp-1", Name: "test", ModelA: "a", ModelB: "b", TrafficPct: 50,
		Metric: experiments.MetricRevenue,
	}
	svc.CreateExperiment(context.Background(), exp)

	a1, _ := svc.AssignVariant(context.Background(), "exp-1", "user-456")
	a2, _ := svc.AssignVariant(context.Background(), "exp-1", "user-456")
	if a1.Variant != a2.Variant {
		t.Error("same user should get same variant consistently")
	}
}

func TestExperimentRecordAndResults(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	exp := &experiments.Experiment{
		ID: "exp-1", Name: "rev-test", ModelA: "control", ModelB: "treatment",
		TrafficPct: 50, Metric: experiments.MetricRevenue,
	}
	svc.CreateExperiment(context.Background(), exp)

	for i := 0; i < 10; i++ {
		svc.RecordResult(context.Background(), "exp-1", "control", "u1", 100.0)
	}
	for i := 0; i < 10; i++ {
		svc.RecordResult(context.Background(), "exp-1", "treatment", "u2", 150.0)
	}

	results, err := svc.GetExperimentResults(context.Background(), "exp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.AResults.SampleSize != 10 {
		t.Errorf("expected 10 control samples, got %d", results.AResults.SampleSize)
	}
	if results.BResults.SampleSize != 10 {
		t.Errorf("expected 10 treatment samples, got %d", results.BResults.SampleSize)
	}
	if results.BResults.Mean != 150.0 {
		t.Errorf("expected treatment mean 150, got %f", results.BResults.Mean)
	}
	if results.Improvement <= 0 {
		t.Errorf("expected positive improvement, got %f", results.Improvement)
	}
}

func TestExperimentClosed(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	_, err := svc.AssignVariant(context.Background(), "nonexistent", "u1")
	if err != experiments.ErrExperimentNotFound {
		t.Errorf("expected ErrExperimentNotFound, got %v", err)
	}
}

func TestExperimentResultsEmpty(t *testing.T) {
	repo := experiments.NewInMemoryRepository()
	svc := experiments.NewService(repo)

	svc.CreateExperiment(context.Background(), &experiments.Experiment{
		ID: "exp-empty", Name: "empty", ModelA: "a", ModelB: "b", TrafficPct: 50, Metric: experiments.MetricCTR,
	})

	results, err := svc.GetExperimentResults(context.Background(), "exp-empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results.AResults.SampleSize != 0 {
		t.Errorf("expected 0 samples, got %d", results.AResults.SampleSize)
	}
}
