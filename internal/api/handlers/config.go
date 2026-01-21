// Package handlers provides HTTP API handlers
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/open-uav/telemetry-bridge/internal/config"
	"gopkg.in/yaml.v3"
)

// ConfigHandler handles configuration API requests
type ConfigHandler struct {
	cfg         *config.Config
	configPath  string
	onConfigChange func(*config.Config)
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(cfg *config.Config, configPath string, onConfigChange func(*config.Config)) *ConfigHandler {
	return &ConfigHandler{
		cfg:            cfg,
		configPath:     configPath,
		onConfigChange: onConfigChange,
	}
}

// SanitizedConfig is the config response with sensitive data redacted
type SanitizedConfig struct {
	Server     config.ServerConfig     `json:"server"`
	MAVLink    config.MAVLinkConfig    `json:"mavlink"`
	DJI        config.DJIConfig        `json:"dji"`
	MQTT       SanitizedMQTTConfig     `json:"mqtt"`
	GB28181    SanitizedGB28181Config  `json:"gb28181"`
	HTTP       SanitizedHTTPConfig     `json:"http"`
	Throttle   config.ThrottleConfig   `json:"throttle"`
	Coordinate config.CoordinateConfig `json:"coordinate"`
	Track      config.TrackConfig      `json:"track"`
}

// SanitizedMQTTConfig is MQTT config with password redacted
type SanitizedMQTTConfig struct {
	Enabled     bool              `json:"enabled"`
	Broker      string            `json:"broker"`
	ClientID    string            `json:"client_id"`
	TopicPrefix string            `json:"topic_prefix"`
	QoS         int               `json:"qos"`
	Username    string            `json:"username"`
	HasPassword bool              `json:"has_password"`
	LWT         config.LWTConfig  `json:"lwt"`
}

// SanitizedGB28181Config is GB28181 config with password redacted
type SanitizedGB28181Config struct {
	Enabled           bool   `json:"enabled"`
	DeviceID          string `json:"device_id"`
	DeviceName        string `json:"device_name"`
	LocalIP           string `json:"local_ip"`
	LocalPort         int    `json:"local_port"`
	ServerID          string `json:"server_id"`
	ServerIP          string `json:"server_ip"`
	ServerPort        int    `json:"server_port"`
	ServerDomain      string `json:"server_domain"`
	Username          string `json:"username"`
	HasPassword       bool   `json:"has_password"`
	Transport         string `json:"transport"`
	RegisterExpires   int    `json:"register_expires"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
	PositionInterval  int    `json:"position_interval"`
}

// SanitizedHTTPConfig is HTTP config with secrets redacted
type SanitizedHTTPConfig struct {
	Enabled      bool                `json:"enabled"`
	Address      string              `json:"address"`
	CORSEnabled  bool                `json:"cors_enabled"`
	CORSOrigins  []string            `json:"cors_origins"`
	WebUIEnabled bool                `json:"webui_enabled"`
	Auth         SanitizedAuthConfig `json:"auth"`
}

// SanitizedAuthConfig is auth config with secrets redacted
type SanitizedAuthConfig struct {
	Enabled          bool   `json:"enabled"`
	Username         string `json:"username"`
	HasPasswordHash  bool   `json:"has_password_hash"`
	HasJWTSecret     bool   `json:"has_jwt_secret"`
	TokenExpiryHours int    `json:"token_expiry_hours"`
}

// GetConfig returns the current configuration (sanitized)
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	sanitized := SanitizedConfig{
		Server:     h.cfg.Server,
		MAVLink:    h.cfg.MAVLink,
		DJI:        h.cfg.DJI,
		MQTT: SanitizedMQTTConfig{
			Enabled:     h.cfg.MQTT.Enabled,
			Broker:      h.cfg.MQTT.Broker,
			ClientID:    h.cfg.MQTT.ClientID,
			TopicPrefix: h.cfg.MQTT.TopicPrefix,
			QoS:         h.cfg.MQTT.QoS,
			Username:    h.cfg.MQTT.Username,
			HasPassword: h.cfg.MQTT.Password != "",
			LWT:         h.cfg.MQTT.LWT,
		},
		GB28181: SanitizedGB28181Config{
			Enabled:           h.cfg.GB28181.Enabled,
			DeviceID:          h.cfg.GB28181.DeviceID,
			DeviceName:        h.cfg.GB28181.DeviceName,
			LocalIP:           h.cfg.GB28181.LocalIP,
			LocalPort:         h.cfg.GB28181.LocalPort,
			ServerID:          h.cfg.GB28181.ServerID,
			ServerIP:          h.cfg.GB28181.ServerIP,
			ServerPort:        h.cfg.GB28181.ServerPort,
			ServerDomain:      h.cfg.GB28181.ServerDomain,
			Username:          h.cfg.GB28181.Username,
			HasPassword:       h.cfg.GB28181.Password != "",
			Transport:         h.cfg.GB28181.Transport,
			RegisterExpires:   h.cfg.GB28181.RegisterExpires,
			HeartbeatInterval: h.cfg.GB28181.HeartbeatInterval,
			PositionInterval:  h.cfg.GB28181.PositionInterval,
		},
		HTTP: SanitizedHTTPConfig{
			Enabled:      h.cfg.HTTP.Enabled,
			Address:      h.cfg.HTTP.Address,
			CORSEnabled:  h.cfg.HTTP.CORSEnabled,
			CORSOrigins:  h.cfg.HTTP.CORSOrigins,
			WebUIEnabled: h.cfg.HTTP.WebUIEnabled,
			Auth: SanitizedAuthConfig{
				Enabled:          h.cfg.HTTP.Auth.Enabled,
				Username:         h.cfg.HTTP.Auth.Username,
				HasPasswordHash:  h.cfg.HTTP.Auth.PasswordHash != "",
				HasJWTSecret:     h.cfg.HTTP.Auth.JWTSecret != "",
				TokenExpiryHours: h.cfg.HTTP.Auth.TokenExpiryHours,
			},
		},
		Throttle:   h.cfg.Throttle,
		Coordinate: h.cfg.Coordinate,
		Track:      h.cfg.Track,
	}

	writeJSON(w, http.StatusOK, sanitized)
}

// UpdateMAVLinkConfig updates the MAVLink adapter configuration
func (h *ConfigHandler) UpdateMAVLinkConfig(w http.ResponseWriter, r *http.Request) {
	var update config.MAVLinkConfig
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if update.ConnectionType != "" &&
	   update.ConnectionType != "udp" &&
	   update.ConnectionType != "tcp" &&
	   update.ConnectionType != "serial" {
		writeError(w, http.StatusBadRequest, "invalid connection_type: must be udp, tcp, or serial")
		return
	}

	h.cfg.MAVLink = update
	writeJSON(w, http.StatusOK, map[string]string{"message": "MAVLink configuration updated"})
}

// UpdateDJIConfig updates the DJI adapter configuration
func (h *ConfigHandler) UpdateDJIConfig(w http.ResponseWriter, r *http.Request) {
	var update config.DJIConfig
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if update.MaxClients < 0 {
		writeError(w, http.StatusBadRequest, "max_clients must be non-negative")
		return
	}

	h.cfg.DJI = update
	writeJSON(w, http.StatusOK, map[string]string{"message": "DJI configuration updated"})
}

// MQTTConfigUpdate is the request body for updating MQTT config
type MQTTConfigUpdate struct {
	Enabled     *bool             `json:"enabled,omitempty"`
	Broker      string            `json:"broker,omitempty"`
	ClientID    string            `json:"client_id,omitempty"`
	TopicPrefix string            `json:"topic_prefix,omitempty"`
	QoS         *int              `json:"qos,omitempty"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	LWT         *config.LWTConfig `json:"lwt,omitempty"`
}

// UpdateMQTTConfig updates the MQTT publisher configuration
func (h *ConfigHandler) UpdateMQTTConfig(w http.ResponseWriter, r *http.Request) {
	var update MQTTConfigUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Apply updates (only non-empty fields)
	if update.Enabled != nil {
		h.cfg.MQTT.Enabled = *update.Enabled
	}
	if update.Broker != "" {
		h.cfg.MQTT.Broker = update.Broker
	}
	if update.ClientID != "" {
		h.cfg.MQTT.ClientID = update.ClientID
	}
	if update.TopicPrefix != "" {
		h.cfg.MQTT.TopicPrefix = update.TopicPrefix
	}
	if update.QoS != nil {
		if *update.QoS < 0 || *update.QoS > 2 {
			writeError(w, http.StatusBadRequest, "qos must be 0, 1, or 2")
			return
		}
		h.cfg.MQTT.QoS = *update.QoS
	}
	if update.Username != "" {
		h.cfg.MQTT.Username = update.Username
	}
	if update.Password != "" {
		h.cfg.MQTT.Password = update.Password
	}
	if update.LWT != nil {
		h.cfg.MQTT.LWT = *update.LWT
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "MQTT configuration updated"})
}

