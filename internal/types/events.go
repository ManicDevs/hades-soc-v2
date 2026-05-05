// Package types provides centralized type definitions for event schemas
// used across all modules in the Hades SOC platform.
//
// This package ensures all modules speak the same language when communicating
// via the EventBus, providing strongly-typed JSON schemas for all event types.
package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// SeverityLevel represents the severity of an event
type SeverityLevel string

const (
	SeverityInfo     SeverityLevel = "Info"
	SeverityWarning  SeverityLevel = "Warning"
	SeverityCritical SeverityLevel = "Critical"
)

// EventType represents the type of event being published
type EventType string

// Core Event Types for Autonomous Cascade
const (
	// Asset Discovery
	EventTypeNewAsset EventType = "asset.new"

	// Threat Detection
	EventTypeThreat EventType = "threat.detected"

	// Action Requests
	EventTypeActionRequest EventType = "action.request"

	// Logging & Reasoning
	EventTypeLog EventType = "log.event"
)

// BaseEvent is the foundation for all events in the system
type BaseEvent struct {
	ID           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	SourceModule string        `json:"source_module"`
	Severity     SeverityLevel `json:"severity"`
}

// NewBaseEvent creates a new base event with defaults
func NewBaseEvent(source string, severity SeverityLevel) BaseEvent {
	return BaseEvent{
		ID:           generateEventID(),
		Timestamp:    time.Now(),
		SourceModule: source,
		Severity:     severity,
	}
}

// NewAssetEvent is published for the Recon-to-Scan cascade
// Triggered when a new asset is discovered (IP, domain, etc.)
type NewAssetEvent struct {
	BaseEvent
	IP       string                 `json:"ip"`
	Domain   string                 `json:"domain,omitempty"`
	Provider string                 `json:"provider,omitempty"` // e.g., "AWS", "GCP", "Azure", "Unknown"
	Metadata map[string]interface{} `json:"metadata"`           // Additional context about the asset
}

// NewNewAssetEvent creates a new asset event
func NewNewAssetEvent(source, ip string) NewAssetEvent {
	return NewAssetEvent{
		BaseEvent: NewBaseEvent(source, SeverityInfo),
		IP:        ip,
		Metadata:  make(map[string]interface{}),
	}
}

// WithDomain sets the domain for the asset
func (e NewAssetEvent) WithDomain(domain string) NewAssetEvent {
	e.Domain = domain
	return e
}

// WithProvider sets the cloud provider
func (e NewAssetEvent) WithProvider(provider string) NewAssetEvent {
	e.Provider = provider
	return e
}

// WithMetadata adds metadata to the asset
func (e NewAssetEvent) WithMetadata(key string, value interface{}) NewAssetEvent {
	e.Metadata[key] = value
	return e
}

// ThreatEvent is published for the Quantum Shield trigger
// Triggered when a threat is detected that requires immediate response
type ThreatEvent struct {
	BaseEvent
	TargetID        string  `json:"target_id"`        // The entity under attack
	AttackType      string  `json:"attack_type"`      // e.g., "BruteForce", "DDoS", "Injection"
	Fingerprint     string  `json:"fingerprint"`      // Attacker identifier (IP, device fingerprint)
	ConfidenceScore float64 `json:"confidence_score"` // 0.0 to 1.0
}

// NewThreatEvent creates a new threat event
func NewThreatEvent(source, targetID, attackType string, confidence float64) ThreatEvent {
	severity := SeverityWarning
	if confidence > 0.8 {
		severity = SeverityCritical
	}
	return ThreatEvent{
		BaseEvent:       NewBaseEvent(source, severity),
		TargetID:        targetID,
		AttackType:      attackType,
		ConfidenceScore: confidence,
	}
}

// WithFingerprint adds attacker fingerprint
func (e ThreatEvent) WithFingerprint(fp string) ThreatEvent {
	e.Fingerprint = fp
	return e
}

// ActionRequest is used by the Orchestrator to command modules
// Requests specific actions like key rotation or IP isolation
type ActionRequest struct {
	BaseEvent
	ActionName string `json:"action_name"` // e.g., "RotateKeys", "IsolateIP", "ScanTarget"
	Target     string `json:"target"`      // The target of the action
	Reasoning  string `json:"reasoning"`   // Why this action is being requested
}

