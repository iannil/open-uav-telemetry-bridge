package statestore

import (
	"testing"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestStateStore(t *testing.T) {
	store := New()

	// Test empty store
	if store.Count() != 0 {
		t.Errorf("Expected empty store, got %d items", store.Count())
	}

	// Test Update and Get
	state := models.NewDroneState("uav-001", "mavlink")
	state.Location.Lat = 22.5431
	state.Location.Lon = 114.0579

	store.Update(state)

	if store.Count() != 1 {
		t.Errorf("Expected 1 item, got %d", store.Count())
	}

	retrieved := store.Get("uav-001")
	if retrieved == nil {
		t.Fatal("Expected to retrieve state, got nil")
	}
	if retrieved.Location.Lat != 22.5431 {
		t.Errorf("Lat mismatch: got %f, want 22.5431", retrieved.Location.Lat)
	}

	// Test Get non-existent
	if store.Get("uav-999") != nil {
		t.Error("Expected nil for non-existent device")
	}

	// Test Update existing
	state.Location.Lat = 23.0
	store.Update(state)

	retrieved = store.Get("uav-001")
	if retrieved.Location.Lat != 23.0 {
		t.Errorf("Lat not updated: got %f, want 23.0", retrieved.Location.Lat)
	}

	// Test GetAll
	state2 := models.NewDroneState("uav-002", "dji")
	store.Update(state2)

	all := store.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 items, got %d", len(all))
	}

	// Test Delete
	store.Delete("uav-001")
	if store.Count() != 1 {
		t.Errorf("Expected 1 item after delete, got %d", store.Count())
	}
	if store.Get("uav-001") != nil {
		t.Error("Expected nil after delete")
	}
}

func TestStateStoreConcurrency(t *testing.T) {
	store := New()
	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 100; i++ {
		go func(id int) {
			state := models.NewDroneState("uav-concurrent", "mavlink")
			state.Location.Lat = float64(id)
			store.Update(state)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		go func() {
			_ = store.Get("uav-concurrent")
			_ = store.GetAll()
			_ = store.Count()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 200; i++ {
		<-done
	}

	// Should have exactly 1 device
	if store.Count() != 1 {
		t.Errorf("Expected 1 device, got %d", store.Count())
	}
}
