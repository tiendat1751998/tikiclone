package unit

import (
	"context"
	"sync"
	"testing"

	"github.com/tikiclone/tiki/platforms/logistics-delivery/internal/dispatch"
)

type memDispatchRepo struct {
	mu         sync.Mutex
	dispatches map[string]*dispatch.Dispatch
}

func newMemDispatchRepo() *memDispatchRepo {
	return &memDispatchRepo{dispatches: make(map[string]*dispatch.Dispatch)}
}

func (r *memDispatchRepo) Create(_ context.Context, d *dispatch.Dispatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dispatches[d.ID] = d
	return nil
}
func (r *memDispatchRepo) GetByID(_ context.Context, id string) (*dispatch.Dispatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.dispatches[id]
	if !ok {
		return nil, dispatch.ErrDispatchNotFound
	}
	return d, nil
}
func (r *memDispatchRepo) GetByShipment(_ context.Context, shipmentID string) (*dispatch.Dispatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, d := range r.dispatches {
		if d.ShipmentID == shipmentID {
			return d, nil
		}
	}
	return nil, dispatch.ErrDispatchNotFound
}
func (r *memDispatchRepo) Update(_ context.Context, d *dispatch.Dispatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dispatches[d.ID] = d
	return nil
}
func (r *memDispatchRepo) List(_ context.Context, filter dispatch.DispatchFilter) ([]*dispatch.Dispatch, int64, error) {
	return nil, 0, nil
}

func TestCreateDispatch(t *testing.T) {
	svc := dispatch.NewService(newMemDispatchRepo(), nil)
	d := &dispatch.Dispatch{ID: "d-001", ShipmentID: "s-001", ZoneID: "zone-1"}
	if err := svc.CreateDispatch(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	if d.Status != dispatch.DispatchPending {
		t.Errorf("expected pending, got %v", d.Status)
	}
}

func TestAssignCourier(t *testing.T) {
	svc := dispatch.NewService(newMemDispatchRepo(), nil)
	svc.CreateDispatch(context.Background(), &dispatch.Dispatch{ID: "d-002", ShipmentID: "s-002"})
	if err := svc.AssignCourier(context.Background(), "d-002", "courier-1"); err != nil {
		t.Fatal(err)
	}
	d, _ := svc.GetByShipment(context.Background(), "s-002")
	if d.CourierID != "courier-1" {
		t.Errorf("expected courier-1, got %s", d.CourierID)
	}
	if d.Status != dispatch.DispatchAssigned {
		t.Errorf("expected assigned, got %v", d.Status)
	}
}

func TestMarkEnRoute(t *testing.T) {
	svc := dispatch.NewService(newMemDispatchRepo(), nil)
	svc.CreateDispatch(context.Background(), &dispatch.Dispatch{ID: "d-003", ShipmentID: "s-003"})
	svc.AssignCourier(context.Background(), "d-003", "courier-2")
	if err := svc.MarkEnRoute(context.Background(), "d-003"); err != nil {
		t.Fatal(err)
	}
	d, _ := svc.GetByShipment(context.Background(), "s-003")
	if d.Status != dispatch.DispatchEnRoute {
		t.Errorf("expected en_route, got %v", d.Status)
	}
	if d.DispatchTime == nil {
		t.Error("dispatch time should be set")
	}
}

func TestMarkCompleted(t *testing.T) {
	svc := dispatch.NewService(newMemDispatchRepo(), nil)
	svc.CreateDispatch(context.Background(), &dispatch.Dispatch{ID: "d-004", ShipmentID: "s-004"})
	svc.AssignCourier(context.Background(), "d-004", "courier-3")
	svc.MarkEnRoute(context.Background(), "d-004")
	if err := svc.MarkCompleted(context.Background(), "d-004"); err != nil {
		t.Fatal(err)
	}
	d, _ := svc.GetByShipment(context.Background(), "s-004")
	if d.Status != dispatch.DispatchCompleted {
		t.Errorf("expected completed, got %v", d.Status)
	}
	if d.CompletedAt == nil {
		t.Error("completed_at should be set")
	}
}
