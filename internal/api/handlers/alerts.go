package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/open-uav/telemetry-bridge/internal/core/alerter"
)

// AlertsHandler handles alert-related API endpoints
type AlertsHandler struct {
	alerter *alerter.Alerter
}

// NewAlertsHandler creates a new alerts handler
func NewAlertsHandler(a *alerter.Alerter) *AlertsHandler {
	return &AlertsHandler{
		alerter: a,
	}
}

// GetAlerts returns all alerts with optional filtering
// GET /api/v1/alerts?device_id=xxx&acknowledged=false&limit=100
func (h *AlertsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	ackedStr := r.URL.Query().Get("acknowledged")
	limitStr := r.URL.Query().Get("limit")

	var acknowledged *bool
	if ackedStr != "" {
		acked := ackedStr == "true"
		acknowledged = &acked
	}

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	alerts := h.alerter.GetAlerts(deviceID, acknowledged, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
		"stats":  h.alerter.GetStats(),
	})
}

// GetAlert returns a single alert by ID
// GET /api/v1/alerts/{id}
func (h *AlertsHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")

	alert, err := h.alerter.GetAlert(alertID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// AcknowledgeAlert marks an alert as acknowledged
// POST /api/v1/alerts/{id}/ack
func (h *AlertsHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")

	// Get user from context if available
	ackedBy := "system"
	if user := r.Context().Value("user"); user != nil {
		if u, ok := user.(string); ok {
			ackedBy = u
		}
	}

	if err := h.alerter.AcknowledgeAlert(alertID, ackedBy); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Alert acknowledged",
	})
}

// ClearAlerts removes all alerts
// DELETE /api/v1/alerts
func (h *AlertsHandler) ClearAlerts(w http.ResponseWriter, r *http.Request) {
	h.alerter.ClearAlerts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Alerts cleared",
	})
}

// Rule endpoints

// GetRules returns all alert rules
// GET /api/v1/alerts/rules
func (h *AlertsHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	rules := h.alerter.GetRules()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRule returns a single rule by ID
// GET /api/v1/alerts/rules/{id}
func (h *AlertsHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")

	rule, err := h.alerter.GetRule(ruleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// CreateRule creates a new alert rule
// POST /api/v1/alerts/rules
func (h *AlertsHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule alerter.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if rule.Name == "" {
		http.Error(w, "Rule name is required", http.StatusBadRequest)
		return
	}

	if rule.Condition.Field == "" {
		http.Error(w, "Condition field is required", http.StatusBadRequest)
		return
	}

	if err := h.alerter.CreateRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// UpdateRule updates an existing alert rule
// PUT /api/v1/alerts/rules/{id}
func (h *AlertsHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")

	var rule alerter.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule.ID = ruleID

	if err := h.alerter.UpdateRule(&rule); err != nil {
		if err == alerter.ErrRuleNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// DeleteRule removes an alert rule
// DELETE /api/v1/alerts/rules/{id}
func (h *AlertsHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")

	if err := h.alerter.DeleteRule(ruleID); err != nil {
		if err == alerter.ErrRuleNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetStats returns alerter statistics
// GET /api/v1/alerts/stats
func (h *AlertsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.alerter.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
