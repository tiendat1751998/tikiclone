package behavior

import (
	"context"
	"strings"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
)

type Analyzer struct {
	profileRepo ProfileRepository
	ruleRepo    RuleRepository
}

func NewAnalyzer(profileRepo ProfileRepository, ruleRepo RuleRepository) *Analyzer {
	return &Analyzer{
		profileRepo: profileRepo,
		ruleRepo:    ruleRepo,
	}
}

func (a *Analyzer) BuildProfile(ctx context.Context, userID string, loginHour int, ipRange string, device string) (*UserBehaviorProfile, error) {
	profile := &UserBehaviorProfile{
		UserID:           userID,
		TypicalLoginHour: loginHour,
		TypicalIPRange:   ipRange,
		TypicalDevice:    device,
		ActionSequence:   []string{},
		LastUpdated:      time.Now().UTC(),
	}

	if err := a.profileRepo.Save(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (a *Analyzer) AnalyzeBehavior(ctx context.Context, ev *core.Event) (*DeviationResult, error) {
	result := &DeviationResult{}
	profile, err := a.profileRepo.Get(ctx, ev.UserID)
	if err != nil {
		return result, nil
	}

	if profile.TypicalLoginHour > 0 {
		eventHour := ev.Timestamp.Hour()
		diff := abs(eventHour - profile.TypicalLoginHour)
		if diff > 6 {
			result.HasDeviation = true
			result.Reasons = append(result.Reasons, "login hour deviates from typical pattern")
			result.Score += 25
		}
	}

	if profile.TypicalIPRange != "" && ev.IP != "" {
		if !strings.HasPrefix(ev.IP, profile.TypicalIPRange) {
			result.HasDeviation = true
			result.Reasons = append(result.Reasons, "ip address outside typical range")
			result.Score += 30
		}
	}

	if profile.TypicalDevice != "" && ev.DeviceID != "" && ev.DeviceID != profile.TypicalDevice {
		result.HasDeviation = true
		result.Reasons = append(result.Reasons, "device differs from typical device")
		result.Score += 20
	}

	profile.ActionSequence = append(profile.ActionSequence, string(ev.Type))
	if len(profile.ActionSequence) > 50 {
		profile.ActionSequence = profile.ActionSequence[len(profile.ActionSequence)-50:]
	}
	profile.LastUpdated = time.Now().UTC()
	a.profileRepo.Save(ctx, profile)

	if result.Score > 100 {
		result.Score = 100
	}

	return result, nil
}

func (a *Analyzer) DetectDeviation(ctx context.Context, profile *UserBehaviorProfile, ev *core.Event) *DeviationResult {
	result := &DeviationResult{}

	if profile.TypicalLoginHour > 0 {
		eventHour := ev.Timestamp.Hour()
		diff := abs(eventHour - profile.TypicalLoginHour)
		if diff > 4 {
			result.HasDeviation = true
			result.Reasons = append(result.Reasons, "unusual login hour")
			result.Score += 20
		}
	}

	if ev.IP != "" && profile.TypicalIPRange != "" && !strings.HasPrefix(ev.IP, profile.TypicalIPRange) {
		result.HasDeviation = true
		result.Reasons = append(result.Reasons, "unusual IP range")
		result.Score += 35
	}

	if ev.DeviceID != "" && profile.TypicalDevice != "" && ev.DeviceID != profile.TypicalDevice {
		result.HasDeviation = true
		result.Reasons = append(result.Reasons, "unusual device")
		result.Score += 25
	}

	if result.Score > 100 {
		result.Score = 100
	}

	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
