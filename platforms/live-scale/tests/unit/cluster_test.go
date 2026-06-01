package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/websocket_cluster"
)

func TestWSClusterRegisterNode(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-001", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	if err := svc.RegisterNode(context.Background(), node); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.Status != websocket_cluster.WSNodeActive {
		t.Errorf("expected active status, got %v", node.Status)
	}
}

func TestWSClusterRegisterDuplicate(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-dup", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	err := svc.RegisterNode(context.Background(), node)
	if err != websocket_cluster.ErrNodeAlreadyExists {
		t.Errorf("expected ErrNodeAlreadyExists, got %v", err)
	}
}

func TestWSClusterAssignRoom(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-room", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	assignment, err := svc.AssignRoom(context.Background(), "room-001", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.RoomID != "room-001" {
		t.Errorf("expected room-001, got %s", assignment.RoomID)
	}
	if assignment.NodeID != "ws-room" {
		t.Errorf("expected ws-room, got %s", assignment.NodeID)
	}
}

func TestWSClusterAssignRoomReusesExisting(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-reuse", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	svc.AssignRoom(context.Background(), "room-002", "")
	assignment, err := svc.AssignRoom(context.Background(), "room-002", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.NodeID != "ws-reuse" {
		t.Errorf("expected same node ws-reuse, got %s", assignment.NodeID)
	}
}

func TestWSClusterBroadcast(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-bc", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	svc.AssignRoom(context.Background(), "room-bc", "")
	delivered, err := svc.BroadcastAcrossCluster(context.Background(), "room-bc", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(delivered) != 1 {
		t.Errorf("expected 1 node delivered, got %d", len(delivered))
	}
}

func TestWSClusterBroadcastNoRoom(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	_, err := svc.BroadcastAcrossCluster(context.Background(), "room-nonexistent", []byte("hello"))
	if err != websocket_cluster.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestWSClusterGetClientNode(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-cn", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	svc.AssignRoom(context.Background(), "room-cn", "")
	clientNode, err := svc.GetClientNode(context.Background(), "client-1", "room-cn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clientNode.ID != "ws-cn" {
		t.Errorf("expected ws-cn, got %s", clientNode.ID)
	}
}

func TestWSClusterDeregisterNode(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	node := &websocket_cluster.WSNode{ID: "ws-del", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), node)
	if err := svc.DeregisterNode(context.Background(), "ws-del"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetClientNode(context.Background(), "client-x", "room-del")
	if err != websocket_cluster.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound after deregister, got %v", err)
	}
}

func TestWSClusterAssignPreferredNode(t *testing.T) {
	svc := websocket_cluster.NewService(websocket_cluster.NewInMemoryRepository(), nil)
	n1 := &websocket_cluster.WSNode{ID: "ws-p1", Address: "10.0.0.1:8080", Region: "us-east", MaxRooms: 100, MaxClients: 1000}
	n2 := &websocket_cluster.WSNode{ID: "ws-p2", Address: "10.0.0.2:8080", Region: "us-west", MaxRooms: 100, MaxClients: 1000}
	svc.RegisterNode(context.Background(), n1)
	svc.RegisterNode(context.Background(), n2)
	assignment, err := svc.AssignRoom(context.Background(), "room-pref", "ws-p2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.NodeID != "ws-p2" {
		t.Errorf("expected ws-p2 (preferred), got %s", assignment.NodeID)
	}
}
