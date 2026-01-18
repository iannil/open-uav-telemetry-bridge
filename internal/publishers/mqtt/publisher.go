package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Publisher implements the core.Publisher interface for MQTT
type Publisher struct {
	cfg    config.MQTTConfig
	client pahomqtt.Client
	mu     sync.RWMutex
	ready  bool
}

// New creates a new MQTT publisher
func New(cfg config.MQTTConfig) *Publisher {
	return &Publisher{
		cfg: cfg,
	}
}

// Name returns the publisher name
func (p *Publisher) Name() string {
	return "mqtt"
}

// Start initializes the MQTT client and connects to the broker
func (p *Publisher) Start(ctx context.Context) error {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(p.cfg.Broker)
	opts.SetClientID(p.cfg.ClientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	// Set credentials if provided
	if p.cfg.Username != "" {
		opts.SetUsername(p.cfg.Username)
		opts.SetPassword(p.cfg.Password)
	}

	// Set Last Will and Testament (LWT)
	if p.cfg.LWT.Enabled {
		lwtTopic := fmt.Sprintf("%s/%s", p.cfg.LWT.Topic, p.cfg.ClientID)
		opts.SetWill(lwtTopic, p.cfg.LWT.Message, byte(p.cfg.QoS), true)
	}

	// Connection handlers
	opts.SetOnConnectHandler(func(c pahomqtt.Client) {
		p.mu.Lock()
		p.ready = true
		p.mu.Unlock()

		// Publish online status
		if p.cfg.LWT.Enabled {
			statusTopic := fmt.Sprintf("%s/%s", p.cfg.LWT.Topic, p.cfg.ClientID)
			c.Publish(statusTopic, byte(p.cfg.QoS), true, "online")
		}
	})

	opts.SetConnectionLostHandler(func(c pahomqtt.Client, err error) {
		p.mu.Lock()
		p.ready = false
		p.mu.Unlock()
	})

	// Create and connect client
	p.client = pahomqtt.NewClient(opts)
	token := p.client.Connect()

	// Wait for connection with timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		if !token.WaitTimeout(0) {
			return fmt.Errorf("mqtt connection timeout")
		}
	}

	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt connection failed: %w", token.Error())
	}

	return nil
}

// Publish sends a DroneState to the MQTT broker
func (p *Publisher) Publish(state *models.DroneState) error {
	p.mu.RLock()
	ready := p.ready
	p.mu.RUnlock()

	if !ready {
		return fmt.Errorf("mqtt client not connected")
	}

	// Serialize state to JSON
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	// Build topic: {prefix}/{device_id}/state
	topic := fmt.Sprintf("%s/%s/state", p.cfg.TopicPrefix, state.DeviceID)

	// Publish message
	token := p.client.Publish(topic, byte(p.cfg.QoS), false, payload)

	// Non-blocking publish - don't wait for confirmation
	go func() {
		if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
			// Log error but don't block
			_ = token.Error()
		}
	}()

	return nil
}

// PublishLocation sends only location data (lighter payload)
func (p *Publisher) PublishLocation(state *models.DroneState) error {
	p.mu.RLock()
	ready := p.ready
	p.mu.RUnlock()

	if !ready {
		return fmt.Errorf("mqtt client not connected")
	}

	// Lightweight location payload
	locationPayload := struct {
		DeviceID  string          `json:"device_id"`
		Timestamp int64           `json:"timestamp"`
		Location  models.Location `json:"location"`
	}{
		DeviceID:  state.DeviceID,
		Timestamp: state.Timestamp,
		Location:  state.Location,
	}

	payload, err := json.Marshal(locationPayload)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	topic := fmt.Sprintf("%s/%s/location", p.cfg.TopicPrefix, state.DeviceID)
	token := p.client.Publish(topic, byte(p.cfg.QoS), false, payload)

	go func() {
		token.WaitTimeout(5 * time.Second)
	}()

	return nil
}

// Stop gracefully stops the publisher
func (p *Publisher) Stop() error {
	if p.client != nil && p.client.IsConnected() {
		// Publish offline status before disconnecting
		if p.cfg.LWT.Enabled {
			statusTopic := fmt.Sprintf("%s/%s", p.cfg.LWT.Topic, p.cfg.ClientID)
			token := p.client.Publish(statusTopic, byte(p.cfg.QoS), true, "offline")
			token.WaitTimeout(2 * time.Second)
		}

		p.client.Disconnect(1000) // 1 second grace period
	}
	return nil
}

// IsConnected returns true if the client is connected
func (p *Publisher) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ready
}
