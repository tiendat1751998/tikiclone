package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

type Service struct {
	repo     Repository
	eventSvc *events.Service
}

func NewService(repo Repository, eventSvc *events.Service) *Service {
	return &Service{repo: repo, eventSvc: eventSvc}
}

func (s *Service) RunQuery(ctx context.Context, query *AnalyticsQuery) (*QueryResult, error) {
	if query == nil {
		return nil, ErrQueryInvalid
	}
	if len(query.Metrics) == 0 {
		return nil, fmt.Errorf("at least one metric is required")
	}

	startTime, endTime := s.resolveTimeRange(query.TimeRange)

	columns := make([]string, 0)
	for _, m := range query.Metrics {
		alias := m.Alias
		if alias == "" {
			alias = string(m.Name)
		}
		columns = append(columns, alias)
	}
	for _, d := range query.Dimensions {
		columns = append(columns, string(d))
	}

	if len(query.GroupBy) > 0 {
		return s.runGroupedQuery(ctx, query, startTime, endTime)
	}

	row := make(map[string]interface{})
	for _, m := range query.Metrics {
		alias := m.Alias
		if alias == "" {
			alias = string(m.Name)
		}
		row[alias] = s.computeMetric(ctx, m, startTime, endTime)
	}
	for _, d := range query.Dimensions {
		row[string(d)] = ""
	}

	rows := []map[string]interface{}{row}
	if query.Limit > 0 && len(rows) > query.Limit {
		rows = rows[:query.Limit]
	}

	return &QueryResult{
		Columns: columns,
		Rows:    rows,
		Total:   int64(len(rows)),
	}, nil
}

func (s *Service) runGroupedQuery(ctx context.Context, query *AnalyticsQuery, startTime, endTime time.Time) (*QueryResult, error) {
	groupField := query.GroupBy[0]
	eventList, _, _ := s.eventSvc.ListEvents(ctx, "", startTime, endTime, 0, 10000)

	groups := make(map[string][]*events.AnalyticsEvent)
	for _, e := range eventList {
		key := s.getGroupKey(e, groupField)
		groups[key] = append(groups[key], e)
	}

	columns := make([]string, 0)
	for _, m := range query.Metrics {
		alias := m.Alias
		if alias == "" {
			alias = string(m.Name)
		}
		columns = append(columns, alias)
	}
	columns = append(columns, groupField)

	var rows []map[string]interface{}
	for groupKey, groupEvents := range groups {
		row := make(map[string]interface{})
		for _, m := range query.Metrics {
			alias := m.Alias
			if alias == "" {
				alias = string(m.Name)
			}
			row[alias] = s.computeMetricForEvents(m, groupEvents)
		}
		row[groupField] = groupKey
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return fmt.Sprintf("%v", rows[i][groupField]) < fmt.Sprintf("%v", rows[j][groupField])
	})

	if query.Limit > 0 && len(rows) > query.Limit {
		rows = rows[:query.Limit]
	}

	return &QueryResult{
		Columns: columns,
		Rows:    rows,
		Total:   int64(len(rows)),
	}, nil
}

func (s *Service) getGroupKey(e *events.AnalyticsEvent, field string) string {
	switch field {
	case "source":
		if e.Source != "" {
			return e.Source
		}
	case "device":
		if e.Device != "" {
			return e.Device
		}
	case "country":
		if e.Country != "" {
			return e.Country
		}
	case "campaign":
		if e.Campaign != "" {
			return e.Campaign
		}
	case "event_type":
		return string(e.EventType)
	case "date":
		return e.Timestamp.Format("2006-01-02")
	}
	return "unknown"
}

