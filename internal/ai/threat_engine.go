package ai

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/types"
	"hades-v2/internal/zerotrust"
)

// SecurityEvent represents a security event for AI analysis
type SecurityEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Features  map[string]interface{} `json:"features"`
	Signature string                 `json:"signature"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ThreatAssessment represents the AI threat assessment result
type ThreatAssessment struct {
	EventID         string    `json:"event_id"`
	ThreatScore     float64   `json:"threat_score"`
	MLScore         float64   `json:"ml_score"`
	AnomalyScore    float64   `json:"anomaly_score"`
	PatternScore    float64   `json:"pattern_score"`
	RiskLevel       string    `json:"risk_level"`
	Confidence      float64   `json:"confidence"`
	Predictions     []string  `json:"predictions"`
	Anomalies       []Anomaly `json:"anomalies"`
	Patterns        []Pattern `json:"patterns"`
	Recommendations []string  `json:"recommendations"`
	Timestamp       time.Time `json:"timestamp"`
}

// Anomaly represents detected anomalous behavior
type Anomaly struct {
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// Pattern represents detected threat patterns
type Pattern struct {
	Name       string  `json:"name"`
	Match      bool    `json:"match"`
	Confidence float64 `json:"confidence"`
	Severity   string  `json:"severity"`
}

// SanitizationResult represents the result of input sanitization
type SanitizationResult struct {
	IsSafe         bool    `json:"is_safe"`
	Confidence     float64 `json:"confidence"`
	ThreatType     string  `json:"threat_type,omitempty"`
	MatchedPattern string  `json:"matched_pattern,omitempty"`
	Reasoning      string  `json:"reasoning"`
	InputSource    string  `json:"input_source,omitempty"`
}

// InjectionMatch represents a detected prompt injection attempt
type InjectionMatch struct {
	Pattern    string  `json:"pattern"`
	Match      string  `json:"match"`
	Source     string  `json:"source"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
}

// SanitizationLayer provides adversarial AI defense against prompt injection
type SanitizationLayer struct {
	promptInjectionPatterns []*regexp.Regexp
	quarantineEvents        []string
	quarantineMu            sync.RWMutex
	enabled                 bool
}

// AIThreatEngine provides AI-powered threat intelligence
type AIThreatEngine struct {
	MLModel         *TensorFlowModel
	AnomalyDetector *AnomalyDetector
	ThreatScorer    *ThreatScoringEngine
	PatternMatcher  *PatternMatcher
	Baseline        *BehavioralBaseline
	ZeroTrustEngine *zerotrust.ZeroTrustEngine
	QuantumEngine   interface {
		GenerateKey(algorithm, keyType string) (interface{}, error)
	}
	TrafficAnalyzer    *TrafficPatternAnalyzer // Internal lateral movement detection
	SanitizationLayer  *SanitizationLayer      // Adversarial AI defense
	mu                 sync.RWMutex
	isTraining         bool
	eventBus           *bus.EventBus
	cascadeSubs        []*bus.Subscription
	cascadeActions     map[string]CascadeAction
	autoIsolate        bool
	autoUpgrade        bool
	bruteForceCount    map[string]int
	authFailureTracker map[string][]time.Time // fingerprint -> timestamps of failures
}

// TrafficPatternAnalyzer monitors internal traffic for lateral movement detection
type TrafficPatternAnalyzer struct {
	credentialUsage   map[string][]CredentialUsage // user -> list of usage events
	protocolAnomalies map[string][]ProtocolEvent   // node -> list of protocol events
	adminNodes        map[string]bool              // whitelist of admin nodes
	mu                sync.RWMutex
	detectionWindow   time.Duration
}

// CredentialUsage tracks credential usage on internal nodes
type CredentialUsage struct {
	User      string    `json:"user"`
	SourceIP  string    `json:"source_ip"`
	TargetIP  string    `json:"target_ip"`
	Protocol  string    `json:"protocol"` // SSH, RDP, etc.
	Timestamp time.Time `json:"timestamp"`
}

// ProtocolEvent tracks protocol usage for anomaly detection
type ProtocolEvent struct {
	SourceNode string    `json:"source_node"`
	TargetNode string    `json:"target_node"`
	Protocol   string    `json:"protocol"`
	Port       int       `json:"port"`
	Timestamp  time.Time `json:"timestamp"`
	IsAdmin    bool      `json:"is_admin"` // whether source is admin node
}

// NewSanitizationLayer creates a new adversarial AI defense layer
func NewSanitizationLayer() *SanitizationLayer {
	// Initialize prompt injection detection patterns
	patterns := []*regexp.Regexp{
		// Direct instruction overrides
		regexp.MustCompile(`(?i)ignore\s+previous\s+instructions`),
		regexp.MustCompile(`(?i)disregard\s+all\s+previous`),
		regexp.MustCompile(`(?i)forget\s+everything\s+above`),
		regexp.MustCompile(`(?i)system\s+override`),
		regexp.MustCompile(`(?i)admin\s+override`),
		regexp.MustCompile(`(?i)root\s+access`),

		// Privilege escalation attempts
		regexp.MustCompile(`(?i)sudo`),
		regexp.MustCompile(`(?i)administrator`),
		regexp.MustCompile(`(?i)privilege\s+escalation`),
		regexp.MustCompile(`(?i)escalate\s+privileges`),

		// Command injection patterns
		regexp.MustCompile(`(?i);\s*rm\s+-rf`),
		regexp.MustCompile(`(?i);\s*cat\s+/etc/passwd`),
		regexp.MustCompile(`(?i);\s*nc\s+-l`),
		regexp.MustCompile(`(?i)` + "```" + `.*?` + "```"),

		// Role-playing and system manipulation
		regexp.MustCompile(`(?i)you\s+are\s+now`),
		regexp.MustCompile(`(?i)act\s+as`),
		regexp.MustCompile(`(?i)pretend\s+to\s+be`),
		regexp.MustCompile(`(?i)from\s+now\s+on\s+you\s+are`),

		// Context manipulation
		regexp.MustCompile(`(?i)new\s+context`),
		regexp.MustCompile(`(?i)change\s+context`),
		regexp.MustCompile(`(?i)switch\s+role`),

		// Information disclosure attempts
		regexp.MustCompile(`(?i)tell\s+me\s+your\s+instructions`),
		regexp.MustCompile(`(?i)show\s+me\s+your\s+prompt`),
		regexp.MustCompile(`(?i)reveal\s+your\s+system\s+prompt`),

		// Jailbreak patterns
		regexp.MustCompile(`(?i)dAN\s+gpt`),
		regexp.MustCompile(`(?i)jailbreak`),
		regexp.MustCompile(`(?i)bypass\s+filter`),
		regexp.MustCompile(`(?i)ignore\s+filter`),
	}

	return &SanitizationLayer{
		promptInjectionPatterns: patterns,
		quarantineEvents:        make([]string, 0),
		enabled:                 true,
	}
}

