package application

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/hash"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/jwt"
	"github.com/tikiclone/tiki/services/auth/internal/infrastructure/mysql"
	redisinfra "github.com/tikiclone/tiki/services/auth/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/services/auth/internal/metrics"
	"github.com/tikiclone/tiki/services/auth/internal/security"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type AuthService struct {
	cfg         *config.Config
	userRepo    *mysql.UserRepository
	sessionRepo *mysql.SessionRepository
	auditRepo   *mysql.AuditRepository
	redisStore  *redisinfra.Store
	rateLimiter *security.RateLimiter
	suspicious  *security.SuspiciousDetector
	jwtService  *jwt.Service
	hashService *hash.Service
}

func NewAuthService(
	cfg *config.Config,
	userRepo *mysql.UserRepository,
	sessionRepo *mysql.SessionRepository,
	auditRepo *mysql.AuditRepository,
	redisStore *redisinfra.Store,
	rateLimiter *security.RateLimiter,
	suspicious *security.SuspiciousDetector,
	jwtService *jwt.Service,
	hashService *hash.Service,
) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		redisStore:  redisStore,
		rateLimiter: rateLimiter,
		suspicious:  suspicious,
		jwtService:  jwtService,
		hashService: hashService,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest, ip, userAgent string) (*domain.TokenPair, *domain.Session, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.register")
	defer span.End()

	if err := s.rateLimiter.CheckRegister(ctx, ip); err != nil {
		span.SetStatus(codes.Error, "rate limited")
		return nil, nil, err
	}

	if err := validatePassword(req.Password); err != nil {
		span.SetStatus(codes.Error, "weak password")
		return nil, nil, err
	}

	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		metrics.RegistrationErrors.WithLabelValues("email_exists").Inc()
		return nil, nil, domain.ErrEmailAlreadyExists
	}

	existingByUser, _ := s.userRepo.FindByUsername(ctx, req.Username)
	if existingByUser != nil {
		metrics.RegistrationErrors.WithLabelValues("username_taken").Inc()
		return nil, nil, domain.ErrUsernameTaken
	}

	passwordHash, err := s.hashService.Hash(ctx, req.Password)
	if err != nil {
		span.SetStatus(codes.Error, "hash failed")
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := domain.NewUser(req.Email, req.Username, passwordHash)
	user.DisplayName = req.DisplayName
	user.Phone = req.Phone

	if err := s.userRepo.Create(ctx, user); err != nil {
		span.SetStatus(codes.Error, "db create failed")
		return nil, nil, fmt.Errorf("user creation failed: %w", err)
	}

	if err := s.assignDefaultRole(ctx, user.ID, domain.RoleBuyer); err != nil {
		observability.LogWithTrace(ctx).Warn("failed to assign default role", zap.String("user_id", user.ID))
	}

	tokens, session, err := s.createSession(ctx, user, ip, userAgent, req.DeviceID)
	if err != nil {
		return nil, nil, err
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		user.ID, domain.AuditRegister, ip, req.DeviceID, userAgent,
	))

	metrics.RegistrationsTotal.Inc()
	span.SetAttributes(attribute.String("user_id", user.ID))

	return tokens, session, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest, ip, userAgent string) (*domain.TokenPair, *domain.Session, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.login")
	defer span.End()

	if err := s.rateLimiter.CheckLogin(ctx, req.Email, ip); err != nil {
		auditLog := domain.NewAuditLog(
			trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
			"", domain.AuditLoginFailed, ip, req.DeviceID, userAgent,
		)
		auditLog.Detail = "rate limited"
		s.auditRepo.Log(ctx, auditLog)
		metrics.FailedLogins.WithLabelValues("rate_limited").Inc()
		return nil, nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		auditLog := domain.NewAuditLog(
			trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
			"", domain.AuditLoginFailed, ip, req.DeviceID, userAgent,
		)
		auditLog.Detail = "user not found"
		s.auditRepo.Log(ctx, auditLog)
		metrics.FailedLogins.WithLabelValues("user_not_found").Inc()
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := user.CanLogin(); err != nil {
		auditLog := domain.NewAuditLog(
			trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
			user.ID, domain.AuditLoginFailed, ip, req.DeviceID, userAgent,
		)
		auditLog.Detail = err.Error()
		s.auditRepo.Log(ctx, auditLog)
		metrics.FailedLogins.WithLabelValues("account_blocked").Inc()
		return nil, nil, err
	}

	if !s.hashService.Verify(ctx, req.Password, user.PasswordHash) {
		user.RecordFailedAttempt()
		s.userRepo.Update(ctx, user)

		if user.FailedAttempts >= s.cfg.RateLimit.AccountLockout {
			user.Lock(s.cfg.RateLimit.LockoutDuration)
			s.userRepo.Update(ctx, user)
			metrics.AccountLockouts.Inc()
		}

		auditLog := domain.NewAuditLog(
			trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
			user.ID, domain.AuditLoginFailed, ip, req.DeviceID, userAgent,
		)
		auditLog.Detail = "invalid password"
		s.auditRepo.Log(ctx, auditLog)
		metrics.FailedLogins.WithLabelValues("wrong_password").Inc()
		return nil, nil, domain.ErrInvalidCredentials
	}

	if s.suspicious.IsSuspicious(ctx, user.ID, ip) {
		auditLog := domain.NewAuditLog(
			trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
			user.ID, domain.AuditSuspicious, ip, req.DeviceID, userAgent,
		)
		auditLog.Detail = "suspicious login from new location"
		s.auditRepo.Log(ctx, auditLog)
		metrics.SuspiciousLogins.Inc()
	}

	user.RecordSuccessfulLogin(ip)
	s.userRepo.Update(ctx, user)

	tokens, session, err := s.createSession(ctx, user, ip, userAgent, req.DeviceID)
	if err != nil {
		return nil, nil, err
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		user.ID, domain.AuditLogin, ip, req.DeviceID, userAgent,
	))

	metrics.LoginsTotal.Inc()
	span.SetAttributes(attribute.String("user_id", user.ID))

	return tokens, session, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*domain.TokenPair, *domain.Session, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.refresh")
	defer span.End()

	claims, err := s.jwtService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		metrics.TokenRefreshErrors.WithLabelValues("invalid_token").Inc()
		return nil, nil, domain.ErrInvalidRefreshToken
	}

	if s.cfg.JWT.BlacklistEnabled && s.cfg.Session.RefreshRotation {
		isReused, err := s.redisStore.IsRefreshTokenReused(ctx, claims.TokenID)
		if err == nil && isReused {
			s.handleTokenReuse(ctx, claims.UserID, claims.SessionID, ip)
			metrics.TokenRefreshErrors.WithLabelValues("reuse_detected").Inc()
			return nil, nil, domain.ErrRefreshReuse
		}
	}

	if s.cfg.JWT.BlacklistEnabled {
		if err := s.redisStore.BlacklistRefreshToken(ctx, claims.TokenID, s.cfg.JWT.RefreshTTL); err != nil {
			observability.LogWithTrace(ctx).Warn("failed to blacklist refresh token", zap.Error(err))
		}
	}

	session, err := s.sessionRepo.FindByID(ctx, claims.SessionID)
	if err != nil {
		metrics.TokenRefreshErrors.WithLabelValues("session_not_found").Inc()
		return nil, nil, domain.ErrSessionExpired
	}

	if session.IsExpired() || session.Status != domain.SessionActive {
		metrics.TokenRefreshErrors.WithLabelValues("session_expired").Inc()
		return nil, nil, domain.ErrSessionExpired
	}

	session.Touch()
	s.sessionRepo.Update(ctx, session)

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, domain.ErrUserNotFound
	}

	newTokens, _, err := s.createSession(ctx, user, ip, userAgent, claims.DeviceID)
	if err != nil {
		return nil, nil, err
	}

	if s.cfg.Session.RefreshRotation {
		if err := s.redisStore.MarkRefreshTokenUsed(ctx, claims.TokenID, s.cfg.JWT.RefreshTTL); err != nil {
			observability.LogWithTrace(ctx).Warn("failed to mark token as used", zap.Error(err))
		}
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		user.ID, domain.AuditTokenRefresh, ip, claims.DeviceID, userAgent,
	))

	metrics.TokenRefreshesTotal.Inc()
	span.SetAttributes(attribute.String("user_id", user.ID))

	return newTokens, session, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string, allDevices bool) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.logout")
	defer span.End()

	claims, err := s.jwtService.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if allDevices {
		sessions, err := s.sessionRepo.FindActiveByUserID(ctx, claims.UserID)
		if err != nil {
			observability.LogWithTrace(ctx).Error("failed to find active sessions for logout-all", zap.Error(err))
		} else {
			for _, session := range sessions {
				session.Revoke()
				if err := s.sessionRepo.Update(ctx, session); err != nil {
					observability.LogWithTrace(ctx).Error("failed to revoke session", zap.String("session_id", session.ID), zap.Error(err))
				}
				if s.cfg.JWT.BlacklistEnabled && s.redisStore != nil {
					if err := s.redisStore.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.cfg.JWT.RefreshTTL); err != nil {
						observability.LogWithTrace(ctx).Error("failed to blacklist refresh token", zap.Error(err))
					}
				}
			}
		}
		if s.redisStore != nil {
			s.redisStore.RevokeAllUserSessions(ctx, claims.UserID)
		}
	} else {
		session, err := s.sessionRepo.FindByID(ctx, claims.SessionID)
		if err != nil {
			observability.LogWithTrace(ctx).Error("failed to find session for logout", zap.Error(err))
		} else {
			session.Revoke()
			if err := s.sessionRepo.Update(ctx, session); err != nil {
				observability.LogWithTrace(ctx).Error("failed to revoke session", zap.String("session_id", session.ID), zap.Error(err))
			}
		}
		if s.cfg.JWT.BlacklistEnabled && s.redisStore != nil {
			s.redisStore.BlacklistAccessToken(ctx, accessToken, s.cfg.JWT.AccessTTL)
			s.redisStore.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.cfg.JWT.RefreshTTL)
		}
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		claims.UserID, domain.AuditLogout, "", "", "",
	))

	return nil
}

