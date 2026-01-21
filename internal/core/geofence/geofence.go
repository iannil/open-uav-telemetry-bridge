// Package geofence provides geofencing functionality for drone tracking
package geofence

import (
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// GeofenceType represents the type of geofence
type GeofenceType string

const (
	GeofenceTypePolygon GeofenceType = "polygon"
	GeofenceTypeCircle  GeofenceType = "circle"
)

// Geofence represents a geographic boundary
type Geofence struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         GeofenceType `json:"type"`
	Coordinates  [][]float64  `json:"coordinates,omitempty"`  // [[lat, lon], ...] for polygon
	Center       []float64    `json:"center,omitempty"`       // [lat, lon] for circle
	Radius       float64      `json:"radius,omitempty"`       // Radius in meters for circle
	MinAltitude  *float64     `json:"min_altitude,omitempty"` // Minimum altitude in meters
	MaxAltitude  *float64     `json:"max_altitude,omitempty"` // Maximum altitude in meters
	AlertOnEnter bool         `json:"alert_on_enter"`
	AlertOnExit  bool         `json:"alert_on_exit"`
	Enabled      bool         `json:"enabled"`
	CreatedAt    int64        `json:"created_at"`
	UpdatedAt    int64        `json:"updated_at"`
}

// BreachType represents the type of geofence breach
type BreachType string

const (
	BreachTypeEnter BreachType = "enter"
	BreachTypeExit  BreachType = "exit"
)

// Breach represents a geofence breach event
type Breach struct {
	ID         string     `json:"id"`
	GeofenceID string     `json:"geofence_id"`
	DeviceID   string     `json:"device_id"`
	Type       BreachType `json:"type"`
	Lat        float64    `json:"lat"`
	Lon        float64    `json:"lon"`
	Alt        float64    `json:"alt"`
	Timestamp  int64      `json:"timestamp"`
}

// Engine handles geofence management and breach detection
type Engine struct {
	geofences    map[string]*Geofence
	deviceStates map[string]map[string]bool // deviceID -> geofenceID -> inside
	breaches     []Breach
	maxBreaches  int
	onBreach     func(*Breach)
	mu           sync.RWMutex
}

// Config holds geofence engine configuration
type Config struct {
	MaxBreaches int // Maximum number of breaches to keep in memory
}

// NewEngine creates a new geofence engine
func NewEngine(cfg Config) *Engine {
	maxBreaches := cfg.MaxBreaches
	if maxBreaches <= 0 {
		maxBreaches = 500
	}

	return &Engine{
		geofences:    make(map[string]*Geofence),
		deviceStates: make(map[string]map[string]bool),
		breaches:     make([]Breach, 0),
		maxBreaches:  maxBreaches,
	}
}

// SetBreachCallback sets a callback function to be called when a breach occurs
func (e *Engine) SetBreachCallback(cb func(*Breach)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onBreach = cb
}

// AddGeofence adds a new geofence
func (e *Engine) AddGeofence(gf *Geofence) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if gf.ID == "" {
		gf.ID = uuid.New().String()
	}

	now := time.Now().UnixMilli()
	gf.CreatedAt = now
	gf.UpdatedAt = now

	e.geofences[gf.ID] = gf
	return nil
}

// UpdateGeofence updates an existing geofence
func (e *Engine) UpdateGeofence(gf *Geofence) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.geofences[gf.ID]; !ok {
		return ErrGeofenceNotFound
	}

	gf.UpdatedAt = time.Now().UnixMilli()
	e.geofences[gf.ID] = gf
	return nil
}

// DeleteGeofence removes a geofence
func (e *Engine) DeleteGeofence(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.geofences[id]; !ok {
		return ErrGeofenceNotFound
	}

	delete(e.geofences, id)

	// Clean up device states for this geofence
	for deviceID := range e.deviceStates {
		delete(e.deviceStates[deviceID], id)
	}

	return nil
}

// GetGeofence returns a geofence by ID
func (e *Engine) GetGeofence(id string) (*Geofence, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	gf, ok := e.geofences[id]
	if !ok {
		return nil, ErrGeofenceNotFound
	}
	return gf, nil
}

// GetGeofences returns all geofences
func (e *Engine) GetGeofences() []*Geofence {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Geofence, 0, len(e.geofences))
	for _, gf := range e.geofences {
		result = append(result, gf)
	}
	return result
}

