package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
	"go.uber.org/zap"
)

type SessionValidator struct {
	redis  *redis.Client
	ttl    time.Duration
	prefix string
}

type SessionData struct {
	UserID    string            `json:"user_id"`
	Roles     []string          `json:"roles"`
	DeviceID  string            `json:"device_id"`
	SessionID string            `json:"session_id"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func NewSessionValidator(rdb *redis.Client, ttl time.Duration) *SessionValidator {
	return &SessionValidator{
		redis:  rdb,
		ttl:    ttl,
		prefix: "session:",
	}
}

func (v *SessionValidator) ValidateSession(ctx context.Context, tokenClaims *Claims) (*SessionData, error) {
	if v.redis == nil {
		return v.createEphemeralSession(tokenClaims), nil
	}

	sessionKey := fmt.Sprintf("%s%s", v.prefix, tokenClaims.SessionID)

	sessionJSON, err := v.redis.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("session lookup failed: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("invalid session data: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		v.redis.Del(ctx, sessionKey)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (v *SessionValidator) CreateSession(ctx context.Context, session *SessionData) error {
	if v.redis == nil {
		return nil
	}

	sessionKey := fmt.Sprintf("%s%s", v.prefix, session.SessionID)
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = v.ttl
	}

	return v.redis.Set(ctx, sessionKey, sessionJSON, ttl).Err()
}

func (v *SessionValidator) RevokeSession(ctx context.Context, sessionID string) error {
	if v.redis == nil {
		return nil
	}
	return v.redis.Del(ctx, fmt.Sprintf("%s%s", v.prefix, sessionID)).Err()
}

func (v *SessionValidator) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if v.redis == nil {
		return nil
	}
	pattern := fmt.Sprintf("%s*:%s", v.prefix, userID)
	iter := v.redis.Scan(ctx, 0, pattern, 100).Iterator()
	batchSize := 100
	keys := make([]string, 0, batchSize)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			if err := v.redis.Del(ctx, keys...).Err(); err != nil {
				observability.GetLogger().Error("failed to batch delete sessions", zap.Error(err))
			}
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		if err := v.redis.Del(ctx, keys...).Err(); err != nil {
			observability.GetLogger().Error("failed to batch delete sessions", zap.Error(err))
		}
	}
	return iter.Err()
}

func (v *SessionValidator) createEphemeralSession(claims *Claims) *SessionData {
	return &SessionData{
		UserID:    claims.UserID,
		Roles:     claims.Roles,
		SessionID: claims.SessionID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(v.ttl),
	}
}

func SessionMiddleware(validator *SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, hasUser := c.Get(string(middleware.UserIDKey))
		if !hasUser {
			c.Next()
			return
		}

		// [SECURITY] Require session ID for all authenticated requests
		// Do not allow bypass when X-Session-ID is missing
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error_code": "SESSION_REQUIRED",
				"message":    "X-Session-ID header is required for authenticated requests",
			})
			return
		}

		session, err := validator.ValidateSession(c.Request.Context(), &Claims{
			UserID:    fmt.Sprintf("%v", userID),
			SessionID: sessionID,
		})
		if err != nil {
			observability.LogWithTrace(c.Request.Context()).Warn("session validation failed",
				zap.Error(err),
			)
			c.AbortWithStatusJSON(401, gin.H{
				"error_code": "SESSION_INVALID",
				"message":    "session expired or invalid",
			})
			return
		}

		c.Set("session_data", session)
		c.Next()
	}
}
