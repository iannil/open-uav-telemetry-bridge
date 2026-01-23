package mqtt

import (
	"testing"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestNew(t *testing.T) {
	cfg := config.MQTTConfig{
		Enabled:     true,
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		TopicPrefix: "uav/telemetry",
		QoS:         1,
	}

	p := New(cfg)

	if p == nil {
		t.Fatal("New should return non-nil publisher")
	}
	if p.cfg.Broker != "tcp://localhost:1883" {
		t.Errorf("Broker = %s, want 'tcp://localhost:1883'", p.cfg.Broker)
	}
	if p.cfg.ClientID != "test-client" {
		t.Errorf("ClientID = %s, want 'test-client'", p.cfg.ClientID)
	}
	if p.cfg.TopicPrefix != "uav/telemetry" {
		t.Errorf("TopicPrefix = %s, want 'uav/telemetry'", p.cfg.TopicPrefix)
	}
	if p.cfg.QoS != 1 {
		t.Errorf("QoS = %d, want 1", p.cfg.QoS)
	}
}

func TestPublisher_Name(t *testing.T) {
	p := New(config.MQTTConfig{})

	name := p.Name()

	if name != "mqtt" {
		t.Errorf("Name() = %s, want 'mqtt'", name)
	}
}

func TestPublisher_IsConnected_NotStarted(t *testing.T) {
	p := New(config.MQTTConfig{})

	connected := p.IsConnected()

	if connected {
		t.Error("IsConnected should return false when not started")
	}
}

func TestPublisher_Publish_NotConnected(t *testing.T) {
	p := New(config.MQTTConfig{})

	state := &models.DroneState{
		DeviceID:  "drone-1",
		Timestamp: 1234567890,
		Location: models.Location{
			Lat: 39.9087,
			Lon: 116.3975,
		},
	}

	err := p.Publish(state)

	if err == nil {
		t.Error("Publish should error when not connected")
	}
	if err.Error() != "mqtt client not connected" {
		t.Errorf("Error = '%v', want 'mqtt client not connected'", err)
	}
}

func TestPublisher_PublishLocation_NotConnected(t *testing.T) {
	p := New(config.MQTTConfig{})

	state := &models.DroneState{
		DeviceID:  "drone-1",
		Timestamp: 1234567890,
		Location: models.Location{
			Lat: 39.9087,
			Lon: 116.3975,
		},
	}

	err := p.PublishLocation(state)

	if err == nil {
		t.Error("PublishLocation should error when not connected")
	}
	if err.Error() != "mqtt client not connected" {
		t.Errorf("Error = '%v', want 'mqtt client not connected'", err)
	}
}

func TestPublisher_Stop_NilClient(t *testing.T) {
	p := New(config.MQTTConfig{})

	// Stop should not panic with nil client
	err := p.Stop()

	if err != nil {
		t.Errorf("Stop should not error with nil client: %v", err)
	}
}

func TestPublisher_ConfigWithAuth(t *testing.T) {
	cfg := config.MQTTConfig{
		Enabled:     true,
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		TopicPrefix: "uav/telemetry",
		QoS:         1,
		Username:    "testuser",
		Password:    "testpass",
	}

	p := New(cfg)

	if p.cfg.Username != "testuser" {
		t.Errorf("Username = %s, want 'testuser'", p.cfg.Username)
	}
	if p.cfg.Password != "testpass" {
		t.Errorf("Password = %s, want 'testpass'", p.cfg.Password)
	}
}

func TestPublisher_ConfigWithLWT(t *testing.T) {
	cfg := config.MQTTConfig{
		Enabled:     true,
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		TopicPrefix: "uav/telemetry",
		LWT: config.LWTConfig{
			Enabled: true,
			Topic:   "uav/status",
			Message: "offline",
		},
	}

	p := New(cfg)

	if !p.cfg.LWT.Enabled {
		t.Error("LWT.Enabled should be true")
	}
	if p.cfg.LWT.Topic != "uav/status" {
		t.Errorf("LWT.Topic = %s, want 'uav/status'", p.cfg.LWT.Topic)
	}
	if p.cfg.LWT.Message != "offline" {
		t.Errorf("LWT.Message = %s, want 'offline'", p.cfg.LWT.Message)
	}
}

func TestPublisher_ReadyState(t *testing.T) {
	p := New(config.MQTTConfig{})

	// Initially not ready
	if p.ready {
		t.Error("Publisher should not be ready initially")
	}

	// Simulate ready state (normally set by connection handler)
	p.mu.Lock()
	p.ready = true
	p.mu.Unlock()

	if !p.IsConnected() {
		t.Error("IsConnected should return true when ready")
	}

	// Reset
	p.mu.Lock()
	p.ready = false
	p.mu.Unlock()

	if p.IsConnected() {
		t.Error("IsConnected should return false when not ready")
	}
}
