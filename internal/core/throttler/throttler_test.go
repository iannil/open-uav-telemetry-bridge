package throttler

import (
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestThrottler(t *testing.T) {
	// 10 Hz = 100ms interval
	throttler := New(10.0)

	state := models.NewDroneState("uav-001", "mavlink")

	// First call should always return true
	if !throttler.ShouldPublish(state) {
		t.Error("First call should return true")
	}

	// Immediate second call should return false
	if throttler.ShouldPublish(state) {
		t.Error("Immediate second call should return false")
	}

	// After waiting longer than interval, should return true
	time.Sleep(110 * time.Millisecond)
	if !throttler.ShouldPublish(state) {
		t.Error("Call after interval should return true")
	}
}

func TestThrottlerSetRate(t *testing.T) {
	throttler := New(1.0)

	if throttler.GetRate() != 1.0 {
		t.Errorf("Expected rate 1.0, got %f", throttler.GetRate())
	}

	throttler.SetRate(5.0)
	if throttler.GetRate() != 5.0 {
		t.Errorf("Expected rate 5.0, got %f", throttler.GetRate())
	}

	// Invalid rate should default to 1.0
	throttler.SetRate(0)
	if throttler.GetRate() != 1.0 {
		t.Errorf("Expected rate 1.0 for invalid input, got %f", throttler.GetRate())
	}
}

func TestThrottlerReset(t *testing.T) {
	throttler := New(10.0)
	state := models.NewDroneState("uav-001", "mavlink")

	// Publish once
	throttler.ShouldPublish(state)

	// Should not publish immediately
	if throttler.ShouldPublish(state) {
		t.Error("Should not publish immediately")
	}

	// Reset and try again
	throttler.Reset("uav-001")
	if !throttler.ShouldPublish(state) {
		t.Error("Should publish after reset")
	}
}

func TestThrottlerMultipleDevices(t *testing.T) {
	throttler := New(10.0)

	state1 := models.NewDroneState("uav-001", "mavlink")
	state2 := models.NewDroneState("uav-002", "mavlink")

	// Both devices should be able to publish initially
	if !throttler.ShouldPublish(state1) {
		t.Error("Device 1 first call should return true")
	}
	if !throttler.ShouldPublish(state2) {
		t.Error("Device 2 first call should return true")
	}

	// Neither should publish immediately after
	if throttler.ShouldPublish(state1) {
		t.Error("Device 1 immediate second call should return false")
	}
	if throttler.ShouldPublish(state2) {
		t.Error("Device 2 immediate second call should return false")
	}
}

func TestThrottlerResetAll(t *testing.T) {
	throttler := New(10.0)

	state1 := models.NewDroneState("uav-001", "mavlink")
	state2 := models.NewDroneState("uav-002", "mavlink")

	// Publish both
	throttler.ShouldPublish(state1)
	throttler.ShouldPublish(state2)

	// Neither should publish
	if throttler.ShouldPublish(state1) || throttler.ShouldPublish(state2) {
		t.Error("Should not publish immediately")
	}

	// Reset all
	throttler.ResetAll()

	// Both should publish now
	if !throttler.ShouldPublish(state1) {
		t.Error("Device 1 should publish after ResetAll")
	}
	if !throttler.ShouldPublish(state2) {
		t.Error("Device 2 should publish after ResetAll")
	}
}
