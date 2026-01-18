package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// serveWs handles WebSocket requests from clients
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	client := &WSClient{
		hub:        s.hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		subscribed: make(map[string]bool),
	}

	client.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}

		// Parse client message
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WebSocket] Invalid message: %v", err)
			continue
		}

		// Handle client commands
		c.handleMessage(&msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WSClient) writePump() {
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
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
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

// handleMessage processes incoming client messages
func (c *WSClient) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case WSMessageTypeSubscribe:
		var payload struct {
			DeviceIDs []string `json:"device_ids"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err == nil {
			c.subscribe(payload.DeviceIDs)
			log.Printf("[WebSocket] Client subscribed to: %v", payload.DeviceIDs)
		}

	case WSMessageTypeUnsubscribe:
		var payload struct {
			DeviceIDs []string `json:"device_ids"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err == nil {
			c.unsubscribe(payload.DeviceIDs)
			log.Printf("[WebSocket] Client unsubscribed from: %v", payload.DeviceIDs)
		}

	default:
		log.Printf("[WebSocket] Unknown message type: %s", msg.Type)
	}
}

// sendError sends an error message to the client
func (c *WSClient) sendError(message string) {
	msg := WSMessage{
		Type: WSMessageTypeError,
	}
	data, _ := json.Marshal(map[string]string{"message": message})
	msg.Data = data

	msgBytes, _ := json.Marshal(msg)
	select {
	case c.send <- msgBytes:
	default:
	}
}
