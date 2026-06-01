package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/region"
)

func TestRegionGetNearestSameRegion(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-east", Name: "US East", Status: region.RegionActive, Latency: 5})
	repo.SaveRegion(context.Background(), &region.Region{Code: "eu-west", Name: "EU West", Status: region.RegionActive, Latency: 80})
	svc := region.NewService(repo, nil)
	r, err := svc.GetNearestRegion(context.Background(), "us-east")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Code != "us-east" {
		t.Errorf("expected us-east, got %s", r.Code)
	}
}

func TestRegionGetNearestCrossRegion(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-east", Name: "US East", Status: region.RegionActive, Latency: 5})
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-west", Name: "US West", Status: region.RegionActive, Latency: 10})
	repo.SaveRegion(context.Background(), &region.Region{Code: "eu-west", Name: "EU West", Status: region.RegionActive, Latency: 80})
	repo.SaveLatency(context.Background(), &region.LatencyMap{FromRegion: "sa-east", ToRegion: "us-east", LatencyMs: 60})
	repo.SaveLatency(context.Background(), &region.LatencyMap{FromRegion: "sa-east", ToRegion: "us-west", LatencyMs: 120})
	repo.SaveLatency(context.Background(), &region.LatencyMap{FromRegion: "sa-east", ToRegion: "eu-west", LatencyMs: 150})
	svc := region.NewService(repo, nil)
	r, err := svc.GetNearestRegion(context.Background(), "sa-east")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Code != "us-east" {
		t.Errorf("expected us-east (lowest latency), got %s", r.Code)
	}
}

func TestRegionRouteToRegion(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "ap-southeast", Name: "Singapore", Status: region.RegionActive, Latency: 5})
	svc := region.NewService(repo, nil)
	r, err := svc.RouteToRegion(context.Background(), "ap-southeast")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Code != "ap-southeast" {
		t.Errorf("expected ap-southeast, got %s", r.Code)
	}
}

func TestRegionFailover(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-east", Name: "US East", Status: region.RegionActive, Latency: 5})
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-west", Name: "US West", Status: region.RegionActive, Latency: 50})
	repo.SaveRegion(context.Background(), &region.Region{Code: "eu-west", Name: "EU West", Status: region.RegionActive, Latency: 100})
	svc := region.NewService(repo, nil)
	r, err := svc.FailoverRegion(context.Background(), "us-east")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Code == "us-east" {
		t.Errorf("expected failover to different region, got us-east")
	}
	if r.Status != region.RegionActive {
		t.Errorf("expected failover region to be active, got %v", r.Status)
	}
}

func TestRegionFailoverNoAvailable(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "only-region", Name: "Only", Status: region.RegionActive, Latency: 5})
	svc := region.NewService(repo, nil)
	_, err := svc.FailoverRegion(context.Background(), "only-region")
	if err != region.ErrNoRegionAvailable {
		t.Errorf("expected ErrNoRegionAvailable, got %v", err)
	}
}

func TestRegionLatency(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "us-east", Name: "US East", Status: region.RegionActive, Latency: 5})
	repo.SaveLatency(context.Background(), &region.LatencyMap{FromRegion: "us-east", ToRegion: "eu-west", LatencyMs: 70})
	svc := region.NewService(repo, nil)
	latency, err := svc.GetRegionLatency(context.Background(), "us-east", "eu-west")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latency != 70 {
		t.Errorf("expected 70ms latency, got %d", latency)
	}
}

func TestRegionNoRegions(t *testing.T) {
	svc := region.NewService(region.NewInMemoryRepository(), nil)
	_, err := svc.GetNearestRegion(context.Background(), "us-east")
	if err != region.ErrNoRegionAvailable {
		t.Errorf("expected ErrNoRegionAvailable, got %v", err)
	}
}

func TestRegionSetRegionStatus(t *testing.T) {
	repo := region.NewInMemoryRepository()
	repo.SaveRegion(context.Background(), &region.Region{Code: "test-region", Name: "Test", Status: region.RegionActive, Latency: 10})
	svc := region.NewService(repo, nil)
	if err := svc.SetRegionStatus(context.Background(), "test-region", region.RegionDegraded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, _ := repo.GetRegion(context.Background(), "test-region")
	if r.Status != region.RegionDegraded {
		t.Errorf("expected degraded, got %v", r.Status)
	}
}

func TestRegionRecordLatencyInvalid(t *testing.T) {
	svc := region.NewService(region.NewInMemoryRepository(), nil)
	err := svc.RecordLatency(context.Background(), "us-east", "eu-west", -1)
	if err != region.ErrInvalidLatencyData {
		t.Errorf("expected ErrInvalidLatencyData, got %v", err)
	}
}