func (s *Service) computeMetric(ctx context.Context, m Metric, startTime, endTime time.Time) float64 {
	switch m.Name {
	case MetricRevenue:
		rev, _ := s.eventSvc.GetRevenue(ctx, startTime, endTime)
		return math.Round(rev*100) / 100
	case MetricTotalUsers:
		users, _ := s.eventSvc.GetUniqueUsers(ctx, startTime, endTime)
		return float64(users)
	case MetricActiveUsers:
		users, _ := s.eventSvc.GetActiveUsers(ctx, startTime, endTime)
		return float64(users)
	case MetricOrders:
		orders, _ := s.eventSvc.GetOrders(ctx, startTime, endTime)
		return float64(orders)
	case MetricPageviews:
		count, _ := s.eventSvc.GetEventCount(ctx, events.EventPageview, startTime, endTime)
		return float64(count)
	case MetricAOV:
		revenue, _ := s.eventSvc.GetRevenue(ctx, startTime, endTime)
		orders, _ := s.eventSvc.GetOrders(ctx, startTime, endTime)
		if orders == 0 {
			return 0
		}
		return math.Round(revenue/float64(orders)*100) / 100
	case MetricSessions:
		return 0
	case MetricConversionRate:
		return 0
	case MetricBounceRate:
		return 0
	case MetricAvgSessionDuration:
		return 0
	}
	return 0
}

func (s *Service) computeMetricForEvents(m Metric, eventsList []*events.AnalyticsEvent) float64 {
	var values []float64
	for _, e := range eventsList {
		switch m.Name {
		case MetricRevenue:
			values = append(values, e.Revenue)
		case MetricOrders:
			if e.EventType == events.EventPurchase {
				values = append(values, 1)
			}
		case MetricPageviews:
			if e.EventType == events.EventPageview {
				values = append(values, 1)
			}
		case MetricTotalUsers, MetricActiveUsers:
			values = append(values, 1)
		}
	}
	if len(values) == 0 {
		return 0
	}
	switch m.Aggregation {
	case AggSum, AggCount:
		var sum float64
		for _, v := range values {
			sum += v
		}
		return sum
	case AggAvg:
		var sum float64
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	case AggMin:
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
	case AggMax:
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
	case AggDistinctCount:
		seen := make(map[float64]bool)
		for _, v := range values {
			seen[v] = true
		}
		return float64(len(seen))
	}
	return 0
}

func (s *Service) GetKeyMetrics(ctx context.Context, timeRange TimeRange) ([]MetricValue, error) {
	startTime, endTime := s.resolveTimeRange(timeRange)

	type metricDef struct {
		name  MetricType
		label string
	}
	metrics := []metricDef{
		{MetricTotalUsers, "Total Users"},
		{MetricActiveUsers, "Active Users"},
		{MetricRevenue, "Revenue"},
		{MetricOrders, "Orders"},
		{MetricAOV, "Average Order Value"},
	}

	result := make([]MetricValue, 0, len(metrics))
	for _, m := range metrics {
		val := s.computeMetric(ctx, Metric{Name: m.name, Aggregation: AggSum}, startTime, endTime)
		result = append(result, MetricValue{
			Name:  m.name,
			Value: val,
			Label: m.label,
		})
	}
	return result, nil
}

func (s *Service) resolveTimeRange(tr TimeRange) (time.Time, time.Time) {
	now := time.Now()
	start := now.Truncate(24 * time.Hour)
	end := now

	switch tr.Type {
	case TimeToday:
		start = now.Truncate(24 * time.Hour)
	case TimeYesterday:
		start = now.Truncate(24 * time.Hour).Add(-24 * time.Hour)
		end = now.Truncate(24 * time.Hour)
	case TimeLast7d:
		start = now.AddDate(0, 0, -7)
	case TimeLast30d:
		start = now.AddDate(0, 0, -30)
	case TimeLast90d:
		start = now.AddDate(0, 0, -90)
	case TimeCustom:
		if tr.StartAt != nil {
			start = *tr.StartAt
		}
		if tr.EndAt != nil {
			end = *tr.EndAt
		}
	}
	return start, end
}

func (s *Service) RunQueryAndSave(ctx context.Context, query *AnalyticsQuery, name string) (*Report, error) {
	result, err := s.RunQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	report := &Report{
		ID:        uuid.New().String(),
		Name:      name,
		Query:     *query,
		Result:    result,
		CreatedAt: time.Now(),
	}
	if err := s.repo.StoreQueryResult(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}
