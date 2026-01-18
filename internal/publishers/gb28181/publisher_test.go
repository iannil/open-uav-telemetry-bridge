package gb28181

import (
	"context"
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
	gbxml "github.com/open-uav/telemetry-bridge/internal/publishers/gb28181/xml"
)

func TestDigestAuth_ParseChallenge(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantErr bool
	}{
		{
			name:    "valid challenge",
			header:  `Digest realm="test-realm", nonce="abc123", qop="auth"`,
			wantErr: false,
		},
		{
			name:    "valid challenge without qop",
			header:  `Digest realm="test-realm", nonce="abc123"`,
			wantErr: false,
		},
		{
			name:    "missing realm",
			header:  `Digest nonce="abc123"`,
			wantErr: true,
		},
		{
			name:    "missing nonce",
			header:  `Digest realm="test-realm"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh DigestAuth for each test case
			auth := NewDigestAuth("user", "password")
			err := auth.ParseChallenge(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChallenge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDigestAuth_GenerateResponse(t *testing.T) {
	auth := NewDigestAuth("user", "password")

	// Parse a valid challenge first
	err := auth.ParseChallenge(`Digest realm="test-realm", nonce="abc123", qop="auth"`)
	if err != nil {
		t.Fatalf("ParseChallenge() failed: %v", err)
	}

	response := auth.GenerateResponse("REGISTER", "sip:server@domain")

	// Check that response contains required fields
	if response == "" {
		t.Error("GenerateResponse() returned empty string")
	}
	if !contains(response, `username="user"`) {
		t.Error("Response missing username")
	}
	if !contains(response, `realm="test-realm"`) {
		t.Error("Response missing realm")
	}
	if !contains(response, `nonce="abc123"`) {
		t.Error("Response missing nonce")
	}
	if !contains(response, `response="`) {
		t.Error("Response missing response hash")
	}
}

func TestDeviceManager_UpdateDrone(t *testing.T) {
	dm := NewDeviceManager("34020000001320000001")

	state := &models.DroneState{
		DeviceID:       "drone-001",
		ProtocolSource: "mavlink",
		Timestamp:      time.Now().UnixMilli(),
		Location: models.Location{
			Lat:     39.9087,
			Lon:     116.3975,
			AltGNSS: 100.0,
		},
	}

	// First update
	ch := dm.UpdateDrone(state)
	if ch == nil {
		t.Fatal("UpdateDrone() returned nil channel")
	}
	if ch.DroneID != "drone-001" {
		t.Errorf("Channel DroneID = %s, want drone-001", ch.DroneID)
	}
	if !ch.Online {
		t.Error("Channel should be online")
	}
	if len(ch.DeviceID) != 20 {
		t.Errorf("Channel DeviceID should be 20 digits, got %d", len(ch.DeviceID))
	}

	// Second update should return same channel
	ch2 := dm.UpdateDrone(state)
	if ch2.DeviceID != ch.DeviceID {
		t.Error("Second update should return same channel")
	}
}

func TestDeviceManager_GetAllChannels(t *testing.T) {
	dm := NewDeviceManager("34020000001320000001")

	// Add multiple drones
	for i := 0; i < 3; i++ {
		state := &models.DroneState{
			DeviceID:       "drone-" + string(rune('a'+i)),
			ProtocolSource: "mavlink",
			Timestamp:      time.Now().UnixMilli(),
		}
		dm.UpdateDrone(state)
	}

	channels := dm.GetAllChannels()
	if len(channels) != 3 {
		t.Errorf("GetAllChannels() returned %d channels, want 3", len(channels))
	}
}

func TestDeviceManager_MarkOffline(t *testing.T) {
	dm := NewDeviceManager("34020000001320000001")

	state := &models.DroneState{
		DeviceID:       "drone-001",
		ProtocolSource: "mavlink",
		Timestamp:      time.Now().UnixMilli(),
	}
	dm.UpdateDrone(state)

	// Channel should be online
	online := dm.GetOnlineChannels()
	if len(online) != 1 {
		t.Errorf("GetOnlineChannels() returned %d, want 1", len(online))
	}

	// Mark offline with 0 timeout (immediate)
	dm.MarkOffline(0)

	// Channel should now be offline
	online = dm.GetOnlineChannels()
	if len(online) != 0 {
		t.Errorf("After MarkOffline, GetOnlineChannels() returned %d, want 0", len(online))
	}
}

