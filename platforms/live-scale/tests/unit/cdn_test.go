package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/cdn"
)

func TestCDNPurgeCache(t *testing.T) {
	svc := cdn.NewService(cdn.NewInMemoryRepository(), nil)
	req := &cdn.CDNPurgeRequest{URLs: []string{"https://cdn.example.com/video/123"}, Reason: "stream ended"}
	if err := svc.PurgeCache(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ID == "" {
		t.Error("purge request ID should be set")
	}
}

func TestCDNPurgeInvalidRequest(t *testing.T) {
	svc := cdn.NewService(cdn.NewInMemoryRepository(), nil)
	req := &cdn.CDNPurgeRequest{}
	err := svc.PurgeCache(context.Background(), req)
	if err != cdn.ErrInvalidPurgeRequest {
		t.Errorf("expected ErrInvalidPurgeRequest, got %v", err)
	}
}

func TestCDNGetEndpointSameRegion(t *testing.T) {
	repo := cdn.NewInMemoryRepository()
	repo.SaveEndpoint(context.Background(), &cdn.CDNEndpoint{
		ID: "cdn-1", URL: "https://cdn-us-east.example.com", Region: "us-east",
		LatencyMs: 5, Capacity: 1000, CurrentLoad: 100, Status: "active",
	})
	repo.SaveEndpoint(context.Background(), &cdn.CDNEndpoint{
		ID: "cdn-2", URL: "https://cdn-us-west.example.com", Region: "us-west",
		LatencyMs: 10, Capacity: 1000, CurrentLoad: 200, Status: "active",
	})
	svc := cdn.NewService(repo, nil)
	ep, err := svc.GetCDNEndpoint(context.Background(), "us-east")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep.Region != "us-east" {
		t.Errorf("expected us-east endpoint, got %s", ep.Region)
	}
}

func TestCDNGetEndpointFallback(t *testing.T) {
	repo := cdn.NewInMemoryRepository()
	repo.SaveEndpoint(context.Background(), &cdn.CDNEndpoint{
		ID: "cdn-eu", URL: "https://cdn-eu.example.com", Region: "eu-west",
		LatencyMs: 20, Capacity: 1000, CurrentLoad: 500, Status: "active",
	})
	svc := cdn.NewService(repo, nil)
	ep, err := svc.GetCDNEndpoint(context.Background(), "ap-southeast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ep == nil {
		t.Fatal("expected an endpoint")
	}
}

func TestCDNGetEndpointNoAvailable(t *testing.T) {
	svc := cdn.NewService(cdn.NewInMemoryRepository(), nil)
	_, err := svc.GetCDNEndpoint(context.Background(), "us-east")
	if err != cdn.ErrNoEndpointsAvailable {
		t.Errorf("expected ErrNoEndpointsAvailable, got %v", err)
	}
}

func TestCDNInvalidateEdge(t *testing.T) {
	repo := cdn.NewInMemoryRepository()
	repo.SaveEndpoint(context.Background(), &cdn.CDNEndpoint{
		ID: "cdn-test", URL: "https://cdn.example.com", Region: "us-east",
		Capacity: 1000, CurrentLoad: 100, Status: "active",
	})
	svc := cdn.NewService(repo, nil)
	if err := svc.InvalidateEdge(context.Background(), "https://cdn.example.com/video/123", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCDNPrefetchContent(t *testing.T) {
	repo := cdn.NewInMemoryRepository()
	repo.SaveEndpoint(context.Background(), &cdn.CDNEndpoint{
		ID: "cdn-prefetch", URL: "https://cdn.example.com", Region: "ap-southeast",
		Capacity: 1000, CurrentLoad: 100, Status: "active",
	})
	svc := cdn.NewService(repo, nil)
	if err := svc.PrefetchContent(context.Background(), []string{"https://cdn.example.com/video/popular"}, "ap-southeast"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