// NewActionRequest creates a new action request
func NewActionRequest(source, actionName, target, reasoning string) ActionRequest {
	return ActionRequest{
		BaseEvent:  NewBaseEvent(source, SeverityInfo),
		ActionName: actionName,
		Target:     target,
		Reasoning:  reasoning,
	}
}

// WithTarget sets the target for the action request (fluent builder)
func (a ActionRequest) WithTarget(target string) ActionRequest {
	a.Target = target
	return a
}

// LogEvent is published for the 'Thought Stream' dashboard
// Shows the agent's real-time internal reasoning
type LogEvent struct {
	BaseEvent
	AgentName         string `json:"agent_name"`         // Name of the agent (e.g., "Orchestrator", "ThreatEngine")
	Message           string `json:"message"`            // Human-readable message
	InternalReasoning string `json:"internal_reasoning"` // Detailed reasoning for the thought stream
}

// NewLogEvent creates a new log event
func NewLogEvent(agentName, message, reasoning string) LogEvent {
	return LogEvent{
		BaseEvent:         NewBaseEvent(agentName, SeverityInfo),
		AgentName:         agentName,
		Message:           message,
		InternalReasoning: reasoning,
	}
}

// VulnerabilityFoundEvent is published when a vulnerability is detected
// Triggers autonomous remediation via hot-swap patching
type VulnerabilityFoundEvent struct {
	BaseEvent
	Vulnerability struct {
		ID          string   `json:"id"`          // CVE ID or internal identifier
		Title       string   `json:"title"`       // Human-readable title
		Description string   `json:"description"` // Detailed description
		Severity    string   `json:"severity"`    // "critical", "high", "medium", "low"
		CVSSScore   float64  `json:"cvss_score"`  // CVSS score (0-10)
		ModuleName  string   `json:"module_name"` // Name of the vulnerable module
		FilePath    string   `json:"file_path"`   // Path to the vulnerable file
		Version     string   `json:"version"`     // Affected version
		References  []string `json:"references"`  // CVE references
	} `json:"vulnerability"`
	Affected struct {
		TargetID string `json:"target_id"` // Target that was scanned
		Host     string `json:"host"`      // Host/IP address
		Port     int    `json:"port"`      // Port if applicable
	} `json:"affected"`
	Remediation struct {
		AutoFixAvailable bool   `json:"auto_fix_available"` // Whether a hot-swap fix exists
		FixedModulePath  string `json:"fixed_module_path"`  // Path to the fixed module if available
		RequiresRestart  bool   `json:"requires_restart"`   // Whether restart is required
	} `json:"remediation"`
}

// NewVulnerabilityFoundEvent creates a new vulnerability found event
func NewVulnerabilityFoundEvent(source, target, moduleName, cveID string, cvssScore float64) VulnerabilityFoundEvent {
	event := VulnerabilityFoundEvent{}
	event.BaseEvent = NewBaseEvent(source, SeverityCritical)
	event.Vulnerability.ID = cveID
	event.Vulnerability.ModuleName = moduleName
	event.Vulnerability.CVSSScore = cvssScore
	event.Affected.TargetID = target
	return event
}

// WithSeverity sets the vulnerability severity
func (e VulnerabilityFoundEvent) WithSeverity(severity string) VulnerabilityFoundEvent {
	e.Vulnerability.Severity = severity
	return e
}

// WithDescription sets the vulnerability description
func (e VulnerabilityFoundEvent) WithDescription(desc string) VulnerabilityFoundEvent {
	e.Vulnerability.Description = desc
	return e
}

// WithFilePath sets the vulnerable file path
func (e VulnerabilityFoundEvent) WithFilePath(path string) VulnerabilityFoundEvent {
	e.Vulnerability.FilePath = path
	return e
}

// WithVersion sets the affected version
func (e VulnerabilityFoundEvent) WithVersion(version string) VulnerabilityFoundEvent {
	e.Vulnerability.Version = version
	return e
}

// WithReferences adds CVE references
func (e VulnerabilityFoundEvent) WithReferences(refs ...string) VulnerabilityFoundEvent {
	e.Vulnerability.References = append(e.Vulnerability.References, refs...)
	return e
}

// WithAutoFix marks that an automatic fix is available
func (e VulnerabilityFoundEvent) WithAutoFix(fixedPath string) VulnerabilityFoundEvent {
	e.Remediation.AutoFixAvailable = true
	e.Remediation.FixedModulePath = fixedPath
	e.Remediation.RequiresRestart = false
	return e
}

