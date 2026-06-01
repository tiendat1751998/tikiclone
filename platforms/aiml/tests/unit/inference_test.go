package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/aiml/internal/inference"
)

func TestInferencePredict(t *testing.T) {
	p := inference.NewMockPredictor()
	svc := inference.NewService(p)

	req := &inference.InferenceRequest{
		ModelName: "fraud-detector",
		Input:     map[string]interface{}{"amount": 100.0, "user_age": 25.0},
	}
	result, err := svc.Predict(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output == nil {
		t.Fatal("expected non-nil output")
	}
	if _, ok := result.Output["amount_prediction"]; !ok {
		t.Error("expected amount_prediction in output")
	}
	if result.Confidence <= 0 || result.Confidence > 1 {
		t.Errorf("expected confidence between 0 and 1, got %f", result.Confidence)
	}
}

func TestInferenceBatchPredict(t *testing.T) {
	p := inference.NewMockPredictor()
	svc := inference.NewService(p)

	reqs := []*inference.InferenceRequest{
		{ModelName: "m1", Input: map[string]interface{}{"x": 1.0}},
		{ModelName: "m1", Input: map[string]interface{}{"x": 2.0}},
	}
	results, err := svc.BatchPredict(context.Background(), reqs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInferenceEmptyInput(t *testing.T) {
	p := inference.NewMockPredictor()
	svc := inference.NewService(p)

	_, err := svc.Predict(context.Background(), &inference.InferenceRequest{
		ModelName: "m1",
	})
	if err != inference.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestInferenceModelInfo(t *testing.T) {
	p := inference.NewMockPredictor()
	p.RegisterModel(&inference.ModelInfo{
		Name: "rec-v1", Version: "1.0", InputSchema: "user_features", OutputSchema: "score",
	})

	info, err := p.GetModelInfo(context.Background(), "rec-v1", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.InputSchema != "user_features" {
		t.Errorf("expected user_features, got %s", info.InputSchema)
	}
}

func TestInferenceModelNotFound(t *testing.T) {
	p := inference.NewMockPredictor()
	_, err := p.GetModelInfo(context.Background(), "nonexistent", "1.0")
	if err != inference.ErrModelNotFound {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}

func TestInferenceResultShape(t *testing.T) {
	p := inference.NewMockPredictor()
	svc := inference.NewService(p)

	result, err := svc.Predict(context.Background(), &inference.InferenceRequest{
		ModelName: "m", ModelVersion: "2.0",
		Input: map[string]interface{}{"feature_a": 10.0, "feature_b": 20.0},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ModelVersionUsed != "2.0" {
		t.Errorf("expected version 2.0, got %s", result.ModelVersionUsed)
	}
	if len(result.Output) != 2 {
		t.Errorf("expected 2 output keys, got %d", len(result.Output))
	}
}
