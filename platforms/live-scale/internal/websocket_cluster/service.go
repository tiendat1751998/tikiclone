package websocket_cluster

import (
	"context"
	"fmt"
	"sort"
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

func (s *Service) RegisterNode(ctx context.Context, node *WSNode) error {
	if node.ID == "" || node.Address == "" {
		return fmt.Errorf("node id and address required")
	}
	node.Status = WSNodeActive
	node.LastHeartbeat = time.Now().UTC()
	node.RegisteredAt = time.Now().UTC()

	if existing, _ := s.repo.GetNode(ctx, node.ID); existing != nil {
		return ErrNodeAlreadyExists
	}

	if err := s.repo.SaveNode(ctx, node); err != nil {
		return fmt.Errorf("register ws node: %w", err)
	}

	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.NodeJoined,
			Source:    "live-scale.wscluster",
			Payload:   node,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) AssignRoom(ctx context.Context, roomID string, preferredNodeID string) (*RoomAssignment, error) {
	if existingNodeID, err := s.repo.GetRoomNode(ctx, roomID); err == nil {
		node, err := s.repo.GetNode(ctx, existingNodeID)
		if err == nil && node.Status == WSNodeActive {
			return &RoomAssignment{
				RoomID:     roomID,
				NodeID:     existingNodeID,
				Clients:    0,
				AssignedAt: time.Now().UTC(),
			}, nil
		}
	}

	if preferredNodeID != "" {
		node, err := s.repo.GetNode(ctx, preferredNodeID)
		if err == nil && node.Status == WSNodeActive && node.RoomCount < node.MaxRooms {
			if err := s.repo.AssignRoom(ctx, roomID, preferredNodeID); err != nil {
				return nil, err
			}
			node.RoomCount++
			s.repo.SaveNode(ctx, node)
			return &RoomAssignment{
				RoomID:     roomID,
				NodeID:     preferredNodeID,
				Clients:    0,
				AssignedAt: time.Now().UTC(),
			}, nil
		}
	}

	nodes, err := s.repo.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	var candidates []*WSNode
	for _, n := range nodes {
		if n.Status == WSNodeActive && n.RoomCount < n.MaxRooms {
			candidates = append(candidates, n)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableNode
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].RoomCount < candidates[j].RoomCount
	})

	selected := candidates[0]
	if err := s.repo.AssignRoom(ctx, roomID, selected.ID); err != nil {
		return nil, err
	}
	selected.RoomCount++
	s.repo.SaveNode(ctx, selected)

	return &RoomAssignment{
		RoomID:     roomID,
		NodeID:     selected.ID,
		Clients:    0,
		AssignedAt: time.Now().UTC(),
	}, nil
}

func (s *Service) GetClientNode(ctx context.Context, clientID, roomID string) (*WSNode, error) {
	nodeID, err := s.repo.GetRoomNode(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetNode(ctx, nodeID)
}

func (s *Service) BroadcastAcrossCluster(ctx context.Context, roomID string, message []byte) ([]string, error) {
	nodeID, err := s.repo.GetRoomNode(ctx, roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	var delivered []string
	if node.Status == WSNodeActive || node.Status == WSNodeDegraded {
		delivered = append(delivered, nodeID)
	}

	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.BroadcastMessage,
			Source:    "live-scale.wscluster",
			Payload:   map[string]interface{}{"room_id": roomID, "node_id": nodeID, "message_size": len(message)},
			Timestamp: time.Now().UTC(),
		})
	}

	return delivered, nil
}

func (s *Service) DeregisterNode(ctx context.Context, nodeID string) error {
	if err := s.repo.DeleteNode(ctx, nodeID); err != nil {
		return fmt.Errorf("deregister ws node: %w", err)
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.NodeLeft,
			Source:    "live-scale.wscluster",
			Payload:   map[string]string{"node_id": nodeID},
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}