func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return domain.ErrSessionExpired
	}

	if session.UserID != userID {
		return domain.ErrInsufficientPerms
	}

	session.Revoke()
	s.sessionRepo.Update(ctx, session)

	if s.cfg.JWT.BlacklistEnabled {
		s.redisStore.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.cfg.JWT.RefreshTTL)
	}

	return nil
}

func (s *AuthService) GetActiveSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	return s.sessionRepo.FindActiveByUserID(ctx, userID)
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.Claims, error) {
	return s.jwtService.ValidateAccessToken(ctx, tokenString)
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User, ip, userAgent, deviceID string) (*domain.TokenPair, *domain.Session, error) {
	activeCount, err := s.sessionRepo.CountActiveByUserID(ctx, user.ID)
	if err == nil && activeCount >= s.cfg.Session.MaxSessionsPerUser {
		oldest, err := s.sessionRepo.FindOldestActiveByUserID(ctx, user.ID)
		if err == nil {
			oldest.Revoke()
			s.sessionRepo.Update(ctx, oldest)
		}
	}

	accessClaims := &jwt.Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Roles:     s.getUserRoles(ctx, user.ID),
		Type:      "access",
		SessionID: "",
	}

	accessToken, _, err := s.jwtService.GenerateAccessToken(ctx, accessClaims)
	if err != nil {
		return nil, nil, fmt.Errorf("access token generation failed: %w", err)
	}

	refreshToken, refreshTokenID, err := s.jwtService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh token generation failed: %w", err)
	}

	session := domain.NewSession(
		user.ID, ip, userAgent, refreshTokenID,
		s.cfg.Session.SessionTTL,
	)
	session.DeviceID = deviceID

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("session creation failed: %w", err)
	}

	accessClaims.SessionID = session.ID

	if s.redisStore != nil {
		s.redisStore.StoreSession(ctx, session)
	}

	tokens := &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
		TokenType:    "Bearer",
		SessionID:    session.ID,
	}

	return tokens, session, nil
}

