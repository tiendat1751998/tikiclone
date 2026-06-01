package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/auth/internal/config"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Claims struct {
	UserID    string       `json:"user_id"`
	Email     string       `json:"email,omitempty"`
	Roles     []domain.Role `json:"roles,omitempty"`
	SessionID string       `json:"session_id,omitempty"`
	DeviceID  string       `json:"device_id,omitempty"`
	TokenID   string       `json:"token_id"`
	Type      string       `json:"type"`
	jwt.RegisteredClaims
}

type Service struct {
	cfg        config.JWTConfig
	redisStore redisStore
}

type redisStore interface {
	IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error)
	IsRefreshTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

func NewService(cfg config.JWTConfig, store redisStore) *Service {
	return &Service{
		cfg:        cfg,
		redisStore: store,
	}
}

func (s *Service) GenerateAccessToken(ctx context.Context, claims *Claims) (string, string, error) {
	ctx, span := otel.Tracer("tiki-auth").Start(ctx, "jwt.generate_access")
	defer span.End()

	tokenID := uuid.New().String()
	now := time.Now()

	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID:        tokenID,
		Subject:   claims.UserID,
		Issuer:    s.cfg.Issuer,
		Audience:  jwt.ClaimStrings{s.cfg.Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
		NotBefore: jwt.NewNumericDate(now.Add(-s.cfg.ClockSkew)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.AccessSecret)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", "", fmt.Errorf("access token signing: %w", err)
	}

	span.SetAttributes(
		attribute.String("token_id", tokenID),
		attribute.String("user_id", claims.UserID),
	)

	return tokenString, tokenID, nil
}

func (s *Service) GenerateRefreshToken(ctx context.Context, userID string) (string, string, error) {
	token, err := s.GenerateRefreshTokenWithSession(ctx, userID, "", uuid.New().String())
	return token, "", err
}

func (s *Service) GenerateRefreshTokenWithSession(ctx context.Context, userID string, sessionID string, tokenID string) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:    userID,
		TokenID:   tokenID,
		SessionID: sessionID,
		Type:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.RefreshSecret)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("refresh token signing: %w", err)
	}

	return tokenString, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	ctx, span := otel.Tracer("tiki-auth").Start(ctx, "jwt.validate_access")
	defer span.End()

	if s.redisStore != nil && s.cfg.BlacklistEnabled {
		blacklisted, err := s.redisStore.IsAccessTokenBlacklisted(ctx, tokenString)
		if err == nil && blacklisted {
			return nil, domain.ErrTokenBlacklisted
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.AccessSecret), nil
	},
		jwt.WithLeeway(s.cfg.ClockSkew),
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithAudience(s.cfg.Audience),
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("access token validation: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.Type != "access" {
		return nil, domain.ErrInvalidToken
	}

	span.SetAttributes(
		attribute.String("user_id", claims.UserID),
		attribute.String("token_id", claims.TokenID),
	)

	return claims, nil
}

func (s *Service) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	ctx, span := otel.Tracer("tiki-auth").Start(ctx, "jwt.validate_refresh")
	defer span.End()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.RefreshSecret), nil
	},
		jwt.WithLeeway(s.cfg.ClockSkew),
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("refresh token validation: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidRefreshToken
	}

	if claims.Type != "refresh" {
		return nil, domain.ErrInvalidRefreshToken
	}

	if s.redisStore != nil && s.cfg.BlacklistEnabled {
		blacklisted, err := s.redisStore.IsRefreshTokenBlacklisted(ctx, claims.TokenID)
		if err == nil && blacklisted {
			return nil, domain.ErrTokenBlacklisted
		}
	}

	span.SetAttributes(
		attribute.String("user_id", claims.UserID),
		attribute.String("token_id", claims.TokenID),
	)

	return claims, nil
}

func (s *Service) ParseTokenUnsafe(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
