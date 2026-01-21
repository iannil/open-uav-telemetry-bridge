// Package alerter provides alert generation and rule evaluation for drone telemetry
package alerter

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-uav/telemetry-bridge/internal/models"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeBatteryLow      AlertType = "battery_low"
	AlertTypeConnectionLost  AlertType = "connection_lost"
	AlertTypeSignalWeak      AlertType = "signal_weak"
	AlertTypeGeofenceBreach  AlertType = "geofence_breach"
	AlertTypeCustom          AlertType = "custom"
)

// AlertSeverity represents alert severity level
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// Alert represents a triggered alert
type Alert struct {
	ID          string        `json:"id"`
	RuleID      string        `json:"rule_id"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	DeviceID    string        `json:"device_id"`
	Message     string        `json:"message"`
	Value       float64       `json:"value,omitempty"`
	Threshold   float64       `json:"threshold,omitempty"`
	Timestamp   int64         `json:"timestamp"`
	Acknowledged bool         `json:"acknowledged"`
	AckedAt     int64         `json:"acked_at,omitempty"`
	AckedBy     string        `json:"acked_by,omitempty"`
}

// Rule represents an alert rule
type Rule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Enabled     bool          `json:"enabled"`
	Condition   Condition     `json:"condition"`
	CooldownMs  int64         `json:"cooldown_ms"` // Minimum time between alerts
	CreatedAt   int64         `json:"created_at"`
	UpdatedAt   int64         `json:"updated_at"`
}

// Condition defines when an alert should trigger
type Condition struct {
	Field     string  `json:"field"`     // e.g., "battery_percent", "satellites_visible"
	Operator  string  `json:"operator"`  // "<", ">", "<=", ">=", "==", "!="
	Threshold float64 `json:"threshold"` // Value to compare against
}

// Alerter is the alert engine that evaluates rules and generates alerts
type Alerter struct {
	rules           map[string]*Rule
	alerts          []Alert
	alertsByDevice  map[string][]string // device_id -> alert_ids
	lastAlertTime   map[string]int64    // rule_id:device_id -> last alert timestamp
	maxAlerts       int
	onAlert         func(*Alert)
	mu              sync.RWMutex
}

// Config holds alerter configuration
type Config struct {
	MaxAlerts int // Maximum number of alerts to keep in memory
}

// New creates a new alerter
func New(cfg Config) *Alerter {
	maxAlerts := cfg.MaxAlerts
	if maxAlerts <= 0 {
		maxAlerts = 1000
	}

	a := &Alerter{
		rules:          make(map[string]*Rule),
		alerts:         make([]Alert, 0),
		alertsByDevice: make(map[string][]string),
		lastAlertTime:  make(map[string]int64),
		maxAlerts:      maxAlerts,
	}

	// Add default rules
	a.addDefaultRules()

	return a
}

// addDefaultRules adds built-in alert rules
func (a *Alerter) addDefaultRules() {
	now := time.Now().UnixMilli()

	defaultRules := []*Rule{
		{
			ID:       "default-battery-low",
			Name:     "Low Battery",
			Type:     AlertTypeBatteryLow,
			Severity: SeverityWarning,
			Enabled:  true,
			Condition: Condition{
				Field:     "battery_percent",
				Operator:  "<",
				Threshold: 20,
			},
			CooldownMs: 60000, // 1 minute
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:       "default-battery-critical",
			Name:     "Critical Battery",
			Type:     AlertTypeBatteryLow,
			Severity: SeverityCritical,
			Enabled:  true,
			Condition: Condition{
				Field:     "battery_percent",
				Operator:  "<",
				Threshold: 10,
			},
			CooldownMs: 30000, // 30 seconds
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:       "default-weak-signal",
			Name:     "Weak Signal",
			Type:     AlertTypeSignalWeak,
			Severity: SeverityWarning,
			Enabled:  true,
			Condition: Condition{
				Field:     "signal_quality",
				Operator:  "<",
				Threshold: 30,
			},
			CooldownMs: 30000,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	for _, rule := range defaultRules {
		a.rules[rule.ID] = rule
	}
}

// SetAlertCallback sets a callback function to be called when an alert is generated
func (a *Alerter) SetAlertCallback(cb func(*Alert)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onAlert = cb
}

// Evaluate checks the drone state against all rules and generates alerts
func (a *Alerter) Evaluate(state *models.DroneState) []*Alert {
	a.mu.Lock()
	defer a.mu.Unlock()

	var generated []*Alert
	now := time.Now().UnixMilli()

	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		value, ok := a.getFieldValue(state, rule.Condition.Field)
		if !ok {
			continue
		}

		if !a.evaluateCondition(value, rule.Condition) {
			continue
		}

		// Check cooldown
		key := rule.ID + ":" + state.DeviceID
		if lastTime, exists := a.lastAlertTime[key]; exists {
			if now-lastTime < rule.CooldownMs {
				continue
			}
		}

		// Generate alert
		alert := &Alert{
			ID:        uuid.New().String(),
			RuleID:    rule.ID,
			Type:      rule.Type,
			Severity:  rule.Severity,
			DeviceID:  state.DeviceID,
			Message:   a.generateMessage(rule, state, value),
			Value:     value,
			Threshold: rule.Condition.Threshold,
			Timestamp: now,
		}

		a.addAlert(alert)
		a.lastAlertTime[key] = now
		generated = append(generated, alert)

		// Call callback if set
		if a.onAlert != nil {
			go a.onAlert(alert)
		}
	}

	return generated
}

// getFieldValue extracts a field value from the drone state
func (a *Alerter) getFieldValue(state *models.DroneState, field string) (float64, bool) {
	switch field {
	case "battery_percent":
		return float64(state.Status.BatteryPercent), true
	case "signal_quality":
		return float64(state.Status.SignalQuality), true
	case "altitude":
		return state.Location.AltGNSS, true
	case "altitude_baro":
		return state.Location.AltBaro, true
	case "speed":
		// Calculate ground speed
		vx := state.Velocity.Vx
		vy := state.Velocity.Vy
		return vx*vx + vy*vy, true // squared for comparison
	default:
		return 0, false
	}
}

// evaluateCondition checks if a value satisfies the condition
func (a *Alerter) evaluateCondition(value float64, cond Condition) bool {
	switch cond.Operator {
	case "<":
		return value < cond.Threshold
	case ">":
		return value > cond.Threshold
	case "<=":
		return value <= cond.Threshold
	case ">=":
		return value >= cond.Threshold
	case "==":
		return value == cond.Threshold
	case "!=":
		return value != cond.Threshold
	default:
		return false
	}
}

// generateMessage creates a human-readable alert message
func (a *Alerter) generateMessage(rule *Rule, state *models.DroneState, value float64) string {
	switch rule.Type {
	case AlertTypeBatteryLow:
		return formatMessage("Battery at %.0f%% (threshold: %.0f%%)", value, rule.Condition.Threshold)
	case AlertTypeSignalWeak:
		return formatMessage("Signal quality: %.0f%% (minimum: %.0f%%)", value, rule.Condition.Threshold)
	default:
		return formatMessage("%s: %.2f %s %.2f", rule.Condition.Field, value, rule.Condition.Operator, rule.Condition.Threshold)
	}
}

func formatMessage(format string, args ...interface{}) string {
	return formatString(format, args...)
}

func formatString(format string, args ...interface{}) string {
	// Simple sprintf implementation
	result := format
	for _, arg := range args {
		// Replace first %X with arg value
		for i := 0; i < len(result)-1; i++ {
			if result[i] == '%' {
				end := i + 2
				if end <= len(result) {
					var formatted string
					switch v := arg.(type) {
					case float64:
						if result[i+1] == '0' || result[i+1] == '.' {
							// Find end of format spec
							for end < len(result) && result[end] != ' ' && result[end] != ')' && result[end] != '(' {
								end++
							}
						}
						formatted = floatToString(v, result[i:end])
					case string:
						formatted = v
					default:
						formatted = "?"
					}
					result = result[:i] + formatted + result[end:]
					break
				}
			}
		}
	}
	return result
}

func floatToString(v float64, format string) string {
	// Simple float formatting
	if len(format) > 2 && format[1] == '.' && format[2] == '0' {
		return intToString(int(v))
	}
	if len(format) > 3 && format[1] == '.' && format[2] == '2' {
		return intToString(int(v)) + "." + intToString(int((v-float64(int(v)))*100))
	}
	return intToString(int(v))
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

// addAlert adds an alert to the list, maintaining max size
func (a *Alerter) addAlert(alert *Alert) {
	a.alerts = append(a.alerts, *alert)

	// Track by device
	a.alertsByDevice[alert.DeviceID] = append(a.alertsByDevice[alert.DeviceID], alert.ID)

	// Trim if over max
	if len(a.alerts) > a.maxAlerts {
		// Remove oldest alerts
		removed := a.alerts[0]
		a.alerts = a.alerts[1:]

		// Clean up device tracking
		if ids, ok := a.alertsByDevice[removed.DeviceID]; ok {
			for i, id := range ids {
				if id == removed.ID {
					a.alertsByDevice[removed.DeviceID] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}
}

// GetAlerts returns all alerts, optionally filtered
func (a *Alerter) GetAlerts(deviceID string, acknowledged *bool, limit int) []Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []Alert

	for i := len(a.alerts) - 1; i >= 0; i-- {
		alert := a.alerts[i]

		if deviceID != "" && alert.DeviceID != deviceID {
			continue
		}

		if acknowledged != nil && alert.Acknowledged != *acknowledged {
			continue
		}

		result = append(result, alert)

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

// AcknowledgeAlert marks an alert as acknowledged
func (a *Alerter) AcknowledgeAlert(alertID, ackedBy string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i := range a.alerts {
		if a.alerts[i].ID == alertID {
			a.alerts[i].Acknowledged = true
			a.alerts[i].AckedAt = time.Now().UnixMilli()
			a.alerts[i].AckedBy = ackedBy
			return nil
		}
	}

	return ErrAlertNotFound
}

// GetAlert returns a single alert by ID
func (a *Alerter) GetAlert(alertID string) (*Alert, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, alert := range a.alerts {
		if alert.ID == alertID {
			return &alert, nil
		}
	}

	return nil, ErrAlertNotFound
}

// Rule management

// GetRules returns all rules
func (a *Alerter) GetRules() []*Rule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rules := make([]*Rule, 0, len(a.rules))
	for _, rule := range a.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetRule returns a single rule by ID
func (a *Alerter) GetRule(ruleID string) (*Rule, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if rule, ok := a.rules[ruleID]; ok {
		return rule, nil
	}
	return nil, ErrRuleNotFound
}

// CreateRule creates a new rule
func (a *Alerter) CreateRule(rule *Rule) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	now := time.Now().UnixMilli()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	a.rules[rule.ID] = rule
	return nil
}

// UpdateRule updates an existing rule
func (a *Alerter) UpdateRule(rule *Rule) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.rules[rule.ID]; !ok {
		return ErrRuleNotFound
	}

	rule.UpdatedAt = time.Now().UnixMilli()
	a.rules[rule.ID] = rule
	return nil
}

// DeleteRule removes a rule
func (a *Alerter) DeleteRule(ruleID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.rules[ruleID]; !ok {
		return ErrRuleNotFound
	}

	delete(a.rules, ruleID)
	return nil
}

// ClearAlerts removes all alerts
func (a *Alerter) ClearAlerts() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.alerts = make([]Alert, 0)
	a.alertsByDevice = make(map[string][]string)
}

// GetStats returns alerter statistics
func (a *Alerter) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	unacked := 0
	for _, alert := range a.alerts {
		if !alert.Acknowledged {
			unacked++
		}
	}

	return map[string]interface{}{
		"total_alerts":        len(a.alerts),
		"unacknowledged":      unacked,
		"rules_count":         len(a.rules),
		"devices_with_alerts": len(a.alertsByDevice),
	}
}

// MarshalJSON for Alert
func (a *Alert) MarshalJSON() ([]byte, error) {
	type alias Alert
	return json.Marshal((*alias)(a))
}

// Errors
var (
	ErrAlertNotFound = &AlertError{"alert not found"}
	ErrRuleNotFound  = &AlertError{"rule not found"}
)

type AlertError struct {
	msg string
}

func (e *AlertError) Error() string {
	return e.msg
}
