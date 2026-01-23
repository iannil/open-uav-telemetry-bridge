package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/core/trackstore"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// mockProvider implements StateProvider for testing
type mockProvider struct {
	states       map[string]*models.DroneState
	tracks       map[string][]trackstore.TrackPoint
	trackEnabled bool
	adapters     []string
	publishers   []string
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		states:       make(map[string]*models.DroneState),
		tracks:       make(map[string][]trackstore.TrackPoint),
		trackEnabled: true,
		adapters:     []string{},
		publishers:   []string{},
	}
}

func (m *mockProvider) GetState(deviceID string) *models.DroneState {
	return m.states[deviceID]
}

func (m *mockProvider) GetAllStates() []*models.DroneState {
	states := make([]*models.DroneState, 0, len(m.states))
	for _, s := range m.states {
		states = append(states, s)
	}
	return states
}

func (m *mockProvider) GetDeviceCount() int {
	return len(m.states)
}

func (m *mockProvider) GetTrack(deviceID string, limit int, since int64) []trackstore.TrackPoint {
	points, ok := m.tracks[deviceID]
	if !ok {
		return []trackstore.TrackPoint{}
	}

	// Apply since filter
	if since > 0 {
		filtered := make([]trackstore.TrackPoint, 0)
		for _, p := range points {
			if p.Timestamp >= since {
				filtered = append(filtered, p)
			}
		}
		points = filtered
	}

	// Apply limit
	if limit > 0 && len(points) > limit {
		points = points[len(points)-limit:]
	}

	return points
}

func (m *mockProvider) ClearTrack(deviceID string) {
	delete(m.tracks, deviceID)
}

func (m *mockProvider) GetTrackSize(deviceID string) int {
	return len(m.tracks[deviceID])
}

func (m *mockProvider) IsTrackEnabled() bool {
	return m.trackEnabled
}

func (m *mockProvider) GetAdapterNames() []string {
	return m.adapters
}

func (m *mockProvider) GetPublisherNames() []string {
	return m.publishers
}

func (m *mockProvider) addState(state *models.DroneState) {
	m.states[state.DeviceID] = state
}

func (m *mockProvider) addTrackPoint(deviceID string, point trackstore.TrackPoint) {
	m.tracks[deviceID] = append(m.tracks[deviceID], point)
}

func createTestServer() (*Server, *mockProvider) {
	provider := newMockProvider()
	cfg := config.HTTPConfig{
		Enabled:     true,
		Address:     "127.0.0.1:0",
		CORSEnabled: true,
		CORSOrigins: []string{"*"},
	}
	server := New(cfg, provider, "test-version")
	return server, provider
}

func TestHandleHealth(t *testing.T) {
	server, _ := createTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestHandleStatus(t *testing.T) {
	server, provider := createTestServer()

	// Add some test data
	provider.addState(&models.DroneState{
		DeviceID:  "test-001",
		Timestamp: time.Now().UnixMilli(),
	})

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", resp.Version)
	}

	if resp.Stats.ActiveDrones != 1 {
		t.Errorf("Expected 1 active drone, got %d", resp.Stats.ActiveDrones)
	}
}

func TestHandleStatusWithAdaptersAndPublishers(t *testing.T) {
	provider := newMockProvider()
	provider.adapters = []string{"mavlink", "dji"}
	provider.publishers = []string{"mqtt", "gb28181"}

	cfg := config.HTTPConfig{
		Enabled:     true,
		Address:     "127.0.0.1:0",
		CORSEnabled: true,
		CORSOrigins: []string{"*"},
	}
	server := New(cfg, provider, "test-version")

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify adapters
	if len(resp.Adapters) != 2 {
		t.Errorf("Expected 2 adapters, got %d", len(resp.Adapters))
	}
	expectedAdapters := map[string]bool{"mavlink": false, "dji": false}
	for _, a := range resp.Adapters {
		if _, ok := expectedAdapters[a.Name]; ok {
			expectedAdapters[a.Name] = true
		}
		if !a.Enabled {
			t.Errorf("Adapter %s should be enabled", a.Name)
		}
	}
	for name, found := range expectedAdapters {
		if !found {
			t.Errorf("Expected adapter %s not found", name)
		}
	}

	// Verify publishers
	if len(resp.Publishers) != 2 {
		t.Errorf("Expected 2 publishers, got %d", len(resp.Publishers))
	}
	expectedPublishers := map[string]bool{"mqtt": false, "gb28181": false}
	for _, p := range resp.Publishers {
		if _, ok := expectedPublishers[p]; ok {
			expectedPublishers[p] = true
		}
	}
	for name, found := range expectedPublishers {
		if !found {
			t.Errorf("Expected publisher %s not found", name)
		}
	}
}

