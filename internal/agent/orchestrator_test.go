package agent

import (
	"context"
	"testing"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/engine"
)

func TestOrchestratorCreation(t *testing.T) {
	config := &engine.DispatcherConfig{
		MaxWorkers: 5,
		QueueSize:  100,
	}
	dispatcher := engine.NewDispatcher(config)

	if dispatcher == nil {
		t.Fatal("NewDispatcher returned nil")
	}
}

func TestOrchestrationRule(t *testing.T) {
	rule := OrchestrationRule{
		Name:      "test_rule",
		EventType: bus.EventTypeThreatDetected,
		Condition: func(e bus.Event) bool {
			return true
		},
		Action: func(ctx context.Context, e bus.Event) (string, error) {
			return "executed", nil
		},
		Priority: 1,
		Enabled:  true,
		Cooldown: time.Minute,
	}

	if rule.Name != "test_rule" {
		t.Errorf("Expected rule name 'test_rule', got '%s'", rule.Name)
	}

	if rule.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", rule.Priority)
	}

	if !rule.Enabled {
		t.Error("Rule should be enabled")
	}
}

func TestEventTypes(t *testing.T) {
	tests := []struct {
		eventType bus.EventType
		expected  string
	}{
		{bus.EventTypeThreatDetected, "threat.detected"},
		{bus.EventTypeIncidentCreated, "incident.created"},
		{bus.EventTypeReconComplete, "recon.complete"},
		{bus.EventTypeExploitationComplete, "exploitation.complete"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.eventType)
		}
	}
}
