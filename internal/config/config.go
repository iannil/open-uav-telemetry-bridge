package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	MAVLink    MAVLinkConfig    `yaml:"mavlink"`
	DJI        DJIConfig        `yaml:"dji"`
	MQTT       MQTTConfig       `yaml:"mqtt"`
	HTTP       HTTPConfig       `yaml:"http"`
	Throttle   ThrottleConfig   `yaml:"throttle"`
	Coordinate CoordinateConfig `yaml:"coordinate"`
	Track      TrackConfig      `yaml:"track"`
}

// ServerConfig contains server-level settings
type ServerConfig struct {
	LogLevel string `yaml:"log_level"`
}

// MAVLinkConfig contains MAVLink adapter settings
type MAVLinkConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ConnectionType string `yaml:"connection_type"` // udp, tcp, serial
	Address        string `yaml:"address"`         // For UDP/TCP: "host:port"
	SerialPort     string `yaml:"serial_port"`     // For serial: "/dev/ttyUSB0"
	SerialBaud     int    `yaml:"serial_baud"`     // For serial: 57600
}

// DJIConfig contains DJI forwarder adapter settings
type DJIConfig struct {
	Enabled       bool   `yaml:"enabled"`
	ListenAddress string `yaml:"listen_address"` // TCP listen address: "host:port"
	MaxClients    int    `yaml:"max_clients"`    // Maximum concurrent clients
}

// MQTTConfig contains MQTT publisher settings
type MQTTConfig struct {
	Enabled     bool      `yaml:"enabled"`
	Broker      string    `yaml:"broker"`
	ClientID    string    `yaml:"client_id"`
	TopicPrefix string    `yaml:"topic_prefix"`
	QoS         int       `yaml:"qos"`
	Username    string    `yaml:"username"`
	Password    string    `yaml:"password"`
	LWT         LWTConfig `yaml:"lwt"`
}

// LWTConfig contains Last Will and Testament settings
type LWTConfig struct {
	Enabled bool   `yaml:"enabled"`
	Topic   string `yaml:"topic"`
	Message string `yaml:"message"`
}

// ThrottleConfig contains frequency control settings
type ThrottleConfig struct {
	DefaultRateHz float64 `yaml:"default_rate_hz"`
	MinRateHz     float64 `yaml:"min_rate_hz"`
	MaxRateHz     float64 `yaml:"max_rate_hz"`
}

// HTTPConfig contains HTTP API server settings
type HTTPConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Address     string   `yaml:"address"`      // Listen address: "host:port"
	CORSEnabled bool     `yaml:"cors_enabled"` // Enable CORS support
	CORSOrigins []string `yaml:"cors_origins"` // Allowed origins for CORS
}

// CoordinateConfig contains coordinate conversion settings
type CoordinateConfig struct {
	ConvertGCJ02 bool `yaml:"convert_gcj02"` // Convert to GCJ02 (Amap, Tencent)
	ConvertBD09  bool `yaml:"convert_bd09"`  // Convert to BD09 (Baidu Maps)
}

// TrackConfig contains trajectory storage settings
type TrackConfig struct {
	Enabled           bool  `yaml:"enabled"`
	MaxPointsPerDrone int   `yaml:"max_points_per_drone"` // Maximum points per drone
	SampleIntervalMs  int64 `yaml:"sample_interval_ms"`   // Minimum sampling interval
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults
	if cfg.Server.LogLevel == "" {
		cfg.Server.LogLevel = "info"
	}
	if cfg.DJI.ListenAddress == "" {
		cfg.DJI.ListenAddress = "0.0.0.0:14560"
	}
	if cfg.DJI.MaxClients == 0 {
		cfg.DJI.MaxClients = 10
	}
	if cfg.Throttle.DefaultRateHz == 0 {
		cfg.Throttle.DefaultRateHz = 1.0
	}
	if cfg.Throttle.MinRateHz == 0 {
		cfg.Throttle.MinRateHz = 0.5
	}
	if cfg.Throttle.MaxRateHz == 0 {
		cfg.Throttle.MaxRateHz = 10.0
	}
	if cfg.HTTP.Address == "" {
		cfg.HTTP.Address = "0.0.0.0:8080"
	}
	if cfg.Track.MaxPointsPerDrone == 0 {
		cfg.Track.MaxPointsPerDrone = 10000
	}
	if cfg.Track.SampleIntervalMs == 0 {
		cfg.Track.SampleIntervalMs = 1000
	}

	return &cfg, nil
}
