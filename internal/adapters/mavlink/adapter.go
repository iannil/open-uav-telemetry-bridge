package mavlink

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/frame"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// Adapter implements the core.Adapter interface for MAVLink protocol
type Adapter struct {
	cfg    config.MAVLinkConfig
	node   *gomavlib.Node
	mu     sync.RWMutex
	states map[uint8]*models.DroneState // keyed by system ID
}

// New creates a new MAVLink adapter
func New(cfg config.MAVLinkConfig) *Adapter {
	return &Adapter{
		cfg:    cfg,
		states: make(map[uint8]*models.DroneState),
	}
}

// Name returns the adapter name
func (a *Adapter) Name() string {
	return "mavlink"
}

// Start begins receiving MAVLink data and sends DroneState events to the channel
func (a *Adapter) Start(ctx context.Context, events chan<- *models.DroneState) error {
	endpoints, err := a.buildEndpoints()
	if err != nil {
		return fmt.Errorf("building endpoints: %w", err)
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints:   endpoints,
		Dialect:     ardupilotmega.Dialect,
		OutVersion:  gomavlib.V2,
		OutSystemID: 255, // GCS system ID
	})
	if err != nil {
		return fmt.Errorf("creating mavlink node: %w", err)
	}
	a.node = node

	go a.receiveLoop(ctx, events)

	return nil
}

// Stop gracefully stops the adapter
func (a *Adapter) Stop() error {
	if a.node != nil {
		a.node.Close()
	}
	return nil
}

// buildEndpoints creates the appropriate endpoint configuration
func (a *Adapter) buildEndpoints() ([]gomavlib.EndpointConf, error) {
	switch a.cfg.ConnectionType {
	case "udp":
		return []gomavlib.EndpointConf{
			gomavlib.EndpointUDPServer{Address: a.cfg.Address},
		}, nil
	case "tcp":
		return []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{Address: a.cfg.Address},
		}, nil
	case "serial":
		return []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: a.cfg.SerialPort,
				Baud:   a.cfg.SerialBaud,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown connection type: %s", a.cfg.ConnectionType)
	}
}

// receiveLoop processes incoming MAVLink messages
func (a *Adapter) receiveLoop(ctx context.Context, events chan<- *models.DroneState) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-a.node.Events():
			if frm, ok := evt.(*gomavlib.EventFrame); ok {
				a.handleFrame(frm.Frame, events)
			}
		}
	}
}

// handleFrame processes a single MAVLink frame
func (a *Adapter) handleFrame(frm frame.Frame, events chan<- *models.DroneState) {
	sysID := frm.GetSystemID()

	a.mu.Lock()
	state, exists := a.states[sysID]
	if !exists {
		state = models.NewDroneState(
			fmt.Sprintf("mavlink-%d", sysID),
			"mavlink",
		)
		a.states[sysID] = state
	}
	a.mu.Unlock()

	// Update timestamp
	state.Timestamp = time.Now().UnixMilli()

	// Process message based on type
	switch msg := frm.GetMessage().(type) {
	case *ardupilotmega.MessageHeartbeat:
		a.handleHeartbeat(state, msg)
	case *ardupilotmega.MessageGlobalPositionInt:
		a.handleGlobalPositionInt(state, msg)
	case *ardupilotmega.MessageAttitude:
		a.handleAttitude(state, msg)
	case *ardupilotmega.MessageSysStatus:
		a.handleSysStatus(state, msg)
	default:
		// Ignore other message types
		return
	}

	// Send state update
	select {
	case events <- state:
	default:
		// Channel full, skip this update
	}
}

// handleHeartbeat processes HEARTBEAT message
func (a *Adapter) handleHeartbeat(state *models.DroneState, msg *ardupilotmega.MessageHeartbeat) {
	state.Status.Armed = (msg.BaseMode & ardupilotmega.MAV_MODE_FLAG_SAFETY_ARMED) != 0
	state.Status.FlightMode = mapFlightMode(msg.CustomMode, msg.Type)
}

// handleGlobalPositionInt processes GLOBAL_POSITION_INT message
func (a *Adapter) handleGlobalPositionInt(state *models.DroneState, msg *ardupilotmega.MessageGlobalPositionInt) {
	state.Location.Lat = float64(msg.Lat) / 1e7
	state.Location.Lon = float64(msg.Lon) / 1e7
	state.Location.AltGNSS = float64(msg.Alt) / 1000.0
	state.Location.AltBaro = float64(msg.RelativeAlt) / 1000.0
	state.Location.CoordinateSystem = "WGS84"

	// Velocity from this message
	state.Velocity.Vx = float64(msg.Vx) / 100.0
	state.Velocity.Vy = float64(msg.Vy) / 100.0
	state.Velocity.Vz = float64(msg.Vz) / 100.0
}

// handleAttitude processes ATTITUDE message
func (a *Adapter) handleAttitude(state *models.DroneState, msg *ardupilotmega.MessageAttitude) {
	state.Attitude.Roll = float64(msg.Roll)
	state.Attitude.Pitch = float64(msg.Pitch)
	// Convert yaw from radians to degrees (0-360)
	yawDeg := float64(msg.Yaw) * 180.0 / 3.14159265359
	if yawDeg < 0 {
		yawDeg += 360.0
	}
	state.Attitude.Yaw = yawDeg
}

// handleSysStatus processes SYS_STATUS message
func (a *Adapter) handleSysStatus(state *models.DroneState, msg *ardupilotmega.MessageSysStatus) {
	if msg.BatteryRemaining >= 0 && msg.BatteryRemaining <= 100 {
		state.Status.BatteryPercent = int(msg.BatteryRemaining)
	}
}