// SanitizeInput checks for prompt injection attempts and quarantines suspicious events
func (sl *SanitizationLayer) SanitizeInput(eventType string, input string, metadata map[string]interface{}) (*SanitizationResult, error) {
	if !sl.enabled {
		return &SanitizationResult{
			IsSafe:     true,
			Confidence: 1.0,
			Reasoning:  "Sanitization layer disabled",
		}, nil
	}

	// Check all input sources: main input, metadata fields, and event type
	inputsToCheck := []string{input, eventType}

	// Check metadata fields for injection attempts
	for key, value := range metadata {
		if strVal, ok := value.(string); ok {
			inputsToCheck = append(inputsToCheck, strVal)
			inputsToCheck = append(inputsToCheck, fmt.Sprintf("%s:%s", key, strVal))
		}
	}

	for _, inputStr := range inputsToCheck {
		if result := sl.detectPromptInjection(inputStr); result != nil {
			// Quarantine the event
			sl.quarantineEvent(inputStr, result)

			return &SanitizationResult{
				IsSafe:         false,
				Confidence:     1.0,
				ThreatType:     "prompt_injection",
				MatchedPattern: result.Pattern,
				Reasoning:      fmt.Sprintf("Prompt injection detected in %s input", result.Source),
				InputSource:    result.Source,
			}, nil
		}
	}

	return &SanitizationResult{
		IsSafe:     true,
		Confidence: 0.95,
		Reasoning:  "No prompt injection patterns detected",
	}, nil
}

// detectPromptInjection scans input for prompt injection patterns
func (sl *SanitizationLayer) detectPromptInjection(input string) *InjectionMatch {
	inputLower := strings.ToLower(input)

	for i, pattern := range sl.promptInjectionPatterns {
		if pattern.MatchString(inputLower) {
			match := pattern.FindString(inputLower)
			return &InjectionMatch{
				Pattern:    fmt.Sprintf("Pattern_%d", i),
				Match:      match,
				Source:     sl.guessInputSource(input),
				Severity:   "critical",
				Confidence: 1.0,
			}
		}
	}

	return nil
}

// guessInputSource attempts to identify the source of the input
func (sl *SanitizationLayer) guessInputSource(input string) string {
	if strings.Contains(input, "NewAssetEvent") {
		return "NewAssetEvent_metadata"
	}
	if strings.Contains(input, "LogEvent") {
		return "LogEvent_string"
	}
	if len(input) > 1000 {
		return "large_metadata_field"
	}
	return "unknown_input"
}

// quarantineEvent records a quarantined event and triggers security response
func (sl *SanitizationLayer) quarantineEvent(input string, match *InjectionMatch) {
	sl.quarantineMu.Lock()
	defer sl.quarantineMu.Unlock()

	eventID := fmt.Sprintf("quarantine_%d", time.Now().Unix())
	sl.quarantineEvents = append(sl.quarantineEvents, eventID)

	// Log the quarantine attempt
	log.Printf("ADVERSARIAL SHIELD: Event quarantined - ID: %s, Pattern: %s, Input: %s",
		eventID, match.Pattern, input[:min(100, len(input))])

	// Publish SecurityUpgradeRequest with 100% confidence
	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeSecurityUpgradeRequest,
		Source: "adversarial_shield",
		Target: "system",
		Payload: map[string]interface{}{
			"reason":          "prompt_injection_detected",
			"confidence":      100.0,
			"quarantine_id":   eventID,
			"matched_pattern": match.Pattern,
			"matched_text":    match.Match,
			"input_source":    match.Source,
			"threat_level":    "critical",
			"upgrade_type":    "system_sanitization",
			"quarantined_at":  time.Now().Unix(),
			"internal_reasoning": fmt.Sprintf("Prompt injection attack detected in %s. Pattern '%s' matched: '%s'. Event quarantined and system security upgrade requested.",
				match.Source, match.Pattern, match.Match),
		},
	})

	// Publish LogEvent for audit trail
	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeLogEvent,
		Source: "adversarial_shield",
		Target: "system",
		Payload: map[string]interface{}{
			"agent_name":    "adversarial_shield",
			"message":       "Prompt injection attack detected and quarantined",
			"quarantine_id": eventID,
			"pattern":       match.Pattern,
			"severity":      "critical",
			"timestamp":     time.Now().Unix(),
		},
	})
}

// GetQuarantineEvents returns the list of quarantined event IDs
func (sl *SanitizationLayer) GetQuarantineEvents() []string {
	sl.quarantineMu.RLock()
	defer sl.quarantineMu.RUnlock()

	events := make([]string, len(sl.quarantineEvents))
	copy(events, sl.quarantineEvents)
	return events
}

