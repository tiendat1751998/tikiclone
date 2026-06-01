package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/sfu"
)

func TestSFURegisterNode(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	node := &sfu.SFUNode{ID: "sfu-001", Address: "192.168.1.1:9000", Region: "us-east", Capacity: 100}
	if err := svc.RegisterNode(context.Background(), node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Status != sfu.NodeStatusActive {
		t.Errorf("expected active status, got %v", node.Status)
	}
}

func TestSFUDuplicateRegistration(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	node := &sfu.SFUNode{ID: "sfu-dup", Address: "10.0.0.1:9000", Region: "ap-southeast", Capacity: 50}
	if err := svc.RegisterNode(context.Background(), node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := svc.RegisterNode(context.Background(), node); err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestSFUSelectOptimalNodeSameRegion(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	n1 := &sfu.SFUNode{ID: "sfu-a", Address: "10.0.0.1:9000", Region: "us-east", Capacity: 100, CurrentLoad: 50}
	n2 := &sfu.SFUNode{ID: "sfu-b", Address: "10.0.0.2:9000", Region: "us-west", Capacity: 100, CurrentLoad: 10}
	n3 := &sfu.SFUNode{ID: "sfu-c", Address: "10.0.0.3:9000", Region: "us-east", Capacity: 100, CurrentLoad: 20}
	svc.RegisterNode(context.Background(), n1)
	svc.RegisterNode(context.Background(), n2)
	svc.RegisterNode(context.Background(), n3)

	selected, err := svc.SelectOptimalNode(context.Background(), "stream-1", "us-east")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected.ID != "sfu-c" {
		t.Errorf("expected sfu-c (lowest load in us-east), got %s", selected.ID)
	}
}

func TestSFUSelectOptimalNodeFallback(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	n1 := &sfu.SFUNode{ID: "sfu-x", Address: "10.0.0.1:9000", Region: "eu-west", Capacity: 100, CurrentLoad: 30}
	n2 := &sfu.SFUNode{ID: "sfu-y", Address: "10.0.0.2:9000", Region: "eu-west", Capacity: 100, CurrentLoad: 60}
	svc.RegisterNode(context.Background(), n1)
	svc.RegisterNode(context.Background(), n2)

	selected, err := svc.SelectOptimalNode(context.Background(), "stream-2", "eu-central")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected.ID != "sfu-x" {
		t.Errorf("expected sfu-x (lowest load), got %s", selected.ID)
	}
}

func TestSFUSelectOptimalNodeNoAvailable(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	_, err := svc.SelectOptimalNode(context.Background(), "stream-3", "us-east")
	if err != sfu.ErrNoAvailableNodes {
		t.Errorf("expected ErrNoAvailableNodes, got %v", err)
	}
}

func TestSFUSelectOptimalNodeFullCapacity(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	n1 := &sfu.SFUNode{ID: "sfu-full", Address: "10.0.0.1:9000", Region: "us-east", Capacity: 100, CurrentLoad: 100}
	svc.RegisterNode(context.Background(), n1)
	_, err := svc.SelectOptimalNode(context.Background(), "stream-4", "us-east")
	if err != sfu.ErrNoAvailableNodes {
		t.Errorf("expected ErrNoAvailableNodes for full node, got %v", err)
	}
}

func TestSFUDeregisterNode(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	node := &sfu.SFUNode{ID: "sfu-del", Address: "10.0.0.1:9000", Region: "us-east", Capacity: 50}
	svc.RegisterNode(context.Background(), node)
	if err := svc.DeregisterNode(context.Background(), "sfu-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.SelectOptimalNode(context.Background(), "stream-5", "us-east")
	if err != sfu.ErrNoAvailableNodes {
		t.Errorf("expected ErrNoAvailableNodes after deregister, got %v", err)
	}
}

func TestSFUHeartbeatNode(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	node := &sfu.SFUNode{ID: "sfu-hb", Address: "10.0.0.1:9000", Region: "us-east", Capacity: 100}
	svc.RegisterNode(context.Background(), node)
	if err := svc.HeartbeatNode(context.Background(), "sfu-hb"); err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
}

func TestSFURebalanceNoNode(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	_, err := svc.RebalanceLoad(context.Background(), "nonexistent")
	if err != sfu.ErrNodeNotFound {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestSFUInvalidNodeData(t *testing.T) {
	svc := sfu.NewService(sfu.NewInMemoryRepository(), nil)
	err := svc.RegisterNode(context.Background(), &sfu.SFUNode{})
	if err != sfu.ErrInvalidNodeData {
		t.Errorf("expected ErrInvalidNodeData, got %v", err)
	}
}
