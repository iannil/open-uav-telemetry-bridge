package models

import (
	"encoding/json"
	"testing"
)

func TestDroneStateJSON(t *testing.T) {
	state := &DroneState{
		DeviceID:       "uav-001",
		Timestamp:      1709882231000,
		ProtocolSource: "mavlink",
		Location: Location{
			Lat:              22.5431,
			Lon:              114.0579,
			AltBaro:          120.5,
			AltGNSS:          125.0,
			CoordinateSystem: "WGS84",
		},
		Attitude: Attitude{
			Roll:  0.05,
			Pitch: -0.12,
			Yaw:   180.0,
		},
		Status: Status{
			BatteryPercent: 85,
			FlightMode:     FlightModeAuto,
			Armed:          true,
			SignalQuality:  95,
		},
		Velocity: Velocity{
			Vx: 10.5,
			Vy: 0.0,
			Vz: -0.5,
		},
	}

	// Test serialization
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Failed to marshal DroneState: %v", err)
	}

	// Test deserialization
	var decoded DroneState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DroneState: %v", err)
	}

	// Verify fields
	if decoded.DeviceID != state.DeviceID {
		t.Errorf("DeviceID mismatch: got %s, want %s", decoded.DeviceID, state.DeviceID)
	}
	if decoded.Location.Lat != state.Location.Lat {
		t.Errorf("Lat mismatch: got %f, want %f", decoded.Location.Lat, state.Location.Lat)
	}
	if decoded.Status.FlightMode != state.Status.FlightMode {
		t.Errorf("FlightMode mismatch: got %s, want %s", decoded.Status.FlightMode, state.Status.FlightMode)
	}
}

func TestNewDroneState(t *testing.T) {
	state := NewDroneState("uav-002", "mavlink")

	if state.DeviceID != "uav-002" {
		t.Errorf("DeviceID mismatch: got %s, want uav-002", state.DeviceID)
	}
	if state.ProtocolSource != "mavlink" {
		t.Errorf("ProtocolSource mismatch: got %s, want mavlink", state.ProtocolSource)
	}
	if state.Location.CoordinateSystem != "WGS84" {
		t.Errorf("CoordinateSystem mismatch: got %s, want WGS84", state.Location.CoordinateSystem)
	}
	if state.Status.FlightMode != FlightModeUnknown {
		t.Errorf("FlightMode mismatch: got %s, want UNKNOWN", state.Status.FlightMode)
	}
}

func TestFlightModeConstants(t *testing.T) {
	modes := []FlightMode{
		FlightModeUnknown,
		FlightModeManual,
		FlightModeStabilize,
		FlightModeAltHold,
		FlightModeLoiter,
		FlightModeAuto,
		FlightModeGuided,
		FlightModeRTL,
		FlightModeLand,
		FlightModeTakeoff,
		FlightModeEmergency,
	}

	for _, mode := range modes {
		if mode == "" {
			t.Errorf("FlightMode should not be empty")
		}
	}
}
