package api

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// WSMessageType defines WebSocket message types
type WSMessageType string

const (
	WSMessageTypeStateUpdate  WSMessageType = "state_update"
	WSMessageTypeDroneOnline  WSMessageType = "drone_online"
	WSMessageTypeDroneOffline WSMessageType = "drone_offline"
	WSMessageTypeSubscribe    WSMessageType = "subscribe"
	WSMessageTypeUnsubscribe  WSMessageType = "unsubscribe"
	WSMessageTypeError        WSMessageType = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type     WSMessageType   `json:"type"`
	DeviceID string          `json:"device_id,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// WSClient represents a WebSocket client connection
type WSClient struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	subscribed  map[string]bool // subscribed device IDs, empty means all
	mu          sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WebSocket] Client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, close the connection
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastState sends a drone state update to all subscribed clients
func (h *Hub) BroadcastState(state *models.DroneState) {
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("[WebSocket] Failed to marshal state: %v", err)
		return
	}

	msg := WSMessage{
		Type:     WSMessageTypeStateUpdate,
		DeviceID: state.DeviceID,
		Data:     data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WebSocket] Failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	for client := range h.clients {
		if client.isSubscribed(state.DeviceID) {
			select {
			case client.send <- msgBytes:
			default:
				// Skip if buffer is full
			}
		}
	}
	h.mu.RUnlock()
}

// BroadcastDroneOnline notifies clients that a drone is online
func (h *Hub) BroadcastDroneOnline(deviceID string) {
	msg := WSMessage{
		Type:     WSMessageTypeDroneOnline,
		DeviceID: deviceID,
	}

	msgBytes, _ := json.Marshal(msg)
	h.broadcast <- msgBytes
}

// BroadcastDroneOffline notifies clients that a drone is offline
func (h *Hub) BroadcastDroneOffline(deviceID string) {
	msg := WSMessage{
		Type:     WSMessageTypeDroneOffline,
		DeviceID: deviceID,
	}

	msgBytes, _ := json.Marshal(msg)
	h.broadcast <- msgBytes
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// isSubscribed checks if the client is subscribed to a device
func (c *WSClient) isSubscribed(deviceID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// If no specific subscriptions, receive all
	if len(c.subscribed) == 0 {
		return true
	}
	return c.subscribed[deviceID]
}

// subscribe adds device IDs to the subscription list
func (c *WSClient) subscribe(deviceIDs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subscribed == nil {
		c.subscribed = make(map[string]bool)
	}
	for _, id := range deviceIDs {
		c.subscribed[id] = true
	}
}

// unsubscribe removes device IDs from the subscription list
func (c *WSClient) unsubscribe(deviceIDs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range deviceIDs {
		delete(c.subscribed, id)
	}
}
