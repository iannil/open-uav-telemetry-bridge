package dji

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestNew(t *testing.T) {
	cfg := config.DJIConfig{
		Enabled:       true,
		ListenAddress: "0.0.0.0:14560",
		MaxClients:    10,
	}

	a := New(cfg)

	if a == nil {
		t.Fatal("New should return non-nil adapter")
	}
	if a.cfg.ListenAddress != "0.0.0.0:14560" {
		t.Errorf("ListenAddress = %s, want '0.0.0.0:14560'", a.cfg.ListenAddress)
	}
	if a.cfg.MaxClients != 10 {
		t.Errorf("MaxClients = %d, want 10", a.cfg.MaxClients)
	}
	if a.clients == nil {
		t.Error("clients map should be initialized")
	}
}

func TestAdapter_Name(t *testing.T) {
	a := New(config.DJIConfig{})

	name := a.Name()

	if name != "dji" {
		t.Errorf("Name() = %s, want 'dji'", name)
	}
}

func TestAdapter_Stop_NilListener(t *testing.T) {
	a := New(config.DJIConfig{})

	// Stop should not panic with nil listener
	err := a.Stop()

	if err != nil {
		t.Errorf("Stop should not error with nil listener: %v", err)
	}
}

func TestAdapter_GetClientCount_Empty(t *testing.T) {
	a := New(config.DJIConfig{})

	count := a.GetClientCount()

	if count != 0 {
		t.Errorf("GetClientCount = %d, want 0", count)
	}
}

func TestAdapter_GetClients_Empty(t *testing.T) {
	a := New(config.DJIConfig{})

	clients := a.GetClients()

	if len(clients) != 0 {
		t.Errorf("GetClients length = %d, want 0", len(clients))
	}
}

func TestAdapter_ClientManagement(t *testing.T) {
	a := New(config.DJIConfig{})

	// Add a mock client
	a.mu.Lock()
	a.clients["drone-1"] = &Client{
		deviceID:   "drone-1",
		sdkVersion: "5.0",
		lastSeen:   time.Now(),
	}
	a.clients["drone-2"] = &Client{
		deviceID:   "drone-2",
		sdkVersion: "5.1",
		lastSeen:   time.Now(),
	}
	a.mu.Unlock()

	// Test GetClientCount
	count := a.GetClientCount()
	if count != 2 {
		t.Errorf("GetClientCount = %d, want 2", count)
	}

	// Test GetClients
	clients := a.GetClients()
	if len(clients) != 2 {
		t.Errorf("GetClients length = %d, want 2", len(clients))
	}

	// Check that both IDs are present
	hasD1, hasD2 := false, false
	for _, id := range clients {
		if id == "drone-1" {
			hasD1 = true
		}
		if id == "drone-2" {
			hasD2 = true
		}
	}
	if !hasD1 || !hasD2 {
		t.Error("GetClients should contain both drone-1 and drone-2")
	}
}

func TestAdapter_removeClient(t *testing.T) {
	a := New(config.DJIConfig{})

	// Add a client
	a.mu.Lock()
	a.clients["drone-1"] = &Client{
		deviceID:   "drone-1",
		sdkVersion: "5.0",
	}
	a.mu.Unlock()

	if a.GetClientCount() != 1 {
		t.Fatal("Should have 1 client before remove")
	}

	// Remove the client
	client := &Client{deviceID: "drone-1"}
	a.removeClient(client)

	if a.GetClientCount() != 0 {
		t.Error("Should have 0 clients after remove")
	}
}

func TestAdapter_removeClient_Empty(t *testing.T) {
	a := New(config.DJIConfig{})

	// Remove with empty deviceID should not panic
	client := &Client{deviceID: ""}
	a.removeClient(client)

	// Should still work
	if a.GetClientCount() != 0 {
		t.Error("GetClientCount should be 0")
	}
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name string
		mt   MessageType
		want string
	}{
		{"hello", MessageTypeHello, "hello"},
		{"state", MessageTypeState, "state"},
		{"heartbeat", MessageTypeHeartbeat, "heartbeat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mt) != tt.want {
				t.Errorf("MessageType = %s, want %s", tt.mt, tt.want)
			}
		})
	}
}

func TestMessage_JSON(t *testing.T) {
	// Test hello message
	helloMsg := Message{
		Type:       MessageTypeHello,
		DeviceID:   "drone-1",
		SDKVersion: "5.0",
	}

	data, err := json.Marshal(helloMsg)
	if err != nil {
		t.Fatalf("Marshal hello message failed: %v", err)
	}

	var parsed Message
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal hello message failed: %v", err)
	}

	if parsed.Type != MessageTypeHello {
		t.Errorf("Type = %s, want 'hello'", parsed.Type)
	}
	if parsed.DeviceID != "drone-1" {
		t.Errorf("DeviceID = %s, want 'drone-1'", parsed.DeviceID)
	}
}

