package estimations

import (
	"context"
	"math"
	"time"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/events"
)

type Repository interface {
	Create(ctx context.Context, e *Estimation) error
	GetByShipment(ctx context.Context, shipmentID string) (*Estimation, error)
	GetLatestByShipment(ctx context.Context, shipmentID string) (*Estimation, error)
}

type Service struct {
	repo     Repository
	producer events.Producer
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) Calculate(ctx context.Context, req *EstimationRequest, shipmentID string) (*Estimation, error) {
	baseDuration := int(req.DistanceKm / 40.0 * 60.0)
	if baseDuration < 1 {
		baseDuration = 1
	}
	trafficDelay := int(float64(baseDuration) * req.TrafficFactor * 0.3)
	if trafficDelay < 0 {
		trafficDelay = 0
	}
	weightFactor := req.PackageWeight / 50.0
	if weightFactor > 1.0 {
		weightFactor = 1.0
	}
	weatherDelay := int(weightFactor * 5.0)
	totalDuration := baseDuration + trafficDelay + weatherDelay
	confidence := 1.0 - (req.TrafficFactor * 0.4)
	if confidence < 0.3 {
		confidence = 0.3
	}
	now := time.Now().UTC()
	est := &Estimation{
		ID:               generateID(shipmentID),
		ShipmentID:       shipmentID,
		DistanceKm:       req.DistanceKm,
		BaseDurationMin:  baseDuration,
		TrafficDelayMin:  trafficDelay,
		WeatherDelayMin:  weatherDelay,
		TotalDurationMin: totalDuration,
		ETA:              now.Add(time.Duration(totalDuration) * time.Minute),
		Confidence:       math.Round(confidence*100) / 100,
		RouteHash:        hashRoute(req.OriginLat, req.OriginLng, req.DestLat, req.DestLng),
		CalculatedAt:     now,
		ExpiresAt:        now.Add(15 * time.Minute),
	}
	if err := s.repo.Create(ctx, est); err != nil {
		return nil, err
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.EstimationCalculated,
			Source:    "logistics.estimations",
			Payload:   est,
			Timestamp: now,
		})
	}
	return est, nil
}

func (s *Service) GetByShipment(ctx context.Context, shipmentID string) (*Estimation, error) {
	return s.repo.GetByShipment(ctx, shipmentID)
}

func generateID(shipmentID string) string {
	return "est-" + shipmentID + "-" + time.Now().UTC().Format("20060102150405")
}

func hashRoute(originLat, originLng, destLat, destLng float64) string {
	return "r" + formatCoord(originLat) + formatCoord(originLng) + formatCoord(destLat) + formatCoord(destLng)
}

func formatCoord(f float64) string {
	return string(rune('A' + int(math.Abs(f)*10)))
}
