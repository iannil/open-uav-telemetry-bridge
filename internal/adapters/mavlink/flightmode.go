package mavlink

import (
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

// ArduPilot Copter flight modes
// https://ardupilot.org/copter/docs/flight-modes.html
const (
	copterModeStabilize   = 0
	copterModeAcro        = 1
	copterModeAltHold     = 2
	copterModeAuto        = 3
	copterModeGuided      = 4
	copterModeLoiter      = 5
	copterModeRTL         = 6
	copterModeCircle      = 7
	copterModeLand        = 9
	copterModeDrift       = 11
	copterModeSport       = 13
	copterModeFlip        = 14
	copterModeAutoTune    = 15
	copterModePosHold     = 16
	copterModeBrake       = 17
	copterModeThrow       = 18
	copterModeAvoidADSB   = 19
	copterModeGuidedNoGPS = 20
	copterModeSmartRTL    = 21
	copterModeFlowHold    = 22
	copterModeFollow      = 23
	copterModeZigZag      = 24
	copterModeSystemID    = 25
	copterModeAutoRotate  = 26
	copterModeTurtle      = 27
)

// ArduPilot Plane flight modes
const (
	planeModeManual       = 0
	planeModeCircle       = 1
	planeModeStabilize    = 2
	planeModeTraining     = 3
	planeModeAcro         = 4
	planeModeFlyByWireA   = 5
	planeModeFlyByWireB   = 6
	planeModeCruise       = 7
	planeModeAutoTune     = 8
	planeModeAuto         = 10
	planeModeRTL          = 11
	planeModeLoiter       = 12
	planeModeTakeoff      = 13
	planeModeAvoidADSB    = 14
	planeModeGuided       = 15
	planeModeInitializing = 16
	planeModeQStabilize   = 17
	planeModeQHover       = 18
	planeModeQLoiter      = 19
	planeModeQLand        = 20
	planeModeQRTL         = 21
	planeModeQAutoTune    = 22
	planeModeQAcro        = 23
	planeModeThermal      = 24
)

// mapFlightMode converts MAVLink custom_mode to unified FlightMode
func mapFlightMode(customMode uint32, vehicleType ardupilotmega.MAV_TYPE) models.FlightMode {
	switch vehicleType {
	case ardupilotmega.MAV_TYPE_QUADROTOR,
		ardupilotmega.MAV_TYPE_HEXAROTOR,
		ardupilotmega.MAV_TYPE_OCTOROTOR,
		ardupilotmega.MAV_TYPE_TRICOPTER,
		ardupilotmega.MAV_TYPE_COAXIAL,
		ardupilotmega.MAV_TYPE_HELICOPTER:
		return mapCopterMode(customMode)

	case ardupilotmega.MAV_TYPE_FIXED_WING,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER_DUOROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER_QUADROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TILTROTOR,
		ardupilotmega.MAV_TYPE_VTOL_FIXEDROTOR,
		ardupilotmega.MAV_TYPE_VTOL_TAILSITTER:
		return mapPlaneMode(customMode)

	default:
		return mapCopterMode(customMode) // Default to copter mapping
	}
}

// mapCopterMode maps ArduCopter custom modes to unified FlightMode
func mapCopterMode(customMode uint32) models.FlightMode {
	switch customMode {
	case copterModeStabilize:
		return models.FlightModeStabilize
	case copterModeAcro:
		return models.FlightModeManual
	case copterModeAltHold:
		return models.FlightModeAltHold
	case copterModeAuto:
		return models.FlightModeAuto
	case copterModeGuided, copterModeGuidedNoGPS:
		return models.FlightModeGuided
	case copterModeLoiter, copterModePosHold:
		return models.FlightModeLoiter
	case copterModeRTL, copterModeSmartRTL:
		return models.FlightModeRTL
	case copterModeLand:
		return models.FlightModeLand
	case copterModeBrake:
		return models.FlightModeLoiter
	default:
		return models.FlightModeUnknown
	}
}

// mapPlaneMode maps ArduPlane custom modes to unified FlightMode
func mapPlaneMode(customMode uint32) models.FlightMode {
	switch customMode {
	case planeModeManual:
		return models.FlightModeManual
	case planeModeStabilize, planeModeTraining, planeModeFlyByWireA, planeModeFlyByWireB:
		return models.FlightModeStabilize
	case planeModeAuto:
		return models.FlightModeAuto
	case planeModeGuided:
		return models.FlightModeGuided
	case planeModeLoiter, planeModeCircle, planeModeQLoiter, planeModeQHover:
		return models.FlightModeLoiter
	case planeModeRTL, planeModeQRTL:
		return models.FlightModeRTL
	case planeModeTakeoff:
		return models.FlightModeTakeoff
	case planeModeQLand:
		return models.FlightModeLand
	default:
		return models.FlightModeUnknown
	}
}
