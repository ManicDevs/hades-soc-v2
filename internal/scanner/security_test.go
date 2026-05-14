package scanner

import (
	"testing"
	"time"
)

func TestSecurityScannerCreation(t *testing.T) {
	scanner := NewSecurityScanner()
	if scanner == nil {
		t.Fatal("NewSecurityScanner returned nil")
	}
}

func TestScanTarget(t *testing.T) {
	target := ScanTarget{
		ID:      "target-001",
		Type:    "host",
		Address: "192.168.1.100",
		Port:    443,
	}

	if target.ID != "target-001" {
		t.Errorf("Expected ID 'target-001', got '%s'", target.ID)
	}
	if target.Type != "host" {
		t.Errorf("Expected Type 'host', got '%s'", target.Type)
	}
	if target.Address != "192.168.1.100" {
		t.Errorf("Expected Address '192.168.1.100', got '%s'", target.Address)
	}
}

func TestScanResult(t *testing.T) {
	result := ScanResult{
		TargetID:    "target-001",
		ScannerType: "vulnerability",
		Status:      "completed",
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}

	if result.TargetID != "target-001" {
		t.Errorf("Expected TargetID 'target-001', got '%s'", result.TargetID)
	}
	if result.Status != "completed" {
		t.Errorf("Expected Status 'completed', got '%s'", result.Status)
	}
}

func TestScanPolicy(t *testing.T) {
	policy := ScanPolicy{
		ID:      "policy-001",
		Name:    "Standard Scan",
		Enabled: true,
	}

	if policy.ID != "policy-001" {
		t.Errorf("Expected ID 'policy-001', got '%s'", policy.ID)
	}
	if !policy.Enabled {
		t.Error("Policy should be enabled")
	}
}
