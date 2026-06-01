package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// bufferPool reduces allocations for WebSocket message serialization
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 512)
	},
}

type Hub struct {
	rooms      map[string]*Room
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	handlers   map[string]MessageHandler
	onMessage  func(ctx context.Context, client *Client, msg map[string]interface{})
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

type MessageHandler func(ctx context.Context, client *Client, payload map[string]interface{})

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Hub{
		rooms:      make(map[string]*Room),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
		handlers:   make(map[string]MessageHandler),
		logger:     observability.LogWithTrace(nil),
		ctx:        ctx,
		cancel:     cancel,
	}
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.RLock()
			room, exists := h.rooms[client.RoomID]
			h.mu.RUnlock()
			if !exists {
				h.mu.Lock()
				room = NewRoom(client.RoomID)
				h.rooms[client.RoomID] = room
				h.mu.Unlock()
			}
			room.Add(client)
			h.logger.Info("client joined room",
				zap.String("client_id", client.ID),
				zap.String("room_id", client.RoomID),
				zap.String("user_id", client.UserID))
		case client := <-h.Unregister:
			h.mu.RLock()
			room, exists := h.rooms[client.RoomID]
			h.mu.RUnlock()
			if exists {
				room.Remove(client.ID)
				if room.Count() == 0 {
					h.mu.Lock()
					if room.Count() == 0 {
						delete(h.rooms, client.RoomID)
					}
					h.mu.Unlock()
				}
			}
			client.Close()
			h.logger.Info("client left room",
				zap.String("client_id", client.ID),
				zap.String("room_id", client.RoomID))
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, roomID, userID, username string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}
	clientID := uuid.New().String()
	client := NewClient(clientID, userID, username, conn, h)
	client.RoomID = roomID
	h.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) BroadcastToRoom(roomID string, message interface{}, excludeID string) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()
	if exists {
		room.Broadcast(message, excludeID)
	}
}

func (h *Hub) BroadcastToRoomBytes(roomID string, data []byte, excludeID string) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()
	if exists {
		room.BroadcastBytes(data, excludeID)
	}
}

func (h *Hub) RoomCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, exists := h.rooms[roomID]
	if !exists {
		return 0
	}
	return room.Count()
}

func (h *Hub) handleMessage(client *Client, msg map[string]interface{}) {
	ctx := context.Background()
	msgType, ok := msg["type"].(string)
	if !ok {
		client.sendError("missing or invalid message type")
		return
	}
	if handler, ok := h.handlers[msgType]; ok {
		handler(ctx, client, msg)
	} else {
		client.sendError(fmt.Sprintf("unknown_message_type: %s", msgType))
	}
}

func (h *Hub) OnMessage(msgType string, handler MessageHandler) {
	h.handlers[msgType] = handler
}

type WSAuthMiddleware func(r *http.Request) (userID, username string, err error)

func WSAuthHandler(hub *Hub, auth WSAuthMiddleware, getRoomID func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, username, err := auth(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		roomID := getRoomID(r)
		hub.HandleWS(w, r, roomID, userID, username)
	}
}

// SonicMarshal uses sonic for faster JSON serialization (fewer allocs than encoding/json).
func SonicMarshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}
