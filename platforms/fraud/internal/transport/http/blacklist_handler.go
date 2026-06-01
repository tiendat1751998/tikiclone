package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud/internal/blacklist"
)

func (h *Handler) CheckBlacklist(c *gin.Context) {
	var req blacklist.CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.blacklistSvc.Check(c.Request.Context(), &req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type addBlacklistRequest struct {
	Type      string  `json:"type" binding:"required"`
	Value     string  `json:"value" binding:"required"`
	Reason    string  `json:"reason" binding:"required"`
	AddedBy   string  `json:"added_by"`
	TTLMinutes int    `json:"ttl_minutes"`
}

func (h *Handler) AddToBlacklist(c *gin.Context) {
	var req addBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := &blacklist.BlacklistEntry{
		Type:      blacklist.BlacklistType(req.Type),
		Value:     req.Value,
		Reason:    blacklist.BlacklistReason(req.Reason),
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if req.TTLMinutes > 0 {
		exp := time.Now().Add(time.Duration(req.TTLMinutes) * time.Minute)
		entry.ExpiresAt = &exp
	}

	if err := h.blacklistSvc.Add(c.Request.Context(), entry); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

type removeBlacklistRequest struct {
	Type  string `json:"type" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func (h *Handler) RemoveFromBlacklist(c *gin.Context) {
	var req removeBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.blacklistSvc.RemoveByValue(c.Request.Context(), blacklist.BlacklistType(req.Type), req.Value); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
