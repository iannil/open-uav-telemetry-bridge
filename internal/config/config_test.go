package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  log_level: debug

mavlink:
  enabled: true
  connection_type: udp
  address: "0.0.0.0:14550"

mqtt:
  enabled: true
  broker: "tcp://localhost:1883"
  client_id: "test-client"
  topic_prefix: "uav/test"
  qos: 1
  lwt:
    enabled: true
    topic: "uav/status"
    message: "offline"

throttle:
  default_rate_hz: 2.0
  min_rate_hz: 0.5
  max_rate_hz: 10.0
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("LogLevel: got %s, want debug", cfg.Server.LogLevel)
	}
	if !cfg.MAVLink.Enabled {
		t.Error("MAVLink should be enabled")
	}
	if cfg.MAVLink.Address != "0.0.0.0:14550" {
		t.Errorf("MAVLink Address: got %s, want 0.0.0.0:14550", cfg.MAVLink.Address)
	}
	if cfg.MQTT.ClientID != "test-client" {
		t.Errorf("MQTT ClientID: got %s, want test-client", cfg.MQTT.ClientID)
	}
	if cfg.Throttle.DefaultRateHz != 2.0 {
		t.Errorf("DefaultRateHz: got %f, want 2.0", cfg.Throttle.DefaultRateHz)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a minimal config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
mavlink:
  enabled: true
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify defaults
	if cfg.Server.LogLevel != "info" {
		t.Errorf("Default LogLevel: got %s, want info", cfg.Server.LogLevel)
	}
	if cfg.Throttle.DefaultRateHz != 1.0 {
		t.Errorf("Default DefaultRateHz: got %f, want 1.0", cfg.Throttle.DefaultRateHz)
	}
	if cfg.Throttle.MinRateHz != 0.5 {
		t.Errorf("Default MinRateHz: got %f, want 0.5", cfg.Throttle.MinRateHz)
	}
	if cfg.Throttle.MaxRateHz != 10.0 {
		t.Errorf("Default MaxRateHz: got %f, want 10.0", cfg.Throttle.MaxRateHz)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
