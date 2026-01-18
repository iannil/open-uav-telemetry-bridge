package dji

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// MessageType defines the type of message from DJI forwarder
type MessageType string

const (
	MessageTypeHello     MessageType = "hello"
	MessageTypeState     MessageType = "state"
	MessageTypeHeartbeat MessageType = "heartbeat"
)

// Message represents a message from DJI forwarder
type Message struct {
	Type      MessageType     `json:"type"`
	DeviceID  string          `json:"device_id,omitempty"`
	SDKVersion string         `json:"sdk_version,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Client represents a connected DJI forwarder client
type Client struct {
	conn       net.Conn
	deviceID   string
	sdkVersion string
	lastSeen   time.Time
}

// Adapter implements the core.Adapter interface for DJI forwarder protocol
type Adapter struct {
	cfg      config.DJIConfig
	listener net.Listener
	clients  map[string]*Client
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// New creates a new DJI adapter
func New(cfg config.DJIConfig) *Adapter {
	return &Adapter{
		cfg:     cfg,
		clients: make(map[string]*Client),
	}
}

// Name returns the adapter name
func (a *Adapter) Name() string {
	return "dji"
}

// Start begins listening for DJI forwarder connections
func (a *Adapter) Start(ctx context.Context, events chan<- *models.DroneState) error {
	listener, err := net.Listen("tcp", a.cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", a.cfg.ListenAddress, err)
	}
	a.listener = listener

	log.Printf("[DJI] TCP server listening on %s", a.cfg.ListenAddress)

	// Accept connections in a goroutine
	a.wg.Add(1)
	go a.acceptLoop(ctx, events)

	return nil
}

// Stop gracefully stops the adapter
func (a *Adapter) Stop() error {
	if a.listener != nil {
		a.listener.Close()
	}

	// Close all client connections
	a.mu.Lock()
	for _, client := range a.clients {
		client.conn.Close()
	}
	a.clients = make(map[string]*Client)
	a.mu.Unlock()

	a.wg.Wait()
	log.Printf("[DJI] Adapter stopped")
	return nil
}

// acceptLoop accepts new connections
func (a *Adapter) acceptLoop(ctx context.Context, events chan<- *models.DroneState) {
	defer a.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Set accept deadline to allow checking context
		if tcpListener, ok := a.listener.(*net.TCPListener); ok {
			tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := a.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			log.Printf("[DJI] Accept error: %v", err)
			continue
		}

		// Check max clients
		a.mu.RLock()
		clientCount := len(a.clients)
		a.mu.RUnlock()

		if clientCount >= a.cfg.MaxClients {
			log.Printf("[DJI] Max clients reached (%d), rejecting connection", a.cfg.MaxClients)
			conn.Close()
			continue
		}

		log.Printf("[DJI] New connection from %s", conn.RemoteAddr())

		a.wg.Add(1)
		go a.handleClient(ctx, conn, events)
	}
}

// handleClient handles a single client connection
func (a *Adapter) handleClient(ctx context.Context, conn net.Conn, events chan<- *models.DroneState) {
	defer a.wg.Done()
	defer conn.Close()

	client := &Client{
		conn:     conn,
		lastSeen: time.Now(),
	}

	reader := bufio.NewReader(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Read message length (4 bytes, big endian)
		lengthBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lengthBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[DJI] Read length error from %s: %v", conn.RemoteAddr(), err)
			}
			a.removeClient(client)
			return
		}

		msgLength := binary.BigEndian.Uint32(lengthBuf)
		if msgLength > 65536 { // Max 64KB message
			log.Printf("[DJI] Message too large from %s: %d bytes", conn.RemoteAddr(), msgLength)
			a.removeClient(client)
			return
		}

		// Read message data
		msgBuf := make([]byte, msgLength)
		_, err = io.ReadFull(reader, msgBuf)
		if err != nil {
			log.Printf("[DJI] Read message error from %s: %v", conn.RemoteAddr(), err)
			a.removeClient(client)
			return
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(msgBuf, &msg); err != nil {
			log.Printf("[DJI] JSON parse error from %s: %v", conn.RemoteAddr(), err)
			continue
		}

		client.lastSeen = time.Now()

		// Handle message by type
		switch msg.Type {
		case MessageTypeHello:
			a.handleHello(client, &msg)
		case MessageTypeState:
			a.handleState(client, &msg, events)
		case MessageTypeHeartbeat:
			// Just update lastSeen, already done above
		default:
			log.Printf("[DJI] Unknown message type from %s: %s", conn.RemoteAddr(), msg.Type)
		}
	}
}

// handleHello processes HELLO message
func (a *Adapter) handleHello(client *Client, msg *Message) {
	client.deviceID = msg.DeviceID
	client.sdkVersion = msg.SDKVersion

	a.mu.Lock()
	a.clients[client.deviceID] = client
	a.mu.Unlock()

	log.Printf("[DJI] Client registered: %s (SDK %s)", client.deviceID, client.sdkVersion)

	// Send ACK
	ack := Message{Type: "ack"}
	a.sendMessage(client.conn, &ack)
}

// handleState processes STATE message
func (a *Adapter) handleState(client *Client, msg *Message, events chan<- *models.DroneState) {
	if client.deviceID == "" {
		log.Printf("[DJI] State received before hello from %s", client.conn.RemoteAddr())
		return
	}

	// Parse DroneState from data
	var state models.DroneState
	if err := json.Unmarshal(msg.Data, &state); err != nil {
		log.Printf("[DJI] Failed to parse state from %s: %v", client.deviceID, err)
		return
	}

	// Ensure device ID and protocol source are set
	if state.DeviceID == "" {
		state.DeviceID = client.deviceID
	}
	state.ProtocolSource = "dji"

	// Send to events channel
	select {
	case events <- &state:
	default:
		// Channel full, skip
	}
}

// removeClient removes a client from the clients map
func (a *Adapter) removeClient(client *Client) {
	if client.deviceID != "" {
		a.mu.Lock()
		delete(a.clients, client.deviceID)
		a.mu.Unlock()
		log.Printf("[DJI] Client disconnected: %s", client.deviceID)
	}
}

// sendMessage sends a message to a connection
func (a *Adapter) sendMessage(conn net.Conn, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write length prefix
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

// GetClientCount returns the number of connected clients
func (a *Adapter) GetClientCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.clients)
}

// GetClients returns information about connected clients
func (a *Adapter) GetClients() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ids := make([]string, 0, len(a.clients))
	for id := range a.clients {
		ids = append(ids, id)
	}
	return ids
}
