package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/identity-auth/internal/application"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type Handler struct {
	authService *application.AuthService
}

func NewHandler(authService *application.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(c *gin.Context) {
	var req domain.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error_code": "VALIDATION_ERROR",
			"message":    err.Error(),
			"details": []gin.H{
				{"field": "email", "issue": "Invalid email format"},
			},
		})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err == domain.ErrEmailAlreadyExists {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error_code": "DUPLICATE_RESOURCE",
				"message":    err.Error(),
			})
			return
		}
		if err == domain.ErrPhoneAlreadyExists {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error_code": "DUPLICATE_RESOURCE",
				"message":    err.Error(),
			})
			return
		}
		if err == domain.ErrPasswordTooWeak {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error_code": "VALIDATION_ERROR",
				"message":    err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error_code": "INTERNAL_ERROR",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error_code": "INVALID_CREDENTIALS",
			"message":    err.Error(),
		})
		return
	}

	ip := clientIP(c)
	resp, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		if err == domain.ErrRateLimited {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error_code": "INVALID_CREDENTIALS",
				"message":    err.Error(),
			})
			return
		}
		if err == domain.ErrAccountLocked || err == domain.ErrAccountDisabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error_code": "FORBIDDEN",
				"message":    err.Error(),
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error_code": "INVALID_CREDENTIALS",
			"message":    "Invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error_code": "VALIDATION_ERROR",
			"message":    "Refresh token is required",
		})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == domain.ErrTokenNotFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "INVALID_CREDENTIALS",
				"message":    "Invalid or expired refresh token",
			})
			return
		}
		if err == domain.ErrTokenRevoked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "INVALID_CREDENTIALS",
				"message":    "Refresh token has been revoked",
			})
			return
		}
		if err == domain.ErrTokenExpired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "INVALID_CREDENTIALS",
				"message":    "Refresh token has expired",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error_code": "INVALID_CREDENTIALS",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	var req domain.RefreshRequest
	_ = c.ShouldBindJSON(&req)

	userID := c.GetString("user_id")
	_ = h.authService.Logout(c.Request.Context(), userID, req.RefreshToken)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *Handler) ValidateToken(c *gin.Context) {
	var req domain.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Token is required",
		})
		return
	}

	valid, err := h.authService.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error_code": "USER_NOT_FOUND",
			"message":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUserByID(c *gin.Context) {
	userID := c.Param("userId")
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error_code": "USER_NOT_FOUND",
			"message":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

func clientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" || ip == "unknown" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" || ip == "unknown" {
		ip = c.ClientIP()
	}
	if ip != "" {
		for i, ch := range ip {
			if ch == ',' {
				return ip[:i]
			}
		}
	}
	return ip
}