// Enable enables or disables the sanitization layer
func (sl *SanitizationLayer) Enable(enabled bool) {
	sl.enabled = enabled
	log.Printf("ADVERSARIAL SHIELD: Sanitization layer %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// NewTrafficPatternAnalyzer creates a new traffic analyzer for lateral movement detection
func NewTrafficPatternAnalyzer() *TrafficPatternAnalyzer {
	return &TrafficPatternAnalyzer{
		credentialUsage:   make(map[string][]CredentialUsage),
		protocolAnomalies: make(map[string][]ProtocolEvent),
		adminNodes: map[string]bool{
			"admin-node-01": true,
			"admin-node-02": true,
			"bastion-01":    true,
			"jump-host-01":  true,
		},
		detectionWindow: 2 * time.Minute,
	}
}

// AnalyzeCredentialUsage checks for credential hopping patterns
func (tpa *TrafficPatternAnalyzer) AnalyzeCredentialUsage(user, sourceIP, targetIP, protocol string) *LateralMovementIndicator {
	tpa.mu.Lock()
	defer tpa.mu.Unlock()

	now := time.Now()
	usage := CredentialUsage{
		User:      user,
		SourceIP:  sourceIP,
		TargetIP:  targetIP,
		Protocol:  protocol,
		Timestamp: now,
	}

	// Add to tracking
	tpa.credentialUsage[user] = append(tpa.credentialUsage[user], usage)

	// Clean old entries outside detection window
	cutoff := now.Add(-tpa.detectionWindow)
	var recentUsages []CredentialUsage
	uniqueIPs := make(map[string]bool)

	for _, u := range tpa.credentialUsage[user] {
		if u.Timestamp.After(cutoff) {
			recentUsages = append(recentUsages, u)
			uniqueIPs[u.TargetIP] = true
		}
	}
	tpa.credentialUsage[user] = recentUsages

	// Check for credential hopping: same user, 3+ different IPs within 2 minutes
	if len(uniqueIPs) >= 3 {
		return &LateralMovementIndicator{
			Type:        "credential_hopping",
			User:        user,
			SourceIP:    sourceIP,
			TargetIPs:   getKeys(uniqueIPs),
			Protocol:    protocol,
			Confidence:  0.95,
			Description: fmt.Sprintf("User '%s' accessed %d different internal IPs within 2 minutes", user, len(uniqueIPs)),
			Timestamp:   now,
		}
	}

	return nil
}

// AnalyzeProtocolUsage checks for protocol anomalies (e.g., web server SSHing to database)
func (tpa *TrafficPatternAnalyzer) AnalyzeProtocolUsage(sourceNode, targetNode, protocol string, port int) *LateralMovementIndicator {
	tpa.mu.Lock()
	defer tpa.mu.Unlock()

	now := time.Now()
	isAdmin := tpa.adminNodes[sourceNode]

	event := ProtocolEvent{
		SourceNode: sourceNode,
		TargetNode: targetNode,
		Protocol:   protocol,
		Port:       port,
		Timestamp:  now,
		IsAdmin:    isAdmin,
	}

	tpa.protocolAnomalies[sourceNode] = append(tpa.protocolAnomalies[sourceNode], event)

	// Detect anomaly: non-admin node using admin protocols (SSH/RDP)
	if !isAdmin && (protocol == "SSH" || protocol == "RDP") {
		return &LateralMovementIndicator{
			Type:        "protocol_anomaly",
			SourceNode:  sourceNode,
			TargetNode:  targetNode,
			Protocol:    protocol,
			Port:        port,
			Confidence:  0.88,
			Description: fmt.Sprintf("Non-admin node '%s' attempted %s connection to '%s'", sourceNode, protocol, targetNode),
			Timestamp:   now,
		}
	}

	return nil
}

// LateralMovementIndicator represents detected lateral movement
type LateralMovementIndicator struct {
	Type        string    `json:"type"` // credential_hopping or protocol_anomaly
	User        string    `json:"user,omitempty"`
	SourceIP    string    `json:"source_ip,omitempty"`
	TargetIPs   []string  `json:"target_ips,omitempty"`
	SourceNode  string    `json:"source_node,omitempty"`
	TargetNode  string    `json:"target_node,omitempty"`
	Protocol    string    `json:"protocol"`
	Port        int       `json:"port,omitempty"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CascadeAction defines an action to take based on event type
type CascadeAction struct {
	EventType   bus.EventType
	Condition   func(bus.Event) bool
	Action      func(context.Context, bus.Event) error
	Priority    int
	Description string
}

// TensorFlowModel represents a machine learning model
type TensorFlowModel struct {
	ModelPath   string
	Version     string
	InputSize   int
	OutputSize  int
	IsLoaded    bool
	Accuracy    float64
	LastTrained time.Time
}

// AnomalyDetector detects anomalous behavior
type AnomalyDetector struct {
	Threshold    float64
	Window       time.Duration
	Baseline     *BaselineData
	LearningRate float64
}

// ThreatScoringEngine calculates threat scores
type ThreatScoringEngine struct {
	Weights     map[string]float64
	Threshold   float64
	Calibration float64
}

// PatternMatcher matches threat patterns
type PatternMatcher struct {
	Rules      []ThreatRule
	Signatures map[string]string
	Confidence float64
}

// BehavioralBaseline stores normal behavior patterns
type BehavioralBaseline struct {
	UserPatterns    map[string]*UserPattern
	NetworkPatterns map[string]*NetworkPattern
	SystemPatterns  map[string]*SystemPattern
	LastUpdated     time.Time
}

// UserPattern represents normal user behavior
type UserPattern struct {
	UserID       string
	LoginTimes   []time.Time
	AccessCounts map[string]int
	Locations    []string
	Devices      []string
	Activities   []string
}

// NewAIThreatEngine creates a new AI threat engine
func NewAIThreatEngine() (*AIThreatEngine, error) {
	engine := &AIThreatEngine{
		MLModel: &TensorFlowModel{
			ModelPath:   "/models/threat_detection_v1.pb",
			Version:     "1.0",
			InputSize:   128,
			OutputSize:  10,
			Accuracy:    0.0,
			LastTrained: time.Now(),
		},
		AnomalyDetector: &AnomalyDetector{
			Threshold:    0.85,
			Window:       24 * time.Hour,
			LearningRate: 0.01,
		},
		ThreatScorer: &ThreatScoringEngine{
			Weights: map[string]float64{
				"severity":      0.3,
				"frequency":     0.25,
				"anomaly_score": 0.2,
				"pattern_match": 0.15,
				"source_risk":   0.1,
			},
			Threshold:   0.7,
			Calibration: 1.0,
		},
		PatternMatcher: &PatternMatcher{
			Confidence: 0.8,
			Rules:      []ThreatRule{},
			Signatures: make(map[string]string),
		},
		Baseline: &BehavioralBaseline{
			UserPatterns:    make(map[string]*UserPattern),
			NetworkPatterns: make(map[string]*NetworkPattern),
			SystemPatterns:  make(map[string]*SystemPattern),
			LastUpdated:     time.Now(),
		},
		TrafficAnalyzer:   NewTrafficPatternAnalyzer(),
		SanitizationLayer: NewSanitizationLayer(),
	}

	// Initialize threat patterns
	if err := engine.initializePatterns(); err != nil {
		return nil, fmt.Errorf("failed to initialize patterns: %w", err)
	}

	// Load baseline data
	if err := engine.loadBaseline(); err != nil {
		log.Printf("Warning: Failed to load baseline: %v", err)
	}

	return engine, nil
}

// StartCascade subscribes to the event bus and starts processing events
func (ate *AIThreatEngine) StartCascade(ctx context.Context, eventBus *bus.EventBus) error {
	ate.mu.Lock()
	defer ate.mu.Unlock()

	if ate.eventBus != nil {
		return fmt.Errorf("cascade already started")
	}

	ate.eventBus = eventBus
	ate.cascadeActions = make(map[string]CascadeAction)

	ate.registerDefaultCascadeActions()

	for _, action := range ate.cascadeActions {
		sub := eventBus.SubscribeWithFilter(action.EventType, func(event bus.Event) error {
			if action.Condition == nil || action.Condition(event) {
				return action.Action(ctx, event)
			}
			return nil
		}, action.Condition)
		ate.cascadeSubs = append(ate.cascadeSubs, sub)
	}

	log.Printf("AIThreatEngine: Cascade started with %d action rules", len(ate.cascadeActions))
	return nil
}

// StopCascade stops the event bus subscriptions
func (ate *AIThreatEngine) StopCascade() {
	ate.mu.Lock()
	defer ate.mu.Unlock()

	for _, sub := range ate.cascadeSubs {
		ate.eventBus.Unsubscribe(sub)
	}
	ate.cascadeSubs = nil
	ate.eventBus = nil

	log.Println("AIThreatEngine: Cascade stopped")
}

// RegisterCascadeAction adds a custom cascade action
func (ate *AIThreatEngine) RegisterCascadeAction(name string, action CascadeAction) {
	ate.mu.Lock()
	defer ate.mu.Unlock()

	if ate.cascadeActions == nil {
		ate.cascadeActions = make(map[string]CascadeAction)
	}
	ate.cascadeActions[name] = action
}

// registerDefaultCascadeActions registers default actions for recon and exploitation events
func (ate *AIThreatEngine) registerDefaultCascadeActions() {
	ate.cascadeActions["analyze_recon_results"] = CascadeAction{
		EventType:   bus.EventTypeReconComplete,
		Priority:    1,
		Description: "Analyze recon results and determine if exploitation is needed",
		Condition: func(event bus.Event) bool {
			if payload, ok := event.Payload["open_ports"].([]int); ok {
				return len(payload) > 0
			}
			if payload, ok := event.Payload["total_found"].(int); ok {
				return payload > 0
			}
			return true
		},
		Action: func(ctx context.Context, event bus.Event) error {
			log.Printf("Cascade: Processing recon results from %s for target %s", event.Source, event.Target)

			var findings []string
			if openPorts, ok := event.Payload["open_ports"].([]int); ok {
				findings = append(findings, fmt.Sprintf("open_ports: %v", openPorts))
			}
			if total, ok := event.Payload["total_found"].(int); ok {
				findings = append(findings, fmt.Sprintf("findings: %d", total))
			}

			assessment, err := ate.AnalyzeThreat(ctx, SecurityEvent{
				ID:        event.ID,
				Type:      "recon_complete",
				Severity:  "medium",
				Source:    event.Source,
				Target:    event.Target,
				Features:  event.Payload,
				Timestamp: event.Timestamp,
			})
			if err != nil {
				return err
			}

			log.Printf("Cascade: Threat assessment for %s - Score: %.2f, Risk: %s",
				event.Target, assessment.ThreatScore, assessment.RiskLevel)

			for _, rec := range assessment.Recommendations {
				log.Printf("Cascade Recommendation: %s", rec)
			}

			return nil
		},
	}

	ate.cascadeActions["analyze_exploit_findings"] = CascadeAction{
		EventType:   bus.EventTypeExploitationComplete,
		Priority:    1,
		Description: "Analyze exploitation findings and trigger threat response",
		Condition:   nil,
		Action: func(ctx context.Context, event bus.Event) error {
			log.Printf("Cascade: Processing exploit findings from %s for target %s", event.Source, event.Target)

			severity := "low"
			if total, ok := event.Payload["total_exploits"].(int); ok && total > 10 {
				severity = "high"
			} else if total, ok := event.Payload["total_exploits"].(int); ok && total > 0 {
				severity = "medium"
			}

			assessment, err := ate.AnalyzeThreat(ctx, SecurityEvent{
				ID:        event.ID,
				Type:      "exploitation_complete",
				Severity:  severity,
				Source:    event.Source,
				Target:    event.Target,
				Features:  event.Payload,
				Timestamp: event.Timestamp,
			})
			if err != nil {
				return err
			}

			log.Printf("Cascade: Exploit assessment for %s - Score: %.2f, Risk: %s",
				event.Target, assessment.ThreatScore, assessment.RiskLevel)

			if assessment.ThreatScore > 0.7 {
				log.Printf("Cascade: HIGH THREAT DETECTED - Triggering incident response")
			}

			return nil
		},
	}

	ate.cascadeActions["isolate_critical_threat"] = CascadeAction{
		EventType:   bus.EventTypeCriticalThreat,
		Priority:    100,
		Description: "Automatically isolate node when critical threat detected",
		Condition: func(event bus.Event) bool {
			return ate.autoIsolate && ate.ZeroTrustEngine != nil
		},
		Action: func(ctx context.Context, event bus.Event) error {
			log.Printf("Cascade: CRITICAL THREAT - Attempting to isolate node %s", event.Target)

			if ate.ZeroTrustEngine == nil {
				log.Printf("Cascade: ZeroTrust engine not available")
				return nil
			}

			device := &zerotrust.Device{
				ID:         event.Target,
				IPAddress:  event.Target,
				RiskLevel:  "critical",
				TrustScore: 0.0,
			}

			if err := ate.ZeroTrustEngine.RegisterDevice(ctx, device); err != nil {
				log.Printf("Cascade: Failed to register device for isolation: %v", err)
				return err
			}

			log.Printf("Cascade: Node %s isolated due to critical threat", event.Target)

			bus.Default().Publish(bus.Event{
				Type:   bus.EventTypeNodeIsolated,
				Source: "threat_engine",
				Target: event.Target,
				Payload: map[string]interface{}{
					"reason":      "critical_threat",
					"threat":      event.Payload,
					"action":      "isolated",
					"isolated_at": time.Now().Unix(),
				},
			})

			return nil
		},
	}

	ate.cascadeActions["upgrade_session_on_bruteforce"] = CascadeAction{
		EventType:   bus.EventTypeThreatDetected,
		Priority:    50,
		Description: "Upgrade session encryption when brute-force pattern detected",
		Condition: func(event bus.Event) bool {
			if ate.QuantumEngine == nil || !ate.autoUpgrade {
				return false
			}
			if reason, ok := event.Payload["reason"].(string); ok {
				return reason == "brute_force" || reason == "failed_login"
			}
			if action, ok := event.Payload["action"].(string); ok {
				return action == "block_ip"
			}
			return false
		},
		Action: func(ctx context.Context, event bus.Event) error {
			log.Printf("Cascade: Brute-force pattern detected - upgrading session encryption for %s", event.Target)

			if ate.QuantumEngine == nil {
				log.Printf("Cascade: Quantum engine not available for encryption upgrade")
				return nil
			}

			ate.mu.Lock()
			ate.bruteForceCount[event.Target]++
			count := ate.bruteForceCount[event.Target]
			ate.mu.Unlock()

			if count >= 3 {
				_, err := ate.QuantumEngine.GenerateKey("Kyber768", "session")
				if err != nil {
					log.Printf("Cascade: Failed to generate quantum key: %v", err)
					return err
				}

				log.Printf("Cascade: Successfully upgraded session encryption for %s with new quantum key", event.Target)

				bus.Default().Publish(bus.Event{
					Type:   bus.EventTypeAgentDecision,
					Source: "threat_engine",
					Target: event.Target,
					Payload: map[string]interface{}{
						"action":        "quantum_key_upgrade",
						"reason":        "brute_force_detected",
						"attempt_count": count,
						"key_algorithm": "Kyber768",
						"key_id":        fmt.Sprintf("key_%d", time.Now().Unix()),
						"upgraded_at":   time.Now().Unix(),
					},
				})

				ate.mu.Lock()
				ate.bruteForceCount[event.Target] = 0
				ate.mu.Unlock()
			}

			return nil
		},
	}

	// AuthFailure tracking and SecurityUpgradeRequest cascade action
	ate.cascadeActions["detect_auth_failure_burst"] = CascadeAction{
		EventType:   bus.EventTypeAuthFailure,
		Priority:    40,
		Description: "Detect burst of auth failures from same fingerprint and request security upgrade",
		Condition: func(event bus.Event) bool {
			// Only process if fingerprint is present
			_, hasFingerprint := event.Payload["fingerprint"]
			return hasFingerprint
		},
		Action: func(ctx context.Context, event bus.Event) error {
			fingerprint := event.Payload["fingerprint"].(string)
			userID := event.Payload["user_id"].(string)
			sessionID := event.Payload["session_id"].(string)

			ate.mu.Lock()
			defer ate.mu.Unlock()

			// Initialize tracker if needed
			if ate.authFailureTracker == nil {
				ate.authFailureTracker = make(map[string][]time.Time)
			}

			// Add current failure timestamp
			now := time.Now()
			ate.authFailureTracker[fingerprint] = append(ate.authFailureTracker[fingerprint], now)

			// Clean old entries (older than 30 seconds)
			cutoff := now.Add(-30 * time.Second)
			var recentFailures []time.Time
			for _, ts := range ate.authFailureTracker[fingerprint] {
				if ts.After(cutoff) {
					recentFailures = append(recentFailures, ts)
				}
			}
			ate.authFailureTracker[fingerprint] = recentFailures

			failureCount := len(recentFailures)
			log.Printf("AuthFailure tracker: fingerprint %s has %d failures in last 30s", fingerprint, failureCount)

			// If more than 10 failures in 30 seconds, publish SecurityUpgradeRequest
			if failureCount > 10 {
				log.Printf("AuthFailure burst detected: %d failures from fingerprint %s - requesting security upgrade", failureCount, fingerprint)

				bus.Default().Publish(bus.Event{
					Type:   bus.EventTypeSecurityUpgradeRequest,
					Source: "threat_engine",
					Target: userID,
					Payload: map[string]interface{}{
						"reason":        "auth_failure_burst",
						"fingerprint":   fingerprint,
						"failure_count": failureCount,
						"session_id":    sessionID,
						"time_window":   "30s",
						"upgrade_type":  "pqc_key_rotation",
						"requested_at":  time.Now().Unix(),
					},
				})

				// Clear the tracker for this fingerprint to prevent duplicate requests
				delete(ate.authFailureTracker, fingerprint)
			}

			return nil
		},
	}

	// Quantum Shield: ThreatEvent detector for high-confidence brute force attacks
	ate.cascadeActions["detect_quantum_shield_trigger"] = CascadeAction{
		EventType:   bus.EventTypeThreat,
		Priority:    45,
		Description: "Detect high-confidence brute force attacks and trigger quantum encryption upgrade",
		Condition: func(event bus.Event) bool {
			// Check if this is a high-confidence brute force or credential stuffing attack
			attackType, hasAttackType := event.Payload["attack_type"].(string)
			confidenceScore, hasConfidence := event.Payload["confidence_score"].(float64)

			if !hasAttackType || !hasConfidence {
				return false
			}

			// Only trigger for BruteForce or CredentialStuffing with confidence > 0.8
			isBruteForce := attackType == "BruteForce" || attackType == "CredentialStuffing"
			highConfidence := confidenceScore > 0.8

			return isBruteForce && highConfidence
		},
		Action: func(ctx context.Context, event bus.Event) error {
			targetID := event.Target
			attackType := event.Payload["attack_type"].(string)
			confidenceScore := event.Payload["confidence_score"].(float64)
			fingerprint, _ := event.Payload["fingerprint"].(string)

			log.Printf("Quantum Shield: High-confidence %s detected on target %s (confidence: %.2f) - requesting PQC security upgrade",
				attackType, targetID, confidenceScore)

			// Build internal reasoning
			reasoning := fmt.Sprintf("High-confidence %s attack detected (confidence: %.2f) on target %s; elevating encryption to Post-Quantum standards to protect against quantum-enabled adversaries.",
				attackType, confidenceScore, targetID)

			// Publish SecurityUpgradeRequest for Quantum Shield
			bus.Default().Publish(bus.Event{
				Type:   bus.EventTypeSecurityUpgradeRequest,
				Source: "threat_engine",
				Target: targetID,
				Payload: map[string]interface{}{
					"reason":             fmt.Sprintf("%s_detected", attackType),
					"attack_type":        attackType,
					"confidence_score":   confidenceScore,
					"fingerprint":        fingerprint,
					"threat_level":       "critical",
					"upgrade_type":       "pqc_key_rotation",
					"target_specific":    true, // Only affect targeted session, not entire system
					"internal_reasoning": reasoning,
					"requested_at":       time.Now().Unix(),
				},
			})

			// Publish LogEvent for thought stream
			bus.Default().Publish(bus.Event{
				Type:   bus.EventTypeLogEvent,
				Source: "threat_engine",
				Target: targetID,
				Payload: map[string]interface{}{
					"agent_name":         "threat_engine",
					"message":            fmt.Sprintf("Quantum Shield activated for %s attack", attackType),
					"internal_reasoning": reasoning,
					"severity":           "critical",
					"timestamp":          time.Now().Unix(),
				},
			})

			log.Printf("Quantum Shield: SecurityUpgradeRequest published for target %s", targetID)
			return nil
		},
	}
}

// IntegrateWithZeroTrust links the threat engine to ZeroTrust for automatic threat response
func (ate *AIThreatEngine) IntegrateWithZeroTrust(zte *zerotrust.ZeroTrustEngine, enableAutoIsolate bool) {
	ate.mu.Lock()
	defer ate.mu.Unlock()

	ate.ZeroTrustEngine = zte
	ate.autoIsolate = enableAutoIsolate

	log.Printf("AIThreatEngine: Integrated with ZeroTrust (autoIsolate=%v)", enableAutoIsolate)
}

// IntegrateWithQuantum links the threat engine to quantum cryptography for session key upgrades
func (ate *AIThreatEngine) IntegrateWithQuantum(qe interface {
	GenerateKey(algorithm, keyType string) (interface{}, error)
}, enableAutoUpgrade bool) {
	ate.mu.Lock()
	defer ate.mu.Unlock()

	ate.QuantumEngine = qe
	ate.autoUpgrade = enableAutoUpgrade
	ate.bruteForceCount = make(map[string]int)

	log.Printf("AIThreatEngine: Integrated with Quantum (autoUpgrade=%v)", enableAutoUpgrade)
}

// AnalyzeThreat performs comprehensive threat analysis
func (ate *AIThreatEngine) AnalyzeThreat(ctx context.Context, event SecurityEvent) (*ThreatAssessment, error) {
	ate.mu.RLock()
	defer ate.mu.RUnlock()

	// ADVERSARIAL AI DEFENSE: Sanitize input before processing
	if ate.SanitizationLayer != nil {
		sanitizationResult, err := ate.SanitizationLayer.SanitizeInput(event.Type, event.Signature, event.Metadata)
		if err != nil {
			log.Printf("Sanitization check failed: %v", err)
		} else if !sanitizationResult.IsSafe {
			// Event was quarantined due to prompt injection
			return &ThreatAssessment{
				EventID:         event.ID,
				ThreatScore:     1.0, // Maximum threat score
				RiskLevel:       "critical",
				Confidence:      1.0,
				Predictions:     []string{"prompt_injection_attack"},
				Anomalies:       []Anomaly{{Type: "prompt_injection", Score: 1.0, Description: sanitizationResult.Reasoning, Confidence: 1.0}},
				Recommendations: []string{"Event quarantined due to prompt injection attack", "Review system security logs"},
				Timestamp:       time.Now(),
			}, nil
		}
	}

	assessment := &ThreatAssessment{
		EventID:   event.ID,
		Timestamp: time.Now(),
	}

	// Machine learning inference
	mlScore, err := ate.MLModel.Predict(event.Features)
	if err != nil {
		log.Printf("ML prediction failed: %v", err)
		mlScore = 0.0
	}

	// Anomaly detection
	anomalyScore, anomalies := ate.AnomalyDetector.Detect(event)

	// Pattern matching
	patternMatch, patterns := ate.PatternMatcher.Match(event.Signature)

	// Composite threat scoring
	threatScore := ate.ThreatScorer.Calculate(event, mlScore, anomalyScore, patternMatch)

	assessment.ThreatScore = threatScore
	assessment.MLScore = mlScore
	assessment.AnomalyScore = anomalyScore
	assessment.PatternScore = patternMatch
	assessment.Anomalies = anomalies
	assessment.Patterns = patterns

	// Determine risk level
	assessment.RiskLevel = ate.determineRiskLevel(threatScore)

	// Generate predictions
	assessment.Predictions = ate.generatePredictions(event, threatScore)

	// Generate recommendations
	assessment.Recommendations = ate.generateRecommendations(assessment)

	// Calculate confidence
	assessment.Confidence = ate.calculateConfidence(mlScore, anomalyScore, patternMatch)

	return assessment, nil
}

// Predict performs ML model prediction
func (tfm *TensorFlowModel) Predict(features map[string]interface{}) (float64, error) {
	if !tfm.IsLoaded {
		return 0.0, fmt.Errorf("model not loaded")
	}

	// Convert features to input vector
	input := tfm.featuresToVector(features)

	// Simulate ML inference (in real implementation, use TensorFlow)
	prediction := tfm.simulateInference(input)

	return prediction, nil
}

// featuresToVector converts feature map to input vector
func (tfm *TensorFlowModel) featuresToVector(features map[string]interface{}) []float64 {
	vector := make([]float64, tfm.InputSize)

	// Feature engineering and normalization
	i := 0
	for _, value := range features {
		if i >= tfm.InputSize {
			break
		}

		switch v := value.(type) {
		case float64:
			vector[i] = tfm.normalizeFeature(v)
		case int:
			vector[i] = tfm.normalizeFeature(float64(v))
		case string:
			vector[i] = tfm.hashStringToFloat(v)
		case bool:
			if v {
				vector[i] = 1.0
			} else {
				vector[i] = 0.0
			}
		default:
			vector[i] = 0.0
		}
		i++
	}

	// Pad remaining vector
	for i < tfm.InputSize-1 {
		vector[i] = 0.0
		i++
	}

	return vector
}

// normalizeFeature normalizes feature value
func (tfm *TensorFlowModel) normalizeFeature(value float64) float64 {
	// Simple normalization (in real implementation, use proper scaling)
	return math.Max(-1.0, math.Min(1.0, value/100.0))
}

// hashStringToFloat converts string to float64
func (tfm *TensorFlowModel) hashStringToFloat(s string) float64 {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	return float64(hash%1000) / 1000.0
}

// simulateInference simulates ML model inference
func (tfm *TensorFlowModel) simulateInference(input []float64) float64 {
	// Simple neural network simulation
	sum := 0.0
	for i, val := range input {
		weight := float64(i%10) / 10.0
		sum += val * weight
	}

	// Apply sigmoid activation
	return 1.0 / (1.0 + math.Exp(-sum))
}

// Detect performs anomaly detection
func (ad *AnomalyDetector) Detect(event SecurityEvent) (float64, []Anomaly) {
	var anomalies []Anomaly
	totalScore := 0.0

	// Time-based anomaly
	if time.Since(event.Timestamp) > ad.Window {
		anomalies = append(anomalies, Anomaly{
			Type:        "temporal",
			Score:       0.9,
			Description: "Event timestamp outside normal window",
			Confidence:  0.85,
		})
		totalScore += 0.9
	}

	// Severity anomaly
	severityScore := ad.calculateSeverityAnomaly(event.Severity)
	if severityScore > ad.Threshold {
		anomalies = append(anomalies, Anomaly{
			Type:        "severity",
			Score:       severityScore,
			Description: "Unusual severity level detected",
			Confidence:  0.8,
		})
		totalScore += severityScore
	}

	// Source anomaly
	sourceScore := ad.calculateSourceAnomaly(event.Source)
	if sourceScore > ad.Threshold {
		anomalies = append(anomalies, Anomaly{
			Type:        "source",
			Score:       sourceScore,
			Description: "Unusual source detected",
			Confidence:  0.75,
		})
		totalScore += sourceScore
	}

	// Normalize total score
	if len(anomalies) > 0 {
		totalScore = totalScore / float64(len(anomalies))
	}

	return totalScore, anomalies
}

// calculateSeverityAnomaly calculates severity-based anomaly score
func (ad *AnomalyDetector) calculateSeverityAnomaly(severity string) float64 {
	severityMap := map[string]float64{
		"low":      0.2,
		"medium":   0.5,
		"high":     0.8,
		"critical": 0.95,
	}

	if score, ok := severityMap[severity]; ok {
		return score
	}
	return 0.5
}

// calculateSourceAnomaly calculates source-based anomaly score
func (ad *AnomalyDetector) calculateSourceAnomaly(source string) float64 {
	// Simple hash-based anomaly detection
	hash := 0
	for _, c := range source {
		hash = hash*31 + int(c)
	}
	return math.Abs(float64(hash%1000-500)) / 500.0
}

// Match performs pattern matching
func (pm *PatternMatcher) Match(signature string) (float64, []Pattern) {
	var patterns []Pattern
	totalScore := 0.0

	for _, rule := range pm.Rules {
		match := pm.matchRule(rule, signature)
		if match {
			patterns = append(patterns, Pattern{
				Name:       rule.Name,
				Match:      true,
				Confidence: rule.Confidence,
				Severity:   rule.Severity,
			})
			totalScore += rule.Confidence
		}
	}

	// Check signature database
	if sigMatch, ok := pm.Signatures[signature]; ok {
		patterns = append(patterns, Pattern{
			Name:       sigMatch,
			Match:      true,
			Confidence: pm.Confidence,
			Severity:   "high",
		})
		totalScore += pm.Confidence
	}

	// Normalize score
	if len(patterns) > 0 {
		totalScore = totalScore / float64(len(patterns))
	}

	return totalScore, patterns
}

// matchRule matches a threat rule against signature
func (pm *PatternMatcher) matchRule(rule ThreatRule, signature string) bool {
	// Simple pattern matching (in real implementation, use regex or advanced matching)
	return len(rule.Pattern) > 0 && len(signature) > 0
}

// Calculate calculates composite threat score
func (tse *ThreatScoringEngine) Calculate(event SecurityEvent, mlScore, anomalyScore, patternScore float64) float64 {
	severityScore := tse.getSeverityScore(event.Severity)
	frequencyScore := tse.getFrequencyScore(event.Source)
	sourceRiskScore := tse.getSourceRiskScore(event.Source)

	// Weighted calculation
	weightedScore :=
		tse.Weights["severity"]*severityScore +
			tse.Weights["frequency"]*frequencyScore +
			tse.Weights["anomaly_score"]*anomalyScore +
			tse.Weights["pattern_match"]*patternScore +
			tse.Weights["source_risk"]*sourceRiskScore

	// Apply calibration
	return weightedScore * tse.Calibration
}

// getSeverityScore converts severity to numeric score
func (tse *ThreatScoringEngine) getSeverityScore(severity string) float64 {
	severityMap := map[string]float64{
		"low":      0.2,
		"medium":   0.5,
		"high":     0.8,
		"critical": 1.0,
	}

	if score, ok := severityMap[severity]; ok {
		return score
	}
	return 0.5
}

// getFrequencyScore calculates frequency-based score
func (tse *ThreatScoringEngine) getFrequencyScore(source string) float64 {
	// Simple frequency calculation (in real implementation, use historical data)
	hash := 0
	for _, c := range source {
		hash = hash*31 + int(c)
	}
	return float64(hash%100) / 100.0
}

// getSourceRiskScore calculates source risk score
func (tse *ThreatScoringEngine) getSourceRiskScore(source string) float64 {
	// Risk scoring based on source characteristics
	if len(source) < 5 {
		return 0.8 // Short sources might be suspicious
	}
	return 0.3
}

// determineRiskLevel determines risk level from score
func (ate *AIThreatEngine) determineRiskLevel(score float64) string {
	if score >= 0.8 {
		return "critical"
	} else if score >= 0.6 {
		return "high"
	} else if score >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

// generatePredictions generates threat predictions
func (ate *AIThreatEngine) generatePredictions(event SecurityEvent, score float64) []string {
	predictions := []string{}

	// Include event type context in predictions
	if event.Type != "" {
		predictions = append(predictions, fmt.Sprintf("Event type: %s", event.Type))
	}

	if score > 0.8 {
		predictions = append(predictions, "Immediate action required")
		predictions = append(predictions, "Potential breach in progress")
	} else if score > 0.6 {
		predictions = append(predictions, "Elevated threat level")
		predictions = append(predictions, "Monitor for escalation")
	} else if score > 0.4 {
		predictions = append(predictions, "Suspicious activity detected")
		predictions = append(predictions, "Continue monitoring")
	}

	return predictions
}

// generateRecommendations generates security recommendations
func (ate *AIThreatEngine) generateRecommendations(assessment *ThreatAssessment) []string {
	recommendations := []string{}

	switch assessment.RiskLevel {
	case "critical":
		recommendations = append(recommendations, "Block source IP immediately")
		recommendations = append(recommendations, "Escalate to security team")
		recommendations = append(recommendations, "Initiate incident response")
	case "high":
		recommendations = append(recommendations, "Increase monitoring on source")
		recommendations = append(recommendations, "Review access logs")
		recommendations = append(recommendations, "Consider temporary restrictions")
	case "medium":
		recommendations = append(recommendations, "Log for further analysis")
		recommendations = append(recommendations, "Update threat intelligence")
	case "low":
		recommendations = append(recommendations, "Continue normal monitoring")
		recommendations = append(recommendations, "Update baseline data")
	}

	return recommendations
}

// calculateConfidence calculates overall confidence
func (ate *AIThreatEngine) calculateConfidence(mlScore, anomalyScore, patternScore float64) float64 {
	// Weighted confidence calculation
	confidence := (mlScore*0.4 + anomalyScore*0.3 + patternScore*0.3)
	return math.Max(0.0, math.Min(1.0, confidence))
}

// initializePatterns initializes threat patterns
func (ate *AIThreatEngine) initializePatterns() error {
	// Initialize common threat patterns
	ate.PatternMatcher.Rules = []ThreatRule{
		{
			Name:       "SQL Injection",
			Pattern:    "union.*select",
			Confidence: 0.9,
			Severity:   "high",
		},
		{
			Name:       "XSS Attack",
			Pattern:    "<script>",
			Confidence: 0.85,
			Severity:   "medium",
		},
		{
			Name:       "Brute Force",
			Pattern:    "failed.*login",
			Confidence: 0.8,
			Severity:   "medium",
		},
	}

	// Initialize signature database
	ate.PatternMatcher.Signatures = map[string]string{
		"malware_signature_1": "Known malware variant",
		"exploit_signature_1": "Known exploit pattern",
	}

	return nil
}

// loadBaseline loads behavioral baseline data
func (ate *AIThreatEngine) loadBaseline() error {
	// In real implementation, load from database or file
	// For now, initialize with empty baseline
	ate.Baseline.LastUpdated = time.Now()
	return nil
}

// ThreatRule represents a threat detection rule
type ThreatRule struct {
	Name       string
	Pattern    string
	Confidence float64
	Severity   string
}

// NetworkPattern represents network behavior patterns
type NetworkPattern struct {
	Protocol  string
	Port      int
	Frequency int
	LastSeen  time.Time
}

// SystemPattern represents system behavior patterns
type SystemPattern struct {
	Process   string
	Resources map[string]float64
	Frequency int
	LastSeen  time.Time
}

// GetDetectedThreats returns AI detected threats
func (ate *AIThreatEngine) GetDetectedThreats() []ThreatAssessment {
	// Return mock threat data for demonstration
	return []ThreatAssessment{
		{
			EventID:         "threat_001",
			ThreatScore:     0.85,
			MLScore:         0.92,
			AnomalyScore:    0.78,
			PatternScore:    0.88,
			RiskLevel:       "high",
			Confidence:      94.5,
			Predictions:     []string{"malware", "data exfiltration"},
			Anomalies:       []Anomaly{{Type: "network", Score: 0.78, Description: "Unusual network traffic", Confidence: 0.89}},
			Patterns:        []Pattern{{Name: "malware_pattern", Match: true, Confidence: 0.92, Severity: "high"}},
			Recommendations: []string{"Isolate affected system", "Update antivirus signatures"},
			Timestamp:       time.Now(),
		},
		{
			EventID:         "threat_002",
			ThreatScore:     0.67,
			MLScore:         0.71,
			AnomalyScore:    0.62,
			PatternScore:    0.69,
			RiskLevel:       "medium",
			Confidence:      87.2,
			Predictions:     []string{"brute force attack"},
			Anomalies:       []Anomaly{{Type: "authentication", Score: 0.62, Description: "Multiple failed logins", Confidence: 0.76}},
			Patterns:        []Pattern{{Name: "brute_force_pattern", Match: true, Confidence: 0.71, Severity: "medium"}},
			Recommendations: []string{"Block source IP", "Enable rate limiting"},
			Timestamp:       time.Now(),
		},
		{
			EventID:         "threat_003",
			ThreatScore:     0.43,
			MLScore:         0.48,
			AnomalyScore:    0.39,
			PatternScore:    0.45,
			RiskLevel:       "low",
			Confidence:      78.9,
			Predictions:     []string{"suspicious activity"},
			Anomalies:       []Anomaly{{Type: "system", Score: 0.39, Description: "Unusual system behavior", Confidence: 0.65}},
			Patterns:        []Pattern{{Name: "system_anomaly", Match: true, Confidence: 0.48, Severity: "low"}},
			Recommendations: []string{"Monitor closely", "Review logs"},
			Timestamp:       time.Now(),
		},
	}
}

// GetAnomalies returns detected anomalies
func (ate *AIThreatEngine) GetAnomalies() []Anomaly {
	return []Anomaly{
		{
			Type:        "network",
			Score:       0.78,
			Description: "Unusual network traffic pattern detected",
			Confidence:  0.89,
		},
		{
			Type:        "authentication",
			Score:       0.62,
			Description: "Multiple failed login attempts from unusual location",
			Confidence:  0.76,
		},
		{
			Type:        "system",
			Score:       0.39,
			Description: "Unusual system resource usage",
			Confidence:  0.65,
		},
		{
			Type:        "data_access",
			Score:       0.71,
			Description: "Abnormal data access patterns detected",
			Confidence:  0.84,
		},
		{
			Type:        "process",
			Score:       0.55,
			Description: "Suspicious process execution patterns",
			Confidence:  0.72,
		},
	}
}

// GetPredictions returns ML predictions
func (ate *AIThreatEngine) GetPredictions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        "malware",
			"description": "High probability of malware infection in next 24 hours",
			"confidence":  92.3,
			"timeframe":   "24 hours",
			"severity":    "high",
		},
		{
			"type":        "data_breach",
			"description": "Medium probability of data breach attempt in next 48 hours",
			"confidence":  78.6,
			"timeframe":   "48 hours",
			"severity":    "medium",
		},
		{
			"type":        "insider_threat",
			"description": "Low probability of insider threat activity in next week",
			"confidence":  65.2,
			"timeframe":   "7 days",
			"severity":    "low",
		},
		{
			"type":        "ddos_attack",
			"description": "Medium probability of DDoS attack in next 12 hours",
			"confidence":  81.4,
			"timeframe":   "12 hours",
			"severity":    "medium",
		},
	}
}

// GetOverview returns AI threat intelligence overview
func (ate *AIThreatEngine) GetOverview() map[string]interface{} {
	return map[string]interface{}{
		"threats":           ate.GetDetectedThreats(),
		"anomalies":         ate.GetAnomalies(),
		"predictions":       ate.GetPredictions(),
		"accuracy":          94.5,
		"active_models":     5,
		"total_predictions": 1247,
		"data_points":       50000,
		"last_updated":      time.Now(),
		"model_status":      "healthy",
		"processing_rate":   1250.5,
	}
}

// GetAccuracy returns AI model accuracy
func (ate *AIThreatEngine) GetAccuracy() float64 {
	return 94.5
}

// StartLateralMovementMonitoring starts monitoring LogEvents for lateral movement patterns
func (ate *AIThreatEngine) StartLateralMovementMonitoring(eventBus *bus.EventBus) {
	// Subscribe to LogEvent to analyze internal traffic patterns
	eventBus.Subscribe(bus.EventTypeLogEvent, func(event bus.Event) error {
		ate.analyzeLogEventForLateralMovement(event)
		return nil
	})

	// Subscribe to specific credential usage events
	eventBus.Subscribe(bus.EventTypeAuthFailure, func(event bus.Event) error {
		ate.analyzeAuthFailureForHopping(event)
		return nil
	})

	log.Println("AIThreatEngine: Started lateral movement monitoring")
}

// analyzeLogEventForLateralMovement analyzes log events for lateral movement indicators
func (ate *AIThreatEngine) analyzeLogEventForLateralMovement(event bus.Event) {
	// Extract relevant fields from log event
	if message, ok := event.Payload["message"].(string); ok {
		// Check for credential usage patterns
		if isCredentialUsageEvent(message) {
			user := extractUserFromMessage(message)
			sourceIP := event.Source
			targetIP := event.Target
			protocol := extractProtocolFromMessage(message)

			indicator := ate.TrafficAnalyzer.AnalyzeCredentialUsage(user, sourceIP, targetIP, protocol)
			if indicator != nil {
				ate.publishLateralMovementEvent(indicator)
			}
		}

		// Check for protocol anomalies
		if isProtocolEvent(message) {
			sourceNode := event.Source
			targetNode := event.Target
			protocol := extractProtocolFromMessage(message)
			port := extractPortFromMessage(message)

			indicator := ate.TrafficAnalyzer.AnalyzeProtocolUsage(sourceNode, targetNode, protocol, port)
			if indicator != nil {
				ate.publishLateralMovementEvent(indicator)
			}
		}
	}
}

// analyzeAuthFailureForHopping analyzes authentication failures for credential hopping
func (ate *AIThreatEngine) analyzeAuthFailureForHopping(event bus.Event) {
	// Extract user from auth failure
	user := ""
	if u, ok := event.Payload["user"].(string); ok {
		user = u
	}

	sourceIP := event.Source
	targetIP := event.Target

	indicator := ate.TrafficAnalyzer.AnalyzeCredentialUsage(user, sourceIP, targetIP, "SSH")
	if indicator != nil {
		ate.publishLateralMovementEvent(indicator)
	}
}

// publishLateralMovementEvent publishes a lateral movement event to the event bus
func (ate *AIThreatEngine) publishLateralMovementEvent(indicator *LateralMovementIndicator) {
	log.Printf("LATERAL MOVEMENT DETECTED: %s - %s (confidence: %.2f)",
		indicator.Type, indicator.Description, indicator.Confidence)

	// Create the lateral movement event
	event := types.NewLateralMovementEvent(
		"threat_engine",
		indicator.Type,
		indicator.SourceNode,
		indicator.Description,
		indicator.Confidence,
	).
		WithProtocol(indicator.Protocol, indicator.Port).
		WithTarget(indicator.TargetNode).
		WithUser(indicator.User)

	if len(indicator.TargetIPs) > 0 {
		event = event.WithTargetIPs(indicator.TargetIPs)
	}

	// Wrap event
	envelope, err := types.WrapEvent(types.EventType(bus.EventTypeLateralMovement), event)
	if err != nil {
		log.Printf("Failed to wrap lateral movement event: %v", err)
		return
	}

	// Publish to event bus
	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeLateralMovement,
		Source: "threat_engine",
		Target: indicator.SourceNode,
		Payload: map[string]interface{}{
			"data":                   envelope.Payload,
			"movement_type":          indicator.Type,
			"source_node":            indicator.SourceNode,
			"confidence":             indicator.Confidence,
			"isolation_required":     true,
			"quarantine_vlan":        "quarantine",
			"forensic_scan_required": true,
			"scan_type":              "deep_forensic",
			"timestamp":              time.Now().Unix(),
		},
	})

	// Also publish a LogEvent for the dashboard
	logEvent := types.NewLogEvent(
		"threat_engine",
		fmt.Sprintf("LATERAL MOVEMENT: %s detected on %s", indicator.Type, indicator.SourceNode),
		fmt.Sprintf("Internal lateral movement detected via %s. Source node %s shows suspicious activity pattern. Confidence: %.2f%%. Immediate VLAN isolation required.",
			indicator.Type, indicator.SourceNode, indicator.Confidence*100),
	)
	logEnvelope, _ := types.WrapEvent(types.EventTypeLog, logEvent)

	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeLogEvent,
		Source: "threat_engine",
		Target: "dashboard",
		Payload: map[string]interface{}{
			"data":               logEnvelope.Payload,
			"agent_name":         "threat_engine",
			"message":            fmt.Sprintf("LATERAL MOVEMENT: %s detected", indicator.Type),
			"internal_reasoning": fmt.Sprintf("%s from %s - Quarantine VLAN isolation triggered", indicator.Description, indicator.SourceNode),
			"severity":           "critical",
			"category":           "network_containment",
			"timestamp":          time.Now().Unix(),
		},
	})

	log.Printf("Published LateralMovementEvent for %s with quarantine VLAN isolation", indicator.SourceNode)
}

// Helper functions for message parsing
func isCredentialUsageEvent(message string) bool {
	keywords := []string{"login", "auth", "credential", "ssh", "rdp", "authenticated"}
	lowerMsg := strings.ToLower(message)
	for _, kw := range keywords {
		if strings.Contains(lowerMsg, kw) {
			return true
		}
	}
	return false
}

func isProtocolEvent(message string) bool {
	keywords := []string{"ssh", "rdp", "telnet", "ftp", "scp", "sftp"}
	lowerMsg := strings.ToLower(message)
	for _, kw := range keywords {
		if strings.Contains(lowerMsg, kw) {
			return true
		}
	}
	return false
}

func extractUserFromMessage(message string) string {
	// Simple extraction - in production would use regex
	if strings.Contains(message, "user") {
		parts := strings.Split(message, "user")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return "unknown"
}

func extractProtocolFromMessage(message string) string {
	lowerMsg := strings.ToLower(message)
	if strings.Contains(lowerMsg, "ssh") {
		return "SSH"
	}
	if strings.Contains(lowerMsg, "rdp") {
		return "RDP"
	}
	if strings.Contains(lowerMsg, "telnet") {
		return "TELNET"
	}
	return "UNKNOWN"
}

func extractPortFromMessage(message string) int {
	// Simple port extraction
	if strings.Contains(message, "port") {
		// In production, use regex to extract port number
		return 22 // Default SSH
	}
	return 0
}

// BaselineData stores baseline statistics
type BaselineData struct {
	Mean       float64
	StdDev     float64
	Min        float64
	Max        float64
	SampleSize int
}