func TestAdapter_sendMessage(t *testing.T) {
	a := New(config.DJIConfig{})

	// Create a pipe for testing
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Send message in goroutine
	go func() {
		msg := &Message{Type: "ack"}
		a.sendMessage(serverConn, msg)
	}()

	// Read length prefix
	lengthBuf := make([]byte, 4)
	clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err := clientConn.Read(lengthBuf)
	if err != nil {
		t.Fatalf("Read length failed: %v", err)
	}

	msgLength := binary.BigEndian.Uint32(lengthBuf)
	if msgLength == 0 {
		t.Error("Message length should not be 0")
	}

	// Read message data
	msgBuf := make([]byte, msgLength)
	_, err = clientConn.Read(msgBuf)
	if err != nil {
		t.Fatalf("Read message failed: %v", err)
	}

	// Parse message
	var msg Message
	if err := json.Unmarshal(msgBuf, &msg); err != nil {
		t.Fatalf("Unmarshal message failed: %v", err)
	}

	if msg.Type != "ack" {
		t.Errorf("Message type = %s, want 'ack'", msg.Type)
	}
}

func TestAdapter_handleHello(t *testing.T) {
	a := New(config.DJIConfig{})

	// Create a pipe for testing
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	client := &Client{
		conn:     serverConn,
		lastSeen: time.Now(),
	}

	msg := &Message{
		Type:       MessageTypeHello,
		DeviceID:   "test-drone",
		SDKVersion: "5.0",
	}

	// Read ACK response in goroutine
	done := make(chan bool)
	go func() {
		lengthBuf := make([]byte, 4)
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		clientConn.Read(lengthBuf)
		msgLength := binary.BigEndian.Uint32(lengthBuf)
		msgBuf := make([]byte, msgLength)
		clientConn.Read(msgBuf)
		done <- true
	}()

	// Handle hello
	a.handleHello(client, msg)

	// Wait for response
	select {
	case <-done:
		// Good
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for ACK")
	}

	// Verify client was registered
	if client.deviceID != "test-drone" {
		t.Errorf("Client deviceID = %s, want 'test-drone'", client.deviceID)
	}
	if client.sdkVersion != "5.0" {
		t.Errorf("Client sdkVersion = %s, want '5.0'", client.sdkVersion)
	}

	// Verify client is in map
	if a.GetClientCount() != 1 {
		t.Error("Client should be registered")
	}
}

func TestAdapter_handleState(t *testing.T) {
	a := New(config.DJIConfig{})

	client := &Client{
		deviceID:   "test-drone",
		sdkVersion: "5.0",
	}

	// Create state data
	stateData := models.DroneState{
		DeviceID:  "test-drone",
		Timestamp: 1234567890,
		Location: models.Location{
			Lat: 39.9087,
			Lon: 116.3975,
		},
	}
	dataBytes, _ := json.Marshal(stateData)

	msg := &Message{
		Type: MessageTypeState,
		Data: dataBytes,
	}

	events := make(chan *models.DroneState, 1)

	// Handle state
	a.handleState(client, msg, events)

	// Check event was sent
	select {
	case state := <-events:
		if state.DeviceID != "test-drone" {
			t.Errorf("State DeviceID = %s, want 'test-drone'", state.DeviceID)
		}
		if state.ProtocolSource != "dji" {
			t.Errorf("State ProtocolSource = %s, want 'dji'", state.ProtocolSource)
		}
		if state.Location.Lat != 39.9087 {
			t.Errorf("State Lat = %f, want 39.9087", state.Location.Lat)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("State event should be sent")
	}
}

func TestAdapter_handleState_NoHello(t *testing.T) {
	a := New(config.DJIConfig{})

	// Create a pipe for the conn
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Client without deviceID (no hello received)
	client := &Client{
		deviceID: "",
		conn:     serverConn,
	}

	msg := &Message{
		Type: MessageTypeState,
		Data: []byte(`{}`),
	}

	events := make(chan *models.DroneState, 1)

	// Handle state - should be rejected
	a.handleState(client, msg, events)

	// Check no event was sent
	select {
	case <-events:
		t.Error("State event should not be sent before hello")
	case <-time.After(100 * time.Millisecond):
		// Good - no event expected
	}
}

func TestAdapter_handleState_SetDeviceID(t *testing.T) {
	a := New(config.DJIConfig{})

	client := &Client{
		deviceID:   "test-drone",
		sdkVersion: "5.0",
	}

	// State without deviceID
	stateData := models.DroneState{
		Timestamp: 1234567890,
	}
	dataBytes, _ := json.Marshal(stateData)

	msg := &Message{
		Type: MessageTypeState,
		Data: dataBytes,
	}

	events := make(chan *models.DroneState, 1)

	// Handle state
	a.handleState(client, msg, events)

	// Check event has deviceID set from client
	select {
	case state := <-events:
		if state.DeviceID != "test-drone" {
			t.Errorf("State DeviceID should be set from client, got %s", state.DeviceID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("State event should be sent")
	}
}
