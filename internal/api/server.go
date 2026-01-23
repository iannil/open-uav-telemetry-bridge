// Package api provides HTTP REST API for the telemetry gateway
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/open-uav/telemetry-bridge/internal/api/auth"
	"github.com/open-uav/telemetry-bridge/internal/api/handlers"
	"github.com/open-uav/telemetry-bridge/internal/api/ratelimit"
	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/core/alerter"
	"github.com/open-uav/telemetry-bridge/internal/core/geofence"
	"github.com/open-uav/telemetry-bridge/internal/core/logger"
	"github.com/open-uav/telemetry-bridge/internal/core/trackstore"
	"github.com/open-uav/telemetry-bridge/internal/models"
	"github.com/open-uav/telemetry-bridge/internal/web"
)

// StateProvider is an interface for getting drone states
type StateProvider interface {
	GetState(deviceID string) *models.DroneState
	GetAllStates() []*models.DroneState
	GetDeviceCount() int
	GetTrack(deviceID string, limit int, since int64) []trackstore.TrackPoint
	ClearTrack(deviceID string)
	GetTrackSize(deviceID string) int
	IsTrackEnabled() bool
	GetAdapterNames() []string
	GetPublisherNames() []string
}

// Server is the HTTP API server
type Server struct {
	cfg               config.HTTPConfig
	fullConfig        *config.Config
	configPath        string
	provider          StateProvider
	server            *http.Server
	router            *chi.Mux
	hub               *Hub
	version           string
	started           time.Time
	webUIEnabled      bool
	authEnabled       bool
	authManager       *auth.Manager
	configHandler     *handlers.ConfigHandler
	logBuffer         *logger.Buffer
	logsHandler       *handlers.LogsHandler
	alerter           *alerter.Alerter
	alertsHandler     *handlers.AlertsHandler
	geofenceEngine    *geofence.Engine
	geofencesHandler  *handlers.GeofencesHandler
}

// New creates a new HTTP API server
func New(cfg config.HTTPConfig, provider StateProvider, version string) *Server {
	return NewWithConfig(cfg, nil, "", provider, version)
}

// NewWithConfig creates a new HTTP API server with full configuration access
func NewWithConfig(cfg config.HTTPConfig, fullConfig *config.Config, configPath string, provider StateProvider, version string) *Server {
	s := &Server{
		cfg:          cfg,
		fullConfig:   fullConfig,
		configPath:   configPath,
		provider:     provider,
		hub:          NewHub(),
		version:      version,
		webUIEnabled: cfg.WebUIEnabled,
		authEnabled:  cfg.Auth.Enabled,
	}

	// Initialize auth manager if authentication is enabled
	if cfg.Auth.Enabled {
		s.authManager = auth.NewManager(
			cfg.Auth.Username,
			cfg.Auth.PasswordHash,
			cfg.Auth.JWTSecret,
			cfg.Auth.TokenExpiryHours,
		)
		log.Printf("[HTTP] Authentication enabled for user: %s", cfg.Auth.Username)
	}

	// Initialize config handler if full config is provided
	if fullConfig != nil {
		s.configHandler = handlers.NewConfigHandler(fullConfig, configPath, nil)
		log.Printf("[HTTP] Configuration management enabled")
	}

	// Initialize log buffer and handler (always enabled)
	logBufferSize := 1000 // Default value
	if fullConfig != nil && fullConfig.Server.LogBufferSize > 0 {
		logBufferSize = fullConfig.Server.LogBufferSize
	}
	s.logBuffer = logger.New(logBufferSize)
	s.logsHandler = handlers.NewLogsHandler(s.logBuffer)
	log.Printf("[HTTP] Log buffer enabled (capacity: %d)", logBufferSize)

	// Initialize alerter (always enabled)
	s.alerter = alerter.New(alerter.Config{MaxAlerts: 1000})
	s.alertsHandler = handlers.NewAlertsHandler(s.alerter)
	log.Printf("[HTTP] Alert system enabled")

	// Initialize geofence engine (always enabled)
	s.geofenceEngine = geofence.NewEngine(geofence.Config{MaxBreaches: 500})
	s.geofencesHandler = handlers.NewGeofencesHandler(s.geofenceEngine)
	log.Printf("[HTTP] Geofence system enabled")

	s.setupRouter()
	return s
}

