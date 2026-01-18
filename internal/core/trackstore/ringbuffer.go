// Package trackstore provides historical trajectory storage for drones
package trackstore

import (
	"sync"
)

// TrackPoint represents a single point in a drone's trajectory
type TrackPoint struct {
	Timestamp int64   `json:"timestamp"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	LatGCJ02  float64 `json:"lat_gcj02,omitempty"`
	LonGCJ02  float64 `json:"lon_gcj02,omitempty"`
	Alt       float64 `json:"alt"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
}

// RingBuffer is a generic circular buffer
type RingBuffer struct {
	data  []TrackPoint
	head  int // Next write position
	size  int // Current number of elements
	cap   int // Maximum capacity
	mu    sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data: make([]TrackPoint, capacity),
		cap:  capacity,
	}
}

// Push adds a new point to the buffer
func (rb *RingBuffer) Push(point TrackPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = point
	rb.head = (rb.head + 1) % rb.cap

	if rb.size < rb.cap {
		rb.size++
	}
}

// GetAll returns all points in chronological order
func (rb *RingBuffer) GetAll() []TrackPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]TrackPoint, rb.size)
	if rb.size == 0 {
		return result
	}

	// Calculate start position (oldest element)
	start := 0
	if rb.size == rb.cap {
		start = rb.head // Buffer is full, oldest is at head
	}

	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.cap
		result[i] = rb.data[idx]
	}

	return result
}

// GetLast returns the last N points in chronological order
func (rb *RingBuffer) GetLast(n int) []TrackPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n > rb.size {
		n = rb.size
	}
	if n == 0 {
		return []TrackPoint{}
	}

	result := make([]TrackPoint, n)

	// Calculate start position for the last n elements
	start := (rb.head - n + rb.cap) % rb.cap

	for i := 0; i < n; i++ {
		idx := (start + i) % rb.cap
		result[i] = rb.data[idx]
	}

	return result
}

// GetSince returns all points since the given timestamp
func (rb *RingBuffer) GetSince(timestamp int64) []TrackPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.size == 0 {
		return []TrackPoint{}
	}

	var result []TrackPoint

	// Calculate start position (oldest element)
	start := 0
	if rb.size == rb.cap {
		start = rb.head
	}

	for i := 0; i < rb.size; i++ {
		idx := (start + i) % rb.cap
		if rb.data[idx].Timestamp >= timestamp {
			result = append(result, rb.data[idx])
		}
	}

	return result
}

// Size returns the current number of elements
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// Clear removes all elements
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.size = 0
}
