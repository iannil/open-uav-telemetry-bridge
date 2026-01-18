package statestore

import (
	"sync"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// StateStore provides thread-safe in-memory caching of drone states
type StateStore struct {
	mu     sync.RWMutex
	states map[string]*models.DroneState
}

// New creates a new StateStore
func New() *StateStore {
	return &StateStore{
		states: make(map[string]*models.DroneState),
	}
}

// Update stores or updates the state for a device
func (s *StateStore) Update(state *models.DroneState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.DeviceID] = state
}

// Get retrieves the current state for a device
// Returns nil if the device is not found
func (s *StateStore) Get(deviceID string) *models.DroneState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[deviceID]
}

// GetAll returns a copy of all current states
func (s *StateStore) GetAll() []*models.DroneState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.DroneState, 0, len(s.states))
	for _, state := range s.states {
		result = append(result, state)
	}
	return result
}

// Delete removes a device from the store
func (s *StateStore) Delete(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, deviceID)
}

// Count returns the number of devices in the store
func (s *StateStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.states)
}
