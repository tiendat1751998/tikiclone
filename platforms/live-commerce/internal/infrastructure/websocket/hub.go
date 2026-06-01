package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/metrics"
	"go.uber.org/zap"
)

type Message struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id"`
	Payload json.RawMessage `json:"payload"`
	Sender  string          `json:"sender,omitempty"`
}

type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *Message, 1024),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			metrics.ConnectionsActive.WithLabelValues(client.RoomID).Inc()
			h.mu.Unlock()
			observability.GetLogger().Debug("client joined room",
				zap.String("user_id", client.UserID),
				zap.String("room_id", client.RoomID))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					metrics.ConnectionsActive.WithLabelValues(client.RoomID).Dec()
					if len(clients) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.rooms[msg.RoomID]
			h.mu.RUnlock()
			if !ok {
				continue
			}
			data, err := json.Marshal(msg)
			if err != nil {
				observability.GetLogger().Error("marshal broadcast message", zap.Error(err))
				continue
			}
			h.mu.RLock()
			for client := range clients {
				if msg.Sender != "" && client.UserID == msg.Sender {
					continue
				}
				select {
				case client.Send <- data:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					delete(clients, client)
					close(client.Send)
					metrics.ConnectionsActive.WithLabelValues(client.RoomID).Dec()
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
			metrics.MessagesBroadcast.WithLabelValues(msg.Type).Inc()

		case <-h.done:
			return
		}
	}
}

func (h *Hub) Broadcast(ctx context.Context, msg *Message) {
	select {
	case h.broadcast <- msg:
	case <-ctx.Done():
	case <-time.After(time.Second):
		observability.GetLogger().Warn("broadcast channel full, dropping message",
			zap.String("type", msg.Type), zap.String("room", msg.RoomID))
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) RoomCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}

func (h *Hub) TotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.rooms {
		count += len(clients)
	}
	return count
}
