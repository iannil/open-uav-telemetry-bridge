package geofence

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name        string
		maxBreaches int
		wantMax     int
	}{
		{"positive max", 100, 100},
		{"zero defaults to 500", 0, 500},
		{"negative defaults to 500", -5, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(Config{MaxBreaches: tt.maxBreaches})
			if e.maxBreaches != tt.wantMax {
				t.Errorf("maxBreaches = %d, want %d", e.maxBreaches, tt.wantMax)
			}
		})
	}
}

func TestEngine_AddGeofence(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       1000,
		AlertOnEnter: true,
		Enabled:      true,
	}

	err := e.AddGeofence(gf)
	if err != nil {
		t.Errorf("AddGeofence should not error: %v", err)
	}

	if gf.ID == "" {
		t.Error("Geofence should have generated ID")
	}
	if gf.CreatedAt == 0 {
		t.Error("Geofence should have CreatedAt timestamp")
	}

	// Verify it was added
	retrieved, err := e.GetGeofence(gf.ID)
	if err != nil {
		t.Errorf("GetGeofence should not error: %v", err)
	}
	if retrieved.Name != "Test Zone" {
		t.Error("Retrieved geofence should match")
	}
}

func TestEngine_UpdateGeofence(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:    "Original",
		Type:    GeofenceTypeCircle,
		Center:  []float64{39.9087, 116.3975},
		Radius:  1000,
		Enabled: true,
	}
	e.AddGeofence(gf)

	gf.Name = "Updated"
	err := e.UpdateGeofence(gf)
	if err != nil {
		t.Errorf("UpdateGeofence should not error: %v", err)
	}

	retrieved, _ := e.GetGeofence(gf.ID)
	if retrieved.Name != "Updated" {
		t.Error("Geofence should be updated")
	}
}

func TestEngine_UpdateGeofence_NotFound(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{ID: "nonexistent"}
	err := e.UpdateGeofence(gf)
	if err != ErrGeofenceNotFound {
		t.Errorf("Should return ErrGeofenceNotFound, got %v", err)
	}
}

func TestEngine_DeleteGeofence(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{Name: "Test", Type: GeofenceTypeCircle, Center: []float64{0, 0}, Radius: 100}
	e.AddGeofence(gf)

	err := e.DeleteGeofence(gf.ID)
	if err != nil {
		t.Errorf("DeleteGeofence should not error: %v", err)
	}

	_, err = e.GetGeofence(gf.ID)
	if err != ErrGeofenceNotFound {
		t.Error("Deleted geofence should not be found")
	}
}

func TestEngine_DeleteGeofence_NotFound(t *testing.T) {
	e := NewEngine(Config{})

	err := e.DeleteGeofence("nonexistent")
	if err != ErrGeofenceNotFound {
		t.Errorf("Should return ErrGeofenceNotFound, got %v", err)
	}
}

func TestEngine_GetGeofences(t *testing.T) {
	e := NewEngine(Config{})

	e.AddGeofence(&Geofence{Name: "Zone 1", Type: GeofenceTypeCircle, Center: []float64{0, 0}, Radius: 100})
	e.AddGeofence(&Geofence{Name: "Zone 2", Type: GeofenceTypeCircle, Center: []float64{1, 1}, Radius: 100})

	gfs := e.GetGeofences()
	if len(gfs) != 2 {
		t.Errorf("Should have 2 geofences, got %d", len(gfs))
	}
}

func TestEngine_Evaluate_CircleEnter(t *testing.T) {
	e := NewEngine(Config{})

	// Beijing coordinates
	gf := &Geofence{
		Name:         "Beijing Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000, // 5km radius
		AlertOnEnter: true,
		AlertOnExit:  true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Drone outside the zone initially (set initial state)
	outsideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{
			Lat: 40.0, // Far from center
			Lon: 117.0,
		},
	}
	e.Evaluate(outsideState) // Initialize state

	// Now drone enters the zone
	insideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{
			Lat: 39.9087, // At center
			Lon: 116.3975,
		},
	}
	breaches := e.Evaluate(insideState)

	if len(breaches) != 1 {
		t.Fatalf("Should detect 1 breach (enter), got %d", len(breaches))
	}
	if breaches[0].Type != BreachTypeEnter {
		t.Errorf("Breach type should be 'enter', got '%s'", breaches[0].Type)
	}
	if breaches[0].DeviceID != "drone-1" {
		t.Errorf("Breach device should be 'drone-1', got '%s'", breaches[0].DeviceID)
	}
}

func TestEngine_Evaluate_CircleExit(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		AlertOnExit:  true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Drone inside first
	insideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 39.9087, Lon: 116.3975},
	}
	e.Evaluate(insideState)

	// Now drone exits
	outsideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 40.5, Lon: 117.0}, // Far from center
	}
	breaches := e.Evaluate(outsideState)

	if len(breaches) != 1 {
		t.Fatalf("Should detect 1 breach (exit), got %d", len(breaches))
	}
	if breaches[0].Type != BreachTypeExit {
		t.Errorf("Breach type should be 'exit', got '%s'", breaches[0].Type)
	}
}

