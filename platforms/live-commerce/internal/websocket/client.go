package websocket

import (
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

type Client struct {
	ID       string
	UserID   string
	Username string
	RoomID   string
	conn     *websocket.Conn
	hub      *Hub
	send     chan []byte
	mu       sync.Mutex
	closed   bool
}

func NewClient(id, userID, username string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:       id,
		UserID:   userID,
		Username: username,
		conn:     conn,
		hub:      hub,
		send:     make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				observability.LogWithTrace(nil).Warn("ws read error", zap.Error(err))
			}
			break
		}
		var msg map[string]interface{}
		if err := sonic.Unmarshal(message, &msg); err != nil {
			c.sendError("invalid_json")
			continue
		}
		c.hub.handleMessage(c, msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		case <-ticker.C:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) SendJSON(v interface{}) {
	data, err := sonic.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		observability.LogWithTrace(nil).Warn("client send buffer full", zap.String("client_id", c.ID))
	}
}

func (c *Client) sendError(code string) {
	c.SendJSON(map[string]string{"type": "error", "code": code})
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.send)
	}
}
