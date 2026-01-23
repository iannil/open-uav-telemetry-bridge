package mavlink

import (
	"testing"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestNew(t *testing.T) {
	cfg := config.MAVLinkConfig{
		Enabled:        true,
		ConnectionType: "udp",
		Address:        "0.0.0.0:14550",
	}

	a := New(cfg)

	if a == nil {
		t.Fatal("New should return non-nil adapter")
	}
	if a.cfg.ConnectionType != "udp" {
		t.Errorf("ConnectionType = %s, want 'udp'", a.cfg.ConnectionType)
	}
	if a.cfg.Address != "0.0.0.0:14550" {
		t.Errorf("Address = %s, want '0.0.0.0:14550'", a.cfg.Address)
	}
	if a.states == nil {
		t.Error("states map should be initialized")
	}
}

func TestAdapter_Name(t *testing.T) {
	a := New(config.MAVLinkConfig{})

	name := a.Name()

	if name != "mavlink" {
		t.Errorf("Name() = %s, want 'mavlink'", name)
	}
}

func TestAdapter_Stop_NilNode(t *testing.T) {
	a := New(config.MAVLinkConfig{})

	// Stop should not panic with nil node
	err := a.Stop()

	if err != nil {
		t.Errorf("Stop should not error with nil node: %v", err)
	}
}

func TestAdapter_buildEndpoints_UDP(t *testing.T) {
	a := New(config.MAVLinkConfig{
		ConnectionType: "udp",
		Address:        "0.0.0.0:14550",
	})

	endpoints, err := a.buildEndpoints()

	if err != nil {
		t.Errorf("buildEndpoints should not error for UDP: %v", err)
	}
	if len(endpoints) != 1 {
		t.Errorf("Should return 1 endpoint, got %d", len(endpoints))
	}
}

func TestAdapter_buildEndpoints_TCP(t *testing.T) {
	a := New(config.MAVLinkConfig{
		ConnectionType: "tcp",
		Address:        "0.0.0.0:14550",
	})

	endpoints, err := a.buildEndpoints()

	if err != nil {
		t.Errorf("buildEndpoints should not error for TCP: %v", err)
	}
	if len(endpoints) != 1 {
		t.Errorf("Should return 1 endpoint, got %d", len(endpoints))
	}
}

func TestAdapter_buildEndpoints_Serial(t *testing.T) {
	a := New(config.MAVLinkConfig{
		ConnectionType: "serial",
		SerialPort:     "/dev/ttyUSB0",
		SerialBaud:     57600,
	})

	endpoints, err := a.buildEndpoints()

	if err != nil {
		t.Errorf("buildEndpoints should not error for Serial: %v", err)
	}
	if len(endpoints) != 1 {
		t.Errorf("Should return 1 endpoint, got %d", len(endpoints))
	}
}

func TestAdapter_buildEndpoints_Unknown(t *testing.T) {
	a := New(config.MAVLinkConfig{
		ConnectionType: "unknown",
	})

	_, err := a.buildEndpoints()

	if err == nil {
		t.Error("buildEndpoints should error for unknown connection type")
	}
}

func TestMapCopterMode(t *testing.T) {
	tests := []struct {
		name       string
		customMode uint32
		want       models.FlightMode
	}{
		{"stabilize", copterModeStabilize, models.FlightModeStabilize},
		{"acro", copterModeAcro, models.FlightModeManual},
		{"alt_hold", copterModeAltHold, models.FlightModeAltHold},
		{"auto", copterModeAuto, models.FlightModeAuto},
		{"guided", copterModeGuided, models.FlightModeGuided},
		{"guided_no_gps", copterModeGuidedNoGPS, models.FlightModeGuided},
		{"loiter", copterModeLoiter, models.FlightModeLoiter},
		{"poshold", copterModePosHold, models.FlightModeLoiter},
		{"rtl", copterModeRTL, models.FlightModeRTL},
		{"smart_rtl", copterModeSmartRTL, models.FlightModeRTL},
		{"land", copterModeLand, models.FlightModeLand},
		{"brake", copterModeBrake, models.FlightModeLoiter},
		{"unknown_mode", 999, models.FlightModeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapCopterMode(tt.customMode)
			if got != tt.want {
				t.Errorf("mapCopterMode(%d) = %s, want %s", tt.customMode, got, tt.want)
			}
		})
	}
}

