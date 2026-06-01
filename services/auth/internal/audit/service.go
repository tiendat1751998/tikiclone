package audit

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel/trace"
)

type Repository interface {
	Log(ctx context.Context, alog *domain.AuditLog)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.AuditLog, error)
	FindByAction(ctx context.Context, action domain.AuditAction, limit, offset int) ([]*domain.AuditLog, error)
	DeleteOlderThan(ctx context.Context, ttl time.Duration) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, action domain.AuditAction, userID, ip, deviceID, userAgent string) {
	if s.repo == nil {
		return
	}
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	s.repo.Log(ctx, domain.NewAuditLog(traceID, userID, action, ip, deviceID, userAgent))
}

func (s *Service) LogWithDetail(ctx context.Context, action domain.AuditAction, userID, ip, deviceID, userAgent, detail string) {
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	alog := domain.NewAuditLog(traceID, userID, action, ip, deviceID, userAgent)
	alog.Detail = detail
	s.repo.Log(ctx, alog)
}

func (s *Service) LogLoginSuccess(ctx context.Context, userID, ip, deviceID, userAgent string) {
	s.Log(ctx, domain.AuditLogin, userID, ip, deviceID, userAgent)
}

func (s *Service) LogLoginFailed(ctx context.Context, userID, ip, deviceID, userAgent, detail string) {
	s.LogWithDetail(ctx, domain.AuditLoginFailed, userID, ip, deviceID, userAgent, detail)
}

func (s *Service) LogRegister(ctx context.Context, userID, ip, deviceID, userAgent string) {
	s.Log(ctx, domain.AuditRegister, userID, ip, deviceID, userAgent)
}

func (s *Service) LogLogout(ctx context.Context, userID string) {
	s.Log(ctx, domain.AuditLogout, userID, "", "", "")
}

func (s *Service) LogTokenRefresh(ctx context.Context, userID, ip, deviceID, userAgent string) {
	s.Log(ctx, domain.AuditTokenRefresh, userID, ip, deviceID, userAgent)
}

func (s *Service) LogSuspicious(ctx context.Context, userID, ip, detail string) {
	s.LogWithDetail(ctx, domain.AuditSuspicious, userID, ip, "", "", detail)
}

func (s *Service) LogSessionRevoke(ctx context.Context, userID, ip string) {
	s.Log(ctx, domain.AuditSessionRevoke, userID, ip, "", "")
}

func (s *Service) LogRoleChange(ctx context.Context, userID, ip string) {
	s.Log(ctx, domain.AuditRoleChange, userID, ip, "", "")
}

func (s *Service) FindByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.AuditLog, error) {
	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

func (s *Service) FindByAction(ctx context.Context, action domain.AuditAction, limit, offset int) ([]*domain.AuditLog, error) {
	return s.repo.FindByAction(ctx, action, limit, offset)
}
