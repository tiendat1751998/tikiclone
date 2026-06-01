package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/aiml/internal/featurestore"
)

func TestFeatureRegistration(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	f := &featurestore.Feature{
		Name:      "user_age",
		ValueType: featurestore.TypeNumber,
		Entity:    featurestore.EntityUser,
		Source:    "profile",
	}
	err := svc.RegisterFeature(context.Background(), f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetFeature(context.Background(), "user_age")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ValueType != featurestore.TypeNumber {
		t.Errorf("expected number, got %s", got.ValueType)
	}
}

func TestFeatureDuplicateRegistration(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	f := &featurestore.Feature{Name: "dup_feat", ValueType: featurestore.TypeString, Entity: featurestore.EntityProduct}
	svc.RegisterFeature(context.Background(), f)
	err := svc.RegisterFeature(context.Background(), f)
	if err != featurestore.ErrFeatureAlreadyExists {
		t.Errorf("expected ErrFeatureAlreadyExists, got %v", err)
	}
}

func TestFeatureNotFound(t *testing.T) {
	svc := featurestore.NewService(featurestore.NewInMemoryRepository())
	_, err := svc.GetFeature(context.Background(), "nonexistent")
	if err != featurestore.ErrFeatureNotFound {
		t.Errorf("expected ErrFeatureNotFound, got %v", err)
	}
}

func TestSetFeatureValue(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	svc.RegisterFeature(context.Background(), &featurestore.Feature{
		Name: "price", ValueType: featurestore.TypeNumber, Entity: featurestore.EntityProduct,
	})

	err := svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{
		FeatureName: "price", EntityID: "prod-1", Value: 29.99,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := svc.GetFeatureValue(context.Background(), "price", "prod-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Value.(float64) != 29.99 {
		t.Errorf("expected 29.99, got %v", val.Value)
	}
}

func TestBatchGetFeatureValues(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "f1", Entity: featurestore.EntityUser})
	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "f2", Entity: featurestore.EntityUser})

	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "f1", EntityID: "u1", Value: "a"})
	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "f2", EntityID: "u1", Value: "b"})

	vals, err := svc.BatchGet(context.Background(), []string{"f1", "f2"}, "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 2 {
		t.Errorf("expected 2 values, got %d", len(vals))
	}
}

func TestGetFeaturesForEntity(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "age", Entity: featurestore.EntityUser})
	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "gender", Entity: featurestore.EntityUser})
	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "price", Entity: featurestore.EntityProduct})

	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "age", EntityID: "u1", Value: 25})
	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "gender", EntityID: "u1", Value: "M"})

	vals, err := svc.GetFeaturesForEntity(context.Background(), "u1", featurestore.EntityUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 2 {
		t.Errorf("expected 2 features for user, got %d", len(vals))
	}
}

func TestFeatureValueVersioning(t *testing.T) {
	repo := featurestore.NewInMemoryRepository()
	svc := featurestore.NewService(repo)

	svc.RegisterFeature(context.Background(), &featurestore.Feature{Name: "score", Entity: featurestore.EntityUser})
	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "score", EntityID: "u1", Value: 1.0})
	svc.SetFeatureValue(context.Background(), &featurestore.FeatureValue{FeatureName: "score", EntityID: "u1", Value: 2.0})

	val, _ := svc.GetFeatureValue(context.Background(), "score", "u1")
	if val.Version != 2 {
		t.Errorf("expected version 2, got %d", val.Version)
	}
}

func TestFeatureValueNotFound(t *testing.T) {
	svc := featurestore.NewService(featurestore.NewInMemoryRepository())
	_, err := svc.GetFeatureValue(context.Background(), "missing", "e1")
	if err != featurestore.ErrFeatureValueNotFound {
		t.Errorf("expected ErrFeatureValueNotFound, got %v", err)
	}
}
