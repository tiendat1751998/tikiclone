package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/global-infra/internal/multiregion"
)

func TestMultiRegionCreate(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	region := &multiregion.Region{
		Name:     "US East",
		Code:     "us-east-1",
		IsActive: true,
		Endpoints: map[string]string{
			"auth": "https://auth.us-east-1.example.com",
		},
	}
	created, err := svc.Create(context.Background(), region)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Code != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", created.Code)
	}
}

func TestMultiRegionGetActive(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	svc.Create(context.Background(), &multiregion.Region{Name: "US East", Code: "us-east-1", IsActive: true})
	svc.Create(context.Background(), &multiregion.Region{Name: "EU West", Code: "eu-west-1", IsActive: false})
	svc.Create(context.Background(), &multiregion.Region{Name: "AP SE", Code: "ap-southeast-1", IsActive: true})

	active, _ := svc.GetActiveRegions(context.Background())
	if len(active) != 2 {
		t.Errorf("expected 2 active regions, got %d", len(active))
	}
}

func TestMultiRegionEndpoints(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	svc.Create(context.Background(), &multiregion.Region{
		Name: "US East", Code: "us-east-1", IsActive: true,
		Endpoints: map[string]string{"auth": "https://auth.us-east-1.example.com"},
	})

	endpoints, err := svc.GetRegionEndpoints(context.Background(), "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoints["auth"] != "https://auth.us-east-1.example.com" {
		t.Errorf("expected auth endpoint, got %s", endpoints["auth"])
	}
}

func TestMultiRegionFailover(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	svc.Create(context.Background(), &multiregion.Region{
		Name: "US East", Code: "us-east-1", IsActive: true,
		Endpoints: map[string]string{"auth": "https://auth.us-east-1.example.com"},
	})
	svc.Create(context.Background(), &multiregion.Region{
		Name: "US West", Code: "us-west-2", IsActive: true,
		FailoverRegion: "us-east-1",
		Endpoints:      map[string]string{"auth": "https://auth.us-west-2.example.com"},
	})

	result, err := svc.GetFailoverStrategy(context.Background(), "us-west-2", "auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsFailover {
		t.Error("expected no failover since region is active")
	}
}

func TestMultiRegionFailoverToBackup(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	svc.Create(context.Background(), &multiregion.Region{
		Name: "Primary", Code: "primary-1", IsActive: false,
		FailoverRegion: "backup-1",
		Endpoints:      map[string]string{"api": "https://api.primary.example.com"},
	})
	svc.Create(context.Background(), &multiregion.Region{
		Name: "Backup", Code: "backup-1", IsActive: true,
		Endpoints: map[string]string{"api": "https://api.backup.example.com"},
	})

	result, err := svc.GetFailoverStrategy(context.Background(), "primary-1", "api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsFailover {
		t.Error("expected failover since region is inactive")
	}
	if result.Endpoint != "https://api.backup.example.com" {
		t.Errorf("expected backup endpoint, got %s", result.Endpoint)
	}
}

func TestMultiRegionList(t *testing.T) {
	repo := multiregion.NewInMemoryRepository()
	svc := multiregion.NewService(repo)

	svc.Create(context.Background(), &multiregion.Region{Name: "US East", Code: "us-east-1", IsActive: true})
	svc.Create(context.Background(), &multiregion.Region{Name: "EU West", Code: "eu-west-1", IsActive: true})

	regions, _ := svc.List(context.Background())
	if len(regions) != 2 {
		t.Errorf("expected 2 regions, got %d", len(regions))
	}
}
