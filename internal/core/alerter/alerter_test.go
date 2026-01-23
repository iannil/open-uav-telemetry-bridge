package alerter

import (
	"sync"
	"testing"
	"time"

	"github.com/open-uav/telemetry-bridge/internal/models"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		maxAlerts int
		wantMax   int
	}{
		{"positive max", 100, 100},
		{"zero defaults to 1000", 0, 1000},
		{"negative defaults to 1000", -5, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(Config{MaxAlerts: tt.maxAlerts})
			if a.maxAlerts != tt.wantMax {
				t.Errorf("maxAlerts = %d, want %d", a.maxAlerts, tt.wantMax)
			}
		})
	}
}

func TestAlerter_DefaultRules(t *testing.T) {
	a := New(Config{})

	rules := a.GetRules()
	if len(rules) != 3 {
		t.Errorf("Should have 3 default rules, got %d", len(rules))
	}

	// Check default rule IDs
	expectedIDs := map[string]bool{
		"default-battery-low":      false,
		"default-battery-critical": false,
		"default-weak-signal":      false,
	}

	for _, rule := range rules {
		if _, ok := expectedIDs[rule.ID]; ok {
			expectedIDs[rule.ID] = true
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Default rule '%s' not found", id)
		}
	}
}

func TestAlerter_Evaluate_BatteryLow(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{
		DeviceID: "drone-1",
		Status: models.Status{
			BatteryPercent: 15, // Below 20% threshold but above 10%
			SignalQuality:  95, // Above threshold to avoid weak signal alert
		},
	}

	alerts := a.Evaluate(state)

	if len(alerts) != 1 {
		t.Fatalf("Should generate 1 alert for low battery, got %d", len(alerts))
	}

	if alerts[0].Type != AlertTypeBatteryLow {
		t.Errorf("Alert type should be battery_low, got %s", alerts[0].Type)
	}
	if alerts[0].Severity != SeverityWarning {
		t.Errorf("Alert severity should be warning, got %s", alerts[0].Severity)
	}
	if alerts[0].DeviceID != "drone-1" {
		t.Errorf("Alert deviceID should be 'drone-1', got '%s'", alerts[0].DeviceID)
	}
}

func TestAlerter_Evaluate_BatteryCritical(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{
		DeviceID: "drone-1",
		Status: models.Status{
			BatteryPercent: 5,  // Below 10% critical threshold (also below 20%)
			SignalQuality:  95, // Above threshold to avoid weak signal alert
		},
	}

	alerts := a.Evaluate(state)

	// Should trigger both warning and critical
	if len(alerts) != 2 {
		t.Fatalf("Should generate 2 alerts for critical battery, got %d", len(alerts))
	}

	hasWarning := false
	hasCritical := false
	for _, alert := range alerts {
		if alert.Severity == SeverityWarning {
			hasWarning = true
		}
		if alert.Severity == SeverityCritical {
			hasCritical = true
		}
	}

	if !hasWarning || !hasCritical {
		t.Error("Should have both warning and critical alerts")
	}
}