// LateralMovementEvent is published when lateral movement is detected
// Triggers network isolation and forensic scanning
type LateralMovementEvent struct {
	BaseEvent
	MovementType string   `json:"movement_type"` // "credential_hopping" or "protocol_anomaly"
	SourceNode   string   `json:"source_node"`   // Source of the lateral movement
	TargetNode   string   `json:"target_node"`   // Target node (if applicable)
	User         string   `json:"user"`          // User involved (for credential hopping)
	Protocol     string   `json:"protocol"`      // Protocol used (SSH, RDP, etc.)
	Port         int      `json:"port"`          // Port used
	TargetIPs    []string `json:"target_ips"`    // Multiple targets (for credential hopping)
	Confidence   float64  `json:"confidence"`    // Detection confidence (0-1)
	Description  string   `json:"description"`   // Human-readable description
	Isolation    struct {
		Required bool   `json:"isolation_required"` // Whether VLAN isolation is needed
		VLAN     string `json:"quarantine_vlan"`    // Target quarantine VLAN
		Reason   string `json:"isolation_reason"`   // Reason for isolation
	} `json:"isolation"`
	ForensicScan struct {
		Required bool   `json:"forensic_scan_required"` // Whether deep forensic scan needed
		ScanType string `json:"scan_type"`              // Type of forensic scan
		Priority int    `json:"priority"`               // Scan priority
	} `json:"forensic_scan"`
}

// NewLateralMovementEvent creates a new lateral movement event
func NewLateralMovementEvent(source, movementType, sourceNode, description string, confidence float64) LateralMovementEvent {
	event := LateralMovementEvent{}
	event.BaseEvent = NewBaseEvent(source, SeverityCritical)
	event.MovementType = movementType
	event.SourceNode = sourceNode
	event.Description = description
	event.Confidence = confidence
	// Default isolation settings
	event.Isolation.Required = true
	event.Isolation.VLAN = "quarantine"
	event.Isolation.Reason = fmt.Sprintf("Internal lateral movement detected via %s. Source node isolated to prevent further spread.", movementType)
	// Default forensic scan settings
	event.ForensicScan.Required = true
	event.ForensicScan.ScanType = "deep_forensic"
	event.ForensicScan.Priority = 1
	return event
}

// WithUser sets the user for credential hopping events
func (e LateralMovementEvent) WithUser(user string) LateralMovementEvent {
	e.User = user
	return e
}

// WithProtocol sets the protocol and port
func (e LateralMovementEvent) WithProtocol(protocol string, port int) LateralMovementEvent {
	e.Protocol = protocol
	e.Port = port
	return e
}

// SecurityUpgradeRequest is published to trigger quantum encryption upgrades
type SecurityUpgradeRequest struct {
	BaseEvent
	UpgradeType       string `json:"upgrade_type"`       // e.g., "pqc_key_rotation"
	Target            string `json:"target"`             // Target node/IP
	Reason            string `json:"reason"`             // Human-readable reason
	InternalReasoning string `json:"internal_reasoning"` // Detailed internal reasoning
}

// NewSecurityUpgradeRequest creates a new security upgrade request
func NewSecurityUpgradeRequest(source, target, reason string) SecurityUpgradeRequest {
	return SecurityUpgradeRequest{
		BaseEvent:   NewBaseEvent(source, SeverityCritical),
		UpgradeType: "pqc_key_rotation",
		Target:      target,
		Reason:      reason,
	}
}

// WithInternalReasoning adds detailed internal reasoning
func (s SecurityUpgradeRequest) WithInternalReasoning(reasoning string) SecurityUpgradeRequest {
	s.InternalReasoning = reasoning
	return s
}

// HoneyFileAccessedEvent is published when a honey-file decoy is accessed
// Indicates unauthorized lateral movement - immediate containment required
type HoneyFileAccessedEvent struct {
	BaseEvent
	FilePath   string  `json:"file_path"`   // Full path to the honey file
	FileName   string  `json:"file_name"`   // Just the filename
	AccessType string  `json:"access_type"` // "read", "write", "rename", "chmod"
	Accessor   string  `json:"accessor"`    // Process/user that accessed the file
	Confidence float64 `json:"confidence"`  // Always 1.0 (100% unauthorized)
	IsBurned   bool    `json:"is_burned"`   // True - prevents duplicate alerts
}

