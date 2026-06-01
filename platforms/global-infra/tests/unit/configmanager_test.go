package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/configmanager"
)

func TestConfigCreateDuplicateIncrementsVersion(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v1", Environment: configmanager.EnvDev, ServiceName: "svc"})
	entry := &configmanager.ConfigEntry{Key: "k1", Value: "v2", Environment: configmanager.EnvDev, ServiceName: "svc"}
	created, _ := svc.Create(context.Background(), entry)
	if created.Version != 2 {
		t.Errorf("expected version 2, got %d", created.Version)
	}
}

func TestConfigGet(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v1", Environment: configmanager.EnvStaging, ServiceName: "svc"})
	got, err := svc.Get(context.Background(), "k1", configmanager.EnvStaging, "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Value != "v1" {
		t.Errorf("expected v1, got %v", got)
	}
}

func TestConfigGetVersion(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v1", Environment: configmanager.EnvDev, ServiceName: "svc"})
	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v2", Environment: configmanager.EnvDev, ServiceName: "svc"})

	v1, _ := svc.GetVersion(context.Background(), "k1", configmanager.EnvDev, "svc", 1)
	if v1 == nil || v1.Value != "v1" {
		t.Errorf("expected v1 for version 1, got %v", v1)
	}
	v2, _ := svc.GetVersion(context.Background(), "k1", configmanager.EnvDev, "svc", 2)
	if v2 == nil || v2.Value != "v2" {
		t.Errorf("expected v2 for version 2, got %v", v2)
	}
}

func TestConfigListByServiceAndEnv(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "a", Value: "1", Environment: configmanager.EnvDev, ServiceName: "svc1"})
	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "b", Value: "2", Environment: configmanager.EnvProd, ServiceName: "svc1"})
	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "c", Value: "3", Environment: configmanager.EnvDev, ServiceName: "svc2"})

	configs, _ := svc.List(context.Background(), "svc1", configmanager.EnvDev)
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
}

func TestConfigRequiresFields(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	_, err := svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "", Value: "v", ServiceName: "s"})
	if err == nil {
		t.Error("expected error for empty key")
	}
	_, err = svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k", Value: "", ServiceName: "s"})
	if err == nil {
		t.Error("expected error for empty value")
	}
	_, err = svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k", Value: "v", ServiceName: ""})
	if err == nil {
		t.Error("expected error for empty service_name")
	}
}

func TestConfigVersioning(t *testing.T) {
	repo := configmanager.NewInMemoryRepository()
	svc := configmanager.NewService(repo, configmanager.NewNoOpPublisher(), logger)

	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v1", Environment: configmanager.EnvProd, ServiceName: "payments"})
	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v2", Environment: configmanager.EnvProd, ServiceName: "payments"})
	svc.Create(context.Background(), &configmanager.ConfigEntry{Key: "k1", Value: "v3", Environment: configmanager.EnvProd, ServiceName: "payments"})

	versions, _ := svc.ListVersions(context.Background(), "k1", configmanager.EnvProd, "payments")
	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}
}
