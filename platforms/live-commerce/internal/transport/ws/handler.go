package ws

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	ws "github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsHandler struct {
	hub     *ws.Hub
	service LiveCommerceService
}

type LiveCommerceService interface {
	ViewerJoined(ctx context.Context, roomID, userID string)
	ViewerLeft(ctx context.Context, roomID, userID string)
}

func NewWsHandler(hub *ws.Hub, svc LiveCommerceService) *WsHandler {
	return &WsHandler{hub: hub, service: svc}
}

func (h *WsHandler) HandleWebSocket(c *gin.Context) {
	roomID := c.Param("room_id")
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		observability.GetLogger().Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := ws.NewClient(h.hub, conn, roomID, userID)
	h.hub.Register(client)

	h.service.ViewerJoined(c.Request.Context(), roomID, userID)

	go client.WritePump()
	go client.ReadPump()

	observability.GetLogger().Info("websocket connected",
		zap.String("user_id", userID),
		zap.String("room_id", roomID))
}
