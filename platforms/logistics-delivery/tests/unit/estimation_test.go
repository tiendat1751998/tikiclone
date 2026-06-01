package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/estimations"
)

type memEstimationRepo struct {
	estimations map[string]*estimations.Estimation
}

func newMemEstimationRepo() *memEstimationRepo {
	return &memEstimationRepo{estimations: make(map[string]*estimations.Estimation)}
}

func (r *memEstimationRepo) Create(_ context.Context, e *estimations.Estimation) error {
	r.estimations[e.ID] = e
	return nil
}
func (r *memEstimationRepo) GetByShipment(_ context.Context, shipmentID string) (*estimations.Estimation, error) {
	for _, e := range r.estimations {
		if e.ShipmentID == shipmentID {
			return e, nil
		}
	}
	return nil, estimations.ErrEstimationNotFound
}
func (r *memEstimationRepo) GetLatestByShipment(ctx context.Context, shipmentID string) (*estimations.Estimation, error) {
	return r.GetByShipment(ctx, shipmentID)
}

func TestCalculateEstimation(t *testing.T) {
	svc := estimations.NewService(newMemEstimationRepo(), nil)
	req := &estimations.EstimationRequest{
		OriginLat:     21.0285,
		OriginLng:     105.8542,
		DestLat:       21.0320,
		DestLng:       105.8600,
		PackageWeight: 5.0,
		DistanceKm:    10.0,
		TrafficFactor: 0.3,
	}
	est, err := svc.Calculate(context.Background(), req, "ship-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if est.DistanceKm != 10.0 {
		t.Errorf("expected 10.0 km, got %f", est.DistanceKm)
	}
	if est.TotalDurationMin <= 0 {
		t.Errorf("expected positive duration, got %d", est.TotalDurationMin)
	}
	if est.Confidence < 0.3 || est.Confidence > 1.0 {
		t.Errorf("confidence out of range: %f", est.Confidence)
	}
	if est.ETA.IsZero() {
		t.Error("ETA should be set")
	}
	if est.ExpiresAt.Before(est.CalculatedAt) {
		t.Error("expiry should be after calculation time")
	}
}

func TestEstimationTrafficImpact(t *testing.T) {
	svc := estimations.NewService(newMemEstimationRepo(), nil)
	reqLow := &estimations.EstimationRequest{DistanceKm: 10.0, TrafficFactor: 0.1, PackageWeight: 1.0}
	reqHigh := &estimations.EstimationRequest{DistanceKm: 10.0, TrafficFactor: 0.8, PackageWeight: 1.0}
	low, _ := svc.Calculate(context.Background(), reqLow, "ship-t1")
	high, _ := svc.Calculate(context.Background(), reqHigh, "ship-t2")
	if low.TrafficDelayMin >= high.TrafficDelayMin {
		t.Errorf("high traffic should have >= delay: low=%d high=%d", low.TrafficDelayMin, high.TrafficDelayMin)
	}
	if low.Confidence <= high.Confidence {
		t.Errorf("low traffic should have higher confidence: low=%.2f high=%.2f", low.Confidence, high.Confidence)
	}
}
