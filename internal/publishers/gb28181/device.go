package gb28181

import (
	"fmt"
	"sync"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Channel represents a GB28181 channel (corresponds to a drone)
type Channel struct {
	DeviceID   string    // 20-digit channel ID
	Name       string    // Channel name
	DroneID    string    // Original drone device ID
	Online     bool      // Online status
	LastUpdate time.Time // Last state update time
	LastState  *models.DroneState
}

// DeviceManager manages channels (drones) for GB28181 reporting
type DeviceManager struct {
	gatewayID string             // Gateway device ID
	civilCode string             // Administrative region code (first 6 digits)
	channels  map[string]*Channel // Map of drone ID to channel
	mu        sync.RWMutex
	nextSeq   int // Next channel sequence number
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(gatewayID string) *DeviceManager {
	civilCode := ""
	if len(gatewayID) >= 6 {
		civilCode = gatewayID[:6]
	}
	return &DeviceManager{
		gatewayID: gatewayID,
		civilCode: civilCode,
		channels:  make(map[string]*Channel),
		nextSeq:   1,
	}
}

// UpdateDrone updates or creates a channel for the given drone state
func (dm *DeviceManager) UpdateDrone(state *models.DroneState) *Channel {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	ch, exists := dm.channels[state.DeviceID]
	if !exists {
		// Create new channel with generated ID
		channelID := dm.generateChannelID()
		ch = &Channel{
			DeviceID: channelID,
			Name:     fmt.Sprintf("UAV-%s", state.DeviceID),
			DroneID:  state.DeviceID,
			Online:   true,
		}
		dm.channels[state.DeviceID] = ch
	}

	ch.LastUpdate = time.Now()
	ch.LastState = state
	ch.Online = true

	return ch
}

// GetChannel returns the channel for a drone ID
func (dm *DeviceManager) GetChannel(droneID string) *Channel {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.channels[droneID]
}

// GetAllChannels returns all channels
func (dm *DeviceManager) GetAllChannels() []*Channel {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	channels := make([]*Channel, 0, len(dm.channels))
	for _, ch := range dm.channels {
		channels = append(channels, ch)
	}
	return channels
}

// GetOnlineChannels returns only online channels
func (dm *DeviceManager) GetOnlineChannels() []*Channel {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var channels []*Channel
	for _, ch := range dm.channels {
		if ch.Online {
			channels = append(channels, ch)
		}
	}
	return channels
}

// MarkOffline marks a channel as offline based on timeout
func (dm *DeviceManager) MarkOffline(timeout time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	for _, ch := range dm.channels {
		if ch.Online && now.Sub(ch.LastUpdate) > timeout {
			ch.Online = false
		}
	}
}

// generateChannelID generates a 20-digit channel ID
// Format: PPPPPPTTTNNNSSSSSSSE (20 digits total)
// P: Administrative region (6 digits)
// T: Type (3 digits, 132=mobile camera device)
// N: Network/Domain (3 digits)
// S: Sequence (7 digits)
// E: Extension (1 digit)
func (dm *DeviceManager) generateChannelID() string {
	seq := dm.nextSeq
	dm.nextSeq++

	// Use gateway's civil code prefix, type 132 (mobile camera), network 200
	return fmt.Sprintf("%s132200%07d0", dm.civilCode, seq)
}

// GatewayID returns the gateway device ID
func (dm *DeviceManager) GatewayID() string {
	return dm.gatewayID
}

// CivilCode returns the administrative region code
func (dm *DeviceManager) CivilCode() string {
	return dm.civilCode
}
