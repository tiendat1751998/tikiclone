package sfu

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/tikiclone/tiki/platforms/live-scale/internal/events"
)

type Service struct {
	repo     Repository
	producer events.Producer
	mu       sync.RWMutex
}

func NewService(repo Repository, producer events.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) RegisterNode(ctx context.Context, node *SFUNode) error {
	if node.ID == "" || node.Address == "" || node.Region == "" {
		return ErrInvalidNodeData
	}
	if existing, _ := s.repo.GetNode(ctx, node.ID); existing != nil {
		return ErrNodeAlreadyExists
	}
	node.Status = NodeStatusActive
	node.LastHeartbeat = time.Now().UTC()
	node.RegisteredAt = time.Now().UTC()
	if err := s.repo.SaveNode(ctx, node); err != nil {
		return fmt.Errorf("register node: %w", err)
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.NodeJoined,
			Source:    "live-scale.sfu",
			Payload:   node,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) DeregisterNode(ctx context.Context, nodeID string) error {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteNode(ctx, nodeID); err != nil {
		return fmt.Errorf("deregister node: %w", err)
	}
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.NodeLeft,
			Source:    "live-scale.sfu",
			Payload:   node,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) SelectOptimalNode(ctx context.Context, streamID string, region string) (*SFUNode, error) {
	nodes, err := s.repo.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	var candidates []*SFUNode
	for _, n := range nodes {
		if n.Status == NodeStatusActive || n.Status == NodeStatusDegraded {
			if n.CurrentLoad < n.Capacity {
				candidates = append(candidates, n)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableNodes
	}

	sort.Slice(candidates, func(i, j int) bool {
		iSame := candidates[i].Region == region
		jSame := candidates[j].Region == region
		if iSame != jSame {
			return iSame
		}
		iLoad := float64(candidates[i].CurrentLoad) / float64(candidates[i].Capacity)
		jLoad := float64(candidates[j].CurrentLoad) / float64(candidates[j].Capacity)
		if math.Abs(iLoad-jLoad) > 0.001 {
			return iLoad < jLoad
		}
		return candidates[i].StreamCount < candidates[j].StreamCount
	})

	return candidates[0], nil
}

func (s *Service) GetStreamSession(ctx context.Context, sessionID string) (*StreamSession, error) {
	return s.repo.GetStreamSession(ctx, sessionID)
}

func (s *Service) CreateStreamSession(ctx context.Context, session *StreamSession) error {
	session.StartedAt = time.Now().UTC()
	session.LastActive = session.StartedAt
	return s.repo.SaveStreamSession(ctx, session)
}

func (s *Service) RebalanceLoad(ctx context.Context, nodeID string) ([]*StreamSession, error) {
	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	sessions, err := s.repo.ListStreamSessionsByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, nil
	}

	availableNodes, err := s.repo.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	var targets []*SFUNode
	for _, n := range availableNodes {
		if n.ID != nodeID && n.Status == NodeStatusActive && n.CurrentLoad < n.Capacity {
			targets = append(targets, n)
		}
	}

	if len(targets) == 0 && node.Status == NodeStatusDown {
		return nil, ErrNoAvailableNodes
	}

	if node.Status == NodeStatusDown && len(targets) == 0 {
		return nil, ErrNoAvailableNodes
	}

	var moved []*StreamSession
	for _, session := range sessions {
		if len(targets) == 0 {
			break
		}

		sort.Slice(targets, func(i, j int) bool {
			iSame := targets[i].Region == session.Region
			jSame := targets[j].Region == session.Region
			if iSame != jSame {
				return iSame
			}
			return targets[i].CurrentLoad < targets[j].CurrentLoad
		})

		target := targets[0]
		target.CurrentLoad++
		target.StreamCount++

		session.NodeID = target.ID
		session.LastActive = time.Now().UTC()
		if err := s.repo.SaveStreamSession(ctx, session); err != nil {
			return moved, fmt.Errorf("rebalance session %s: %w", session.ID, err)
		}
		moved = append(moved, session)
	}

	if node.Status != NodeStatusDown {
		node.CurrentLoad = 0
		node.StreamCount = 0
		s.repo.SaveNode(ctx, node)
	}

	return moved, nil
}

func (s *Service) HeartbeatNode(ctx context.Context, nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, err := s.repo.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}
	node.LastHeartbeat = time.Now().UTC()
	node.Status = NodeStatusActive
	return s.repo.SaveNode(ctx, node)
}

func (s *Service) CheckNodeHealth(ctx context.Context) {
	nodes, _ := s.repo.ListNodes(ctx)
	timeout := 30 * time.Second
	now := time.Now().UTC()
	for _, n := range nodes {
		if now.Sub(n.LastHeartbeat) > timeout {
			n.Status = NodeStatusDown
			s.repo.SaveNode(ctx, n)
		}
	}
}
