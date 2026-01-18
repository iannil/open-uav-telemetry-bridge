package trackstore

import (
	"testing"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestRingBuffer_Push(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Push(TrackPoint{Timestamp: 1, Lat: 1.0})
	rb.Push(TrackPoint{Timestamp: 2, Lat: 2.0})
	rb.Push(TrackPoint{Timestamp: 3, Lat: 3.0})

	if rb.Size() != 3 {
		t.Errorf("Size() = %d, want 3", rb.Size())
	}

	// Push one more, should overwrite oldest
	rb.Push(TrackPoint{Timestamp: 4, Lat: 4.0})

	if rb.Size() != 3 {
		t.Errorf("Size() = %d, want 3 after overflow", rb.Size())
	}

	points := rb.GetAll()
	if len(points) != 3 {
		t.Errorf("GetAll() returned %d points, want 3", len(points))
	}

	// Should be in chronological order: 2, 3, 4
	if points[0].Timestamp != 2 {
		t.Errorf("First point timestamp = %d, want 2", points[0].Timestamp)
	}
	if points[2].Timestamp != 4 {
		t.Errorf("Last point timestamp = %d, want 4", points[2].Timestamp)
	}
}

func TestRingBuffer_GetLast(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := 1; i <= 5; i++ {
		rb.Push(TrackPoint{Timestamp: int64(i)})
	}

	// Get last 3
	points := rb.GetLast(3)
	if len(points) != 3 {
		t.Errorf("GetLast(3) returned %d points, want 3", len(points))
	}

	if points[0].Timestamp != 3 {
		t.Errorf("First point = %d, want 3", points[0].Timestamp)
	}
	if points[2].Timestamp != 5 {
		t.Errorf("Last point = %d, want 5", points[2].Timestamp)
	}

	// Get more than available
	points = rb.GetLast(10)
	if len(points) != 5 {
		t.Errorf("GetLast(10) returned %d points, want 5", len(points))
	}
}

func TestRingBuffer_GetSince(t *testing.T) {
	rb := NewRingBuffer(10)

	for i := 1; i <= 5; i++ {
		rb.Push(TrackPoint{Timestamp: int64(i * 1000)})
	}

	// Get since timestamp 3000
	points := rb.GetSince(3000)
	if len(points) != 3 {
		t.Errorf("GetSince(3000) returned %d points, want 3", len(points))
	}

	if points[0].Timestamp != 3000 {
		t.Errorf("First point = %d, want 3000", points[0].Timestamp)
	}

	// Get since future timestamp
	points = rb.GetSince(10000)
	if len(points) != 0 {
		t.Errorf("GetSince(10000) returned %d points, want 0", len(points))
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer(5)

	rb.Push(TrackPoint{Timestamp: 1})
	rb.Push(TrackPoint{Timestamp: 2})

	rb.Clear()

	if rb.Size() != 0 {
		t.Errorf("Size() after Clear() = %d, want 0", rb.Size())
	}

	points := rb.GetAll()
	if len(points) != 0 {
		t.Errorf("GetAll() after Clear() returned %d points, want 0", len(points))
	}
}

func TestStore_Record(t *testing.T) {
	cfg := Config{
		MaxPointsPerDrone: 100,
		SampleIntervalMs:  100,
	}
	store := New(cfg)

	state := &models.DroneState{
		DeviceID:  "test-001",
		Timestamp: 1000,
		Location: models.Location{
			Lat:     39.9,
			Lon:     116.4,
			AltGNSS: 100,
		},
		Attitude: models.Attitude{
			Yaw: 45,
		},
		Velocity: models.Velocity{
			Vx: 3,
			Vy: 4,
		},
	}

	// First record should succeed
	if !store.Record(state) {
		t.Error("First Record() should return true")
	}

	// Second record within interval should fail
	state.Timestamp = 1050
	if store.Record(state) {
		t.Error("Record() within interval should return false")
	}

	// Record after interval should succeed
	state.Timestamp = 1200
	if !store.Record(state) {
		t.Error("Record() after interval should return true")
	}

	if store.GetTrackSize("test-001") != 2 {
		t.Errorf("GetTrackSize() = %d, want 2", store.GetTrackSize("test-001"))
	}
}

func TestStore_GetTrack(t *testing.T) {
	cfg := Config{
		MaxPointsPerDrone: 100,
		SampleIntervalMs:  0, // No sampling limit for test
	}
	store := New(cfg)

	// Add 5 points
	for i := 1; i <= 5; i++ {
		state := &models.DroneState{
			DeviceID:  "test-001",
			Timestamp: int64(i * 1000),
			Location: models.Location{
				Lat: float64(i),
			},
		}
		store.Record(state)
	}

	// Get all
	points := store.GetTrack("test-001", 0, 0)
	if len(points) != 5 {
		t.Errorf("GetTrack() returned %d points, want 5", len(points))
	}

	// Get last 3
	points = store.GetTrack("test-001", 3, 0)
	if len(points) != 3 {
		t.Errorf("GetTrack(limit=3) returned %d points, want 3", len(points))
	}

	// Get since timestamp
	points = store.GetTrack("test-001", 0, 3000)
	if len(points) != 3 {
		t.Errorf("GetTrack(since=3000) returned %d points, want 3", len(points))
	}

	// Get non-existent device
	points = store.GetTrack("unknown", 0, 0)
	if len(points) != 0 {
		t.Errorf("GetTrack(unknown) returned %d points, want 0", len(points))
	}
}

func TestStore_ClearTrack(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SampleIntervalMs = 0
	store := New(cfg)

	state := &models.DroneState{
		DeviceID:  "test-001",
		Timestamp: 1000,
	}
	store.Record(state)

	store.ClearTrack("test-001")

	if store.GetTrackSize("test-001") != 0 {
		t.Error("GetTrackSize() after ClearTrack() should be 0")
	}
}

func TestStore_SpeedCalculation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SampleIntervalMs = 0
	store := New(cfg)

	state := &models.DroneState{
		DeviceID:  "test-001",
		Timestamp: 1000,
		Velocity: models.Velocity{
			Vx: 3,
			Vy: 4,
			Vz: 0,
		},
	}
	store.Record(state)

	points := store.GetTrack("test-001", 1, 0)
	if len(points) != 1 {
		t.Fatalf("Expected 1 point, got %d", len(points))
	}

	// Speed should be sqrt(3^2 + 4^2) = 5
	expectedSpeed := 5.0
	if points[0].Speed != expectedSpeed {
		t.Errorf("Speed = %f, want %f", points[0].Speed, expectedSpeed)
	}
}
