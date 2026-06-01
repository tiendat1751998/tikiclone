package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/registry"
)

func TestServiceRegister(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, registry.NewHealthChecker())

	inst := &registry.ServiceInstance{
		ID:      "inst-1",
		Name:    "auth-service",
		Version: "1.0.0",
		Address: "10.0.0.1",
		Port:    8080,
		Region:  "us-east-1",
	}
	created, err := svc.Register(context.Background(), inst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "inst-1" {
		t.Errorf("expected inst-1, got %s", created.ID)
	}
}

func TestServiceDeregister(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, registry.NewHealthChecker())

	svc.Register(context.Background(), &registry.ServiceInstance{ID: "inst-1", Name: "svc", Address: "addr", Port: 8080})
	err := svc.Deregister(context.Background(), "inst-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.Get(context.Background(), "inst-1")
	if got != nil {
		t.Error("expected instance to be removed")
	}
}

func TestServiceHeartbeat(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, nil)

	svc.Register(context.Background(), &registry.ServiceInstance{ID: "inst-1", Name: "svc", Address: "addr", Port: 8080, HealthEndpoint: ""})
	err := svc.Heartbeat(context.Background(), "inst-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceDiscover(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, nil)

	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i1", Name: "svc-a", Address: "addr1", Port: 8080, Region: "us-east-1"})
	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i2", Name: "svc-a", Address: "addr2", Port: 8081, Region: "us-east-1"})
	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i3", Name: "svc-b", Address: "addr3", Port: 8082, Region: "eu-west-1"})

	instances, err := svc.Discover(context.Background(), "svc-a", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestServiceDiscoverFiltersDownInstances(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, nil)

	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i1", Name: "svc-a", Address: "addr1", Port: 8080, Status: registry.StatusUp})
	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i2", Name: "svc-a", Address: "addr2", Port: 8081, Status: registry.StatusDown})

	instances, _ := svc.Discover(context.Background(), "svc-a", "")
	if len(instances) != 1 {
		t.Errorf("expected 1 up instance, got %d", len(instances))
	}
}

func TestServiceList(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, nil)

	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i1", Name: "svc-a", Address: "a", Port: 1})
	svc.Register(context.Background(), &registry.ServiceInstance{ID: "i2", Name: "svc-b", Address: "b", Port: 2})

	all, _ := svc.List(context.Background())
	if len(all) != 2 {
		t.Errorf("expected 2 instances, got %d", len(all))
	}
}

func TestServiceValidateRegistration(t *testing.T) {
	repo := registry.NewInMemoryRepository()
	svc := registry.NewService(repo, nil)

	_, err := svc.Register(context.Background(), &registry.ServiceInstance{ID: "", Name: "", Address: "", Port: 0})
	if err == nil {
		t.Error("expected validation error")
	}
}
