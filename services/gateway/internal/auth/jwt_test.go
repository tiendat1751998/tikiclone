package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
)

func generateTestKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func createTestToken(privateKey *rsa.PrivateKey, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"
	return token.SignedString(privateKey)
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"valid bearer", "Bearer token123", "token123"},
		{"no bearer", "token123", ""},
		{"empty header", "", ""},
		{"malformed", "Bearer ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			result := ExtractToken(req)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTokenHash(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3QifQ.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"
	hash := tokenHash(token)
	if hash == "" {
		t.Error("hash should not be empty")
	}
	if len(hash) != 64 {
		t.Errorf("hash length expected 64, got %d", len(hash))
	}
}

func TestParsePublicKey(t *testing.T) {
	_, pubKey, err := generateTestKey()
	if err != nil {
		t.Fatal(err)
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ParsePublicKey(pubKeyBytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	if parsed.N == nil {
		t.Error("parsed key N should not be nil")
	}
}

func TestNewJWTValidator(t *testing.T) {
	cfg := config.AuthConfig{
		JWKSEndpoint: "http://localhost:8080/.well-known/jwks.json",
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   168 * time.Hour,
		EnableRBAC:   true,
	}

	validator := NewJWTValidator(cfg, nil)
	if validator == nil {
		t.Fatal("validator should not be nil")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	cfg := config.AuthConfig{JWKSEndpoint: "http://localhost:9999/.well-known/jwks.json"}
	validator := NewJWTValidator(cfg, nil)

	_, err := validator.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	privateKey, _, err := generateTestKey()
	if err != nil {
		t.Fatal(err)
	}

	claims := jwt.MapClaims{
		"sub":     "user123",
		"user_id": "user123",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	}

	token, err := createTestToken(privateKey, claims)
	if err != nil {
		t.Fatal(err)
	}

	validator := NewJWTValidator(config.AuthConfig{}, nil)
	_, err = validator.ValidateToken(context.Background(), token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}
