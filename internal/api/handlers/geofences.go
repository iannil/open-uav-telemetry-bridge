// Package handlers provides HTTP API handlers
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/open-uav/telemetry-bridge/internal/core/geofence"
)

// GeofencesHandler handles geofence-related API requests
type GeofencesHandler struct {
	engine *geofence.Engine
}

// NewGeofencesHandler creates a new geofences handler
func NewGeofencesHandler(engine *geofence.Engine) *GeofencesHandler {
	return &GeofencesHandler{
		engine: engine,
	}
}

// GetGeofences returns all geofences
func (h *GeofencesHandler) GetGeofences(w http.ResponseWriter, r *http.Request) {
	geofences := h.engine.GetGeofences()

	resp := map[string]interface{}{
		"geofences": geofences,
		"count":     len(geofences),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetGeofence returns a single geofence by ID
func (h *GeofencesHandler) GetGeofence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	gf, err := h.engine.GetGeofence(id)
	if err != nil {
		if err == geofence.ErrGeofenceNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "geofence not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, gf)
}

// CreateGeofenceRequest represents a create geofence request
type CreateGeofenceRequest struct {
	Name         string                `json:"name"`
	Type         geofence.GeofenceType `json:"type"`
	Coordinates  [][]float64           `json:"coordinates,omitempty"`
	Center       []float64             `json:"center,omitempty"`
	Radius       float64               `json:"radius,omitempty"`
	MinAltitude  *float64              `json:"min_altitude,omitempty"`
	MaxAltitude  *float64              `json:"max_altitude,omitempty"`
	AlertOnEnter bool                  `json:"alert_on_enter"`
	AlertOnExit  bool                  `json:"alert_on_exit"`
	Enabled      bool                  `json:"enabled"`
}

// CreateGeofence creates a new geofence
func (h *GeofencesHandler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
	var req CreateGeofenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	if req.Type != geofence.GeofenceTypePolygon && req.Type != geofence.GeofenceTypeCircle {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type must be 'polygon' or 'circle'"})
		return
	}

	if req.Type == geofence.GeofenceTypePolygon && len(req.Coordinates) < 3 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "polygon requires at least 3 coordinates"})
		return
	}

	if req.Type == geofence.GeofenceTypeCircle {
		if len(req.Center) < 2 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "circle requires center [lat, lon]"})
			return
		}
		if req.Radius <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "circle requires positive radius"})
			return
		}
	}

	gf := &geofence.Geofence{
		Name:         req.Name,
		Type:         req.Type,
		Coordinates:  req.Coordinates,
		Center:       req.Center,
		Radius:       req.Radius,
		MinAltitude:  req.MinAltitude,
		MaxAltitude:  req.MaxAltitude,
		AlertOnEnter: req.AlertOnEnter,
		AlertOnExit:  req.AlertOnExit,
		Enabled:      req.Enabled,
	}

	if err := h.engine.AddGeofence(gf); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, gf)
}

// UpdateGeofence updates an existing geofence
func (h *GeofencesHandler) UpdateGeofence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get existing geofence
	existing, err := h.engine.GetGeofence(id)
	if err != nil {
		if err == geofence.ErrGeofenceNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "geofence not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var req CreateGeofenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Update fields
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Coordinates != nil {
		existing.Coordinates = req.Coordinates
	}
	if req.Center != nil {
		existing.Center = req.Center
	}
	if req.Radius > 0 {
		existing.Radius = req.Radius
	}
	existing.MinAltitude = req.MinAltitude
	existing.MaxAltitude = req.MaxAltitude
	existing.AlertOnEnter = req.AlertOnEnter
	existing.AlertOnExit = req.AlertOnExit
	existing.Enabled = req.Enabled

	if err := h.engine.UpdateGeofence(existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// DeleteGeofence removes a geofence
func (h *GeofencesHandler) DeleteGeofence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.engine.DeleteGeofence(id); err != nil {
		if err == geofence.ErrGeofenceNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "geofence not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBreaches returns breach history
func (h *GeofencesHandler) GetBreaches(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	geofenceID := r.URL.Query().Get("geofence_id")

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	breaches := h.engine.GetBreaches(deviceID, geofenceID, limit)

	resp := map[string]interface{}{
		"breaches": breaches,
		"count":    len(breaches),
	}

	writeJSON(w, http.StatusOK, resp)
}

// ClearBreaches removes all breach history
func (h *GeofencesHandler) ClearBreaches(w http.ResponseWriter, r *http.Request) {
	h.engine.ClearBreaches()
	writeJSON(w, http.StatusOK, map[string]string{"message": "breaches cleared"})
}

// GetStats returns geofence statistics
func (h *GeofencesHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.engine.GetStats()
	writeJSON(w, http.StatusOK, stats)
}