// Evaluate checks if a drone state triggers any geofence breaches
func (e *Engine) Evaluate(state *models.DroneState) []*Breach {
	e.mu.Lock()
	defer e.mu.Unlock()

	var breaches []*Breach

	// Get or create device state map
	if e.deviceStates[state.DeviceID] == nil {
		e.deviceStates[state.DeviceID] = make(map[string]bool)
	}
	deviceState := e.deviceStates[state.DeviceID]

	for _, gf := range e.geofences {
		if !gf.Enabled {
			continue
		}

		// Check if drone is inside geofence
		inside := e.isInside(state, gf)
		wasInside := deviceState[gf.ID]

		// Check for breach
		var breach *Breach

		if inside && !wasInside && gf.AlertOnEnter {
			// Entered geofence
			breach = &Breach{
				ID:         uuid.New().String(),
				GeofenceID: gf.ID,
				DeviceID:   state.DeviceID,
				Type:       BreachTypeEnter,
				Lat:        state.Location.Lat,
				Lon:        state.Location.Lon,
				Alt:        state.Location.AltGNSS,
				Timestamp:  time.Now().UnixMilli(),
			}
		} else if !inside && wasInside && gf.AlertOnExit {
			// Exited geofence
			breach = &Breach{
				ID:         uuid.New().String(),
				GeofenceID: gf.ID,
				DeviceID:   state.DeviceID,
				Type:       BreachTypeExit,
				Lat:        state.Location.Lat,
				Lon:        state.Location.Lon,
				Alt:        state.Location.AltGNSS,
				Timestamp:  time.Now().UnixMilli(),
			}
		}

		// Update state
		deviceState[gf.ID] = inside

		if breach != nil {
			e.addBreach(breach)
			breaches = append(breaches, breach)

			// Call callback
			if e.onBreach != nil {
				go e.onBreach(breach)
			}
		}
	}

	return breaches
}

// isInside checks if a drone is inside a geofence
func (e *Engine) isInside(state *models.DroneState, gf *Geofence) bool {
	lat := state.Location.Lat
	lon := state.Location.Lon
	alt := state.Location.AltGNSS

	// Check altitude bounds
	if gf.MinAltitude != nil && alt < *gf.MinAltitude {
		return false
	}
	if gf.MaxAltitude != nil && alt > *gf.MaxAltitude {
		return false
	}

	switch gf.Type {
	case GeofenceTypeCircle:
		return e.isInsideCircle(lat, lon, gf)
	case GeofenceTypePolygon:
		return e.isInsidePolygon(lat, lon, gf)
	default:
		return false
	}
}

// isInsideCircle checks if a point is inside a circular geofence
func (e *Engine) isInsideCircle(lat, lon float64, gf *Geofence) bool {
	if len(gf.Center) < 2 {
		return false
	}

	centerLat := gf.Center[0]
	centerLon := gf.Center[1]

	// Calculate distance using Haversine formula
	distance := haversineDistance(lat, lon, centerLat, centerLon)

	return distance <= gf.Radius
}

// isInsidePolygon checks if a point is inside a polygon using ray casting
func (e *Engine) isInsidePolygon(lat, lon float64, gf *Geofence) bool {
	if len(gf.Coordinates) < 3 {
		return false
	}

	// Ray casting algorithm
	inside := false
	n := len(gf.Coordinates)

	for i := 0; i < n; i++ {
		j := (i + 1) % n

		yi := gf.Coordinates[i][0]
		xi := gf.Coordinates[i][1]
		yj := gf.Coordinates[j][0]
		xj := gf.Coordinates[j][1]

		if ((yi > lat) != (yj > lat)) &&
			(lon < (xj-xi)*(lat-yi)/(yj-yi)+xi) {
			inside = !inside
		}
	}

	return inside
}

// addBreach adds a breach to the history
func (e *Engine) addBreach(breach *Breach) {
	e.breaches = append(e.breaches, *breach)

	// Trim if over max
	if len(e.breaches) > e.maxBreaches {
		e.breaches = e.breaches[1:]
	}
}

// GetBreaches returns breach history
func (e *Engine) GetBreaches(deviceID string, geofenceID string, limit int) []Breach {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []Breach

	for i := len(e.breaches) - 1; i >= 0; i-- {
		breach := e.breaches[i]

		if deviceID != "" && breach.DeviceID != deviceID {
			continue
		}
		if geofenceID != "" && breach.GeofenceID != geofenceID {
			continue
		}

		result = append(result, breach)

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

// ClearBreaches removes all breach history
func (e *Engine) ClearBreaches() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.breaches = make([]Breach, 0)
}

// GetStats returns geofence engine statistics
func (e *Engine) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	enabled := 0
	for _, gf := range e.geofences {
		if gf.Enabled {
			enabled++
		}
	}

	return map[string]interface{}{
		"total_geofences":   len(e.geofences),
		"enabled_geofences": enabled,
		"total_breaches":    len(e.breaches),
		"tracked_devices":   len(e.deviceStates),
	}
}

// haversineDistance calculates the distance between two points on Earth in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// MarshalJSON for Geofence
func (gf *Geofence) MarshalJSON() ([]byte, error) {
	type alias Geofence
	return json.Marshal((*alias)(gf))
}

// Errors
var (
	ErrGeofenceNotFound = &GeofenceError{"geofence not found"}
)

type GeofenceError struct {
	msg string
}

func (e *GeofenceError) Error() string {
	return e.msg
}
