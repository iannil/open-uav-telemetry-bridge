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
	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/core/trackstore"
	"github.com/open-uav/telemetry-bridge/internal/models"
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
}

// Server is the HTTP API server
type Server struct {
	cfg      config.HTTPConfig
	provider StateProvider
	server   *http.Server
	router   *chi.Mux
	hub      *Hub
	version  string
	started  time.Time
}

// New creates a new HTTP API server
func New(cfg config.HTTPConfig, provider StateProvider, version string) *Server {
	s := &Server{
		cfg:      cfg,
		provider: provider,
		hub:      NewHub(),
		version:  version,
	}
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
		r.Get("/status", s.handleStatus)
		r.Get("/drones", s.handleGetDrones)
		r.Get("/drones/{deviceID}", s.handleGetDrone)
		r.Get("/drones/{deviceID}/track", s.handleGetTrack)
		r.Delete("/drones/{deviceID}/track", s.handleDeleteTrack)
		r.Get("/ws", s.serveWs) // WebSocket endpoint
	})

	// Health check
	r.Get("/health", s.handleHealth)

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
		log.Printf("[HTTP] Server listening on %s", s.cfg.Address)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP] Server error: %v", err)
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
	resp := StatusResponse{
		Version:       s.version,
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
		Adapters:      []AdapterStatus{}, // Would need to be passed from engine
		Publishers:    []string{},        // Would need to be passed from engine
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
