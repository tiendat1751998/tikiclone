package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search-indexing/internal/monitoring"
)

func TestReportMetrics(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	metric := &monitoring.IndexMetric{
		IndexName:        "products",
		ShardCount:       5,
		DocCount:         10000,
		SizeBytes:        524288000,
		IndexingRate:     150.5,
		SearchLatencyP99: 45.2,
		Health:           monitoring.HealthGreen,
	}

	if err := svc.ReportMetrics(ctx, metric); err != nil {
		t.Fatalf("ReportMetrics failed: %v", err)
	}
}

func TestReportMetricsDefaultsHealth(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	metric := &monitoring.IndexMetric{
		IndexName: "orders",
		ShardCount: 3,
		DocCount:   5000,
	}

	if err := svc.ReportMetrics(ctx, metric); err != nil {
		t.Fatalf("ReportMetrics failed: %v", err)
	}

	retrieved, _ := svc.GetIndexMetric(ctx, "orders")
	if retrieved.Health != monitoring.HealthGreen {
		t.Errorf("expected default health green, got %s", retrieved.Health)
	}
}

func TestGetIndexMetrics(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "products", ShardCount: 5, DocCount: 10000})
	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "orders", ShardCount: 3, DocCount: 5000})
	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "users", ShardCount: 2, DocCount: 2000})

	metrics, err := svc.GetIndexMetrics(ctx)
	if err != nil {
		t.Fatalf("GetIndexMetrics failed: %v", err)
	}
	if len(metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(metrics))
	}
}

func TestGetClusterHealth(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "products", ShardCount: 5, DocCount: 10000, Health: monitoring.HealthGreen})
	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "orders", ShardCount: 3, DocCount: 5000, Health: monitoring.HealthYellow})
	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "users", ShardCount: 2, DocCount: 2000, Health: monitoring.HealthRed})

	health, err := svc.GetClusterHealth(ctx)
	if err != nil {
		t.Fatalf("GetClusterHealth failed: %v", err)
	}

	if health.ActiveShards != 10 {
		t.Errorf("expected 10 active shards, got %d", health.ActiveShards)
	}
	if health.RelocatingShards != 1 {
		t.Errorf("expected 1 relocating shard, got %d", health.RelocatingShards)
	}
	if health.UnassignedShards != 1 {
		t.Errorf("expected 1 unassigned shard, got %d", health.UnassignedShards)
	}
}

func TestGetClusterHealthEmpty(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	health, err := svc.GetClusterHealth(ctx)
	if err != nil {
		t.Fatalf("GetClusterHealth failed: %v", err)
	}
	if health.NodesCount != 1 {
		t.Errorf("expected default 1 node, got %d", health.NodesCount)
	}
	if health.ActiveShards != 0 {
		t.Errorf("expected 0 shards, got %d", health.ActiveShards)
	}
}

func TestGetIndexMetric(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	svc.ReportMetrics(ctx, &monitoring.IndexMetric{IndexName: "products", ShardCount: 5, DocCount: 10000})

	metric, err := svc.GetIndexMetric(ctx, "products")
	if err != nil {
		t.Fatalf("GetIndexMetric failed: %v", err)
	}
	if metric.DocCount != 10000 {
		t.Errorf("expected 10000 docs, got %d", metric.DocCount)
	}
}

func TestIndexMetricNotFound(t *testing.T) {
	repo := monitoring.NewInMemoryRepository()
	svc := monitoring.NewService(repo)
	ctx := context.Background()

	_, err := svc.GetIndexMetric(ctx, "nonexistent")
	if err != monitoring.ErrIndexMetricNotFound {
		t.Errorf("expected ErrIndexMetricNotFound, got %v", err)
	}
}