func (s *AuthService) getUserRoles(ctx context.Context, userID string) []domain.Role {
	roles, err := s.userRepo.FindRolesByUserID(ctx, userID)
	if err != nil {
		return []domain.Role{domain.RoleBuyer}
	}
	return roles
}

func (s *AuthService) assignDefaultRole(ctx context.Context, userID string, role domain.Role) error {
	return s.userRepo.AssignRole(ctx, userID, string(role))
}

func (s *AuthService) handleTokenReuse(ctx context.Context, userID, sessionID, ip string) {
	sessions, _ := s.sessionRepo.FindActiveByUserID(ctx, userID)
	for _, session := range sessions {
		session.Revoke()
		s.sessionRepo.Update(ctx, session)
		if s.redisStore != nil {
			s.redisStore.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.cfg.JWT.RefreshTTL)
		}
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		userID, domain.AuditSuspicious, ip, "", "",
	))
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, req *domain.PasswordResetRequest, ip string) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.request_password_reset")
	defer span.End()

	if err := s.rateLimiter.CheckPasswordReset(ctx, req.Email); err != nil {
		span.SetStatus(codes.Error, "rate limited")
		return err
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	token := uuid.New().String()
	tokenHash := sha256Hex(token)
	if s.redisStore != nil {
		if err := s.redisStore.SetResetToken(ctx, user.ID, tokenHash, 15*time.Minute); err != nil {
			observability.LogWithTrace(ctx).Warn("failed to store reset token", zap.Error(err))
		}
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		user.ID, domain.AuditPasswordReset, ip, "", "",
	))

	metrics.PasswordResetsTotal.Inc()
	span.SetAttributes(attribute.String("user_id", user.ID))
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest, ip string) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.reset_password")
	defer span.End()

	if req.NewPassword != req.ConfirmPassword {
		return domain.ErrPasswordMismatch
	}

	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}

	tokenHash := sha256Hex(req.Token)

	userID, err := s.redisStore.ValidateAndConsumeResetToken(ctx, tokenHash)
	if err != nil {
		span.SetStatus(codes.Error, "invalid reset token")
		return domain.ErrInvalidResetToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	passwordHash, err := s.hashService.Hash(ctx, req.NewPassword)
	if err != nil {
		span.SetStatus(codes.Error, "hash failed")
		return fmt.Errorf("password hashing failed: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
		return fmt.Errorf("password update failed: %w", err)
	}

	sessions, _ := s.sessionRepo.FindActiveByUserID(ctx, user.ID)
	for _, session := range sessions {
		session.Revoke()
		s.sessionRepo.Update(ctx, session)
		if s.redisStore != nil {
			s.redisStore.BlacklistRefreshToken(ctx, session.RefreshTokenID, s.cfg.JWT.RefreshTTL)
		}
	}

	if s.redisStore != nil {
		s.redisStore.RevokeAllUserSessions(ctx, user.ID)
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		user.ID, domain.AuditPasswordChange, ip, "", "",
	))

	metrics.PasswordResetsTotal.Inc()
	span.SetAttributes(attribute.String("user_id", user.ID))
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *domain.VerifyEmailRequest, ip string) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "auth.verify_email")
	defer span.End()

	tokenHash := sha256Hex(req.Token)

	userID, err := s.redisStore.ValidateAndConsumeVerifyToken(ctx, tokenHash)
	if err != nil {
		span.SetStatus(codes.Error, "invalid verify token")
		return domain.ErrInvalidVerifyToken
	}

	if err := s.userRepo.UpdateEmailVerified(ctx, userID); err != nil {
		return fmt.Errorf("email verification update failed: %w", err)
	}

	s.auditRepo.Log(ctx, domain.NewAuditLog(
		trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
		userID, domain.AuditEmailVerify, ip, "", "",
	))

	metrics.EmailVerificationsTotal.Inc()
	span.SetAttributes(attribute.String("user_id", userID))
	return nil
}

func (s *AuthService) SendVerificationEmail(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if user.EmailVerified {
		return nil
	}

	token := uuid.New().String()
	tokenHash := sha256Hex(token)

	if s.redisStore != nil {
		if err := s.redisStore.SetVerifyToken(ctx, user.ID, tokenHash, 24*time.Hour); err != nil {
			return fmt.Errorf("failed to store verify token: %w", err)
		}
	}

	return nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooWeak
	}
	if len(password) > 128 {
		return domain.ErrPasswordTooWeak
	}
	// Accept pre-hashed passwords (SHA-256 hex: 64 lowercase hex chars)
	if len(password) == 64 {
		isHex := true
		for _, c := range password {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				isHex = false
				break
			}
		}
		if isHex {
			return nil
		}
	}
	hasUpper, hasLower, hasDigit := false, false, false
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return domain.ErrPasswordTooWeak
	}
	return nil
}

func constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
