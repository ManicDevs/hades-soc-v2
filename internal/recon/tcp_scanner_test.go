package recon

import (
	"testing"
	"time"
)

func TestTCPScannerCreation(t *testing.T) {
	scanner := NewTCPScanner()
	if scanner == nil {
		t.Fatal("NewTCPScanner returned nil")
	}
}

func TestTCPScannerConfig(t *testing.T) {
	scanner := &TCPScanner{
		target:  "192.168.1.1",
		ports:   []int{80, 443, 8080},
		timeout: 5 * time.Second,
		threads: 50,
	}

	if scanner.target != "192.168.1.1" {
		t.Errorf("Expected target '192.168.1.1', got '%s'", scanner.target)
	}
	if len(scanner.ports) != 3 {
		t.Errorf("Expected 3 ports, got %d", len(scanner.ports))
	}
	if scanner.timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", scanner.timeout)
	}
	if scanner.threads != 50 {
		t.Errorf("Expected threads 50, got %d", scanner.threads)
	}
}