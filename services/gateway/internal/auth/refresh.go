package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tikiclone/tiki/services/gateway/internal/config"
	"github.com/tikiclone/tiki/services/gateway/internal/middleware"
)

type RefreshTokenHandler struct {
	validator *JWTValidator
	redis     *redis.Client
	cfg       config.AuthConfig
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func NewRefreshTokenHandler(validator *JWTValidator, rdb *redis.Client, cfg config.AuthConfig) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		validator: validator,
		redis:     rdb,
		cfg:       cfg,
	}
}

func (h *RefreshTokenHandler) HandleRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error_code": "MISSING_REFRESH_TOKEN",
				"message":    "refresh_token is required",
			})
			return
		}

		claims, err := h.validator.ValidateToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"error_code": "INVALID_REFRESH_TOKEN",
				"message":    "refresh token is invalid or expired",
			})
			return
		}

		refreshStoreKey := fmt.Sprintf("refresh:%s:%s", claims.UserID, claims.SessionID)
		if h.redis != nil {
			exists, err := h.redis.Exists(c.Request.Context(), refreshStoreKey).Result()
			if err != nil || exists == 0 {
				c.AbortWithStatusJSON(401, gin.H{
					"error_code": "REFRESH_TOKEN_REVOKED",
					"message":    "refresh token has been revoked",
				})
				return
			}
		}

		deviceInfo, exists := c.Get(string(middleware.DeviceInfoKey))
		if !exists || deviceInfo == nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error_code": "MISSING_DEVICE_INFO",
				"message":    "device info not found in context",
			})
			return
		}
		deviceMap, ok := deviceInfo.(map[string]string)
		if !ok {
			c.AbortWithStatusJSON(400, gin.H{
				"error_code": "INVALID_DEVICE_INFO",
				"message":    "invalid device info format",
			})
			return
		}

		newAccessToken, err := h.issueToken(claims.UserID, claims.Roles, claims.SessionID, h.cfg.AccessTTL, deviceMap)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"error_code": "TOKEN_ISSUE_FAILED",
				"message":    "failed to issue new access token",
			})
			return
		}

		c.JSON(200, TokenResponse{
			AccessToken:  newAccessToken,
			RefreshToken: req.RefreshToken,
			ExpiresIn:    int64(h.cfg.AccessTTL.Seconds()),
			TokenType:    "Bearer",
		})
	}
}

func (h *RefreshTokenHandler) HandleLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(string(middleware.UserIDKey))
		sessionID := c.GetHeader("X-Session-ID")

		if h.redis != nil && userID != nil && sessionID != "" {
			refreshKey := fmt.Sprintf("refresh:%s:%s", userID, sessionID)
			h.redis.Del(c.Request.Context(), refreshKey)
		}

		tokenString := ExtractToken(c.Request)
		if tokenString != "" && h.redis != nil {
			h.validator.BlacklistToken(c.Request.Context(), tokenString, h.cfg.TokenBlacklistTTL)
		}

		c.JSON(200, gin.H{
			"message": "logged out successfully",
		})
	}
}

func (h *RefreshTokenHandler) StoreRefreshToken(ctx context.Context, userID, sessionID, refreshToken string) error {
	if h.redis == nil {
		return nil
	}

	key := fmt.Sprintf("refresh:%s:%s", userID, sessionID)
	return h.redis.Set(ctx, key, refreshToken, h.cfg.RefreshTTL).Err()
}

func (h *RefreshTokenHandler) RevokeRefreshToken(ctx context.Context, userID, sessionID string) error {
	if h.redis == nil {
		return nil
	}

	key := fmt.Sprintf("refresh:%s:%s", userID, sessionID)
	return h.redis.Del(ctx, key).Err()
}

func (h *RefreshTokenHandler) issueToken(userID string, roles []string, sessionID string, ttl time.Duration, deviceInfo map[string]string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"roles":      roles,
		"session_id": sessionID,
		"sub":        userID,
		"iat":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
		"iss":        "tiki-gateway",
		"type":       "access",
	}

	if deviceInfo != nil {
		for k, v := range deviceInfo {
			if v != "" {
				claims[k] = v
			}
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(h.cfg.AccessTokenKey)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
