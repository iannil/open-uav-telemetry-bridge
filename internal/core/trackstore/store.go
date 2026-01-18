package trackstore

import (
	"math"
	"sync"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Config holds configuration for the track store
type Config struct {
	MaxPointsPerDrone  int   // Maximum points to store per drone
	SampleIntervalMs   int64 // Minimum interval between samples in milliseconds
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MaxPointsPerDrone: 10000,
		SampleIntervalMs:  1000, // 1 second
	}
}

// Store manages trajectory data for multiple drones
type Store struct {
	tracks     map[string]*RingBuffer
	lastSample map[string]int64 // Last sample timestamp per device
	cfg        Config
	mu         sync.RWMutex
}

// New creates a new track store
func New(cfg Config) *Store {
	return &Store{
		tracks:     make(map[string]*RingBuffer),
		lastSample: make(map[string]int64),
		cfg:        cfg,
	}
}

// Record adds a new state to the trajectory
// Returns true if the point was recorded, false if it was skipped due to sampling
func (s *Store) Record(state *models.DroneState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check sampling interval
	lastTime := s.lastSample[state.DeviceID]
	now := state.Timestamp
	if now == 0 {
		now = time.Now().UnixMilli()
	}

	if now-lastTime < s.cfg.SampleIntervalMs {
		return false // Skip this sample
	}

	// Get or create ring buffer for this device
	rb, exists := s.tracks[state.DeviceID]
	if !exists {
		rb = NewRingBuffer(s.cfg.MaxPointsPerDrone)
		s.tracks[state.DeviceID] = rb
	}

	// Calculate speed from velocity
	speed := math.Sqrt(
		state.Velocity.Vx*state.Velocity.Vx +
		state.Velocity.Vy*state.Velocity.Vy +
		state.Velocity.Vz*state.Velocity.Vz,
	)

	// Create track point
	point := TrackPoint{
		Timestamp: now,
		Lat:       state.Location.Lat,
		Lon:       state.Location.Lon,
		Alt:       state.Location.AltGNSS,
		Heading:   state.Attitude.Yaw,
		Speed:     speed,
	}

	// Add converted coordinates if available
	if state.Location.LatGCJ02 != nil {
		point.LatGCJ02 = *state.Location.LatGCJ02
	}
	if state.Location.LonGCJ02 != nil {
		point.LonGCJ02 = *state.Location.LonGCJ02
	}

	rb.Push(point)
	s.lastSample[state.DeviceID] = now

	return true
}

// GetTrack returns the trajectory for a device
func (s *Store) GetTrack(deviceID string, limit int, since int64) []TrackPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rb, exists := s.tracks[deviceID]
	if !exists {
		return []TrackPoint{}
	}

	if since > 0 {
		points := rb.GetSince(since)
		if limit > 0 && len(points) > limit {
			return points[len(points)-limit:]
		}
		return points
	}

	if limit > 0 {
		return rb.GetLast(limit)
	}

	return rb.GetAll()
}

// ClearTrack removes all trajectory data for a device
func (s *Store) ClearTrack(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rb, exists := s.tracks[deviceID]; exists {
		rb.Clear()
	}
	delete(s.lastSample, deviceID)
}

// GetTrackSize returns the number of points stored for a device
func (s *Store) GetTrackSize(deviceID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if rb, exists := s.tracks[deviceID]; exists {
		return rb.Size()
	}
	return 0
}

// GetDeviceIDs returns all device IDs with stored tracks
func (s *Store) GetDeviceIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.tracks))
	for id := range s.tracks {
		ids = append(ids, id)
	}
	return ids
}

// GetTotalPoints returns the total number of points across all devices
func (s *Store) GetTotalPoints() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, rb := range s.tracks {
		total += rb.Size()
	}
	return total
}
