package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/platforms/live-scale/internal/websocket_cluster"
)

func (h *Handler) RegisterWSNode(c *gin.Context) {
	var req struct {
		ID         string `json:"id" binding:"required"`
		Address    string `json:"address" binding:"required"`
		Region     string `json:"region" binding:"required"`
		MaxRooms   int    `json:"max_rooms"`
		MaxClients int    `json:"max_clients"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.MaxRooms <= 0 {
		req.MaxRooms = 100
	}
	if req.MaxClients <= 0 {
		req.MaxClients = 1000
	}
	node := &websocket_cluster.WSNode{
		ID:         req.ID,
		Address:    req.Address,
		Region:     req.Region,
		MaxRooms:   req.MaxRooms,
		MaxClients: req.MaxClients,
	}
	if err := h.cluster.RegisterNode(c.Request.Context(), node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, node)
}

func (h *Handler) AssignRoom(c *gin.Context) {
	var req struct {
		RoomID          string `json:"room_id" binding:"required"`
		PreferredNodeID string `json:"preferred_node_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assignment, err := h.cluster.AssignRoom(c.Request.Context(), req.RoomID, req.PreferredNodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *Handler) BroadcastMessage(c *gin.Context) {
	var req struct {
		RoomID  string `json:"room_id" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delivered, err := h.cluster.BroadcastAcrossCluster(c.Request.Context(), req.RoomID, []byte(req.Message))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"delivered_to": delivered})
}