func TestSubscriptionManager_AddGet(t *testing.T) {
	sm := NewSubscriptionManager()

	sub := &Subscription{
		ID:        "sub-001",
		DeviceID:  "*",
		Interval:  5,
		Expires:   time.Now().Add(time.Hour),
		EventType: "presence",
	}

	sm.Add(sub)

	retrieved := sm.Get("sub-001")
	if retrieved == nil {
		t.Fatal("Get() returned nil for existing subscription")
	}
	if retrieved.ID != "sub-001" {
		t.Errorf("Get() returned wrong subscription ID: %s", retrieved.ID)
	}
}

func TestSubscriptionManager_GetActive(t *testing.T) {
	sm := NewSubscriptionManager()

	// Add active subscription
	sm.Add(&Subscription{
		ID:      "active",
		Expires: time.Now().Add(time.Hour),
	})

	// Add expired subscription
	sm.Add(&Subscription{
		ID:      "expired",
		Expires: time.Now().Add(-time.Hour),
	})

	active := sm.GetActive()
	if len(active) != 1 {
		t.Errorf("GetActive() returned %d subscriptions, want 1", len(active))
	}
	if active[0].ID != "active" {
		t.Errorf("GetActive() returned wrong subscription: %s", active[0].ID)
	}
}

func TestSubscriptionManager_Cleanup(t *testing.T) {
	sm := NewSubscriptionManager()

	// Add expired subscription
	sm.Add(&Subscription{
		ID:      "expired",
		Expires: time.Now().Add(-time.Hour),
	})

	// Add active subscription
	sm.Add(&Subscription{
		ID:      "active",
		Expires: time.Now().Add(time.Hour),
	})

	sm.Cleanup()

	if sm.Get("expired") != nil {
		t.Error("Cleanup() did not remove expired subscription")
	}
	if sm.Get("active") == nil {
		t.Error("Cleanup() incorrectly removed active subscription")
	}
}

func TestMobilePositionNotify(t *testing.T) {
	state := &models.DroneState{
		DeviceID:       "34020000001320000001",
		ProtocolSource: "mavlink",
		Timestamp:      1705639200000, // 2024-01-19T10:00:00
		Location: models.Location{
			Lat:     39.9087,
			Lon:     116.3975,
			AltGNSS: 100.5,
		},
		Velocity: models.Velocity{
			Vx: 3.0,
			Vy: 4.0,
			Vz: 0.0,
		},
		Attitude: models.Attitude{
			Yaw: 45.0,
		},
	}

	notify := gbxml.NewMobilePositionNotify(state, 1)

	if notify.DeviceID != state.DeviceID {
		t.Errorf("DeviceID = %s, want %s", notify.DeviceID, state.DeviceID)
	}
	if notify.Longitude != 116.3975 {
		t.Errorf("Longitude = %f, want 116.3975", notify.Longitude)
	}
	if notify.Latitude != 39.9087 {
		t.Errorf("Latitude = %f, want 39.9087", notify.Latitude)
	}
	if notify.Altitude != 100.5 {
		t.Errorf("Altitude = %f, want 100.5", notify.Altitude)
	}
	// Speed should be sqrt(3^2 + 4^2) = 5
	if notify.Speed != 5.0 {
		t.Errorf("Speed = %f, want 5.0", notify.Speed)
	}
	if notify.Direction != 45.0 {
		t.Errorf("Direction = %f, want 45.0", notify.Direction)
	}
}

func TestMobilePositionNotify_Marshal(t *testing.T) {
	state := &models.DroneState{
		DeviceID:  "34020000001320000001",
		Timestamp: 1705639200000,
		Location: models.Location{
			Lat:     39.9087,
			Lon:     116.3975,
			AltGNSS: 100.0,
		},
	}

	notify := gbxml.NewMobilePositionNotify(state, 1)
	xml, err := notify.Marshal()
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	if !contains(xml, `<?xml version="1.0" encoding="GB2312"?>`) {
		t.Error("XML missing declaration")
	}
	if !contains(xml, `<CmdType>MobilePosition</CmdType>`) {
		t.Error("XML missing CmdType")
	}
	if !contains(xml, `<DeviceID>34020000001320000001</DeviceID>`) {
		t.Error("XML missing DeviceID")
	}
}

