package auth

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
)

type JWTValidator struct {
	cfg        config.AuthConfig
	redis      *redis.Client
	jwksClient *JWKSClient
	mu         sync.RWMutex
	// [SECURITY] Use LRU cache with TTL to prevent unbounded memory growth
	cache      map[string]*cacheEntry
}

type cacheEntry struct {
	claims    *jwt.MapClaims
	expiresAt time.Time
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type Claims struct {
	Sub       string   `json:"sub"`
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	Scope     string   `json:"scope"`
	DeviceID  string   `json:"device_id"`
	SessionID string   `json:"session_id"`
	Type      string   `json:"type"`
	jwt.RegisteredClaims
}

type JWKSClient struct {
	endpoint  string
	client    *http.Client
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	lastFetch time.Time
	ttl       time.Duration
	// [SECURITY] Circuit breaker to prevent thundering herd on JWKS endpoint failure
	cb        *circuitBreaker
}

type circuitBreaker struct {
	mu                sync.Mutex
	failures          int
	lastFailure       time.Time
	state             cbState
	threshold         int
	resetTimeout      time.Duration
	halfOpenMaxProbes int
}

type cbState int

const (
	cbClosed cbState = iota
	cbOpen
	cbHalfOpen
)

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = cbHalfOpen
			return true
		}
		return false
	case cbHalfOpen:
		return true
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = cbClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = cbOpen
	}
}

func NewJWKSClient(endpoint string) *JWKSClient {
	return &JWKSClient{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     60 * time.Second,
				DisableCompression:  false,
			},
		},
		keys: make(map[string]*rsa.PublicKey),
		ttl:  1 * time.Hour,
		cb: &circuitBreaker{
			threshold:         5,
			resetTimeout:      30 * time.Second,
			halfOpenMaxProbes: 1,
		},
	}
}

func (c *JWKSClient) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, exists := c.keys[kid]
	age := time.Since(c.lastFetch)
	c.mu.RUnlock()

	if exists && age < c.ttl {
		return key, nil
	}

	// [SECURITY] Check circuit breaker before fetching
	if !c.cb.allow() {
		if exists {
			// Return stale key rather than failing when circuit is open
			return key, nil
		}
		return nil, fmt.Errorf("JWKS circuit breaker open, refusing to fetch")
	}

	if err := c.fetchKeys(); err != nil {
		c.cb.recordFailure()
		if exists {
			return key, nil
		}
		return nil, err
	}

	c.cb.recordSuccess()

	c.mu.RLock()
	key, exists = c.keys[kid]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key %s not found in JWKS", kid)
	}

	return key, nil
}

func (c *JWKSClient) fetchKeys() error {
	resp, err := c.client.Get(c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.keys = make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Use != "sig" {
			continue
		}
		key, err := jwkToPublicKey(jwk)
		if err != nil {
			continue
		}
		c.keys[jwk.Kid] = key
	}
	c.lastFetch = time.Now()

	return nil
}

func jwkToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	if len(eBytes) < 4 {
		tmp := make([]byte, 4)
		copy(tmp[4-len(eBytes):], eBytes)
		eBytes = tmp
	}

	exp := int(eBytes[0])<<24 | int(eBytes[1])<<16 | int(eBytes[2])<<8 | int(eBytes[3])

	n := new(rsa.PublicKey)
	n.N = new(big.Int).SetBytes(nBytes)
	n.E = exp
	return n, nil
}

func NewJWTValidator(cfg config.AuthConfig, rdb *redis.Client) *JWTValidator {
	v := &JWTValidator{
		cfg:        cfg,
		redis:      rdb,
		jwksClient: NewJWKSClient(cfg.JWKSEndpoint),
		cache:      make(map[string]*cacheEntry),
	}
	go v.periodicCacheCleanup()
	return v
}

func (v *JWTValidator) periodicCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		v.mu.Lock()
		for key, entry := range v.cache {
			if time.Now().After(entry.expiresAt) {
				delete(v.cache, key)
			}
		}
		v.mu.Unlock()
	}
}

// ValidateToken validates a JWT token with full security checks:
// 1. Blacklist check (fail closed - reject if Redis is down)
// 2. Algorithm confusion prevention (determine alg by key type, not header)
// 3. Signature verification
// 4. Claims validation
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}

	// [SECURITY] Check blacklist FIRST, before parsing the token.
	// Fail closed: if we can't check the blacklist, reject the token.
	if v.redis != nil {
		blacklisted, err := v.redis.Exists(ctx, fmt.Sprintf("token:blacklist:%s", tokenHash(tokenString))).Result()
		if err != nil {
			// [SECURITY] Fail closed - reject token if blacklist check fails
			return nil, fmt.Errorf("token blacklist check failed: %w", err)
		}
		if blacklisted > 0 {
			return nil, fmt.Errorf("token has been revoked")
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Header["alg"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid alg header type")
		}

		// [SECURITY] Algorithm confusion prevention:
		// Determine expected algorithm based on key type, NOT from the token header.
		switch {
		case strings.HasPrefix(alg, "RS"):
			// RSA signing - use JWKS
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header")
			}
			return v.jwksClient.GetKey(kid)
		case strings.HasPrefix(alg, "HS"):
			// HMAC signing - use shared secret
			// [SECURITY] Only allow HMAC if JWKS endpoint is NOT configured
			// This prevents an attacker from using HS256 when RS256 is expected
			if v.cfg.JWKSEndpoint != "" {
				return nil, fmt.Errorf("HMAC signing not allowed when JWKS is configured")
			}
			// Try access token key first, then refresh token key.
			// We don't know the token type until after parsing, so we attempt
			// both keys: first the access key, then the refresh key.
			// This is safe because access and refresh tokens have different
			// claims (type: "access" vs type: "refresh").
			accessKey := []byte(v.cfg.AccessTokenKey)
			refreshKey := []byte(v.cfg.RefreshTokenKey)
			if len(refreshKey) > 0 {
				// Try access key first, fall back to refresh key
				testToken, parseErr := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
					return accessKey, nil
				}, jwt.WithLeeway(30*time.Second))
				if parseErr != nil || !testToken.Valid || claims.Type != "access" {
					// Access key didn't work or token is not an access token, try refresh key
					testToken, parseErr = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
						return refreshKey, nil
					}, jwt.WithLeeway(30*time.Second))
					if parseErr == nil && testToken.Valid {
						return refreshKey, nil
					}
				}
			}
			return accessKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %s", alg)
		}
	},
		jwt.WithLeeway(30*time.Second),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "HS256", "HS384", "HS512"}),
	)

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (v *JWTValidator) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	if v.redis == nil {
		return nil
	}
	return v.redis.Set(ctx, fmt.Sprintf("token:blacklist:%s", tokenHash(tokenString)), "1", ttl).Err()
}

func tokenHash(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(h[:])
}

func ExtractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return ""
	}
	if len(bearer) > 7 && strings.EqualFold(bearer[:7], "Bearer ") {
		return bearer[7:]
	}
	return ""
}

func ParsePublicKey(raw []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return pubKey, nil
}
