package jwt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/identity-auth/internal/config"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type Claims struct {
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
	Scope  string `json:"scope,omitempty"`
	Type   string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

type Service struct {
	cfg    config.JWTConfig
	keyID  string
	hasRSA bool
}

func NewService(cfg config.JWTConfig) *Service {
	keyID := uuid.New().String()
	hasRSA := cfg.RSAPrivateKey != "" && cfg.RSAPublicKey != ""
	return &Service{
		cfg:    cfg,
		keyID:  keyID,
		hasRSA: hasRSA,
	}
}

func (s *Service) GenerateAccessToken(ctx context.Context, user *domain.User) (string, error) {
	now := time.Now()
	expiry := now.Add(s.cfg.AccessTTL)

	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		Scope:  s.buildScope(user),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "tiki-clone",
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.AccessSecret)
	return token.SignedString(secret)
}

func (s *Service) GenerateRefreshToken(ctx context.Context, user *domain.User) (string, error) {
	now := time.Now()
	expiry := now.Add(s.cfg.RefreshTTL)

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "tiki-clone",
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
		Type: "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(s.cfg.RefreshSecret)
	return token.SignedString(secret)
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.AccessSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid access token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidRefreshToken
	}

	if claims.Type != "refresh" {
		return nil, domain.ErrInvalidRefreshToken
	}

	return claims, nil
}

func (s *Service) buildScope(user *domain.User) string {
	var scope []string

	switch user.Role {
	case "SELLER":
		scope = []string{"products:read", "products:write", "orders:read", "inventory:read", "inventory:write"}
	case "ADMIN", "SUPER_ADMIN":
		scope = []string{"admin:access", "users:read", "users:write", "products:read", "products:write", "orders:read", "payments:read"}
	default:
		scope = []string{"products:read", "orders:read", "orders:write"}
	}

	return strings.Join(scope, " ")
}