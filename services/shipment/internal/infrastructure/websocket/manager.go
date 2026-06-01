package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tikiclone/tiki/services/shipment/internal/domain/delivery"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure per environment in production
	},
}

// Client represents a single WebSocket connection
type Client struct {
	id       string
	conn     *websocket.Conn
	send     chan []byte
	rooms    map[string]bool
	manager  *Manager
	mu       sync.RWMutex
	lastPing time.Time
	userID   string
	userType string // "customer" or "driver"
}

// Manager manages all WebSocket connections and room subscriptions
type Manager struct {
	clients    map[string]*Client
	rooms      map[string]map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *RoomMessage
	logger     *zap.Logger
	maxConns   int
	mu         sync.RWMutex
}

// RoomMessage is a message broadcast to a room
type RoomMessage struct {
	RoomID  string
	Message []byte
}

func NewManager(logger *zap.Logger, maxConns int) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *RoomMessage, 1024),
		logger:     logger.Named("ws_manager"),
		maxConns:   maxConns,
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			if len(m.clients) >= m.maxConns {
				m.mu.Unlock()
				m.logger.Warn("max connections reached, rejecting client",
					zap.Int("max", m.maxConns),
				)
				close(client.send)
				client.conn.Close()
				continue
			}
			m.clients[client.id] = client
			m.mu.Unlock()
			m.logger.Info("client registered",
				zap.String("client_id", client.id),
				zap.String("user_type", client.userType),
				zap.String("user_id", client.userID),
				zap.Int("total_clients", len(m.clients)),
			)

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.id]; ok {
				delete(m.clients, client.id)
				// Remove from all rooms
				for roomID := range client.rooms {
					if room, exists := m.rooms[roomID]; exists {
						delete(room, client.id)
						if len(room) == 0 {
							delete(m.rooms, roomID)
						}
					}
				}
				close(client.send)
				client.conn.Close()
				m.logger.Info("client unregistered",
					zap.String("client_id", client.id),
					zap.Int("remaining_clients", len(m.clients)),
				)
			}
			m.mu.Unlock()

		case msg := <-m.broadcast:
			m.mu.RLock()
			if room, ok := m.rooms[msg.RoomID]; ok {
				for _, client := range room {
					select {
					case client.send <- msg.Message:
					default:
						// Client's send buffer is full, skip
						m.logger.Warn("client send buffer full, skipping",
							zap.String("client_id", client.id),
						)
					}
				}
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) JoinRoom(clientID, roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[clientID]
	if !ok {
		return
	}

	client.mu.Lock()
	client.rooms[roomID] = true
	client.mu.Unlock()

	if _, exists := m.rooms[roomID]; !exists {
		m.rooms[roomID] = make(map[string]*Client)
	}
	m.rooms[roomID][clientID] = client

	m.logger.Debug("client joined room",
		zap.String("client_id", clientID),
		zap.String("room_id", roomID),
	)
}

func (m *Manager) LeaveRoom(clientID, roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[clientID]; ok {
		client.mu.Lock()
		delete(client.rooms, roomID)
		client.mu.Unlock()
	}

	if room, exists := m.rooms[roomID]; exists {
		delete(room, clientID)
		if len(room) == 0 {
			delete(m.rooms, roomID)
		}
	}
}

// BroadcastDriverLocation sends location update to tracking room
func (m *Manager) BroadcastDriverLocation(driverID string, lat, lng float64) {
	payload := delivery.TrackingUpdate{
		DriverID:  driverID,
		Lat:       lat,
		Lng:       lng,
		Status:    "location_update",
		Timestamp: time.Now().UTC(),
	}
	msg := delivery.WSEvent{
		Type:    "driver:location:update",
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Error("failed to marshal location update", zap.Error(err))
		return
	}

	// Broadcast to driver's location room
	m.broadcast <- &RoomMessage{
		RoomID:  "driver:" + driverID,
		Message: data,
	}

	// Also broadcast to any active order tracking rooms for this driver
	m.mu.RLock()
	defer m.mu.RUnlock()
	for roomID, clients := range m.rooms {
		if len(roomID) > 7 && roomID[:7] == "order:_" {
			for _, c := range clients {
				if c.userType == "customer" {
					select {
					case c.send <- data:
					default:
					}
				}
			}
		}
	}
}

// BroadcastTrackingUpdate sends order tracking update
func (m *Manager) BroadcastTrackingUpdate(orderID, driverID string, lat, lng float64, status string) {
	payload := delivery.TrackingUpdate{
		OrderID:   orderID,
		DriverID:  driverID,
		Lat:       lat,
		Lng:       lng,
		Status:    status,
		Timestamp: time.Now().UTC(),
	}
	msg := delivery.WSEvent{
		Type:    "order:tracking:update",
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Error("failed to marshal tracking update", zap.Error(err))
		return
	}

	m.broadcast <- &RoomMessage{
		RoomID:  "order:" + orderID,
		Message: data,
	}
}

// BroadcastDriverAssigned sends driver assignment notification
func (m *Manager) BroadcastDriverAssigned(orderID, driverID string) {
	payload := delivery.DriverAssignedEvent{
		OrderID:   orderID,
		DriverID:  driverID,
		Status:    "driver_assigned",
		Timestamp: time.Now().UTC(),
	}
	msg := delivery.WSEvent{
		Type:    "driver:assigned",
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Error("failed to marshal driver assigned event", zap.Error(err))
		return
	}

	m.broadcast <- &RoomMessage{
		RoomID:  "order:" + orderID,
		Message: data,
	}
}

// HandleWS upgrades HTTP to WebSocket and manages the connection lifecycle
func (m *Manager) HandleWS(c *gin.Context) {
	userID := c.Query("user_id")
	userType := c.Query("user_type")
	roomID := c.Query("room_id")

	if userID == "" {
		userID = uuid.New().String()
	}
	if userType == "" {
		userType = "customer"
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		m.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		id:       uuid.New().String(),
		conn:     conn,
		send:     make(chan []byte, 256),
		rooms:    make(map[string]bool),
		manager:  m,
		lastPing: time.Now(),
		userID:   userID,
		userType: userType,
	}

	m.register <- client

	// Auto-join room if specified
	if roomID != "" {
		m.JoinRoom(client.id, roomID)
	}

	// Auto-join driver location room
	if userType == "driver" {
		m.JoinRoom(client.id, "driver:"+userID)
	}

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastPing = time.Now()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived,
			) {
				c.manager.logger.Warn("websocket unexpected close",
					zap.String("client_id", c.id),
					zap.Error(err),
				)
			}
			break
		}

		// Handle incoming messages (room subscriptions, etc.)
		var msg struct {
			Action string `json:"action"`
			Room   string `json:"room"`
		}
		if err := json.Unmarshal(message, &msg); err == nil {
			switch msg.Action {
			case "join":
				c.manager.JoinRoom(c.id, msg.Room)
			case "leave":
				c.manager.LeaveRoom(c.id, msg.Room)
			case "ping":
				// Client-side ping, respond with pong
				pong := delivery.WSEvent{Type: "pong", Payload: time.Now().Unix()}
				if data, err := json.Marshal(pong); err == nil {
					select {
					case c.send <- data:
					default:
					}
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch any pending messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetStats returns connection statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalRooms := len(m.rooms)
	roomSizes := make(map[string]int)
	for id, clients := range m.rooms {
		roomSizes[id] = len(clients)
	}

	return map[string]interface{}{
		"total_clients":  len(m.clients),
		"total_rooms":    totalRooms,
		"room_sizes":     roomSizes,
		"max_connections": m.maxConns,
	}
}
