package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/shipments"
)

type memShipmentRepo struct {
	shipments map[string]*shipments.Shipment
}

func newMemShipmentRepo() *memShipmentRepo {
	return &memShipmentRepo{shipments: make(map[string]*shipments.Shipment)}
}

func (r *memShipmentRepo) Create(_ context.Context, s *shipments.Shipment) error {
	r.shipments[s.ID] = s
	return nil
}
func (r *memShipmentRepo) GetByID(_ context.Context, id string) (*shipments.Shipment, error) {
	s, ok := r.shipments[id]
	if !ok {
		return nil, shipments.ErrShipmentNotFound
	}
	return s, nil
}
func (r *memShipmentRepo) Update(_ context.Context, s *shipments.Shipment) error {
	r.shipments[s.ID] = s
	return nil
}
func (r *memShipmentRepo) List(_ context.Context, filter shipments.ShipmentFilter) ([]*shipments.Shipment, int64, error) {
	var result []*shipments.Shipment
	for _, s := range r.shipments {
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		if filter.CourierID != "" && s.CourierID != filter.CourierID {
			continue
		}
		result = append(result, s)
	}
	return result, int64(len(result)), nil
}
func (r *memShipmentRepo) Delete(_ context.Context, id string) error {
	delete(r.shipments, id)
	return nil
}
func (r *memShipmentRepo) TransitionStatus(_ context.Context, txnID string, from, to shipments.ShipmentStatus, replayID string) error {
	s, ok := r.shipments[txnID]
	if !ok {
		return shipments.ErrShipmentNotFound
	}
	if s.Status != from {
		return shipments.ErrInvalidStatusTransition
	}
	s.Status = to
	return nil
}
func (r *memShipmentRepo) GetByOrderID(_ context.Context, orderID string) ([]*shipments.Shipment, error) {
	var result []*shipments.Shipment
	for _, s := range r.shipments {
		if s.OrderID == orderID {
			result = append(result, s)
		}
	}
	return result, nil
}

func TestShipmentCreate(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	sh := &shipments.Shipment{
		ID:      "ship-001",
		OrderID: "order-001",
		Status:  shipments.StatusPending,
	}
	if err := svc.Create(context.Background(), sh); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sh.Status != shipments.StatusPending {
		t.Errorf("expected pending, got %v", sh.Status)
	}
}

func TestShipmentStatusTransition(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	sh := &shipments.Shipment{ID: "ship-002", OrderID: "order-002"}
	if err := svc.Create(context.Background(), sh); err != nil {
		t.Fatal(err)
	}
	if err := svc.TransitionStatus(context.Background(), "ship-002", shipments.StatusConfirmed, "", ""); err != nil {
		t.Fatalf("transition failed: %v", err)
	}
	s, _ := svc.GetByID(context.Background(), "ship-002")
	if s.Status != shipments.StatusConfirmed {
		t.Errorf("expected confirmed, got %v", s.Status)
	}
}

func TestShipmentInvalidTransition(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	sh := &shipments.Shipment{ID: "ship-003", OrderID: "order-003"}
	if err := svc.Create(context.Background(), sh); err != nil {
		t.Fatal(err)
	}
	err := svc.TransitionStatus(context.Background(), "ship-003", shipments.StatusDelivered, "", "")
	if err != shipments.ErrInvalidStatusTransition {
		t.Errorf("expected invalid transition error, got %v", err)
	}
}

func TestShipmentNotFound(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent shipment")
	}
}

func TestShipmentValidTransitions(t *testing.T) {
	tests := []struct {
		from shipments.ShipmentStatus
		to   shipments.ShipmentStatus
		ok   bool
	}{
		{shipments.StatusPending, shipments.StatusConfirmed, true},
		{shipments.StatusPending, shipments.StatusCancelled, true},
		{shipments.StatusConfirmed, shipments.StatusAwaitingPickup, true},
		{shipments.StatusPickedUp, shipments.StatusInTransit, true},
		{shipments.StatusInTransit, shipments.StatusOutForDelivery, true},
		{shipments.StatusOutForDelivery, shipments.StatusDelivered, true},
		{shipments.StatusPending, shipments.StatusDelivered, false},
		{shipments.StatusDelivered, shipments.StatusPending, false},
		{shipments.StatusCancelled, shipments.StatusPending, false},
	}
	for _, tc := range tests {
		got := shipments.IsValidTransition(tc.from, tc.to)
		if got != tc.ok {
			t.Errorf("IsValidTransition(%v -> %v) = %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestShipmentCreatedAt(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	sh := &shipments.Shipment{ID: "ship-t1", OrderID: "order-t1"}
	if err := svc.Create(context.Background(), sh); err != nil {
		t.Fatal(err)
	}
	if sh.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}
	if sh.UpdatedAt.Before(sh.CreatedAt) {
		t.Error("updated_at should not be before created_at")
	}
}

func TestShipmentVersion(t *testing.T) {
	svc := shipments.NewService(newMemShipmentRepo(), nil)
	sh := &shipments.Shipment{ID: "ship-v1", OrderID: "order-v1"}
	if err := svc.Create(context.Background(), sh); err != nil {
		t.Fatal(err)
	}
	if sh.Version != 1 {
		t.Errorf("expected version 1, got %d", sh.Version)
	}
}
