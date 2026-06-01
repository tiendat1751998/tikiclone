package transactionmon

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/core"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/devicefp"
)

const (
	maxDailyCount  = 50
	maxHourlyCount = 10
	maxDailyVolume = 100000.0
	maxHourlyVolume = 20000.0
)

type Monitor struct {
	repo       Repository
	deviceSvc  *devicefp.Service
}

func NewMonitor(repo Repository, deviceSvc *devicefp.Service) *Monitor {
	return &Monitor{
		repo:      repo,
		deviceSvc: deviceSvc,
	}
}

func (m *Monitor) RecordTransaction(ctx context.Context, rec *TransactionRecord) (*TransactionMonitor, error) {
	now := time.Now().UTC()
	mon, err := m.repo.Get(ctx, rec.UserID)
	if err != nil {
		mon = &TransactionMonitor{
			UserID:          rec.UserID,
			LastDailyReset:  now,
			LastHourlyReset: now,
		}
	}

	mon.DailyCount++
	mon.HourlyCount++
	mon.DailyVolume += rec.Amount
	mon.HourlyVolume += rec.Amount
	mon.LastTransactions = append(mon.LastTransactions, rec.Amount)

	totalSum := mon.DailyVolume
	totalCount := mon.DailyCount
	if totalCount > 0 {
		mon.AvgTicket = math.Round((totalSum/float64(totalCount))*100) / 100
	}

	if len(mon.LastTransactions) > 100 {
		mon.LastTransactions = mon.LastTransactions[len(mon.LastTransactions)-100:]
	}

	if err := m.repo.Save(ctx, mon); err != nil {
		return nil, err
	}

	return mon, nil
}

func (m *Monitor) GetPattern(ctx context.Context, userID string) (*TransactionMonitor, error) {
	return m.repo.Get(ctx, userID)
}

func (m *Monitor) DetectAnomaly(ctx context.Context, rec *TransactionRecord) (*AnomalyResult, error) {
	result := &AnomalyResult{}

	mon, err := m.repo.Get(ctx, rec.UserID)
	if err != nil {
		return result, nil
	}

	if mon.DailyCount > maxDailyCount {
		result.HasAnomaly = true
		result.Reasons = append(result.Reasons, "daily transaction count exceeded")
	}
	if mon.HourlyCount > maxHourlyCount {
		result.HasAnomaly = true
		result.Reasons = append(result.Reasons, "hourly transaction count exceeded")
	}
	if mon.DailyVolume > maxDailyVolume {
		result.HasAnomaly = true
		result.Reasons = append(result.Reasons, "daily transaction volume exceeded")
	}
	if mon.HourlyVolume > maxHourlyVolume {
		result.HasAnomaly = true
		result.Reasons = append(result.Reasons, "hourly transaction volume exceeded")
	}

	if rec.Amount > 3*mon.AvgTicket && mon.AvgTicket > 0 {
		result.HasAnomaly = true
		result.Reasons = append(result.Reasons, "amount anomaly: exceeds 3x average ticket")
	}

	if rec.Location != "" {
		locationMismatch := m.checkLocationMismatch(ctx, rec.UserID, rec.Location)
		if locationMismatch {
			result.HasAnomaly = true
			result.Reasons = append(result.Reasons, "location mismatch detected")
		}
	}

	result.CurrentVelocity = mon.HourlyCount
	result.AvgTicket = mon.AvgTicket

	return result, nil
}

func (m *Monitor) checkLocationMismatch(ctx context.Context, userID string, currentLocation string) bool {
	mon, err := m.repo.Get(ctx, userID)
	if err != nil || len(mon.LastTransactions) == 0 {
		return false
	}

	if strings.EqualFold(currentLocation, "foreign") && strings.EqualFold(mon.UserID, userID) {
		return true
	}

	return false
}

func (m *Monitor) evaluateEvent(ctx context.Context, ev *core.Event) (*AnomalyResult, error) {
	rec := &TransactionRecord{
		UserID:    ev.UserID,
		Amount:    ev.Amount,
		Timestamp: ev.Timestamp,
		Location:  "",
		IP:        ev.IP,
		DeviceID:  ev.DeviceID,
	}

	if ev.Metadata != nil {
		if loc, ok := ev.Metadata["location"].(string); ok {
			rec.Location = loc
		}
	}

	return m.DetectAnomaly(ctx, rec)
}