func TestHandleGetDrones(t *testing.T) {
	server, provider := createTestServer()

	// Add test drones
	provider.addState(&models.DroneState{
		DeviceID:  "test-001",
		Timestamp: 1000,
		Location:  models.Location{Lat: 39.9, Lon: 116.4},
	})
	provider.addState(&models.DroneState{
		DeviceID:  "test-002",
		Timestamp: 2000,
		Location:  models.Location{Lat: 31.2, Lon: 121.5},
	})

	req := httptest.NewRequest("GET", "/api/v1/drones", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp DronesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Count != 2 {
		t.Errorf("Expected 2 drones, got %d", resp.Count)
	}

	if len(resp.Drones) != 2 {
		t.Errorf("Expected 2 drone objects, got %d", len(resp.Drones))
	}
}

func TestHandleGetDrone(t *testing.T) {
	server, provider := createTestServer()

	provider.addState(&models.DroneState{
		DeviceID:  "test-001",
		Timestamp: 1000,
		Location:  models.Location{Lat: 39.9, Lon: 116.4},
	})

	// Test existing drone
	req := httptest.NewRequest("GET", "/api/v1/drones/test-001", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var state models.DroneState
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if state.DeviceID != "test-001" {
		t.Errorf("Expected device_id 'test-001', got '%s'", state.DeviceID)
	}
}

func TestHandleGetDroneNotFound(t *testing.T) {
	server, _ := createTestServer()

	req := httptest.NewRequest("GET", "/api/v1/drones/nonexistent", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != "drone not found" {
		t.Errorf("Expected 'drone not found' error, got '%s'", resp.Error)
	}
}

func TestHandleGetTrack(t *testing.T) {
	server, provider := createTestServer()

	// Add track points
	for i := 1; i <= 5; i++ {
		provider.addTrackPoint("test-001", trackstore.TrackPoint{
			Timestamp: int64(i * 1000),
			Lat:       39.9 + float64(i)*0.001,
			Lon:       116.4 + float64(i)*0.001,
			Alt:       100 + float64(i)*10,
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/drones/test-001/track", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp TrackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Count != 5 {
		t.Errorf("Expected 5 points, got %d", resp.Count)
	}

	if resp.TotalSize != 5 {
		t.Errorf("Expected total_size 5, got %d", resp.TotalSize)
	}
}

func TestHandleGetTrackWithLimit(t *testing.T) {
	server, provider := createTestServer()

	// Add track points
	for i := 1; i <= 10; i++ {
		provider.addTrackPoint("test-001", trackstore.TrackPoint{
			Timestamp: int64(i * 1000),
			Lat:       39.9 + float64(i)*0.001,
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/drones/test-001/track?limit=3", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp TrackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Count != 3 {
		t.Errorf("Expected 3 points, got %d", resp.Count)
	}
}

func TestHandleGetTrackWithSince(t *testing.T) {
	server, provider := createTestServer()

	// Add track points
	for i := 1; i <= 5; i++ {
		provider.addTrackPoint("test-001", trackstore.TrackPoint{
			Timestamp: int64(i * 1000),
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/drones/test-001/track?since=3000", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp TrackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Count != 3 {
		t.Errorf("Expected 3 points (since=3000), got %d", resp.Count)
	}
}

func TestHandleGetTrackInvalidLimit(t *testing.T) {
	server, _ := createTestServer()

	req := httptest.NewRequest("GET", "/api/v1/drones/test-001/track?limit=abc", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGetTrackDisabled(t *testing.T) {
	server, provider := createTestServer()
	provider.trackEnabled = false

	req := httptest.NewRequest("GET", "/api/v1/drones/test-001/track", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestHandleDeleteTrack(t *testing.T) {
	server, provider := createTestServer()

	// Add track points
	provider.addTrackPoint("test-001", trackstore.TrackPoint{Timestamp: 1000})
	provider.addTrackPoint("test-001", trackstore.TrackPoint{Timestamp: 2000})

	if provider.GetTrackSize("test-001") != 2 {
		t.Fatal("Expected 2 track points before delete")
	}

	req := httptest.NewRequest("DELETE", "/api/v1/drones/test-001/track", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if provider.GetTrackSize("test-001") != 0 {
		t.Errorf("Expected 0 track points after delete, got %d", provider.GetTrackSize("test-001"))
	}
}

func TestHandleDeleteTrackDisabled(t *testing.T) {
	server, provider := createTestServer()
	provider.trackEnabled = false

	req := httptest.NewRequest("DELETE", "/api/v1/drones/test-001/track", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}
