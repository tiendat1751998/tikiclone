package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/returns"
)

func (h *Handler) RequestReturn(c *gin.Context) {
	var req struct {
		ID     string            `json:"id"`
		OrderID string           `json:"order_id"`
		UserID string           `json:"user_id"`
		Items  []returns.ReturnItem `json:"items"`
		Reason string            `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ret := &returns.Return{
		ID:      req.ID,
		OrderID: req.OrderID,
		UserID:  req.UserID,
		Items:   req.Items,
		Reason:  req.Reason,
	}
	if err := h.returns.RequestReturn(c.Request.Context(), ret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ret)
}

func (h *Handler) ApproveReturn(c *gin.Context) {
	id := c.Param("id")
	if err := h.returns.ApproveReturn(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "return approved"})
}

func (h *Handler) ReceiveReturn(c *gin.Context) {
	id := c.Param("id")
	if err := h.returns.ReceiveReturn(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "return received"})
}
