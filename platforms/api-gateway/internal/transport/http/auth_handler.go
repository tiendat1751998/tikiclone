package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/api-gateway/internal/auth"
)

type createAPIKeyRequest struct {
	Service string `json:"service" binding:"required"`
}

type validateAPIKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

type signJWTRequest struct {
	Subject   string            `json:"sub" binding:"required"`
	Issuer    string            `json:"iss"`
	ExpiresAt int64             `json:"exp"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type verifyJWTRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.APIKeyValidator.Create(req.Service)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, key)
}

func (h *Handler) ValidateAPIKey(c *gin.Context) {
	var req validateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := h.APIKeyValidator.Validate(req.Key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, key)
}

func (h *Handler) SignJWT(c *gin.Context) {
	var req signJWTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := auth.JWTClaims{
		Subject:   req.Subject,
		Issuer:    req.Issuer,
		ExpiresAt: req.ExpiresAt,
		Extra:     req.Extra,
	}

	token, err := h.JWTHandler.Sign(claims)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) VerifyJWT(c *gin.Context) {
	var req verifyJWTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.JWTHandler.Verify(req.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, claims)
}
