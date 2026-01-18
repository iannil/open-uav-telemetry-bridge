// Package api provides HTTP REST API for the telemetry gateway
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/open-uav/telemetry-bridge/internal/config"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// StateProvider is an interface for getting drone states
type StateProvider interface {
	GetState(deviceID string) *models.DroneState
	GetAllStates() []*models.DroneState
	GetDeviceCount() int
}

// Server is the HTTP API server
type Server struct {
	cfg      config.HTTPConfig
	provider StateProvider
	server   *http.Server
	router   *chi.Mux
	version  string
	started  time.Time
}

// New creates a new HTTP API server
func New(cfg config.HTTPConfig, provider StateProvider, version string) *Server {
	s := &Server{
		cfg:      cfg,
		provider: provider,
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
	ActiveDrones int `json:"active_drones"`
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
			ActiveDrones: s.provider.GetDeviceCount(),
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

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
