package stream_health

import (
	"context"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/events"
)

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) ReportHealth(ctx context.Context, streamID string, metric HealthMetric) (*StreamHealth, error) {
	health, err := s.repo.GetStreamHealth(ctx, streamID)
	if err != nil {
		health = &StreamHealth{
			StreamID:   streamID,
			Status:     StreamHealthy,
			Threshold:  AlertThreshold{MaxLatencyMs: 500, MaxPacketLoss: 0.05, MinBitrate: 500000, MinFrameRate: 24},
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
	}

	metric.RecordedAt = time.Now().UTC()
	health.Metrics = append(health.Metrics, metric)
	if len(health.Metrics) > 100 {
		health.Metrics = health.Metrics[len(health.Metrics)-100:]
	}

	previousStatus := health.Status
	health.Status = s.evaluateHealth(metric, health.Threshold)
	health.LastChecked = time.Now().UTC()
	health.UpdatedAt = time.Now().UTC()

	if err := s.repo.SaveStreamHealth(ctx, health); err != nil {
		return nil, fmt.Errorf("report health: %w", err)
	}

	if health.Status != previousStatus && s.producer != nil {
		eventType := events.StreamDegraded
		if health.Status == StreamDown {
			eventType = events.StreamDown
		}
		s.producer.Publish(ctx, events.Event{
			Type:      eventType,
			Source:    "live-scale.streamhealth",
			Payload:   health,
			Timestamp: time.Now().UTC(),
		})
	}

	return health, nil
}

func (s *Service) CheckHealth(ctx context.Context, streamID string) (*StreamHealth, error) {
	health, err := s.repo.GetStreamHealth(ctx, streamID)
	if err != nil {
		return nil, err
	}

	if len(health.Metrics) == 0 {
		health.Status = StreamDown
		s.repo.SaveStreamHealth(ctx, health)
		return health, nil
	}

	latest := health.Metrics[len(health.Metrics)-1]
	if time.Since(latest.RecordedAt) > 30*time.Second {
		health.Status = StreamDown
		s.repo.SaveStreamHealth(ctx, health)
		return health, nil
	}

	health.Status = s.evaluateHealth(latest, health.Threshold)
	health.LastChecked = time.Now().UTC()
	s.repo.SaveStreamHealth(ctx, health)
	return health, nil
}

func (s *Service) GetStreamHealth(ctx context.Context, streamID string) (*StreamHealth, error) {
	return s.repo.GetStreamHealth(ctx, streamID)
}

func (s *Service) GenerateAlert(ctx context.Context, streamID string) (string, error) {
	health, err := s.repo.GetStreamHealth(ctx, streamID)
	if err != nil {
		return "", err
	}

	if health.Status == StreamHealthy {
		return "", nil
	}

	alert := fmt.Sprintf("Stream %s is %s", streamID, health.Status)
	if len(health.Metrics) > 0 {
		latest := health.Metrics[len(health.Metrics)-1]
		alert += fmt.Sprintf(" | bitrate=%d latency=%dms packet_loss=%.2f frame_rate=%.1f",
			latest.Bitrate, latest.LatencyMs, latest.PacketLoss, latest.FrameRate)
	}

	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.AlertGenerated,
			Source:    "live-scale.streamhealth",
			Payload:   map[string]interface{}{"stream_id": streamID, "alert": alert, "status": health.Status},
			Timestamp: time.Now().UTC(),
		})
	}

	return alert, nil
}

func (s *Service) evaluateHealth(metric HealthMetric, threshold AlertThreshold) StreamStatus {
	if metric.Bitrate < threshold.MinBitrate/2 || metric.FrameRate < threshold.MinFrameRate/2 {
		return StreamDown
	}
	if metric.LatencyMs > threshold.MaxLatencyMs || metric.PacketLoss > threshold.MaxPacketLoss ||
		metric.Bitrate < threshold.MinBitrate || metric.FrameRate < threshold.MinFrameRate {
		return StreamDegraded
	}
	return StreamHealthy
}