func TestEngine_Evaluate_Polygon(t *testing.T) {
	e := NewEngine(Config{})

	// Square polygon
	gf := &Geofence{
		Name: "Square Zone",
		Type: GeofenceTypePolygon,
		Coordinates: [][]float64{
			{0, 0},
			{0, 10},
			{10, 10},
			{10, 0},
		},
		AlertOnEnter: true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Set initial state outside
	outsideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 20, Lon: 20}, // Outside
	}
	e.Evaluate(outsideState)

	// Move inside
	insideState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 5, Lon: 5}, // Inside
	}
	breaches := e.Evaluate(insideState)

	if len(breaches) != 1 {
		t.Fatalf("Should detect polygon entry, got %d breaches", len(breaches))
	}
}

func TestEngine_Evaluate_AltitudeBounds(t *testing.T) {
	e := NewEngine(Config{})

	minAlt := 50.0
	maxAlt := 200.0
	gf := &Geofence{
		Name:         "Altitude Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       10000,
		MinAltitude:  &minAlt,
		MaxAltitude:  &maxAlt,
		AlertOnEnter: true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Drone at correct position but too low altitude
	lowState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 39.9087, Lon: 116.3975, AltGNSS: 30},
	}
	breaches := e.Evaluate(lowState)
	if len(breaches) != 0 {
		t.Error("Should not detect breach when altitude is below minimum")
	}

	// Now at correct altitude
	correctState := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 39.9087, Lon: 116.3975, AltGNSS: 100},
	}
	breaches = e.Evaluate(correctState)
	if len(breaches) != 1 {
		t.Error("Should detect breach when altitude is within bounds")
	}
}

func TestEngine_Evaluate_DisabledGeofence(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Disabled Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		Enabled:      false, // Disabled
	}
	e.AddGeofence(gf)

	// Initialize outside
	e.Evaluate(&models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 40.5, Lon: 117.0},
	})

	// Enter zone
	state := &models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 39.9087, Lon: 116.3975},
	}
	breaches := e.Evaluate(state)

	if len(breaches) != 0 {
		t.Error("Disabled geofence should not trigger breaches")
	}
}

func TestEngine_Evaluate_NoAlertOnEnter(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Exit Only Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: false, // Don't alert on enter
		AlertOnExit:  true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Initialize outside
	e.Evaluate(&models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 40.5, Lon: 117.0},
	})

	// Enter - should not alert
	breaches := e.Evaluate(&models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 39.9087, Lon: 116.3975},
	})
	if len(breaches) != 0 {
		t.Error("Should not alert on enter when AlertOnEnter is false")
	}

	// Exit - should alert
	breaches = e.Evaluate(&models.DroneState{
		DeviceID: "drone-1",
		Location: models.Location{Lat: 40.5, Lon: 117.0},
	})
	if len(breaches) != 1 {
		t.Error("Should alert on exit when AlertOnExit is true")
	}
}

func TestEngine_GetBreaches(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Generate breach
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 40.5, Lon: 117.0}})
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 39.9087, Lon: 116.3975}})

	// Get all breaches
	breaches := e.GetBreaches("", "", 0)
	if len(breaches) == 0 {
		t.Error("Should have at least 1 breach")
	}

	// Filter by device
	breaches = e.GetBreaches("drone-1", "", 0)
	for _, b := range breaches {
		if b.DeviceID != "drone-1" {
			t.Error("Filtered breaches should only contain drone-1")
		}
	}

	// Filter by geofence
	breaches = e.GetBreaches("", gf.ID, 0)
	for _, b := range breaches {
		if b.GeofenceID != gf.ID {
			t.Error("Filtered breaches should only contain specific geofence")
		}
	}

	// Limit
	breaches = e.GetBreaches("", "", 1)
	if len(breaches) > 1 {
		t.Errorf("Limit should return max 1 breach, got %d", len(breaches))
	}
}

func TestEngine_ClearBreaches(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Generate breach
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 40.5, Lon: 117.0}})
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 39.9087, Lon: 116.3975}})

	breaches := e.GetBreaches("", "", 0)
	if len(breaches) == 0 {
		t.Fatal("Should have breaches before clear")
	}

	e.ClearBreaches()

	breaches = e.GetBreaches("", "", 0)
	if len(breaches) != 0 {
		t.Errorf("Should have no breaches after clear, got %d", len(breaches))
	}
}

func TestEngine_GetStats(t *testing.T) {
	e := NewEngine(Config{})

	e.AddGeofence(&Geofence{Name: "Zone 1", Type: GeofenceTypeCircle, Center: []float64{0, 0}, Radius: 100, Enabled: true})
	e.AddGeofence(&Geofence{Name: "Zone 2", Type: GeofenceTypeCircle, Center: []float64{1, 1}, Radius: 100, Enabled: false})

	stats := e.GetStats()

	if stats["total_geofences"].(int) != 2 {
		t.Errorf("total_geofences should be 2, got %v", stats["total_geofences"])
	}
	if stats["enabled_geofences"].(int) != 1 {
		t.Errorf("enabled_geofences should be 1, got %v", stats["enabled_geofences"])
	}
}

