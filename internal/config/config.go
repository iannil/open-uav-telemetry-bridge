package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	MAVLink  MAVLinkConfig  `yaml:"mavlink"`
	MQTT     MQTTConfig     `yaml:"mqtt"`
	Throttle ThrottleConfig `yaml:"throttle"`
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
	if cfg.Throttle.DefaultRateHz == 0 {
		cfg.Throttle.DefaultRateHz = 1.0
	}
	if cfg.Throttle.MinRateHz == 0 {
		cfg.Throttle.MinRateHz = 0.5
	}
	if cfg.Throttle.MaxRateHz == 0 {
		cfg.Throttle.MaxRateHz = 10.0
	}

	return &cfg, nil
}
