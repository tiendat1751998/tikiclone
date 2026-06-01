package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/advertising/internal/campaign"
	"github.com/tikiclone/tiki/platforms/advertising/internal/targeting"
)

func TestMatchAudienceDemographic(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	target := &campaign.Targeting{
		Demographics: &campaign.Demographic{
			MinAge: 18,
			MaxAge: 35,
			Gender: "male",
		},
	}

	profile := &targeting.UserProfile{Age: 25, Gender: "male"}
	match, _ := svc.MatchAudience(ctx, target, profile)
	if !match {
		t.Error("Expected match for 25yo male")
	}

	profile2 := &targeting.UserProfile{Age: 40, Gender: "male"}
	match2, _ := svc.MatchAudience(ctx, target, profile2)
	if match2 {
		t.Error("Expected no match for 40yo male (over max age)")
	}

	profile3 := &targeting.UserProfile{Age: 25, Gender: "female"}
	match3, _ := svc.MatchAudience(ctx, target, profile3)
	if match3 {
		t.Error("Expected no match for female (gender mismatch)")
	}

	profile4 := &targeting.UserProfile{Age: 15, Gender: "male"}
	match4, _ := svc.MatchAudience(ctx, target, profile4)
	if match4 {
		t.Error("Expected no match for 15yo male (under min age)")
	}
}

func TestMatchAudienceLocation(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	target := &campaign.Targeting{
		Locations: []string{"US", "CA"},
	}

	profile := &targeting.UserProfile{Location: "US"}
	match, _ := svc.MatchAudience(ctx, target, profile)
	if !match {
		t.Error("Expected match for US")
	}

	profile2 := &targeting.UserProfile{Location: "GB"}
	match2, _ := svc.MatchAudience(ctx, target, profile2)
	if match2 {
		t.Error("Expected no match for GB")
	}
}

func TestMatchAudienceDevice(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	target := &campaign.Targeting{
		Devices: []string{"mobile"},
	}

	profile := &targeting.UserProfile{Devices: []string{"mobile", "tablet"}}
	match, _ := svc.MatchAudience(ctx, target, profile)
	if !match {
		t.Error("Expected match for mobile user")
	}

	profile2 := &targeting.UserProfile{Devices: []string{"desktop"}}
	match2, _ := svc.MatchAudience(ctx, target, profile2)
	if match2 {
		t.Error("Expected no match for desktop-only user")
	}
}

func TestMatchAudienceInterests(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	target := &campaign.Targeting{
		Interests: []string{"electronics", "gaming"},
	}

	profile := &targeting.UserProfile{Interests: []string{"gaming", "music"}}
	match, _ := svc.MatchAudience(ctx, target, profile)
	if !match {
		t.Error("Expected match for gaming interest")
	}

	profile2 := &targeting.UserProfile{Interests: []string{"fashion", "beauty"}}
	match2, _ := svc.MatchAudience(ctx, target, profile2)
	if match2 {
		t.Error("Expected no match for unrelated interests")
	}
}

func TestMatchAudienceNilTargeting(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	profile := &targeting.UserProfile{Age: 25, Gender: "male", Location: "US"}
	match, _ := svc.MatchAudience(ctx, nil, profile)
	if !match {
		t.Error("Expected match for nil targeting (no restrictions)")
	}
}

func TestSegmentUsers(t *testing.T) {
	svc := targeting.NewService(targeting.NewInMemoryRepository())
	ctx := context.Background()

	segments := svc.SegmentUsers(ctx, &targeting.UserProfile{
		IsNewUser:      true,
		IsHighValue:    false,
		IsCartAbandoner: false,
	})
	if len(segments) != 1 || segments[0].ID != "new" {
		t.Errorf("Expected 1 segment (new), got %d", len(segments))
	}

	segments2 := svc.SegmentUsers(ctx, &targeting.UserProfile{
		IsNewUser:       false,
		IsHighValue:     true,
		IsCartAbandoner: true,
	})
	if len(segments2) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments2))
	}
}