func TestEngine_SetBreachCallback(t *testing.T) {
	e := NewEngine(Config{})

	var called bool
	var mu sync.Mutex
	var receivedBreach *Breach

	e.SetBreachCallback(func(breach *Breach) {
		mu.Lock()
		defer mu.Unlock()
		called = true
		receivedBreach = breach
	})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Generate breach
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 40.5, Lon: 117.0}})
	e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 39.9087, Lon: 116.3975}})

	// Wait for async callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Error("Callback should be called")
	}
	if receivedBreach == nil {
		t.Error("Callback should receive breach")
	}
}

func TestEngine_MaxBreaches(t *testing.T) {
	e := NewEngine(Config{MaxBreaches: 3})

	gf := &Geofence{
		Name:         "Test Zone",
		Type:         GeofenceTypeCircle,
		Center:       []float64{39.9087, 116.3975},
		Radius:       5000,
		AlertOnEnter: true,
		AlertOnExit:  true,
		Enabled:      true,
	}
	e.AddGeofence(gf)

	// Generate many breaches by going in and out
	for i := 0; i < 5; i++ {
		e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 40.5, Lon: 117.0}})
		e.Evaluate(&models.DroneState{DeviceID: "drone-1", Location: models.Location{Lat: 39.9087, Lon: 116.3975}})
	}

	breaches := e.GetBreaches("", "", 0)
	if len(breaches) > 3 {
		t.Errorf("Should not exceed max breaches, got %d", len(breaches))
	}
}

func TestHaversineDistance(t *testing.T) {
	// Beijing to Shanghai approximate distance: ~1068 km
	beijing := []float64{39.9042, 116.4074}
	shanghai := []float64{31.2304, 121.4737}

	distance := haversineDistance(beijing[0], beijing[1], shanghai[0], shanghai[1])

	// Should be approximately 1068 km (1068000 meters) with some tolerance
	expectedMeters := 1068000.0
	tolerance := 50000.0 // 50km tolerance

	if math.Abs(distance-expectedMeters) > tolerance {
		t.Errorf("Distance between Beijing and Shanghai should be ~%v meters, got %v", expectedMeters, distance)
	}

	// Same point should have 0 distance
	samePoint := haversineDistance(beijing[0], beijing[1], beijing[0], beijing[1])
	if samePoint != 0 {
		t.Errorf("Distance to same point should be 0, got %v", samePoint)
	}
}

func TestEngine_IsInsideCircle(t *testing.T) {
	e := NewEngine(Config{})

	gf := &Geofence{
		Type:   GeofenceTypeCircle,
		Center: []float64{39.9087, 116.3975},
		Radius: 1000, // 1km
	}

	tests := []struct {
		name   string
		lat    float64
		lon    float64
		inside bool
	}{
		{"at center", 39.9087, 116.3975, true},
		{"within radius", 39.9087, 116.4000, true}, // ~200m from center
		{"outside radius", 39.95, 116.45, false},   // ~6km from center
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.isInsideCircle(tt.lat, tt.lon, gf)
			if result != tt.inside {
				t.Errorf("isInsideCircle(%v, %v) = %v, want %v", tt.lat, tt.lon, result, tt.inside)
			}
		})
	}
}

func TestEngine_IsInsidePolygon(t *testing.T) {
	e := NewEngine(Config{})

	// Triangle polygon
	gf := &Geofence{
		Type: GeofenceTypePolygon,
		Coordinates: [][]float64{
			{0, 0},
			{0, 10},
			{10, 0},
		},
	}

	tests := []struct {
		name   string
		lat    float64
		lon    float64
		inside bool
	}{
		{"inside triangle", 2, 2, true},
		{"outside triangle", 8, 8, false},
		{"on edge", 5, 5, false}, // On the hypotenuse, implementation may vary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.isInsidePolygon(tt.lat, tt.lon, gf)
			if result != tt.inside {
				t.Errorf("isInsidePolygon(%v, %v) = %v, want %v", tt.lat, tt.lon, result, tt.inside)
			}
		})
	}
}

func TestEngine_IsInsidePolygon_InvalidPolygon(t *testing.T) {
	e := NewEngine(Config{})

	// Too few points
	gf := &Geofence{
		Type: GeofenceTypePolygon,
		Coordinates: [][]float64{
			{0, 0},
			{0, 10},
		},
	}

	result := e.isInsidePolygon(5, 5, gf)
	if result {
		t.Error("Invalid polygon (< 3 points) should return false")
	}
}

func TestEngine_IsInsideCircle_InvalidCircle(t *testing.T) {
	e := NewEngine(Config{})

	// Missing center
	gf := &Geofence{
		Type:   GeofenceTypeCircle,
		Center: []float64{}, // Empty
		Radius: 1000,
	}

	result := e.isInsideCircle(0, 0, gf)
	if result {
		t.Error("Invalid circle (no center) should return false")
	}
}
