package gb28181

import (
	"sync"
	"time"
)

// Subscription represents a position subscription from the platform
type Subscription struct {
	ID        string    // Subscription ID (from SUBSCRIBE dialog)
	DeviceID  string    // Target device ID (or "*" for all)
	Interval  int       // Report interval in seconds
	Expires   time.Time // Subscription expiry time
	EventType string    // Event type (e.g., "presence")
}

// SubscriptionManager manages active subscriptions
type SubscriptionManager struct {
	subscriptions map[string]*Subscription
	mu            sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]*Subscription),
	}
}

// Add adds or updates a subscription
func (sm *SubscriptionManager) Add(sub *Subscription) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.subscriptions[sub.ID] = sub
}

// Remove removes a subscription
func (sm *SubscriptionManager) Remove(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.subscriptions, id)
}

// Get returns a subscription by ID
func (sm *SubscriptionManager) Get(id string) *Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.subscriptions[id]
}

// GetActive returns all active (non-expired) subscriptions
func (sm *SubscriptionManager) GetActive() []*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	var active []*Subscription
	for _, sub := range sm.subscriptions {
		if sub.Expires.After(now) {
			active = append(active, sub)
		}
	}
	return active
}

// GetForDevice returns subscriptions that apply to a specific device
func (sm *SubscriptionManager) GetForDevice(deviceID string) []*Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	var subs []*Subscription
	for _, sub := range sm.subscriptions {
		if sub.Expires.After(now) {
			if sub.DeviceID == "*" || sub.DeviceID == deviceID {
				subs = append(subs, sub)
			}
		}
	}
	return subs
}

// HasActiveSubscriptions returns true if there are any active subscriptions
func (sm *SubscriptionManager) HasActiveSubscriptions() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	for _, sub := range sm.subscriptions {
		if sub.Expires.After(now) {
			return true
		}
	}
	return false
}

// Cleanup removes expired subscriptions
func (sm *SubscriptionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, sub := range sm.subscriptions {
		if sub.Expires.Before(now) {
			delete(sm.subscriptions, id)
		}
	}
}

// StartCleanupLoop starts periodic cleanup of expired subscriptions
func (sm *SubscriptionManager) StartCleanupLoop(done <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			sm.Cleanup()
		}
	}
}