func TestAlerter_Evaluate_Cooldown(t *testing.T) {
	a := New(Config{})

	// Modify cooldown for testing
	for _, rule := range a.rules {
		rule.CooldownMs = 100 // 100ms cooldown
	}

	state := &models.DroneState{
		DeviceID: "drone-1",
		Status: models.Status{
			BatteryPercent: 15,
		},
	}

	// First evaluation
	alerts1 := a.Evaluate(state)
	if len(alerts1) == 0 {
		t.Fatal("First evaluation should generate alerts")
	}

	// Immediate second evaluation should be blocked by cooldown
	alerts2 := a.Evaluate(state)
	if len(alerts2) != 0 {
		t.Error("Second evaluation within cooldown should not generate alerts")
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// Third evaluation should work
	alerts3 := a.Evaluate(state)
	if len(alerts3) == 0 {
		t.Error("Third evaluation after cooldown should generate alerts")
	}
}

func TestAlerter_Evaluate_NoAlert(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{
		DeviceID: "drone-1",
		Status: models.Status{
			BatteryPercent: 85,  // Above threshold
			SignalQuality:  95,  // Above threshold
		},
	}

	alerts := a.Evaluate(state)

	if len(alerts) != 0 {
		t.Errorf("Should not generate alerts for healthy state, got %d", len(alerts))
	}
}

func TestAlerter_GetAlerts(t *testing.T) {
	a := New(Config{})

	// Disable cooldown for testing
	for _, rule := range a.rules {
		rule.CooldownMs = 0
	}

	state1 := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	state2 := &models.DroneState{DeviceID: "drone-2", Status: models.Status{BatteryPercent: 15}}

	a.Evaluate(state1)
	a.Evaluate(state2)

	// Get all alerts
	allAlerts := a.GetAlerts("", nil, 0)
	if len(allAlerts) < 2 {
		t.Errorf("Should have at least 2 alerts, got %d", len(allAlerts))
	}

	// Filter by device
	drone1Alerts := a.GetAlerts("drone-1", nil, 0)
	for _, alert := range drone1Alerts {
		if alert.DeviceID != "drone-1" {
			t.Error("Filtered alerts should only contain drone-1")
		}
	}

	// Filter by acknowledged
	acked := false
	unackedAlerts := a.GetAlerts("", &acked, 0)
	for _, alert := range unackedAlerts {
		if alert.Acknowledged {
			t.Error("Filter should only return unacknowledged alerts")
		}
	}

	// Limit
	limitedAlerts := a.GetAlerts("", nil, 1)
	if len(limitedAlerts) != 1 {
		t.Errorf("Limit should return 1 alert, got %d", len(limitedAlerts))
	}
}

func TestAlerter_AcknowledgeAlert(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	alerts := a.Evaluate(state)

	if len(alerts) == 0 {
		t.Fatal("Need at least one alert to test")
	}

	alertID := alerts[0].ID

	err := a.AcknowledgeAlert(alertID, "admin")
	if err != nil {
		t.Errorf("AcknowledgeAlert should not error: %v", err)
	}

	// Verify acknowledgment
	alert, err := a.GetAlert(alertID)
	if err != nil {
		t.Fatalf("GetAlert should not error: %v", err)
	}
	if !alert.Acknowledged {
		t.Error("Alert should be acknowledged")
	}
	if alert.AckedBy != "admin" {
		t.Errorf("AckedBy should be 'admin', got '%s'", alert.AckedBy)
	}
	if alert.AckedAt == 0 {
		t.Error("AckedAt should be set")
	}
}

func TestAlerter_AcknowledgeAlert_NotFound(t *testing.T) {
	a := New(Config{})

	err := a.AcknowledgeAlert("nonexistent", "admin")
	if err != ErrAlertNotFound {
		t.Errorf("Should return ErrAlertNotFound, got %v", err)
	}
}

func TestAlerter_GetAlert(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	alerts := a.Evaluate(state)

	if len(alerts) == 0 {
		t.Fatal("Need at least one alert")
	}

	alert, err := a.GetAlert(alerts[0].ID)
	if err != nil {
		t.Errorf("GetAlert should not error: %v", err)
	}
	if alert.ID != alerts[0].ID {
		t.Error("GetAlert should return correct alert")
	}

	// Test not found
	_, err = a.GetAlert("nonexistent")
	if err != ErrAlertNotFound {
		t.Errorf("Should return ErrAlertNotFound, got %v", err)
	}
}

func TestAlerter_RuleManagement(t *testing.T) {
	a := New(Config{})

	// Create rule
	rule := &Rule{
		Name:     "Custom Rule",
		Type:     AlertTypeCustom,
		Severity: SeverityWarning,
		Enabled:  true,
		Condition: Condition{
			Field:     "altitude",
			Operator:  ">",
			Threshold: 500,
		},
	}

	err := a.CreateRule(rule)
	if err != nil {
		t.Errorf("CreateRule should not error: %v", err)
	}
	if rule.ID == "" {
		t.Error("Rule should have generated ID")
	}
	if rule.CreatedAt == 0 {
		t.Error("Rule should have CreatedAt timestamp")
	}

	// Get rule
	retrieved, err := a.GetRule(rule.ID)
	if err != nil {
		t.Errorf("GetRule should not error: %v", err)
	}
	if retrieved.Name != "Custom Rule" {
		t.Error("Retrieved rule should match")
	}

	// Update rule
	rule.Name = "Updated Rule"
	err = a.UpdateRule(rule)
	if err != nil {
		t.Errorf("UpdateRule should not error: %v", err)
	}

	retrieved, _ = a.GetRule(rule.ID)
	if retrieved.Name != "Updated Rule" {
		t.Error("Rule should be updated")
	}

	// Delete rule
	err = a.DeleteRule(rule.ID)
	if err != nil {
		t.Errorf("DeleteRule should not error: %v", err)
	}

	_, err = a.GetRule(rule.ID)
	if err != ErrRuleNotFound {
		t.Error("Deleted rule should not be found")
	}
}

func TestAlerter_UpdateRule_NotFound(t *testing.T) {
	a := New(Config{})

	rule := &Rule{ID: "nonexistent"}
	err := a.UpdateRule(rule)
	if err != ErrRuleNotFound {
		t.Errorf("Should return ErrRuleNotFound, got %v", err)
	}
}

func TestAlerter_DeleteRule_NotFound(t *testing.T) {
	a := New(Config{})

	err := a.DeleteRule("nonexistent")
	if err != ErrRuleNotFound {
		t.Errorf("Should return ErrRuleNotFound, got %v", err)
	}
}

func TestAlerter_ClearAlerts(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	a.Evaluate(state)

	alerts := a.GetAlerts("", nil, 0)
	if len(alerts) == 0 {
		t.Fatal("Should have alerts before clear")
	}

	a.ClearAlerts()

	alerts = a.GetAlerts("", nil, 0)
	if len(alerts) != 0 {
		t.Errorf("Should have no alerts after clear, got %d", len(alerts))
	}
}

func TestAlerter_GetStats(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	a.Evaluate(state)

	stats := a.GetStats()

	if stats["total_alerts"].(int) == 0 {
		t.Error("total_alerts should be > 0")
	}
	if stats["rules_count"].(int) == 0 {
		t.Error("rules_count should be > 0")
	}
	if _, ok := stats["unacknowledged"]; !ok {
		t.Error("stats should contain unacknowledged")
	}
}

func TestAlerter_MaxAlerts(t *testing.T) {
	a := New(Config{MaxAlerts: 3})

	// Disable cooldown
	for _, rule := range a.rules {
		rule.CooldownMs = 0
	}

	// Generate more alerts than max
	for i := 0; i < 5; i++ {
		state := &models.DroneState{
			DeviceID: "drone-1",
			Status:   models.Status{BatteryPercent: 15},
		}
		a.Evaluate(state)
	}

	alerts := a.GetAlerts("", nil, 0)
	if len(alerts) > 3 {
		t.Errorf("Should not exceed max alerts, got %d", len(alerts))
	}
}

func TestAlerter_SetAlertCallback(t *testing.T) {
	a := New(Config{})

	var called bool
	var mu sync.Mutex
	var receivedAlert *Alert

	a.SetAlertCallback(func(alert *Alert) {
		mu.Lock()
		defer mu.Unlock()
		called = true
		receivedAlert = alert
	})

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 15}}
	a.Evaluate(state)

	// Wait for async callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Error("Callback should be called")
	}
	if receivedAlert == nil {
		t.Error("Callback should receive alert")
	}
}

