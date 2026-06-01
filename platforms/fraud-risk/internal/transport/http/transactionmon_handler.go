package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/fraud-risk/internal/transactionmon"
)

type recordTransactionRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Location string  `json:"location"`
	IP       string  `json:"ip"`
	DeviceID string  `json:"device_id"`
}

func (h *Handler) RecordTransaction(c *gin.Context) {
	var req recordTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec := &transactionmon.TransactionRecord{
		UserID:   req.UserID,
		Amount:   req.Amount,
		Location: req.Location,
		IP:       req.IP,
		DeviceID: req.DeviceID,
	}

	mon, err := h.txnMon.RecordTransaction(c.Request.Context(), rec)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mon)
}

type checkTransactionRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Location string  `json:"location"`
	IP       string  `json:"ip"`
	DeviceID string  `json:"device_id"`
}

func (h *Handler) CheckTransaction(c *gin.Context) {
	var req checkTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rec := &transactionmon.TransactionRecord{
		UserID:   req.UserID,
		Amount:   req.Amount,
		Location: req.Location,
		IP:       req.IP,
		DeviceID: req.DeviceID,
	}

	result, err := h.txnMon.DetectAnomaly(c.Request.Context(), rec)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
