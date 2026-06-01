package cdn

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

func (s *Service) PurgeCache(ctx context.Context, req *CDNPurgeRequest) error {
	if req.Reason == "" {
		return ErrInvalidPurgeRequest
	}
	if len(req.URLs) == 0 && req.Pattern == "" && len(req.Tags) == 0 {
		return ErrInvalidPurgeRequest
	}
	req.ID = fmt.Sprintf("purge-%d", time.Now().UnixNano())
	req.RequestedAt = time.Now().UTC()

	if err := s.repo.CreatePurgeRequest(ctx, req); err != nil {
		return fmt.Errorf("purge cache: %w", err)
	}

	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.CDNPurged,
			Source:    "live-scale.cdn",
			Payload:   req,
			Timestamp: time.Now().UTC(),
		})
	}
	return nil
}

func (s *Service) GetCDNEndpoint(ctx context.Context, viewerRegion string) (*CDNEndpoint, error) {
	endpoints, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	var candidates []*CDNEndpoint
	for _, ep := range endpoints {
		if ep.Status == "active" && ep.CurrentLoad < ep.Capacity {
			candidates = append(candidates, ep)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoEndpointsAvailable
	}

	sort.Slice(candidates, func(i, j int) bool {
		iSame := candidates[i].Region == viewerRegion
		jSame := candidates[j].Region == viewerRegion
		if iSame != jSame {
			return iSame
		}
		return candidates[i].LatencyMs < candidates[j].LatencyMs
	})

	return candidates[0], nil
}

func (s *Service) InvalidateEdge(ctx context.Context, url string, region string) error {
	_, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return err
	}
	req := &CDNPurgeRequest{
		URLs:         []string{url},
		Reason:       "edge_invalidation",
		RequestedAt:  time.Now().UTC(),
	}
	return s.PurgeCache(ctx, req)
}

func (s *Service) PrefetchContent(ctx context.Context, urls []string, region string) error {
	_, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return err
	}
	return nil
}