func TestMapPlaneMode(t *testing.T) {
	tests := []struct {
		name       string
		customMode uint32
		want       models.FlightMode
	}{
		{"manual", planeModeManual, models.FlightModeManual},
		{"stabilize", planeModeStabilize, models.FlightModeStabilize},
		{"training", planeModeTraining, models.FlightModeStabilize},
		{"fly_by_wire_a", planeModeFlyByWireA, models.FlightModeStabilize},
		{"fly_by_wire_b", planeModeFlyByWireB, models.FlightModeStabilize},
		{"auto", planeModeAuto, models.FlightModeAuto},
		{"guided", planeModeGuided, models.FlightModeGuided},
		{"loiter", planeModeLoiter, models.FlightModeLoiter},
		{"circle", planeModeCircle, models.FlightModeLoiter},
		{"qloiter", planeModeQLoiter, models.FlightModeLoiter},
		{"qhover", planeModeQHover, models.FlightModeLoiter},
		{"rtl", planeModeRTL, models.FlightModeRTL},
		{"qrtl", planeModeQRTL, models.FlightModeRTL},
		{"takeoff", planeModeTakeoff, models.FlightModeTakeoff},
		{"qland", planeModeQLand, models.FlightModeLand},
		{"unknown_mode", 999, models.FlightModeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapPlaneMode(tt.customMode)
			if got != tt.want {
				t.Errorf("mapPlaneMode(%d) = %s, want %s", tt.customMode, got, tt.want)
			}
		})
	}
}

func TestMapFlightMode_Copter(t *testing.T) {
	copterTypes := []ardupilotmega.MAV_TYPE{
		ardupilotmega.MAV_TYPE_QUADROTOR,
		ardupilotmega.MAV_TYPE_HEXAROTOR,
		ardupilotmega.MAV_TYPE_OCTOROTOR,
		ardupilotmega.MAV_TYPE_TRICOPTER,
		ardupilotmega.MAV_TYPE_COAXIAL,
		ardupilotmega.MAV_TYPE_HELICOPTER,
	}

	for _, vType := range copterTypes {
		t.Run(vType.String(), func(t *testing.T) {
			got := mapFlightMode(copterModeAuto, vType)
			if got != models.FlightModeAuto {
				t.Errorf("mapFlightMode(auto, %s) = %s, want AUTO", vType.String(), got)
			}
		})
	}
}

func TestMapFlightMode_Plane(t *testing.T) {
	planeTypes := []ardupilotmega.MAV_TYPE{
		ardupilotmega.MAV_TYPE_FIXED_WING,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER_DUOROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER_QUADROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TILTROTOR,
		ardupilotmega.MAV_TYPE_VTOL_FIXEDROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER,
	}

	for _, vType := range planeTypes {
		t.Run(vType.String(), func(t *testing.T) {
			got := mapFlightMode(planeModeAuto, vType)
			if got != models.FlightModeAuto {
				t.Errorf("mapFlightMode(auto, %s) = %s, want AUTO", vType.String(), got)
			}
		})
	}
}

func TestMapFlightMode_Unknown(t *testing.T) {
	// Unknown vehicle types should default to copter mapping
	got := mapFlightMode(copterModeAuto, ardupilotmega.MAV_TYPE_GENERIC)
	if got != models.FlightModeAuto {
		t.Errorf("mapFlightMode for unknown type = %s, want AUTO (copter default)", got)
	}
}

func TestAdapter_ConfigWithSerial(t *testing.T) {
	cfg := config.MAVLinkConfig{
		Enabled:        true,
		ConnectionType: "serial",
		SerialPort:     "/dev/ttyUSB0",
		SerialBaud:     57600,
	}

	a := New(cfg)

	if a.cfg.SerialPort != "/dev/ttyUSB0" {
		t.Errorf("SerialPort = %s, want '/dev/ttyUSB0'", a.cfg.SerialPort)
	}
	if a.cfg.SerialBaud != 57600 {
		t.Errorf("SerialBaud = %d, want 57600", a.cfg.SerialBaud)
	}
}
