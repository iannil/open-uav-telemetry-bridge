package throttler

import (
	"sync"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Throttler controls the rate of state updates per device
type Throttler struct {
	mu          sync.RWMutex
	rateHz      float64
	interval    time.Duration
	lastPublish map[string]time.Time
}

// New creates a new Throttler with the specified rate in Hz
func New(rateHz float64) *Throttler {
	if rateHz <= 0 {
		rateHz = 1.0
	}
	return &Throttler{
		rateHz:      rateHz,
		interval:    time.Duration(float64(time.Second) / rateHz),
		lastPublish: make(map[string]time.Time),
	}
}

// ShouldPublish returns true if enough time has passed since the last publish
// for this device. If true, it also updates the last publish time.
func (t *Throttler) ShouldPublish(state *models.DroneState) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	lastTime, exists := t.lastPublish[state.DeviceID]

	if !exists || now.Sub(lastTime) >= t.interval {
		t.lastPublish[state.DeviceID] = now
		return true
	}
	return false
}

// SetRate updates the throttle rate
func (t *Throttler) SetRate(rateHz float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if rateHz <= 0 {
		rateHz = 1.0
	}
	t.rateHz = rateHz
	t.interval = time.Duration(float64(time.Second) / rateHz)
}

// GetRate returns the current rate in Hz
func (t *Throttler) GetRate() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.rateHz
}

// Reset clears the last publish time for a device
func (t *Throttler) Reset(deviceID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastPublish, deviceID)
}

// ResetAll clears all last publish times
func (t *Throttler) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastPublish = make(map[string]time.Time)
}
