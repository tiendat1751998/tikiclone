package unit

import (
	"testing"

	"github.com/tikiclone/tiki/services/shipment/internal/domain"
)

func TestNewShipment(t *testing.T) {
	origin := domain.Address{Street1: "123 Warehouse Rd", City: "Singapore", Country: "SG"}
	dest := domain.Address{Street1: "456 Customer St", City: "Singapore", Country: "SG"}
	s := domain.NewShipment("order-1", "user-1", "ninja_van", "idem-1", "SGD", origin, dest, 1.5)
	if s.OrderID != "order-1" { t.Errorf("expected order-1, got %s", s.OrderID) }
	if s.Status != domain.ShipmentStatusPending { t.Errorf("expected pending, got %s", s.Status) }
	if s.Weight != 1.5 { t.Errorf("expected weight 1.5, got %f", s.Weight) }
}

func TestShipment_CanTransitionTo(t *testing.T) {
	tests := []struct{ from, to domain.ShipmentStatus; expected bool }{
		{domain.ShipmentStatusPending, domain.ShipmentStatusBooked, true},
		{domain.ShipmentStatusPending, domain.ShipmentStatusPickedUp, false},
		{domain.ShipmentStatusBooked, domain.ShipmentStatusPickedUp, true},
		{domain.ShipmentStatusDelivered, domain.ShipmentStatusReturned, true},
		{domain.ShipmentStatusCancelled, domain.ShipmentStatusBooked, false},
	}
	for _, tt := range tests {
		s := &domain.Shipment{Status: tt.from}
		if s.CanTransitionTo(tt.to) != tt.expected {
			t.Errorf("CanTransitionTo(%s,%s)=%v want %v", tt.from, tt.to, s.CanTransitionTo(tt.to), tt.expected)
		}
	}
}

func TestShipment_TransitionTo(t *testing.T) {
	s := &domain.Shipment{ID: "ship-1", Status: domain.ShipmentStatusPending, Version: 1}
	if err := s.TransitionTo(domain.ShipmentStatusBooked); err != nil { t.Fatalf("unexpected error: %v", err) }
	if s.Status != domain.ShipmentStatusBooked { t.Errorf("expected booked, got %s", s.Status) }
	if s.Version != 2 { t.Errorf("expected version 2, got %d", s.Version) }
}

func TestShipment_IsTerminal(t *testing.T) {
	tests := []struct{ status domain.ShipmentStatus; expected bool }{
		{domain.ShipmentStatusPending, false},
		{domain.ShipmentStatusInTransit, false},
		{domain.ShipmentStatusDelivered, true},
		{domain.ShipmentStatusReturned, true},
		{domain.ShipmentStatusCancelled, true},
	}
	for _, tt := range tests {
		s := &domain.Shipment{Status: tt.status}
		if s.IsTerminal() != tt.expected {
			t.Errorf("IsTerminal(%s)=%v want %v", tt.status, s.IsTerminal(), tt.expected)
		}
	}
}

func TestTrackingEvent(t *testing.T) {
	event := domain.NewTrackingEvent("ship-1", "in_transit", "Singapore Hub", "Package in transit")
	if event.ShipmentID != "ship-1" { t.Errorf("expected ship-1, got %s", event.ShipmentID) }
	if event.Status != "in_transit" { t.Errorf("expected in_transit, got %s", event.Status) }
}
