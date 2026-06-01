package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/platforms/analytics/internal/events"
)

type Service struct {
	repo     Repository
	eventSvc *events.Service
	timeout  time.Duration
}

func NewService(repo Repository, eventSvc *events.Service, timeoutMinutes int) *Service {
	return &Service{
		repo:     repo,
		eventSvc: eventSvc,
		timeout:  time.Duration(timeoutMinutes) * time.Minute,
	}
}

func (s *Service) TrackSession(ctx context.Context, event *events.AnalyticsEvent) (*Session, error) {
	existing, _ := s.repo.GetActiveSession(ctx, event.UserID)

	if existing != nil {
		if time.Since(existing.StartTime) > s.timeout {
			existing.IsActive = false
			existing.EndTime = time.Now()
			existing.Duration = existing.EndTime.Sub(existing.StartTime)
			s.repo.UpdateSession(ctx, existing)

			return s.createNewSession(ctx, event)
		}

		existing.EventsCount++
		if event.EventType == events.EventPageview {
			existing.Pageviews++
		}
		if event.EventType == events.EventPurchase || event.EventType == events.EventCheckout {
			existing.HasConversion = true
			existing.Revenue += event.Revenue
		}

		sessionEvent := SessionEvent{
			EventType: string(event.EventType),
			Timestamp: event.Timestamp,
		}
		if event.Properties != nil {
			sessionEvent.Data = event.Properties
		}
		existing.Events = append(existing.Events, sessionEvent)

		s.repo.UpdateSession(ctx, existing)
		return existing, nil
	}

	return s.createNewSession(ctx, event)
}

func (s *Service) createNewSession(ctx context.Context, event *events.AnalyticsEvent) (*Session, error) {
	session := &Session{
		SessionID:   uuid.New().String(),
		UserID:      event.UserID,
		StartTime:   event.Timestamp,
		IsActive:    true,
		EventsCount: 1,
		Events: []SessionEvent{
			{
				EventType: string(event.EventType),
				Timestamp: event.Timestamp,
			},
		},
		Device:  event.Device,
		Source:  event.Source,
		Country: event.Country,
	}

	if event.EventType == events.EventPageview {
		session.Pageviews = 1
	}
	if event.EventType == events.EventPurchase || event.EventType == events.EventCheckout {
		session.HasConversion = true
		session.Revenue = event.Revenue
	}

	s.repo.StoreSession(ctx, session)
	return session, nil
}

func (s *Service) GetSessions(ctx context.Context, filter *SessionFilter) ([]*Session, int64, error) {
	return s.repo.ListSessions(ctx, filter)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.repo.GetSession(ctx, sessionID)
}

func (s *Service) CalculateSessionMetrics(ctx context.Context, filter *SessionFilter) (*SessionMetrics, error) {
	return s.repo.GetSessionMetrics(ctx, filter)
}

func (s *Service) EndSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}
	session.IsActive = false
	session.EndTime = time.Now()
	session.Duration = session.EndTime.Sub(session.StartTime)
	return s.repo.UpdateSession(ctx, session)
}
