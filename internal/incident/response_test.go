package incident

import (
	"testing"
	"time"
)

func TestIncidentResponseManagerCreation(t *testing.T) {
	manager := NewIncidentResponseManager(nil)
	if manager == nil {
		t.Fatal("NewIncidentResponseManager returned nil")
	}
}

func TestWorkflow(t *testing.T) {
	workflow := Workflow{
		ID:          "test_workflow",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Enabled:     true,
		Priority:    1,
	}

	if workflow.ID != "test_workflow" {
		t.Errorf("Expected ID 'test_workflow', got '%s'", workflow.ID)
	}
	if !workflow.Enabled {
		t.Error("Workflow should be enabled")
	}
}

func TestIncident(t *testing.T) {
	incident := Incident{
		ID:          "INC-001",
		Title:       "Test Incident",
		Description: "A test incident",
		Severity:    "high",
		Status:      "new",
		Priority:    1,
		Source:      "manual",
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	if incident.ID != "INC-001" {
		t.Errorf("Expected ID 'INC-001', got '%s'", incident.ID)
	}
	if incident.Severity != "high" {
		t.Errorf("Expected Severity 'high', got '%s'", incident.Severity)
	}
	if incident.Status != "new" {
		t.Errorf("Expected Status 'new', got '%s'", incident.Status)
	}
}

func TestIncidentAction(t *testing.T) {
	action := IncidentAction{
		ID:          "ACT-001",
		Type:        "isolation",
		Description: "Isolate compromised host",
		Status:      "pending",
		ExecutedAt:  time.Now(),
		ExecutedBy:  "system",
	}

	if action.ID != "ACT-001" {
		t.Errorf("Expected ID 'ACT-001', got '%s'", action.ID)
	}
	if action.Type != "isolation" {
		t.Errorf("Expected Type 'isolation', got '%s'", action.Type)
	}
	if action.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", action.Status)
	}
}

func TestWorkflowStep(t *testing.T) {
	step := WorkflowStep{
		ID:         "step_001",
		Name:       "Detect Threat",
		Type:       "action",
		Action:     "detect",
		Timeout:    30 * time.Second,
		RetryCount: 3,
		Parallel:   false,
	}

	if step.ID != "step_001" {
		t.Errorf("Expected ID 'step_001', got '%s'", step.ID)
	}
	if step.Type != "action" {
		t.Errorf("Expected Type 'action', got '%s'", step.Type)
	}
	if step.RetryCount != 3 {
		t.Errorf("Expected RetryCount 3, got %d", step.RetryCount)
	}
}
