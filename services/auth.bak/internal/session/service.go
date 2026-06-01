package session

import (
	"context"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByID(ctx context.Context, id string) (*domain.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	CountActiveByUserID(ctx context.Context, userID string) (int, error)
	FindOldestActiveByUserID(ctx context.Context, userID string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
}

type CacheStore interface {
	StoreSession(ctx context.Context, session *domain.Session) error
	DeleteSession(ctx context.Context, sessionID, userID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	BlacklistRefreshToken(ctx context.Context, tokenID string, ttl time.Duration) error
	BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error
	MarkRefreshTokenUsed(ctx context.Context, tokenID string, ttl time.Duration) error
	IsRefreshTokenReused(ctx context.Context, tokenID string) (bool, error)
}

type Service struct {
	repo   SessionRepository
	cache  CacheStore
	cfg    config.SessionConfig
	jwtCfg config.JWTConfig
}

func NewService(repo SessionRepository, cache CacheStore, cfg config.SessionConfig, jwtCfg config.JWTConfig) *Service {
	return &Service{repo: repo, cache: cache, cfg: cfg, jwtCfg: jwtCfg}
}

func (s *Service) Create(ctx context.Context, user *domain.User, ip, userAgent, deviceID string, roles []domain.Role) (*domain.TokenPair, *domain.Session, error) {
	return nil, nil, fmt.Errorf("session creation requires JWT service, use application layer")
}

func (s *Service) CreateSession(ctx context.Context, userID, refreshTokenID, ip, userAgent, deviceID string, expiresAt time.Time) (*domain.Session, error) {
	session := domain.NewSession(userID, ip, userAgent, refreshTokenID, s.cfg.SessionTTL)
	session.DeviceID = deviceID

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("session create: %w", err)
	}

	if s.cache != nil {
		if err := s.cache.StoreSession(ctx, session); err != nil {
			return nil, fmt.Errorf("session cache: %w", err)
		}
	}

	return session, nil
}

func (s *Service) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.repo.FindByID(ctx, sessionID)
}

func (s *Service) GetActive(ctx context.Context, userID string) ([]*domain.Session, error) {
	return s.repo.FindActiveByUserID(ctx, userID)
}

func (s *Service) Revoke(ctx context.Context, session *domain.Session) error {
	session.Revoke()
	if err := s.repo.Update(ctx, session); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.DeleteSession(ctx, session.ID, session.UserID)
	}
	if s.jwtCfg.BlacklistEnabled && s.cache != nil {
		s.cache.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.jwtCfg.RefreshTTL)
	}
	return nil
}

func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	sessions, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		session.Revoke()
		s.repo.Update(ctx, session)
		if s.jwtCfg.BlacklistEnabled && s.cache != nil {
			s.cache.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.jwtCfg.RefreshTTL)
		}
	}
	if s.cache != nil {
		s.cache.RevokeAllUserSessions(ctx, userID)
	}
	return nil
}

func (s *Service) EnforceMaxSessions(ctx context.Context, userID string) error {
	activeCount, err := s.repo.CountActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if activeCount >= s.cfg.MaxSessionsPerUser {
		oldest, err := s.repo.FindOldestActiveByUserID(ctx, userID)
		if err != nil {
			return err
		}
		oldest.Revoke()
		s.repo.Update(ctx, oldest)
	}
	return nil
}

func (s *Service) IsExpired(session *domain.Session) bool {
	return session.IsExpired() || session.Status != domain.SessionActive
}

func (s *Service) Touch(ctx context.Context, session *domain.Session) error {
	session.Touch()
	return s.repo.Update(ctx, session)
}

func (s *Service) HandleTokenReuse(ctx context.Context, userID, ip string) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "session.handle_reuse")
	defer span.End()

	sessions, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		observability.LogWithTrace(ctx).Error("token reuse: failed to find active sessions", zap.String("user_id", userID), zap.Error(err))
		return
	}
	for _, session := range sessions {
		session.Revoke()
		if err := s.repo.Update(ctx, session); err != nil {
			observability.LogWithTrace(ctx).Error("token reuse: failed to revoke session", zap.String("session_id", session.ID), zap.Error(err))
		}
		if s.cache != nil {
			if err := s.cache.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.jwtCfg.RefreshTTL); err != nil {
				observability.LogWithTrace(ctx).Error("token reuse: failed to blacklist token", zap.Error(err))
			}
		}
	}
	if s.cache != nil {
		s.cache.RevokeAllUserSessions(ctx, userID)
	}
}
