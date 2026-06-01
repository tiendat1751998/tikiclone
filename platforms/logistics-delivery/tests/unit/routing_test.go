package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/routing"
)

type memRoutingRepo struct{}

func newMemRoutingRepo() *memRoutingRepo { return &memRoutingRepo{} }
func (r *memRoutingRepo) CreateRoute(_ context.Context, rt *routing.Route) error { return nil }
func (r *memRoutingRepo) GetRoutesByShipment(_ context.Context, shipmentID string) ([]*routing.Route, error) { return nil, nil }
func (r *memRoutingRepo) GetActiveRoute(_ context.Context, shipmentID string) (*routing.Route, error) { return nil, routing.ErrRouteNotFound }
func (r *memRoutingRepo) UpdateRoute(_ context.Context, rt *routing.Route) error { return nil }
func (r *memRoutingRepo) GetZones(_ context.Context) ([]*routing.Zone, error) { return nil, nil }
func (r *memRoutingRepo) GetZoneByCity(_ context.Context, city string) (*routing.Zone, error) {
	return &routing.Zone{ID: city + "-zone", Name: city + " Zone", City: city, IsActive: true}, nil
}

func TestAssignWarehouse(t *testing.T) {
	svc := routing.NewService(newMemRoutingRepo())
	assignment, err := svc.AssignWarehouse(context.Background(), "ship-001", "Hanoi", "HN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.ShipmentID != "ship-001" {
		t.Errorf("expected ship-001, got %s", assignment.ShipmentID)
	}
	if assignment.WarehouseID == "" {
		t.Error("expected warehouse id to be assigned")
	}
}

func TestOptimizeWaypoints(t *testing.T) {
	svc := routing.NewService(newMemRoutingRepo())
	waypoints := []routing.Waypoint{
		{Sequence: 0, Latitude: 21.0285, Longitude: 105.8542, Name: "Hub A"},
		{Sequence: 1, Latitude: 21.0320, Longitude: 105.8600, Name: "Drop 1"},
		{Sequence: 2, Latitude: 21.0220, Longitude: 105.8500, Name: "Drop 2"},
	}
	optimized, err := svc.OptimizeWaypoints(context.Background(), waypoints)
	if err != nil {
		t.Fatal(err)
	}
	if len(optimized) != 3 {
		t.Errorf("expected 3 waypoints, got %d", len(optimized))
	}
	if optimized[0].Name != "Hub A" {
		t.Errorf("first waypoint should be Hub A, got %s", optimized[0].Name)
	}
}

func TestOptimizeWaypointsSmall(t *testing.T) {
	svc := routing.NewService(newMemRoutingRepo())
	wps := []routing.Waypoint{
		{Sequence: 0, Latitude: 21.0285, Longitude: 105.8542},
		{Sequence: 1, Latitude: 21.0320, Longitude: 105.8600},
	}
	result, err := svc.OptimizeWaypoints(context.Background(), wps)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 waypoints, got %d", len(result))
	}
}
