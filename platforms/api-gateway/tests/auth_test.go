package tests

import (
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/api-gateway/internal/auth"
)

func TestJWTSignAndVerify(t *testing.T) {
	jwt := auth.NewJWTHandler("test-secret-key-12345")

	token, err := jwt.Sign(auth.JWTClaims{
		Subject:   "user-123",
		Issuer:    "api-gateway",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	claims, err := jwt.Verify(token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("expected subject user-123, got %s", claims.Subject)
	}
}

func TestJWTExpiredToken(t *testing.T) {
	jwt := auth.NewJWTHandler("test-secret-key")

	token, err := jwt.Sign(auth.JWTClaims{
		Subject:   "user-123",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	_, err = jwt.Verify(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	jwt := auth.NewJWTHandler("secret-1")
	jwt2 := auth.NewJWTHandler("secret-2")

	token, err := jwt.Sign(auth.JWTClaims{
		Subject:   "user-123",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	_, err = jwt2.Verify(token)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestJWTParseClaims(t *testing.T) {
	jwt := auth.NewJWTHandler("test-key")

	token, _ := jwt.Sign(auth.JWTClaims{
		Subject: "user-456",
		Issuer:  "test",
	})

	claims, err := jwt.ParseClaims(token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.Subject != "user-456" {
		t.Errorf("expected user-456, got %s", claims.Subject)
	}
}

func TestAPIKeyCreateAndValidate(t *testing.T) {
	store := auth.NewAPIKeyStore()
	validator := auth.NewAPIKeyValidator(store)

	key, err := validator.Create("payment-service")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if key.Service != "payment-service" {
		t.Errorf("expected payment-service, got %s", key.Service)
	}
	if !key.IsActive {
		t.Error("key should be active")
	}

	validated, err := validator.Validate(key.Key)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	if validated.Key != key.Key {
		t.Error("validated key mismatch")
	}
}

func TestAPIKeyInvalid(t *testing.T) {
	store := auth.NewAPIKeyStore()
	validator := auth.NewAPIKeyValidator(store)

	_, err := validator.Validate("nonexistent-key")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAPIKeyInactive(t *testing.T) {
	store := auth.NewAPIKeyStore()
	validator := auth.NewAPIKeyValidator(store)

	key, _ := validator.Create("test-service")
	key.IsActive = false
	store.Store(key)

	_, err := validator.Validate(key.Key)
	if err == nil {
		t.Error("expected error for inactive key")
	}
}

func TestAPIKeyCreateEmptyService(t *testing.T) {
	store := auth.NewAPIKeyStore()
	validator := auth.NewAPIKeyValidator(store)

	_, err := validator.Create("")
	if err == nil {
		t.Error("expected error for empty service")
	}
}

func TestKeyRateLimiter(t *testing.T) {
	krl := auth.NewKeyRateLimiter()

	if !krl.RateLimitByKey("test-key", 2, time.Minute) {
		t.Error("first request should be allowed")
	}
	if !krl.RateLimitByKey("test-key", 2, time.Minute) {
		t.Error("second request should be allowed")
	}
	if krl.RateLimitByKey("test-key", 2, time.Minute) {
		t.Error("third request should be denied")
	}
}

func TestKeyRateLimiterWindowReset(t *testing.T) {
	krl := auth.NewKeyRateLimiter()

	if !krl.RateLimitByKey("reset-key", 1, 50*time.Millisecond) {
		t.Error("first request should be allowed")
	}
	if krl.RateLimitByKey("reset-key", 1, 50*time.Millisecond) {
		t.Error("second request should be denied within window")
	}

	time.Sleep(60 * time.Millisecond)

	if !krl.RateLimitByKey("reset-key", 1, 50*time.Millisecond) {
		t.Error("request after window reset should be allowed")
	}
}
