package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/secrets"
)

func TestSecretCreate(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	secret := &secrets.Secret{
		Name:           "db-password",
		Value:          "supersecret123",
		ServiceName:    "auth-service",
		RotationPeriod: 30,
	}
	created, err := svc.Create(context.Background(), secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Name != "db-password" {
		t.Errorf("expected db-password, got %s", created.Name)
	}
	if created.Version != 1 {
		t.Errorf("expected version 1, got %d", created.Version)
	}
}

func TestSecretGet(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	svc.Create(context.Background(), &secrets.Secret{Name: "api-key", Value: "key-123", ServiceName: "gateway"})
	got, err := svc.GetByName(context.Background(), "api-key", "gateway")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected secret to exist")
	}
	if got.Name != "api-key" {
		t.Errorf("expected api-key, got %s", got.Name)
	}
}

func TestSecretRotate(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	created, _ := svc.Create(context.Background(), &secrets.Secret{Name: "rotate-me", Value: "original", ServiceName: "svc"})
	if created.Version != 1 {
		t.Errorf("expected version 1, got %d", created.Version)
	}
	rotated, err := svc.Rotate(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rotated.Version != 2 {
		t.Errorf("expected version 2 after rotation, got %d", rotated.Version)
	}
	if len(rotated.Value) == 0 {
		t.Error("expected rotated value to be non-empty")
	}
}

func TestSecretList(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	svc.Create(context.Background(), &secrets.Secret{Name: "s1", Value: "v1", ServiceName: "svc-a"})
	svc.Create(context.Background(), &secrets.Secret{Name: "s2", Value: "v2", ServiceName: "svc-a"})
	svc.Create(context.Background(), &secrets.Secret{Name: "s3", Value: "v3", ServiceName: "svc-b"})

	all, _ := svc.List(context.Background(), "")
	if len(all) != 3 {
		t.Errorf("expected 3 secrets, got %d", len(all))
	}

	filtered, _ := svc.List(context.Background(), "svc-a")
	if len(filtered) != 2 {
		t.Errorf("expected 2 secrets for svc-a, got %d", len(filtered))
	}
}

func TestSecretRotateNotFound(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	_, err := svc.Rotate(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent secret")
	}
}

func TestSecretValidateFields(t *testing.T) {
	repo := secrets.NewInMemoryRepository()
	svc := secrets.NewService(repo)

	_, err := svc.Create(context.Background(), &secrets.Secret{Name: "", Value: "v", ServiceName: "s"})
	if err == nil {
		t.Error("expected error for empty name")
	}
	_, err = svc.Create(context.Background(), &secrets.Secret{Name: "n", Value: "", ServiceName: "s"})
	if err == nil {
		t.Error("expected error for empty value")
	}
	_, err = svc.Create(context.Background(), &secrets.Secret{Name: "n", Value: "v", ServiceName: ""})
	if err == nil {
		t.Error("expected error for empty service_name")
	}
}
