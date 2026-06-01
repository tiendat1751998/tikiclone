package targeting

import (
	"context"

	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
)

type Service interface {
	MatchAudience(ctx context.Context, targeting *campaign.Targeting, profile *UserProfile) (bool, error)
	SegmentUsers(ctx context.Context, profile *UserProfile) []Segment
	BuildTargetingProfile(ctx context.Context, userID string) (*UserProfile, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) MatchAudience(ctx context.Context, targeting *campaign.Targeting, profile *UserProfile) (bool, error) {
	if targeting == nil {
		return true, nil
	}

	if targeting.Demographics != nil {
		d := targeting.Demographics
		if d.MinAge > 0 && profile.Age < d.MinAge {
			return false, nil
		}
		if d.MaxAge > 0 && profile.Age > d.MaxAge {
			return false, nil
		}
		if d.Gender != "" && d.Gender != profile.Gender {
			return false, nil
		}
	}

	if len(targeting.Locations) > 0 {
		found := false
		for _, loc := range targeting.Locations {
			if loc == profile.Location {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	if len(targeting.Devices) > 0 {
		found := false
		for _, d := range targeting.Devices {
			for _, pd := range profile.Devices {
				if d == pd {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	if len(targeting.Interests) > 0 {
		found := false
		for _, interest := range targeting.Interests {
			for _, pi := range profile.Interests {
				if interest == pi {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (s *service) SegmentUsers(ctx context.Context, profile *UserProfile) []Segment {
	var segments []Segment
	if profile.IsNewUser {
		segments = append(segments, Segment{ID: "new", Name: "New Users"})
	}
	if !profile.IsNewUser {
		segments = append(segments, Segment{ID: "returning", Name: "Returning Users"})
	}
	if profile.IsHighValue {
		segments = append(segments, Segment{ID: "high_value", Name: "High Value Users"})
	}
	if profile.IsCartAbandoner {
		segments = append(segments, Segment{ID: "cart_abandoner", Name: "Cart Abandoners"})
	}
	return segments
}

func (s *service) BuildTargetingProfile(ctx context.Context, userID string) (*UserProfile, error) {
	return s.repo.GetProfile(ctx, userID)
}
