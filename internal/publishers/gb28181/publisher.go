package gb28181

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
	gbxml "github.com/open-uav/telemetry-bridge/internal/publishers/gb28181/xml"
)

// Publisher implements the core.Publisher interface for GB/T 28181
type Publisher struct {
	cfg       config.GB28181Config
	sipClient *SIPClient
	deviceMgr *DeviceManager
	subMgr    *SubscriptionManager
	handler   *RequestHandler

	mu            sync.RWMutex
	running       bool
	lastStates    map[string]*models.DroneState
	lastSentTimes map[string]time.Time

	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	done       chan struct{}
}

// New creates a new GB28181 publisher
func New(cfg config.GB28181Config) *Publisher {
	return &Publisher{
		cfg:           cfg,
		lastStates:    make(map[string]*models.DroneState),
		lastSentTimes: make(map[string]time.Time),
		done:          make(chan struct{}),
	}
}

// Name returns the publisher name
func (p *Publisher) Name() string {
	return "gb28181"
}

// Start initializes the GB28181 publisher and connects to the SIP server
func (p *Publisher) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("publisher already running")
	}
	p.running = true
	p.mu.Unlock()

	// Create child context
	p.ctx, p.cancel = context.WithCancel(ctx)

	// Initialize components
	p.sipClient = NewSIPClient(p.cfg)
	p.deviceMgr = NewDeviceManager(p.cfg.DeviceID)
	p.subMgr = NewSubscriptionManager()
	p.handler = NewRequestHandler(p.deviceMgr, p.subMgr, p.sipClient)

	// Start SIP client
	if err := p.sipClient.Start(p.ctx); err != nil {
		return fmt.Errorf("start SIP client: %w", err)
	}

	// Set request handler
	p.sipClient.SetRequestHandler(p.handler.HandleRequest)

	// Initial registration
	if err := p.sipClient.Register(p.ctx); err != nil {
		return fmt.Errorf("initial registration: %w", err)
	}

	// Start background tasks
	p.wg.Add(3)

	// Registration refresh loop
	go func() {
		defer p.wg.Done()
		p.sipClient.StartRegistrationLoop(p.ctx)
	}()

	// Heartbeat loop
	go func() {
		defer p.wg.Done()
		p.heartbeatLoop()
	}()

	// Subscription cleanup loop
	go func() {
		defer p.wg.Done()
		p.subMgr.StartCleanupLoop(p.done)
	}()

	log.Printf("[GB28181] Publisher started (device: %s, server: %s:%d)",
		p.cfg.DeviceID, p.cfg.ServerIP, p.cfg.ServerPort)

	return nil
}

// Publish sends a DroneState to the GB28181 platform
func (p *Publisher) Publish(state *models.DroneState) error {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return fmt.Errorf("publisher not running")
	}
	p.mu.RUnlock()

	// Check if SIP client is registered
	if !p.sipClient.IsRegistered() {
		return fmt.Errorf("not registered with SIP server")
	}

	// Update device manager
	p.deviceMgr.UpdateDrone(state)

	// Store latest state
	p.mu.Lock()
	p.lastStates[state.DeviceID] = state
	p.mu.Unlock()

	// Check if we should send position based on interval
	if !p.shouldSendPosition(state.DeviceID) {
		return nil
	}

	// Send position notification
	return p.sendPositionNotify(state)
}

// shouldSendPosition checks if enough time has passed since the last send
func (p *Publisher) shouldSendPosition(deviceID string) bool {
	p.mu.RLock()
	lastSent, exists := p.lastSentTimes[deviceID]
	p.mu.RUnlock()

	if !exists {
		return true
	}

	interval := time.Duration(p.cfg.PositionInterval) * time.Second
	return time.Since(lastSent) >= interval
}

// sendPositionNotify sends a MobilePosition notification
func (p *Publisher) sendPositionNotify(state *models.DroneState) error {
	// Create MobilePosition XML
	notify := gbxml.NewMobilePositionNotify(state, p.sipClient.NextSN())
	body, err := notify.Marshal()
	if err != nil {
		return fmt.Errorf("marshal position notify: %w", err)
	}

	// Send via SIP NOTIFY
	if err := p.sipClient.SendNotify(p.ctx, "Application/MANSCDP+xml", body); err != nil {
		return fmt.Errorf("send position notify: %w", err)
	}

	// Update last sent time
	p.mu.Lock()
	p.lastSentTimes[state.DeviceID] = time.Now()
	p.mu.Unlock()

	return nil
}

// heartbeatLoop sends periodic keepalive messages
func (p *Publisher) heartbeatLoop() {
	interval := time.Duration(p.cfg.HeartbeatInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a keepalive message
func (p *Publisher) sendHeartbeat() {
	if !p.sipClient.IsRegistered() {
		return
	}

	notify := gbxml.NewKeepaliveNotify(p.cfg.DeviceID, p.sipClient.NextSN())
	body, err := notify.Marshal()
	if err != nil {
		log.Printf("[GB28181] Failed to marshal keepalive: %v", err)
		return
	}

	if err := p.sipClient.SendMessage(p.ctx, "Application/MANSCDP+xml", body); err != nil {
		log.Printf("[GB28181] Failed to send keepalive: %v", err)
	}
}

// Stop gracefully stops the publisher
func (p *Publisher) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = false
	p.mu.Unlock()

	// Signal shutdown
	close(p.done)
	if p.cancel != nil {
		p.cancel()
	}

	// Wait for background tasks
	p.wg.Wait()

	// Stop SIP client
	if p.sipClient != nil {
		if err := p.sipClient.Stop(); err != nil {
			log.Printf("[GB28181] Error stopping SIP client: %v", err)
		}
	}

	log.Printf("[GB28181] Publisher stopped")
	return nil
}

// IsConnected returns true if the publisher is registered with the SIP server
func (p *Publisher) IsConnected() bool {
	if p.sipClient == nil {
		return false
	}
	return p.sipClient.IsRegistered()
}

// GetOnlineDevices returns the count of online devices
func (p *Publisher) GetOnlineDevices() int {
	return len(p.deviceMgr.GetOnlineChannels())
}

// GetActiveSubscriptions returns the count of active subscriptions
func (p *Publisher) GetActiveSubscriptions() int {
	return len(p.subMgr.GetActive())
}
