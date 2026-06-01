package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/auth/internal/application"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type Handler struct {
	authService *application.AuthService
}

func NewHandler(authService *application.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.register")
	defer span.End()

	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "INVALID_REQUEST",
			"message":    err.Error(),
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokens, _, err := h.authService.Register(ctx, &req, ip, userAgent)
	if err != nil {
		handleError(c, err)
		return
	}

	observability.LogWithTrace(ctx).Info("user registered",
		zap.String("session_id", tokens.SessionID),
	)

	c.JSON(http.StatusCreated, tokens)
}

func (h *Handler) Login(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.login")
	defer span.End()

	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "INVALID_REQUEST",
			"message":    err.Error(),
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokens, _, err := h.authService.Login(ctx, &req, ip, userAgent)
	if err != nil {
		handleError(c, err)
		return
	}

	span.SetAttributes(attribute.String("session_id", tokens.SessionID))
	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.refresh")
	defer span.End()

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "MISSING_REFRESH_TOKEN",
			"message":    "refresh_token is required",
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokens, _, err := h.authService.RefreshToken(ctx, req.RefreshToken, ip, userAgent)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.logout")
	defer span.End()

	accessToken := extractToken(c)
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}

	err := h.authService.Logout(ctx, accessToken, refreshToken, false)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *Handler) LogoutAll(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.logout_all")
	defer span.End()

	accessToken := extractToken(c)

	err := h.authService.Logout(ctx, accessToken, "", true)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out from all devices"})
}

func (h *Handler) GetSessions(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.sessions")
	defer span.End()

	accessToken := extractToken(c)
	claims, err := h.authService.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		handleError(c, err)
		return
	}

	sessions, err := h.authService.GetActiveSessions(ctx, claims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (h *Handler) RevokeSession(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.revoke_session")
	defer span.End()

	sessionID := c.Param("session_id")
	accessToken := extractToken(c)

	claims, err := h.authService.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		handleError(c, err)
		return
	}

	err = h.authService.RevokeSession(ctx, claims.UserID, sessionID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session revoked"})
}

func (h *Handler) GetProfile(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.profile")
	defer span.End()

	accessToken := extractToken(c)
	claims, err := h.authService.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		handleError(c, err)
		return
	}

	user, err := h.authService.GetUser(ctx, claims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"username":       user.Username,
		"display_name":   user.DisplayName,
		"phone":          user.Phone,
		"status":         user.Status,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
	})
}

func (h *Handler) ValidateToken(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.validate_token")
	defer span.End()

	tokenString := extractToken(c)
	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error_code": "MISSING_TOKEN",
			"message":    "authorization header required",
		})
		return
	}

	claims, err := h.authService.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error_code": "INVALID_TOKEN",
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"user_id":    claims.UserID,
		"email":      claims.Email,
		"roles":      claims.Roles,
		"session_id": claims.SessionID,
	})
}

func (h *Handler) RequestPasswordReset(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.request_password_reset")
	defer span.End()

	var req domain.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "INVALID_REQUEST",
			"message":    "valid email is required",
		})
		return
	}

	ip := c.ClientIP()
	if err := h.authService.RequestPasswordReset(ctx, &req, ip); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset email sent if account exists"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.reset_password")
	defer span.End()

	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "INVALID_REQUEST",
			"message":    "token, new_password, and confirm_password are required",
		})
		return
	}

	ip := c.ClientIP()
	if err := h.authService.ResetPassword(ctx, &req, ip); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.verify_email")
	defer span.End()

	var req domain.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error_code": "INVALID_REQUEST",
			"message":    "token is required",
		})
		return
	}

	ip := c.ClientIP()
	if err := h.authService.VerifyEmail(ctx, &req, ip); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

func (h *Handler) SendVerificationEmail(c *gin.Context) {
	ctx, span := otel.Tracer("shopee-auth").Start(c.Request.Context(), "http.send_verification_email")
	defer span.End()

	accessToken := extractToken(c)
	claims, err := h.authService.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		handleError(c, err)
		return
	}

	if err := h.authService.SendVerificationEmail(ctx, claims.UserID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification email sent if not already verified"})
}

var errorStatusMap = map[error]int{
	domain.ErrInvalidCredentials:  http.StatusUnauthorized,
	domain.ErrInvalidToken:        http.StatusUnauthorized,
	domain.ErrTokenExpired:        http.StatusUnauthorized,
	domain.ErrInvalidRefreshToken: http.StatusUnauthorized,
	domain.ErrRefreshReuse:        http.StatusUnauthorized,
	domain.ErrTokenBlacklisted:    http.StatusUnauthorized,
	domain.ErrAccountLocked:       http.StatusTooManyRequests,
	domain.ErrAccountInactive:     http.StatusForbidden,
	domain.ErrAccountSuspended:    http.StatusForbidden,
	domain.ErrEmailNotVerified:    http.StatusForbidden,
	domain.ErrRateLimited:         http.StatusTooManyRequests,
	domain.ErrEmailAlreadyExists:  http.StatusConflict,
	domain.ErrUsernameTaken:       http.StatusConflict,
	domain.ErrPasswordTooWeak:     http.StatusUnprocessableEntity,
	domain.ErrSessionExpired:      http.StatusUnauthorized,
	domain.ErrSessionRevoked:      http.StatusUnauthorized,
	domain.ErrInsufficientPerms:   http.StatusForbidden,
	domain.ErrMaxSessions:         http.StatusConflict,
	domain.ErrUserNotFound:        http.StatusNotFound,
	domain.ErrInvalidResetToken:   http.StatusBadRequest,
	domain.ErrInvalidVerifyToken:  http.StatusBadRequest,
	domain.ErrPasswordMismatch:    http.StatusUnprocessableEntity,
}

func handleError(c *gin.Context, err error) {
	errCode := "INTERNAL_ERROR"
	status := http.StatusInternalServerError

	for domainErr, httpStatus := range errorStatusMap {
		if errors.Is(err, domainErr) {
			status = httpStatus
			errCode = domainErr.Error()
			break
		}
	}

	if status == http.StatusInternalServerError {
		zap.L().Error("unhandled auth error", zap.Error(err))
	}

	c.AbortWithStatusJSON(status, gin.H{
		"error_code": errCode,
		"message":    err.Error(),
	})
}

func extractToken(c *gin.Context) string {
	bearer := c.GetHeader("Authorization")
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		return bearer[7:]
	}
	return ""
}