func TestKeepaliveNotify(t *testing.T) {
	notify := gbxml.NewKeepaliveNotify("34020000001320000001", 1)

	if notify.DeviceID != "34020000001320000001" {
		t.Errorf("DeviceID = %s, want 34020000001320000001", notify.DeviceID)
	}
	if notify.Status != "OK" {
		t.Errorf("Status = %s, want OK", notify.Status)
	}
}

func TestCatalogResponse(t *testing.T) {
	items := []gbxml.CatalogItem{
		gbxml.NewCatalogItem("34020000001320000001", "UAV-001", "34020000002000000001", "340200", true),
		gbxml.NewCatalogItem("34020000001320000002", "UAV-002", "34020000002000000001", "340200", false),
	}

	resp := gbxml.NewCatalogResponse("34020000002000000001", 1, items)

	if resp.SumNum != 2 {
		t.Errorf("SumNum = %d, want 2", resp.SumNum)
	}
	if resp.DeviceList.Num != 2 {
		t.Errorf("DeviceList.Num = %d, want 2", resp.DeviceList.Num)
	}

	xml, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	if !contains(xml, `<CmdType>Catalog</CmdType>`) {
		t.Error("XML missing CmdType")
	}
	if !contains(xml, `<SumNum>2</SumNum>`) {
		t.Error("XML missing SumNum")
	}
}

func TestPublisher_New(t *testing.T) {
	cfg := config.GB28181Config{
		Enabled:           true,
		DeviceID:          "34020000001320000001",
		DeviceName:        "UAV-Gateway",
		LocalIP:           "192.168.1.100",
		LocalPort:         5060,
		ServerID:          "34020000002000000001",
		ServerIP:          "192.168.1.1",
		ServerPort:        5060,
		ServerDomain:      "3402000000",
		Username:          "34020000001320000001",
		Password:          "password",
		Transport:         "udp",
		RegisterExpires:   3600,
		HeartbeatInterval: 60,
		PositionInterval:  5,
	}

	pub := New(cfg)

	if pub.Name() != "gb28181" {
		t.Errorf("Name() = %s, want gb28181", pub.Name())
	}
}

func TestPublisher_PublishNotRunning(t *testing.T) {
	cfg := config.GB28181Config{
		DeviceID: "34020000001320000001",
	}
	pub := New(cfg)

	state := &models.DroneState{
		DeviceID: "drone-001",
	}

	err := pub.Publish(state)
	if err == nil {
		t.Error("Publish() should fail when publisher is not running")
	}
}

func TestPublisher_StopNotRunning(t *testing.T) {
	cfg := config.GB28181Config{
		DeviceID: "34020000001320000001",
	}
	pub := New(cfg)

	// Stop should not error when not running
	err := pub.Stop()
	if err != nil {
		t.Errorf("Stop() returned error when not running: %v", err)
	}
}

func TestPublisher_StartStop(t *testing.T) {
	// Skip if running in CI without network
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := config.GB28181Config{
		Enabled:           true,
		DeviceID:          "34020000001320000001",
		DeviceName:        "UAV-Gateway",
		LocalIP:           "127.0.0.1",
		LocalPort:         15060, // Use non-standard port to avoid conflicts
		ServerID:          "34020000002000000001",
		ServerIP:          "127.0.0.1",
		ServerPort:        15061, // Non-existent server
		ServerDomain:      "3402000000",
		Username:          "34020000001320000001",
		Password:          "password",
		Transport:         "udp",
		RegisterExpires:   3600,
		HeartbeatInterval: 60,
		PositionInterval:  5,
	}

	pub := New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start will fail on registration since there's no server
	err := pub.Start(ctx)
	if err == nil {
		// If it somehow succeeded, make sure to stop
		pub.Stop()
		t.Log("Start succeeded (unexpected but not an error)")
	} else {
		// Expected to fail since there's no server
		t.Logf("Start failed as expected: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
