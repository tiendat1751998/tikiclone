package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/identity-auth/internal/config"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
	"github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/jwt"
	"github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/mysql"
	rediss "github.com/tikiclone/tiki/services/identity-auth/internal/infrastructure/redis"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cfg          *config.Config
	userRepo       *mysql.UserRepository
	tokenRepo      *mysql.RefreshTokenRepository
	failedLoginRepo *mysql.FailedLoginAttemptRepository
	outboxRepo     *mysql.OutboxEventRepository
	roleRepo       *mysql.RoleRepository
	userRoleRepo   *mysql.UserRoleRepository
	jwtService     *jwt.Service
	redisStore     *rediss.Store
}

func NewAuthService(
	cfg *config.Config,
	userRepo *mysql.UserRepository,
	tokenRepo *mysql.RefreshTokenRepository,
	failedLoginRepo *mysql.FailedLoginAttemptRepository,
	outboxRepo *mysql.OutboxEventRepository,
	roleRepo *mysql.RoleRepository,
	userRoleRepo *mysql.UserRoleRepository,
	jwtService *jwt.Service,
	redisStore *rediss.Store,
) *AuthService {
	return &AuthService{
		cfg:     cfg,
		userRepo: userRepo,
		tokenRepo: tokenRepo,
		failedLoginRepo: failedLoginRepo,
		outboxRepo: outboxRepo,
		roleRepo: roleRepo,
		userRoleRepo: userRoleRepo,
		jwtService: jwtService,
		redisStore: redisStore,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.AuthRequest) (*domain.AuthResponse, error) {
	email := req.Email

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	if req.Phone != "" {
		phoneExists, err := s.userRepo.ExistsByPhone(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if phoneExists {
			return nil, domain.ErrPhoneAlreadyExists
		}
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed")
	}

	user := domain.NewUser(email, req.Phone, req.FullName, string(passwordHash))

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.userRoleRepo.AssignDefaultRole(ctx, user.ID.String()); err != nil {
		fmt.Printf("warning: failed to assign default role: %v\n", err)
	}

	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "user", user.ID.String(), "user.registered", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
	})

	return &domain.AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Phone:        user.Phone,
		FullName:     user.FullName,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest, ip string) (*domain.AuthResponse, error) {
	email := req.Email

	if s.redisStore != nil {
		allowed, err := s.redisStore.CheckLoginRate(ctx, email, s.cfg.RateLimit.LoginMaxAttempts, s.cfg.RateLimit.LoginWindow)
		if err == nil && !allowed {
			_ = s.failedLoginRepo.Create(ctx, &domain.FailedLoginAttempt{
				ID:         uuid.New(),
				Email:      email,
				IPAddress:  ip,
				AttemptedAt: time.Now(),
			})
			return nil, domain.ErrRateLimited
		}
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if s.redisStore != nil {
			_ = s.redisStore.RecordLoginAttempt(ctx, email, s.cfg.RateLimit.LoginWindow)
		}
		_ = s.failedLoginRepo.Create(ctx, &domain.FailedLoginAttempt{
			ID:         uuid.New(),
			Email:      email,
			IPAddress:  ip,
			AttemptedAt: time.Now(),
		})
		return nil, domain.ErrInvalidCredentials
	}

	if !verifyPassword(req.Password, user.PasswordHash) {
		if s.redisStore != nil {
			_ = s.redisStore.RecordLoginAttempt(ctx, email, s.cfg.RateLimit.LoginWindow)
			_ = s.redisStore.RecordIPAttempt(ctx, ip, s.cfg.RateLimit.IPWindow)
		}
		_ = s.failedLoginRepo.Create(ctx, &domain.FailedLoginAttempt{
			ID:         uuid.New(),
			Email:      email,
			IPAddress:  ip,
			AttemptedAt: time.Now(),
		})
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	if s.redisStore != nil {
		_ = s.redisStore.ResetLoginRate(ctx, email)
	}
	_ = s.failedLoginRepo.DeleteByEmail(ctx, email)

	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "user", user.ID.String(), "user.logged_in", map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"ip":      ip,
	})

	return &domain.AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Phone:        user.Phone,
		FullName:     user.FullName,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenNotFound
	}

	if token.Revoked {
		return nil, domain.ErrTokenRevoked
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	token.Revoked = true
	_ = s.tokenRepo.Create(ctx, &domain.RefreshToken{ID: token.ID, Token: "", UserID: token.UserID, ExpiresAt: time.Now(), Revoked: true})

	user, err := s.userRepo.FindByID(ctx, claims.Subject)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	accessToken, err := s.jwtService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	newToken := &domain.RefreshToken{
		ID:        uuid.New(),
		Token:     newRefreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, newToken); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		UserID:       user.ID,
		Email:        user.Email,
		Phone:        user.Phone,
		FullName:     user.FullName,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string, refreshToken string) error {
	if refreshToken != "" {
		_ = s.tokenRepo.Revoke(ctx, refreshToken)
	}
	s.publishEvent(ctx, "user", userID, "user.logged_out", map[string]interface{}{"user_id": userID})
	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (bool, error) {
	_, err := s.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		UserID:    user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		FullName:  user.FullName,
		Role:      user.Role,
		Verified:  user.IsVerified,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *AuthService) publishEvent(ctx context.Context, aggregateType, aggregateID, eventType string, payload map[string]interface{}) {
	event := &domain.OutboxEvent{
		EventID:        uuid.New(),
		AggregateType:  aggregateType,
		AggregateID:    aggregateID,
		EventType:      eventType,
		Payload:        serializePayload(payload),
		Processed:      false,
		RetryCount:     0,
		CreatedAt:      time.Now(),
	}
	_ = s.outboxRepo.Create(ctx, event)
}

func serializePayload(payload map[string]interface{}) string {
	data, _ := json.Marshal(payload)
	return string(data)
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooWeak
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