// GB28181ConfigUpdate is the request body for updating GB28181 config
type GB28181ConfigUpdate struct {
	Enabled           *bool  `json:"enabled,omitempty"`
	DeviceID          string `json:"device_id,omitempty"`
	DeviceName        string `json:"device_name,omitempty"`
	LocalIP           string `json:"local_ip,omitempty"`
	LocalPort         *int   `json:"local_port,omitempty"`
	ServerID          string `json:"server_id,omitempty"`
	ServerIP          string `json:"server_ip,omitempty"`
	ServerPort        *int   `json:"server_port,omitempty"`
	ServerDomain      string `json:"server_domain,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	Transport         string `json:"transport,omitempty"`
	RegisterExpires   *int   `json:"register_expires,omitempty"`
	HeartbeatInterval *int   `json:"heartbeat_interval,omitempty"`
	PositionInterval  *int   `json:"position_interval,omitempty"`
}

// UpdateGB28181Config updates the GB28181 publisher configuration
func (h *ConfigHandler) UpdateGB28181Config(w http.ResponseWriter, r *http.Request) {
	var update GB28181ConfigUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Apply updates
	if update.Enabled != nil {
		h.cfg.GB28181.Enabled = *update.Enabled
	}
	if update.DeviceID != "" {
		h.cfg.GB28181.DeviceID = update.DeviceID
	}
	if update.DeviceName != "" {
		h.cfg.GB28181.DeviceName = update.DeviceName
	}
	if update.LocalIP != "" {
		h.cfg.GB28181.LocalIP = update.LocalIP
	}
	if update.LocalPort != nil {
		h.cfg.GB28181.LocalPort = *update.LocalPort
	}
	if update.ServerID != "" {
		h.cfg.GB28181.ServerID = update.ServerID
	}
	if update.ServerIP != "" {
		h.cfg.GB28181.ServerIP = update.ServerIP
	}
	if update.ServerPort != nil {
		h.cfg.GB28181.ServerPort = *update.ServerPort
	}
	if update.ServerDomain != "" {
		h.cfg.GB28181.ServerDomain = update.ServerDomain
	}
	if update.Username != "" {
		h.cfg.GB28181.Username = update.Username
	}
	if update.Password != "" {
		h.cfg.GB28181.Password = update.Password
	}
	if update.Transport != "" {
		if update.Transport != "udp" && update.Transport != "tcp" {
			writeError(w, http.StatusBadRequest, "transport must be udp or tcp")
			return
		}
		h.cfg.GB28181.Transport = update.Transport
	}
	if update.RegisterExpires != nil {
		h.cfg.GB28181.RegisterExpires = *update.RegisterExpires
	}
	if update.HeartbeatInterval != nil {
		h.cfg.GB28181.HeartbeatInterval = *update.HeartbeatInterval
	}
	if update.PositionInterval != nil {
		h.cfg.GB28181.PositionInterval = *update.PositionInterval
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "GB28181 configuration updated"})
}

// UpdateThrottleConfig updates the throttle configuration
func (h *ConfigHandler) UpdateThrottleConfig(w http.ResponseWriter, r *http.Request) {
	var update config.ThrottleConfig
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if update.MinRateHz <= 0 {
		writeError(w, http.StatusBadRequest, "min_rate_hz must be positive")
		return
	}
	if update.MaxRateHz < update.MinRateHz {
		writeError(w, http.StatusBadRequest, "max_rate_hz must be >= min_rate_hz")
		return
	}
	if update.DefaultRateHz < update.MinRateHz || update.DefaultRateHz > update.MaxRateHz {
		writeError(w, http.StatusBadRequest, "default_rate_hz must be between min and max")
		return
	}

	h.cfg.Throttle = update
	writeJSON(w, http.StatusOK, map[string]string{"message": "Throttle configuration updated"})
}

// UpdateCoordinateConfig updates the coordinate conversion configuration
func (h *ConfigHandler) UpdateCoordinateConfig(w http.ResponseWriter, r *http.Request) {
	var update config.CoordinateConfig
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.cfg.Coordinate = update
	writeJSON(w, http.StatusOK, map[string]string{"message": "Coordinate configuration updated"})
}

// UpdateTrackConfig updates the track storage configuration
func (h *ConfigHandler) UpdateTrackConfig(w http.ResponseWriter, r *http.Request) {
	var update config.TrackConfig
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if update.MaxPointsPerDrone < 0 {
		writeError(w, http.StatusBadRequest, "max_points_per_drone must be non-negative")
		return
	}
	if update.SampleIntervalMs < 0 {
		writeError(w, http.StatusBadRequest, "sample_interval_ms must be non-negative")
		return
	}

	h.cfg.Track = update
	writeJSON(w, http.StatusOK, map[string]string{"message": "Track configuration updated"})
}

// ExportConfig exports the current configuration as YAML
func (h *ConfigHandler) ExportConfig(w http.ResponseWriter, r *http.Request) {
	// Create a copy with sensitive data masked
	exportCfg := *h.cfg
	exportCfg.MQTT.Password = maskIfSet(h.cfg.MQTT.Password)
	exportCfg.GB28181.Password = maskIfSet(h.cfg.GB28181.Password)
	exportCfg.HTTP.Auth.PasswordHash = maskIfSet(h.cfg.HTTP.Auth.PasswordHash)
	exportCfg.HTTP.Auth.JWTSecret = maskIfSet(h.cfg.HTTP.Auth.JWTSecret)

	data, err := yaml.Marshal(exportCfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to marshal configuration")
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=config.yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ApplyConfig signals to apply configuration changes (requires restart for some settings)
func (h *ConfigHandler) ApplyConfig(w http.ResponseWriter, r *http.Request) {
	if h.onConfigChange != nil {
		h.onConfigChange(h.cfg)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Configuration changes noted. Some changes may require a restart to take effect.",
		"restart_required": true,
	})
}

// Helper functions

func maskIfSet(s string) string {
	if s != "" {
		return strings.Repeat("*", 8)
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