// setupRouter configures the HTTP routes
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Rate Limiting
	if s.cfg.RateLimit.Enabled {
		requestsPerSec := s.cfg.RateLimit.RequestsPerSec
		if requestsPerSec <= 0 {
			requestsPerSec = 100 // Default: 100 requests per second
		}
		burstSize := s.cfg.RateLimit.BurstSize
		if burstSize <= 0 {
			burstSize = 200 // Default: burst size of 200
		}
		limiter := ratelimit.NewIPRateLimiter(requestsPerSec, burstSize)
		r.Use(ratelimit.Middleware(limiter))
		log.Printf("[HTTP] Rate limiting enabled (%.0f req/s, burst %d)", requestsPerSec, burstSize)
	}

	// CORS
	if s.cfg.CORSEnabled {
		origins := s.cfg.CORSOrigins
		if len(origins) == 0 {
			origins = []string{"*"}
		}
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes (always available, even when auth is disabled)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", s.handleLogin)
			r.Post("/logout", s.handleLogout)
			r.Get("/me", s.handleGetMe)
		})

		// Protected routes (conditionally apply auth middleware)
		r.Group(func(r chi.Router) {
			if s.authEnabled {
				r.Use(auth.Middleware(s.authManager))
			}
			r.Get("/status", s.handleStatus)
			r.Get("/drones", s.handleGetDrones)
			r.Get("/drones/{deviceID}", s.handleGetDrone)
			r.Get("/drones/{deviceID}/track", s.handleGetTrack)
			r.Delete("/drones/{deviceID}/track", s.handleDeleteTrack)

			// Configuration management routes (only if config handler is available)
			if s.configHandler != nil {
				r.Route("/config", func(r chi.Router) {
					r.Get("/", s.configHandler.GetConfig)
					r.Put("/adapters/mavlink", s.configHandler.UpdateMAVLinkConfig)
					r.Put("/adapters/dji", s.configHandler.UpdateDJIConfig)
					r.Put("/publishers/mqtt", s.configHandler.UpdateMQTTConfig)
					r.Put("/publishers/gb28181", s.configHandler.UpdateGB28181Config)
					r.Put("/throttle", s.configHandler.UpdateThrottleConfig)
					r.Put("/coordinate", s.configHandler.UpdateCoordinateConfig)
					r.Put("/track", s.configHandler.UpdateTrackConfig)
					r.Post("/apply", s.configHandler.ApplyConfig)
					r.Post("/export", s.configHandler.ExportConfig)
				})
			}

			// Logs routes (always enabled)
			if s.logsHandler != nil {
				r.Route("/logs", func(r chi.Router) {
					r.Get("/", s.logsHandler.GetLogs)
					r.Get("/stream", s.logsHandler.StreamLogs)
					r.Delete("/", s.logsHandler.ClearLogs)
				})
			}

			// Alerts routes (always enabled)
			if s.alertsHandler != nil {
				r.Route("/alerts", func(r chi.Router) {
					r.Get("/", s.alertsHandler.GetAlerts)
					r.Delete("/", s.alertsHandler.ClearAlerts)
					r.Get("/stats", s.alertsHandler.GetStats)
					r.Get("/{id}", s.alertsHandler.GetAlert)
					r.Post("/{id}/ack", s.alertsHandler.AcknowledgeAlert)

					// Rules sub-routes
					r.Route("/rules", func(r chi.Router) {
						r.Get("/", s.alertsHandler.GetRules)
						r.Post("/", s.alertsHandler.CreateRule)
						r.Get("/{id}", s.alertsHandler.GetRule)
						r.Put("/{id}", s.alertsHandler.UpdateRule)
						r.Delete("/{id}", s.alertsHandler.DeleteRule)
					})
				})
			}

			// Geofences routes (always enabled)
			if s.geofencesHandler != nil {
				r.Route("/geofences", func(r chi.Router) {
					r.Get("/", s.geofencesHandler.GetGeofences)
					r.Post("/", s.geofencesHandler.CreateGeofence)
					r.Get("/stats", s.geofencesHandler.GetStats)
					r.Get("/breaches", s.geofencesHandler.GetBreaches)
					r.Delete("/breaches", s.geofencesHandler.ClearBreaches)
					r.Get("/{id}", s.geofencesHandler.GetGeofence)
					r.Put("/{id}", s.geofencesHandler.UpdateGeofence)
					r.Delete("/{id}", s.geofencesHandler.DeleteGeofence)
				})
			}
		})

		// WebSocket endpoint (with optional auth)
		r.Group(func(r chi.Router) {
			if s.authEnabled {
				r.Use(auth.OptionalMiddleware(s.authManager))
			}
			r.Get("/ws", s.serveWs)
		})
	})

	// Health check
	r.Get("/health", s.handleHealth)

	// Web UI static file serving
	if s.webUIEnabled {
		fsys, err := web.GetFS()
		if err != nil {
			log.Printf("[HTTP] Failed to load Web UI: %v", err)
		} else {
			spaHandler := web.NewSPAHandler(fsys)
			r.NotFound(spaHandler.ServeHTTP)
			log.Printf("[HTTP] Web UI enabled")
		}
	}

	s.router = r
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	s.started = time.Now()
	s.server = &http.Server{
		Addr:         s.cfg.Address,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start WebSocket hub
	go s.hub.Run()

	go func() {
		if s.cfg.TLS.Enabled {
			log.Printf("[HTTP] HTTPS server listening on %s (TLS enabled)", s.cfg.Address)
			if err := s.server.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("[HTTP] Server error: %v", err)
			}
		} else {
			log.Printf("[HTTP] Server listening on %s", s.cfg.Address)
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[HTTP] Server error: %v", err)
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("[HTTP] Server shutting down...")
	return s.server.Shutdown(ctx)
}

// Response types

// StatusResponse is the response for /api/v1/status
type StatusResponse struct {
	Version       string          `json:"version"`
	UptimeSeconds int64           `json:"uptime_seconds"`
	Adapters      []AdapterStatus `json:"adapters"`
	Publishers    []string        `json:"publishers"`
	Stats         Stats           `json:"stats"`
}

// AdapterStatus represents adapter status in the response
type AdapterStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Stats represents gateway statistics
type Stats struct {
	ActiveDrones     int `json:"active_drones"`
	WebSocketClients int `json:"websocket_clients"`
}

// DronesResponse is the response for /api/v1/drones
type DronesResponse struct {
	Count  int                   `json:"count"`
	Drones []*models.DroneState `json:"drones"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error    string `json:"error"`
	DeviceID string `json:"device_id,omitempty"`
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Build adapter status list from provider
	adapterNames := s.provider.GetAdapterNames()
	adapters := make([]AdapterStatus, len(adapterNames))
	for i, name := range adapterNames {
		adapters[i] = AdapterStatus{
			Name:    name,
			Enabled: true, // All registered adapters are enabled
		}
	}

	resp := StatusResponse{
		Version:       s.version,
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
		Adapters:      adapters,
		Publishers:    s.provider.GetPublisherNames(),
		Stats: Stats{
			ActiveDrones:     s.provider.GetDeviceCount(),
			WebSocketClients: s.hub.ClientCount(),
		},
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetDrones(w http.ResponseWriter, r *http.Request) {
	drones := s.provider.GetAllStates()
	resp := DronesResponse{
		Count:  len(drones),
		Drones: drones,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetDrone(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")

	state := s.provider.GetState(deviceID)
	if state == nil {
		s.writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:    "drone not found",
			DeviceID: deviceID,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, state)
}

// TrackResponse is the response for /api/v1/drones/{deviceID}/track
type TrackResponse struct {
	DeviceID   string                  `json:"device_id"`
	Count      int                     `json:"count"`
	Points     []trackstore.TrackPoint `json:"points"`
	TotalSize  int                     `json:"total_size"`
}

func (s *Server) handleGetTrack(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")

	// Check if track storage is enabled
	if !s.provider.IsTrackEnabled() {
		s.writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:    "track storage is disabled",
			DeviceID: deviceID,
		})
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	sinceStr := r.URL.Query().Get("since")

	var limit int
	var since int64

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			s.writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid limit parameter",
			})
			return
		}
	}

	if sinceStr != "" {
		var err error
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil || since < 0 {
			s.writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid since parameter",
			})
			return
		}
	}

	points := s.provider.GetTrack(deviceID, limit, since)
	totalSize := s.provider.GetTrackSize(deviceID)

	s.writeJSON(w, http.StatusOK, TrackResponse{
		DeviceID:  deviceID,
		Count:     len(points),
		Points:    points,
		TotalSize: totalSize,
	})
}

func (s *Server) handleDeleteTrack(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")

	// Check if track storage is enabled
	if !s.provider.IsTrackEnabled() {
		s.writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:    "track storage is disabled",
			DeviceID: deviceID,
		})
		return
	}

	s.provider.ClearTrack(deviceID)

	w.WriteHeader(http.StatusNoContent)
}

// Authentication handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// If auth is not enabled, return auth status
	if !s.authEnabled {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"auth_enabled": false,
			"message":      "authentication is disabled",
		})
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
		return
	}

	// Validate credentials
	if err := s.authManager.ValidateCredentials(req.Username, req.Password); err != nil {
		s.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: "invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := s.authManager.GenerateToken(req.Username)
	if err != nil {
		log.Printf("[HTTP] Failed to generate token: %v", err)
		s.writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: "failed to generate token",
		})
		return
	}

	user := s.authManager.GetUser()
	s.writeJSON(w, http.StatusOK, auth.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// JWT is stateless, so logout is handled client-side by removing the token
	// This endpoint exists for API completeness
	s.writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	// If auth is not enabled, return anonymous user
	if !s.authEnabled {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"auth_enabled": false,
			"user": auth.User{
				Username: "anonymous",
				Role:     "admin",
			},
		})
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		s.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: "missing authorization header",
		})
		return
	}

	// Parse Bearer token
	parts := make([]string, 0)
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		parts = append(parts, "Bearer", authHeader[7:])
	}
	if len(parts) != 2 {
		s.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: "invalid authorization header format",
		})
		return
	}

	tokenInfo, err := s.authManager.ValidateToken(parts[1])
	if err != nil {
		if err == auth.ErrTokenExpired {
			s.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: "token has expired",
			})
		} else {
			s.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: "invalid token",
			})
		}
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"auth_enabled": true,
		"user": auth.User{
			Username: tokenInfo.Username,
			Role:     tokenInfo.Role,
		},
	})
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// BroadcastState sends a drone state update to all WebSocket clients
func (s *Server) BroadcastState(state *models.DroneState) {
	if s.hub != nil {
		s.hub.BroadcastState(state)
	}
}

// GetHub returns the WebSocket hub
func (s *Server) GetHub() *Hub {
	return s.hub
}

// GetLogBuffer returns the log buffer for integration with the global logger
func (s *Server) GetLogBuffer() *logger.Buffer {
	return s.logBuffer
}

// GetAlerter returns the alerter for integration with the engine
func (s *Server) GetAlerter() *alerter.Alerter {
	return s.alerter
}

// EvaluateAlerts checks a drone state against alert rules
func (s *Server) EvaluateAlerts(state *models.DroneState) {
	if s.alerter != nil {
		s.alerter.Evaluate(state)
	}
}

// GetGeofenceEngine returns the geofence engine for integration
func (s *Server) GetGeofenceEngine() *geofence.Engine {
	return s.geofenceEngine
}

// EvaluateGeofences checks a drone state against geofences
func (s *Server) EvaluateGeofences(state *models.DroneState) {
	if s.geofenceEngine != nil {
		s.geofenceEngine.Evaluate(state)
	}
}
