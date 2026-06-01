package region

import (
	"context"
	"math"
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

func (s *Service) GetNearestRegion(ctx context.Context, viewerRegion string) (*Region, error) {
	regions, err := s.repo.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	var candidates []*Region
	for _, r := range regions {
		if r.Status == RegionActive {
			if r.Code == viewerRegion {
				return r, nil
			}
			candidates = append(candidates, r)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoRegionAvailable
	}

	latencies, err := s.repo.ListLatencies(ctx)
	if err != nil {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Latency < candidates[j].Latency
		})
		return candidates[0], nil
	}

	type scored struct {
		region  *Region
		latency int
	}

	var scoredCandidates []scored
	for _, r := range candidates {
		latency := r.Latency
		for _, lm := range latencies {
			if lm.FromRegion == viewerRegion && lm.ToRegion == r.Code {
				latency = lm.LatencyMs
				break
			}
		}
		scoredCandidates = append(scoredCandidates, scored{region: r, latency: latency})
	}

	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].latency < scoredCandidates[j].latency
	})

	return scoredCandidates[0].region, nil
}

func (s *Service) RouteToRegion(ctx context.Context, viewerRegion string) (*Region, error) {
	regions, err := s.repo.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range regions {
		if r.Code == viewerRegion && r.Status == RegionActive {
			return r, nil
		}
	}

	return s.GetNearestRegion(ctx, viewerRegion)
}

func (s *Service) GetRegionLatency(ctx context.Context, from, to string) (int, error) {
	latencies, err := s.repo.ListLatencies(ctx)
	if err != nil {
		return 0, err
	}

	for _, lm := range latencies {
		if lm.FromRegion == from && lm.ToRegion == to {
			return lm.LatencyMs, nil
		}
	}

	regions, err := s.repo.ListRegions(ctx)
	if err != nil {
		return 0, err
	}
	for _, r := range regions {
		if r.Code == to {
			return r.Latency, nil
		}
	}

	return 0, ErrRegionNotFound
}

func (s *Service) FailoverRegion(ctx context.Context, failedRegion string) (*Region, error) {
	regions, err := s.repo.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	region, err := s.repo.GetRegion(ctx, failedRegion)
	if err != nil {
		return nil, err
	}

	region.Status = RegionDown
	region.UpdatedAt = time.Now().UTC()
	s.repo.SaveRegion(ctx, region)

	var candidates []*Region
	for _, r := range regions {
		if r.Code != failedRegion && (r.Status == RegionActive || r.Status == RegionDegraded) {
			candidates = append(candidates, r)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoRegionAvailable
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Latency < candidates[j].Latency
	})

	selected := candidates[0]
	if s.producer != nil {
		s.producer.Publish(ctx, events.Event{
			Type:      events.RegionFailover,
			Source:    "live-scale.region",
			Payload: map[string]interface{}{
				"failed_region":  failedRegion,
				"failover_to":    selected.Code,
				"failover_latency_ms": selected.Latency,
			},
			Timestamp: time.Now().UTC(),
		})
	}

	return selected, nil
}

func (s *Service) SetRegionStatus(ctx context.Context, regionCode string, status RegionStatus) error {
	region, err := s.repo.GetRegion(ctx, regionCode)
	if err != nil {
		return err
	}
	region.Status = status
	region.UpdatedAt = time.Now().UTC()
	return s.repo.SaveRegion(ctx, region)
}

func (s *Service) RecordLatency(ctx context.Context, fromRegion, toRegion string, latencyMs int) error {
	if latencyMs < 0 {
		return ErrInvalidLatencyData
	}
	if math.Abs(float64(latencyMs)) > 10000 {
		return ErrInvalidLatencyData
	}
	return s.repo.SaveLatency(ctx, &LatencyMap{
		FromRegion: fromRegion,
		ToRegion:   toRegion,
		LatencyMs:  latencyMs,
	})
}

func (s *Service) CheckRegionHealth(ctx context.Context, regionCode string) (*Region, error) {
	region, err := s.repo.GetRegion(ctx, regionCode)
	if err != nil {
		return nil, err
	}
	if time.Since(region.UpdatedAt) > 60*time.Second {
		region.Status = RegionDegraded
		s.repo.SaveRegion(ctx, region)
	}
	return region, nil
}