func TestAlerter_EvaluateCondition(t *testing.T) {
	a := New(Config{})

	tests := []struct {
		value    float64
		operator string
		thresh   float64
		want     bool
	}{
		{10, "<", 20, true},
		{10, "<", 5, false},
		{20, ">", 10, true},
		{5, ">", 10, false},
		{10, "<=", 10, true},
		{10, "<=", 5, false},
		{10, ">=", 10, true},
		{10, ">=", 15, false},
		{10, "==", 10, true},
		{10, "==", 5, false},
		{10, "!=", 5, true},
		{10, "!=", 10, false},
		{10, "invalid", 10, false},
	}

	for _, tt := range tests {
		cond := Condition{Operator: tt.operator, Threshold: tt.thresh}
		got := a.evaluateCondition(tt.value, cond)
		if got != tt.want {
			t.Errorf("evaluateCondition(%v %s %v) = %v, want %v",
				tt.value, tt.operator, tt.thresh, got, tt.want)
		}
	}
}

func TestAlerter_GetFieldValue(t *testing.T) {
	a := New(Config{})

	state := &models.DroneState{
		Status: models.Status{
			BatteryPercent: 85,
			SignalQuality:  90,
		},
		Location: models.Location{
			AltGNSS: 100.5,
			AltBaro: 99.0,
		},
		Velocity: models.Velocity{
			Vx: 3.0,
			Vy: 4.0,
		},
	}

	tests := []struct {
		field   string
		wantVal float64
		wantOk  bool
	}{
		{"battery_percent", 85, true},
		{"signal_quality", 90, true},
		{"altitude", 100.5, true},
		{"altitude_baro", 99.0, true},
		{"speed", 25, true}, // 3^2 + 4^2 = 25
		{"unknown", 0, false},
	}

	for _, tt := range tests {
		val, ok := a.getFieldValue(state, tt.field)
		if ok != tt.wantOk {
			t.Errorf("getFieldValue(%s) ok = %v, want %v", tt.field, ok, tt.wantOk)
		}
		if ok && val != tt.wantVal {
			t.Errorf("getFieldValue(%s) = %v, want %v", tt.field, val, tt.wantVal)
		}
	}
}

func TestAlerter_DisabledRule(t *testing.T) {
	a := New(Config{})

	// Disable all rules
	for _, rule := range a.rules {
		rule.Enabled = false
	}

	state := &models.DroneState{DeviceID: "drone-1", Status: models.Status{BatteryPercent: 5}}
	alerts := a.Evaluate(state)

	if len(alerts) != 0 {
		t.Error("Disabled rules should not generate alerts")
	}
}