// NewHoneyFileAccessedEvent creates a new honey file access event
func NewHoneyFileAccessedEvent(source, filePath, fileName, accessType, accessor string) HoneyFileAccessedEvent {
	return HoneyFileAccessedEvent{
		BaseEvent:  NewBaseEvent(source, SeverityCritical),
		FilePath:   filePath,
		FileName:   fileName,
		AccessType: accessType,
		Accessor:   accessor,
		Confidence: 1.0, // 100% confidence - honey files are never legitimate access
		IsBurned:   true,
	}
}

// HoneyTokenTriggeredEvent is published when a honey-token identity is accessed
// Indicates 100% malicious intent - immediate isolation required
type HoneyTokenTriggeredEvent struct {
	BaseEvent
	Username        string  `json:"username"`      // The honey-token username accessed
	SourceIP        string  `json:"source_ip"`     // Attacker's IP address
	Fingerprint     string  `json:"fingerprint"`   // Device/browser fingerprint
	LoginTime       string  `json:"login_time"`    // Time of access attempt
	PasswordUsed    string  `json:"password_used"` // Password attempt (for forensics)
	Confidence      float64 `json:"confidence"`    // Always 1.0 (100% malicious)
	ImmediateAction struct {
		IsolateSource  bool `json:"isolate_source"`  // Immediate network isolation
		RevokeSessions bool `json:"revoke_sessions"` // Revoke all attacker sessions
		AlertSentinel  bool `json:"alert_sentinel"`  // Alert security team
	} `json:"immediate_action"`
}

// NewHoneyTokenTriggeredEvent creates a new honey-token trigger event
func NewHoneyTokenTriggeredEvent(source, username, sourceIP, fingerprint string) HoneyTokenTriggeredEvent {
	event := HoneyTokenTriggeredEvent{}
	event.BaseEvent = NewBaseEvent(source, SeverityCritical)
	event.Username = username
	event.SourceIP = sourceIP
	event.Fingerprint = fingerprint
	event.Confidence = 1.0 // 100% confidence - honey tokens are never legitimate
	event.ImmediateAction.IsolateSource = true
	event.ImmediateAction.RevokeSessions = true
	event.ImmediateAction.AlertSentinel = true
	return event
}

// WithPassword adds the password used in the attempt (for forensics)
func (h HoneyTokenTriggeredEvent) WithPassword(password string) HoneyTokenTriggeredEvent {
	h.PasswordUsed = password
	return h
}

// WithTarget sets the target node
func (e LateralMovementEvent) WithTarget(targetNode string) LateralMovementEvent {
	e.TargetNode = targetNode
	return e
}

// WithTargetIPs sets multiple target IPs (for credential hopping)
func (e LateralMovementEvent) WithTargetIPs(ips []string) LateralMovementEvent {
	e.TargetIPs = ips
	return e
}

// EventEnvelope wraps any event with type information for the EventBus
type EventEnvelope struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WrapEvent wraps an event into an envelope
func WrapEvent(eventType EventType, event interface{}) (EventEnvelope, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return EventEnvelope{}, err
	}
	return EventEnvelope{
		Type:    eventType,
		Payload: data,
	}, nil
}

// UnwrapEvent unwraps an envelope into a specific event type
func (e EventEnvelope) UnwrapEvent(target interface{}) error {
	return json.Unmarshal(e.Payload, target)
}

// ToJSON serializes the event to JSON bytes
func (e BaseEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes JSON bytes to the event
func (e *BaseEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// Validate checks if the event has required fields
func (e BaseEvent) Validate() error {
	if e.ID == "" {
		return ErrMissingEventID
	}
	if e.SourceModule == "" {
		return ErrMissingEventSource
	}
	if e.Timestamp.IsZero() {
		return ErrMissingEventTimestamp
	}
	return nil
}

// Event validation errors
var (
	ErrMissingEventID        = NewValidationError("event ID is required")
	ErrMissingEventType      = NewValidationError("event type is required")
	ErrMissingEventSource    = NewValidationError("event source is required")
	ErrMissingEventTimestamp = NewValidationError("event timestamp is required")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(msg string) error {
	return &ValidationError{Message: msg}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102_150405_") + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
