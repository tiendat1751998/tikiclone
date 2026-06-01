package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/tikiclone/tiki/services/shipment/internal/infrastructure/websocket"
)

type WSHandler struct {
	wsManager *websocket.Manager
}

func NewWSHandler(wsManager *websocket.Manager) *WSHandler {
	return &WSHandler{wsManager: wsManager}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	h.wsManager.HandleWS(c)
}

func (h *WSHandler) GetStats(c *gin.Context) {
	c.JSON(200, h.wsManager.GetStats())
}
