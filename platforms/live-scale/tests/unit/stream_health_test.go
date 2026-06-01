package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/stream_health"
)

func TestStreamHealthReport(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 100, PacketLoss: 0.01, Viewers: 50}
	health, err := svc.ReportHealth(context.Background(), "stream-001", metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != stream_health.StreamHealthy {
		t.Errorf("expected healthy, got %v", health.Status)
	}
	if health.StreamID != "stream-001" {
		t.Errorf("expected stream-001, got %s", health.StreamID)
	}
}

func TestStreamHealthDegradedLatency(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 600, PacketLoss: 0.01, Viewers: 50}
	health, err := svc.ReportHealth(context.Background(), "stream-002", metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != stream_health.StreamDegraded {
		t.Errorf("expected degraded due to high latency, got %v", health.Status)
	}
}

func TestStreamHealthDegradedPacketLoss(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 100, PacketLoss: 0.10, Viewers: 50}
	health, err := svc.ReportHealth(context.Background(), "stream-003", metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != stream_health.StreamDegraded {
		t.Errorf("expected degraded due to packet loss, got %v", health.Status)
	}
}

func TestStreamHealthDownLowBitrate(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 50000, FrameRate: 10, LatencyMs: 100, PacketLoss: 0.01, Viewers: 50}
	health, err := svc.ReportHealth(context.Background(), "stream-004", metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != stream_health.StreamDown {
		t.Errorf("expected down due to low bitrate, got %v", health.Status)
	}
}

func TestStreamHealthCheck(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 100, PacketLoss: 0.01, Viewers: 50}
	svc.ReportHealth(context.Background(), "stream-005", metric)
	health, err := svc.CheckHealth(context.Background(), "stream-005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != stream_health.StreamHealthy {
		t.Errorf("expected healthy, got %v", health.Status)
	}
}

func TestStreamHealthNotFound(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	_, err := svc.GetStreamHealth(context.Background(), "nonexistent")
	if err != stream_health.ErrStreamNotFound {
		t.Errorf("expected ErrStreamNotFound, got %v", err)
	}
}

func TestStreamHealthGenerateAlert(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 50000, FrameRate: 10, LatencyMs: 100, PacketLoss: 0.01, Viewers: 50}
	svc.ReportHealth(context.Background(), "stream-006", metric)
	alert, err := svc.GenerateAlert(context.Background(), "stream-006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert == "" {
		t.Error("expected non-empty alert for degraded stream")
	}
}

func TestStreamHealthNoAlertForHealthy(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 100, PacketLoss: 0.01, Viewers: 50}
	svc.ReportHealth(context.Background(), "stream-007", metric)
	alert, err := svc.GenerateAlert(context.Background(), "stream-007")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert != "" {
		t.Errorf("expected empty alert for healthy stream, got %s", alert)
	}
}

func TestStreamHealthMetricsHistory(t *testing.T) {
	svc := stream_health.NewService(stream_health.NewInMemoryRepository(), nil)
	for i := 0; i < 110; i++ {
		metric := stream_health.HealthMetric{Bitrate: 3000000, FrameRate: 30, LatencyMs: 100, PacketLoss: 0.01, Viewers: i}
		svc.ReportHealth(context.Background(), "stream-008", metric)
	}
	health, _ := svc.GetStreamHealth(context.Background(), "stream-008")
	if len(health.Metrics) > 100 {
		t.Errorf("expected max 100 metrics, got %d", len(health.Metrics))
	}
}
