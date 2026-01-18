package models

// DroneState represents the unified telemetry data model
// This is the core data structure that all protocol adapters convert to
type DroneState struct {
	DeviceID       string   `json:"device_id"`        // Unique device identifier
	Timestamp      int64    `json:"timestamp"`        // Unix timestamp in milliseconds
	ProtocolSource string   `json:"protocol_source"`  // Data source: mavlink, dji, gb28181
	Location       Location `json:"location"`         // Position data
	Attitude       Attitude `json:"attitude"`         // Orientation data
	Status         Status   `json:"status"`           // System status
	Velocity       Velocity `json:"velocity"`         // Velocity data
}

// Location contains position information
type Location struct {
	Lat              float64 `json:"lat"`               // Latitude in degrees
	Lon              float64 `json:"lon"`               // Longitude in degrees
	AltBaro          float64 `json:"alt_baro"`          // Barometric altitude in meters
	AltGNSS          float64 `json:"alt_gnss"`          // GNSS altitude in meters
	CoordinateSystem string  `json:"coordinate_system"` // WGS84, GCJ02, BD09
}

// Attitude contains orientation information
type Attitude struct {
	Roll  float64 `json:"roll"`  // Roll angle in radians
	Pitch float64 `json:"pitch"` // Pitch angle in radians
	Yaw   float64 `json:"yaw"`   // Yaw angle in degrees (0-360)
}

// Status contains system status information
type Status struct {
	BatteryPercent int        `json:"battery_percent"` // Battery level 0-100
	FlightMode     FlightMode `json:"flight_mode"`     // Unified flight mode
	Armed          bool       `json:"armed"`           // Whether motors are armed
	SignalQuality  int        `json:"signal_quality"`  // Signal strength 0-100
}

// Velocity contains velocity information
type Velocity struct {
	Vx float64 `json:"vx"` // Velocity in X (North) direction, m/s
	Vy float64 `json:"vy"` // Velocity in Y (East) direction, m/s
	Vz float64 `json:"vz"` // Velocity in Z (Down) direction, m/s
}

// FlightMode represents unified flight modes across different protocols
type FlightMode string

const (
	FlightModeUnknown    FlightMode = "UNKNOWN"
	FlightModeManual     FlightMode = "MANUAL"
	FlightModeStabilize  FlightMode = "STABILIZE"
	FlightModeAltHold    FlightMode = "ALT_HOLD"
	FlightModeLoiter     FlightMode = "LOITER"  // Position hold
	FlightModeAuto       FlightMode = "AUTO"    // Autonomous mission
	FlightModeGuided     FlightMode = "GUIDED"  // External control
	FlightModeRTL        FlightMode = "RTL"     // Return to launch
	FlightModeLand       FlightMode = "LAND"    // Landing
	FlightModeTakeoff    FlightMode = "TAKEOFF" // Taking off
	FlightModeEmergency  FlightMode = "EMERGENCY"
)

// NewDroneState creates a new DroneState with default values
func NewDroneState(deviceID string, protocolSource string) *DroneState {
	return &DroneState{
		DeviceID:       deviceID,
		ProtocolSource: protocolSource,
		Location: Location{
			CoordinateSystem: "WGS84",
		},
		Status: Status{
			FlightMode: FlightModeUnknown,
		},
	}
}
